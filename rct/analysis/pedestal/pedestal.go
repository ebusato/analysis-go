package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"gitlab.in2p3.fr/avirm/analysis-go/rct/rw"
	"gitlab.in2p3.fr/avirm/analysis-go/rct/trees"
)

var (
	inFile       = flag.String("i", "", "Name of input file")
	noEvents     = flag.Uint("n", 100000, "Number of events")
	evtFreq      = flag.Uint("ef", 100, "Event printing frequency")
	treeFileName = flag.String("ot", "", "Name of the TFile containing the output tree")
)

func main() {
	flag.Parse()

	// Reader
	f, err := os.Open(*inFile)
	if err != nil {
		log.Fatalf("could not open data file: %v\n", err)
	}
	defer f.Close()

	/*r, err := rw.NewReader(bufio.NewReader(f))
	if err != nil {
		log.Fatalf("could not open stream: %v\n", err)
	}*/
	var r *rw.Reader
	r, _ = rw.NewReader(bufio.NewReader(f))
	r.SetSigThreshold(1)

	outRootFileName := strings.Replace(*inFile, ".bin", ".root", 1)
	var tree *trees.Tree
	if *treeFileName != "" {
		tree = trees.NewTree(*treeFileName)
	} else {
		tree = trees.NewTree(outRootFileName)
	}

	var iEvent uint

	// Read and do nothing with first event
	// because it is corrupted for an unknown reason
	//  -> remove once fixed
	r.ReadNextEvent()

	for iEvent < *noEvents {
		if iEvent%*evtFreq == 0 {
			fmt.Printf("event %v\n", iEvent)
		}
		event, err := r.ReadNextEvent()
		if err != nil {
			panic(err)
		}
		//event.PlotPulses(pulse.XaxisCapacitor, false, pulse.YRangeAuto, pulse.XRangeFull)

		if tree != nil {
			tree.Fill(0, event)
		}
		iEvent++
	}
	if tree != nil {
		tree.Close()
	}
}
