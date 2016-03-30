package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"gitlab.in2p3.fr/avirm/analysis-go/dpga/rw"
)

func main() {
	log.SetFlags(log.Llongfile | log.LstdFlags)

	var (
		infileName = flag.String("i", "", "Name of the input file")
	)

	flag.Parse()

	f, err := os.Open(*infileName)
	if err != nil {
		log.Fatalf("could not open data file: %v\n", err)
	}
	defer f.Close()

	r, err := rw.NewReader(bufio.NewReader(f), rw.HeaderCAL, false)
	if err != nil {
		log.Fatalf("could not open asm file: %v\n", err)
	}

	r.Header()

	nFrames := uint(0)
	for {
		frame, err := r.Frame()
		if err != nil {
			if err != io.EOF {
				log.Fatalf("error loading frame: %v\n", err)
			}
			if frame.ID != rw.LastFrame() {
				log.Fatalf("invalid last frame id. got=%d. want=%d", frame.ID, rw.LastFrame())
			}
			break
		}
		nFrames++
	}
	fmt.Printf("Informations for file %v\n", *infileName)
	fmt.Printf("  -> number of frames = %v", nFrames)
	if nFrames%120 != 0 {
		fmt.Printf(" (WARNING: not a multiple of 120)")
	}
	fmt.Printf("\n")
	fmt.Printf("  -> number of events = %v\n", nFrames/120)
}
