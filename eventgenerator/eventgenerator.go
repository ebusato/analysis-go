package main

import (
	"math/rand"

	"github.com/gonum/stat/distuv"
)

type source struct {
	Name string
	Activity float64
	N0 float64
}

type process struct {
	Name string
	Rate float64
}

type event struct {
	Proc *process
	T    float64 // time
}

func newEvent(proc *process) *event {
	dist := distuv.Exponential{
		Rate:   proc.Rate,
		Source: rand.New(rand.NewSource(0)),
	}

	var e *event
	e.Proc = proc
	e.T = dist.Rand()
	return e

	/*
		hist := hbook.NewH1D(20, 0, 20)
		for i := 0; i < npoints; i++ {
			v := dist.Rand()
			hist.Fill(v, 1)
		}

		// Make a plot and set its title.
		p, err := hplot.New()
		if err != nil {
			log.Fatalf("error: %v\n", err)
		}
		p.Title.Text = "Histogram"
		p.X.Label.Text = "X"
		p.Y.Label.Text = "Y"

		// Create a histogram of our values drawn
		// from the standard normal.
		h, err := hplot.NewH1D(hist)
		if err != nil {
			log.Fatal(err)
		}
		h.Infos.Style = hplot.HInfoSummary
		p.Add(h)

		// Save the plot to a PNG file.
		if err := p.Save(6*vg.Inch, -1, "h1d_plot.png"); err != nil {
			log.Fatalf("error saving plot: %v\n", err)
		}
	*/
}

func main() {
	const npoints = 10000
	
	var na22_2MBq
	
	for i := 0; i < npoints; i++ {
		newEvent(nil)
	}
}
