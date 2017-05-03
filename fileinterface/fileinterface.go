// Package fileinterface provides access to the underlying operating system files.
// Files are organized in blocks.
package fileinterface

import (
	"container/list"
	"errors"
	"io/ioutil"
	"os"
)

// Blocksize of a block in bytes
const Blocksize = 4096

// FID to manage the file ids
type FID int

// Block definition
type Block [Blocksize]byte

// Association from the FIDs to the file pointers
var fileMap = make(map[FID](*os.File))

// Create a new file with a given name. Return the FID and nil if succesful.
// The file is then open for reading and writing.
// If unsuccessful, return any FID and an error value describing the error.
func Create(name string) (FID, error) {
	var file, err = os.Create(name)
	fileMap[FID(file.Fd())] = file

	return FID(file.Fd()), err
}

// Delete a file with a given name. Return nil if succesful.
// If unsuccessful, return an error value describing the error.
func Delete(name string) error {
	return os.Remove(name)
}

// Open a file with a given name for reading and writing. Return nil if succesful.
// If unsuccessful, return  any FID and an error value describing the error.
// Possible errors include FileNotFoundError or FileAlreadyOpenError
func Open(name string) (FID, error) {
	var file, err = os.Open(name)
	fileMap[FID(file.Fd())] = file

	return FID(file.Fd()), err
}

// Length calculates the number of blocks available in the file given by fileNo. Return nil if succesful.
// If unsuccessful, return  any FID and an error value describing the error.
// Possible errors include FileNotOpenError
func Length(fileNo FID) (int, error) {
	var file = fileMap[fileNo]

	if file == nil {
		return 0, errors.New("FileNotOpenException")
	}

	var stats, statErr = file.Stat()

	if statErr != nil {
		return 0, statErr
	}

	return int(stats.Size() / Blocksize), nil
}

// Read the block number blockNo from the file fileNo. Counting starts at 0.
// Return a pointer to the block and nil if succesful.
// If unsuccessful, return nil and an error value describing the error.
// Possible errors include FileNotOpenError
func Read(fileNo FID, blockNo int) (*Block, error) {
	var file = fileMap[fileNo]

	if file == nil {
		return nil, errors.New("FileNotOpenException")
	}

	var blockBytes = make([]byte, Blocksize)
	var _, err = file.ReadAt(blockBytes, int64(blockNo*Blocksize))

	if err != nil {
		return nil, err
	}

	var block Block

	for i, curByte := range blockBytes {
		block[i] = curByte
	}

	return &block, nil
}

// Write the block given by the pointer block to the block number blockNo in
// the file fileNo. Counting starts at 0.
// Return nil if succesful.
// If unsuccessful, return an error value describing the error.
// Possible errors include FileNotOpen or WriteError
func Write(fileNo FID, blockNo int, block *Block) error { //  FileNotOpenException, IOException;
	var file = fileMap[fileNo]

	if file == nil {
		return errors.New("FileNotOpenException")
	}

	var blockBytes = make([]byte, Blocksize)

	for i, curBlockData := range block {
		blockBytes[i] = curBlockData
	}

	var _, err = file.WriteAt(blockBytes, int64(blockNo*Blocksize))
	return err
}

// Close the file given by fileNo. Return nil if succesful.
// If unsuccessful, return an error value describing the error.
// Possible errors include FileNotOpenError
func Close(fileNo FID) error {
	if fileMap[fileNo] == nil {
		return errors.New("InvalidFileNumber")

	} else if fileMap[fileNo].Close() != nil {
		return errors.New("FileNotOpenError")

	} else {
		return nil
	}
}

// Return a list of possible file names
func ls() []string {
	var files, _ = ioutil.ReadDir("./")
	var fileList = list.New()

	for _, curInfo := range files {
		if !curInfo.IsDir() {
			fileList.PushBack(curInfo.Name())
		}
	}

	var fileNames = make([]string, fileList.Len())

	for elemPointer, i := fileList.Front(), 0; elemPointer != nil; elemPointer, i = elemPointer.Next(), i+1 {
		fileNames[i] = elemPointer.Value.(string)
	}

	return fileNames
}
