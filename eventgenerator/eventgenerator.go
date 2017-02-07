package main

import (
	"fmt"
	"math"
	"math/rand"

	"github.com/gonum/stat/distuv"
)

type Isotope struct {
	Name string
	T    float64 // half life in seconds
}

// return value is in s-1
func (i *Isotope) Lambda() float64 {
	return math.Ln2 / i.T
}

type Source struct {
	Isotope   Isotope
	Activity0 float64 // initial activity in Bq
	T0        float64 // initial time in seconds
}

// t is in seconds
func (s *Source) Activity(t float64) float64 {
	return s.Activity0 * math.Exp(-1*s.Isotope.Lambda()*(t-s.T0))
}

// t is in seconds
func (s *Source) NoNuclei(t float64) float64 {
	return s.Activity(t) / s.Isotope.Lambda()
}

type Event struct {
	Source *Source
	Time   float64
}

func NewEvent(s *Source) *Event {
	dist := distuv.Exponential{
		Rate:   s.Isotope.Lambda(),
		Source: rand.New(rand.NewSource(99)),
	}

	e := &Event{Source: s, Time: dist.Rand()}
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
	// Set initial and final run time
	const ti = 0   // seconds
	const tf = 200 // seconds

	na22 := Isotope{Name: "Na22", T: 8.2e7}
	na22_2MBq := &Source{Isotope: na22, Activity0: 2.69e6, T0: -1.077e8}

	fmt.Println(" act =", na22_2MBq.Activity(ti), na22_2MBq.NoNuclei(ti))

	na22_2MBq_noDecays := na22_2MBq.NoNuclei(ti) - na22_2MBq.NoNuclei(tf)
	fmt.Println(" na22_2MBq_noDecays=", na22_2MBq_noDecays)

	for i := int64(0); i < 10; i++ {
		evt_Na22_2MBq := NewEvent(na22_2MBq)
		fmt.Println("t =", evt_Na22_2MBq.Time)
	}
}
