package fileinterface

import (
	"testing"
)

func TestRequest(t *testing.T) {
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
	}

	//Request page 3
	if page, err := Request(3); err != nil {
		t.Error(err)
	} else {

		if page == nil {
			t.Error("Page 3 is nil")
		}
	}
}

func TestFix(t *testing.T) {
}

func TestUnfix(t *testing.T) {
}

func TestUpdate(t *testing.T) {
}

func TestWrite(t *testing.T) {
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

		//write the page to disk
		if err := Write(1); err != nil {
			t.Error(err)
		}
	}
}
