package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	"gitlab.in2p3.fr/AVIRM/Analysis-go/testbench/applyCorrCalib"
	"gitlab.in2p3.fr/AVIRM/Analysis-go/testbench/event"
	"gitlab.in2p3.fr/AVIRM/Analysis-go/testbench/reader"
	"gitlab.in2p3.fr/AVIRM/Analysis-go/testbench/tbdetector"
)

func main() {
	log.SetFlags(log.Llongfile | log.LstdFlags)

	var (
		infileName = flag.String("i", "testdata/tenevents_hex.txt", "Name of the input file")
		//outFileNamePulses = flag.String("oP", "output/pulses.csv", "Name of the output file containing pulse data")
		//outFileNameGlobal = flag.String("oG", "output/globalEventVariables.csv", "Name of the output file containing global event variables")
		noEvents         = flag.Uint("n", 10, "Number of events to process")
		applyCorrections = flag.Bool("applyCorr", false, "Do corrections and calibration or not")
		inputType        = reader.HexInput
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

	tbdetector.Det.ReadPedestalsFile("../computePedestals/output/pedestals.csv")

	s := reader.NewScanner(bufio.NewScanner(file))

	var data event.Data

	for event, status := s.ReadNextEvent(inputType); status && event.ID < *noEvents; event, status = s.ReadNextEvent(inputType) {
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
		data = append(data, *event)
	}

	data.CheckIntegrity()
	//data.PrintPulsesToFile(*outFileNamePulses)
	//data.PrintGlobalVarsToFile(*outFileNameGlobal)
	data.Plot()
	data.AmplitudeCorrelationWithinCluster()
}
