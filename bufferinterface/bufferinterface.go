// Package fileinterface provides access to the underlying operating system files.
// Files are organized in blocks.
package fileinterface

import (
	"errors"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/henrik-peters/DBTI/fileinterface"
)

// FileName of the database file
const fileName = "simple.db"

// FileID of the database file
var fileID = fileinterface.FID(-1)

// PageSize of a page in bytes
const PageSize = fileinterface.Blocksize

// BlocksPerPage will be the number of blocks in page
const BlocksPerPage = PageSize / fileinterface.Blocksize

// PufferSize will set the maximum number of pages in the puffer
const PufferSize = 128

// Page data
type Page [PageSize]byte

// PageFrame with the puffer data and the page data
type PageFrame struct {
	page      *Page
	pageNo    int
	isFixed   bool
	isUpdated bool
}

// CacheDisplacement definies the strategy to find slots in the puffer
type CacheDisplacement int

const (
	// RANDOM will select a random position in the puffer
	RANDOM CacheDisplacement = iota

	// FIFO will select the puffer position based on the
	// First In – First Out principle (Queue)
	FIFO

	// LRU will select the puffer position based on the
	// Least recently used principle (based on the time)
	LRU
)

// The Current strategy to find slots in the puffer
const cacheDisplacementStrategy = RANDOM

// Contains the currently puffered pages
var puffer [PufferSize]PageFrame

// The pageMap will map the pageNumbers to pufferNumbers
var pageMap = make(map[int]int)

// CacheMissCounter will count the number of pages that must be loaded from disk (or created)
var CacheMissCounter = 0

// CacheHitCounter will count the number of pages found in the puffer on a request
var CacheHitCounter = 0

// Request page with number pageNo. Returns pointer to page data in system buffer (and err is nil)
// If unsuccessful, return nil and an error value describing the error.
func Request(PageNo int) (*Page, error) {
	var pufferIndex = pageMap[PageNo]

	if pufferIndex != 0 && puffer[pufferIndex].page != nil {
		//cache hit; load the page from the puffer
		CacheHitCounter++
		return puffer[pufferIndex].page, nil
	}

	//Cache miss; check for load or creation
	CacheMissCounter++
	var err error
	var blockLength = -1

	if error := initFileSystem(PageNo); error != nil {
		return nil, error

	} else if blockLength, err = fileinterface.Length(fileID); err != nil {
		return nil, err
	}

	//Divide by the number of blocks per page to get the pageLength
	var pageLength = blockLength / BlocksPerPage

	//Create a new page
	var newPage PageFrame
	var newPageData Page

	newPage.page = &newPageData

	if PageNo < pageLength {
		//Page exists on disk; load the page
		for blockIndex := 0; blockIndex < BlocksPerPage; blockIndex++ {
			if fileBlock, err := fileinterface.Read(fileID, 0); err == nil {

				//Copy the data into the page
				for byteIndex := 0; byteIndex < fileinterface.Blocksize; byteIndex++ {
					newPage.page[byteIndex*blockIndex] = fileBlock[byteIndex]
				}

			} else {
				return nil, err
			}
		}
	}

	//Find a slot for the page
	newPage.pageNo = PageNo
	pufferIndex, err = requestPufferSlot()

	if err != nil {
		return nil, err
	}

	//Store the new slot position in the pageMap
	pageMap[PageNo] = pufferIndex

	//Save the page in the puffer
	puffer[pufferIndex] = newPage
	return newPage.page, nil
}

// Create a new free puffer slot based on the selected cache displacement strategy
func requestPufferSlot() (int, error) {

	switch cacheDisplacementStrategy {
	case RANDOM:
		rand.Seed(time.Now().Unix())
		randomIndex := rand.Intn(PufferSize+1) + 1

		// Check if there is an old page in the slot that must be written to disk
		if puffer[randomIndex].pageNo != 0 && puffer[randomIndex].isUpdated {
			if err := Write(puffer[randomIndex].pageNo); err != nil {
				return 0, err
			}
		}

		return randomIndex, nil

	case FIFO:
		// TODO
		return 0, nil

	case LRU:
		// TODO
		return 0, nil

	default:
		panic("unregistered cache displacement strategy selected")
	}
}

