package rw

import (
	"bufio"
	"fmt"
	"os"
	"testing"
)

//var rhdr *Header
//var revents []event.Event

func TestRW(t *testing.T) {
	fmt.Println("starting TestRW")

	inFileName := "/home/ebusato/Travail/Imaging/DPGA/Soft/FirmwareTests/ServeurUdp/datas/MyFile_eno1@0_0.bin"
	// 	outFileName := strings.Replace(inFileName, ".bin", "_out.bin", 1)

	// Reader
	f, err := os.Open(inFileName)
	if err != nil {
		t.Fatalf("could not open data file: %v\n", err)
	}
	defer f.Close()

	r, err := NewReader(bufio.NewReader(f))
	if err != nil {
		t.Fatalf("could not open asm file: %v\n", err)
	}
	r.ReadMode = Default
	r.FileHeader.Print()

	// Writer
	// 	filew, err := os.Create(outFileName)
	// 	if err != nil {
	// 		log.Fatalf("could not create data file: %v\n", err)
	// 	}
	// 	defer filew.Close()
	//
	// 	w := NewWriter(bufio.NewWriter(filew))
	// 	if err != nil {
	// 		log.Fatalf("could not open file: %v\n", err)
	// 	}
	// 	defer w.Close()

	for {
		frame, _ := r.Frame()
		frame.Header.Print()
		frame.Data.Print()
	}

	/*
			// Writer
			filew, err := os.Create("testdata/w50evtsNewHeader.bin")
			if err != nil {
				log.Fatalf("could not create data file: %v\n", err)
			}
			defer filew.Close()

			w := NewWriter(bufio.NewWriter(filew))
			if err != nil {
				log.Fatalf("could not open file: %v\n", err)
			}
			defer w.Close()


		rhdr = r.Header()


			err = w.Header(rhdr, false)
			if err != nil {
				t.Fatalf("error writing header: %v\n", err)
			}
	*/
}

/*
func TestWIntegrity(t *testing.T) {
	fmt.Println("starting TestWIntegrity")
	f, err := os.Open("testdata/w50evtsNewHeader.bin")
	if err != nil {
		t.Fatalf("could not open data file: %v\n", err)
	}
	defer f.Close()

	r, err := NewReader(bufio.NewReader(f), HeaderCAL)
	if err != nil {
		t.Fatalf("could not open asm file: %v\n", err)
	}

	whdr := r.Header()

	var wevents []event.Event

	for {
		event, status := r.ReadNextEvent()
		wevents = append(wevents, *event)
		if r.Err() != io.EOF {
			if status == false {
				t.Fatalf("error: status is false\n")
			}
		} else {
			break
		}
	}

	fmt.Println("in TestWIntegrity, starting deepEqual")
	if !reflect.DeepEqual(rhdr, whdr) {
		fmt.Println("Printing original header")
		rhdr.Print()
		fmt.Println("Printing written header")
		whdr.Print()
		t.Fatalf("headers differ.")
	}

	if !reflect.DeepEqual(revents, wevents) {
		t.Fatalf("events differ.\ngot= %#v\nwant=%v\n", wevents, revents)
	}
}
*/
