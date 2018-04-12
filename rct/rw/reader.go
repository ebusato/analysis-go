package rw

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"

	"gitlab.in2p3.fr/avirm/analysis-go/dpga/dpgadetector"
	"gitlab.in2p3.fr/avirm/analysis-go/event"
	"gitlab.in2p3.fr/avirm/analysis-go/pulse"
)

type ReadMode byte

const (
	Default ReadMode = iota
	UDPHalfDRS
)

// Reader wraps an io.Reader and reads avirm data files
type Reader struct {
	r                io.Reader
	err              error
	FileHeader       FileHeader
	SigThreshold     uint
	Debug            bool
	ReadMode         ReadMode
	UDPHalfDRSBuffer []byte // relevant only when reading from UDP with packet = half DRS

	IDPrevFrame       uint32
	firstFrameOfEvent *Frame
}

// NewReader returns a new ASM stream in read mode
func NewReader(r io.Reader) (*Reader, error) {
	rr := &Reader{
		r:                r,
		IDPrevFrame:      0,
		SigThreshold:     800,
		ReadMode:         Default,
		UDPHalfDRSBuffer: make([]byte, 8270), //8238),
	}
	rr.readFileHeader(&rr.FileHeader)
	return rr, rr.err
}

// SetDebug() sets debug mode
func (r *Reader) SetDebug() {
	r.Debug = true
}

// SetSigThreshold sets the signal SetSigThreshold
func (r *Reader) SetSigThreshold(val uint) {
	r.SigThreshold = val
}

// Read implements io.Reader
// func (r *Reader) Read(data []byte) (int, error) {
// 	return r.r.Read(data)
// }

// NoSamples returns the number of samples
func (r *Reader) NoSamples() uint16 {
	return r.FileHeader.NoSamples
}

// Err return the reader error
func (r *Reader) Err() error {
	return r.err
}

func (r *Reader) read(v interface{}, byteOrder binary.ByteOrder) {
	if r.err != nil {
		return
	}
	r.err = binary.Read(r.r, byteOrder, v)
	if r.Debug {
		switch v := v.(type) {
		case *uint32:
			fmt.Printf("word = %x\n", *v)
		case *[]uint32:
			for _, vv := range *v {
				fmt.Printf("word = %x\n", vv)
			}
		}
		//fmt.Printf("word = %x\n", *(v.(*uint32)))
	}
}

func (r *Reader) Read(v interface{}, byteOrder binary.ByteOrder) {
	r.read(v, byteOrder)
}

func (r *Reader) readU16(v *uint16, byteOrder binary.ByteOrder) {
	if r.err != nil {
		return
	}
	var buf [2]byte
	_, r.err = r.r.Read(buf[:])
	if r.err != nil {
		return
	}
	*v = byteOrder.Uint16(buf[:])
	// 	if r.Debug {
	// 		fmt.Printf("word = %x\n", *v)
	// 	}
}

func (r *Reader) ReadU16(v *uint16, byteOrder binary.ByteOrder) {
	r.readU16(v, byteOrder)
}

func (r *Reader) readFileHeader(f *FileHeader) {
	r.read(&f.ModeFile, binary.LittleEndian)
	r.read(&f.FEId, binary.LittleEndian)
	r.readU16(&f.NoSamples, binary.LittleEndian)
	r.read(&f.Time, binary.LittleEndian)
	r.read(&f.Time, binary.LittleEndian)
}

