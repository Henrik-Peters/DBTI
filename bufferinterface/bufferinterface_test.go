package fileinterface

import (
	"fmt"
	"testing"
)

func TestRequest(t *testing.T) {
	ResetCounters()

	//Request page 1
	if page, err := Request(1); err != nil {
		t.Error(err)
	} else {

		if page == nil {
			t.Error("Page 1 is nil")
		}
	}

	//Request page 2
	if page, err := Request(2); err != nil {
		t.Error(err)
	} else {

		if page == nil {
			t.Error("Page 2 is nil")
		}

		page[0] = 0xE5
		page[5] = 0xAB
		Update(2)
	}

	//Request page 3
	if page, err := Request(3); err != nil {
		t.Error(err)
	} else {

		if page == nil {
			t.Error("Page 3 is nil")
		}

		page[100] = 0xFF
		Update(3)
	}

	//Request page 2 and see if we still get the saved value
	if page, err := Request(2); err != nil {
		t.Error(err)
	} else {

		if page == nil {
			t.Error("Page 2 is nil")
		}

		if page[0] != 0xE5 || page[5] != 0xAB {
			t.Error("Page 2 does not contain saved data")
		}
	}

	//Request page 3 and see if we still get the saved value
	if page, err := Request(3); err != nil {
		t.Error(err)
	} else {

		if page == nil {
			t.Error("Page 3 is nil")
		}

		if page[100] != 0xFF {
			t.Error("Page 3 does not contain saved data")
		}
	}

	//Request page 4
	if page, err := Request(4); err != nil {
		t.Error(err)
	} else {

		if page == nil {
			t.Error("Page 4 is nil")
		}
	}

	//Request page 5
	if page, err := Request(5); err != nil {
		t.Error(err)
	} else {

		if page == nil {
			t.Error("Page 5 is nil")
		}
	}

	WriteCacheStats("TestRequest")
}

func TestFix(t *testing.T) {
	ResetCounters()
	
	//Fill the puffer with pages
	for i := 0; i < PufferSize; i++ {
		if _, err := Request(i); err != nil {
			t.Error(err)
		}
	}
	
	//Fix some pages
	Fix(5)
	Fix(7)
	
	if !puffer[pageMap[5]].isFixed {
		t.Error("Page 5 is not fixed")
	}
	
	if !puffer[pageMap[7]].isFixed {
		t.Error("Page 7 is not fixed")
	}
	
	//Fill the puffer with new pages
	for i := PufferSize; i < PufferSize * 2; i++ {
		if _, err := Request(i); err != nil {
			t.Error(err)
		}
	}
	
	//Test if the fixed pages are still puffered
	if !pageAvailInPuffer(5) {
		t.Error("Page 5 is fixed but not in puffer")
	}
	
	if !pageAvailInPuffer(7) {
		t.Error("Page 5 is fixed but not in puffer")
	}
	
	if !puffer[pageMap[5]].isFixed {
		t.Error("Page 5 is not fixed")
	}
	
	if !puffer[pageMap[7]].isFixed {
		t.Error("Page 7 is not fixed")
	}
	
	WriteCacheStats("TestFix")
}

func TestUnfix(t *testing.T) {
	ResetCounters()
	
	if page, err := Request(1); err != nil {
		t.Error(err)
	} else {

		if page == nil {
			t.Error("Page 1 is nil")
		}
	}
	
	Fix(1)
	
	if !puffer[pageMap[1]].isFixed {
		t.Error("Page 1 is not fixed")
	}
	
	UnFix(1)
	
	if puffer[pageMap[1]].isFixed {
		t.Error("Page 1 should be unfixed")
	}
	
	WriteCacheStats("TestUnfix")
}

func TestUpdate(t *testing.T) {
	ResetCounters()
	
	//Fill the puffer with pages
	for i := 0; i < PufferSize; i++ {
		if _, err := Request(i); err != nil {
			t.Error(err)
		}
	}
	
	//Request page 8 and write some data
	if page, err := Request(8); err != nil {
		t.Error(err)
	} else {

		if page == nil {
			t.Error("Page 8 is nil")
		}

		page[0] = 0x12
		page[1] = 0x34
		page[2] = 0x56
		page[3] = 0xA8
		page[4] = 0x9E
		
		//Mark this page as updated
		Update(8)
		
		//Check the update-flag
		if !puffer[pageMap[8]].isUpdated {
			t.Error("Page 8 should have the update-flag set")
		}
	}
	
	//Fill the puffer with new pages
	//Page 8 should be written to disk on a cache displacement
	for i := PufferSize; i < PufferSize * 2; i++ {
		if _, err := Request(i); err != nil {
			t.Error(err)
		}
	}
	
	//Request page 8 and see if we get the correct data
	if page, err := Request(8); err != nil {
		t.Error(err)
	} else {

		if page == nil {
			t.Error("Page 8 is nil")
		}

		if page[0] != 0x12 || page[1] != 0x34 || page[2] != 0x56 || page[3] != 0xA8 || page[4] != 0x9E {
			t.Error("Page 8 does not contain saved data")
		}
	}
	
	WriteCacheStats("TestUpdate")
}

func TestWrite(t *testing.T) {
	ResetCounters()

	//request page 1
	if page, err := Request(1); err != nil {
		t.Error(err)
	} else {
		if page == nil {
			t.Error("Page 1 is nil")
		}

		//insert some test data
		page[0] = 0xAB
		page[1] = 0xCF
		page[2] = 0xFF
		page[4094] = 0x4C
		page[4095] = 0x7F
		Update(1)

		//write the page to disk
		if err := Write(1); err != nil {
			t.Error(err)
		}
	}

	WriteCacheStats("TestWrite")
}

func WriteCacheStats(title string) {
	fmt.Printf("\n%s - Cache stats\n", title)
	fmt.Println("-----------------------------------")
	fmt.Printf("Cache hits:   %d\n", CacheHitCounter)
	fmt.Printf("Cache misses: %d\n", CacheMissCounter)
	fmt.Printf("-----------------------------------\n")
}
