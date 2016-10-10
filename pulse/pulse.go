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
	if SRout > 1023 {
		log.Fatalf("SRout = %v (>1023)\n", SRout)
	}
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

// Subtract subtracts the provided number to the sample amplitude.
func (s *Sample) Subtract(a float64) {
	s.Amplitude -= a
}

// SubtractPedestal subtracts pedestal for the sample.
// This assumes that pedestals have been computed before.
func (s *Sample) SubtractPedestal() {
	if s.Capacitor == nil {
		log.Print("warning ! no capacitor associated to this sample")
	} else {
		s.Amplitude -= s.Capacitor.PedestalMean()
	}
}

// Pulse describes a DRS channel.
// A pulse is made of several samples
type Pulse struct {
	Samples             []Sample
	TimeStep            float64
	SRout               uint16 // SRout is the number of the first capacitor for this pulse, it can take 1024 values (0 -> 1023)
	HasSignal           bool
	HasSatSignal        bool
	Ampl                float64
	AmplIndex           int
	Charg               float64 // Charge, removed the final "e" because the name "Charge" is already used by the method
	Time20              float64 // time at 20% on rising front of the pulse
	Time30              float64 // time at 30% on rising front of the pulse
	Time80              float64 // time at 80% on rising front of the pulse
	NoLocMaxRisingFront int     // number of local maxima on rising front (counted between 20% and 80%)
	Channel             *detector.Channel
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
	if s.Amplitude >= thres {
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

// SubtractTimeDepOffset subtracts time dependent offsets for all the samples of the pulse
func (p *Pulse) SubtractTimeDepOffsets() {
	ch := p.Channel
	for iSample := range p.Samples {
		sample := &p.Samples[iSample]
		sample.Subtract(ch.TimeDepOffsetMean(iSample))
	}
}

// Amplitude returns the amplitude of the sample having the highest amplitude
func (p *Pulse) Amplitude() (int, float64) {
	var ampl float64
	var amplIndex int
	for i, s := range p.Samples {
		if s.Amplitude > ampl {
			ampl = s.Amplitude
			amplIndex = i
		}
	}
	p.Ampl = ampl
	p.AmplIndex = amplIndex
	return amplIndex, ampl
}

// Charge returns the area under the pulse
func (p *Pulse) Charge() float64 {
	var sum float64 = 0
	for _, s := range p.Samples {
		sum += s.Amplitude
	}
	p.Charg = sum * float64(p.TimeStep)
	return p.Charg
}

// Time returns the time at which the signal is a factor "frac" (second parameter) of its amplitude plus the index ilow.
// If recomputeAmpl is true, then idxStart is ignored and iteration start at p.AmplIndex.
// If recomputeAmpl is false, then downwards iterations start at idxStart (it is assumed in this case that the amplitude
// has been computed prior to the call to this method and thus that p.AmplIndex is available in case the user wants
// idxStart=p.AmplIndex).
func (p *Pulse) T(recomputeAmpl bool, frac float64, idxStart int) (float64, int, int) {
	if recomputeAmpl {
		p.Amplitude()
	} else if p.Ampl == 0 {
		panic("pulse amplitude is 0, meaning that the amplitude was never calculated before. You should set the recomputeAmpl flag to true")
	}
	//fmt.Println("debug Time:", p.Ampl)
	// Determination of ilow
	// ilow is the index of the sample for which the amplitude is just below frac of the pulse amplitude
	var ilow int
	var ampllow float64
	var amplprev float64
	PrevDerivativeIspositive := true // derivative sign computed in order of increasing time
	noLocMax := 0
	if recomputeAmpl {
		idxStart = p.AmplIndex
	}
	for i := idxStart; i >= 0; i-- {
		//fmt.Println("   ->", i, p.Samples[i].Amplitude)
		ampl := p.Samples[i].Amplitude
		if i != idxStart {
			if ampl < amplprev && PrevDerivativeIspositive == false {
				noLocMax++
				//fmt.Println("frac, i, ampl, TimelocMax =", frac, i, ampl, p.Samples[i].Time)
				PrevDerivativeIspositive = true
			}
			if ampl > amplprev {
				PrevDerivativeIspositive = false
			}
		}
		amplprev = ampl
		if ampl < frac*p.Ampl {
			ilow = i
			ampllow = p.Samples[i].Amplitude
			//fmt.Println("   -> found ilow, breaking")
			break
		}
	}
	// if ilow == 0, then we do not know Time -> return 0
	if ilow == 0 {
		return 0, 0, 0
	}
	ihigh := ilow + 1
	amplhigh := p.Samples[ihigh].Amplitude
	// Sanity check
	if amplhigh < frac*p.Ampl {
		panic("p.Samples[i30high] < frac * p.Ampl, this should never happen")
	}
	if amplhigh == ampllow {
		// This should never happen but just to be sure
		// Following calculations are undefined in this case
		return 0, 0, 0
	}
	// As of now, work with time rather than with indices
	tlow := p.Samples[ilow].Time
	thigh := p.Samples[ihigh].Time
	// Linear interpolation between tlow and thigh
	//fmt.Println("  -> ilow, ihigh, tlow, thigh, ampllow, amplhigh:", ilow, ihigh, tlow, thigh, ampllow, amplhigh)
	t := ((frac*p.Ampl-ampllow)*thigh + (amplhigh-frac*p.Ampl)*tlow) / (amplhigh - ampllow)
	//fmt.Println("  -> T30 =", T30)
	if t > thigh || t < tlow {
		panic("t > thigh || t < tlow")
	}
	return t, ilow, noLocMax
}

// CalcRisingFront returns various quantities calculated on the rising front:
//  - Time at 80%
//  - Time at 30%
//  - Time at 20%
//  - Number of local maxima on rising front, between 20% and 80%
func (p *Pulse) CalcRisingFront(recomputeAmpl bool) (float64, float64, float64, int) {
	time80, i80low, _ := p.T(recomputeAmpl, 0.8, p.AmplIndex)
	time30, i30low, noLocMax8030 := p.T(false, 0.3, i80low)
	time20, _, noLocMax3020 := p.T(false, 0.2, i30low)
	p.Time80 = time80
	p.Time30 = time30
	p.Time20 = time20
	p.NoLocMaxRisingFront = noLocMax8030 + noLocMax3020
	return p.Time80, p.Time30, p.Time20, p.NoLocMaxRisingFront
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
	Pulses        [4]Pulse
	ID            uint8
	CountersFifo1 []uint32 // Counters stores the counters present in the binary/hexa/decimal files
	CountersFifo2 []uint32 // Counters stores the counters present in the binary/hexa/decimal files
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
	fmt.Printf(" o Number of counters first fifo=%v\n", len(c.CountersFifo1))
	fmt.Printf(" o Number of counters second fifo=%v\n", len(c.CountersFifo2))
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
	amplitude := 0.
	for i := range c.Pulses {
		_, amp := c.Pulses[i].Amplitude()
		amplitude += amp
	}
	return amplitude
}

// Charge returns the sum of the pulse charges
func (c *Cluster) Charge(recalcPulsesCharges bool) float64 {
	charge := 0.
	for i := range c.Pulses {
		var pulseCharge float64
		switch recalcPulsesCharges {
		case true:
			pulseCharge = c.Pulses[i].Charge()
		case false:
			if c.Pulses[i].Charg == 0 {
				log.Printf("warning, c.Pulses[i].Charg == 0 (this is probably a mistake, fix it !)\n")
			}
			pulseCharge = c.Pulses[i].Charg
		}
		charge += pulseCharge
	}
	return charge
}

// XY return the X and Y coordinates of the incoming particle
// Calculated according a a formula found in
//
//     An Improved Multicrystal 2-D BGO Detector for PET J.G. Rogers
//     IEEE TRANSACTIONS ON NUCLEAR SCIENCE, VOL. 39. NO. 4,1992
//
// which is also available on our gitlab:
//     https://gitlab.in2p3.fr/avirm/Docs/blob/master/MiscPapers/Rogers_1992_IEEE.pdf
func (c *Cluster) XY(recalcPulsesCharges bool) (x, y float64) {
	var ch0, ch1, ch2, ch3 float64
	if c.Pulses[0].HasSignal {
		switch recalcPulsesCharges {
		case true:
			ch0 = c.Pulses[0].Charge()
		case false:
			ch0 = c.Pulses[0].Charg
		}
	}
	if c.Pulses[1].HasSignal {
		switch recalcPulsesCharges {
		case true:
			ch1 = c.Pulses[1].Charge()
		case false:
			ch1 = c.Pulses[1].Charg
		}
	}
	if c.Pulses[2].HasSignal {
		switch recalcPulsesCharges {
		case true:
			ch2 = c.Pulses[2].Charge()
		case false:
			ch2 = c.Pulses[2].Charg
		}
	}
	if c.Pulses[3].HasSignal {
		switch recalcPulsesCharges {
		case true:
			ch3 = c.Pulses[3].Charge()
		case false:
			ch3 = c.Pulses[3].Charg
		}
	}

	chTot := ch0 + ch1 + ch2 + ch3

	x = ((ch1 + ch3) - (ch0 + ch2)) / chTot
	y = ((ch1 + ch0) - (ch3 + ch2)) / chTot

	// 	fmt.Println("  -> ch0, ch1, ch2, ch3 =", ch0, ch1, ch2, ch3)
	// 	fmt.Println("  -> x, y =", x, y)

	return
}

// Counter return the value of the first fifo counter for the given index
func (c *Cluster) CounterFifo1(i int) uint32 {
	if i >= len(c.CountersFifo1) {
		log.Fatalf("pulse: counter index out of range")
	}
	return c.CountersFifo1[i]
}

// Counter return the value of the second fifo counter for the given index
func (c *Cluster) CounterFifo2(i int) uint32 {
	if i >= len(c.CountersFifo2) {
		log.Fatalf("pulse: counter index out of range")
	}
	return c.CountersFifo2[i]
}

type YRange byte

const (
	YRangeAuto YRange = iota
	YRangePedestal
	YRangeFullDynamics
)

// PlotPulses plots the four pulses of the cluster in one canvas
func (c *Cluster) PlotPulses(evtID uint, x XaxisType, yrange YRange, xRangeZoomAround500 bool) string {
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

	switch yrange {
	case YRangePedestal:
		p.Y.Min = 300
		p.Y.Max = 700
	case YRangeFullDynamics:
		p.Y.Min = -500
		p.Y.Max = 4096
	default:
		// do nothing, automatic range
	}

	if xRangeZoomAround500 {
		p.X.Min = 400
		p.X.Max = 600
	}

	outFile := "output/pulses_event" + strconv.Itoa(int(evtID)) + "_cluster" + strconv.Itoa(int(c.ID)) + ".png"

	if err := p.Save(14*vg.Inch, 5*vg.Inch, outFile); err != nil {
		panic(err)
	}

	return outFile
}
