// Package pulse defines the physical signals recorded by the ASM cards.
package pulse

import (
	"fmt"
	"log"
	"math"
	"strconv"

	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/plotutil"
	"github.com/gonum/plot/vg"
	"github.com/gonum/stat"
	"gitlab.in2p3.fr/avirm/analysis-go/detector"
)

// Sample describes the signal recorded by a DRS capacitor
type Sample struct {
	Amplitude float64
	Time      float64
	Index     uint16 // index of sample within pulse
	Capacitor *detector.Capacitor
}

// NewSample constructs a new sample
func NewSample(amp float64, index uint16, time float64) *Sample {
	s := &Sample{
		Amplitude: amp,
		Index:     index,
		Time:      time,
		Capacitor: nil,
	}
	return s
}

// CapaIndex returns the index of the capacitor associated the the sample (from 0 to 1023)
func (s *Sample) CapaIndex(SRout uint16) uint16 {
	iCapa := SRout + s.Index
	if iCapa > 1023 {
		iCapa -= 1024
	}
	return iCapa
}

// Print prints sample informations
func (s *Sample) Print() {
	fmt.Printf("Printing sample: \n")
	fmt.Printf("  -> Amplitude, Time, Index, Capacitor = %v, %v, %v, %p\n", s.Amplitude, s.Time, s.Index, s.Capacitor)
}

// SubtractPedestal subtracts pedestal for the sample.
// This assumes that pedestals have been computed before.
func (s *Sample) SubtractPedestal() {
	if s.Capacitor == nil {
		log.Fatal("Error ! no capacitor associated to this sample")
	}
	s.Amplitude -= s.Capacitor.PedestalMean()
}

// Pulse describes a DRS channel.
// A pulse is made of several samples
type Pulse struct {
	Samples      []Sample
	TimeStep     float64
	SRout        uint16 // SRout is the number of the first capacitor for this pulse, it can take 1024 values (0 -> 1023)
	HasSignal    bool
	HasSatSignal bool
	Channel      *detector.Channel
}

// NewPulse constructs a new pulse
func NewPulse(channel *detector.Channel) *Pulse {
	return &Pulse{
		TimeStep:     0.0,
		SRout:        0,
		HasSignal:    false,
		HasSatSignal: false,
		Channel:      channel,
	}
}

// Print prints pulse informations
func (p *Pulse) Print(detailed bool) {
	fmt.Println("-> Printing pulse informations")
	fmt.Printf("  o len(Samples) = %v\n", len(p.Samples))
	fmt.Printf("  o TimeStep = %v\n", p.TimeStep)
	fmt.Printf("  o SRout = %v\n", p.SRout)
	fmt.Printf("  o HasSignal = %v\n", p.HasSignal)
	if p.Channel != nil {
		fmt.Printf("  o Channel name = %v\n", p.Channel.Name())
	} else {
		fmt.Printf("  o Channel = nil\n")
	}
	if detailed {
		fmt.Println("  o Capacitors = ")
		for i := range p.Samples {
			sample := &p.Samples[i]
			if sample.Capacitor != nil {
				if i == 0 {
					fmt.Printf("%v", sample.Capacitor.ID())
				} else {
					fmt.Printf(", %v", sample.Capacitor.ID())
				}
			}
		}
		fmt.Println("\n")
	}
}

// NoSamples return the number of samples the pulse is made of
func (p *Pulse) NoSamples() uint16 {
	return uint16(len(p.Samples))
}

// Copy makes a copy of the pulse
func (p *Pulse) Copy() *Pulse {
	newsamples := make([]Sample, len(p.Samples))
	copy(newsamples, p.Samples)
	newpulse := &Pulse{
		Samples:      newsamples,
		TimeStep:     p.TimeStep,
		SRout:        p.SRout,
		HasSignal:    p.HasSignal,
		HasSatSignal: p.HasSatSignal,
		Channel:      p.Channel,
	}
	return newpulse
}

