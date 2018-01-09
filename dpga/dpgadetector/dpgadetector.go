// Package dpgadetector describes the physical structure of the dpga detector and its electronics.
// It is based on package detector.
package dpgadetector

import (
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"time"

	"go-hep.org/x/hep/csvutil"

	"gitlab.in2p3.fr/avirm/analysis-go/detector"
	"gitlab.in2p3.fr/avirm/analysis-go/utils"
)

type HemisphereType uint8

const (
	Right HemisphereType = iota
	Left
)

type Hemisphere struct {
	asm   [6]detector.ASMCard
	which HemisphereType
}

func (h *Hemisphere) ASMs() [6]detector.ASMCard {
	return h.asm
}

func (h *Hemisphere) ASM(i int) *detector.ASMCard {
	return &h.asm[i]
}

func (h *Hemisphere) Print() {
	fmt.Printf("Printing Hemisphere %v\n", h.which)
	for iASM := range h.asm {
		h.asm[iASM].Print()
	}
}

func (h *Hemisphere) Which() HemisphereType {
	return h.which
}

func (h *Hemisphere) GoString() string {
	switch h.which {
	case Right:
		return "right"
	case Left:
		return "left"
	default:
		return fmt.Sprintf("InputType(%d)", h.which)
	}
}

// Detector describes the 288 channels of the 12 ASM cards used in the DPGA acquisition.
// Out of these 288 channels, 240 are associated to the DPGA physical channels.
// The 48 remaining ones are not associated to any physical object.
type Detector struct {
	scintillator string
	samplingFreq float64 // sampling frequency in ns
	noSamples    int     // number of samples
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
			hemi.which = Right
		case 1:
			hemi.which = Left
		}
		for iASM := range hemi.asm {
			asm := &hemi.asm[iASM]
			asm.UpStr = hemi
			// The value 148.4 for the radial distance is the one used by Christophe Insa is Catia.
			// We use this same value here for consistency, but note that it is slighlty different
			// than the one we find in some papers (e.g. 148.336, but the difference is
			// way smaller than the incertainty (of the order of 1 mm) so it's no big deal)
			switch hemi.which {
			case Right:
				// on the right hemisphere, the ASM card with index iASM=0
				// is the one at the very top of the detector (corresponding to line 1)
				asm.SetCylCoord(148.4, math.Pi*(180.-37.5+float64(iASM)*15)/180., -(3*19.5+6.5/2.+13-6.5/2)-float64(iASM%3)*6.5)
			case Left:
				// on the left hemisphere, the ASM card with index iASM=0
				// is the one at the very bottom of the detector (corresponding to line 7)
				asm.SetCylCoord(148.4, math.Pi*(-37.5+float64(iASM)*15)/180., -(3*19.5+6.5/2.+13-6.5/2)-float64(iASM%3)*6.5)
			}
			for iDRS := range asm.DRSs() {
				drs := asm.DRS(uint8(iDRS))
				drs.SetID(uint8(iDRS))
				drs.ASMCard = asm
				for iQuartet := range drs.Quartets() {
					quartet := drs.Quartet(uint8(iQuartet))
					quartet.SetID(uint8(iQuartet))
					quartet.DRS = drs
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
						ch.Quartet = quartet
						ch.SetName("PMT" + strconv.FormatUint(uint64(iChannel), 10))
						ch.SetID(uint8(iChannel))
						_, iChannelAbs288 := RelIdxToAbsIdx288(uint8(iHemi), uint8(iASM), uint8(iDRS), uint8(iQuartet), uint8(iChannel))
						ch.SetAbsID288(iChannelAbs288)
						switch iDRS == 2 && iQuartet == 1 {
						case false:
							_, iChannelAbs240 := RelIdxToAbsIdx240(uint8(iHemi), uint8(iASM), uint8(iDRS), uint8(iQuartet), uint8(iChannel))
							ch.SetAbsID240(iChannelAbs240)

							// Determine channel cartesian coordinates from quartet's cylindrical coordinates
							// The position of the PMT within a quartet is given by its index on the PMT base card. Let's call this index idxPMT (0 -> 3).
							// When looking from the inside of the detector to the quartets, we have
							//
							//      _________________________
							//      |           |           |
							//      |  idxPMT2  |  idxPMT3  |
							//      |           |           |
							//      |-----------|-----------|
							//      |           |           |
							//      |  idxPMT0  |  idxPMT1  |
							//      |___________|___________|
							//
							//
							// This idxPMT index is different from the electronics channel index (or the index within the Quartet.channels array, which is
							// the same as the electronics channel index). The electronics channel index is the index specifying on which pin of the connector
							// on the HV divider card the signal is leaving the card.
							// Let's call this index idxElec (0 -> 3). We have, representing the connector on the HV divider card :
							//
							//            ________________________________________________
							//           |                                                |
							//           |      pin        pin       pin      pin (brown) |
							//           |       |          |         |        |          |
							//           --------------------------------------------------
							//                idxElec3  idxElec2  idxElec1  idxElec0
							//
							//
							// A mapping thus needs to be done to map the electronic channels to the right position.
							// The mapping is the following
							//
							//     idxPMT   |    idxElec
							//     ---------------------
							//        0     |       3
							//        1     |       2
							//        2     |       1
							//        3     |       0

							var xSign int
							var ySign int
							var zSign int
							// Set xSign and ySign (channels 0 and 1 have the same X and Y coordinates and thus the same xSign and ySign)
							if iChannel == 3 || iChannel == 2 {
								switch hemi.which {
								case Right:
									xSign = -1
									ySign = -1
								case Left:
									xSign = +1
									ySign = -1
								}
							} else { // iChannel == 0 || iChannel == 1
								switch hemi.which {
								case Right:
									xSign = +1
									ySign = +1
								case Left:
									xSign = -1
									ySign = +1
								}
							}
							// Set zSign
							if iChannel == 3 || iChannel == 1 {
								switch hemi.which {
								case Right:
									zSign = +1
								case Left:
									zSign = -1
								}
							} else { // iChannel == 2 || iChannel == 0
								switch hemi.which {
								case Right:
									zSign = -1
								case Left:
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
							case Left:
								y = qCartCoord.Y + float64(ySign)*19.5/2.*math.Cos(quartet.CylCoord.Phi)
							case Right:
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
							case Right:
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
							case Left:
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
							ch.CrystCenter.X = 0
							ch.CrystCenter.Y = 0
							ch.CrystCenter.Z = 0
							for i := range ch.ScintCoords {
								ch.CrystCenter.X += ch.ScintCoords[i].X
								ch.CrystCenter.Y += ch.ScintCoords[i].Y
								ch.CrystCenter.Z += ch.ScintCoords[i].Z
							}
							ch.CrystCenter.X /= float64(len(ch.ScintCoords))
							ch.CrystCenter.Y /= float64(len(ch.ScintCoords))
							ch.CrystCenter.Z /= float64(len(ch.ScintCoords))
						case true: // channel not used in DPGA (one of the four channels per ASM card which is not used)
							ch.SetAbsID240(math.MaxUint16)
						}
						ch.SetFifoID144(ChannelAbsIdx288ToFifoID144(iChannelAbs288))
						for iCapacitor := range ch.Capacitors() {
							capa := ch.Capacitor(uint16(iCapacitor))
							capa.SetID(uint16(iCapacitor))
							capa.Channel = ch
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
func ChannelAbsIdx288ToFifoID144(iChannelAbs uint16) uint16 {
	return iChannelAbs / 2
}

// iFifo can go from 0 to 143
// The returned value can go from 0 to 119
func FifoID144ToFifoID120(iFifo uint16, throwErr bool) uint16 {
	i := iFifo % 12
	if throwErr && (i == 10 || i == 11) {
		log.Fatalf("iFifo=%v -> corresponds to an unused fifo.\n", iFifo)
	}
	return iFifo - 2*(iFifo/12)
}

// iFifo can go from 0 to 143
// QuartetAbsIdx60 can go from 0 to 59
func FifoID144ToQuartetAbsIdx60(iFifo uint16, throwErr bool) uint8 {
	return uint8(FifoID144ToFifoID120(iFifo, throwErr) / 2)
}

// iFifo can go from 0 to 143
// QuartetAbsIdx72 can go from 0 to 71
func FifoID144ToQuartetAbsIdx72(iFifo uint16) uint8 {
	return uint8(iFifo) / 2
}

func QuartetAbsIdx60ToLineAbsIdx12(iQuartet uint8) uint8 {
	return iQuartet / 5
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

// FEIdAndChanIdToQuartetAbsIdx60 return the quartet absolute Id (0->59)
// from an input front end (ASM) board id and a channel Id (0 -> 23)
func FEIdAndChanIdToQuartetAbsIdx60(FEId uint16, ChanId uint16, throwErr bool) uint8 {
	var FEIdNew uint16
	switch FEId {
	case 0x10:
		FEIdNew = 0
	case 0x11:
		FEIdNew = 1
	case 0x12:
		FEIdNew = 2
	case 0x13:
		FEIdNew = 3
	case 0x14:
		FEIdNew = 4
	case 0x15:
		FEIdNew = 5
	case 0x16:
		FEIdNew = 6
	case 0x17:
		FEIdNew = 7
	case 0x18:
		FEIdNew = 8
	case 0x19:
		FEIdNew = 9
	case 0x1a:
		FEIdNew = 10
	case 0x1b:
		FEIdNew = 11
	case 0x1e:
		FEIdNew = 0
	default:
		panic("dpgadetector.FEIdAndChanIdToQuartetAbsIdx60: FEId not know")
	}
	ChanAbsIdx288 := FEIdNew*24 + ChanId
	FifoId144 := ChannelAbsIdx288ToFifoID144(ChanAbsIdx288)
	return FifoID144ToQuartetAbsIdx60(FifoId144, throwErr)
}

// FEIdAndChanIdToQuartetAbsIdx72 return the quartet absolute Id (0->71)
// from an input front end (ASM) board id and a channel Id (0 -> 23)
// There is much in common with FEIdAndChanIdToQuartetAbsIdx60 -> merge at some point
func FEIdAndChanIdToQuartetAbsIdx72(FEId uint16, ChanId uint16) uint8 {
	var FEIdNew uint16
	switch FEId {
	case 0x10:
		FEIdNew = 0
	case 0x11:
		FEIdNew = 1
	case 0x12:
		FEIdNew = 2
	case 0x13:
		FEIdNew = 3
	case 0x14:
		FEIdNew = 4
	case 0x15:
		FEIdNew = 5
	case 0x16:
		FEIdNew = 6
	case 0x17:
		FEIdNew = 7
	case 0x18:
		FEIdNew = 8
	case 0x19:
		FEIdNew = 9
	case 0x1a:
		FEIdNew = 10
	case 0x1b:
		FEIdNew = 11
	case 0x1e:
		FEIdNew = 0
	default:
		panic("dpgadetector.FEIdAndChanIdToQuartetAbsIdx60: FEId not know")
	}
	ChanAbsIdx288 := FEIdNew*24 + ChanId
	FifoId144 := ChannelAbsIdx288ToFifoID144(ChanAbsIdx288)
	return FifoID144ToQuartetAbsIdx72(FifoId144)
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
	fileName := os.Getenv("GOPATH") + "/src/gitlab.in2p3.fr/avirm/analysis-go/dpga/dpgadetector/geom/dpgageom.csv"
	tbl, err := csvutil.Create(fileName)
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
	fileName := os.Getenv("GOPATH") + "/src/gitlab.in2p3.fr/avirm/analysis-go/dpga/dpgadetector/geom/dpgafullgeom.csv"
	tbl, err := csvutil.Create(fileName)
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

// FinalizePedestalsMeanErr finalizes the computations of the pedestals for all capacitors.
func (d *Detector) FinalizePedestalsMeanErr() {
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
							capa.FinalizePedestalMeanErr()
						}
					}
				}
			}
		}
	}
}

// FinalizeTimeDepOffsetsMeanErr finalizes the computations of the time dependent offsets for all channels.
func (d *Detector) FinalizeTimeDepOffsetsMeanErr() {
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
						ch.FinalizeTimeDepOffsetMeanErr()
					}
				}
			}
		}
	}
}