func (r *Reader) readFrameHeader(f *FrameHeader) {
	r.readU16(&f.StartOfFrame, binary.BigEndian)
	if r.err != nil { // true for EOF
		return
	}

	//fmt.Printf("Start of frame = %x\n", f.StartOfFrame)
	r.readU16(&f.NbFrameAmcMsb, binary.BigEndian)
	r.readU16(&f.NbFrameAmcLsb, binary.BigEndian)
	r.readU16(&f.FEIdK30, binary.LittleEndian)
	r.readU16(&f.Mode, binary.BigEndian)
	r.readU16(&f.TriggerType, binary.BigEndian)
	r.readU16(&f.NoFrameAsmMsb, binary.BigEndian)
	r.readU16(&f.NoFrameAsmOsb, binary.BigEndian)
	r.readU16(&f.NoFrameAsmUsb, binary.BigEndian)
	r.readU16(&f.NoFrameAsmLsb, binary.BigEndian)
	r.readU16(&f.Cafe, binary.BigEndian)
	r.readU16(&f.Deca, binary.BigEndian)
	r.readU16(&f.UndefinedMsb, binary.BigEndian)
	r.readU16(&f.UndefinedOsb, binary.BigEndian)
	r.readU16(&f.UndefinedUsb, binary.BigEndian)
	r.readU16(&f.UndefinedLsb, binary.BigEndian)
	r.readU16(&f.TimeStampAsmMsb, binary.BigEndian)
	r.readU16(&f.TimeStampAsmOsb, binary.BigEndian)
	r.readU16(&f.TimeStampAsmUsb, binary.BigEndian)
	r.readU16(&f.TimeStampAsmLsb, binary.BigEndian)
	r.readU16(&f.TimeStampTrigThorAsmMsb, binary.BigEndian)
	r.readU16(&f.TimeStampTrigThorAsmOsb, binary.BigEndian)
	r.readU16(&f.TimeStampTrigThorAsmUsb, binary.BigEndian)
	r.readU16(&f.TimeStampTrigThorAsmLsb, binary.BigEndian)
	r.readU16(&f.ThorTT, binary.BigEndian)
	r.readU16(&f.PatternMsb, binary.BigEndian)
	r.readU16(&f.PatternOsb, binary.BigEndian)
	r.readU16(&f.PatternLsb, binary.BigEndian)
	r.readU16(&f.Bobo, binary.BigEndian)
	r.readU16(&f.ThorTrigTimeStampMsb, binary.BigEndian)
	r.readU16(&f.ThorTrigTimeStampOsb, binary.BigEndian)
	r.readU16(&f.ThorTrigTimeStampLsb, binary.BigEndian)
	r.readU16(&f.CptTriggerThorMsb, binary.BigEndian)
	r.readU16(&f.CptTriggerThorLsb, binary.BigEndian)
	r.readU16(&f.CptTriggerAsmMsb, binary.BigEndian)
	r.readU16(&f.CptTriggerAsmLsb, binary.BigEndian)
	r.readU16(&f.NoSamples, binary.BigEndian)
	// 	f.AMCFrameCounter = (uint32(f.AMCFrameCounters[0]) << 16) + uint32(f.AMCFrameCounters[1])
	// 	f.FrontEndId = (f.ParityFEIdCtrl & 0x7fff) >> 8
	// 	f.ASMFrameCounter = (uint64(f.ASMFrameCounters[0]) << 48) + (uint64(f.ASMFrameCounters[1]) << 32) + (uint64(f.ASMFrameCounters[2]) << 16) + uint64(f.ASMFrameCounters[3])
	// 	temp := (uint64(f.TimeStampsASM[0]) << 16) | uint64(f.TimeStampsASM[1])
	// 	temp = (temp << 32)
	// 	temp1 := (uint64(f.TimeStampsASM[2]) << 16) | uint64(f.TimeStampsASM[3])
	// 	// 	temp |= temp1
	// 	f.TimeStampASM = temp | temp1
	///////////////////////////////////////////////////////////////////////
	// This +11 is necessary but currently not really understood
	// 11 clock periods are generated by "machine d'etat" in ASM firmware
	// These additionnal 11 samples should currently be considered junk
	//f.Data.SetNoSamples(f.NoSamples + 11)
	///////////////////////////////////////////////////////////////////////

	f.FEId = f.FEIdK30 & 0x7f
	f.CptTriggerThor = (uint32(f.CptTriggerThorMsb) << 16) | uint32(f.CptTriggerThorLsb)
}