// AddSample adds a sample to the pulse
func (p *Pulse) AddSample(s *Sample, capa *detector.Capacitor, thres float64) {
	if s.Capacitor != nil {
		log.Fatal("capacitor is not nil")
	}
	s.Capacitor = capa
	p.Samples = append(p.Samples, *s)
	if s.Amplitude > thres {
		p.HasSignal = true
		if s.Amplitude == 4095 {
			p.HasSatSignal = true
		}
	}
	noSamples := len(p.Samples)
	if noSamples == 2 {
		p.TimeStep = s.Time - p.Samples[noSamples-2].Time
	}
	if noSamples >= 3 {
		tStep := s.Time - p.Samples[noSamples-2].Time
		if math.Abs(tStep-p.TimeStep)/p.TimeStep > 0.0001 {
			log.Fatalf("time step varies: tStep = %v, p.TimeStep = %v", tStep, p.TimeStep)
		}
	}
}

// SubtractPedestal subtracts pedestals for all the samples of the pulse
func (p *Pulse) SubtractPedestal() {
	for i := range p.Samples {
		sample := &p.Samples[i]
		sample.SubtractPedestal()
	}
}

// Amplitude returns the amplitude of the sample having the highest amplitude
func (p *Pulse) Amplitude() float64 {
	var ampl float64 = 0
	for _, s := range p.Samples {
		if s.Amplitude > ampl {
			ampl = s.Amplitude
		}
	}
	return ampl
}

// Charge returns the area under the pulse
func (p *Pulse) Charge() float64 {
	var sum float64 = 0
	for _, s := range p.Samples {
		sum += s.Amplitude
	}
	return sum * float64(p.TimeStep)
}

// XaxisType defines the x axis type for plotting
type XaxisType byte

const (
	XaxisTime XaxisType = iota
	XaxisIndex
	XaxisCapacitor
)

// MakeXY returns a plotter.XYs used for plotting
func (p *Pulse) MakeXY(x XaxisType) plotter.XYs {
	pts := make(plotter.XYs, len(p.Samples))
	for i, sample := range p.Samples {
		var xval float64
		switch x {
		case XaxisTime:
			xval = float64(sample.Time)
		case XaxisIndex:
			xval = float64(sample.Index)
		case XaxisCapacitor:
			xval = float64(sample.Capacitor.ID())
		}
		pts[i].X = xval
		pts[i].Y = sample.Amplitude
	}
	return pts
}

// MakeAmpSlice makes a slice containing amplitudes of all samples
func (p *Pulse) MakeAmpSlice() []float64 {
	var amps []float64
	for _, samp := range p.Samples {
		amps = append(amps, samp.Amplitude)
	}
	return amps
}

// Correlation computes the correlation between two pulses
func (p *Pulse) Correlation(pu *Pulse) float64 {
	amplitudes1 := p.MakeAmpSlice()
	amplitudes2 := pu.MakeAmpSlice()
	var weights []float64
	return stat.Correlation(amplitudes1, amplitudes2, weights)
}

// Cluster describes a quartet.
type Cluster struct {
	Pulses   [4]Pulse
	ID       uint8
	Counters []uint32 // Counters stores the counters present in the binary/hexa/decimal files
}

// NewClusterFromID constructs a new cluster from ID only
func NewClusterFromID(id uint8) *Cluster {
	return &Cluster{
		ID: id,
	}
}

// NewCluster constructs a new cluster
func NewCluster(id uint8, pulses [4]Pulse) *Cluster {
	return &Cluster{
		Pulses: pulses,
		ID:     id,
	}
}

// Print prints cluster informations
func (c *Cluster) Print(detailed bool) {
	fmt.Printf("-> Printing cluster with ID=%v\n", c.ID)
	fmt.Printf(" o Number of counters=%v\n", len(c.Counters))
	for i := range c.Pulses {
		c.Pulses[i].Print(detailed)
	}
}

// NoSamples returns the number of samples of the pulses in the cluster
func (c *Cluster) NoSamples() uint16 {
	noSamples := c.Pulses[0].NoSamples()
	for i := 1; i < len(c.Pulses); i++ {
		n := c.Pulses[i].NoSamples()
		if n != noSamples {
			log.Fatal("all pulses don't have the same number of samples")
		}
	}
	return noSamples
}

