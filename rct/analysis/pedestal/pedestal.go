package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	"gitlab.in2p3.fr/avirm/analysis-go/pulse"
	"gitlab.in2p3.fr/avirm/analysis-go/rct/rw"
)

var (
	inFile   = flag.String("i", "", "Name of input file")
	outFile  = flag.String("o", "", "Name of the output csv file")
	noEvents = flag.Uint("n", 100000, "Number of events")
	evtFreq  = flag.Uint("ef", 100, "Event printing frequency")
	vme      = flag.Bool("vme", false, "If set, uses VME reader")
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

	var iEvent uint

	for iEvent < *noEvents {
		if iEvent%*evtFreq == 0 {
			fmt.Printf("event %v\n", iEvent)
		}
		event, err := r.ReadNextEvent()
		if err != nil {
			panic(err)
		}
		event.PlotPulses(pulse.XaxisCapacitor, false, pulse.YRangePedestal, false)
		iEvent++
	}
}
