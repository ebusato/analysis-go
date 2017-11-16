package event

import (
	"fmt"
	"image/color"
	"log"
	"time"

	"go-hep.org/x/hep/csvutil"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/vg"
	"gitlab.in2p3.fr/avirm/analysis-go/pulse"
)

type Data struct {
	Events []Event
}

func (d *Data) CheckIntegrity() {
	if len(d.Events) == 0 {
		panic("data has 0 events")
	}
	noPulses := d.Events[0].NoPulses()
	for i := 1; i < len(d.Events); i++ {
		no := d.Events[i].NoPulses()
		if no != noPulses {
			panic("not all events have the same number of pulses")
		}
	}
}

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
		for j := range e.Cluster.Pulses[0].Samples {
			data := PulsesCSV{
				EventID: e.ID,
				Time:    e.Cluster.Pulses[0].Samples[j].Time,
				Ampl1:   e.Cluster.Pulses[0].Samples[j].Amplitude,
				Ampl2:   e.Cluster.Pulses[1].Samples[j].Amplitude,
				Ampl3:   e.Cluster.Pulses[2].Samples[j].Amplitude,
				Ampl4:   e.Cluster.Pulses[3].Samples[j].Amplitude,
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
		for j := range e.Cluster.Pulses {
			pulse := &e.Cluster.Pulses[j]
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

func (d *Data) PlotPulses(xaxis pulse.XaxisType, pedestalRange bool, savePulses bool) { /*
		var gsOptions = []string{"-dNOPAUSE", "-dBATCH", "-sDEVICE=pdfwrite", "-sOutputFile=output/merged.pdf"}*/
	var outPulseFiles []string

	for i, event := range d.Events {
		outPulseFiles = append(outPulseFiles, event.PlotPulses(xaxis, pedestalRange))
		if i >= 20 {
			break
		}
	}

	// 	gsOptions = append(gsOptions, outPulseFiles...)
	// 	err := exec.Command("gs", gsOptions...).Run()
	// 	if err != nil {
	// 		log.Fatal("error merging files", err)
	// 	}
	// 	if !savePulses {
	// 		for _, fileName := range outPulseFiles {
	// 			err := exec.Command("rm", "-f", fileName).Run()
	// 			if err != nil {
	// 				log.Fatal("error removing file", err)
	// 			}
	// 		}
	// 	}
}

func (d *Data) PlotAmplitudeCorrelationWithinCluster() {
	var data plotter.XYZs
	for i := range d.Events {
		cluster := d.Events[i].Cluster
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
			data = append(data, mydata)
		}
	}
	p, err := plot.New()
	if err != nil {
		panic(err)
	}
	p.Title.Text = "Correlation of amplitudes for clusters with 2 pulses"
	p.X.Label.Text = "amplitude 1"
	p.Y.Label.Text = "amplitude 2"

	bs, err := plotter.NewBubbles(data, vg.Points(1), vg.Points(3))
	if err != nil {
		panic(err)
	}
	bs.Color = color.RGBA{R: 196, B: 128, A: 255}
	p.Add(bs)

	if err := p.Save(4*vg.Inch, 4*vg.Inch, "output/bubble.png"); err != nil {
		panic(err)
	}
}
