package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	"gitlab.in2p3.fr/avirm/analysis-go/dpga/applyCorrCalib"
	"gitlab.in2p3.fr/avirm/analysis-go/dpga/dpgadetector"
	"gitlab.in2p3.fr/avirm/analysis-go/dpga/dq"
	"gitlab.in2p3.fr/avirm/analysis-go/dpga/event"
	"gitlab.in2p3.fr/avirm/analysis-go/dpga/reader"
	"gitlab.in2p3.fr/avirm/analysis-go/pulse"
)

func main() {
	log.SetFlags(log.Llongfile | log.LstdFlags)

	var (
		infileName       = flag.String("i", "testdata/tenevents_hex.txt", "Name of the input file")
		noEvents         = flag.Uint("n", 10000000, "Number of events to process")
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

	r, err := reader.NewReader(bufio.NewReader(file))
	if err != nil {
		log.Fatalf("could not open asm file: %v\n", err)
	}

	dpgadetector.Det.ReadPedestalsFile("../calibConstants/pedestals.csv")

	var data event.Data

	dqplots := dq.NewDQPlot()

	for event, status := r.ReadNextEvent(); status && event.ID < *noEvents; event, status = r.ReadNextEvent() {
		if event.ID%50 == 0 {
			fmt.Printf("Processing event %v\n", event.ID)
		}
		if *applyCorrections {
			event = applyCorrCalib.RemovePedestal(event)
		}
		dqplots.FillHistos(event)
		//event.Print(true)
		data = append(data, *event)
	}

	data.CheckIntegrity()

	dqplots.Finalize()
	dqplots.WriteHistosToFile("../dqref/dqplots.gob")
	dqplots.WriteGob("dqplots.gob")
	data.PlotPulses(pulse.XaxisTime, false)
}
