// Package dpgadetector describes the physical structure of the test bench detector and its electronics.
// It is based on package detector.
package tbdetector

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/go-hep/csvutil"

	"gitlab.in2p3.fr/avirm/analysis-go/detector"
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
		//samplingFreq: 1,
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
				ch.SetFifoID(FifoID(uint8(iDRS), uint8(iQuartet), uint8(iChannel)))
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

func (d *Detector) NoClusters() uint8 {
	return uint8(6)
}

func FifoID(iDRS uint8, iQuartet uint8, iChannel uint8) uint16 {
	return uint16(iChannel/2 + iQuartet*2 + iDRS*4)
}

func (d *Detector) SamplingFreq() float64 {
	return d.samplingFreq
}

func (d *Detector) Capacitor(iDRS uint8, iQuartet uint8, iChannel uint8, iCapacitor uint16) *detector.Capacitor {
	return d.asm.DRS(iDRS).Quartet(iQuartet).Channel(iChannel).Capacitor(iCapacitor)
}

func (d *Detector) Channel(iDRS uint8, iQuartet uint8, iChannel uint8) *detector.Channel {
	return d.asm.DRS(iDRS).Quartet(iQuartet).Channel(iChannel)
}

func (d *Detector) Quartet(iDRS uint8, iQuartet uint8) *detector.Quartet {
	return d.asm.DRS(iDRS).Quartet(iQuartet)
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

func (d *Detector) PlotPedestals(plotStat bool) {
	for iDRS := range d.asm.DRSs() {
		drs := d.asm.DRS(uint8(iDRS))
		for iQuartet := range drs.Quartets() {
			quartet := drs.Quartet(uint8(iQuartet))
			text := fmt.Sprintf("iDRS%v_iQuartet%v",
				strconv.FormatUint(uint64(iDRS), 10),
				strconv.FormatUint(uint64(iQuartet), 10))
			quartet.PlotPedestals(plotStat, text)
		}
	}
}

type PedestalCSV struct {
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

	err = tbl.WriteHeader(fmt.Sprintf("# Test bench pedestal file (creation date: %v)\n", time.Now()))
	err = tbl.WriteHeader("# iDRS iQuartet iChannel iCapacitor pedestalMean pedestalStdDev")

	if err != nil {
		log.Fatalf("error writing header: %v\n", err)
	}

	for iDRS := range d.asm.DRSs() {
		drs := d.asm.DRS(uint8(iDRS))
		for iQuartet := range drs.Quartets() {
			quartet := drs.Quartet(uint8(iQuartet))
			for iChannel := range quartet.Channels() {
				ch := quartet.Channel(uint8(iChannel))
				for iCapacitor := range ch.Capacitors() {
					capa := ch.Capacitor(uint16(iCapacitor))
					data := PedestalCSV{
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

	var data PedestalCSV

	for rows.Next() {
		err = rows.Scan(&data)
		if err != nil {
			log.Fatalf("error reading row: %v\n", err)
		}
		//fmt.Printf("data: %+v\n", data)
		capacitor := d.Capacitor(data.IDRS, data.IQuartet, data.IChannel, data.ICapacitor)
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
