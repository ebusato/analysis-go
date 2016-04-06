// Package dpgadetector describes the physical structure of the dpga detector and its electronics.
// It is based on package detector.
package dpgadetector

import (
	"fmt"
	"log"
	"math"
	"strconv"
	"time"

	"github.com/go-hep/csvutil"

	"gitlab.in2p3.fr/avirm/analysis-go/detector"
	"gitlab.in2p3.fr/avirm/analysis-go/utils"
)

type HemisphereType uint8

const (
	right HemisphereType = iota
	left
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
		switch iHemi {
		case 0:
			hemi.which = right
		case 1:
			hemi.which = left
		}
		for iASM := range hemi.asm {
			asm := &hemi.asm[iASM]
			switch hemi.which {
			case right:
				// on the right hemisphere, the ASM card with index iASM=0
				// is the one at the very top of the detector (corresponding to line 1)
				asm.SetCylCoord(148.336, math.Pi*(180.-37.5+float64(iASM)*15)/180., -(3*19.5+6.5/2.+13)-float64(iASM%3)*6.5)
			case left:
				// on the left hemisphere, the ASM card with index iASM=0
				// is the one at the very bottom of the detector (corresponding to line 7)
				asm.SetCylCoord(148.336, math.Pi*(-37.5+float64(iASM)*15)/180., -(3*19.5+6.5/2.+13)-float64(iASM%3)*6.5)
			}
			for iDRS := range asm.DRSs() {
				drs := asm.DRS(uint8(iDRS))
				drs.SetID(uint8(iDRS))
				for iQuartet := range drs.Quartets() {
					quartet := drs.Quartet(uint8(iQuartet))
					quartet.SetID(uint8(iQuartet))
					// iqtemp is a temporary index refering to the quartet within a line.
					// iqtemp goes from 0 to 4 (5 quartets per line). It is equal to 0 for
					// the quartet on the front side of the DPGA and equal to 4 for the
					// quartet on the rear side of the DPGA.
					// iqtemp corresponds to iDRS=2 and iQuartet=1, hence the non-physical values
					// of the coordinates.
					iqtemp := len(drs.Quartets())*iDRS + iQuartet
					switch iqtemp == 5 {
					case false:
						quartet.SetCylCoord(asm.CylCoord.R, asm.CylCoord.Phi, asm.CylCoord.Z+float64(iqtemp)*2*19.5)
					case true:
						quartet.SetCylCoord(0, 0, 0)
					}
					// Convert right away cylindrical coordinates to cartesian ones.
					// To be used in the following channel loop
					qCartCoord := utils.CylindricalToCartesian(quartet.CylCoord)
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

							// determine channel cartesian coordinates from quartet's cylindrical coordinates
							var xSign int
							var ySign int
							var zSign int
							// Set xSign and ySign (channels 0 and 1 have the same X and Y coordinates and thus the same xSign and ySign)
							if iChannel == 0 || iChannel == 1 {
								switch hemi.which {
								case right:
									xSign = -1
									ySign = -1
								case left:
									xSign = +1
									ySign = -1
								}
							} else { // iChannel == 2 || iChannel == 3
								switch hemi.which {
								case right:
									xSign = +1
									ySign = +1
								case left:
									xSign = -1
									ySign = +1
								}
							}
							// Set zSign
							if iChannel == 0 || iChannel == 2 {
								switch hemi.which {
								case right:
									zSign = +1
								case left:
									zSign = -1
								}
							} else { // iChannel == 1 || iChannel == 3
								switch hemi.which {
								case right:
									zSign = -1
								case left:
									zSign = +1
								}
							}
							// fmt.Printf("iChannelAbs240=%v, phi=%v, Yquartet=%v\n", iChannelAbs240, quartet.CylCoord.Phi, qCartCoord.Y)
							x := qCartCoord.X + float64(xSign)*19.5/2.*math.Sin(quartet.CylCoord.Phi)
							var y float64
							// in the case of the y coordinate, the angle needs to be adjusted when projecting as cos(pi - phi) = -cos(phi).
							// (in the case of the x coordinate, the projection is done with sin, it's therefore not necessary to change phi to pi - phi
							// as sin(pi - phi) = sin(phi)
							switch hemi.which {
							case left:
								y = qCartCoord.Y + float64(ySign)*19.5/2.*math.Cos(quartet.CylCoord.Phi)
							case right:
								y = qCartCoord.Y + float64(ySign)*19.5/2.*math.Cos(math.Pi-quartet.CylCoord.Phi)
							}
							z := qCartCoord.Z + float64(zSign)*(6.5/2.+6.5)
							ch.SetCartCoord(x, y, z)
							// set the scintillator coordinates (8 values, each corresponding to one corner of the rectangular parallelepiped)
							// corners are counted in the following order
							//
							//            7  _____________ 6
							//             /|            /|
							//            / |           / |
							//         3 /__|__ _____ 2/  |
							//           |  |_________ |__|   rear face
							//           |  /4         |  /5
							//           | /           | /
							//           |/____________|/
							//           0            1   front face
							//
							switch hemi.which {
							case right:
								ch.ScintCoords[0].X = x - 13/2.*math.Sin(quartet.CylCoord.Phi)
								ch.ScintCoords[1].X = ch.ScintCoords[0].X
								ch.ScintCoords[2].X = x + 13/2.*math.Sin(quartet.CylCoord.Phi)
								ch.ScintCoords[3].X = ch.ScintCoords[2].X
								ch.ScintCoords[4].X = ch.ScintCoords[0].X - 15*math.Cos(math.Pi-quartet.CylCoord.Phi)
								ch.ScintCoords[5].X = ch.ScintCoords[4].X
								ch.ScintCoords[6].X = ch.ScintCoords[2].X - 15*math.Cos(math.Pi-quartet.CylCoord.Phi)
								ch.ScintCoords[7].X = ch.ScintCoords[6].X

								ch.ScintCoords[0].Y = y - 13/2.*math.Cos(math.Pi-quartet.CylCoord.Phi)
								ch.ScintCoords[1].Y = ch.ScintCoords[0].Y
								ch.ScintCoords[2].Y = y + 13/2.*math.Cos(math.Pi-quartet.CylCoord.Phi)
								ch.ScintCoords[3].Y = ch.ScintCoords[2].Y
								ch.ScintCoords[4].Y = ch.ScintCoords[0].Y + 15*math.Sin(quartet.CylCoord.Phi) // in fact pi - phi, but sin(pi - phi) = sin(phi)
								ch.ScintCoords[5].Y = ch.ScintCoords[4].Y
								ch.ScintCoords[6].Y = ch.ScintCoords[2].Y + 15*math.Sin(quartet.CylCoord.Phi) // in fact pi - phi, but sin(pi - phi) = sin(phi)
								ch.ScintCoords[7].Y = ch.ScintCoords[6].Y

								ch.ScintCoords[0].Z = z + 13/2.
								ch.ScintCoords[1].Z = z - 13/2.
								ch.ScintCoords[2].Z = ch.ScintCoords[1].Z
								ch.ScintCoords[3].Z = ch.ScintCoords[0].Z
								ch.ScintCoords[4].Z = ch.ScintCoords[0].Z
								ch.ScintCoords[5].Z = ch.ScintCoords[1].Z
								ch.ScintCoords[6].Z = ch.ScintCoords[5].Z
								ch.ScintCoords[7].Z = ch.ScintCoords[3].Z
							case left:
								ch.ScintCoords[0].X = x + 13/2.*math.Sin(quartet.CylCoord.Phi)
								ch.ScintCoords[1].X = ch.ScintCoords[0].X
								ch.ScintCoords[2].X = x - 13/2.*math.Sin(quartet.CylCoord.Phi)
								ch.ScintCoords[3].X = ch.ScintCoords[2].X
								ch.ScintCoords[4].X = ch.ScintCoords[0].X + 15*math.Cos(quartet.CylCoord.Phi)
								ch.ScintCoords[5].X = ch.ScintCoords[4].X
								ch.ScintCoords[6].X = ch.ScintCoords[2].X + 15*math.Cos(quartet.CylCoord.Phi)
								ch.ScintCoords[7].X = ch.ScintCoords[6].X

								ch.ScintCoords[0].Y = y - 13/2.*math.Cos(quartet.CylCoord.Phi)
								ch.ScintCoords[1].Y = ch.ScintCoords[0].Y
								ch.ScintCoords[2].Y = y + 13/2.*math.Cos(quartet.CylCoord.Phi)
								ch.ScintCoords[3].Y = ch.ScintCoords[2].Y
								ch.ScintCoords[4].Y = ch.ScintCoords[0].Y + 15*math.Sin(quartet.CylCoord.Phi)
								ch.ScintCoords[5].Y = ch.ScintCoords[4].Y
								ch.ScintCoords[6].Y = ch.ScintCoords[2].Y + 15*math.Sin(quartet.CylCoord.Phi)
								ch.ScintCoords[7].Y = ch.ScintCoords[6].Y

								ch.ScintCoords[0].Z = z - 13/2.
								ch.ScintCoords[1].Z = z + 13/2.
								ch.ScintCoords[2].Z = ch.ScintCoords[1].Z
								ch.ScintCoords[3].Z = ch.ScintCoords[0].Z
								ch.ScintCoords[4].Z = ch.ScintCoords[0].Z
								ch.ScintCoords[5].Z = ch.ScintCoords[1].Z
								ch.ScintCoords[6].Z = ch.ScintCoords[5].Z
								ch.ScintCoords[7].Z = ch.ScintCoords[3].Z
							}
						case true: // channel not used in DPGA (one of the four channels per ASM card which is not used)
							ch.SetAbsID240(math.MaxUint16)
						}
						ch.SetFifoID(ChannelAbsIdx288ToFifoID(iChannelAbs288))
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

// iChannelAbs can go from 0 to 287
func ChannelAbsIdx288ToFifoID(iChannelAbs uint16) uint16 {
	return iChannelAbs / 2
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

// DumpGeom dumps the x, y and z coordinates of the 240 channels (indentified by the iChannelAbs240 index) of the DPGA.
// Coordinates are those of the detector.Channel.CartCoord object. They correspond to the coordinates of the center
// of the front face of the
func (d *Detector) DumpGeom() {
	tbl, err := csvutil.Create("dpgageom.csv")
	if err != nil {
		log.Fatalf("could not create dpgageom.csv: %v\n", err)
	}
	defer tbl.Close()
	tbl.Writer.Comma = ' '

	err = tbl.WriteHeader(fmt.Sprintf("# DPGA geometry file (creation date: %v)\n", time.Now()))
	err = tbl.WriteHeader("# iChannelAbs240 X Y Z")

	for i := uint16(0); i < 240; i++ {
		ch := d.ChannelFromIdAbs240(i)
		data := GeomCSV{
			i,
			ch.X,
			ch.Y,
			ch.Z,
		}
		err = tbl.WriteRow(data)
		if err != nil {
			log.Fatalf("error writing row: %v\n", err)
		}
	}
	err = tbl.Close()
	if err != nil {
		log.Fatalf("error closing table: %v\n", err)
	}
}

type FullGeomCSV struct {
	IChannelAbs240 uint16
	X0, Y0, Z0     float64 // cartesian coords of corner 0
	X1, Y1, Z1     float64 // cartesian coords of corner 1
	X2, Y2, Z2     float64 // cartesian coords of corner 2
	X3, Y3, Z3     float64 // cartesian coords of corner 3
	X4, Y4, Z4     float64 // cartesian coords of corner 4
	X5, Y5, Z5     float64 // cartesian coords of corner 5
	X6, Y6, Z6     float64 // cartesian coords of corner 6
	X7, Y7, Z7     float64 // cartesian coords of corner 7
}

// DumpFullGeom dumps the 8 coordinates of the 240 PMTs of the DPGA
func (d *Detector) DumpFullGeom() {
	tbl, err := csvutil.Create("dpgafullgeom.csv")
	if err != nil {
		log.Fatalf("could not create dpgafullgeom.csv: %v\n", err)
	}
	defer tbl.Close()
	tbl.Writer.Comma = ' '

	err = tbl.WriteHeader(fmt.Sprintf("# DPGA full geometry file (creation date: %v)\n", time.Now()))
	err = tbl.WriteHeader("# iChannelAbs240 X0 Y0 Z0 X1 Y1 Z1 X2 Y2 Z2 X3 Y3 Z3 X4 Y4 Z4 X5 Y5 Z5 X6 Y6 Z6 X7 Y7 Z7")

	for i := uint16(0); i < 240; i++ {
		ch := d.ChannelFromIdAbs240(i)
		data := FullGeomCSV{
			i,
			ch.ScintCoords[0].X, ch.ScintCoords[0].Y, ch.ScintCoords[0].Z,
			ch.ScintCoords[1].X, ch.ScintCoords[1].Y, ch.ScintCoords[1].Z,
			ch.ScintCoords[2].X, ch.ScintCoords[2].Y, ch.ScintCoords[2].Z,
			ch.ScintCoords[3].X, ch.ScintCoords[3].Y, ch.ScintCoords[3].Z,
			ch.ScintCoords[4].X, ch.ScintCoords[4].Y, ch.ScintCoords[4].Z,
			ch.ScintCoords[5].X, ch.ScintCoords[5].Y, ch.ScintCoords[5].Z,
			ch.ScintCoords[6].X, ch.ScintCoords[6].Y, ch.ScintCoords[6].Z,
			ch.ScintCoords[7].X, ch.ScintCoords[7].Y, ch.ScintCoords[7].Z,
		}
		err = tbl.WriteRow(data)
		if err != nil {
			log.Fatalf("error writing row: %v\n", err)
		}
	}
	err = tbl.Close()
	if err != nil {
		log.Fatalf("error closing table: %v\n", err)
	}
}

// func (d *Detector) ReadGeom() {
// 	fileName := os.Getenv("GOPATH") + "/src/gitlab.in2p3.fr/avirm/analysis-go/dpga/dpgadetector/mapping_IDchanneltoCoordinates_Arnaud.txt"
// 	tbl, err := csvutil.Open(fileName)
// 	if err != nil {
// 		log.Fatalf("could not open %s: %v\n", fileName, err)
// 	}
// 	defer tbl.Close()
// 	tbl.Reader.Comma = ' '
// 	tbl.Reader.Comment = '#'
//
// 	rows, err := tbl.ReadRows(0, -1)
// 	if err != nil {
// 		log.Fatalf("could read rows [0, -1): %v\n", err)
// 	}
// 	defer rows.Close()
//
// 	var data GeomCSV
//
// 	for rows.Next() {
// 		err = rows.Scan(&data)
// 		if err != nil {
// 			log.Fatalf("error reading row: %v\n", err)
// 		}
// 		channel := d.ChannelFromIdAbs240(data.IChannelAbs240)
// 		channel.SetCartCoord(data.X, data.Y, data.Z)
// 	}
// 	err = rows.Err()
// 	if err != nil && err.Error() != "EOF" {
// 		log.Fatalf("error: %v\n", err)
// 	}
// }

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
	Det.DumpGeom()
	Det.DumpFullGeom()
	//Det.Print()
}
