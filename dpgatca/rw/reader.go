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
		blk.UDPPayloadSize = n
		if r.err != nil {
			panic(err)
		}
		// 	for i := range r.UDPHalfDRSBuffer {
		// 		fmt.Printf(" r.UDPHalfDRSBuffer[%v] = %x \n", i, r.UDPHalfDRSBuffer[i])
		// 	}
	}

	r.readBlockHeader(blk)
	r.err = blk.IntegrityHeader()
	if r.err != nil {
		fmt.Println("IntegrityHeader check failed")
		blk.Print("short")
		return
	}
	r.readBlockData(blk)
	r.err = blk.IntegrityData()
	if r.err != nil {
		fmt.Println("IntegrityData check failed")
		blk.Print("short")
		return
	}
	r.readBlockTrailer(blk)
	r.err = blk.IntegrityTrailer()
	if r.err != nil {
		fmt.Println("IntegrityTrailer check failed")
		blk.Print("medium")
		return
	}
	return f, r.err
}

func (r *Reader) readBlockHeader(blk *Block) {
	switch r.ReadMode {
	case UDPHalfDRS:
		blk.FirstBlockWord = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[0:2])
		blk.AMCFrameCounters[0] = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[2:4])
		blk.AMCFrameCounters[1] = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[4:6])
		blk.ParityFEIdCtrl = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[6:8])
		blk.TriggerMode = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[8:10])
		blk.Trigger = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[10:12])
		blk.ASMFrameCounters[0] = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[12:14])
		blk.ASMFrameCounters[1] = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[14:16])
		blk.ASMFrameCounters[2] = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[16:18])
		blk.ASMFrameCounters[3] = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[18:20])
		blk.Cafe = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[20:22])
		blk.Deca = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[22:24])
		blk.Counters[0] = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[24:26])
		blk.Counters[1] = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[26:28])
		blk.Counters[2] = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[28:30])
		blk.Counters[3] = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[30:32])
		blk.TimeStamps[0] = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[32:34])
		blk.TimeStamps[1] = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[34:36])
		blk.TimeStamps[2] = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[36:38])
		blk.TimeStamps[3] = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[38:40])
		blk.NoSamples = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[40:42])
	case Default:
		r.readU16(&blk.FirstBlockWord)
		r.read(&blk.AMCFrameCounters)
		r.readU16(&blk.ParityFEIdCtrl)
		r.readU16(&blk.TriggerMode)
		r.readU16(&blk.Trigger)
		r.read(&blk.ASMFrameCounters)
		r.readU16(&blk.Cafe)
		r.readU16(&blk.Deca)
		r.read(&blk.Counters)
		r.read(&blk.TimeStamps)
		r.readU16(&blk.NoSamples)
	}
	blk.AMCFrameCounter = (uint32(blk.AMCFrameCounters[0]) << 16) + uint32(blk.AMCFrameCounters[1])
	blk.FrontEndId = (blk.ParityFEIdCtrl & 0x7fff) >> 8
	blk.ASMFrameCounter = (uint64(blk.ASMFrameCounters[0]) << 48) + (uint64(blk.ASMFrameCounters[1]) << 32) + (uint64(blk.ASMFrameCounters[2]) << 16) + uint64(blk.ASMFrameCounters[3])
	temp := (uint64(blk.TimeStamps[0]) << 16) | uint64(blk.TimeStamps[1])
	temp = (temp << 32)
	temp1 := (uint64(blk.TimeStamps[2]) << 16) | uint64(blk.TimeStamps[3])
	// 	temp |= temp1
	blk.TimeStamp = temp | temp1
	///////////////////////////////////////////////////////////////////////
	// This +11 is necessary but currently not really understood
	// 11 clock periods are generated by "machine d'etat" in ASM firmware
	// These additionnal 11 samples should currently be considered junk
	blk.Data.SetNoSamples(blk.NoSamples + 11)
	///////////////////////////////////////////////////////////////////////
}

var (
	noAttempts         int
	QuartetAbsIdx60old uint8
)

// readParityChanIdCtrl is a temporary fix, until we understand where the additionnal 16 bits words come from
func (r *Reader) readParityChanIdCtrl(blk *Block, i int) bool {
	data := &blk.Data.Data[i]
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
	if data.Channel != blk.Data.Data[0].Channel+uint16(i) {
		//panic("reader.readParityChanIdCtrl: data.Channel != blk.Data.Data[0].Channel+uint16(i)")
		return true
	}
	blk.QuartetAbsIdx60 = dpgadetector.FEIdAndChanIdToQuartetAbsIdx60(blk.FrontEndId, data.Channel)
	//fmt.Printf("   -> %v, %v, %v\n", data.Channel, blk.QuartetAbsIdx60, QuartetAbsIdx60old)
	if i > 0 && blk.QuartetAbsIdx60 != QuartetAbsIdx60old {
		//panic("i > 0 && blk.QuartetAbsIdx60 != QuartetAbsIdx60old")
		return true
	}
	QuartetAbsIdx60old = blk.QuartetAbsIdx60
	return false
}

func (r *Reader) readBlockData(blk *Block) {
	if r.err != nil {
		return
	}
	//blk.Print("short")
	for i := range blk.Data.Data {
		data := &blk.Data.Data[i]
		for r.readParityChanIdCtrl(blk, i) {
			noAttempts++
			if noAttempts >= 4 {
				log.Fatalf("reader.readParityChanIdCtrl: noAttempts >= 4\n")
			}
		}
		if noAttempts == 1 {
			blk.Err = ErrorCode1
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

func (r *Reader) readBlockTrailer(blk *Block) {
	switch r.ReadMode {
	case UDPHalfDRS:
		if blk.Err == ErrorCode1 {
			blk.CRC = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[len(r.UDPHalfDRSBuffer)-4 : len(r.UDPHalfDRSBuffer)-2])
			blk.ParityFEIdCtrl2 = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[len(r.UDPHalfDRSBuffer)-2 : len(r.UDPHalfDRSBuffer)])
		} else {
			blk.CRC = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[len(r.UDPHalfDRSBuffer)-12 : len(r.UDPHalfDRSBuffer)-10])
			blk.ParityFEIdCtrl2 = binary.BigEndian.Uint16(r.UDPHalfDRSBuffer[len(r.UDPHalfDRSBuffer)-10 : len(r.UDPHalfDRSBuffer)-8])
		}
	case Default:
		r.readU16(&blk.CRC)
		// Temporary fix, until we understand where these additionnal 16 bits come from
		if blk.CRC != ctrl0xCRC {
			//fmt.Printf("CRC = %x (should be %x)\n", blk.CRC, ctrl0xCRC)
			r.readU16(&blk.CRC)
			//fmt.Printf("new CRC = %x\n", blk.CRC)
		}
		// End of temporary fix
		r.readU16(&blk.ParityFEIdCtrl2)
	}
}
