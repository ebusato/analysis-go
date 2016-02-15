package dq

import (
	"encoding/gob"
	"fmt"
	"os"

	"github.com/go-hep/hbook"
	"github.com/gonum/plot/vg"
	"github.com/gonum/plot/vg/draw"
	"gitlab.in2p3.fr/avirm/analysis-go/dpga/event"
	"gitlab.in2p3.fr/avirm/analysis-go/utils"
)

type DQPlot struct {
	Nevents          uint
	HFrequency       *hbook.H1D
	HSatFrequency    *hbook.H1D
	HMultiplicity    *hbook.H1D
	HSatMultiplicity *hbook.H1D
}

func NewDQPlot() *DQPlot {
	dqp := &DQPlot{
		HFrequency:       hbook.NewH1D(240, 0, 240),
		HSatFrequency:    hbook.NewH1D(240, 0, 240),
		HMultiplicity:    hbook.NewH1D(8, 0, 8),
		HSatMultiplicity: hbook.NewH1D(8, 0, 8),
	}
	return dqp
}

func NewDQPlotFromGob(fileName string) *DQPlot {
	fmt.Printf("Opening gob file %s\n", fileName)
	f, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	dec := gob.NewDecoder(f)

	fmt.Printf("Loading DQ from %s\n", fileName)

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
}

func (d *DQPlot) WriteHistosToFile(refs ...string) {
	doplot := utils.MakeHPl
	// 	doplot := utils.MakeGonumPlot

	dqplotref := &DQPlot{}

	if len(refs) != 0 {
		dqplotref = NewDQPlotFromGob(refs[0])
	}

	linestyle := draw.LineStyle{Width: vg.Points(2)}
	linestyleref := draw.LineStyle{Width: vg.Points(1), Dashes: []vg.Length{vg.Points(5), vg.Points(5)}}

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
	fmt.Printf("Creating gob file %s\n", fileName)
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