// SRout returns the srout (common to all pulses in the cluster)
func (c *Cluster) SRout() uint16 {
	srout := c.Pulses[0].SRout
	for i := 1; i < len(c.Pulses); i++ {
		if srout != c.Pulses[i].SRout {
			fmt.Printf("not all pulses have the same SRout in this cluster. Mismatching SRouts are %v and %v.\n", srout, c.Pulses[i].SRout)
			log.Fatalf(" -> quitting")
		}
	}
	return srout
}

// PulsesWithSignal returns a slice with pointers to pulses with signal
func (c *Cluster) PulsesWithSignal() []*Pulse {
	var pulses []*Pulse
	for i := range c.Pulses {
		if c.Pulses[i].HasSignal {
			pulses = append(pulses, &c.Pulses[i])
		}
	}
	return pulses
}

// PulsesWithSignal returns a slice with pointers to pulses with saturating signal
func (c *Cluster) PulsesWithSatSignal() []*Pulse {
	var pulses []*Pulse
	for i := range c.Pulses {
		if c.Pulses[i].HasSatSignal {
			pulses = append(pulses, &c.Pulses[i])
		}
	}
	return pulses
}

// Amplitude returns the sum of the pulse amplitudes
func (c *Cluster) Amplitude() float64 {
	amp := 0.
	for i := range c.Pulses {
		amp += c.Pulses[i].Amplitude()
	}
	return amp
}

// Charge returns the sum of the pulse charges
func (c *Cluster) Charge() float64 {
	charge := 0.
	for i := range c.Pulses {
		charge += c.Pulses[i].Charge()
	}
	return charge
}

// Counter return the value of counter for the given index
func (c *Cluster) Counter(i int) uint32 {
	if i >= len(c.Counters) {
		log.Fatalf("pulse: counter index out of range")
	}
	return c.Counters[i]
}

// PlotPulses plots the four pulses of the cluster in one canvas
func (c *Cluster) PlotPulses(evtID uint, x XaxisType, pedestalRange bool) string {
	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	p.Title.Text = "Pulse for event " + strconv.Itoa(int(evtID)) + " cluster " + strconv.Itoa(int(c.ID))
	switch x {
	case XaxisTime:
		p.X.Label.Text = "time (ns)"
	case XaxisIndex:
		p.X.Label.Text = "sample index"
	case XaxisCapacitor:
		p.X.Label.Text = "capacitor"
	}
	p.Y.Label.Text = "amplitude"
	p.Legend.Top = true

	p.Add(plotter.NewGrid())
	var TextsAndXYs []interface{}
	for i := range c.Pulses {
		if c.Pulses[i].Channel != nil {
			TextsAndXYs = append(TextsAndXYs, c.Pulses[i].Channel.Name()+" (HasSignal = "+strconv.FormatBool(c.Pulses[i].HasSignal)+
				", Sat = "+strconv.FormatBool(c.Pulses[i].HasSatSignal)+
				", SRout = "+strconv.Itoa(int(c.Pulses[i].SRout))+
				")")
			TextsAndXYs = append(TextsAndXYs, c.Pulses[i].MakeXY(x))
		}
	}
	err = plotutil.AddLinePoints(p, TextsAndXYs...)
	if err != nil {
		panic(err)
	}

	switch pedestalRange {
	case true:
		p.Y.Min = 300
		p.Y.Max = 700
		// 				p.Y.Min = -50
		// 				p.Y.Max = 50

	case false:
		p.Y.Min = -500
		p.Y.Max = 4096
	}

	outFile := "output/pulses_event" + strconv.Itoa(int(evtID)) + "_cluster" + strconv.Itoa(int(c.ID)) + ".png"

	if err := p.Save(14*vg.Inch, 5*vg.Inch, outFile); err != nil {
		panic(err)
	}

	return outFile
}
