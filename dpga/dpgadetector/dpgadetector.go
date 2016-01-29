package dpgadetector

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

type HemisphereType byte

const (
	rightHemisphere HemisphereType = iota
	leftHemisphere
)

type Hemisphere struct {
	asm   [6]detector.ASMCard
	which HemisphereType
}

func (h *Hemisphere) Print() {
	fmt.Printf("Printing Hemisphere %v\n", h.which)
	for iASM := range h.asm {
		h.asm[iASM].Print()
	}
}

type Detector struct {
	scintillator string
	samplingFreq float64 // sampling frequency in ns
	hemisphere   [2]Hemisphere
}

func NewDetector() *Detector {
	det := &Detector{
		scintillator: "LYSO",
		samplingFreq: 0.200,
	}
	for iHemi := range det.hemisphere {
		hemi := &det.hemisphere[iHemi]
		for iASM := range hemi.asm {
			asm := &hemi.asm[iASM]
			for iDRS := range asm.DRSs() {
				drs := asm.DRS(uint8(iDRS))
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
		}
	}
	return det
}

func (d *Detector) Print() {
	fmt.Printf("Printing information for detector %v (sampling freq = %v)\n", d.scintillator, d.samplingFreq)
	for iHemi := range d.hemisphere {
		d.hemisphere[iHemi].Print()
	}
}

func (d *Detector) SamplingFreq() float64 {
	return d.samplingFreq
}

func (d *Detector) Capacitor(iHemi uint8, iASM uint8, iDRS uint8, iQuartet uint8, iChannel uint8, iCapacitor uint16) *detector.Capacitor {
	return d.hemisphere[int(iHemi)].asm[int(iASM)].DRS(iDRS).Quartet(iQuartet).Channel(iChannel).Capacitor(iCapacitor)
}

func (d *Detector) Channel(iHemi uint8, iASM uint8, iDRS uint8, iQuartet uint8, iChannel uint8) *detector.Channel {
	return d.hemisphere[int(iHemi)].asm[int(iASM)].DRS(iDRS).Quartet(iQuartet).Channel(iChannel)
}

func (d *Detector) Quartet(iHemi uint8, iASM uint8, iDRS uint8, iQuartet uint8) *detector.Quartet {
	return d.hemisphere[int(iHemi)].asm[int(iASM)].DRS(iDRS).Quartet(iQuartet)
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
	fmt.Fprintf(w, "# iHemi iASM iDRS iQuartet iChannel iCapacitor pedestalMean pedestalStdDev\n")

	for iHemi := range d.hemisphere {
		hemi := &d.hemisphere[iHemi]
		for iASM := range hemi.asm {
			asm := &hemi.asm[iASM]
			for iDRS := range asm.DRSs() {
				drs := asm.DRS(uint8(iDRS))
				for iQuartet := range drs.Quartets() {
					quartet := drs.Quartet(uint8(iQuartet))
					for iChannel := range quartet.Channels() {
						ch := quartet.Channel(uint8(iChannel))
						for iCapacitor := range ch.Capacitors() {
							capa := ch.Capacitor(uint16(iCapacitor))
							fmt.Fprint(w, iChannel, iCapacitor, capa.PedestalMean(), capa.PedestalStdDev(), "\n")
						}
					}
				}
			}
		}
	}
}

func (d *Detector) ComputePedestalsMeanStdDevFromSamples() {
	for iHemi := range d.hemisphere {
		hemi := &d.hemisphere[iHemi]
		for iASM := range hemi.asm {
			asm := &hemi.asm[iASM]
			for iDRS := range asm.DRSs() {
				drs := asm.DRS(uint8(iDRS))
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
		if len(fields) != 8 {
			log.Fatalf("number of fields per line in file %v != 5", fileName)
		}
		iHemi, err := strconv.ParseUint(fields[0], 10, 8)
		if err != nil {
			log.Fatalf("error parsing %q: %v\n", text, err)
		}
		iASM, err := strconv.ParseUint(fields[1], 10, 8)
		if err != nil {
			log.Fatalf("error parsing %q: %v\n", text, err)
		}
		iDRS, err := strconv.ParseUint(fields[2], 10, 8)
		if err != nil {
			log.Fatalf("error parsing %q: %v\n", text, err)
		}
		iQuartet, err := strconv.ParseUint(fields[3], 10, 8)
		if err != nil {
			log.Fatalf("error parsing %q: %v\n", text, err)
		}
		iChannel, err := strconv.ParseUint(fields[4], 10, 8)
		if err != nil {
			log.Fatalf("error parsing %q: %v\n", text, err)
		}
		iCapacitor, err := strconv.ParseUint(fields[5], 10, 16)
		if err != nil {
			log.Fatalf("error parsing %q: %v\n", text, err)
		}
		pedestalMean, err := strconv.ParseFloat(fields[6], 64)
		if err != nil {
			log.Fatalf("error parsing %q: %v\n", text, err)
		}
		pedestalVariance, err := strconv.ParseFloat(fields[7], 64)
		if err != nil {
			log.Fatalf("error parsing %q: %v\n", text, err)
		}
		capacitor := d.Capacitor(uint8(iHemi), uint8(iASM), uint8(iDRS), uint8(iQuartet), uint8(iChannel), uint16(iCapacitor))
		capacitor.SetPedestalMeanStdDev(pedestalMean, pedestalVariance)
	}
}

func (d *Detector) PlotPedestals(plotStat bool) {

}

var Det *Detector

func init() {
	Det = NewDetector()
}
