package rw

import (
	"encoding/binary"
	"fmt"
	"log"
)

const (
	numAMCFrameCounters uint8  = 2
	numASMFrameCounters uint8  = 4
	numCounters         uint8  = 4
	numTimeStamps       uint8  = 4
	ctrlFirstWord       uint16 = 0x1230
	ctrl0xfe            uint16 = 0xfe
	ctrl0xfd            uint16 = 0xfd
	ctrl0xCafe          uint16 = 0xCAFE
	ctrl0xDeca          uint16 = 0xDECA
	ctrl0xCRC           uint16 = 0x9876
	ctrl0xfb            uint16 = 0xfb
)

type ChanData struct {
	// Raw quantities
	ParityChanIdCtrl uint16
	Amplitudes       []uint16

	// Derived quantities
	Channel uint16
}

type HalfDRSData struct {
	Data [4]ChanData
}

func (h *HalfDRSData) SetNoSamples(n uint16) {
	for i := range h.Data {
		h.Data[i].Amplitudes = make([]uint16, n)
	}
}

type ErrorCode int

const (
	ErrorCode1 ErrorCode = iota + 1 // value of error code if block has 4 extra 16 bits words after each sample block
)

// Frame is a single data frame produced by AMC
// Each frame is associated to one half DRS
type Frame struct {
	// Raw quantities
	FirstBlockWord   uint16
	AMCFrameCounters [numAMCFrameCounters]uint16
	ParityFEIdCtrl   uint16
	TriggerMode      uint16
	Trigger          uint16
	ASMFrameCounters [numASMFrameCounters]uint16
	Cafe             uint16
	Deca             uint16
	Counters         [numCounters]uint16
	TimeStamps       [numTimeStamps]uint16
	NoSamples        uint16
	Data             HalfDRSData
	CRC              uint16
	ParityFEIdCtrl2  uint16

	// Derived quantities
	AMCFrameCounter uint32
	FrontEndId      uint16
	ASMFrameCounter uint64
	TimeStamp       uint64
	QuartetAbsIdx60 uint8

	// Error handling
	Err ErrorCode

	// UDP Payload size in octects
	UDPPayloadSize int
}

func NewFrame(udppayloadsize int, buffer []byte) *Frame {
	f := &Frame{}
	f.UDPPayloadSize = udppayloadsize
	f.ReadHeader(buffer)
	err := f.IntegrityHeader()
	if err != nil {
		fmt.Println("IntegrityHeader check failed")
		f.Print("short")
		return nil
	}
	return f
}

func (f *Frame) IntegrityHeader() error {
	if f.FirstBlockWord != ctrlFirstWord {
		return fmt.Errorf("asm: missing %x magic\n", ctrlFirstWord)
	}
	if (f.ParityFEIdCtrl & 0xff) != ctrl0xfe {
		return fmt.Errorf("asm: missing %x magic\n", ctrl0xfe)
	}
	if f.Cafe != ctrl0xCafe {
		return fmt.Errorf("asm: missing %x magic\n", ctrl0xCafe)
	}
	if f.Deca != ctrl0xDeca {
		return fmt.Errorf("asm: missing %x magic\n", ctrl0xDeca)
	}
	return nil
}

func (f *Frame) IntegrityData() error {
	for i := range f.Data.Data {
		if (f.Data.Data[i].ParityChanIdCtrl & 0xff) != ctrl0xfd {
			return fmt.Errorf("asm: missing %x magic\n", ctrl0xfd)
		}
	}
	return nil
}

func (f *Frame) IntegrityTrailer() error {
	if f.CRC != ctrl0xCRC {
		return fmt.Errorf("asm: missing %x magic\n", ctrl0xCRC)
	}
	if (f.ParityFEIdCtrl2 & 0xff) != ctrl0xfb {
		return fmt.Errorf("asm: missing %x magic\n", ctrl0xfb)
	}
	if (f.ParityFEIdCtrl2&0x7fff)>>8 != f.FrontEndId {
		log.Fatalf("Front end ids in header and trailer don't match\n")
	}
	return nil
}

func (f *Frame) ReadHeader(buffer []byte) {
	f.FirstBlockWord = binary.BigEndian.Uint16(buffer[0:2])
	f.AMCFrameCounters[0] = binary.BigEndian.Uint16(buffer[2:4])
	f.AMCFrameCounters[1] = binary.BigEndian.Uint16(buffer[4:6])
	f.ParityFEIdCtrl = binary.BigEndian.Uint16(buffer[6:8])
	f.TriggerMode = binary.BigEndian.Uint16(buffer[8:10])
	f.Trigger = binary.BigEndian.Uint16(buffer[10:12])
	f.ASMFrameCounters[0] = binary.BigEndian.Uint16(buffer[12:14])
	f.ASMFrameCounters[1] = binary.BigEndian.Uint16(buffer[14:16])
	f.ASMFrameCounters[2] = binary.BigEndian.Uint16(buffer[16:18])
	f.ASMFrameCounters[3] = binary.BigEndian.Uint16(buffer[18:20])
	f.Cafe = binary.BigEndian.Uint16(buffer[20:22])
	f.Deca = binary.BigEndian.Uint16(buffer[22:24])
	f.Counters[0] = binary.BigEndian.Uint16(buffer[24:26])
	f.Counters[1] = binary.BigEndian.Uint16(buffer[26:28])
	f.Counters[2] = binary.BigEndian.Uint16(buffer[28:30])
	f.Counters[3] = binary.BigEndian.Uint16(buffer[30:32])
	f.TimeStamps[0] = binary.BigEndian.Uint16(buffer[32:34])
	f.TimeStamps[1] = binary.BigEndian.Uint16(buffer[34:36])
	f.TimeStamps[2] = binary.BigEndian.Uint16(buffer[36:38])
	f.TimeStamps[3] = binary.BigEndian.Uint16(buffer[38:40])
	f.NoSamples = binary.BigEndian.Uint16(buffer[40:42])
}

