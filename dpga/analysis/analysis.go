// Package analysis is used to perform the dpga data analysis.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gonum/plot/vg"

	"gitlab.in2p3.fr/avirm/analysis-go/applyCorrCalib"
	"gitlab.in2p3.fr/avirm/analysis-go/dpga/dpgadetector"
	"gitlab.in2p3.fr/avirm/analysis-go/dpga/dq"
	"gitlab.in2p3.fr/avirm/analysis-go/dpga/rw"
	"gitlab.in2p3.fr/avirm/analysis-go/dpga/trees"
	"gitlab.in2p3.fr/avirm/analysis-go/reconstruction"
)

func main() {
	log.SetFlags(log.Llongfile | log.LstdFlags)

	var (
		infileName = flag.String("i", "", "Name of the input file.")
		noEvents   = flag.Uint("n", 10000000, "Number of events to process.")
		ped        = flag.String("ped", "", "Name of the csv file containing pedestal constants. If not set, pedestal corrections are not applied.")
		tdo        = flag.String("tdo", "", "Name of the csv file containing time dependent offsets. If not set, time dependent offsets are not applied. Relevant only when ped!=\"\".")
		en         = flag.String("en", "", "Name of the csv file containing energy calibration constants. If not set, energy calibration is not applied.")
		wGob       = flag.String("wgob", "dqplots.gob", "Name of the output gob file containing dq plots. If not set, the gob file is not produced.")
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

	if *ped != "" {
		dpgadetector.Det.ReadPedestalsFile(*ped)
	}
	if *tdo != "" {
		dpgadetector.Det.ReadTimeDepOffsetsFile(*tdo)
	}
	if *en != "" {
		dpgadetector.Det.ReadEnergyCalibFile(*en)
	}
	dqplots := dq.NewDQPlot()

	outrootfileName := strings.Replace(*infileName, ".bin", ".root", 1)
	var treeMult2 *trees.TreeMult2 = trees.NewTreeMult2(outrootfileName)

	hdr := r.Header()

	for event, status := r.ReadNextEvent(); status && event.ID < *noEvents; event, status = r.ReadNextEvent() {
		if event.ID%500 == 0 {
			fmt.Printf("Processing event %v\n", event.ID)
		}
		///////////////////////////////////////////////////////////
		// Corrections
		doPedestal := false
		doTimeDepOffset := false
		doEnergyCalib := false
		if *ped != "" {
			doPedestal = true
		}
		if *tdo != "" {
			doTimeDepOffset = true
		}
		if *en != "" {
			doEnergyCalib = true
		}
		event = applyCorrCalib.CorrectEvent(event, doPedestal, doTimeDepOffset, doEnergyCalib)
		///////////////////////////////////////////////////////////

		///////////////////////////////////////////////////////////
		// Plotting
		// pulses
		// 		if event.ID < 20 {
		// 			event.PlotPulses(pulse.XaxisCapacitor, false, pulse.YRangePedestal, true)
		// 		}
		// dq
		dqplots.FillHistos(event)
		////////////////////////////////////////////////////////////

		///////////////////////////////////////////////////////////
		// ROOT Tree making
		mult, pulsesWithSignal := event.Multiplicity()
		if mult == 2 {
			if len(pulsesWithSignal) != 2 {
				panic("mult == 2 but len(pulsesWithSignal) != 2: this should NEVER happen !")
			}
			ch0 := pulsesWithSignal[0].Channel
			ch1 := pulsesWithSignal[1].Channel
			doMinRec := true
			if r.Header().TriggerEq == 3 {
				// In case TriggerEq = 3 (pulser), one has to check that the two pulses are
				// on different hemispheres, otherwise the minimal reconstruction is not well
				// defined
				hemi0, ok := ch0.Quartet.DRS.ASMCard.UpStr.(*dpgadetector.Hemisphere)
				if !ok {
					panic("ch0.Quartet.DRS.ASMCard.UpStr type assertion failed")
				}
				hemi1, ok := ch1.Quartet.DRS.ASMCard.UpStr.(*dpgadetector.Hemisphere)
				if !ok {
					panic("ch0.Quartet.DRS.ASMCard.UpStr type assertion failed")
				}
				if hemi0.Which() == hemi1.Which() {
					doMinRec = false
				}
			}
			if doMinRec {
				xbeam, ybeam := 0., 0.
				x, y, z := reconstruction.Minimal(ch0, ch1, xbeam, ybeam)
				dqplots.HMinRecX.Fill(x, 1)
				dqplots.HMinRecY.Fill(y, 1)
				dqplots.HMinRecZ.Fill(z, 1)

				pulsesWithSignal[0].CalcRisingFront(true)
				pulsesWithSignal[1].CalcRisingFront(true)
				pulsesWithSignal[0].CalcFallingFront(false)
				pulsesWithSignal[1].CalcFallingFront(false)
				treeMult2.Fill(hdr.RunNumber, uint32(event.ID), pulsesWithSignal[0], pulsesWithSignal[1], x, y, z)
			}
		}
		////////////////////////////////////////////////////////////

		//event.Print(true)
	}
	dqplots.Finalize()

	tpL := dqplots.MakeChargeAmplTiledPlot(dq.Amplitude, dpgadetector.Left)
	tpL.Save(150*vg.Centimeter, 100*vg.Centimeter, "ChargeDistribTiledLeftHemi.png")
	tpR := dqplots.MakeChargeAmplTiledPlot(dq.Amplitude, dpgadetector.Right)
	tpR.Save(150*vg.Centimeter, 100*vg.Centimeter, "ChargeDistribTiledRightHemi.png")

	dqplots.WriteGob(*wGob)
	dqplots.SaveHistos()

	treeMult2.Close()
}
