// Package computePedestal computes pedestals.
// It should be run before applyCorrCalib is used.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	"gitlab.in2p3.fr/avirm/analysis-go/applyCorrCalib"
	"gitlab.in2p3.fr/avirm/analysis-go/testbench/rw"
	"gitlab.in2p3.fr/avirm/analysis-go/testbench/tbdetector"
)

func main() {
	log.SetFlags(log.Llongfile | log.LstdFlags)

	var (
		infileName  = flag.String("i", "testdata/tenevents_hex.txt", "Name of the input file")
		outfileName = flag.String("o", "output/timeDepOffsets.csv", "Name of the output file")
		noEvents    = flag.Uint("n", 10000000, "Number of events to process")
		ped         = flag.String("ped", "", "Name of the csv file containing pedestal constants. If not set, pedestal corrections are not applied.")
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

	switch *ped == "" {
	case true:
		panic("pedestal correction should be applied in order to determine time dependent offset.")
	case false:
		tbdetector.Det.ReadPedestalsFile(*ped)
	}

	for event, status := r.ReadNextEvent(); status && event.ID < *noEvents; event, status = r.ReadNextEvent() {
		if event.ID%500 == 0 {
			fmt.Printf("Processing event %v\n", event.ID)
		}

		// this should be safe as pedestals calibration coefficients have been loaded previously.
		event = applyCorrCalib.HV(event, true, false)

		event.PushTimeDepOffsetSamples()
	}

	tbdetector.Det.FinalizeTimeDepOffsetsMeanErr()
	tbdetector.Det.WriteTimeDepOffsetsToFile(*outfileName, *infileName)

	tbdetector.Det.PlotTimeDepOffsets()
}
