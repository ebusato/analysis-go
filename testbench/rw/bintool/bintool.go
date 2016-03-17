package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"gitlab.in2p3.fr/avirm/analysis-go/testbench/rw"
)

var nFramesPerEvent uint = 12

func main() {
	log.SetFlags(log.Llongfile | log.LstdFlags)

	var (
		infileName = flag.String("i", "", "Name of the input file")

		// set of flags when dumping to ascii
		ascii     = flag.Bool("ascii", false, "Dump binary to ascii if set to true")
		noFrames  = flag.Uint("n", 10000000000000000000, "Number of frames")
		verbosity = flag.String("v", "full", "Verbosity of the output (short, medium, long, full)")

		// set of flags when printing general informations
		info = flag.Bool("info", false, "Print general informations about binary file when set tot true")

		// set of flags when writting new binary file with only good events
		wgoodevts   = flag.Bool("wgoodevts", false, "Write good events to new binary file (skip events having corrupted frames)")
		outfileName = flag.String("o", "", "Name of the output file")
		nevts       = flag.Int("nevts", -1, "Number of events to write")
	)

	flag.Parse()

	// Reader
	f, err := os.Open(*infileName)
	if err != nil {
		log.Fatalf("could not open data file: %v\n", err)
	}
	defer f.Close()

	r, err := rw.NewReader(bufio.NewReader(f))
	if err != nil {
		log.Fatalf("could not open asm file: %v\n", err)
	}

	// Check flags
	if !*ascii && !*info && !*wgoodevts {
		fmt.Println("you need to set some flags, otherwise you'll get nothing out of bintool.")
		return
	}

	if *ascii && *wgoodevts {
		fmt.Println("ascii and wgoodevts flags can't be set to true at the same time.")
		return
	}

	if *wgoodevts && *outfileName == "" {
		fmt.Println("you need to set the output file name (-o) when wgoodevts is set to true.")
		return
	}

	if *wgoodevts && *nevts < 0 && *nevts != -1 {
		fmt.Println("you've set nevts to be < 0 and != -1. This case is not supported, please fix.")
		return
	}

	// Writer, used if wgoodevts set to true
	var filew *os.File
	var w *rw.Writer
	eventFrames := make([]*rw.Frame, nFramesPerEvent)
	if *wgoodevts {
		filew, err = os.Create(*outfileName)
		if err != nil {
			log.Fatalf("could not create data file: %v\n", err)
		}
		defer filew.Close()
		w = rw.NewWriter(bufio.NewWriter(filew))
		if err != nil {
			log.Fatalf("could not open file: %v\n", err)
		}
		defer w.Close()
	}

	//r.Debug = true

	// Start processing input file
	hdr := r.Header()
	switch {
	case *ascii:
		hdr.Print()
	case *wgoodevts:
		err = w.Header(hdr)
		if err != nil {
			log.Fatalf("error writing header: %v\n", err)
		}
	}

	nFrames := uint(0)
	for {
		frame, err := r.Frame()
		if err != nil {
			if err != io.EOF {
				fmt.Printf("error loading frame: %v\n", err)
			}
			if frame.ID != rw.LastFrame() {
				fmt.Printf("invalid last frame id. got=%d. want=%d (nFrames=%v)", frame.ID, rw.LastFrame(), nFrames)
			}
			break
		}
		switch {
		case *ascii:
			switch nFrames < *noFrames {
			case true:
				frame.Print(*verbosity)
			case false:
				goto A
			}
		case *wgoodevts:
			iEvent := nFrames / nFramesPerEvent
			if *nevts == -1 || iEvent < uint(*nevts) {
				if iEvent%500 == 0 && nFrames%nFramesPerEvent == 0 {
					fmt.Println("writing event", iEvent)
				}
				eventFrames[nFrames%nFramesPerEvent] = frame
				if nFrames%nFramesPerEvent == nFramesPerEvent-1 {
					writeFrames(w, eventFrames)
				}
			}
		}
		nFrames++
	}
A:
	if *info {
		fmt.Printf("\nInformations for file %v\n", *infileName)
		fmt.Printf("  -> number of frames = %v", nFrames)
		if nFrames%nFramesPerEvent != 0 {
			fmt.Printf(" (WARNING: not a multiple of %v)", nFramesPerEvent)
		}
		fmt.Printf("\n")
		fmt.Printf("  -> number of events = %v\n", float64(nFrames)/float64(nFramesPerEvent))
	}
}

func writeFrames(w *rw.Writer, frames []*rw.Frame) {
	for i := range frames {
		err := w.Frame(*frames[i])
		if err != nil {
			log.Fatalf("error writing frame: %v\n", err)
		}
		frames[i] = nil
	}
}
