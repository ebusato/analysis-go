package rw

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"testing"

	"gitlab.in2p3.fr/avirm/analysis-go/event"
)

var rhdr Header
var revents []event.Event

func TestRW(t *testing.T) {
	fmt.Println("starting TestRW")

	// Reader
	f, err := os.Open("testdata/50evtsNewHeader.bin")
	if err != nil {
		t.Fatalf("could not open data file: %v\n", err)
	}
	defer f.Close()

	r, err := NewReader(bufio.NewReader(f), HeaderCAL)
	if err != nil {
		t.Fatalf("could not open asm file: %v\n", err)
	}

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

	err = w.Header(&rhdr, false)
	if err != nil {
		t.Fatalf("error writing header: %v\n", err)
	}

	nevents := 0
	for {
		event, status := r.ReadNextEvent()
		if r.Err() != io.EOF {
			revents = append(revents, *event)
			w.Event(event)
			if status == false {
				t.Fatalf("error: status is false\n")
			}
		} else {
			break
		}
		nevents++
	}
	if nevents != 50 {
		t.Fatalf("got %d events. want 50\n", nevents)
	}
}

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
		if r.Err() != io.EOF {
			wevents = append(wevents, *event)
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
