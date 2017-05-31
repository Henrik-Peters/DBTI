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
		Update(2)
	}

	//Request page 3
	if page, err := Request(3); err != nil {
		t.Error(err)
	} else {

		if page == nil {
			t.Error("Page 3 is nil")
		}
	}

	//Request page 2 and see if we still out saved value
	if page, err := Request(2); err != nil {
		t.Error(err)
	} else {

		if page == nil {
			t.Error("Page 2 is nil")
		}

		if page[0] != 0xE5 {
			t.Error("Page 2 does not contain saved data")
		}
	}

	WriteCacheStats("TestRequest")
}

func TestFix(t *testing.T) {
}

func TestUnfix(t *testing.T) {
}

func TestUpdate(t *testing.T) {
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