// PlotPedestals plots pedestals
func (d *Detector) PlotPedestals(outDir string, plotStat bool) {
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
					quartet.PlotPedestals(outDir, plotStat, text)
				}
			}
		}
	}
}

// PlotTimeDepOffsets plots time dependent offsets
func (d *Detector) PlotTimeDepOffsets() {
	fmt.Println("Plotting time dependent offsets -> to be implemented")
	/*for iHemi := range d.hemispheres {
		for iASM := range d.hemispheres[iHemi].asm {
			for iDRS := range d.hemispheres[iHemi].asm[iASM].DRSs() {
				for iQuartet := range d.hemispheres[iHemi].asm[iASM].DRS(uint8(iDRS)).Quartets() {
					quartet := d.Quartet(uint8(iHemi), uint8(iASM), uint8(iDRS), uint8(iQuartet))
					text := fmt.Sprintf("iHemi%v_iASM%v_iDRS%v_iQuartet%v",
						strconv.FormatUint(uint64(iHemi), 10),
						strconv.FormatUint(uint64(iASM), 10),
						strconv.FormatUint(uint64(iDRS), 10),
						strconv.FormatUint(uint64(iQuartet), 10))
					//quartet.PlotPedestals(plotStat, text)
				}
			}
		}
	}*/
}

