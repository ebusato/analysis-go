// Package utils defines a few utilities for plotting.
package utils

import (
	"bytes"

	"go-hep.org/x/hep/hbook"
	"go-hep.org/x/hep/hplot"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
	"gonum.org/v1/plot/vg/vgsvg"
)

// H1DToGonum converts hbook.H1D objects to plotter.Histogram objects.
func H1DToGonum(histo ...hbook.H1D) []plotter.Histogram {
	output := make([]plotter.Histogram, len(histo))
	for i, h := range histo {
		h, err := plotter.NewHistogram(&h, len(h.Binning().Bins()))
		if err != nil {
			panic(err)
		}
		h.FillColor = nil //plotutil.Color(i)
		h.Color = plotutil.Color(i)
		output[i] = *h
	}
	return output
}

// H1dToHplot converts hbook.H1D objects to hplot.Histogram objects.
func H1dToHplot(lineStyle draw.LineStyle, histo ...hbook.H1D) []hplot.H1D {
	output := make([]hplot.H1D, len(histo))
	for i := range histo {
		hi := hplot.NewH1D(&histo[i])
		if len(histo) == 1 {
			hi.Infos.Style = hplot.HInfoSummary
		}
		hi.LineStyle = lineStyle
		hi.FillColor = nil //plotutil.Color(i)
		hi.Color = plotutil.Color(i)
		output[i] = *hi
	}
	return output
}

// H1dptrToHplot converts hbook.H1D objects to hplot.Histogram objects.
func H1dptrToHplot(lineStyle draw.LineStyle, histos ...*hbook.H1D) []hplot.H1D {
	var output []hplot.H1D
	for i := range histos {
		histo := histos[i]
		if histo != nil {
			hi := hplot.NewH1D(histo)
			hi.Infos.Style = hplot.HInfoSummary

			hi.LineStyle = lineStyle
			hi.FillColor = nil //plotutil.Color(i)
			hi.Color = plotutil.Color(i)
			output = append(output, *hi)
		}
	}
	return output
}

// MakeHPlot makes a hplot with hbook.H1D objects.
func MakeHPlot(xTitle string, yTitle string, outFile string, histo ...hbook.H1D) {
	p := hplot.New()
	p.X.Label.Text = xTitle
	p.Y.Label.Text = yTitle

	p.Y.Min = 0

	hHplot := H1dToHplot(draw.LineStyle{}, histo...)
	for i := range hHplot {
		p.Add(&hHplot[i])
	}

	if err := p.Save(4*vg.Inch, 4*vg.Inch, outFile); err != nil {
		panic(err)
	}
}

// MakeHPl makes a hplot with hplot.Histogram objects.
func MakeHPl(xTitle string, yTitle string, outFile string, hHplot ...hplot.H1D) {
	p := hplot.New()
	p.X.Label.Text = xTitle
	p.Y.Label.Text = yTitle

	p.Y.Min = 0

	for i := range hHplot {
		p.Add(&hHplot[i])
	}

	if err := p.Save(4*vg.Inch, 4*vg.Inch, outFile); err != nil {
		panic(err)
	}
}

// MakeGonumPlot makes a plot with hbook.H1D objects.
func MakeGonumPlot(xTitle string, yTitle string, outFile string, histo ...hbook.H1D) {
	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	p.X.Label.Text = xTitle
	p.Y.Label.Text = yTitle

	hGonum := H1DToGonum(histo...)
	for i := range hGonum {
		p.Add(&hGonum[i])
	}

	if err := p.Save(4*vg.Inch, 4*vg.Inch, outFile); err != nil {
		panic(err)
	}
}

type Drawer interface {
	Draw(c draw.Canvas)
}

// RenderSVG takes a gonum plot and returns string encoding svg graphics.
func RenderSVG(d Drawer, w vg.Length, h vg.Length) string {
	width := w * vg.Centimeter
	height := h * vg.Centimeter
	//canvas := vgsvg.New(size, size/vg.Length(math.Phi))
	canvas := vgsvg.New(width, height)
	d.Draw(draw.New(canvas))
	out := new(bytes.Buffer)
	_, err := canvas.WriteTo(out)
	if err != nil {
		panic(err)
	}
	return string(out.Bytes())
}
