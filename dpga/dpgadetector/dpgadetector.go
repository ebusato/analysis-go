package dpgadetector

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/go-hep/csvutil"

	"gitlab.in2p3.fr/AVIRM/Analysis-go/detector"
)

type HemisphereType uint8

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

func (d *Detector) PlotPedestals(plotStat bool) {

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

type PedestalFile struct {
	IHemi      uint8
	IASM       uint8
	IDRS       uint8
	IQuartet   uint8
	IChannel   uint8
	ICapacitor uint16
	Mean       float64
	StdDev     float64
}

func (d *Detector) WritePedestalsToFile(outFileName string) {
	tbl, err := csvutil.Create(outFileName)
	if err != nil {
		log.Fatalf("could not create %s: %v\n", outFileName, err)
	}
	defer tbl.Close()
	tbl.Writer.Comma = ' '

	err = tbl.WriteHeader(fmt.Sprintf("# DPGA pedestal file (creation date: %v)\n", time.Now()))
	err = tbl.WriteHeader("# iHemi iASM iDRS iQuartet iChannel iCapacitor pedestalMean pedestalStdDev")

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
							data := PedestalFile{
								IHemi:      uint8(iHemi),
								IASM:       uint8(iASM),
								IDRS:       uint8(iDRS),
								IQuartet:   uint8(iQuartet),
								IChannel:   uint8(iChannel),
								ICapacitor: uint16(iCapacitor),
								Mean:       capa.PedestalMean(),
								StdDev:     capa.PedestalStdDev(),
							}
							err = tbl.WriteRow(data)
							if err != nil {
								log.Fatalf("error writing row: %v\n", err)
							}
						}
					}
				}
			}
		}
	}

	err = tbl.Close()
	if err != nil {
		log.Fatalf("error closing table: %v\n", err)
	}
}

func (d *Detector) ReadPedestalsFile(fileName string) {
	tbl, err := csvutil.Open(fileName)
	if err != nil {
		log.Fatalf("could not open %s: %v\n", fileName, err)
	}
	defer tbl.Close()
	tbl.Reader.Comma = ' '
	tbl.Reader.Comment = '#'

	rows, err := tbl.ReadRows(0, -1)
	if err != nil {
		log.Fatalf("could read rows [0, -1): %v\n", err)
	}
	defer rows.Close()

	var data PedestalFile

	for rows.Next() {
		err = rows.Scan(&data)
		if err != nil {
			log.Fatalf("error reading row: %v\n", err)
		}
		capacitor := d.Capacitor(data.IHemi, data.IASM, data.IDRS, data.IQuartet, data.IChannel, data.ICapacitor)
		capacitor.SetPedestalMeanStdDev(data.Mean, data.StdDev)
	}
	err = rows.Err()
	if err != nil && err.Error() != "EOF" {
		log.Fatalf("error: %v\n", err)
	}
}

var Det *Detector

func init() {
	Det = NewDetector()
}
