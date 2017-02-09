package main

import (
	"fmt"
	"log"
	"math"

	"github.com/go-hep/hbook"
	"github.com/go-hep/hplot"
	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/plotutil"
	"github.com/gonum/plot/vg"
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
	Name      string  // string with rounded activity
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

// Time is absolute (not relative to previous event)
type Event struct {
	Source *Source
	Time   float64
}

func NewEvent(s *Source, t float64) Event {
	dist := distuv.Exponential{
		Rate:   s.Activity(t),
		Source: nil, //rand.New(rand.NewSource(0)),
	}

	e := Event{Time: dist.Rand() + t}
	return e
}

type EventColl struct {
	Source *Source
	Events []Event
}

func NewEventColl(s *Source, ti float64, nEvents int) *EventColl {
	evtColl := &EventColl{Source: s}
	time := ti
	evtColl.Events = make([]Event, nEvents)
	for i := 0; i < nEvents; i++ {
		evtColl.Events[i].Source = s
		evtColl.Events[i] = NewEvent(s, time)
		//fmt.Println(time, events_Na22_2MBq[i].Time, events_Na22_2MBq[i].Time-time)
		time = evtColl.Events[i].Time
	}
	return evtColl
}

func (ec *EventColl) DeadTimeLoss(dt float64, paralizable bool) int {
	var nEventsLost int
	evtGeneratingDeadTime := &ec.Events[0] // the first event is generating the first dead time
	for i := 1; i < len(ec.Events); i++ {
		switch ec.Events[i].Time-evtGeneratingDeadTime.Time < dt {
		case true:
			nEventsLost++
			if paralizable {
				evtGeneratingDeadTime = &ec.Events[i]
			}
		case false:
			evtGeneratingDeadTime = &ec.Events[i]
		}
	}
	return nEventsLost
}

func (ec *EventColl) Len() int {
	return len(ec.Events)
}

func (ec *EventColl) XY(i int) (x, y float64) {
	x = float64(i)
	y = ec.Events[i].Time
	return
}

func (ec *EventColl) PlotTimeVsEvtIndex(name string) {
	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	p.Title.Text = "Time vs event index"
	p.X.Label.Text = "event index"
	p.Y.Label.Text = "Time (s)"
	p.Add(plotter.NewGrid())

	err = plotutil.AddLinePoints(p, "", ec)

	if err != nil {
		panic(err)
	}

	if err := p.Save(14*vg.Inch, 5*vg.Inch, name); err != nil {
		panic(err)
	}
}

func (ec *EventColl) PlotRelTimeHist(name string) {
	hRelTime := hbook.NewH1D(200, 0, 0.00001)
	for i := 1; i < len(ec.Events); i++ {
		hRelTime.Fill(ec.Events[i].Time-ec.Events[i-1].Time, 1)
	}

	p, err := hplot.New()
	if err != nil {
		log.Fatalf("error: %v\n", err)
	}
	p.Title.Text = "Histogram"
	p.X.Label.Text = "relative time (s)"
	p.Y.Label.Text = "A. U. "

	h, err := hplot.NewH1D(hRelTime)
	if err != nil {
		log.Fatal(err)
	}
	h.Infos.Style = hplot.HInfoSummary
	p.Add(h)

	// Save the plot to a PNG file.
	if err := p.Save(6*vg.Inch, -1, name); err != nil {
		log.Fatalf("error saving plot: %v\n", err)
	}
}

type mVsrResults struct {
	r []float64
	m []float64
}

func (m *mVsrResults) Len() int {
	return len(m.r)
}

func (m *mVsrResults) XY(i int) (x, y float64) {
	x = m.r[i]
	y = m.m[i]
	return
}

func mVsr(dt float64) {
	const ti = 0   // seconds
	const tf = 100 // seconds
	// Take Na22 period (2.6 years)

	mVsrRes := &mVsrResults{}

	isotope := Isotope{Name: "Na22", T: 8.2e7}
	for i := 1; i < 1000; i += 1 {
		if i%100 == 0 {
			fmt.Printf("i=%v\n", i)
		}
		r := float64(i) * 2
		source := &Source{Isotope: isotope, Name: "Na22", Activity0: r, T0: -1.077e8}
		//fmt.Printf("N(ti)=%v, N(tf)=%v\n", source.NoNuclei(ti), source.NoNuclei(tf))
		noDecays := int(math.Ceil(source.NoNuclei(ti) - source.NoNuclei(tf)))
		//fmt.Println("noDecays =", noDecays)
		events := NewEventColl(source, ti, noDecays)
		nLost := events.DeadTimeLoss(dt, false)
		nMeasured := noDecays - nLost
		m := float64(nMeasured) / (tf - ti)

		mVsrRes.r = append(mVsrRes.r, r)
		mVsrRes.m = append(mVsrRes.m, m)
	}
	p, err := plot.New()
	if err != nil {
		panic(err)
	}
	p.Title.Text = "measured rate vs true rate"
	p.X.Label.Text = "r"
	p.Y.Label.Text = "m"
	p.Y.Max = 2 * 1 / dt
	p.Add(plotter.NewGrid())

	line := make(plotter.XYs, 2)
	line[0].X = mVsrRes.r[0]
	line[0].Y = 1 / dt
	line[1].X = mVsrRes.r[len(mVsrRes.r)-1]
	line[1].Y = 1 / dt

	err = plotutil.AddLinePoints(p, "", mVsrRes, "Asymptotic value", line)

	if err != nil {
		panic(err)
	}

	if err := p.Save(14*vg.Inch, 5*vg.Inch, "mVsr.png"); err != nil {
		panic(err)
	}
}

func main() {
	// Set initial and final run time
	const ti = 0 // seconds
	const tf = 1 // seconds

	mVsr(0.04)

	/*
		na22 := Isotope{Name: "Na22", T: 8.2e7}
		na22_2MBq := &Source{Isotope: na22, Name: "2MBq", Activity0: 2.69e6, T0: -1.077e8}
		na22_2MBq_noDecays := int(math.Ceil(na22_2MBq.NoNuclei(ti) - na22_2MBq.NoNuclei(tf)))

		fmt.Println(" na22_2MBq_noDecays=", na22_2MBq_noDecays)

			events_Na22_2MBq := NewEventColl(na22_2MBq, ti, na22_2MBq_noDecays)
			nLost_Na22_2MBq := events_Na22_2MBq.DeadTimeLoss(0.04)
			fmt.Println(nLost_Na22_2MBq)

			// Cross check plots
			events_Na22_2MBq.PlotTimeVsEvtIndex("events_Na22_2MBq_TimeVsEvtIndex.png")
			events_Na22_2MBq.PlotRelTimeHist("events_Na22_2MBq_RelTimeHist.png")
	*/
}
