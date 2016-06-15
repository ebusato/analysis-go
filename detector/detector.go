// Package detector describes the physical structure of the detector and its electronics.
package detector

import (
	"fmt"
	"log"
	"math"

	"gitlab.in2p3.fr/avirm/analysis-go/utils"

	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/plotutil"
	"github.com/gonum/plot/vg"
)

// Capacitor describes a DRS capacitor.
type Capacitor struct {
	id                uint16 // possible values: 0, 1, ..., 1023 (1024 capacitors per channel)
	noPedestalSamples uint64
	pedestalMean      float64
	pedestalMeanErr   float64
	Channel           *Channel
}

// Print prints capacitor informations.
func (c *Capacitor) Print() {
	fmt.Printf("    # Capacitor: id = %v, pedestal mean = %v, stddev = %v (address=%p)\n", c.id, c.pedestalMean, c.pedestalMeanErr, c)
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
func (c *Capacitor) NoPedestalSamples() uint64 {
	return c.noPedestalSamples
}

// AddPedestalSample adds a sample for pedestal calculation.
func (c *Capacitor) AddPedestalSample(ampl float64) {
	c.noPedestalSamples++
	c.pedestalMean += ampl
	c.pedestalMeanErr += ampl * ampl
}

// FinalizePedestalMeanErr computes the pedestal sample mean and standard deviation on mean.
func (c *Capacitor) FinalizePedestalMeanErr() {
	if c.noPedestalSamples != 0 && c.noPedestalSamples != 1 {
		c.pedestalMean /= float64(c.noPedestalSamples)
		c.pedestalMeanErr /= float64(c.noPedestalSamples) - 1
		c.pedestalMeanErr -= float64(c.noPedestalSamples) / (float64(c.noPedestalSamples) - 1) * c.pedestalMean * c.pedestalMean
		c.pedestalMeanErr = math.Sqrt(c.pedestalMeanErr / float64(c.noPedestalSamples))
	}
}

// ComputePedestalMeanErrFromSamples computes the pedestal mean and standard deviation on mean.
// func (c *Capacitor) ComputePedestalMeanErrFromSamples() {
// 	var weights []float64
// 	mean, variance := stat.MeanVariance(c.pedestalSamples, weights)
// 	if math.IsNaN(mean) {
// 		mean = 0
// 	}
// 	if math.IsNaN(variance) {
// 		variance = 0
// 	}
// 	c.pedestalMean = mean
// 	switch c.NoPedestalSamples() != 0 {
// 	case true:
// 		c.pedestalMeanErr = math.Sqrt(variance / float64(c.NoPedestalSamples()))
// 	default:
// 		c.pedestalMeanErr = 0
// 	}
// }

// PedestalMean returns the pedestal mean.
func (c *Capacitor) PedestalMean() float64 {
	return c.pedestalMean
}

// PedestalMeanErr returns the pedestal standard deviation.
func (c *Capacitor) PedestalMeanErr() float64 {
	return c.pedestalMeanErr
}

// SetPedestalMeanErr sets the pedestal mean and standard deviation to the given values.
func (c *Capacitor) SetPedestalMeanErr(mean float64, err float64) {
	c.pedestalMean = mean
	c.pedestalMeanErr = err
}

// RectParallelepiped describes the geometry of the scintillator used in the detector (which
// are rectangular parallelepiped). It's nothing but an array of size 8, each component beeing
// the cartesian coordinates of one corner of the rectangular parallelepiped.
type RectParallelepiped [8]utils.CartCoord

// Channel describes a DRS channel.
// A channel is made of 1024 capacitors.
type Channel struct {
	capacitors [1024]Capacitor
	id         uint8  // relative id: 0 -> 3 (because there are 4 channels per quartet)
	absid288   uint16 // absolute id: 0 -> 287 for DPGA, irrelevant for test bench
	absid240   uint16 // absolute id: 0 -> 239 for DPGA, irrelevant for test bench
	fifoid144  uint16 // fifo id: 0 -> 143 for DPGA
	name       string
	plotStat   bool

	// The following three quantities are used for the the determination and usage of
	// the time dependent offset calibration
	noTimeDepOffsetSamples uint64    // number of samples used to compute the time dependent offsets
	timeDepOffsetMean      []float64 // slice containing time dependent offset calibration coefficients (to be applied after per-capacitor pedestal correction has been applied)
	timeDepOffsetMeanErr   []float64 // slice containing time dependent offset calibration coefficient errors

	// The coordinates are those of the center of the front face of the cristal.
	utils.CartCoord

	// Full coordinates of the scintillator
	ScintCoords RectParallelepiped

	Quartet *Quartet
}

// SetNoSamples sets the size of the timeDepOffsetMean and timeDepOffsetMeanErr slices to the number of samples
func (c *Channel) SetNoSamples(noSamples int) {
	c.timeDepOffsetMean = make([]float64, noSamples)
	c.timeDepOffsetMeanErr = make([]float64, noSamples)
}

// IncrementNoTimeDepOffsetSamples increments c.noTimeDepOffsetSamples by one unit
func (c *Channel) IncrementNoTimeDepOffsetSamples() {
	c.noTimeDepOffsetSamples++
}

// AddTimeDepOffsetSample adds a sample for the calculation of the time dependent offset mean and error
func (c *Channel) AddTimeDepOffsetSample(iSample int, ampl float64) {
	c.timeDepOffsetMean[iSample] += ampl
	c.timeDepOffsetMeanErr[iSample] += ampl * ampl
}

// FinalizePedestalMeanErr computes the pedestal sample mean and standard deviation on mean.
func (c *Channel) FinalizeTimeDepOffsetMeanErr() {
	if c.noTimeDepOffsetSamples != 0 && c.noTimeDepOffsetSamples != 1 {
		for iSample := 0; iSample < len(c.timeDepOffsetMean); iSample++ {
			c.timeDepOffsetMean[iSample] /= float64(c.noTimeDepOffsetSamples)
			c.timeDepOffsetMeanErr[iSample] /= float64(c.noTimeDepOffsetSamples) - 1
			c.timeDepOffsetMeanErr[iSample] -= float64(c.noTimeDepOffsetSamples) / (float64(c.noTimeDepOffsetSamples) - 1) * c.timeDepOffsetMean[iSample] * c.timeDepOffsetMean[iSample]
			c.timeDepOffsetMeanErr[iSample] = math.Sqrt(c.timeDepOffsetMeanErr[iSample] / float64(c.noTimeDepOffsetSamples))
		}
	}
}

// TimeDepOffsetMeans returns the time dependent offset slice.
func (c *Channel) TimeDepOffsetMeans() []float64 {
	return c.timeDepOffsetMean
}

// TimeDepOffsetMean returns the time dependent offset mean for sample iSample.
func (c *Channel) TimeDepOffsetMean(iSample int) float64 {
	return c.timeDepOffsetMean[iSample]
}

// PedestalMeanErr returns the pedestal standard deviation for sample iSample.
func (c *Channel) TimeDepOffsetMeanErr(iSample int) float64 {
	return c.timeDepOffsetMeanErr[iSample]
}

// SetTimDepOffsetMeanErr sets the time dependent offset mean and standard deviation to the given values for sample iSample.
func (c *Channel) SetTimeDepOffsetMeanErr(iSample int, mean float64, err float64) {
	c.timeDepOffsetMean[iSample] = mean
	c.timeDepOffsetMeanErr[iSample] = err
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
		down = capacitor.pedestalMeanErr
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
func (c *Channel) FifoID144() uint16 {
	return c.fifoid144
}

// SetFifoID144 sets the fifo id. (from 0 to 143)
func (c *Channel) SetFifoID144(id uint16) {
	c.fifoid144 = id
}

// SetCartCoord sets the cartesian coordinates for this channel.
// Coordinates are those of the center of the PMT at its front surface.
func (c *Channel) SetCartCoord(x, y, z float64) {
	c.X = x
	c.Y = y
	c.Z = z
}

// Print print channel informations.
func (c *Channel) Print() {
	fmt.Printf("   o Channel: id = %v absid288 = %v absid240 = %v coord = (%v, %v, %v) (address=%p)\n", c.id, c.absid288, c.absid240, c.X, c.Y, c.Z, c)
	for i := range c.capacitors {
		c.capacitors[i].Print()
	}
}

// Quartet describes a quartet.
// A quartet is made of 4 channels.
type Quartet struct {
	channels [4]Channel
	id       uint8

	utils.CylCoord

	DRS *DRS
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

// SetCylCoord sets the cylindrical coordinates for this quartet.
// Coordinates are those of the center of the quartet at its front.
// r is in mm
// phi is in radian
// z is in mm
func (q *Quartet) SetCylCoord(r, phi, z float64) {
	q.R = r
	q.Phi = phi
	q.Z = z
}

// DRS describes a DRS chip of the ASM cards.
// A DRS treats signals of 2 quartets.
type DRS struct {
	// A DRS is made of 8 channels (in fact 9 but the 9-th is not used)
	// The first four and last four channels correspond to two different quartets
	quartets [2]Quartet
	id       uint8

	ASMCard *ASMCard
}

// Print prints DRS informations.
func (d *DRS) Print() {
	fmt.Printf(" * DRS: id = %v (address=%p)\n", d.id, d)
	for i := range d.quartets {
		d.quartets[i].Print()
	}
}

// ID returns the DRS id.
func (d *DRS) ID() uint8 {
	return d.id
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

// UpperStruct is an interface to the upper structure to which
// the ASM card belong in the full detector.
// For example, for the LAPD detector UpperStruct is a hemisphere
type UpperStruct interface{}

// ASMCard describes an ASM card.
// An ASMCard is made of 3 DRS, one on each mezzanine.
// In the DPGA, an ASM card processes signals from one line of the detector.
// There is thus a one-to-one mapping between ASM cards and detectors lines (cassettes).
// A line of the DPGA can be described by its cylindrical coordinates, hence the embedding of utils.CylCoord.
type ASMCard struct {
	drss [3]DRS
	id   uint8

	utils.CylCoord

	UpStr UpperStruct
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

// SetCylCoord sets the cylindrical coordinates for this ASM card.
// Coordinates are those of the center of its first quartet (i.e. the one closest to the front side of the DPGA).
// r is in mm
// phi is in radian
// z is in mm
func (a *ASMCard) SetCylCoord(r, phi, z float64) {
	a.R = r
	a.Phi = phi
	a.Z = z
}
