package pulse

import (
	"github.com/go-hep/hbook"
	"gitlab.in2p3.fr/avirm/analysis-go/utils"
)

type ClusterPlot struct {
	Nclusters                        uint
	HCharge                          []hbook.H1D
	HAmplitude                       []hbook.H1D
	HHasSignal                       []hbook.H1D
	HHasSatSignal                    []hbook.H1D
	HFrequency                       *hbook.H1D
	HSatFrequency                    *hbook.H1D
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
		HHasSignal:                       make([]hbook.H1D, N),
		HHasSatSignal:                    make([]hbook.H1D, N),
		HFrequency:                       hbook.NewH1D(4, 0, 4),
		HSatFrequency:                    hbook.NewH1D(4, 0, 4),
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
		cp.HHasSignal[i] = *hbook.NewH1D(2, 0, 2)
		cp.HHasSatSignal[i] = *hbook.NewH1D(2, 0, 2)
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
		hasSig, hasSatSig := 0, 0
		switch pulse.HasSignal {
		case true:
			hasSig = 1
			c.HFrequency.Fill(float64(j), 1)
		case false:
			hasSig = 0
		}
		c.HHasSignal[j].Fill(float64(hasSig), 1)
		switch pulse.HasSatSignal {
		case true:
			hasSatSig = 1
			c.HSatFrequency.Fill(float64(j), 1)
		case false:
			hasSatSig = 0
		}
		c.HHasSatSignal[j].Fill(float64(hasSatSig), 1)
	}
}

func (c *ClusterPlot) Finalize() {
	c.HFrequency.Scale(1 / float64(c.Nclusters))
	c.HSatFrequency.Scale(1 / float64(c.Nclusters))
	N := 4
	for i := 0; i < N; i++ {
		c.HHasSignal[i].Scale(1 / c.HHasSignal[i].Integral())
		c.HHasSatSignal[i].Scale(1 / c.HHasSatSignal[i].Integral())
	}
}

func (c *ClusterPlot) WriteHistosToFile() {
	doplot := utils.MakeHPlot
	// 	doplot := utils.MakeGonumPlot
	doplot("Charge", "Entries", "output/distribCharge.png", c.HCharge...)
	doplot("Amplitude", "Entries", "output/distribAmplitude.png", c.HAmplitude...)
	doplot("HasSignal", "Entries", "output/distribHasSignal.png", c.HHasSignal...)
	doplot("HasSatSignal", "Entries", "output/distribHasSatSignal.png", c.HHasSatSignal...)
	doplot("Channel", "Frequency", "output/distribFrequency.png", *c.HFrequency)
	doplot("Channel", "Saturation frequency", "output/distribSatFrequency.png", *c.HSatFrequency)
	doplot("SRout", "Entries", "output/distribSRout.png", *c.HSRout)
	doplot("Multiplicity", "Entries", "output/distribMultiplicity.png", *c.HMultiplicity)
	doplot("Cluster charge", "Entries", "output/distribClusterCharge.png", *c.HClusterCharge)
	doplot("Cluster charge (multiplicity = 1)", "Entries", "output/distribClusterChargeMultiplicityEq1.png", *c.HClusterChargeMultiplicityEq1)
	doplot("Cluster charge (multiplicity = 2)", "Entries", "output/distribClusterChargeMultiplicityEq2.png", *c.HClusterChargeMultiplicityEq2)
	doplot("Cluster amplitude", "Entries", "output/distribClusterAmplitude.png", *c.HClusterAmplitude)
	doplot("Cluster amplitude (multiplicity = 1)", "Entries", "output/distribClusterAmplitudeMultiplicityEq1.png", *c.HClusterAmplitudeMultiplicityEq1)
	doplot("Cluster amplitude (multiplicity = 2)", "Entries", "output/distribClusterAmplitudeMultiplicityEq2.png", *c.HClusterAmplitudeMultiplicityEq2)

}
