package event

import (
	"log"
	"os/exec"

	"github.com/go-hep/hbook"
	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/plotutil"
	"github.com/gonum/plot/vg"
	"gitlab.in2p3.fr/AVIRM/Analysis-go/pulse"
)

type Data []Event

func (d *Data) CheckIntegrity() {
	if len(*d) == 0 {
		panic("data has 0 events")
	}
	noPulses := (*d)[0].NoPulses()
	for i := 1; i < len(*d); i++ {
		no := (*d)[i].NoPulses()
		if no != noPulses {
			panic("not all events have the same number of pulses")
		}
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

func (d *Data) PlotDistribs() {
	const N = 4
	hCharge := make([]hbook.H1D, N)
	hAmplitude := make([]hbook.H1D, N)
	hHasSignal := make([]hbook.H1D, N)
	hSRout := *hbook.NewH1D(1024, 0, 1023)
	hMultiplicity := *hbook.NewH1D(5, 0, 5)
	hClusterCharge := *hbook.NewH1D(100, -2e4, 400e3)
	hClusterChargeMultiplicityEq1 := *hbook.NewH1D(100, -2e4, 400e3)
	hClusterChargeMultiplicityEq2 := *hbook.NewH1D(100, -2e4, 400e3)
	hClusterAmplitude := *hbook.NewH1D(100, 0, 15000)
	hClusterAmplitudeMultiplicityEq1 := *hbook.NewH1D(100, 0, 15000)
	hClusterAmplitudeMultiplicityEq2 := *hbook.NewH1D(100, 0, 15000)

	for i := 0; i < N; i++ {
		hCharge[i] = *hbook.NewH1D(100, -2e4, 100e3)
		hAmplitude[i] = *hbook.NewH1D(100, 0, 4200)
		hHasSignal[i] = *hbook.NewH1D(2, 0, 2)
	}
	for i := range *d {
		cluster := &(*d)[i].Cluster
		hSRout.Fill(float64(cluster.SRout()), 1)
		hClusterCharge.Fill(float64(cluster.Charge()), 1)
		hClusterAmplitude.Fill(float64(cluster.Amplitude()), 1)

		multi := len(cluster.PulsesWithSignal())
		hMultiplicity.Fill(float64(multi), 1)
		switch multi {
		case 1:
			hClusterChargeMultiplicityEq1.Fill(float64(cluster.Charge()), 1)
			hClusterAmplitudeMultiplicityEq1.Fill(float64(cluster.Amplitude()), 1)
		case 2:
			hClusterChargeMultiplicityEq2.Fill(float64(cluster.Charge()), 1)
			hClusterAmplitudeMultiplicityEq2.Fill(float64(cluster.Amplitude()), 1)
		}

		for j := range cluster.Pulses {
			pulse := &cluster.Pulses[j]
			hCharge[j].Fill(float64(pulse.Charge()), 1)
			hAmplitude[j].Fill(float64(pulse.Amplitude()), 1)
			hasSig := 0.
			switch pulse.HasSignal {
			case true:
				hasSig = 1
			case false:
				hasSig = 0
			}
			hHasSignal[j].Fill(hasSig, 1)
		}
	}

	MakePlot("Charge", "Entries (A. U.)", "output/distribCharge.png", hCharge...)
	MakePlot("Amplitude", "Entries (A. U.)", "output/distribAmplitude.png", hAmplitude...)
	MakePlot("HasSignal", "Entries (A. U.)", "output/distribHasSignal.png", hHasSignal...)
	MakePlot("SRout", "Entries (A. U.)", "output/distribSRout.png", hSRout)
	MakePlot("Multiplicity", "Entries (A. U.)", "output/distribMultiplicity.png", hMultiplicity)
	MakePlot("Cluster charge", "Entries (A. U.)", "output/distribClusterCharge.png", hClusterCharge)
	MakePlot("Cluster charge (multiplicity = 1)", "Entries (A. U.)", "output/distribClusterChargeMultiplicityEq1.png", hClusterChargeMultiplicityEq1)
	MakePlot("Cluster charge (multiplicity = 2)", "Entries (A. U.)", "output/distribClusterChargeMultiplicityEq2.png", hClusterChargeMultiplicityEq2)
	MakePlot("Cluster amplitude", "Entries (A. U.)", "output/distribClusterAmplitude.png", hClusterAmplitude)
	MakePlot("Cluster amplitude (multiplicity = 1)", "Entries (A. U.)", "output/distribClusterAmplitudeMultiplicityEq1.png", hClusterAmplitudeMultiplicityEq1)
	MakePlot("Cluster amplitude (multiplicity = 2)", "Entries (A. U.)", "output/distribClusterAmplitudeMultiplicityEq2.png", hClusterAmplitudeMultiplicityEq2)
}

func (d *Data) PlotPulses(xaxis pulse.XaxisType, pedestalRange bool, savePulses bool) {
	var gsOptions = []string{"-dNOPAUSE", "-dBATCH", "-sDEVICE=pdfwrite", "-sOutputFile=output/merged.pdf"}
	var outPulseFiles []string

	for i, event := range *d {
		outPulseFiles = append(outPulseFiles, event.PlotPulses(xaxis, pedestalRange))
		if i >= 20 {
			break
		}
	}

	gsOptions = append(gsOptions, outPulseFiles...)
	err := exec.Command("gs", gsOptions...).Run()
	if err != nil {
		log.Fatal("error merging files", err)
	}
	if !savePulses {
		for _, fileName := range outPulseFiles {
			err := exec.Command("rm", "-f", fileName).Run()
			if err != nil {
				log.Fatal("error removing file", err)
			}
		}
	}
}

func (d *Data) AmplitudeCorrelationWithinCluster() plotter.XYZs {
	var data plotter.XYZs
	for i := range *d {
		cluster := (*d)[i].Cluster
		pulses := cluster.PulsesWithSignal()
		multiplicity := len(pulses)
		if multiplicity == 2 {
			mydata := struct {
				X, Y, Z float64
			}{
				X: pulses[0].Amplitude(),
				Y: pulses[1].Amplitude(),
				Z: 1,
			}
			data = append(data, mydata)
		}
	}
	return data
}

func (d *Data) Plot() {
	d.PlotDistribs()
	d.PlotPulses(pulse.XaxisTime, false, true)
}
