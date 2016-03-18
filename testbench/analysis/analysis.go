package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gitlab.in2p3.fr/avirm/analysis-go/applyCorrCalib"
	"gitlab.in2p3.fr/avirm/analysis-go/event"
	"gitlab.in2p3.fr/avirm/analysis-go/pulse"
	"gitlab.in2p3.fr/avirm/analysis-go/testbench/dq"
	"gitlab.in2p3.fr/avirm/analysis-go/testbench/reader"
	"gitlab.in2p3.fr/avirm/analysis-go/testbench/rw"
	"gitlab.in2p3.fr/avirm/analysis-go/testbench/tbdetector"
)

type ReadNextEventer interface {
	ReadNextEvent() (*event.Event, bool)
}

func main() {
	log.SetFlags(log.Llongfile | log.LstdFlags)

	var (
		infileName = flag.String("i", "testdata/tenevents_hex.txt", "Name of the input file")
		//outFileNamePulses = flag.String("oP", "output/pulses.csv", "Name of the output file containing pulse data")
		//outFileNameGlobal = flag.String("oG", "output/globalEventVariables.csv", "Name of the output file containing global event variables")
		noEvents  = flag.Int("n", -1, "Number of events to process (-1 means all events are processed)")
		pedCorr   = flag.String("pedcorr", "", "Name of the csv file containing pedestal constants. If not set, pedestal corrections are not applied.")
		wGob      = flag.String("wgob", "dqplots.gob", "Name of the output gob file containing dq plots. If not set, the gob file is not produced.")
		refGob    = flag.String("refgob", "", "Name of the gob file containing reference dq plots. If not set, reference dq plots are not overlaid to current dq plots.")
		inputType = reader.HexInput
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

	var rner ReadNextEventer

	if filepath.Ext(*infileName) == ".bin" {
		// Binary file reader
		rner, err = rw.NewReader(bufio.NewReader(file))
		if err != nil {
			log.Fatalf("could not open asm file: %v\n", err)
		}
	} else if filepath.Ext(*infileName) == ".txt" {
		// ASCII file scanner
		s := reader.NewScanner(bufio.NewScanner(file))
		s.SetInputType(inputType)
		rner = s
	} else {
		log.Fatalf("file extension not recognized.")
	}

	// Start doing concrete analysis

	if *pedCorr != "" {
		tbdetector.Det.ReadPedestalsFile(*pedCorr)
	}
	dqplot := dq.NewDQPlot()

	//var dataCorrelation plotter.XYZs

	for event, status := rner.ReadNextEvent(); status && (*noEvents == -1 || int(event.ID) < *noEvents); event, status = rner.ReadNextEvent() {
		if event.ID%500 == 0 {
			fmt.Printf("Processing event %v\n", event.ID)
		}
		///////////////////////////////////////////////////////////
		// Corrections
		if *pedCorr != "" {
			event = applyCorrCalib.RemovePedestal(event)
		}
		///////////////////////////////////////////////////////////

		///////////////////////////////////////////////////////////
		// Plotting
		// pulses
		if event.ID < 40 {
			event.PlotPulses(pulse.XaxisIndex, false)
		}
		// dq
		dqplot.FillHistos(event)
		// correlation
		/*
			cluster := event.Clusters[0]
			pulses := cluster.PulsesWithSignal()
			multiplicity := len(pulses)
			if multiplicity == 2 {
				mydata := struct {
					X, Y, Z float64
				}{
					X: pulses[0].Amplitude(),
					Y: pulses[1].Amplitude(),
					Z: 0,
				}
				dataCorrelation = append(dataCorrelation, mydata)
			}
		*/
		////////////////////////////////////////////////////////////
	}

	/*
		p, err := plot.New()
		if err != nil {
			panic(err)
		}
		p.Title.Text = "Correlation of amplitudes for clusters with 2 pulses"
		p.X.Label.Text = "amplitude 1"
		p.Y.Label.Text = "amplitude 2"

		bs, err := plotter.NewBubbles(dataCorrelation, vg.Points(1), vg.Points(3))
		if err != nil {
			panic(err)
		}
		bs.Color = color.RGBA{R: 196, B: 128, A: 255}
		p.Add(bs)

		if err := p.Save(4*vg.Inch, 4*vg.Inch, "output/bubble.png"); err != nil {
			panic(err)
		}
	*/
	///////////////////////////////////////////////////
	// These two line should not be uncomment, it won't
	// work. They are here just to remind me that I should
	// include these calculations in the new analysis.go.
	// Original functions from event/data.go are
	// copied/pasted below
	//
	//PrintPulsesToFile(*outFileNamePulses)
	//PrintGlobalVarsToFile(*outFileNameGlobal)
	///////////////////////////////////////////////////
	dqplot.Finalize()
	dqplot.WriteGob(*wGob)
	dqplot.SaveHistos(*refGob)
}

/*
// To be updated when used
// This has been copied/pasted from the event/data.go file, which is now deprecated
type PulsesCSV struct {
	EventID uint
	Time    float64
	Ampl1   float64
	Ampl2   float64
	Ampl3   float64
	Ampl4   float64
}

func (d *Data) PrintPulsesToFile(outFileName string) {
	tbl, err := csvutil.Create(outFileName)
	if err != nil {
		log.Fatalf("could not create %s: %v\n", outFileName, err)
	}
	defer tbl.Close()
	tbl.Writer.Comma = ' '

	err = tbl.WriteHeader(fmt.Sprintf("# Pulses file (on line per sample) (creation date: %v)\n", time.Now()))
	err = tbl.WriteHeader("# eventID time ampl1 ampl2 ampl3 ampl4")

	if err != nil {
		log.Fatalf("error writing header: %v\n", err)
	}

	for i := range d.Events {
		e := d.Events[i]
		cluster := &e.Clusters[0]
		for j := range cluster.Pulses[0].Samples {
			data := PulsesCSV{
				EventID: e.ID,
				Time:    cluster.Pulses[0].Samples[j].Time,
				Ampl1:   cluster.Pulses[0].Samples[j].Amplitude,
				Ampl2:   cluster.Pulses[1].Samples[j].Amplitude,
				Ampl3:   cluster.Pulses[2].Samples[j].Amplitude,
				Ampl4:   cluster.Pulses[3].Samples[j].Amplitude,
			}
			err = tbl.WriteRow(data)
			if err != nil {
				log.Fatalf("error writing row: %v\n", err)
			}
		}
	}

	err = tbl.Close()
	if err != nil {
		log.Fatalf("error closing table: %v\n", err)
	}
}

type ClusterCSV struct {
	EventID   uint
	PulseID   uint
	HasSignal uint8
	Amplitude float64
	Charge    float64
	SRout     uint16
}

func (d *Data) PrintGlobalVarsToFile(outFileName string) {
	tbl, err := csvutil.Create(outFileName)
	if err != nil {
		log.Fatalf("could not create %s: %v\n", outFileName, err)
	}
	defer tbl.Close()
	tbl.Writer.Comma = ' '

	err = tbl.WriteHeader(fmt.Sprintf("# Cluster file (on line per pulse) (creation date: %v)\n", time.Now()))
	err = tbl.WriteHeader("# eventID PulseID HasSignal Amplitude Charge SRout")

	if err != nil {
		log.Fatalf("error writing header: %v\n", err)
	}

	for i := range d.Events {
		e := d.Events[i]
		cluster := &e.Clusters[0]
		for j := range cluster.Pulses {
			pulse := &cluster.Pulses[j]
			hasSignal := uint8(0)
			if pulse.HasSignal {
				hasSignal = 1
			}
			data := ClusterCSV{
				EventID:   e.ID,
				PulseID:   uint(j),
				HasSignal: hasSignal,
				Amplitude: pulse.Amplitude(),
				Charge:    pulse.Charge(),
				SRout:     pulse.SRout,
			}
			err = tbl.WriteRow(data)
			if err != nil {
				log.Fatalf("error writing row: %v\n", err)
			}
		}
	}

	err = tbl.Close()
	if err != nil {
		log.Fatalf("error closing table: %v\n", err)
	}
}

*/
