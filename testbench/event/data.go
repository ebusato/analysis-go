package event

import (
	"fmt"
	"image/color"
	"log"
	"time"

	"github.com/go-hep/csvutil"
	"github.com/go-hep/hbook"
	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/plotutil"
	"github.com/gonum/plot/vg"
	"gitlab.in2p3.fr/avirm/analysis-go/pulse"
)

type Data struct {
	Events                           []Event
	HCharge                          []hbook.H1D
	HAmplitude                       []hbook.H1D
	HHasSignal                       []hbook.H1D
	HHasSatSignal                    []hbook.H1D
	HSRout                           *hbook.H1D
	HMultiplicity                    *hbook.H1D
	HClusterCharge                   *hbook.H1D
	HClusterChargeMultiplicityEq1    *hbook.H1D
	HClusterChargeMultiplicityEq2    *hbook.H1D
	HClusterAmplitude                *hbook.H1D
	HClusterAmplitudeMultiplicityEq1 *hbook.H1D
	HClusterAmplitudeMultiplicityEq2 *hbook.H1D
}

func NewData() *Data {
	const N = 4
	data := &Data{
		HCharge:                          make([]hbook.H1D, N),
		HAmplitude:                       make([]hbook.H1D, N),
		HHasSignal:                       make([]hbook.H1D, N),
		HHasSatSignal:                    make([]hbook.H1D, N),
		HSRout:                           hbook.NewH1D(1024, 0, 1023),
		HMultiplicity:                    hbook.NewH1D(5, 0, 5),
		HClusterCharge:                   hbook.NewH1D(100, -2e4, 400e3),
		HClusterChargeMultiplicityEq1:    hbook.NewH1D(100, -2e4, 400e3),
		HClusterChargeMultiplicityEq2:    hbook.NewH1D(100, -2e4, 400e3),
		HClusterAmplitude:                hbook.NewH1D(100, 0, 15000),
		HClusterAmplitudeMultiplicityEq1: hbook.NewH1D(100, 0, 15000),
		HClusterAmplitudeMultiplicityEq2: hbook.NewH1D(100, 0, 15000),
	}

	for i := 0; i < N; i++ {
		data.HCharge[i] = *hbook.NewH1D(100, -2e4, 100e3)
		data.HAmplitude[i] = *hbook.NewH1D(100, 0, 4200)
		data.HHasSignal[i] = *hbook.NewH1D(2, 0, 2)
		data.HHasSatSignal[i] = *hbook.NewH1D(2, 0, 2)
	}

	return data
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

func (d *Data) FillHistos(event *Event) {
	cluster := &event.Cluster
	d.HSRout.Fill(float64(cluster.SRout()), 1)
	d.HClusterCharge.Fill(float64(cluster.Charge()), 1)
	d.HClusterAmplitude.Fill(float64(cluster.Amplitude()), 1)

	multi := len(cluster.PulsesWithSignal())
	d.HMultiplicity.Fill(float64(multi), 1)
	switch multi {
	case 1:
		d.HClusterChargeMultiplicityEq1.Fill(float64(cluster.Charge()), 1)
		d.HClusterAmplitudeMultiplicityEq1.Fill(float64(cluster.Amplitude()), 1)
	case 2:
		d.HClusterChargeMultiplicityEq2.Fill(float64(cluster.Charge()), 1)
		d.HClusterAmplitudeMultiplicityEq2.Fill(float64(cluster.Amplitude()), 1)
	}

	for j := range cluster.Pulses {
		pulse := &cluster.Pulses[j]
		d.HCharge[j].Fill(float64(pulse.Charge()), 1)
		d.HAmplitude[j].Fill(float64(pulse.Amplitude()), 1)
		hasSig := 0
		switch pulse.HasSignal {
		case true:
			hasSig = 1
		case false:
			hasSig = 0
		}
		d.HHasSignal[j].Fill(float64(hasSig), 1)
		hasSatSig := 0
		switch pulse.HasSatSignal {
		case true:
			hasSatSig = 1
		case false:
			hasSatSig = 0
		}
		d.HHasSatSignal[j].Fill(float64(hasSatSig), 1)
	}
}

func HbookToGonum(histo ...hbook.H1D) []plotter.Histogram {
	output := make([]plotter.Histogram, len(histo))
	for i, h := range histo {
		h, err := plotter.NewHistogram(&h, h.Axis().Bins())
		if err != nil {
			panic(err)
		}
		h.FillColor = nil //plotutil.Color(i)
		h.Color = plotutil.Color(i)
		output[i] = *h
	}
	return output
}

func MakePlot(xTitle string, yTitle string, outFile string, histo ...hbook.H1D) {
	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	p.X.Label.Text = xTitle
	p.Y.Label.Text = yTitle

	hGonum := HbookToGonum(histo...)
	for i := range hGonum {
		p.Add(&hGonum[i])
	}

	if err := p.Save(4*vg.Inch, 4*vg.Inch, outFile); err != nil {
		panic(err)
	}
}

// func H1dToHplot(histo ...hbook.H1D) []hplot.Histogram {
// 	output := make([]hplot.Histogram, len(histo))
// 	for i, h := range histo {
// 		// 		hi, err := hplot.NewHistogram(&h, h.Axis().Bins())
// 		// 		if err != nil {
// 		// 			panic(err)
// 		// 		}
// 		hi, err := hplot.NewH1D(&h)
// 		if err != nil {
// 			panic(err)
// 		}
// 		//hi.FillColor = nil //plotutil.Color(i)
// 		hi.Color = plotutil.Color(i)
// 		output[i] = *hi
// 	}
// 	return output
// }
//
// func MakePlot(xTitle string, yTitle string, outFile string, histo ...hbook.H1D) {
// 	p, err := hplot.New()
// 	if err != nil {
// 		panic(err)
// 	}
// 	p.X.Label.Text = xTitle
// 	p.Y.Label.Text = yTitle
//
// 	p.Y.Min = 0
//
// 	hHplot := H1dToHplot(histo...)
// 	for i := range hHplot {
// 		p.Add(&hHplot[i])
// 	}
// 	// 	p.Add(hHplot)
//
// 	if err := p.Save(4*vg.Inch, 4*vg.Inch, outFile); err != nil {
// 		panic(err)
// 	}
// }

func (d *Data) WriteHistosToFile() {
	MakePlot("Charge", "Entries (A. U.)", "output/distribCharge.png", d.HCharge...)
	MakePlot("Amplitude", "Entries (A. U.)", "output/distribAmplitude.png", d.HAmplitude...)
	MakePlot("HasSignal", "Entries (A. U.)", "output/distribHasSignal.png", d.HHasSignal...)
	MakePlot("HasSatSignal", "Entries (A. U.)", "output/distribHasSatSignal.png", d.HHasSatSignal...)
	MakePlot("SRout", "Entries (A. U.)", "output/distribSRout.png", *d.HSRout)
	MakePlot("Multiplicity", "Entries (A. U.)", "output/distribMultiplicity.png", *d.HMultiplicity)
	MakePlot("Cluster charge", "Entries (A. U.)", "output/distribClusterCharge.png", *d.HClusterCharge)
	MakePlot("Cluster charge (multiplicity = 1)", "Entries (A. U.)", "output/distribClusterChargeMultiplicityEq1.png", *d.HClusterChargeMultiplicityEq1)
	MakePlot("Cluster charge (multiplicity = 2)", "Entries (A. U.)", "output/distribClusterChargeMultiplicityEq2.png", *d.HClusterChargeMultiplicityEq2)
	MakePlot("Cluster amplitude", "Entries (A. U.)", "output/distribClusterAmplitude.png", *d.HClusterAmplitude)
	MakePlot("Cluster amplitude (multiplicity = 1)", "Entries (A. U.)", "output/distribClusterAmplitudeMultiplicityEq1.png", *d.HClusterAmplitudeMultiplicityEq1)
	MakePlot("Cluster amplitude (multiplicity = 2)", "Entries (A. U.)", "output/distribClusterAmplitudeMultiplicityEq2.png", *d.HClusterAmplitudeMultiplicityEq2)
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

func (d *Data) Plot() {
	//d.PlotDistribs()
	d.PlotPulses(pulse.XaxisCapacitor, false, true)
}
