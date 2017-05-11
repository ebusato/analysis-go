// Package analysis is used to perform the dpga data analysis.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"gitlab.in2p3.fr/avirm/analysis-go/applyCorrCalib"
	"gitlab.in2p3.fr/avirm/analysis-go/dpga/calib/selectCalib"
	"gitlab.in2p3.fr/avirm/analysis-go/dpga/dq"
	"gitlab.in2p3.fr/avirm/analysis-go/dpga/rw"
	"gitlab.in2p3.fr/avirm/analysis-go/dpga/trees"
	"gitlab.in2p3.fr/avirm/analysis-go/pulse"
)

func main() {
	log.SetFlags(log.Llongfile | log.LstdFlags)

	var (
		infileName = flag.String("i", "", "Name of the input file.")
		noEvents   = flag.Uint("n", 10000000, "Number of events to process.")
		evtStart   = flag.Uint("evtstart", 0, "First event to process.")
		calib      = flag.String("calib", "", "String indicating which calib to use (e.g. A1 for period A, version 1)")
		noped      = flag.Bool("noped", false, "If specified, no pedestal correction applied")
		notdo      = flag.Bool("notdo", false, "If specified, no time dependent offset correction applied")
		noen       = flag.Bool("noen", false, "If specified, no energy calibration applied")
		dotree     = flag.Bool("dotree", false, "If specified, tree with all pulses is written")
		dotree2    = flag.Bool("dotree2", false, "If specified, treeMult2 is written")
		dotreeLOR  = flag.Bool("dotreeLOR", false, "If specified, treeLOR is written")
		//wGob       = flag.String("wgob", "dqplots.gob", "Name of the output gob file containing dq plots. If not set, the gob file is not produced.")
		rfcutmean  = flag.Float64("rfcutmean", 7, "Mean used to apply RF selection cut.")
		rfcutwidth = flag.Float64("rfcutwidth", 5, "Width used to apply RF selection cut.")
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
	doPedestal := false
	doTimeDepOffset := false
	doEnergyCalib := false
	if *calib != "" {
		selectCalib.Which(*calib)
		if !*noped {
			doPedestal = true
		}
		if !*notdo {
			doTimeDepOffset = true
		}
		if !*noen {
			doEnergyCalib = true
		}
	}
	dqplots := dq.NewDQPlot()

	outrootfileName := strings.Replace(*infileName, ".bin", ".root", 1)
	var tree *trees.Tree = nil
	if *dotree {
		tree = trees.NewTree(outrootfileName)
	}

	outrootfileNameMult2 := strings.Replace(*infileName, ".bin", "Mult2.root", 1)
	var treeMult2 *trees.TreeMult2 = nil
	if *dotree2 {
		treeMult2 = trees.NewTreeMult2(outrootfileNameMult2)
	}

	outrootfileNameLOR := strings.Replace(*infileName, ".bin", "LOR.root", 1)
	var treeLOR *trees.TreeLOR = nil
	if *dotreeLOR {
		treeLOR = trees.NewTreeLOR(outrootfileNameLOR)
	}

	hdr := r.Header()

	for event, status := r.ReadNextEvent(); status && event.ID < *noEvents; event, status = r.ReadNextEvent() {
		if event.ID%500 == 0 {
			fmt.Printf("Processing event %v\n", event.ID)
		}
		if event.ID < *evtStart {
			continue
		}
		// 		if event.ID < 86000 {
		// 			continue
		// 		}
		///////////////////////////////////////////////////////////
		// Corrections
		event = applyCorrCalib.CorrectEvent(event, doPedestal, doTimeDepOffset, doEnergyCalib)
		///////////////////////////////////////////////////////////

		///////////////////////////////////////////////////////////
		// Plotting
		// pulses
		// 		if event.ID < 20 {
		// 			event.PlotPulses(pulse.XaxisCapacitor, false, pulse.YRangePedestal, true)
		// 		}
		// dq
		dqplots.FillHistos(event, *rfcutmean, *rfcutwidth)
		////////////////////////////////////////////////////////////

		///////////////////////////////////////////////////////////
		// ROOT Tree making
		if tree != nil {
			tree.Fill(hdr.RunNumber, r.Header(), event)
		}
		if treeLOR != nil {
			timesRF := event.FindTimesRF()
			treeLOR.Fill(hdr.RunNumber, r.Header(), event, timesRF)
		}

		//fmt.Println(len(pulses511keV))

		if treeMult2 != nil {
			pulses511keV := event.PulsesInEnergyWindow(511, 3, 28.3)
			if len(pulses511keV) == 2 && !pulse.SameHemi(pulses511keV[0], pulses511keV[1]) {
				treeMult2.Fill(hdr.RunNumber, r.Header(), event, pulses511keV[0], pulses511keV[1])
			}
		}
		////////////////////////////////////////////////////////////

		//event.Print(true)
	}
	if tree != nil {
		tree.Close()
	}
	if treeLOR != nil {
		treeLOR.Close()
	}
	if treeMult2 != nil {
		treeMult2.Close()
	}
}
