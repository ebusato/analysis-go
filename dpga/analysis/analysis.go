// Package analysis is used to perform the dpga data analysis.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/gonum/plot/vg"

	"gitlab.in2p3.fr/avirm/analysis-go/applyCorrCalib"
	"gitlab.in2p3.fr/avirm/analysis-go/dpga/dpgadetector"
	"gitlab.in2p3.fr/avirm/analysis-go/dpga/dq"
	"gitlab.in2p3.fr/avirm/analysis-go/dpga/rw"
	"gitlab.in2p3.fr/avirm/analysis-go/pulse"
)

func main() {
	log.SetFlags(log.Llongfile | log.LstdFlags)

	var (
		infileName = flag.String("i", "testdata/tenevents_hex.txt", "Name of the input file.")
		noEvents   = flag.Uint("n", 10000000, "Number of events to process.")
		pedCorr    = flag.String("pedcorr", "", "Name of the csv file containing pedestal constants. If not set, pedestal corrections are not applied.")
		wGob       = flag.String("wgob", "dqplots.gob", "Name of the output gob file containing dq plots. If not set, the gob file is not produced.")
		refGob     = flag.String("refgob", "", "Name of the gob file containing reference dq plots. If not set, reference dq plots are not overlaid to current dq plots.")
	)

	flag.Parse()

	err := os.RemoveAll("output")
	if err != nil {
		log.Fatalf("error removing output directory", err)
	}

	err = os.Mkdir("output", 0777)
	if err != nil {
		log.Fatalf("error creating output directory", err)
	}

	// Reader
	filer, err := os.Open(*infileName)
	if err != nil {
		log.Fatalf("error opening file %v", err)
	}
	defer filer.Close()

	r, err := rw.NewReader(bufio.NewReader(filer), rw.HeaderCAL)
	if err != nil {
		log.Fatalf("could not open asm file: %v\n", err)
	}

	// Start doing concrete analysis

	if *pedCorr != "" {
		dpgadetector.Det.ReadPedestalsFile(*pedCorr)
	}
	dqplots := dq.NewDQPlot()

	for event, status := r.ReadNextEvent(); status && event.ID < *noEvents; event, status = r.ReadNextEvent() {
		if event.ID%50 == 0 {
			fmt.Printf("Processing event %v\n", event.ID)
		}
		///////////////////////////////////////////////////////////
		// Corrections
		if *pedCorr != "" {
			event = applyCorrCalib.RemovePedestal(event)
		}
		///////////////////////////////////////////////////////////

		///////////////////////////////////////////////////////////
		// Plotting
		// pulses
		if event.ID < 40 {
			event.PlotPulses(pulse.XaxisCapacitor, false)
		}
		// dq
		dqplots.FillHistos(event)
		////////////////////////////////////////////////////////////

		//event.Print(true)
	}

	dqplots.Finalize()

	tp := dqplots.MakeChargeTiledPlot(dq.Charge)
	tp.Save(50*vg.Centimeter, 50*vg.Centimeter, "ChargeTiled.png")

	dqplots.WriteGob(*wGob)
	dqplots.SaveHistos(*refGob)

}
