package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	"gitlab.in2p3.fr/AVIRM/Analysis-go/dpga/event"
	"gitlab.in2p3.fr/AVIRM/Analysis-go/dpga/reader"
	"gitlab.in2p3.fr/AVIRM/Analysis-go/dpga/tbdetector"
)

func ComputePedestals(data *event.Data) {
	for iEvent := range *data {
		event := &(*data)[iEvent]
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
	tbdetector.Det.ComputePedestalsMeanStdDevFromSamples()
}

func main() {
	var (
		infileName  = flag.String("i", "testdata/tenevents_hex.txt", "Name of the input file")
		outfileName = flag.String("o", "output/pedestals.csv", "Name of the output file")
		noEvents    = flag.Uint("n", 10000000, "Number of events to process")
		inputType   = reader.HexInput
	)
	flag.Var(&inputType, "inType", "Type of input file (possible values: Dec,Hex,Bin)")

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

	s := reader.NewScanner(bufio.NewScanner(file))

	var data event.Data

	for event, status := s.ReadNextEvent(inputType); status && event.ID < *noEvents; event, status = s.ReadNextEvent(inputType) {
		if event.ID%500 == 0 {
			fmt.Printf("Processing event %v\n", event.ID)
		}
		data = append(data, *event)
	}

	data.CheckIntegrity()
	// 	data.PlotPulses(pulse.XaxisCapacitor, true, false)

	ComputePedestals(&data)

	tbdetector.Det.WritePedestalsToFile(*outfileName)

	tbdetector.Det.PlotPedestals(true)
	tbdetector.Det.PlotPedestals(false)

	// detector.TBDet.Print()

}
