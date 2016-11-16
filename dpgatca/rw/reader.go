package rw

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"

	"gitlab.in2p3.fr/avirm/analysis-go/dpga/dpgadetector"
)

type ReadMode byte

const (
	Default ReadMode = iota
	UDPHalfDRS
)

// Reader wraps an io.Reader and reads avirm data files
type Reader struct {
	r   io.Reader
	err error
	//hdr               Header
	noSamples uint16
	//evtIDPrevFrame    uint32
	//firstFrameOfEvent *Frame
	SigThreshold     uint
	Debug            bool
	ReadMode         ReadMode
	UDPHalfDRSBuffer []byte // relevant only when reading from UDP with packet = half DRS
}

// NoSamples returns the number of samples
func (r *Reader) NoSamples() uint16 {
	return r.noSamples
}

// Err return the reader error
func (r *Reader) Err() error {
	return r.err
}

// NewReader returns a new ASM stream in read mode
func NewReader(r io.Reader) (*Reader, error) {
	rr := &Reader{
		r: r,
		//evtIDPrevFrame: 0,
		SigThreshold:     800,
		ReadMode:         Default,
		UDPHalfDRSBuffer: make([]byte, 8238),
	}
	//rr.readHeader(&rr.hdr)
	return rr, rr.err
}

// Read implements io.Reader
// func (r *Reader) Read(data []byte) (int, error) {
// 	return r.r.Read(data)
// }

func (r *Reader) read(v interface{}) {
	if r.err != nil {
		return
	}
	r.err = binary.Read(r.r, binary.BigEndian, v)
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

func (r *Reader) Read(v interface{}) {
	r.read(v)
}

func (r *Reader) readU16(v *uint16) {
	if r.err != nil {
		return
	}
	var buf [2]byte
	_, r.err = r.r.Read(buf[:])
	if r.err != nil {
		return
	}
	*v = binary.BigEndian.Uint16(buf[:])
	if r.Debug {
		fmt.Printf("word = %x\n", *v)
	}
}

func (r *Reader) ReadU16(v *uint16) {
	r.readU16(v)
}

func (r *Reader) Frame() (*Frame, error) {
	f := &Frame{}
	if r.Debug {
		fmt.Printf("rw: start reading frame\n")
	}
	if r.ReadMode == UDPHalfDRS {
		for i := range r.UDPHalfDRSBuffer {
			r.UDPHalfDRSBuffer[i] = 0
		}
		n, err := r.r.Read(r.UDPHalfDRSBuffer)
		f.UDPPayloadSize = n
		if r.err != nil {
			panic(err)
		}
		// 	for i := range r.UDPHalfDRSBuffer {
		// 		fmt.Printf(" r.UDPHalfDRSBuffer[%v] = %x \n", i, r.UDPHalfDRSBuffer[i])
		// 	}
	}

	r.readHeader(f)
	r.err = f.IntegrityHeader()
	if r.err != nil {
		fmt.Println("IntegrityHeader check failed")
		f.Print("short")
		return nil, nil
	}
	r.readData(f)
	r.err = f.IntegrityData()
	if r.err != nil {
		fmt.Println("IntegrityData check failed")
		f.Print("short")
		return nil, nil
	}
	r.readTrailer(f)
	r.err = f.IntegrityTrailer()
	if r.err != nil {
		fmt.Println("IntegrityTrailer check failed")
		f.Print("medium")
		return nil, nil
	}
	return f, r.err
}

func (r *Reader) readHeader(f *Frame) {
	switch r.ReadMode {
	case UDPHalfDRS:
		f.FirstBlockWord = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[0:2])
		f.AMCFrameCounters[0] = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[2:4])
		f.AMCFrameCounters[1] = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[4:6])
		f.ParityFEIdCtrl = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[6:8])
		f.TriggerMode = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[8:10])
		f.Trigger = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[10:12])
		f.ASMFrameCounters[0] = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[12:14])
		f.ASMFrameCounters[1] = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[14:16])
		f.ASMFrameCounters[2] = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[16:18])
		f.ASMFrameCounters[3] = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[18:20])
		f.Cafe = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[20:22])
		f.Deca = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[22:24])
		f.Counters[0] = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[24:26])
		f.Counters[1] = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[26:28])
		f.Counters[2] = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[28:30])
		f.Counters[3] = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[30:32])
		f.TimeStamps[0] = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[32:34])
		f.TimeStamps[1] = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[34:36])
		f.TimeStamps[2] = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[36:38])
		f.TimeStamps[3] = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[38:40])
		f.NoSamples = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[40:42])
	case Default:
		r.readU16(&f.FirstBlockWord)
		r.read(&f.AMCFrameCounters)
		r.readU16(&f.ParityFEIdCtrl)
		r.readU16(&f.TriggerMode)
		r.readU16(&f.Trigger)
		r.read(&f.ASMFrameCounters)
		r.readU16(&f.Cafe)
		r.readU16(&f.Deca)
		r.read(&f.Counters)
		r.read(&f.TimeStamps)
		r.readU16(&f.NoSamples)
	}
	f.AMCFrameCounter = (uint32(f.AMCFrameCounters[0]) << 16) + uint32(f.AMCFrameCounters[1])
	f.FrontEndId = (f.ParityFEIdCtrl & 0x7fff) >> 8
	f.ASMFrameCounter = (uint64(f.ASMFrameCounters[0]) << 48) + (uint64(f.ASMFrameCounters[1]) << 32) + (uint64(f.ASMFrameCounters[2]) << 16) + uint64(f.ASMFrameCounters[3])
	temp := (uint64(f.TimeStamps[0]) << 16) | uint64(f.TimeStamps[1])
	temp = (temp << 32)
	temp1 := (uint64(f.TimeStamps[2]) << 16) | uint64(f.TimeStamps[3])
	// 	temp |= temp1
	f.TimeStamp = temp | temp1
	///////////////////////////////////////////////////////////////////////
	// This +11 is necessary but currently not really understood
	// 11 clock periods are generated by "machine d'etat" in ASM firmware
	// These additionnal 11 samples should currently be considered junk
	f.Data.SetNoSamples(f.NoSamples + 11)
	///////////////////////////////////////////////////////////////////////
}

