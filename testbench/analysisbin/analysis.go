package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	"gitlab.in2p3.fr/avirm/analysis-go/pulse"
	"gitlab.in2p3.fr/avirm/analysis-go/testbench/applyCorrCalib"
	"gitlab.in2p3.fr/avirm/analysis-go/testbench/dq"
	"gitlab.in2p3.fr/avirm/analysis-go/testbench/event"
	"gitlab.in2p3.fr/avirm/analysis-go/testbench/rw"
	"gitlab.in2p3.fr/avirm/analysis-go/testbench/tbdetector"
)

func main() {

	log.SetFlags(log.Llongfile | log.LstdFlags)

	var (
		infileName = flag.String("i", "testdata/tenevents_hex.txt", "Name of the input file")
		//outFileNamePulses = flag.String("oP", "output/pulses.csv", "Name of the output file containing pulse data")
		//outFileNameGlobal = flag.String("oG", "output/globalEventVariables.csv", "Name of the output file containing global event variables")
		noEvents         = flag.Uint("n", 10, "Number of events to process")
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

	var data event.Data

	dqplot := dq.NewDQPlot()

	for event, status := r.ReadNextEvent(); status && event.ID < *noEvents; event, status = r.ReadNextEvent() {
		if event.ID%500 == 0 {
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
	//data.PrintPulsesToFile(*outFileNamePulses)
	//data.PrintGlobalVarsToFile(*outFileNameGlobal)
	dqplot.Finalize()
	//dqplot.WriteHistosToFile("../dqref/dqplots_ref.gob")
	//dqplot.WriteGob("dqplots.gob")
	data.PlotPulses(pulse.XaxisTime, false, true)
	data.PlotAmplitudeCorrelationWithinCluster()
}
