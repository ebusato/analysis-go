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
			quartet.DRS = drs
			for iChannel := range quartet.Channels() {
				ch := quartet.Channel(uint8(iChannel))
				ch.Quartet = quartet
				ch.SetName("PMT" + strconv.FormatUint(uint64(iChannel), 10))
				ch.SetID(uint8(iChannel))
				ch.SetFifoID144(FifoID(uint8(iDRS), uint8(iQuartet), uint8(iChannel)))
				for iCapacitor := range ch.Capacitors() {
					capa := ch.Capacitor(uint16(iCapacitor))
					capa.SetID(uint16(iCapacitor))
					capa.Channel = ch
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

// FinalizePedestalsMeanErr finalizes the computations of the pedestals for all capacitors.
func (d *Detector) FinalizePedestalsMeanErr() {
	for iDRS := range d.asm.DRSs() {
		drs := d.asm.DRS(uint8(iDRS))
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

// FinalizeTimeDepOffsetsMeanErr finalizes the computations of the time dependent offsets for all channels.
func (d *Detector) FinalizeTimeDepOffsetsMeanErr() {
	for iDRS := range d.asm.DRSs() {
		drs := d.asm.DRS(uint8(iDRS))
		for iQuartet := range drs.Quartets() {
			quartet := drs.Quartet(uint8(iQuartet))
			for iChannel := range quartet.Channels() {
				ch := quartet.Channel(uint8(iChannel))
				ch.FinalizeTimeDepOffsetMeanErr()
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
	for iDRS := range d.asm.DRSs() {
		drs := d.asm.DRS(uint8(iDRS))
		for iQuartet := range drs.Quartets() {
			quartet := drs.Quartet(uint8(iQuartet))
			for iChannel := range quartet.Channels() {
				ch := quartet.Channel(uint8(iChannel))
				ch.SetNoSamples(noSamples)
			}
		}
	}
}

type PedestalFile struct {
	IDRS       uint8
	IQuartet   uint8
	IChannel   uint8
	ICapacitor uint16
	Mean       float64
	MeanErr    float64
}

type PedROOTData struct {
	IDRS            float64
	IQuartet        float64
	IChannel        float64
	ICapacitor      float64
	PedestalSamples []float64
}

func (p *PedROOTData) Print() {
	fmt.Println(p.IDRS, p.IQuartet, p.IChannel, p.ICapacitor)
}

func (d *Detector) WritePedestalsToFile(outFileName string, inFileName string, outrootfileName string) {
	tbl, err := csvutil.Create(outFileName)
	if err != nil {
		log.Fatalf("could not create %s: %v\n", outFileName, err)
	}
	defer tbl.Close()
	tbl.Writer.Comma = ' '

	err = tbl.WriteHeader(fmt.Sprintf("# Test bench pedestal file (creation date: %v, input file: %v)\n", time.Now(), inFileName))
	err = tbl.WriteHeader("# iDRS iQuartet iChannel iCapacitor pedestalMean pedestalMeanErr")

	if err != nil {
		log.Fatalf("error writing header: %v\n", err)
	}

	/*ofile, err := croot.OpenFile(outrootfileName, "recreate", "Pedestal file", 1, 0)
	if err != nil {
		panic(err)
	}
	defer ofile.Close("")

	tree := croot.NewTree("tree", "tree", 32)
	peddata := PedROOTData{}
	const bufsiz = 32000

	_, err = tree.Branch("data", &peddata, bufsiz, 0)
	if err != nil {
		panic(err)
	}
	*/
	for iDRS := range d.asm.DRSs() {
		drs := d.asm.DRS(uint8(iDRS))
		for iQuartet := range drs.Quartets() {
			quartet := drs.Quartet(uint8(iQuartet))
			for iChannel := range quartet.Channels() {
				ch := quartet.Channel(uint8(iChannel))
				for iCapacitor := range ch.Capacitors() {
					capa := ch.Capacitor(uint16(iCapacitor))
					data := PedestalFile{
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

					/*
						peddata.IDRS = float64(iDRS)
						peddata.IQuartet = float64(iQuartet)
						peddata.IChannel = float64(iChannel)
						peddata.ICapacitor = float64(iCapacitor)
						//peddata.PedestalSamples = capa.PedestalSamples()
						//peddata.Print()
						_, err = tree.Fill()
						if err != nil {
							panic(err)
						}
					*/
				}
			}
		}
	}

	err = tbl.Close()
	if err != nil {
		log.Fatalf("error closing table: %v\n", err)
	}

	//ofile.Write("", 0, 0)
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
		//fmt.Printf("data: %+v\n", data)
		capacitor := d.Capacitor(data.IDRS, data.IQuartet, data.IChannel, data.ICapacitor)
		capacitor.SetPedestalMeanErr(data.Mean, data.MeanErr)
	}
	err = rows.Err()
	if err != nil && err.Error() != "EOF" {
		log.Fatalf("error: %v\n", err)
	}
}

type TimeDepOffsetFile struct {
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

	err = tbl.WriteHeader(fmt.Sprintf("# Test bench time dependent offset file (creation date: %v, input file: %v)\n", time.Now(), inFileName))
	err = tbl.WriteHeader("# iDRS iQuartet iChannel iSample timeDepOffsetMean timeDepOffsetMeanErr")

	for iDRS := range d.asm.DRSs() {
		drs := d.asm.DRS(uint8(iDRS))
		for iQuartet := range drs.Quartets() {
			quartet := drs.Quartet(uint8(iQuartet))
			for iChannel := range quartet.Channels() {
				ch := quartet.Channel(uint8(iChannel))
				for iSample := 0; iSample < len(ch.TimeDepOffsetMeans()); iSample++ {
					data := TimeDepOffsetFile{
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
		ch := d.Channel(data.IDRS, data.IQuartet, data.IChannel)
		iSample := int(data.ISample)
		ch.SetTimeDepOffsetMeanErr(iSample, data.Mean, data.MeanErr)
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
