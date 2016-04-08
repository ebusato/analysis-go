// Package dq implements structures and functions to be used for data quality assessment.
package dq

import (
	"encoding/gob"
	"fmt"
	"os"
	"strconv"

	"github.com/go-hep/hbook"
	"github.com/go-hep/hplot"
	"github.com/gonum/plot/plotutil"
	"github.com/gonum/plot/vg"
	"github.com/gonum/plot/vg/draw"
	"gitlab.in2p3.fr/avirm/analysis-go/dpga/dpgadetector"
	"gitlab.in2p3.fr/avirm/analysis-go/event"
	"gitlab.in2p3.fr/avirm/analysis-go/utils"
)

type DQPlot struct {
	Nevents          uint
	HFrequency       *hbook.H1D
	HSatFrequency    *hbook.H1D
	HMultiplicity    *hbook.H1D
	HSatMultiplicity *hbook.H1D
	HCharge          [][]hbook.H1D
	HAmplitude       [][]hbook.H1D
}

func NewDQPlot() *DQPlot {
	const N = 4
	NoClusters := dpgadetector.Det.NoClusters()
	dqp := &DQPlot{
		HFrequency:       hbook.NewH1D(240, 0, 240),
		HSatFrequency:    hbook.NewH1D(240, 0, 240),
		HMultiplicity:    hbook.NewH1D(8, 0, 8),
		HSatMultiplicity: hbook.NewH1D(8, 0, 8),
		HCharge:          make([][]hbook.H1D, NoClusters),
		HAmplitude:       make([][]hbook.H1D, NoClusters),
	}
	for i := uint8(0); i < NoClusters; i++ {
		dqp.HCharge[i] = make([]hbook.H1D, N)
		dqp.HAmplitude[i] = make([]hbook.H1D, N)
		for j := 0; j < N; j++ {
			dqp.HCharge[i][j] = *hbook.NewH1D(50, 1e4, 1000e3)
			dqp.HAmplitude[i][j] = *hbook.NewH1D(100, 0, 6000)
		}
	}
	return dqp
}

func NewDQPlotFromGob(fileName string) *DQPlot {
	fmt.Printf("Reading gob file %s\n", fileName)
	f, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	dec := gob.NewDecoder(f)

	dqplot := NewDQPlot()

	err = dec.Decode(dqplot)
	if err != nil {
		panic(err)
	}

	err = f.Close()
	if err != nil {
		panic(err)
	}
	return dqplot
}

func (d *DQPlot) FillHistos(event *event.Event) {
	d.Nevents++

	var mult uint8 = 0
	var satmult uint8 = 0
	var counter float64 = 0

	for i := range event.Clusters {
		cluster := &event.Clusters[i]
		mult += uint8(len(cluster.PulsesWithSignal()))
		satmult += uint8(len(cluster.PulsesWithSatSignal()))
		for j := range cluster.Pulses {
			pulse := &cluster.Pulses[j]
			d.HCharge[i][j].Fill(float64(pulse.Charge()), 1)
			d.HAmplitude[i][j].Fill(float64(pulse.Amplitude()), 1)
			if pulse.HasSignal {
				d.HFrequency.Fill(counter, 1)
			}
			if pulse.HasSatSignal {
				d.HSatFrequency.Fill(counter, 1)
			}
			counter++
		}
	}

	d.HMultiplicity.Fill(float64(mult), 1)
	d.HSatMultiplicity.Fill(float64(satmult), 1)
}

func (d *DQPlot) Finalize() {
	d.HFrequency.Scale(1 / float64(d.Nevents))
	d.HSatFrequency.Scale(1 / float64(d.Nevents))
	d.HMultiplicity.Scale(1 / d.HMultiplicity.Integral())
	d.HSatMultiplicity.Scale(1 / d.HSatMultiplicity.Integral())
	// Take len of HCharge and HCharge[0] as it should be the same for all other
	// objects used here
	for i := 0; i < len(d.HCharge); i++ {
		for j := 0; j < len(d.HCharge[0]); j++ {
			d.HCharge[i][j].Scale(1 / d.HCharge[i][j].Integral())
			d.HAmplitude[i][j].Scale(1 / d.HAmplitude[i][j].Integral())
		}
	}
}

type WhichVar byte

const (
	Charge WhichVar = iota
	Amplitude
)

