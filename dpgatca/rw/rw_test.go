package rw

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"testing"

	"gitlab.in2p3.fr/avirm/analysis-go/event"
)

//var rhdr *Header
//var revents []event.Event

var (
	evtChan = make(chan *event.Event)
)

func TestRW(t *testing.T) {
	fmt.Println("starting TestRW")

	inFileName := "../data/run0011/MyFile1e.bin"
	outFileName := strings.Replace(inFileName, "MyFile1e", "MyFile1e_out", 1)

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
	// Writer
	filew, err := os.Create(outFileName)
	if err != nil {
		log.Fatalf("could not create data file: %v\n", err)
	}
	defer filew.Close()

	w := NewWriter(bufio.NewWriter(filew))
	if err != nil {
		log.Fatalf("could not open file: %v\n", err)
	}
	defer w.Close()

	const N = 1
	var wg sync.WaitGroup
	wg.Add(N)

	go r.ReadFrames(evtChan, w, &wg)
	for {
		<-evtChan
	}

	wg.Wait()

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