// SetNoSamples sets the number of physical samples
func (d *Detector) SetNoSamples(noSamples int) {
	d.noSamples = noSamples
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
						ch.SetNoSamples(noSamples)
					}
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

func (d *Detector) Hemispheres() [2]Hemisphere {
	return d.hemispheres
}

func (d *Detector) Hemisphere(i int) *Hemisphere {
	return &d.hemispheres[i]
}

func (d *Detector) NoClusters() uint8 {
	return uint8(5 * len(d.hemispheres) * len(d.hemispheres[0].asm))
}

func (d *Detector) NoSamples() int {
	return d.noSamples
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

func (d *Detector) QuartetFromIdAbs60(iQuartetAbs60 uint8) *detector.Quartet {
	iHemi, iASM, iDRS, iQuartet := QuartetAbsIdx60ToRelIdx(iQuartetAbs60)
	return d.Quartet(iHemi, iASM, iDRS, iQuartet)
}

type PedestalFile struct {
	IHemi      uint8
	IASM       uint8
	IDRS       uint8
	IQuartet   uint8
	IChannel   uint8
	ICapacitor uint16
	Mean       float64
	MeanErr    float64
}

func (d *Detector) WritePedestalsToFile(outFileName string, inFileName string) {
	fmt.Println("Writing pedestals to", outFileName)
	tbl, err := csvutil.Create(outFileName)
	if err != nil {
		log.Fatalf("could not create %s: %v\n", outFileName, err)
	}
	defer tbl.Close()
	tbl.Writer.Comma = ' '

	err = tbl.WriteHeader(fmt.Sprintf("# DPGA pedestal file (creation date: %v, input file: %v)\n", time.Now(), inFileName))
	err = tbl.WriteHeader("# iHemi iASM iDRS iQuartet iChannel iCapacitor pedestalMean pedestalMeanErr")

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
								MeanErr:    capa.PedestalMeanErr(),
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
		capacitor.SetPedestalMeanErr(data.Mean, data.MeanErr)
	}
	err = rows.Err()
	if err != nil && err.Error() != "EOF" {
		log.Fatalf("error: %v\n", err)
	}
}

