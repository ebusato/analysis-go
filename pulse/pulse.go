package pulse

import (
	"fmt"
	"log"
	"math"
	"strconv"

	"gitlab.in2p3.fr/AVIRM/Analysis-go/detector"

	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/plotutil"
	"github.com/gonum/plot/vg"
	"github.com/gonum/stat"
)

type Sample struct {
	Amplitude float64
	Time      float64
	Index     uint16 // index of sample within pulse
	Capacitor *detector.Capacitor
}

func NewSample(amp float64, index uint16, time float64) *Sample {
	s := &Sample{
		Amplitude: amp,
		Index:     index,
		Time:      time,
		Capacitor: nil,
	}
	return s
}

func (s *Sample) CapaIndex(SRout uint16) uint16 {
	iCapa := SRout + s.Index
	if iCapa > 1023 {
		iCapa -= 1024
	}
	return iCapa
}

func (s *Sample) Print() {
	fmt.Printf("Printing sample: \n")
	fmt.Printf("  -> Amplitude, Time, Index, Capacitor = %v, %v, %v, %p\n", s.Amplitude, s.Time, s.Index, s.Capacitor)
}

func (s *Sample) SubtractPedestal() {
	s.Amplitude -= s.Capacitor.PedestalMean()
}

type Pulse struct {
	Samples      []Sample
	TimeStep     float64
	SRout        uint16 // SRout is the number of the first capacitor for this pulse, it can take 1024 values (0 -> 1023)
	HasSignal    bool
	HasSatSignal bool
	Channel      *detector.Channel
}

func NewPulse(channel *detector.Channel) *Pulse {
	return &Pulse{
		TimeStep:     0.0,
		SRout:        0,
		HasSignal:    false,
		HasSatSignal: false,
		Channel:      channel,
	}
}

func (p *Pulse) Print() {
	fmt.Println("-> Printing pulse informations")
	fmt.Printf("  o len(Samples) = %v\n", len(p.Samples))
	fmt.Printf("  o TimeStep = %v\n", p.TimeStep)
	fmt.Printf("  o SRout = %v\n", p.SRout)
	fmt.Printf("  o HasSignal = %v\n", p.HasSignal)
	if p.Channel != nil {
		fmt.Printf("  o Channel name = %v\n", p.Channel.Name())
	}
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

func (p *Pulse) NoSamples() uint16 {
	return uint16(len(p.Samples))
}

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

func (p *Pulse) AddSample(s *Sample, capa *detector.Capacitor) {
	if s.Capacitor != nil {
		log.Fatal("capacitor is not nil")
	}
	s.Capacitor = capa
	p.Samples = append(p.Samples, *s)
	if s.Amplitude > 1000 {
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

func (p *Pulse) SubtractPedestal() {
	for i := range p.Samples {
		sample := &p.Samples[i]
		sample.SubtractPedestal()
	}
}

func (p *Pulse) Amplitude() float64 {
	var ampl float64 = 0
	for _, s := range p.Samples {
		if s.Amplitude > ampl {
			ampl = s.Amplitude
		}
	}
	return ampl
}

func (p *Pulse) Charge() float64 {
	var sum float64 = 0
	for _, s := range p.Samples {
		sum += s.Amplitude
	}
	return sum * float64(p.TimeStep)
}

type XaxisType byte

const (
	XaxisTime XaxisType = iota
	XaxisIndex
	XaxisCapacitor
)

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

func (p *Pulse) MakeAmpSlice() []float64 {
	var amps []float64
	for _, samp := range p.Samples {
		amps = append(amps, samp.Amplitude)
	}
	return amps
}

func (p *Pulse) Correlation(pu *Pulse) float64 {
	amplitudes1 := p.MakeAmpSlice()
	amplitudes2 := pu.MakeAmpSlice()
	var weights []float64
	return stat.Correlation(amplitudes1, amplitudes2, weights)
}

type Cluster struct {
	Pulses [4]Pulse
}

func NewCluster(pulses [4]Pulse) *Cluster {
	return &Cluster{
		Pulses: pulses,
	}
}

func (c *Cluster) Print() {
	for i := range c.Pulses {
		fmt.Printf("->Printing cluster")
		c.Pulses[i].Print()
	}
}

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

func (c *Cluster) SRout() uint16 {
	srout := c.Pulses[0].SRout
	for i := 1; i < len(c.Pulses); i++ {
		if srout != c.Pulses[i].SRout {
			log.Fatalf("not all pulses have the same SRout in this event")
		}
	}
	return srout
}

func (c *Cluster) PulsesWithSignal() []*Pulse {
	var pulses []*Pulse
	for i := range c.Pulses {
		if c.Pulses[i].HasSignal {
			pulses = append(pulses, &c.Pulses[i])
		}
	}
	return pulses
}

func (c *Cluster) Amplitude() float64 {
	amp := 0.
	for i := range c.Pulses {
		amp += c.Pulses[i].Amplitude()
	}
	return amp
}

func (c *Cluster) Charge() float64 {
	charge := 0.
	for i := range c.Pulses {
		charge += c.Pulses[i].Charge()
	}
	return charge
}

func (c *Cluster) PlotPulses(ID uint, x XaxisType, pedestalRange bool) string {
	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	p.Title.Text = "Pulse for event " + strconv.Itoa(int(ID)) + " (SRout = " + strconv.Itoa(int(c.SRout())) + ")"
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
	err = plotutil.AddLinePoints(p,
		c.Pulses[0].Channel.Name()+" (HasSignal = "+strconv.FormatBool(c.Pulses[0].HasSignal)+", Sat = "+strconv.FormatBool(c.Pulses[0].HasSatSignal)+")", c.Pulses[0].MakeXY(x),
		c.Pulses[1].Channel.Name()+" (HasSignal = "+strconv.FormatBool(c.Pulses[1].HasSignal)+", Sat = "+strconv.FormatBool(c.Pulses[1].HasSatSignal)+")", c.Pulses[1].MakeXY(x),
		c.Pulses[2].Channel.Name()+" (HasSignal = "+strconv.FormatBool(c.Pulses[2].HasSignal)+", Sat = "+strconv.FormatBool(c.Pulses[2].HasSatSignal)+")", c.Pulses[2].MakeXY(x),
		c.Pulses[3].Channel.Name()+" (HasSignal = "+strconv.FormatBool(c.Pulses[3].HasSignal)+", Sat = "+strconv.FormatBool(c.Pulses[3].HasSatSignal)+")", c.Pulses[3].MakeXY(x))

	if err != nil {
		panic(err)
	}

	switch pedestalRange {
	case true:
		p.Y.Min = 300
		p.Y.Max = 700
	case false:
		p.Y.Min = -500
		p.Y.Max = 4096
	}

	outFile := "output/pulses_event" + strconv.Itoa(int(ID)) + ".png"

	if err := p.Save(14*vg.Inch, 5*vg.Inch, outFile); err != nil {
		panic(err)
	}

	return outFile
}
