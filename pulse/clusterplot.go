package pulse

import (
	"github.com/go-hep/hbook"
	"gitlab.in2p3.fr/avirm/analysis-go/utils"
)

type ClusterPlot struct {
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
		hasSig := 0
		switch pulse.HasSignal {
		case true:
			hasSig = 1
			c.HFrequency.Fill(float64(j), 1)
		case false:
			hasSig = 0
		}
		c.HHasSignal[j].Fill(float64(hasSig), 1)
		hasSatSig := 0
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

func (c *ClusterPlot) WriteHistosToFile() {
	utils.MakeGonumPlot("Charge", "Entries", "output/distribCharge.png", false, c.HCharge...)
	utils.MakeGonumPlot("Amplitude", "Entries", "output/distribAmplitude.png", false, c.HAmplitude...)
	utils.MakeGonumPlot("HasSignal", "Entries", "output/distribHasSignal.png", false, c.HHasSignal...)
	utils.MakeGonumPlot("HasSatSignal", "Entries", "output/distribHasSatSignal.png", false, c.HHasSatSignal...)
	utils.MakeGonumPlot("Channel", "Frequency", "output/distribFrequency.png", true, *c.HFrequency)
	utils.MakeGonumPlot("Channel", "Saturation frequency", "output/distribSatFrequency.png", true, *c.HSatFrequency)
	utils.MakeGonumPlot("SRout", "Entries", "output/distribSRout.png", false, *c.HSRout)
	utils.MakeGonumPlot("Multiplicity", "Entries", "output/distribMultiplicity.png", false, *c.HMultiplicity)
	utils.MakeGonumPlot("Cluster charge", "Entries", "output/distribClusterCharge.png", false, *c.HClusterCharge)
	utils.MakeGonumPlot("Cluster charge (multiplicity = 1)", "Entries", "output/distribClusterChargeMultiplicityEq1.png", false, *c.HClusterChargeMultiplicityEq1)
	utils.MakeGonumPlot("Cluster charge (multiplicity = 2)", "Entries", "output/distribClusterChargeMultiplicityEq2.png", false, *c.HClusterChargeMultiplicityEq2)
	utils.MakeGonumPlot("Cluster amplitude", "Entries", "output/distribClusterAmplitude.png", false, *c.HClusterAmplitude)
	utils.MakeGonumPlot("Cluster amplitude (multiplicity = 1)", "Entries", "output/distribClusterAmplitudeMultiplicityEq1.png", false, *c.HClusterAmplitudeMultiplicityEq1)
	utils.MakeGonumPlot("Cluster amplitude (multiplicity = 2)", "Entries", "output/distribClusterAmplitudeMultiplicityEq2.png", false, *c.HClusterAmplitudeMultiplicityEq2)
}
