package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"gitlab.in2p3.fr/avirm/analysis-go/dpgatca/rw"
	"gitlab.in2p3.fr/avirm/analysis-go/dpgatca/rwi"
	"gitlab.in2p3.fr/avirm/analysis-go/dpgatca/rwvme"
	"go-hep.org/x/hep/csvutil"
)

var (
	inFile   = flag.String("i", "", "Name of input file")
	outFile  = flag.String("o", "", "Name of the output csv file")
	noEvents = flag.Uint("n", 100000, "Number of events")
	evtFreq  = flag.Uint("ef", 100, "Event printing frequency")
	vme      = flag.Bool("vme", false, "If set, uses VME reader")
)

func main() {
	flag.Parse()

	// Reader
	f, err := os.Open(*inFile)
	if err != nil {
		log.Fatalf("could not open data file: %v\n", err)
	}
	defer f.Close()

	/*r, err := rw.NewReader(bufio.NewReader(f))
	if err != nil {
		log.Fatalf("could not open stream: %v\n", err)
	}*/
	var r rwi.Reader
	if !*vme {
		r, _ = rw.NewReader(bufio.NewReader(f))
	} else {
		r, _ = rwvme.NewReader(bufio.NewReader(f), rwvme.HeaderCAL)
	}

	// csv file containing output of pedestal analysis
	tbl, err := csvutil.Create(*outFile)
	if err != nil {
		log.Fatalf("could not create output file: %v\n", err)
	}
	defer tbl.Close()
	tbl.Writer.Comma = ','

	//err = tbl.WriteHeader(fmt.Sprintf("# Output of pedestal analysis (creation date: %v).\n", time.Now()))
	//ss := "# "
	ss := ""
	for i := 0; i < 240; i++ {
		ss += fmt.Sprintf("ChanWithData240_%v,", i)
	}
	for i := 0; i < 48; i++ {
		if i < 47 {
			ss += fmt.Sprintf("ChanWoData48_%v,", i)
		} else {
			ss += fmt.Sprintf("ChanWoData48_%v", i)
		}
	}
	err = tbl.WriteHeader(ss)

	var iEvent uint

	for iEvent < *noEvents {
		if iEvent%*evtFreq == 0 {
			fmt.Printf("event %v\n", iEvent)
		}
		event, status := r.ReadNextEvent()
		if status == false {
			panic("error: status is false\n")
		}
		// 		event.Print(false, false)
		// 		fmt.Println(event.HasSignal())
		sigInClusters, sigInClustersWoData := event.HasSignal()
		if sigInClusters == true || sigInClustersWoData == true {
			continue
		}
		amps := event.AmpsPerChannel()
		stamps := strings.Fields(strings.Trim(fmt.Sprint(amps), "[]"))
		err := tbl.Writer.Write(stamps)
		if err != nil {
			log.Fatalf("error writing row: %v\n", err)
		}
		iEvent++
	}
}
