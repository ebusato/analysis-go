package dpgadetector

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"time"

	"github.com/go-hep/csvutil"

	"gitlab.in2p3.fr/avirm/analysis-go/detector"
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

// Detector describes the 288 channels of the 12 ASM cards used the DPGA acquisition
// Out of these 288 channels, 240 are associated to the DPGA physical channels.
// The 48 remaining ones are not associated to any physical object.
type Detector struct {
	scintillator string
	samplingFreq float64 // sampling frequency in ns
	hemispheres  [2]Hemisphere
}

func NewDetector() *Detector {
	det := &Detector{
		scintillator: "LYSO",
		samplingFreq: 0.200,
	}
	for iHemi := range det.hemispheres {
		hemi := &det.hemispheres[iHemi]
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
						_, iChannelAbs288 := RelIdxToAbsIdx288(uint8(iHemi), uint8(iASM), uint8(iDRS), uint8(iQuartet), uint8(iChannel))
						ch.SetAbsID288(iChannelAbs288)
						switch iDRS == 2 && iQuartet == 1 {
						case false:
							_, iChannelAbs240 := RelIdxToAbsIdx240(uint8(iHemi), uint8(iASM), uint8(iDRS), uint8(iQuartet), uint8(iChannel))
							ch.SetAbsID240(iChannelAbs240)
						case true: // channel not used in DPGA (one of the four channels per ASM card which is not used)
							ch.SetAbsID240(math.MaxUint16)
						}
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

// iQuartetAbs can go from 0 to 59
func QuartetAbsIdx60ToRelIdx(iQuartetAbs uint8) (iHemi uint8, iASM uint8, iDRS uint8, iQuartet uint8) {
	iHemi = iQuartetAbs / 30
	iASM = iQuartetAbs/5 - iHemi*6
	iDRS = (iQuartetAbs - iASM*5 - iHemi*30) / 2
	iQuartet = iQuartetAbs - iASM*5 - iHemi*30 - iDRS*2
	return
}

// iQuartetAbs can go from 0 to 72
func QuartetAbsIdx72ToRelIdx(iQuartetAbs uint8) (iHemi uint8, iASM uint8, iDRS uint8, iQuartet uint8) {
	iHemi = iQuartetAbs / 36
	iASM = iQuartetAbs/6 - iHemi*6
	iDRS = (iQuartetAbs - iASM*6 - iHemi*36) / 2
	iQuartet = iQuartetAbs - iASM*6 - iHemi*36 - iDRS*2
	return
}

// iChannelAbs can go from 0 to 239
func ChannelAbsIdx240ToRelIdx(iChannelAbs uint16) (iHemi uint8, iASM uint8, iDRS uint8, iQuartet uint8, iChannel uint8) {
	iQuartetAbs := uint8(iChannelAbs / 4)
	iHemi, iASM, iDRS, iQuartet = QuartetAbsIdx60ToRelIdx(iQuartetAbs)
	iChannel = uint8(iChannelAbs % 4)
	return
}

// iChannelAbs can go from 0 to 287
func ChannelAbsIdx288ToRelIdx(iChannelAbs uint16) (iHemi uint8, iASM uint8, iDRS uint8, iQuartet uint8, iChannel uint8) {
	iQuartetAbs := uint8(iChannelAbs / 4)
	iHemi, iASM, iDRS, iQuartet = QuartetAbsIdx72ToRelIdx(iQuartetAbs)
	iChannel = uint8(iChannelAbs % 4)
	return
}

func RelIdxToAbsIdx240(iHemi uint8, iASM uint8, iDRS uint8, iQuartet uint8, iChannel uint8) (iQuartetAbs uint8, iChannelAbs uint16) {
	// 	if iDRS == 2 && iQuartet == 1 {
	// 		panic("dpgadetector: iDRS == 2 && iQuartet == 1")
	// 	}
	iQuartetAbs = iQuartet + iHemi*30 + iASM*5 + iDRS*2
	iChannelAbs = uint16(iQuartetAbs)*4 + uint16(iChannel)
	return
}

func RelIdxToAbsIdx288(iHemi uint8, iASM uint8, iDRS uint8, iQuartet uint8, iChannel uint8) (iQuartetAbs uint8, iChannelAbs uint16) {
	// 	if iDRS == 2 && iQuartet == 1 {
	// 		panic("dpgadetector: iDRS == 2 && iQuartet == 1")
	// 	}
	iQuartetAbs = iQuartet + iHemi*36 + iASM*6 + iDRS*2
	iChannelAbs = uint16(iQuartetAbs)*4 + uint16(iChannel)
	return
}

type GeomCSV struct {
	IChannelAbs240 uint16
	X              float64
	Y              float64
	Z              float64
}

func (d *Detector) ReadGeom() {
	fileName := "../dpgadetector/mapping_IDchanneltoCoordinates_Arnaud.txt"
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

	var data GeomCSV

	for rows.Next() {
		err = rows.Scan(&data)
		if err != nil {
			log.Fatalf("error reading row: %v\n", err)
		}
		channel := d.ChannelFromIdAbs240(data.IChannelAbs240)
		channel.SetCoord(data.X, data.Y, data.Z)
	}
	err = rows.Err()
	if err != nil && err.Error() != "EOF" {
		log.Fatalf("error: %v\n", err)
	}
}

func (d *Detector) ComputePedestalsMeanStdDevFromSamples() {
	for iHemi := range d.hemispheres {
		hemi := &d.hemispheres[iHemi]
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
	fmt.Println("Plotting pedestals")
	for iHemi := range d.hemispheres {
		for iASM := range d.hemispheres[iHemi].asm {
			for iDRS := range d.hemispheres[iHemi].asm[iASM].DRSs() {
				for iQuartet := range d.hemispheres[iHemi].asm[iASM].DRS(uint8(iDRS)).Quartets() {
					quartet := d.Quartet(uint8(iHemi), uint8(iASM), uint8(iDRS), uint8(iQuartet))
					text := fmt.Sprintf("iHemi%v_iASM%v_iDRS%v_iQuartet%v",
						strconv.FormatUint(uint64(iHemi), 10),
						strconv.FormatUint(uint64(iASM), 10),
						strconv.FormatUint(uint64(iDRS), 10),
						strconv.FormatUint(uint64(iQuartet), 10))
					quartet.PlotPedestals(plotStat, text)
				}
			}
		}
	}
}

func (d *Detector) Print() {
	fmt.Printf("Printing information for detector %v (sampling freq = %v)\n", d.scintillator, d.samplingFreq)
	for iHemi := range d.hemispheres {
		d.hemispheres[iHemi].Print()
	}
}

func (d *Detector) NoClusters() uint8 {
	return uint8(5 * len(d.hemispheres) * len(d.hemispheres[0].asm))
}

func (d *Detector) SamplingFreq() float64 {
	return d.samplingFreq
}

func (d *Detector) Capacitor(iHemi uint8, iASM uint8, iDRS uint8, iQuartet uint8, iChannel uint8, iCapacitor uint16) *detector.Capacitor {
	return d.hemispheres[int(iHemi)].asm[int(iASM)].DRS(iDRS).Quartet(iQuartet).Channel(iChannel).Capacitor(iCapacitor)
}

func (d *Detector) Channel(iHemi uint8, iASM uint8, iDRS uint8, iQuartet uint8, iChannel uint8) *detector.Channel {
	return d.hemispheres[int(iHemi)].asm[int(iASM)].DRS(iDRS).Quartet(iQuartet).Channel(iChannel)
}

func (d *Detector) ChannelFromIdAbs240(iChannelAbs240 uint16) *detector.Channel {
	iHemi, iASM, iDRS, iQuartet, iChannel := ChannelAbsIdx240ToRelIdx(iChannelAbs240)
	return d.hemispheres[int(iHemi)].asm[int(iASM)].DRS(iDRS).Quartet(iQuartet).Channel(iChannel)
}

func (d *Detector) ChannelFromIdAbs288(iChannelAbs288 uint16) *detector.Channel {
	iHemi, iASM, iDRS, iQuartet, iChannel := ChannelAbsIdx288ToRelIdx(iChannelAbs288)
	return d.hemispheres[int(iHemi)].asm[int(iASM)].DRS(iDRS).Quartet(iQuartet).Channel(iChannel)
}

func (d *Detector) Quartet(iHemi uint8, iASM uint8, iDRS uint8, iQuartet uint8) *detector.Quartet {
	return d.hemispheres[int(iHemi)].asm[int(iASM)].DRS(iDRS).Quartet(iQuartet)
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
	fmt.Println("Writing pedestals to", outFileName)
	tbl, err := csvutil.Create(outFileName)
	if err != nil {
		log.Fatalf("could not create %s: %v\n", outFileName, err)
	}
	defer tbl.Close()
	tbl.Writer.Comma = ' '

	err = tbl.WriteHeader(fmt.Sprintf("# DPGA pedestal file (creation date: %v)\n", time.Now()))
	err = tbl.WriteHeader("# iHemi iASM iDRS iQuartet iChannel iCapacitor pedestalMean pedestalStdDev")

	for iHemi := range d.hemispheres {
		hemi := &d.hemispheres[iHemi]
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
	Det.ReadGeom()
	//Det.Print()
}
