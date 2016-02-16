package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"os"

	"gitlab.in2p3.fr/avirm/analysis-go/dpga/rw"
)

func main() {
	log.SetFlags(log.Llongfile | log.LstdFlags)

	var (
		infileName = flag.String("i", "../rw/testdata/data-03-frames.bin", "Name of the input file")
		noFrames   = flag.Uint("n", 10000000, "Number of frames")
		verbosity  = flag.String("v", "full", "Verbosity of the output (short, medium, long, full)")
	)

	flag.Parse()

	f, err := os.Open(*infileName)
	if err != nil {
		log.Fatalf("could not open data file: %v\n", err)
	}
	defer f.Close()

	r, err := rw.NewReader(bufio.NewReader(f))
	if err != nil {
		log.Fatalf("could not open asm file: %v\n", err)
	}

	hdr := r.Header()
	hdr.Print()

	nFrames := uint(0)
	for nFrames < *noFrames {
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
		frame.Print(*verbosity)
		nFrames++
	}
}
