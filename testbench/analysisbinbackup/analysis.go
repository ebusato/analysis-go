package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	"gitlab.in2p3.fr/avirm/analysis-go/applyCorrCalib"
	"gitlab.in2p3.fr/avirm/analysis-go/event"
	"gitlab.in2p3.fr/avirm/analysis-go/pulse"
	"gitlab.in2p3.fr/avirm/analysis-go/testbench/dq"
	"gitlab.in2p3.fr/avirm/analysis-go/testbench/rw"
	"gitlab.in2p3.fr/avirm/analysis-go/testbench/tbdetector"
)

func main() {

	log.SetFlags(log.Llongfile | log.LstdFlags)

	var (
		infileName = flag.String("i", "testdata/tenevents_hex.txt", "Name of the input file")
		//outFileNamePulses = flag.String("oP", "output/pulses.csv", "Name of the output file containing pulse data")
		//outFileNameGlobal = flag.String("oG", "output/globalEventVariables.csv", "Name of the output file containing global event variables")
		noEvents = flag.Int("n", -1, "Number of events to process (-1 means all events are processed)")

		// flags specific to pedestal computation
		computePed  = flag.Bool("computePed", false, "Compute pedestals if set to true")
		outfileName = flag.String("o", "output/pedestals.csv", "Name of the file containing pedestal data")

		// flags specific to analysis
		applyCorrections = flag.Bool("corr", false, "Do corrections and calibration or not")
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

	file, err := os.Open(*infileName)
	if err != nil {
		log.Fatalf("error opening file %v", err)
	}
	defer file.Close()

	if *applyCorrections {
		tbdetector.Det.ReadPedestalsFile("../calibConstants/pedestals.csv")
	}

	r, err := rw.NewReader(bufio.NewReader(file))
	if err != nil {
		log.Fatalf("could not open asm file: %v\n", err)
	}

	data := event.NewData(5000)

	dqplot := dq.NewDQPlot()

	for event, status := r.ReadNextEvent(); status && (*noEvents == -1 || int(event.ID) < *noEvents); event, status = r.ReadNextEvent() {
		if event.ID%100 == 0 {
			fmt.Printf("Processing event %v\n", event.ID)
		}
		if *applyCorrections {
			event = applyCorrCalib.RemovePedestal(event)
		}
		// 		if event.ID == 17 {
		// 			event.Print(true)
		// 		}
		//fmt.Println("correlation=", event.GlobalCorrelation("PMT1", "PMT2"))
		data.Events = append(data.Events, *event)
		dqplot.FillHistos(event)
	}

	data.CheckIntegrity()
	switch {
	case *computePed:
		ComputePedestals(&data)
		tbdetector.Det.WritePedestalsToFile(*outfileName)
		tbdetector.Det.PlotPedestals(true)
		tbdetector.Det.PlotPedestals(false)
	default: // analysis
		//data.PrintPulsesToFile(*outFileNamePulses)
		//data.PrintGlobalVarsToFile(*outFileNameGlobal)
		dqplot.Finalize()
		dqplot.WriteHistosToFile("../dqref/dqplots_ref.gob")
		dqplot.WriteGob("dqplots.gob")
		data.PlotPulses(pulse.XaxisIndex, false)
		data.PlotAmplitudeCorrelationWithinCluster()
	}
}

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
	tbdetector.Det.ComputePedestalsMeanStdDevFromSamples()
}
