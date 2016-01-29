package event

import (
	"fmt"
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
		fmt.Printf("histo: %+v\n", h)
		h.FillColor = nil //plotutil.Color(i)
		h.Color = plotutil.Color(i)
		output[i] = *h
	}
	return output
}

func PrintHbookH1D(h hbook.H1D) {
	for i := 0; i < h.Len(); i++ {
		x, y := h.XY(i)
		fmt.Printf("%v %v\n", x, y)
	}
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
	hHasSignal := make([]hbook.H1D, N)
	hSRout := *hbook.NewH1D(1024, 0, 1023)

	for i := 0; i < N; i++ {
		hCharge[i] = *hbook.NewH1D(100, -100e3, 600e3)
		hHasSignal[i] = *hbook.NewH1D(2, 0, 2)
	}
	for _, event := range *d {
		for i := 0; i < N; i++ {
			pulse := event.Cluster.Pulses[i]
			hCharge[i].Fill(float64(pulse.Charge()), 1)
			hasSig := 0.
			switch pulse.HasSignal {
			case true:
				hasSig = 1
			case false:
				hasSig = 0
			}
			hHasSignal[i].Fill(hasSig, 1)
			if i == 0 {
				hSRout.Fill(float64(pulse.SRout), 1)
			}
		}
	}

	PrintHbookH1D(hHasSignal[0])
	MakePlot("Charge", "Entries (A. U.)", "output/distribCharge.png", hCharge...)
	MakePlot("HasSignal", "Entries (A. U.)", "output/distribHasSignal.png", hHasSignal...)
	MakePlot("SRout", "Entries (A. U.)", "output/distribSRout.png", hSRout)
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

func (d *Data) Plot() {
	d.PlotDistribs()
	d.PlotPulses(pulse.XaxisTime, false, true)
}
