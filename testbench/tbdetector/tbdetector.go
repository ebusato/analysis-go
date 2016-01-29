package tbdetector

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"gitlab.in2p3.fr/AVIRM/Analysis-go/detector"
)

type Detector struct {
	asm          detector.ASMCard
	scintillator string
	samplingFreq float64 // sampling frequency in ns
}

func NewDetector() *Detector {
	det := &Detector{
		scintillator: "LYSO",
		samplingFreq: 0.200,
	}
	for iDRS := range det.asm.DRSs() {
		drs := det.asm.DRS(uint8(iDRS))
		drs.SetID(uint8(iDRS))
		for iQuartet := range drs.Quartets() {
			quartet := drs.Quartet(uint8(iQuartet))
			quartet.SetID(uint8(iQuartet))
			for iChannel := range quartet.Channels() {
				ch := quartet.Channel(uint8(iChannel))
				ch.SetName("PMT" + strconv.FormatUint(uint64(iChannel), 10))
				ch.SetID(uint8(iChannel))
				for iCapacitor := range ch.Capacitors() {
					capa := ch.Capacitor(uint16(iCapacitor))
					capa.SetID(uint16(iCapacitor))
				}
			}
		}
	}
	return det
}

func (d *Detector) Print() {
	fmt.Printf("Printing information for detector %v\n", d.scintillator)
	d.asm.Print()
}

func (d *Detector) SamplingFreq() float64 {
	return d.samplingFreq
}

func (d *Detector) Capacitor(iChannel uint8, iCapacitor uint16) *detector.Capacitor {
	return d.asm.DRS(0).Quartet(0).Channel(iChannel).Capacitor(iCapacitor)
}

func (d *Detector) Channel(iChannel uint8) *detector.Channel {
	return d.asm.DRS(0).Quartet(0).Channel(iChannel)
}

func (d *Detector) Quartet() *detector.Quartet {
	return d.asm.DRS(0).Quartet(0)
}

func (d *Detector) WritePedestalsToFile(outfileName string) {
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
	fmt.Fprintf(w, "# iChannel iCapacitor pedestalMean pedestalStdDev\n")

	for iChannel := range d.Quartet().Channels() {
		quartet := d.Quartet()
		ch := quartet.Channel(uint8(iChannel))
		for iCapacitor := range ch.Capacitors() {
			capa := ch.Capacitor(uint16(iCapacitor))
			fmt.Fprint(w, iChannel, iCapacitor, capa.PedestalMean(), capa.PedestalStdDev(), "\n")

		}

	}
}

func (d *Detector) ComputePedestalsMeanStdDevFromSamples() {
	for iDRS := range d.asm.DRSs() {
		drs := d.asm.DRS(uint8(iDRS))
		for iQuartet := range drs.Quartets() {
			quartet := drs.Quartet(uint8(iQuartet))
			for iChannel := range quartet.Channels() {
				ch := quartet.Channel(uint8(iChannel))
				for iCapacitor := range ch.Capacitors() {
					capa := ch.Capacitor(uint16(iCapacitor))
					capa.ComputePedestalMeanStdDevFromSamples()
				}
			}
		}
	}
}

func (d *Detector) ReadPedestalsFile(fileName string) {
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
		if len(fields) != 4 {
			log.Fatalf("number of fields per line in file %v != 5", fileName)
		}
		iChannel, err := strconv.ParseUint(fields[0], 10, 8)
		if err != nil {
			log.Fatalf("error parsing %q: %v\n", text, err)
		}
		iCapacitor, err := strconv.ParseUint(fields[1], 10, 16)
		if err != nil {
			log.Fatalf("error parsing %q: %v\n", text, err)
		}
		pedestalMean, err := strconv.ParseFloat(fields[2], 64)
		if err != nil {
			log.Fatalf("error parsing %q: %v\n", text, err)
		}
		pedestalVariance, err := strconv.ParseFloat(fields[3], 64)
		if err != nil {
			log.Fatalf("error parsing %q: %v\n", text, err)
		}
		capacitor := d.Capacitor(uint8(iChannel), uint16(iCapacitor))
		capacitor.SetPedestalMeanStdDev(pedestalMean, pedestalVariance)
	}
}

func (d *Detector) PlotPedestals(plotStat bool) {
	d.Quartet().PlotPedestals(plotStat)
}

var Det *Detector

func init() {
	Det = NewDetector()
}
