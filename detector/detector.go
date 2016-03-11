// Package detector describes the physical structure of the detector and its electronics.
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

// Capacitor describes a DRS capacitor.
type Capacitor struct {
	id              uint16 // possible values: 0, 1, ..., 1023 (1024 capacitors per channel)
	pedestalSamples []float64
	pedestalMean    float64
	pedestalStdDev  float64
}

// Print prints capacitor informations.
func (c *Capacitor) Print() {
	fmt.Printf("    # Capacitor: id = %v, pedestal mean = %v, stddev = %v (address=%p)\n", c.id, c.pedestalMean, c.pedestalStdDev, c)
}

// ID returns the capacitor id.
func (c *Capacitor) ID() uint16 {
	return c.id
}

// SetID sets capacitor id to the given value.
func (c *Capacitor) SetID(id uint16) {
	c.id = id
}

// NoPedestalSamples returns the number of samples used for the pedestal calculation.
func (c *Capacitor) NoPedestalSamples() int {
	return len(c.pedestalSamples)
}

// AddPedestalSample adds a sample for pedestal calculation.
func (c *Capacitor) AddPedestalSample(n float64) {
	c.pedestalSamples = append(c.pedestalSamples, n)
}

// ComputePedestalMeanStdDevFromSamples computes the pedestal mean and standard deviation on mean.
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

// PedestalMean returns the pedestal mean.
func (c *Capacitor) PedestalMean() float64 {
	return c.pedestalMean
}

// PedestalStdDev returns the pedestal standard deviation.
func (c *Capacitor) PedestalStdDev() float64 {
	return c.pedestalStdDev
}

// SetPedestalMeanStdDev sets the pedestal mean and standard deviation to the given values.
func (c *Capacitor) SetPedestalMeanStdDev(mean float64, stddev float64) {
	c.pedestalMean = mean
	c.pedestalStdDev = stddev
}

// Coord describes the coordinates of a scintillator cristal.
// The coordinates are those of the center of the front face of the cristal.
type Coord struct {
	x, y, z float64
}

// Channel describes a DRS channel.
// A channel is made of 1024 capacitors.
type Channel struct {
	capacitors [1024]Capacitor
	id         uint8  // relative id: 0 -> 3 (because there are 4 channels per quartet)
	absid288   uint16 // absolute id: 0 -> 287 for DPGA, irrelevant for test bench
	absid240   uint16 // absolute id: 0 -> 239 for DPGA, irrelevant for test bench
	fifoid     uint16 // fifo id: 0 -> 143 for DPGA
	name       string
	coord      Coord
	plotStat   bool
}

// Capacitors returns the capacitors for this channel
func (c *Channel) Capacitors() [1024]Capacitor {
	return c.capacitors
}

// Capacitor return a pointer to the capacitor corresponding to the given index
func (c *Channel) Capacitor(iCapacitor uint16) *Capacitor {
	return &c.capacitors[iCapacitor]
}

// PlotStat should be used to set Channel.plotStat to true (by default to false).
// When set to true, the number of samples used to compute pedestals is plotted rather
// the pedestal mean and standard deviation of the pedestals.
func (c *Channel) PlotStat(plotStat bool) {
	c.plotStat = plotStat
}

// Len return the number of capacitors.
// It implements gonum/plot/plotter/XYer interface
func (c *Channel) Len() int {
	return len(c.capacitors)
}

// XY return an x, y pair for the capacitor corresponding to the given index
// It implements gonum/plot/plotter/XYer interface
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

// YError returns a pair of errors corresponding respectively to the downward and upward error of the pedestal mean.
// It implements the gonum/plot/plotter/YErrorer interface.
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

// XError returns a pair of errors in the x direction.
// It implements the gonum/plot/plotter/XErrorer interface.
func (c *Channel) XError(iCapacitor int) (down float64, up float64) {
	return 0, 0
}

// Name returns the name of the channel.
func (c *Channel) Name() string {
	return c.name
}

// SetName sets the channel name to the given value.
func (c *Channel) SetName(name string) {
	c.name = name
}

// ID returns the channel id.
func (c *Channel) ID() uint8 {
	return c.id
}