type TimeDepOffsetFile struct {
	IHemi    uint8
	IASM     uint8
	IDRS     uint8
	IQuartet uint8
	IChannel uint8
	ISample  uint16
	Mean     float64
	MeanErr  float64
}

func (d *Detector) WriteTimeDepOffsetsToFile(outFileName string, inFileName string) {
	fmt.Println("Writing time dependent offsets to", outFileName)
	tbl, err := csvutil.Create(outFileName)
	if err != nil {
		log.Fatalf("could not create %s: %v\n", outFileName, err)
	}
	defer tbl.Close()
	tbl.Writer.Comma = ' '

	err = tbl.WriteHeader(fmt.Sprintf("# DPGA time dependent offset file (creation date: %v, input file: %v)\n", time.Now(), inFileName))
	err = tbl.WriteHeader("# iHemi iASM iDRS iQuartet iChannel iSample timeDepOffsetMean timeDepOffsetMeanErr")

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
						for iSample := 0; iSample < len(ch.TimeDepOffsetMeans()); iSample++ {
							data := TimeDepOffsetFile{
								IHemi:    uint8(iHemi),
								IASM:     uint8(iASM),
								IDRS:     uint8(iDRS),
								IQuartet: uint8(iQuartet),
								IChannel: uint8(iChannel),
								ISample:  uint16(iSample),
								Mean:     ch.TimeDepOffsetMean(iSample),
								MeanErr:  ch.TimeDepOffsetMeanErr(iSample),
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

func (d *Detector) ReadTimeDepOffsetsFile(fileName string) {
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

	var data TimeDepOffsetFile

	for rows.Next() {
		err = rows.Scan(&data)
		if err != nil {
			log.Fatalf("error reading row: %v\n", err)
		}
		ch := d.Channel(data.IHemi, data.IASM, data.IDRS, data.IQuartet, data.IChannel)
		iSample := int(data.ISample)
		ch.SetTimeDepOffsetMeanErr(iSample, data.Mean, data.MeanErr)
	}
	err = rows.Err()
	if err != nil && err.Error() != "EOF" {
		log.Fatalf("error: %v\n", err)
	}
}

type EnergyCalibFile struct {
	IChannelAbs240      uint16
	ADCper511keVMean    float64
	ADCper511keVMeanErr float64
}

func (d *Detector) ReadEnergyCalibFile(fileName string) {
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

	var data EnergyCalibFile

	for rows.Next() {
		err = rows.Scan(&data)
		if err != nil {
			log.Fatalf("error reading row: %v\n", err)
		}
		ch := d.ChannelFromIdAbs240(data.IChannelAbs240)
		switch data.ADCper511keVMean != 0 {
		case true:
			ch.EnergyCalib.A = 511 / data.ADCper511keVMean
		case false:
			ch.EnergyCalib.A = 0
		}
		ch.EnergyCalib.B = 0
	}
	err = rows.Err()
	if err != nil && err.Error() != "EOF" {
		log.Fatalf("error: %v\n", err)
	}
}

type HVSerialChan struct {
	SerialNumber  uint8
	ChannelNumber uint8
}

var HVmap map[uint]HVSerialChan

var Det *Detector

func init() {
	Det = NewDetector()
	//Det.DumpGeom()
	//Det.DumpFullGeom()
	//Det.Print()

	HVmap = make(map[uint]HVSerialChan)
	HVmap[0] = HVSerialChan{SerialNumber: 1, ChannelNumber: 1}
	HVmap[1] = HVSerialChan{SerialNumber: 1, ChannelNumber: 2}
	HVmap[2] = HVSerialChan{SerialNumber: 1, ChannelNumber: 3}
	HVmap[3] = HVSerialChan{SerialNumber: 1, ChannelNumber: 4}
	HVmap[4] = HVSerialChan{SerialNumber: 1, ChannelNumber: 5}
	HVmap[5] = HVSerialChan{SerialNumber: 4, ChannelNumber: 15}
	HVmap[6] = HVSerialChan{SerialNumber: 4, ChannelNumber: 14}
	HVmap[7] = HVSerialChan{SerialNumber: 4, ChannelNumber: 13}
	HVmap[8] = HVSerialChan{SerialNumber: 4, ChannelNumber: 12}
	HVmap[9] = HVSerialChan{SerialNumber: 4, ChannelNumber: 11}
	HVmap[10] = HVSerialChan{SerialNumber: 1, ChannelNumber: 6}
	HVmap[11] = HVSerialChan{SerialNumber: 1, ChannelNumber: 7}
	HVmap[12] = HVSerialChan{SerialNumber: 1, ChannelNumber: 8}
	HVmap[13] = HVSerialChan{SerialNumber: 1, ChannelNumber: 9}
	HVmap[14] = HVSerialChan{SerialNumber: 1, ChannelNumber: 10}
	HVmap[15] = HVSerialChan{SerialNumber: 4, ChannelNumber: 10} // this inversion between 15 and 16 is weird -> irrelevant now
	HVmap[16] = HVSerialChan{SerialNumber: 4, ChannelNumber: 9}  // this inversion between 15 and 16 is weird -> irrelevant now
	HVmap[17] = HVSerialChan{SerialNumber: 4, ChannelNumber: 8}
	HVmap[18] = HVSerialChan{SerialNumber: 4, ChannelNumber: 7}
	HVmap[19] = HVSerialChan{SerialNumber: 4, ChannelNumber: 6}
	HVmap[20] = HVSerialChan{SerialNumber: 1, ChannelNumber: 11}
	HVmap[21] = HVSerialChan{SerialNumber: 1, ChannelNumber: 12}
	HVmap[22] = HVSerialChan{SerialNumber: 1, ChannelNumber: 13}
	HVmap[23] = HVSerialChan{SerialNumber: 1, ChannelNumber: 14}
	HVmap[24] = HVSerialChan{SerialNumber: 1, ChannelNumber: 15}
	HVmap[25] = HVSerialChan{SerialNumber: 4, ChannelNumber: 5}
	HVmap[26] = HVSerialChan{SerialNumber: 4, ChannelNumber: 4}
	HVmap[27] = HVSerialChan{SerialNumber: 4, ChannelNumber: 3}
	HVmap[28] = HVSerialChan{SerialNumber: 4, ChannelNumber: 2}
	HVmap[29] = HVSerialChan{SerialNumber: 4, ChannelNumber: 1}
	HVmap[30] = HVSerialChan{SerialNumber: 2, ChannelNumber: 1}
	HVmap[31] = HVSerialChan{SerialNumber: 2, ChannelNumber: 2}
	HVmap[32] = HVSerialChan{SerialNumber: 2, ChannelNumber: 3}
	HVmap[33] = HVSerialChan{SerialNumber: 2, ChannelNumber: 4}
	HVmap[34] = HVSerialChan{SerialNumber: 2, ChannelNumber: 5}
	HVmap[35] = HVSerialChan{SerialNumber: 3, ChannelNumber: 15}
	HVmap[36] = HVSerialChan{SerialNumber: 3, ChannelNumber: 14}
	HVmap[37] = HVSerialChan{SerialNumber: 3, ChannelNumber: 13}
	HVmap[38] = HVSerialChan{SerialNumber: 3, ChannelNumber: 12}
	HVmap[39] = HVSerialChan{SerialNumber: 3, ChannelNumber: 11}
	HVmap[40] = HVSerialChan{SerialNumber: 2, ChannelNumber: 6}
	HVmap[41] = HVSerialChan{SerialNumber: 2, ChannelNumber: 7}
	HVmap[42] = HVSerialChan{SerialNumber: 2, ChannelNumber: 8}
	HVmap[43] = HVSerialChan{SerialNumber: 2, ChannelNumber: 9}
	HVmap[44] = HVSerialChan{SerialNumber: 2, ChannelNumber: 10}
	HVmap[45] = HVSerialChan{SerialNumber: 3, ChannelNumber: 10}
	HVmap[46] = HVSerialChan{SerialNumber: 3, ChannelNumber: 9}
	HVmap[47] = HVSerialChan{SerialNumber: 3, ChannelNumber: 8}
	HVmap[48] = HVSerialChan{SerialNumber: 3, ChannelNumber: 7}
	HVmap[49] = HVSerialChan{SerialNumber: 3, ChannelNumber: 6}
	HVmap[50] = HVSerialChan{SerialNumber: 2, ChannelNumber: 11}
	HVmap[51] = HVSerialChan{SerialNumber: 2, ChannelNumber: 12}
	HVmap[52] = HVSerialChan{SerialNumber: 2, ChannelNumber: 13}
	HVmap[53] = HVSerialChan{SerialNumber: 2, ChannelNumber: 14}
	HVmap[54] = HVSerialChan{SerialNumber: 2, ChannelNumber: 15}
	HVmap[55] = HVSerialChan{SerialNumber: 3, ChannelNumber: 5}
	HVmap[56] = HVSerialChan{SerialNumber: 3, ChannelNumber: 4}
	HVmap[57] = HVSerialChan{SerialNumber: 3, ChannelNumber: 3}
	HVmap[58] = HVSerialChan{SerialNumber: 3, ChannelNumber: 2}
	HVmap[59] = HVSerialChan{SerialNumber: 3, ChannelNumber: 1}
}
