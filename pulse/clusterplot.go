package pulse

import (
	"github.com/go-hep/hbook"
	"gitlab.in2p3.fr/avirm/analysis-go/utils"
)

type ClusterPlot struct {
	Nclusters                        uint
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

func NewClusterPlot() *ClusterPlot {
	const N = 4
	cp := &ClusterPlot{
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
		cp.HCharge[i] = *hbook.NewH1D(100, -2e4, 100e3)
		cp.HAmplitude[i] = *hbook.NewH1D(100, 0, 4200)
	}

	return cp
}

func (c *ClusterPlot) FillHistos(cluster *Cluster) {
	c.Nclusters++
	c.HSRout.Fill(float64(cluster.SRout()), 1)
	c.HClusterCharge.Fill(float64(cluster.Charge()), 1)
	c.HClusterAmplitude.Fill(float64(cluster.Amplitude()), 1)

	multi := len(cluster.PulsesWithSignal())
	c.HMultiplicity.Fill(float64(multi), 1)
	switch multi {
	case 1:
		c.HClusterChargeMultiplicityEq1.Fill(float64(cluster.Charge()), 1)
		c.HClusterAmplitudeMultiplicityEq1.Fill(float64(cluster.Amplitude()), 1)
	case 2:
		c.HClusterChargeMultiplicityEq2.Fill(float64(cluster.Charge()), 1)
		c.HClusterAmplitudeMultiplicityEq2.Fill(float64(cluster.Amplitude()), 1)
	}

	for j := range cluster.Pulses {
		pulse := &cluster.Pulses[j]
		c.HCharge[j].Fill(float64(pulse.Charge()), 1)
		c.HAmplitude[j].Fill(float64(pulse.Amplitude()), 1)
		if pulse.HasSignal {
			c.HFrequency.Fill(float64(j), 1)
			c.HFrequencyTot.Fill(1, 1)
		}
		if pulse.HasSatSignal {
			c.HSatFrequency.Fill(float64(j), 1)
			c.HSatFrequencyTot.Fill(1, 1)
		}
	}
}

func (c *ClusterPlot) Finalize() {
	c.HFrequency.Scale(1 / float64(c.Nclusters))
	c.HSatFrequency.Scale(1 / float64(c.Nclusters))
	c.HFrequencyTot.Scale(1 / float64(c.Nclusters))
	c.HSatFrequencyTot.Scale(1 / float64(c.Nclusters))
}

func (c *ClusterPlot) WriteHistosToFile() {
	doplot := utils.MakeHPlot
	// 	doplot := utils.MakeGonumPlot
	doplot("Charge", "Entries", "output/distribCharge.png", c.HCharge...)
	doplot("Amplitude", "Entries", "output/distribAmplitude.png", c.HAmplitude...)
	doplot("Channel", "# pulses / cluster", "output/distribFrequency.png", *c.HFrequency, *c.HFrequencyTot)
	doplot("Channel", "# pulses with saturation / cluster", "output/distribSatFrequency.png", *c.HSatFrequency, *c.HSatFrequencyTot)
	doplot("SRout", "Entries", "output/distribSRout.png", *c.HSRout)
	doplot("Multiplicity", "Entries", "output/distribMultiplicity.png", *c.HMultiplicity)
	doplot("Cluster charge", "Entries", "output/distribClusterCharge.png", *c.HClusterCharge, *c.HClusterChargeMultiplicityEq1, *c.HClusterChargeMultiplicityEq2)
	doplot("Cluster amplitude", "Entries", "output/distribClusterAmplitude.png", *c.HClusterAmplitude, *c.HClusterAmplitudeMultiplicityEq1, *c.HClusterAmplitudeMultiplicityEq2)
}
