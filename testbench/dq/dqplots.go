package dq

import (
	"encoding/gob"
	"fmt"
	"os"

	"github.com/go-hep/hbook"
	"github.com/gonum/plot/vg"
	"github.com/gonum/plot/vg/draw"
	"gitlab.in2p3.fr/avirm/analysis-go/event"
	"gitlab.in2p3.fr/avirm/analysis-go/utils"
)

type DQPlot struct {
	Nevents                          uint
	HCharge                          []hbook.H1D
	HAmplitude                       []hbook.H1D
	HFrequency                       *hbook.H1D
	HSatFrequency                    *hbook.H1D
	HSRout                           *hbook.H1D
	HMultiplicity                    *hbook.H1D
	HSatMultiplicity                 *hbook.H1D
	HClusterCharge                   *hbook.H1D
	HClusterChargeMultiplicityEq1    *hbook.H1D
	HClusterChargeMultiplicityEq2    *hbook.H1D
	HClusterAmplitude                *hbook.H1D
	HClusterAmplitudeMultiplicityEq1 *hbook.H1D
	HClusterAmplitudeMultiplicityEq2 *hbook.H1D
}

func NewDQPlot() *DQPlot {
	const N = 4
	dqp := &DQPlot{
		HCharge:                          make([]hbook.H1D, N),
		HAmplitude:                       make([]hbook.H1D, N),
		HFrequency:                       hbook.NewH1D(4, 0, 4),
		HSatFrequency:                    hbook.NewH1D(4, 0, 4),
		HSRout:                           hbook.NewH1D(1024, 0, 1023),
		HMultiplicity:                    hbook.NewH1D(5, 0, 5),
		HSatMultiplicity:                 hbook.NewH1D(5, 0, 5),
		HClusterCharge:                   hbook.NewH1D(100, -2e4, 400e3),
		HClusterChargeMultiplicityEq1:    hbook.NewH1D(100, -2e4, 400e3),
		HClusterChargeMultiplicityEq2:    hbook.NewH1D(100, -2e4, 400e3),
		HClusterAmplitude:                hbook.NewH1D(100, 0, 15000),
		HClusterAmplitudeMultiplicityEq1: hbook.NewH1D(100, 0, 15000),
		HClusterAmplitudeMultiplicityEq2: hbook.NewH1D(100, 0, 15000),
	}
	for i := 0; i < N; i++ {
		dqp.HCharge[i] = *hbook.NewH1D(100, -2e4, 100e3)
		dqp.HAmplitude[i] = *hbook.NewH1D(100, 0, 4200)
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
	cluster := &event.Clusters[0]
	d.Nevents++
	d.HSRout.Fill(float64(cluster.SRout()), 1)
	d.HClusterCharge.Fill(float64(cluster.Charge()), 1)
	d.HClusterAmplitude.Fill(float64(cluster.Amplitude()), 1)

	multi := len(cluster.PulsesWithSignal())
	satmulti := len(cluster.PulsesWithSatSignal())
	d.HMultiplicity.Fill(float64(multi), 1)
	d.HSatMultiplicity.Fill(float64(satmulti), 1)
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
		if pulse.HasSignal {
			d.HFrequency.Fill(float64(j), 1)
		}
		if pulse.HasSatSignal {
			d.HSatFrequency.Fill(float64(j), 1)
		}
	}
}