// SetID sets the channel id to the given value.
func (c *Channel) SetID(id uint8) {
	c.id = id
}

// AbsID288 returns the channel absolute id (0 -> 287).
func (c *Channel) AbsID288() uint16 {
	return c.absid288
}

// SetAbsID288 sets the channel absolute id (0 -> 287).
func (c *Channel) SetAbsID288(id uint16) {
	c.absid288 = id
}

// AbsID240 returns the channel absolute id (0 -> 239).
func (c *Channel) AbsID240() uint16 {
	return c.absid240
}

// SetAbsID240 sets the channel absolute id (0 -> 239).
func (c *Channel) SetAbsID240(id uint16) {
	c.absid240 = id
}

// FifoID returns the fifo id.
func (c *Channel) FifoID() uint16 {
	return c.fifoid
}

// SetFifoID sets the fifo id.
func (c *Channel) SetFifoID(id uint16) {
	c.fifoid = id
}

// SetCoord sets the cartesian coordinates for this channel.
func (c *Channel) SetCoord(x, y, z float64) {
	c.coord.x = x
	c.coord.y = y
	c.coord.z = z
}

// Print print channel informations.
func (c *Channel) Print() {
	fmt.Printf("   o Channel: id = %v absid288 = %v absid240 = %v coord = (%v, %v, %v) (address=%p)\n", c.id, c.absid288, c.absid240, c.coord.x, c.coord.y, c.coord.z, c)
	for i := range c.capacitors {
		c.capacitors[i].Print()
	}
}

// Quartet describes a quartet.
// A quartet is made of 4 channels.
type Quartet struct {
	channels [4]Channel
	id       uint8
}

// SetID sets the quartet id.
func (q *Quartet) SetID(id uint8) {
	q.id = id
}

// ID returns the quartet id.
func (q *Quartet) ID() uint8 {
	return q.id
}

// Print prints quartet informations.
func (q *Quartet) Print() {
	fmt.Printf("  - Quartet: id= %v (address=%p)\n", q.id, q)
	for i := range q.channels {
		q.channels[i].Print()
	}
}

// Channels returns an array with the four channels for this quartet.
func (q *Quartet) Channels() [4]Channel {
	return q.channels
}

// Channel returns the channel corresponding to the given index.
func (q *Quartet) Channel(iChannel uint8) *Channel {
	return &q.channels[iChannel]
}

// PlotPedestals plots pedestals for this quartet.
// If plotStat is true, the statistics used to compute pedestals are plotted.
// Otherwise the pedestals are plotted.
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

// DRS describes a DRS chip of the ASM cards.
// A DRS treats signals of 2 quartets.
type DRS struct {
	// A DRS is made of 8 channels (in fact 9 but the 9-th is not used)
	// The first four and last four channels correspond to two different quartets
	quartets [2]Quartet
	id       uint8
}

// Print prints DRS informations.
func (d *DRS) Print() {
	fmt.Printf(" * DRS: id = %v (address=%p)\n", d.id, d)
	for i := range d.quartets {
		d.quartets[i].Print()
	}
}

// SetID sets the DRS id.
func (d *DRS) SetID(id uint8) {
	d.id = id
}

// Quartets returns an array with the 2 quartets of this DRS.
func (d *DRS) Quartets() [2]Quartet {
	return d.quartets
}

// Quartet returns the quartet corresponding to the given index.
func (d *DRS) Quartet(iQuartet uint8) *Quartet {
	return &d.quartets[iQuartet]
}

// ASMCard describes an ASM card.
// An ASMCard is made of 3 DRS.
type ASMCard struct {
	// An ASM card is made of 3 DRS, one on each mezzanine
	drss [3]DRS
	id   uint8
}

// Print prints the ASMCard informations.
func (a *ASMCard) Print() {
	fmt.Printf("+ ASM: id = %v (address=%p)\n", a.id, a)
	for i := range a.drss {
		a.drss[i].Print()
	}
}

// DRSs returns an array with the 3 DRS for this ASMCard.
func (a *ASMCard) DRSs() [3]DRS {
	return a.drss
}

// DRS returns the DRS corresponding to the given index.
func (a *ASMCard) DRS(iDRS uint8) *DRS {
	return &a.drss[iDRS]
}
