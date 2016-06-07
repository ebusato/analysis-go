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
	"gitlab.in2p3.fr/avirm/analysis-go/event"
)

func ComputePedestals(data *event.Data) {
	for iEvent := range data.Events {
		event := data.Events[iEvent]
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
					if iEvent == 0 && noSamples != 0 {
						log.Fatal("len(capacitor.Pedestal()) != 0!")
					}
					capacitor.AddPedestalSample(sample.Amplitude)
				}
			}
		}
	}
	dpgadetector.Det.ComputePedestalsMeanStdDevFromSamples()
}

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

	var data event.Data

	for event, status := r.ReadNextEvent(); status && event.ID < *noEvents; event, status = r.ReadNextEvent() {
		if event.ID%500 == 0 {
			fmt.Printf("Processing event %v\n", event.ID)
		}
		//event.Print(true)
		data.Events = append(data.Events, *event)
	}

	//data.CheckIntegrity()
	//data.PlotPulses(pulse.XaxisCapacitor, true, false)

	ComputePedestals(&data)

	dpgadetector.Det.WritePedestalsToFile(*outfileName)

	//dpgadetector.Det.PlotPedestals(true)
	dpgadetector.Det.PlotPedestals(false)
}
