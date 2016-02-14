package dq

import (
	"encoding/gob"
	"fmt"
	"os"

	"github.com/go-hep/hbook"
	"github.com/gonum/plot/vg"
	"github.com/gonum/plot/vg/draw"
	"gitlab.in2p3.fr/avirm/analysis-go/testbench/event"
	"gitlab.in2p3.fr/avirm/analysis-go/utils"
)

type DQPlot struct {
	Nevents                          uint
	HCharge                          []hbook.H1D
	HAmplitude                       []hbook.H1D
	HFrequency                       *hbook.H1D
	HSatFrequency                    *hbook.H1D
	HFrequencyTot                    *hbook.H1D
	HSatFrequencyTot                 *hbook.H1D
	HSRout                           *hbook.H1D
	HMultiplicity                    *hbook.H1D
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
		HFrequencyTot:                    hbook.NewH1D(1, 0, 4),
		HSatFrequencyTot:                 hbook.NewH1D(1, 0, 4),
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
		dqp.HCharge[i] = *hbook.NewH1D(100, -2e4, 100e3)
		dqp.HAmplitude[i] = *hbook.NewH1D(100, 0, 4200)
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
	cluster := &event.Cluster
	d.Nevents++
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
		if pulse.HasSignal {
			d.HFrequency.Fill(float64(j), 1)
			d.HFrequencyTot.Fill(1, 1)
		}
		if pulse.HasSatSignal {
			d.HSatFrequency.Fill(float64(j), 1)
			d.HSatFrequencyTot.Fill(1, 1)
		}
	}
}

func (d *DQPlot) Finalize() {
	d.HFrequency.Scale(1 / float64(d.Nevents))
	d.HSatFrequency.Scale(1 / float64(d.Nevents))
	d.HFrequencyTot.Scale(1 / float64(d.Nevents))
	d.HSatFrequencyTot.Scale(1 / float64(d.Nevents))
}

func (d *DQPlot) WriteHistosToFile(suffix string) {
	doplot := utils.MakeHPl
	// 	doplot := utils.MakeGonumPlot
	doplot("Charge", "Entries", "output/distribCharge"+suffix+".png", utils.H1dToHplot(draw.LineStyle{Width: vg.Points(10)}, d.HCharge...)...)
	/*doplot("Amplitude", "Entries", "output/distribAmplitude"+suffix+".png", d.HAmplitude...)
	doplot("Channel", "# pulses / cluster", "output/distribFrequency"+suffix+".png", *d.HFrequency, *d.HFrequencyTot)
	doplot("Channel", "# pulses with saturation / cluster", "output/distribSatFrequency"+suffix+".png", *d.HSatFrequency, *d.HSatFrequencyTot)
	doplot("SRout", "Entries", "output/distribSRout"+suffix+".png", *d.HSRout)
	doplot("Multiplicity", "Entries", "output/distribMultiplicity"+suffix+".png", *d.HMultiplicity)
	doplot("Cluster charge", "Entries", "output/distribClusterCharge"+suffix+".png", *d.HClusterCharge, *d.HClusterChargeMultiplicityEq1, *d.HClusterChargeMultiplicityEq2)
	doplot("Cluster amplitude", "Entries", "output/distribClusterAmplitude"+suffix+".png", *d.HClusterAmplitude, *d.HClusterAmplitudeMultiplicityEq1, *d.HClusterAmplitudeMultiplicityEq2)*/
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