func (f *Frame) Print(s string) {
	fmt.Printf(" Printing block (UDP payload size=%v):\n", f.UDPPayloadSize)
	fmt.Printf("   -> FirstBlockWord = %x\n", f.FirstBlockWord)
	fmt.Printf("   -> AMCFrameCounters = %x (AMCFrameCounter = %v)\n", f.AMCFrameCounters, f.AMCFrameCounter)
	fmt.Printf("   -> ParityFEIdCtrl = %x (FrontEndId = %x)\n", f.ParityFEIdCtrl, f.FrontEndId)
	fmt.Printf("   -> TriggerMode = %x\n", f.TriggerMode)
	fmt.Printf("   -> Trigger = %x\n", f.Trigger)
	fmt.Printf("   -> ASMFrameCounters = %x (ASMFrameCounter = %v)\n", f.ASMFrameCounters, f.ASMFrameCounter)
	fmt.Printf("   -> Cafe = %x\n", f.Cafe)
	fmt.Printf("   -> Deca = %x\n", f.Deca)
	fmt.Printf("   -> Counters = %x\n", f.Counters)
	fmt.Printf("   -> TimeStamps = %x (TimeStamp = %v)\n", f.TimeStamps, f.TimeStamp)
	fmt.Printf("   -> NoSamples = %x\n", f.NoSamples)

	switch s {
	case "short":
	case "medium":
		for i := range f.Data.Data {
			data := &f.Data.Data[i]
			fmt.Printf("   -> ParityChanIdCtrl = %x (channel = %v)\n", data.ParityChanIdCtrl, data.Channel)
			fmt.Printf("   -> Amplitudes[0] = %x\n", data.Amplitudes[0])
			fmt.Printf("   -> Amplitudes[1] = %x\n", data.Amplitudes[1])
			fmt.Printf("   -> Amplitudes[2] = %x\n", data.Amplitudes[2])
			fmt.Printf("   -> Amplitudes[3] = %x\n", data.Amplitudes[3])
			fmt.Printf("   ->    ...\n")
			fmt.Printf("   -> Amplitudes[1008] = %x\n", data.Amplitudes[1008])
			fmt.Printf("   -> Amplitudes[1009] = %x\n", data.Amplitudes[1009])
			fmt.Printf("   -> Amplitudes[1010] = %x\n", data.Amplitudes[1010])
		}
		/*
			case "long":
				fmt.Printf("  Data %v = %x\n", 0, f.Data[0])
				fmt.Printf("  Data %v = %x\n", 1, f.Data[1])
				fmt.Printf("  Data %v = %x\n", 2, f.Data[2])
				fmt.Printf("  Data %v = %x\n", 3, f.Data[3])
				fmt.Println("\t.\n\t.")
				fmt.Printf("  Data %v = %x\n", len(f.Data)-3, f.Data[len(f.Data)-3])
				fmt.Printf("  Data %v = %x\n", len(f.Data)-2, f.Data[len(f.Data)-2])
				fmt.Printf("  Data %v = %x\n", len(f.Data)-1, f.Data[len(f.Data)-1])
				fmt.Printf("  SRout = %v\n", f.SRout)
				for i := range f.Counters {
					fmt.Printf("  Counter %v = %v\n", i, f.Counters[i])
				}
		*/
	case "full":
		for i := range f.Data.Data {
			data := &f.Data.Data[i]
			fmt.Printf("   -> ParityChanIdCtrl = %x\n", data.ParityChanIdCtrl)
			fmt.Printf("   -> Amplitudes = %x\n", data.Amplitudes)
		}
	}

}

func (f *Frame) Buffer() []byte {
	var buffer []uint16
	buffer = append(buffer, f.FirstBlockWord)
	buffer = append(buffer, f.AMCFrameCounters[:]...)
	buffer = append(buffer, f.ParityFEIdCtrl)
	buffer = append(buffer, f.TriggerMode)
	buffer = append(buffer, f.Trigger)
	buffer = append(buffer, f.ASMFrameCounters[:]...)
	buffer = append(buffer, f.Cafe)
	buffer = append(buffer, f.Deca)
	buffer = append(buffer, f.Counters[:]...)
	buffer = append(buffer, f.TimeStamps[:]...)
	buffer = append(buffer, f.NoSamples)
	for i := range f.Data.Data {
		data := &f.Data.Data[i]
		buffer = append(buffer, data.ParityChanIdCtrl)
		buffer = append(buffer, data.Amplitudes...)
		if f.Err == ErrorCode1 {
			//fmt.Println("ErrorCode1, add extra word")
			buffer = append(buffer, uint16(0))
		}
	}
	buffer = append(buffer, f.CRC)
	buffer = append(buffer, f.ParityFEIdCtrl2)

	var buffer8 []byte
	for i := range buffer {
		buffer8 = append(buffer8, uint8(buffer[i]>>8))
		buffer8 = append(buffer8, uint8(buffer[i]&0xFFFF))
		//fmt.Printf("buffer8 = %x %x\n", buffer8[len(buffer8)-2], buffer8[len(buffer8)-1])
	}
	return buffer8
}
