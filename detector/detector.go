package detector

import (
	"bufio"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

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
	fmt.Printf("    - Capacitor: id = %v, pedestal mean = %v (address=%p)\n", c.id, c.pedestalMean, c)
}

func (c *Capacitor) ID() uint16 {
	return c.id
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
	c.pedestalStdDev = math.Sqrt(variance)
}

func (c *Capacitor) PedestalMean() float64 {
	return c.pedestalMean
}

type Channel struct {
	// A channel is made of 1024 capacitors
	capacitors [1024]Capacitor
	id         uint8
	name       string
	plotStat   bool
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

func (c *Channel) ID() uint8 {
	return c.id
}

func (c *Channel) Print() {
	fmt.Printf("  o Channel: id = %v (address=%p)\n", c.id, c)
	for i := range c.capacitors {
		c.capacitors[i].Print()
	}
}

type DRS struct {
	// A DRS is made of 8 channels (in fact 9 but the 9-th is not used)
	channels [8]Channel
	id       uint8
}

func (d *DRS) Print() {
	fmt.Printf(" * DRS: id = %v (address=%p)\n", d.id, d)
	for i := range d.channels {
		d.channels[i].Print()
	}
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

type TestBench struct {
	asm          ASMCard
	scintillator string
}

func (t *TestBench) Print() {
	fmt.Printf("Printing information for detector %v\n", t.scintillator)
	t.asm.Print()
}

func (t *TestBench) Capacitor(iDRS uint8, iChannel uint8, iCapacitor uint16) *Capacitor {
	return &t.asm.drss[iDRS].channels[iChannel].capacitors[iCapacitor]
}

func (t *TestBench) Channel(iDRS uint8, iChannel uint8) *Channel {
	return &t.asm.drss[iDRS].channels[iChannel]
}

func (t *TestBench) WritePedestalsToFile(outfileName string) {
	outFile, err := os.Create(outfileName)
	if err != nil {
		log.Fatalf("os.Create: %s", err)
	}
	defer func() {
		err = outFile.Close()
		if err != nil {
			log.Fatalf("error closing file %q: %v\n", outfileName, err)
		}
	}()

	w := bufio.NewWriter(outFile)
	defer func() {
		err = w.Flush()
		if err != nil {
			log.Fatalf("error flushing file %q: %v\n", outfileName, err)
		}
	}()

	fmt.Fprintf(w, "# Test bench pedestal file (creation date: %v)\n", time.Now())
	fmt.Fprintf(w, "# iDRS iChannel iCapacitor pedestalMean pedestalStdDev\n")

	for iDRS := range t.asm.drss {
		drs := &t.asm.drss[iDRS]
		for iChannel := range drs.channels {
			ch := &drs.channels[iChannel]
			for iCapacitor := range ch.capacitors {
				capa := &ch.capacitors[iCapacitor]
				fmt.Fprint(w, iDRS, iChannel, iCapacitor, capa.pedestalMean, capa.pedestalStdDev, "\n")
			}
		}
	}
}

func (t *TestBench) ComputePedestalsMeanStdDevFromSamples() {
	for iDRS := range t.asm.drss {
		drs := &t.asm.drss[iDRS]
		for iChannel := range drs.channels {
			ch := &drs.channels[iChannel]
			for iCapacitor := range ch.capacitors {
				capa := &ch.capacitors[iCapacitor]
				capa.ComputePedestalMeanStdDevFromSamples()
			}
		}
	}
}

func (t *TestBench) ReadPedestalsFile(fileName string) {
	file, err := os.Open(fileName)
	if err != nil {
		log.Fatalf("error opening file %v", err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		text := scanner.Text()
		if strings.HasPrefix(text, "#") {
			continue
		}
		fields := strings.Split(text, " ")
		if len(fields) != 5 {
			log.Fatalf("number of fields per line in file %v != 5", fileName)
		}
		iDRS, err := strconv.ParseUint(fields[0], 10, 8)
		if err != nil {
			log.Fatalf("error parsing %q: %v\n", text, err)
		}
		iChannel, err := strconv.ParseUint(fields[1], 10, 8)
		if err != nil {
			log.Fatalf("error parsing %q: %v\n", text, err)
		}
		iCapacitor, err := strconv.ParseUint(fields[2], 10, 16)
		if err != nil {
			log.Fatalf("error parsing %q: %v\n", text, err)
		}
		pedestalMean, err := strconv.ParseFloat(fields[3], 64)
		if err != nil {
			log.Fatalf("error parsing %q: %v\n", text, err)
		}
		pedestalVariance, err := strconv.ParseFloat(fields[4], 64)
		if err != nil {
			log.Fatalf("error parsing %q: %v\n", text, err)
		}
		capacitor := t.Capacitor(uint8(iDRS), uint8(iChannel), uint16(iCapacitor))
		capacitor.pedestalMean = pedestalMean
		capacitor.pedestalStdDev = pedestalVariance
	}
}

func (t *TestBench) PlotPedestals(iDRS uint8, plotStat bool) {
	for i := uint8(0); i < 8; i++ {
		t.Channel(iDRS, i).plotStat = plotStat
	}
	channel0 := t.Channel(iDRS, 0)
	channel1 := t.Channel(iDRS, 1)
	channel2 := t.Channel(iDRS, 2)
	channel3 := t.Channel(iDRS, 3)
	channel4 := t.Channel(iDRS, 4)
	channel5 := t.Channel(iDRS, 5)
	channel6 := t.Channel(iDRS, 6)
	channel7 := t.Channel(iDRS, 7)

	p, err := plot.New()
	if err != nil {
		panic(err)
	}

	p.Title.Text = "Pedestal"
	p.X.Label.Text = "capacitor"
	switch plotStat {
	case false:
		p.Y.Label.Text = "mean +- variance"
	case true:
		p.Y.Label.Text = "number of samples"
	}
	p.Add(plotter.NewGrid())

	err = plotutil.AddScatters(p,
		channel0.name, channel0,
		channel1.name, channel1,
		channel2.name, channel2,
		channel3.name, channel3,
		channel4.name, channel4,
		channel5.name, channel5,
		channel6.name, channel6,
		channel7.name, channel7)

	if err != nil {
		panic(err)
	}

	err = plotutil.AddErrorBars(p,
		channel0,
		channel1,
		channel2,
		channel3,
		channel4,
		channel5,
		channel6,
		channel7)

	if err != nil {
		panic(err)
	}

	outFile := "output/pedestal"
	if plotStat {
		outFile += "Stat"
	}
	outFile += ".pdf"
	if err := p.Save(14*vg.Inch, 5*vg.Inch, outFile); err != nil {
		panic(err)
	}
}

func NewTestBenchDetector() *TestBench {
	det := &TestBench{
		scintillator: "LYSO",
	}
	asm := &det.asm
	asm.id = 0
	for iDRS := range asm.drss {
		drs := &asm.drss[iDRS]
		drs.id = uint8(iDRS)
		for iChannel := range drs.channels {
			ch := &drs.channels[iChannel]
			ch.name = "PMT" + strconv.FormatUint(uint64(iChannel), 10)
			ch.id = uint8(iChannel)
			for iCapacitor := range ch.capacitors {
				capa := &ch.capacitors[iCapacitor]
				capa.id = uint16(iCapacitor)
			}
		}
	}
	return det
}

var TBDet *TestBench

func init() {
	TBDet = NewTestBenchDetector()
}