func (d *DQPlot) Finalize() {
	const N = 4
	for i := 0; i < N; i++ {
		d.HCharge[i].Scale(1 / d.HCharge[i].Integral())
		d.HAmplitude[i].Scale(1 / d.HAmplitude[i].Integral())
	}
	d.HFrequency.Scale(1 / float64(d.Nevents))
	d.HSatFrequency.Scale(1 / float64(d.Nevents))
	d.HSRout.Scale(1 / d.HSRout.Integral())
	d.HMultiplicity.Scale(1 / d.HMultiplicity.Integral())
	d.HSatMultiplicity.Scale(1 / d.HSatMultiplicity.Integral())
	//d.HClusterCharge.Scale(1 / d.HClusterCharge.Integral())
	//d.HClusterChargeMultiplicityEq1.Scale(1 / d.HClusterChargeMultiplicityEq1.Integral())
	//d.HClusterChargeMultiplicityEq2.Scale(1 / d.HClusterChargeMultiplicityEq2.Integral())
	//d.HClusterAmplitude.Scale(1 / d.HClusterAmplitude.Integral())
	//d.HClusterAmplitudeMultiplicityEq1.Scale(1 / d.HClusterAmplitudeMultiplicityEq1.Integral())
	//d.HClusterAmplitudeMultiplicityEq2.Scale(1 / d.HClusterAmplitudeMultiplicityEq2.Integral())
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

	//doplot("Charge", "Entries", "output/distribCharge.png", utils.H1dToHplot(linestyle, d.HCharge...)...) //, utils.H1dToHplot(linestyleref, dqplotref.HCharge...)...)
	doplot("Charge",
		"Entries",
		"output/distribCharge.png",
		append(utils.H1dToHplot(linestyle, d.HCharge...),
			utils.H1dToHplot(linestyleref, dqplotref.HCharge...)...)...)
	doplot("Amplitude",
		"Entries",
		"output/distribAmplitude.png",
		append(utils.H1dToHplot(linestyle, d.HAmplitude...),
			utils.H1dToHplot(linestyleref, dqplotref.HAmplitude...)...)...)
	doplot("Channel",
		"# pulses / cluster",
		"output/distribFrequency.png",
		append(utils.H1dptrToHplot(linestyle, d.HFrequency),
			utils.H1dptrToHplot(linestyleref, dqplotref.HFrequency)...)...)
	doplot("Channel",
		"# pulses with saturation / cluster",
		"output/distribSatFrequency.png",
		append(utils.H1dptrToHplot(linestyle, d.HSatFrequency),
			utils.H1dptrToHplot(linestyleref, dqplotref.HSatFrequency)...)...)
	doplot("SRout",
		"Entries",
		"output/distribSRout.png",
		append(utils.H1dptrToHplot(linestyle, d.HSRout),
			utils.H1dptrToHplot(linestyleref, dqplotref.HSRout)...)...)
	doplot("Multiplicity",
		"Entries",
		"output/distribMultiplicity.png",
		append(utils.H1dptrToHplot(linestyle, d.HMultiplicity),
			utils.H1dptrToHplot(linestyleref, dqplotref.HMultiplicity)...)...)
	doplot("Multiplicity of saturating pulses",
		"Entries",
		"output/distribSatMultiplicity.png",
		append(utils.H1dptrToHplot(linestyle, d.HSatMultiplicity),
			utils.H1dptrToHplot(linestyleref, dqplotref.HSatMultiplicity)...)...)
	doplot("Cluster charge",
		"Entries",
		"output/distribClusterCharge.png",
		append(utils.H1dptrToHplot(linestyle, d.HClusterCharge, d.HClusterChargeMultiplicityEq1, d.HClusterChargeMultiplicityEq2),
			utils.H1dptrToHplot(linestyleref, dqplotref.HClusterCharge, dqplotref.HClusterChargeMultiplicityEq1, dqplotref.HClusterChargeMultiplicityEq2)...)...)
	doplot("Cluster amplitude",
		"Entries",
		"output/distribClusterAmplitude.png",
		append(utils.H1dptrToHplot(linestyle, d.HClusterAmplitude, d.HClusterAmplitudeMultiplicityEq1, d.HClusterAmplitudeMultiplicityEq2),
			utils.H1dptrToHplot(linestyleref, dqplotref.HClusterAmplitude, dqplotref.HClusterAmplitudeMultiplicityEq1, dqplotref.HClusterAmplitudeMultiplicityEq2)...)...)
}

func (d *DQPlot) WriteGob(fileName string) error {
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := gob.NewEncoder(f)
	fmt.Printf("Saving DQplot object to %s\n", fileName)
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
