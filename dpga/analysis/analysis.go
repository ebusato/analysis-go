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
		pedCorr    = flag.String("ped", "", "Name of the csv file containing pedestal constants. If not set, pedestal corrections are not applied.")
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

		// begin here
		for iCluster := range event.Clusters {
			cluster := &event.Clusters[iCluster]
			for iPulse := range cluster.Pulses {
				pulse := &cluster.Pulses[iPulse]
				if pulse.HasSignal {
					continue
				}
				for iSample := range pulse.Samples {
					sample := &pulse.Samples[iSample]
					capacitor := sample.Capacitor
					noSamples := capacitor.NoPedestalSamples()
					if event.ID == 0 && noSamples != 0 {
						log.Fatal("noSamples != 0!")
					}
					capacitor.TempSum += sample.Amplitude
				}
			}
		}
		// end here

		///////////////////////////////////////////////////////////
		// Plotting
		// pulses
		if event.ID < 2 {
			event.PlotPulses(pulse.XaxisCapacitor, false, true)
		}
		// dq
		dqplots.FillHistos(event)
		////////////////////////////////////////////////////////////

		//event.Print(true)
	}
	dqplots.Finalize()
	// begin here
	for iHemi := range dpgadetector.Det.Hemispheres() {
		hemi := dpgadetector.Det.Hemisphere(iHemi)
		for iASM := range hemi.ASMs() {
			asm := hemi.ASM(iASM)
			for iDRS := range asm.DRSs() {
				drs := asm.DRS(uint8(iDRS))
				for iQuartet := range drs.Quartets() {
					quartet := drs.Quartet(uint8(iQuartet))
					for iChannel := range quartet.Channels() {
						ch := quartet.Channel(uint8(iChannel))
						for iCapacitor := range ch.Capacitors() {
							capa := ch.Capacitor(uint16(iCapacitor))
							fmt.Println("value (", iCapacitor, iChannel, iQuartet, iDRS, iASM, iHemi, ") =", capa.TempSum)
						}
					}
				}
			}
		}
	}
	// end here
	tpL := dqplots.MakeChargeAmplTiledPlot(dq.Amplitude, dpgadetector.Left)
	tpL.Save(150*vg.Centimeter, 100*vg.Centimeter, "ChargeDistribTiledLeftHemi.png")
	tpR := dqplots.MakeChargeAmplTiledPlot(dq.Amplitude, dpgadetector.Right)
	tpR.Save(150*vg.Centimeter, 100*vg.Centimeter, "ChargeDistribTiledRightHemi.png")

	dqplots.WriteGob(*wGob)
	dqplots.SaveHistos()

}
