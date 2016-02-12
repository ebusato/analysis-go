package detector

import (
	"fmt"
	"log"
	"math"

	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/plotutil"
	"github.com/gonum/plot/vg"
	"github.com/gonum/stat"
)

type Capacitor struct {
	id              uint16 // possible values: 0, 1, ..., 1023 (1024 capacitors per channel)
	pedestalSamples []float64
	pedestalMean    float64
	pedestalStdDev  float64
}

func (c *Capacitor) Print() {
	fmt.Printf("    # Capacitor: id = %v, pedestal mean = %v, stddev = %v (address=%p)\n", c.id, c.pedestalMean, c.pedestalStdDev, c)
}

func (c *Capacitor) ID() uint16 {
	return c.id
}

func (c *Capacitor) SetID(id uint16) {
	c.id = id
}

func (c *Capacitor) NoPedestalSamples() int {
	return len(c.pedestalSamples)
}

func (c *Capacitor) AddPedestalSample(n float64) {
	c.pedestalSamples = append(c.pedestalSamples, n)
}

func (c *Capacitor) ComputePedestalMeanStdDevFromSamples() {
	var weights []float64
	mean, variance := stat.MeanVariance(c.pedestalSamples, weights)
	if math.IsNaN(mean) {
		mean = 0
	}
	if math.IsNaN(variance) {
		variance = 0
	}
	c.pedestalMean = mean
	switch c.NoPedestalSamples() != 0 {
	case true:
		c.pedestalStdDev = math.Sqrt(variance / float64(c.NoPedestalSamples()))
	default:
		c.pedestalStdDev = 0
	}
}

func (c *Capacitor) PedestalMean() float64 {
	return c.pedestalMean
}

func (c *Capacitor) PedestalStdDev() float64 {
	return c.pedestalStdDev
}

func (c *Capacitor) SetPedestalMeanStdDev(mean float64, stddev float64) {
	c.pedestalMean = mean
	c.pedestalStdDev = stddev
}

type Coord struct {
	x, y, z float64
}

type Channel struct {
	// A channel is made of 1024 capacitors
	capacitors [1024]Capacitor
	id         uint8  // relative id: 0 -> 3 (because there are 4 channels per quartet)
	absid      uint16 // absolute id: 0 -> 239 for DPGA, 0 -> 3 for TestBench
	name       string
	coord      Coord
	plotStat   bool
}

func (c *Channel) Capacitors() [1024]Capacitor {
	return c.capacitors
}

func (c *Channel) Capacitor(iCapacitor uint16) *Capacitor {
	return &c.capacitors[iCapacitor]
}

func (c *Channel) PlotStat(plotStat bool) {
	c.plotStat = plotStat
}

// implement gonum/plot/plotter/XYer interface
func (c *Channel) Len() int {
	return len(c.capacitors)
}

func (c *Channel) XY(iCapacitor int) (x, y float64) {
	if iCapacitor > 1023 {
		log.Fatal("iCapacitor out of range")
	}
	capacitor := &c.capacitors[iCapacitor]
	x = float64(capacitor.id)

	switch c.plotStat {
	case false:
		y = capacitor.pedestalMean
	case true:
		y = float64(capacitor.NoPedestalSamples())

	}
	return
}

// implement gonum/plot/plotter/YErrorer interface
func (c *Channel) YError(iCapacitor int) (down float64, up float64) {
	if iCapacitor > 1023 {
		log.Fatal("iCapacitor out of range")
	}
	capacitor := &c.capacitors[iCapacitor]
	switch c.plotStat {
	case false:
		down = capacitor.pedestalStdDev
		up = down
	case true:
		down = 0
		up = 0
	}
	return down, up
}

// implement gonum/plot/plotter/XErrorer interface
func (c *Channel) XError(iCapacitor int) (down float64, up float64) {
	return 0, 0
}

func (c *Channel) Name() string {
	return c.name
}

func (c *Channel) SetName(name string) {
	c.name = name
}

func (c *Channel) ID() uint8 {
	return c.id
}

func (c *Channel) SetID(id uint8) {
	c.id = id
}

func (c *Channel) AbsID() uint16 {
	return c.absid
}

func (c *Channel) SetAbsID(id uint16) {
	c.absid = id
}

func (c *Channel) Print() {
	fmt.Printf("   o Channel: id = %v absid = %v (address=%p)\n", c.id, c.absid, c)
	for i := range c.capacitors {
		c.capacitors[i].Print()
	}
}

type Quartet struct {
	channels [4]Channel
	id       uint8
}

func (q *Quartet) SetID(id uint8) {
	q.id = id
}

func (q *Quartet) ID() uint8 {
	return q.id
}

func (q *Quartet) Print() {
	fmt.Printf("  - Quartet: id= %v (address=%p)\n", q.id, q)
	for i := range q.channels {
		q.channels[i].Print()
	}
}

func (q *Quartet) Channels() [4]Channel {
	return q.channels
}

func (q *Quartet) Channel(iChannel uint8) *Channel {
	return &q.channels[iChannel]
}

func (q *Quartet) PlotPedestals(plotStat bool, text string) {
	for i := range q.channels {
		q.channels[i].PlotStat(plotStat)
	}

	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	p.Title.Text = "Pedestal " + text
	p.X.Label.Text = "capacitor"
	switch plotStat {
	case false:
		p.Y.Label.Text = "mean +- variance"
	case true:
		p.Y.Label.Text = "number of samples"
	}
	p.Add(plotter.NewGrid())

	err = plotutil.AddLinePoints(p, //AddScatters(p,
		q.channels[0].Name(), &q.channels[0],
		q.channels[1].Name(), &q.channels[1],
		q.channels[2].Name(), &q.channels[2],
		q.channels[3].Name(), &q.channels[3])

	if err != nil {
		panic(err)
	}

	err = plotutil.AddErrorBars(p,
		&q.channels[0],
		&q.channels[1],
		&q.channels[2],
		&q.channels[3])

	if err != nil {
		panic(err)
	}

	//outFile := "output/pedestal_quartet_" + strconv.FormatUint(uint64(q.id), 10)
	outFile := "output/pedestal_quartet_" + text
	if plotStat {
		outFile += "_Stat"
	}
	outFile += ".png"
	if err := p.Save(14*vg.Inch, 5*vg.Inch, outFile); err != nil {
		panic(err)
	}
}

type DRS struct {
	// A DRS is made of 8 channels (in fact 9 but the 9-th is not used)
	// The first four and last four channels correspond to two different quartets
	quartets [2]Quartet
	id       uint8
}

func (d *DRS) Print() {
	fmt.Printf(" * DRS: id = %v (address=%p)\n", d.id, d)
	for i := range d.quartets {
		d.quartets[i].Print()
	}
}

func (d *DRS) SetID(id uint8) {
	d.id = id
}

func (d *DRS) Quartets() [2]Quartet {
	return d.quartets
}

func (d *DRS) Quartet(iQuartet uint8) *Quartet {
	return &d.quartets[iQuartet]
}

type ASMCard struct {
	// An ASM card is made of 3 DRS, one on each mezzanine
	drss [3]DRS
	id   uint8
}

func (a *ASMCard) Print() {
	fmt.Printf("+ ASM: id = %v (address=%p)\n", a.id, a)
	for i := range a.drss {
		a.drss[i].Print()
	}
}

func (a *ASMCard) DRSs() [3]DRS {
	return a.drss
}

func (a *ASMCard) DRS(iDRS uint8) *DRS {
	return &a.drss[iDRS]
}
