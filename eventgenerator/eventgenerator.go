package main

import (
	"fmt"
	"log"
	"math"
	"sort"

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

func (s *Source) NoDecays(ti, tf float64) int {
	return int(math.Ceil(s.NoNuclei(ti) - s.NoNuclei(tf)))
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

	e := Event{Source: s, Time: dist.Rand() + t}
	return e
}

func Print(evts []Event, n int) {
	fmt.Printf("Source: %v\n", evts[0].Source.Name)
	fmt.Printf("Printing first %v events:\n", n)
	for i := 0; i < n; i++ {
		fmt.Printf("  -> event %v: source = %v, time = %v\n", i, evts[i].Source.Name, evts[i].Time)
	}
}

// Events in the Events slice are sorted with increasing event time
type EventColl struct {
	Source *Source
	Events []Event
}

func NewEventColl(s *Source, ti float64, nEvents int) *EventColl {
	evtColl := &EventColl{Source: s}
	time := ti
	evtColl.Events = make([]Event, nEvents)
	for i := 0; i < nEvents; i++ {
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

type ByTime []Event

func (b ByTime) Len() int           { return len(b) }
func (b ByTime) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }
func (b ByTime) Less(i, j int) bool { return b[i].Time < b[j].Time }

// Mixture is a mixture of events from different sources
type Mixture struct {
	Name         string
	NoEventsTot  map[string]int
	NoEventsLost map[string]int
	Events       []Event
}

// Gather multiple event collections into a single event collection
func NewMixture(name string, eColls ...*EventColl) *Mixture {
	mixture := &Mixture{Name: name}
	mixture.NoEventsTot = make(map[string]int)
	mixture.NoEventsLost = make(map[string]int)
	for i := range eColls {
		mixture.NoEventsTot[eColls[i].Source.Name] = len(eColls[i].Events)
		mixture.Events = append(mixture.Events, eColls[i].Events...)
	}
	sort.Sort(ByTime(mixture.Events))
	return mixture
}

// MakeDPGAMixture tries to emulate mixture of processes in DPGA
func MakeDPGAMixture(ti, tf float64, factor float64) *Mixture {
	na22 := Isotope{Name: "Na22", T: 8.2e7}
	/*
		// Activities below account for detection efficiency
		// They are taken from trigger rates for equation 3 vs 3L
		na22_2MBq := &Source{Isotope: na22, Name: "Na22_2MBq", Activity0: 12500 * factor, T0: 0}
		na22_16kBq := &Source{Isotope: na22, Name: "Na22_16kBq", Activity0: 50 * factor, T0: 0}
		lyso := Isotope{Name: "lyso", T: 1e9}
		lyso_1kBq := &Source{Isotope: lyso, Name: "LYSO", Activity0: 400 * factor, T0: 0}
	*/

	///////////////////////////////////////////////////////////////////////////////
	// for tests
	na22_2MBq := &Source{Isotope: na22, Name: "Na22_2MBq", Activity0: 0.15 * factor, T0: 0}
	na22_16kBq := &Source{Isotope: na22, Name: "Na22_16kBq", Activity0: 0.2 * factor, T0: 0}
	lyso := Isotope{Name: "lyso", T: 1e9}
	lyso_1kBq := &Source{Isotope: lyso, Name: "LYSO", Activity0: 0.3 * factor, T0: 0}
	///////////////////////////////////////////////////////////////////////////////

	fmt.Printf("activities=%v, %v, %v\n", na22_2MBq.Activity(ti), na22_16kBq.Activity(ti), lyso_1kBq.Activity(ti))
	na22_2MBq_noDecays := na22_2MBq.NoDecays(ti, tf)
	na22_16kBq_noDecays := na22_16kBq.NoDecays(ti, tf)
	lyso_1kBq_noDecays := lyso_1kBq.NoDecays(ti, tf)
	fmt.Printf("noDecays=%v, %v, %v\n", na22_2MBq_noDecays, na22_16kBq_noDecays, lyso_1kBq_noDecays)
	events_Na22_2MBq := NewEventColl(na22_2MBq, ti, na22_2MBq_noDecays)
	events_Na22_16kBq := NewEventColl(na22_16kBq, ti, na22_16kBq_noDecays)
	events_Lyso_1kBq := NewEventColl(lyso_1kBq, ti, lyso_1kBq_noDecays)

	mixture := NewMixture("DPGA_nominal", events_Na22_2MBq, events_Na22_16kBq, events_Lyso_1kBq)
	return mixture
}

func (m *Mixture) RateTrue(name string, ti, tf float64) float64 {
	return float64(m.NoEventsTot[name]) / (tf - ti)
}

func (m *Mixture) RateMeasured(name string, ti, tf float64) float64 {
	noEventsMeasured := m.NoEventsTot[name] - m.NoEventsLost[name]
	return float64(noEventsMeasured) / (tf - ti)
}

func (m *Mixture) DeadTimeLoss(ti, tf, dt float64, paralizable bool) (float64, float64) {
	evtGeneratingDeadTime := &m.Events[0] // the first event is generating the first dead time
	for i := 1; i < len(m.Events); i++ {
		switch m.Events[i].Time-evtGeneratingDeadTime.Time < dt {
		case true:
			m.NoEventsLost[m.Events[i].Source.Name]++
			if paralizable {
				evtGeneratingDeadTime = &m.Events[i]
			}
		case false:
			evtGeneratingDeadTime = &m.Events[i]
		}
	}

	fmt.Printf("Summary:\n")
	fmt.Println("NoEventsTot:", m.NoEventsTot)
	fmt.Println("NoEventsLost:", m.NoEventsLost)

	// Compute total measured and true rates
	rateTrueTot := 0.
	rateMeasTot := 0.
	for name := range m.NoEventsTot {
		rateTrueTot += m.RateTrue(name, ti, tf)
		rateMeasTot += m.RateMeasured(name, ti, tf)
	}
	fmt.Printf("Total true and measured rates = %v, %v\n", rateTrueTot, rateMeasTot)

	for name := range m.NoEventsTot {
		fmt.Printf(" -> %v: %9.5v (%7.5v %% of total) %9.7v (%7.5v %% of total)\n",
			name,
			m.RateTrue(name, ti, tf), m.RateTrue(name, ti, tf)/rateTrueTot*100,
			m.RateMeasured(name, ti, tf), m.RateMeasured(name, ti, tf)/rateMeasTot*100)
	}

	return rateTrueTot, rateMeasTot
}

type XYResults struct {
	X []float64
	Y []float64
}

func (m *XYResults) Len() int {
	return len(m.X)
}

func (m *XYResults) XY(i int) (x, y float64) {
	x = m.X[i]
	y = m.Y[i]
	return
}

func mVsr(dt float64) {
	const ti = 0   // seconds
	const tf = 100 // seconds
	// Take Na22 period (2.6 years)

	mVsrRes := &XYResults{}

	isotope := Isotope{Name: "Na22", T: 8.2e7}
	for i := 1; i < 1000; i += 1 {
		if i%100 == 0 {
			fmt.Printf("i=%v\n", i)
		}
		r := float64(i) * 2
		source := &Source{Isotope: isotope, Name: "Na22", Activity0: r, T0: -1.077e8}
		//fmt.Printf("N(ti)=%v, N(tf)=%v\n", source.NoNuclei(ti), source.NoNuclei(tf))
		noDecays := source.NoDecays(ti, tf)
		//fmt.Println("noDecays =", noDecays)
		events := NewEventColl(source, ti, noDecays)
		nLost := events.DeadTimeLoss(dt, false)
		nMeasured := noDecays - nLost
		m := float64(nMeasured) / (tf - ti)

		mVsrRes.X = append(mVsrRes.X, r)
		mVsrRes.Y = append(mVsrRes.Y, m)
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
	line[0].X = mVsrRes.X[0]
	line[0].Y = 1 / dt
	line[1].X = mVsrRes.X[len(mVsrRes.X)-1]
	line[1].Y = 1 / dt

	err = plotutil.AddLinePoints(p, "", mVsrRes, "Asymptotic value", line)

	if err != nil {
		panic(err)
	}

	if err := p.Save(14*vg.Inch, 5*vg.Inch, "mVsr.png"); err != nil {
		panic(err)
	}
}

func mVsrDPGAMixture(dt, ti, tf float64) {
	// Components:
	//   0 -> mTot and rTot
	//   1 -> mLyso and rLyso
	//   2 -> mNa22_16kBq and rNa22_16kBq
	//   3 -> mNa22_2MBq and rNa22_2MBq
	var mVsrRes [4]XYResults
	var ratioVsrRes [3]XYResults

	for i := 1; i < 100; i += 1 {
		if i%1 == 0 {
			fmt.Printf("i=%v\n", i)
		}
		mixture := MakeDPGAMixture(ti, tf, float64(i))
		rateTrueTot, rateMeasTot := mixture.DeadTimeLoss(ti, tf, dt, false)
		mVsrRes[0].X = append(mVsrRes[0].X, rateTrueTot)
		mVsrRes[1].X = append(mVsrRes[1].X, rateTrueTot)
		mVsrRes[2].X = append(mVsrRes[2].X, rateTrueTot)
		mVsrRes[3].X = append(mVsrRes[3].X, rateTrueTot)

		mVsrRes[0].Y = append(mVsrRes[0].Y, rateMeasTot)
		mVsrRes[1].Y = append(mVsrRes[1].Y, mixture.RateMeasured("LYSO", ti, tf))
		mVsrRes[2].Y = append(mVsrRes[2].Y, mixture.RateMeasured("Na22_16kBq", ti, tf))
		mVsrRes[3].Y = append(mVsrRes[3].Y, mixture.RateMeasured("Na22_2MBq", ti, tf))

		ratioVsrRes[0].X = append(ratioVsrRes[0].X, rateTrueTot)
		ratioVsrRes[1].X = append(ratioVsrRes[1].X, rateTrueTot)
		ratioVsrRes[2].X = append(ratioVsrRes[2].X, rateTrueTot)

		ratioVsrRes[0].Y = append(ratioVsrRes[0].Y, mixture.RateMeasured("LYSO", ti, tf)/rateMeasTot)
		ratioVsrRes[1].Y = append(ratioVsrRes[1].Y, mixture.RateMeasured("Na22_16kBq", ti, tf)/rateMeasTot)
		ratioVsrRes[2].Y = append(ratioVsrRes[2].Y, mixture.RateMeasured("Na22_2MBq", ti, tf)/rateMeasTot)
	}

	////////////////////////////////////////////
	// plot m vs r
	p0, err := plot.New()
	if err != nil {
		panic(err)
	}
	p0.Title.Text = "measured rates vs total true rate"
	p0.X.Label.Text = "rTot"
	p0.Y.Label.Text = "m"
	p0.Y.Max = 2 * 1 / dt
	p0.Add(plotter.NewGrid())

	line := make(plotter.XYs, 2)
	line[0].X = mVsrRes[0].X[0]
	line[0].Y = 1 / dt
	line[1].X = mVsrRes[0].X[len(mVsrRes[0].X)-1]
	line[1].Y = 1 / dt

	err = plotutil.AddLinePoints(p0, "Tot", &mVsrRes[0], "LYSO", &mVsrRes[1], "Na22_16kBq", &mVsrRes[2], "Na22_2MBq", &mVsrRes[3], "Asymptotic value", line)

	if err != nil {
		panic(err)
	}

	if err := p0.Save(14*vg.Inch, 5*vg.Inch, "mVsrDPGAMixture.png"); err != nil {
		panic(err)
	}
	////////////////////////////////////////////

	////////////////////////////////////////////
	// plot ratios
	p1, err := plot.New()
	if err != nil {
		panic(err)
	}
	p1.Title.Text = "measured rates/total measured rate vs total true rate"
	p1.X.Label.Text = "rTot"
	p1.Y.Label.Text = "measured rate/total measured rate"
	p1.Y.Min = 0
	p1.Add(plotter.NewGrid())

	err = plotutil.AddLinePoints(p1, "LYSO", &ratioVsrRes[0], "Na22_16kBq", &ratioVsrRes[1], "Na22_2MBq", &ratioVsrRes[2])

	if err != nil {
		panic(err)
	}

	if err := p1.Save(14*vg.Inch, 5*vg.Inch, "ratiosVsrDPGAMixture.png"); err != nil {
		panic(err)
	}
	////////////////////////////////////////////
}

func main() {
	// Set initial and final run time
	const ti = 0    // seconds
	const tf = 2000 // seconds
	const dt = 0.5   // seconds

	// 	mVsr(dt)
	mVsrDPGAMixture(dt, ti, tf)

	/*
		na22 := Isotope{Name: "Na22", T: 8.2e7}
		na22_2MBq := &Source{Isotope: na22, Name: "2MBq", Activity0: 2.69e6, T0: -1.077e8}
		na22_2MBq_noDecays := na22_2MBq.NoDecays(ti, tf)

		fmt.Println(" na22_2MBq_noDecays=", na22_2MBq_noDecays)

			events_Na22_2MBq := NewEventColl(na22_2MBq, ti, na22_2MBq_noDecays)
			nLost_Na22_2MBq := events_Na22_2MBq.DeadTimeLoss(0.04)
			fmt.Println(nLost_Na22_2MBq)

			// Cross check plots
			events_Na22_2MBq.PlotTimeVsEvtIndex("events_Na22_2MBq_TimeVsEvtIndex.png")
			events_Na22_2MBq.PlotRelTimeHist("events_Na22_2MBq_RelTimeHist.png")
	*/
}