// Fix the page pageNo as pinned. It's pointer will stay valid and the page
// is nerver removed from the system bufffer.
// If unsuccessful, return an error value describing the error.
func Fix(pageNo int) error {
	var pufferIndex = pageMap[pageNo]

	if !pageAvailInPuffer(pageNo) {
		return errors.New("PageNotPuffered")
	}

	puffer[pufferIndex].isFixed = true
	return nil
}

// UnFix the page pageNo as no longer pinned. The page might be subsequently
// removed from the system buffer.
// If unsuccessful, return an error value describing the error.
func UnFix(pageNo int) error {
	var pufferIndex = pageMap[pageNo]

	if !pageAvailInPuffer(pageNo) {
		return errors.New("PageNotPuffered")
	}

	puffer[pufferIndex].isFixed = false
	return nil
}

// Update the page pageNo as modified. If the page ist later removed form the
// system buffer, it has to be written to mass storage.
// If unsuccessful, return an error value describing the error.
func Update(pageNo int) error {
	var pufferIndex = pageMap[pageNo]

	if !pageAvailInPuffer(pageNo) {
		return errors.New("PageNotPuffered")
	}

	puffer[pufferIndex].isUpdated = true
	return nil
}

// Write this page to mass storage.
// The page address stays valid.
// If unsuccessful, return an error value describing the error.
func Write(pageNo int) error {
	var pufferIndex = pageMap[pageNo]

	if !pageAvailInPuffer(pageNo) {
		return errors.New("PageNotPuffered")

	} else if err := initFileSystem(pageNo); err != nil {
		return err
	}

	for blockIndex := 0; blockIndex < BlocksPerPage; blockIndex++ {
		var block fileinterface.Block

		for byteIndex := 0; byteIndex < fileinterface.Blocksize; byteIndex++ {
			block[byteIndex] = (*puffer[pufferIndex].page)[byteIndex]
		}

		if err := fileinterface.Write(fileID, pageNo*BlocksPerPage+blockIndex, &block); err != nil {
			return err
		}
	}

	return nil
}

// ResetCounters will store zero in both the
// hit and miss counter
func ResetCounters() {
	CacheHitCounter = 0
	CacheMissCounter = 0
}

// ---------------------- helper functions ----------------------

// Check if a page is stored in the puffer at the moment
// and that the pageData are not nil
func pageAvailInPuffer(pageNo int) bool {
	var pufferIndex = pageMap[pageNo]

	if pufferIndex == 0 {
		return false

	} else if puffer[pufferIndex].page == nil {
		return false
	}

	return true
}

// Prepare the database file to be ready to read or write.
// If the pageNo is higher than the maximum pageNo of the file,
// the length of the file will be increased
func initFileSystem(pageNo int) error {
	log.SetOutput(os.Stdout)
	var err error

	if fileID == -1 {

		if _, err := os.Stat(fileName); os.IsNotExist(err) {
			//database file not existing
			fileinterface.Create(fileName)
			log.Printf("New file created: %s", fileName)
		}

		fileID, err = fileinterface.Open(fileName)

		if err != nil {
			return err
		}

		log.Printf("Open file: %s with ID: %d", fileName, fileID)
	}

	var blockLength = -1
	if blockLength, err = fileinterface.Length(fileID); err != nil {
		return err
	}

	//Divide by the number of blocks per page to get the pageLength
	var pageLength = blockLength / BlocksPerPage

	//Extend the file when the pageNo can not be stored in the file
	if pageNo > pageLength {
		log.Printf("Extending file from pageLength: %d to %d", pageLength, pageNo)

		var emptyBlock fileinterface.Block

		var targetBlockLength = pageNo * BlocksPerPage

		for blockNo := blockLength; blockNo < targetBlockLength; blockNo++ {
			if err := fileinterface.Write(fileID, blockNo, &emptyBlock); err != nil {
				return err
			}
		}
	}

	return nil
}
