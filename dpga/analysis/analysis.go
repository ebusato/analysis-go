package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"

	"gitlab.in2p3.fr/avirm/analysis-go/dpga/event"
	"gitlab.in2p3.fr/avirm/analysis-go/dpga/reader"
	"gitlab.in2p3.fr/avirm/analysis-go/pulse"
)

func main() {
	log.SetFlags(log.Llongfile | log.LstdFlags)

	var (
		infileName = flag.String("i", "testdata/tenevents_hex.txt", "Name of the input file")
		noEvents   = flag.Uint("n", 10000000, "Number of events to process")
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

	var data event.Data

	for event, status := r.ReadNextEvent(); status && event.ID < *noEvents; event, status = r.ReadNextEvent() {
		if event.ID%1 == 0 {
			fmt.Printf("Processing event %v\n", event.ID)
		}
		//event.Print(true)
		data = append(data, *event)
	}

	//data.CheckIntegrity()
	data.PlotPulses(pulse.XaxisCapacitor, true, false)
}
