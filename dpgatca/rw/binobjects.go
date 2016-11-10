package rw

import (
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

// Block is a single data frame produced by AMC
// Each block is associated to one half DRS
type Block struct {
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

func (b *Block) IntegrityHeader() error {
	if b.FirstBlockWord != ctrlFirstWord {
		return fmt.Errorf("asm: missing %x magic\n", ctrlFirstWord)
	}
	if (b.ParityFEIdCtrl & 0xff) != ctrl0xfe {
		return fmt.Errorf("asm: missing %x magic\n", ctrl0xfe)
	}
	if b.Cafe != ctrl0xCafe {
		return fmt.Errorf("asm: missing %x magic\n", ctrl0xCafe)
	}
	if b.Deca != ctrl0xDeca {
		return fmt.Errorf("asm: missing %x magic\n", ctrl0xDeca)
	}
	return nil
}

func (b *Block) IntegrityData() error {
	for i := range b.Data.Data {
		if (b.Data.Data[i].ParityChanIdCtrl & 0xff) != ctrl0xfd {
			return fmt.Errorf("asm: missing %x magic\n", ctrl0xfd)
		}
	}
	return nil
}

func (b *Block) IntegrityTrailer() error {
	if b.CRC != ctrl0xCRC {
		return fmt.Errorf("asm: missing %x magic\n", ctrl0xCRC)
	}
	if (b.ParityFEIdCtrl2 & 0xff) != ctrl0xfb {
		return fmt.Errorf("asm: missing %x magic\n", ctrl0xfb)
	}
	if (b.ParityFEIdCtrl2&0x7fff)>>8 != b.FrontEndId {
		log.Fatalf("Front end ids in header and trailer don't match\n")
	}
	return nil
}

func (b *Block) Print(s string) {
	fmt.Printf(" Printing block (UDP payload size=%v):\n", b.UDPPayloadSize)
	fmt.Printf("   -> FirstBlockWord = %x\n", b.FirstBlockWord)
	fmt.Printf("   -> AMCFrameCounters = %x (AMCFrameCounter = %v)\n", b.AMCFrameCounters, b.AMCFrameCounter)
	fmt.Printf("   -> ParityFEIdCtrl = %x (FrontEndId = %x)\n", b.ParityFEIdCtrl, b.FrontEndId)
	fmt.Printf("   -> TriggerMode = %x\n", b.TriggerMode)
	fmt.Printf("   -> Trigger = %x\n", b.Trigger)
	fmt.Printf("   -> ASMFrameCounters = %x (ASMFrameCounter = %v)\n", b.ASMFrameCounters, b.ASMFrameCounter)
	fmt.Printf("   -> Cafe = %x\n", b.Cafe)
	fmt.Printf("   -> Deca = %x\n", b.Deca)
	fmt.Printf("   -> Counters = %x\n", b.Counters)
	fmt.Printf("   -> TimeStamps = %x (TimeStamp = %v)\n", b.TimeStamps, b.TimeStamp)
	fmt.Printf("   -> NoSamples = %x\n", b.NoSamples)

	switch s {
	case "short":
	case "medium":
		for i := range b.Data.Data {
			data := &b.Data.Data[i]
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
				fmt.Printf("  Data %v = %x\n", 0, b.Data[0])
				fmt.Printf("  Data %v = %x\n", 1, b.Data[1])
				fmt.Printf("  Data %v = %x\n", 2, b.Data[2])
				fmt.Printf("  Data %v = %x\n", 3, b.Data[3])
				fmt.Println("\t.\n\t.")
				fmt.Printf("  Data %v = %x\n", len(b.Data)-3, b.Data[len(b.Data)-3])
				fmt.Printf("  Data %v = %x\n", len(b.Data)-2, b.Data[len(b.Data)-2])
				fmt.Printf("  Data %v = %x\n", len(b.Data)-1, b.Data[len(b.Data)-1])
				fmt.Printf("  SRout = %v\n", b.SRout)
				for i := range b.Counters {
					fmt.Printf("  Counter %v = %v\n", i, b.Counters[i])
				}
		*/
	case "full":
		for i := range b.Data.Data {
			data := &b.Data.Data[i]
			fmt.Printf("   -> ParityChanIdCtrl = %x\n", data.ParityChanIdCtrl)
			fmt.Printf("   -> Amplitudes = %x\n", data.Amplitudes)
		}
	}

}

// Frame is a single frame in an ASM stream
type Frame struct {
	ID    uint32 // id of the frame in the ASM stream
	Block Block  // data payload for this frame
}

func (f *Frame) Print(s string) {
	fmt.Printf("Printing frame ID = %v\n", f.ID)
	f.Block.Print(s)
}

func (f *Frame) Buffer() []byte {
	var buffer []uint16
	buffer = append(buffer, f.Block.FirstBlockWord)
	buffer = append(buffer, f.Block.AMCFrameCounters[:]...)
	buffer = append(buffer, f.Block.ParityFEIdCtrl)
	buffer = append(buffer, f.Block.TriggerMode)
	buffer = append(buffer, f.Block.Trigger)
	buffer = append(buffer, f.Block.ASMFrameCounters[:]...)
	buffer = append(buffer, f.Block.Cafe)
	buffer = append(buffer, f.Block.Deca)
	buffer = append(buffer, f.Block.Counters[:]...)
	buffer = append(buffer, f.Block.TimeStamps[:]...)
	buffer = append(buffer, f.Block.NoSamples)
	for i := range f.Block.Data.Data {
		data := &f.Block.Data.Data[i]
		buffer = append(buffer, data.ParityChanIdCtrl)
		buffer = append(buffer, data.Amplitudes...)
		if f.Block.Err == ErrorCode1 {
			fmt.Println("ErrorCode1, add extra word")
			buffer = append(buffer, uint16(0))
		}
	}
	buffer = append(buffer, f.Block.CRC)
	buffer = append(buffer, f.Block.ParityFEIdCtrl2)

	var buffer8 []byte
	for i := range buffer {
		buffer8 = append(buffer8, uint8(buffer[i]>>8))
		buffer8 = append(buffer8, uint8(buffer[i]&0xFFFF))
		//fmt.Printf("buffer8 = %x %x\n", buffer8[len(buffer8)-2], buffer8[len(buffer8)-1])
	}
	return buffer8
}