var (
	noAttempts         int
	QuartetAbsIdx60old uint8
)

// readParityChanIdCtrl is a temporary fix, until we understand where the additionnal 16 bits words come from
func (r *Reader) readParityChanIdCtrl(f *Frame, i int) bool {
	data := &f.Data.Data[i]
	switch r.ReadMode {
	case UDPHalfDRS:
		data.ParityChanIdCtrl = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[42+i*2*1023+2*noAttempts : 44+i*2*1023+2*noAttempts])
	case Default:
		r.readU16(&data.ParityChanIdCtrl)
	}
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
	f.QuartetAbsIdx60 = dpgadetector.FEIdAndChanIdToQuartetAbsIdx60(f.FrontEndId, data.Channel)
	//fmt.Printf("   -> %v, %v, %v\n", data.Channel, f.QuartetAbsIdx60, QuartetAbsIdx60old)
	if i > 0 && f.QuartetAbsIdx60 != QuartetAbsIdx60old {
		//panic("i > 0 && f.QuartetAbsIdx60 != QuartetAbsIdx60old")
		return true
	}
	QuartetAbsIdx60old = f.QuartetAbsIdx60
	return false
}

func (r *Reader) readData(f *Frame) {
	if r.err != nil {
		return
	}
	//f.Print("short")
	for i := range f.Data.Data {
		data := &f.Data.Data[i]
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
		switch r.ReadMode {
		case UDPHalfDRS:
			for j := range data.Amplitudes {
				data.Amplitudes[j] = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[44+2*j+i*2*1023 : 46+2*j+i*2*1023])
			}
		case Default:
			r.read(&data.Amplitudes)
		}
		// 		for j := range data.Amplitudes {
		// 			fmt.Printf("data.Amplitudes[%v] = %x\n", j, data.Amplitudes[j])
		// 		}
	}
}

func (r *Reader) readTrailer(f *Frame) {
	switch r.ReadMode {
	case UDPHalfDRS:
		if f.Err == ErrorCode1 {
			f.CRC = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[len(r.UDPHalfDRSBuffer)-4 : len(r.UDPHalfDRSBuffer)-2])
			f.ParityFEIdCtrl2 = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[len(r.UDPHalfDRSBuffer)-2 : len(r.UDPHalfDRSBuffer)])
		} else {
			f.CRC = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[len(r.UDPHalfDRSBuffer)-12 : len(r.UDPHalfDRSBuffer)-10])
			f.ParityFEIdCtrl2 = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[len(r.UDPHalfDRSBuffer)-10 : len(r.UDPHalfDRSBuffer)-8])
		}
	case Default:
		r.readU16(&f.CRC)
		// Temporary fix, until we understand where these additionnal 16 bits come from
		if f.CRC != ctrl0xCRC {
			//fmt.Printf("CRC = %x (should be %x)\n", f.CRC, ctrl0xCRC)
			r.readU16(&f.CRC)
			//fmt.Printf("new CRC = %x\n", f.CRC)
		}
		// End of temporary fix
		r.readU16(&f.ParityFEIdCtrl2)
	}
}