func (d *DQPlot) MakeChargeAmplTiledPlot(which WhichVar) *hplot.TiledPlot {
	tp, err := hplot.NewTiledPlot(draw.Tiles{Cols: 6, Rows: 10, PadY: 1 * vg.Centimeter})
	//tp, err := hplot.NewTiledPlot(draw.Tiles{Cols: 2, Rows: 2, PadY: 1 * vg.Centimeter})
	if err != nil {
		panic(err)
	}
	iCluster := 0
	histos := make([]hbook.H1D, len(d.HCharge[0]))

	switch which {
	case Charge:
		histos = d.HCharge[iCluster]
	case Amplitude:
		histos = d.HAmplitude[iCluster]
	}
	for irows := 0; irows < tp.Tiles.Rows; irows++ {
		for icols := 0; icols < tp.Tiles.Cols; icols++ {
			p := tp.Plot(irows, icols)
			p.X.Tick.Marker = &hplot.FreqTicks{N: 10, Freq: 4}
			p.Add(hplot.NewGrid())
			hplotcharge0, err := hplot.NewH1D(&histos[0])
			if err != nil {
				panic(err)
			}
			hplotcharge1, err := hplot.NewH1D(&histos[1])
			if err != nil {
				panic(err)
			}
			hplotcharge2, err := hplot.NewH1D(&histos[2])
			if err != nil {
				panic(err)
			}
			hplotcharge3, err := hplot.NewH1D(&histos[3])
			if err != nil {
				panic(err)
			}
			hplotcharge0.Color = plotutil.Color(1)
			hplotcharge1.Color = plotutil.Color(2)
			hplotcharge2.Color = plotutil.Color(3)
			hplotcharge3.Color = plotutil.Color(4)
			p.Add(hplotcharge0)
			p.Add(hplotcharge1)
			p.Add(hplotcharge2)
			p.Add(hplotcharge3)
			iCluster++
		}
	}
	return tp
}

// SaveHistos saves histograms on disk.
// If refs is specified, current histograms
// are overlaid with reference histograms
// located in the gob file provided.
func (d *DQPlot) SaveHistos(refs ...string) {
	doplot := utils.MakeHPl
	// 	doplot := utils.MakeGonumPlot

	dqplotref := &DQPlot{}

	if len(refs) != 0 && refs[0] != "" {
		dqplotref = NewDQPlotFromGob(refs[0])
	}

	linestyle := draw.LineStyle{Width: vg.Points(2)}
	linestyleref := draw.LineStyle{Width: vg.Points(1), Dashes: []vg.Length{vg.Points(5), vg.Points(5)}}

	for i := range d.HAmplitude {
		doplot("Amplitude",
			"Entries",
			"output/distribAmplitude"+strconv.FormatInt(int64(i), 10)+".png",
			utils.H1dToHplot(linestyle, d.HAmplitude[i]...)...)
		doplot("Charge",
			"Entries",
			"output/distribCharge"+strconv.FormatInt(int64(i), 10)+".png",
			utils.H1dToHplot(linestyle, d.HCharge[i]...)...)
	}
	doplot("Channel",
		"# pulses / event",
		"output/distribFrequency.png",
		append(utils.H1dptrToHplot(linestyle, d.HFrequency),
			utils.H1dptrToHplot(linestyleref, dqplotref.HFrequency)...)...)
	doplot("Channel",
		"# pulses with saturation / event",
		"output/distribSatFrequency.png",
		append(utils.H1dptrToHplot(linestyle, d.HSatFrequency),
			utils.H1dptrToHplot(linestyleref, dqplotref.HSatFrequency)...)...)
	doplot("Multiplicity",
		"Entries",
		"output/distribMultiplicity.png",
		append(utils.H1dptrToHplot(linestyle, d.HMultiplicity),
			utils.H1dptrToHplot(linestyleref, dqplotref.HMultiplicity)...)...)
	doplot("Multiplicity of pulses with saturation",
		"Entries",
		"output/distribSatMultiplicity.png",
		append(utils.H1dptrToHplot(linestyle, d.HSatMultiplicity),
			utils.H1dptrToHplot(linestyleref, dqplotref.HSatMultiplicity)...)...)
}

func (d *DQPlot) WriteGob(fileName string) error {
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := gob.NewEncoder(f)
	fmt.Printf("Saving DQplot to %s\n", fileName)
	err = enc.Encode(d)
	if err != nil {
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}

	return err
}
