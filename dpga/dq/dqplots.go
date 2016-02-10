package dq

import (
	"github.com/go-hep/hbook"
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
}

func (d *DQPlot) WriteHistosToFile() {
	doplot := utils.MakeHPlot
	// 	doplot := utils.MakeGonumPlot
	doplot("Channel", "# pulses / event", "output/distribFrequency.png", *d.HFrequency)
	doplot("Channel", "# pulses with saturation / event", "output/distribSatFrequency.png", *d.HSatFrequency)
	doplot("Multiplicity", "Entries", "output/distribMultiplicity.png", *d.HMultiplicity)
	doplot("Multiplicity of pulses with saturation", "Entries", "output/distribSatMultiplicity.png", *d.HSatMultiplicity)
}
