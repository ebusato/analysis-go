package rw

import (
	"encoding/binary"
	"fmt"
	"io"
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
}

// NewReader returns a new ASM stream in read mode
func NewReader(r io.Reader) (*Reader, error) {
	rr := &Reader{
		r: r,
		//evtIDPrevFrame: 0,
		SigThreshold:     800,
		ReadMode:         Default,
		UDPHalfDRSBuffer: make([]byte, 8270), //8238),
	}
	rr.readFileHeader(&rr.FileHeader)
	return rr, rr.err
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
	if r.Debug {
		fmt.Printf("word = %x\n", *v)
	}
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
	r.readU16(&f.NbFrameAmcMsb, binary.BigEndian)
	r.readU16(&f.NbFrameAmcLsb, binary.BigEndian)
	r.readU16(&f.FEIdK30, binary.BigEndian)
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
		r.readU16(&chanData.FirstChanWord, binary.BigEndian)
		r.readU16(&chanData.SecondChanWord, binary.BigEndian)
		r.read(&chanData.Amplitudes, binary.BigEndian)
	}
}

func (r *Reader) readFrameTrailer(f *FrameTrailer) {
	r.readU16(&f.Crc, binary.BigEndian)
	// Temporary fix, until we understand where these additionnal 16 bits come from
	if f.Crc != ctrl0xCRC {
		//fmt.Printf("CRC = %x (should be %x)\n", f.CRC, ctrl0xCRC)
		r.readU16(&f.Crc, binary.BigEndian)
		//fmt.Printf("new CRC = %x\n", f.CRC)
	}
	// End of temporary fix
	r.readU16(&f.EoF, binary.BigEndian)
}

func (r *Reader) Frame() (*Frame, error) {
	f := &Frame{}
	if r.Debug {
		fmt.Printf("rw: start reading frame\n")
	}
	switch r.ReadMode {
	case Default:
		r.readFrameHeader(&f.Header)
		r.err = f.Header.Integrity()
		if r.err != nil {
			f.Header.Print()
			panic(r.err)
		}
		f.SetDataSliceLen(int(f.Header.NoSamples))
		r.readFrameData(&f.Data)
		r.readFrameTrailer(&f.Trailer)
		r.err = f.Trailer.Integrity()
		if r.err != nil {
			f.Trailer.Print()
			panic(r.err)
		}
		// 		r.err = f.IntegrityData()
		// 		if r.err != nil {
		// 			fmt.Println("IntegrityData check failed")
		// 			f.Print("short")
		// 			return nil, nil
		// 		}
		// 		r.readTrailer(f)
		// 		r.err = f.IntegrityTrailer()
		// 		if r.err != nil {
		// 			fmt.Println("IntegrityTrailer check failed")
		// 			f.Print("medium")
		// 			return nil, nil
		// 		}
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
	return f, r.err
}

/*
var (
	noAttempts         int
	QuartetAbsIdx60old uint8
)

// readParityChanIdCtrl is a temporary fix, until we understand where the additionnal 16 bits words come from
func (r *Reader) readParityChanIdCtrl(f *Frame, i int) bool {
	data := &f.Data.Data[i]
	r.readU16(&data.ParityChanIdCtrl, binary.BigEndian)

	//fmt.Printf("%v, %x (noAttempts=%v)\n", i, data.ParityChanIdCtrl, noAttempts)
	if (data.ParityChanIdCtrl & 0xff) != ctrl0xfd {
		//panic("(data.ParityChanIdCtrl & 0xff) != ctrl0xfd")
		return true
	}
	data.Channel = (data.ParityChanIdCtrl & 0x7f00) >> 8
	if data.Channel != f.Data.Data[0].Channel+uint16(i) {
		//panic("reader.readParityChanIdCtrl: data.Channel != f.Data.Data[0].Channel+uint16(i)")
		return true
	}
	f.QuartetAbsIdx60 = dpgadetector.FEIdAndChanIdToQuartetAbsIdx60(f.FrontEndId, data.Channel, false)
	//fmt.Printf("   -> %v, %v, %v\n", data.Channel, f.QuartetAbsIdx60, QuartetAbsIdx60old)
	if i > 0 && f.QuartetAbsIdx60 != QuartetAbsIdx60old {
		//panic("i > 0 && f.QuartetAbsIdx60 != QuartetAbsIdx60old")
		return true
	}
	QuartetAbsIdx60old = f.QuartetAbsIdx60
	return false
}
*/
