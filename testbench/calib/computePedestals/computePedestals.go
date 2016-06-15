package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gitlab.in2p3.fr/avirm/analysis-go/event"
	"gitlab.in2p3.fr/avirm/analysis-go/testbench/reader"
	"gitlab.in2p3.fr/avirm/analysis-go/testbench/rw"
	"gitlab.in2p3.fr/avirm/analysis-go/testbench/tbdetector"
)

type ReadNextEventer interface {
	ReadNextEvent() (*event.Event, bool)
}

func main() {
	log.SetFlags(log.Llongfile | log.LstdFlags)

	var (
		infileName      = flag.String("i", "testdata/tenevents_hex.txt", "Name of the input file")
		outfileName     = flag.String("o", "output/pedestals.csv", "Name of the output pedestal csv file")
		outrootfileName = flag.String("oroot", "output/pedestals.root", "Name of the output pedestal root file")
		//outFileNamePulses = flag.String("oP", "output/pulses.csv", "Name of the output file containing pulse data")
		//outFileNameGlobal = flag.String("oG", "output/globalEventVariables.csv", "Name of the output file containing global event variables")
		noEvents  = flag.Int("n", -1, "Number of events to process (-1 means all events are processed)")
		inputType = reader.HexInput
		hdrType   = rw.HeaderCAL
	)
	flag.Var(&inputType, "inType", "Type of input file (possible values: Dec,Hex,Bin)")
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

	var rner ReadNextEventer

	if filepath.Ext(*infileName) == ".bin" {
		// Binary file reader
		rner, err = rw.NewReader(bufio.NewReader(file), hdrType)
		if err != nil {
			log.Fatalf("could not open asm file: %v\n", err)
		}
	} else if filepath.Ext(*infileName) == ".txt" {
		// ASCII file scanner
		s := reader.NewScanner(bufio.NewScanner(file))
		s.SetInputType(inputType)
		rner = s
	} else {
		log.Fatalf("file extension not recognized.")
	}

	for event, status := rner.ReadNextEvent(); status && (*noEvents == -1 || int(event.ID) < *noEvents); event, status = rner.ReadNextEvent() {
		if event.ID%500 == 0 {
			fmt.Printf("Processing event %v\n", event.ID)
		}
		event.PushPedestalSamples()
	}

	tbdetector.Det.FinalizePedestalsMeanErr()
	tbdetector.Det.WritePedestalsToFile(*outfileName, *infileName, *outrootfileName)

	tbdetector.Det.PlotPedestals(true)
	tbdetector.Det.PlotPedestals(false)
	// detector.TBDet.Print()
}