func (r *Reader) readFrameData(data *HalfDRSData) {
	if r.err != nil {
		return
	}
	//f.Print("short")
	for i := range data.Data {
		chanData := &data.Data[i]
		/*
			for r.readParityChanIdCtrl(f, i) {
				noAttempts++
				if noAttempts >= 4 {
					log.Fatalf("reader.readParityChanIdCtrl: noAttempts >= 4\n")
				}
			}
			if noAttempts == 1 {
				f.Err = ErrorCode1
			}
			noAttempts = 0
			//fmt.Printf("data.ParityChanIdCtrl = %x\n", data.ParityChanIdCtrl)
		*/
		r.readU16(&chanData.FirstChanWord, binary.LittleEndian)
		r.readU16(&chanData.SecondChanWord, binary.BigEndian)
		r.read(&chanData.Amplitudes, binary.BigEndian)

		chanData.Channel = chanData.FirstChanWord & 0x7f
		// 		fmt.Printf("SecondChanWord = %x\n", chanData.SecondChanWord)
		chanData.SRout = chanData.SecondChanWord & 0x3ff
		// 		fmt.Println("SRout here =", chanData.SRout)
	}
}

func (r *Reader) readFrameTrailer(f *FrameTrailer) {
	r.readU16(&f.Crc, binary.BigEndian)
	// Temporary fix, until we understand where these additionnal 16 bits come from
	// 	fmt.Printf("CRC = %x\n", f.Crc)
	if f.Crc != ctrl0xCRC {
		//fmt.Printf("CRC = %x (should be %x)\n", f.CRC, ctrl0xCRC)
		r.readU16(&f.Crc, binary.BigEndian)
		// 		fmt.Printf("new CRC = %x\n", f.Crc)
	}
	// End of temporary fix
	r.readU16(&f.EoF, binary.BigEndian)
}

func (r *Reader) Frame() *Frame {
	f := &Frame{}
	if r.Debug {
		fmt.Printf("\nrw: start reading frame\n")
	}
	switch r.ReadMode {
	case Default:
		r.readFrameHeader(&f.Header)
		if r.err == io.EOF {
			return nil
		}
		//f.Header.Print()
		r.err = f.Header.Integrity()
		if r.err != nil {
			f.Header.Print()
			panic(r.err)
		}
		f.SetDataSliceLen(int(f.Header.NoSamples))
		r.readFrameData(&f.Data)
		// 		fmt.Println("Channels = ", f.Data.Data[0].Channel, f.Data.Data[1].Channel, f.Data.Data[2].Channel, f.Data.Data[3].Channel)
		f.QuartetAbsIdx60 = dpgadetector.FEIdAndChanIdToQuartetAbsIdx60(f.Header.FEId, f.Data.Data[0].Channel, false)
		f.QuartetAbsIdx72 = dpgadetector.FEIdAndChanIdToQuartetAbsIdx72(f.Header.FEId, f.Data.Data[0].Channel)
		r.readFrameTrailer(&f.Trailer)
		r.err = f.Trailer.Integrity()
		if r.err != nil {
			f.Trailer.Print()
			panic(r.err)
		}
		/*case UDPHalfDRS:
		for i := range r.UDPHalfDRSBuffer {
			r.UDPHalfDRSBuffer[i] = 0
		}
		n, err := r.r.Read(r.UDPHalfDRSBuffer)
		f.UDPPayloadSize = n
		if r.err != nil {
			panic(err)
		}
		f.FillHeader(r.UDPHalfDRSBuffer)
		err = f.IntegrityHeader()
		if err != nil {
			panic(err)
		}
		f.FillData(r.UDPHalfDRSBuffer)
		err = f.IntegrityData()
		if err != nil {
			panic(err)
		}
		f.FillTrailer(r.UDPHalfDRSBuffer)
		err = f.IntegrityTrailer()
		if err != nil {
			panic(err)
		}
		// 	for i := range r.UDPHalfDRSBuffer {
		// 		fmt.Printf(" r.UDPHalfDRSBuffer[%v] = %x \n", i, r.UDPHalfDRSBuffer[i])
		// 	}
		*/
	}
	return f
}

