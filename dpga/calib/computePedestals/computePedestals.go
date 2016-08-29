// Package computePedestal computes pedestals.
// It should be run before applyCorrCalib is used.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	"gitlab.in2p3.fr/avirm/analysis-go/dpga/dpgadetector"
	"gitlab.in2p3.fr/avirm/analysis-go/dpga/rw"
)

func main() {
	log.SetFlags(log.Llongfile | log.LstdFlags)

	var (
		infileName  = flag.String("i", "testdata/tenevents_hex.txt", "Name of the input file")
		outfileName = flag.String("o", "output/pedestals.csv", "Name of the output file")
		noEvents    = flag.Uint("n", 10000000, "Number of events to process")
		hdrType     = rw.HeaderCAL
	)

	flag.Var(&hdrType, "h", "Type of header: HeaderCAL or HeaderOld")
	flag.Parse()

	err := os.RemoveAll("output")
	if err != nil {
		log.Fatalf("error removing output directory", err)
	}

	err = os.Mkdir("output", 0777)
	if err != nil {
		log.Fatalf("error creating output directory", err)
	}

	file, err := os.Open(*infileName)
	if err != nil {
		log.Fatalf("error opening file %v", err)
	}
	defer file.Close()

	r, err := rw.NewReader(bufio.NewReader(file), hdrType)
	if err != nil {
		log.Fatalf("could not open asm file: %v\n", err)
	}

	for event, status := r.ReadNextEvent(); status && event.ID < *noEvents; event, status = r.ReadNextEvent() {
		if event.ID%500 == 0 {
			fmt.Printf("Processing event %v\n", event.ID)
		}
		//event.Print(true)

		event.PushPedestalSamples()
	}

	dpgadetector.Det.FinalizePedestalsMeanErr()
	dpgadetector.Det.WritePedestalsToFile(*outfileName, *infileName)

	//dpgadetector.Det.PlotPedestals(true)
	dpgadetector.Det.PlotPedestals("output", false)
}
