package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"os"
	"time"

	"gitlab.in2p3.fr/avirm/analysis-go/dpga/rw"
)

func main() {
	log.SetFlags(log.Llongfile | log.LstdFlags)

	var (
		hdrType  = rw.HeaderCAL
		fileName = flag.String("i", "", "Input file name")
		ip       = flag.String("ip", "localhost", "IP address")
		port     = flag.String("p", "5556", "Port number")
		freq     = flag.Uint("freq", 100, "Event number printing frequency")
		firstEvt = flag.Uint("firstEvt", 0, "First event")
		sleep    = flag.Bool("s", false, "If set, sleep a bit between events")
	)
	flag.Var(&hdrType, "h", "Type of header: HeaderCAL or HeaderOld")
	flag.Parse()

	// Reader
	filew, err := os.Open(*fileName)
	if err != nil {
		log.Fatalf("could not create data file: %v\n", err)
	}
	defer filew.Close()
	r, err := rw.NewReader(bufio.NewReader(filew), hdrType)
	if err != nil {
		log.Fatalf("could not open stream: %v\n", err)
	}

	// Writer
	ln, err := net.Listen("tcp", *ip+":"+*port)
	if err != nil {
		log.Fatal(err)
	}
	conn, err := ln.Accept()

	w := rw.NewWriter(bufio.NewWriter(conn))
	if err != nil {
		log.Fatalf("could not open file: %v\n", err)
	}
	defer w.Close()

	// Start writing stream to TCP
	hdr := r.Header()
	hdr.Print()

	err = w.Header(hdr, false)
	if err != nil {
		log.Fatalf("error writing header: %v\n", err)
	}

	nFrames := uint(0)
	evtIDprev := float64(-1)
	for {
		if *sleep {
			fmt.Println("Warning: sleeping for 1 second")
			time.Sleep(1 * time.Second)
		}

		frame, err := r.Frame()
		evtID := float64(frame.Block.Evt)
		if evtID != evtIDprev {
			if math.Mod(evtID, float64(*freq)) == 0 {
				fmt.Printf("event %v\n", evtID)
			}
			if evtID < float64(*firstEvt) {
				continue
			}
		}
		evtIDprev = evtID
		if err != nil {
			log.Println("error is not nil", err)
			if err != io.EOF {
				log.Fatalf("error loading frame: %v\n", err)
			}
			if frame.ID != rw.LastFrame() {
				log.Fatalf("invalid last frame id. got=%d. want=%d", frame.ID, rw.LastFrame())
			}
			break
		}
		if frame.FirstOfEvent {
			//fmt.Println("FirstOfEvent=", frame.FirstOfEvent)
			w.WriteU32(rw.FirstEventWord)
			for _, v := range r.Counters {
				w.WriteU32(v)
			}
		}
		err = w.Frame(frame)
		if err != nil {
			log.Fatalf("error writing frame: %v\n", err)
		}
		nFrames++
	}

}