func MakePulse(c *ChanData, quartetAbsIdx72 uint8, sigThreshold uint) *pulse.Pulse {
	iHemi, iASM, iDRS, iQuartet := dpgadetector.QuartetAbsIdx72ToRelIdx(quartetAbsIdx72)
	_, iChannelAbs288 := dpgadetector.RelIdxToAbsIdx288(iHemi, iASM, iDRS, iQuartet, uint8(c.Channel)%4)
	// 	fmt.Println("iChannelAbs288 =", iChannelAbs288)
	if iChannelAbs288 >= 288 {
		panic("reader: iChannelAbs288 >= 288")
	}
	detChannel := dpgadetector.Det.ChannelFromIdAbs288(iChannelAbs288)
	mypulse := pulse.NewPulse(detChannel)
	mypulse.SRout = uint16(c.SRout)
	iChannel := uint8(c.Channel % 4)
	for i := range c.Amplitudes {
		ampl := float64(c.Amplitudes[i])
		sample := pulse.NewSample(ampl, uint16(i), float64(i)*dpgadetector.Det.SamplingFreq())
		mypulse.AddSample(sample, dpgadetector.Det.Capacitor(iHemi, iASM, iDRS, iQuartet, iChannel, sample.CapaIndex(mypulse.SRout)), float64(sigThreshold))
	}
	return mypulse
}

func MakePulses(f *Frame, sigThreshold uint) []*pulse.Pulse {
	var pulses []*pulse.Pulse
	for i := range f.Data.Data {
		data := &f.Data.Data[i]
		pulses = append(pulses, MakePulse(data, f.QuartetAbsIdx72, sigThreshold))
	}
	return pulses
}

var ID uint

// for rct
func (r *Reader) ReadNextEvent() (*event.Event, error) {
	//////////////////////////////////////////////////////
	// Temporary fix:
	// Read first frame and do nothing with it (remove it)
	// 	if ID == 0 {
	// 		r.Frame()
	// 	}
	/////////////////////////////////////////////////////////

	event := event.NewEvent(5, 1)
	event.ID = ID
	event.NoFrames = 2

	var SRout1, SRout2 uint16 // for debug

	for i := 0; i < 4; i++ {
		frame := r.Frame()
		// 		frame.Print()
		pulses := MakePulses(frame, r.SigThreshold)
		if i == 0 {
			SRout1 = pulses[0].SRout
		} else {
			SRout2 = pulses[0].SRout
		}
		if frame.QuartetAbsIdx72%6 != 5 {
			iCluster := frame.QuartetAbsIdx60
			if iCluster >= 60 {
				log.Fatalf("error ! iCluster=%v (>= 60)\n", iCluster)
			}
			// 				fmt.Printf("iCluster = %v\n", iCluster)
			event.Clusters[iCluster].ID = iCluster
			event.Clusters[iCluster].Quartet = dpgadetector.Det.QuartetFromIdAbs60(iCluster)
			// 			fmt.Printf("Quartet in reader %p\n", event.Clusters[iCluster].Quartet)
			////////////////////////////////////////////////////////
			// Put pulses in event
			event.Clusters[iCluster].Pulses[0] = *pulses[0]
			event.Clusters[iCluster].Pulses[1] = *pulses[1]
			event.Clusters[iCluster].Pulses[2] = *pulses[2]
			event.Clusters[iCluster].Pulses[3] = *pulses[3]
			////////////////////////////////////////////////////////
			event.Clusters[iCluster].SetSRout()
		} else {
			iClusterWoData := frame.QuartetAbsIdx72 / 6
			// 				fmt.Printf("iClusterWoData = %v\n", iClusterWoData)
			event.ClustersWoData[iClusterWoData].ID = uint8(iClusterWoData)
			////////////////////////////////////////////////////////
			// Put pulses in event
			event.ClustersWoData[iClusterWoData].Pulses[0] = *pulses[0]
			event.ClustersWoData[iClusterWoData].Pulses[1] = *pulses[1]
			event.ClustersWoData[iClusterWoData].Pulses[2] = *pulses[2]
			event.ClustersWoData[iClusterWoData].Pulses[3] = *pulses[3]
			////////////////////////////////////////////////////////
			event.ClustersWoData[iClusterWoData].SetSRout()
		}
	}

	var err error

	if SRout1 != SRout2 {
		fmt.Printf("SRout1 (%v) != SRout2 (%v)\n", SRout1, SRout2)
		err = errors.New("SRout1 != SRout2")
	}
	ID += 1
	return event, err
}
