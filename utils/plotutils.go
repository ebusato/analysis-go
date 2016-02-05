package utils

import (
	"github.com/go-hep/hbook"
	"github.com/go-hep/hplot"
	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/plotutil"
	"github.com/gonum/plot/vg"
)

func H1DToGonum(histo ...hbook.H1D) []plotter.Histogram {
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

func H1dToHplot(histo ...hbook.H1D) []hplot.Histogram {
	output := make([]hplot.Histogram, len(histo))
	for i, h := range histo {
		// 		hi, err := hplot.NewHistogram(&h, h.Axis().Bins())
		// 		if err != nil {
		// 			panic(err)
		// 		}
		hi, err := hplot.NewH1D(&h)
		if err != nil {
			panic(err)
		}
		hi.FillColor = nil //plotutil.Color(i)
		hi.Color = plotutil.Color(i)
		output[i] = *hi
	}
	return output
}

func MakeHPlot(xTitle string, yTitle string, outFile string, normalize bool, histo ...hbook.H1D) {
	p, err := hplot.New()
	if err != nil {
		panic(err)
	}
	p.X.Label.Text = xTitle
	p.Y.Label.Text = yTitle

	p.Y.Min = 0

	hHplot := H1dToHplot(histo...)
	for i := range hHplot {
		p.Add(&hHplot[i])
	}
	// 	p.Add(hHplot)

	if err := p.Save(4*vg.Inch, 4*vg.Inch, outFile); err != nil {
		panic(err)
	}
}

func MakeGonumPlot(xTitle string, yTitle string, outFile string, normalize bool, histo ...hbook.H1D) {
	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	p.X.Label.Text = xTitle
	p.Y.Label.Text = yTitle

	hGonum := H1DToGonum(histo...)
	for i := range hGonum {
		if normalize {
			hGonum[i].Normalize(1)
		}
		p.Add(&hGonum[i])
	}

	if err := p.Save(4*vg.Inch, 4*vg.Inch, outFile); err != nil {
		panic(err)
	}
}
