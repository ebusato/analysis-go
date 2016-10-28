package rw

import "fmt"

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

/*
// Header holds metadata about the run configuration
type Header struct {
	HdrType                 HeaderType // type of header
	History                 uint32     // history word (0 means that it's a raw file produced by the DAQ software, not reprocessed by any subsequent software.)
	RunNumber               uint32     // run number
	FreeField               uint32     // free fields (for the moment not used, here just in case it's needed in the future)
	TimeStart               uint32     // start time (number of seconds since Jan 01 1970)
	TimeStop                uint32     // end time (number of seconds since Jan 01 1970)
	NoEvents                uint32     // total number of events in file
	NoASMCards              uint32     // number of ASM cards
	NoSamples               uint32     // number of samples
	DataToRead              uint32     // data to read
	TriggerEq               uint32     // trigger equation
	TriggerDelay            uint32     // trigger delay
	ChanUsedForTrig         uint32     // channels used for trigger
	Threshold               uint32     // threshold above which fifo data are sent to socket and received on DAQ PC
	LowHighThres            uint32     // low and high trigger thresholds
	TrigSigShapingHighThres uint32     // trigger signal shaping for high threshold
	TrigSigShapingLowThres  uint32     // trigger signal shaping for low threshold
	Size                    uint32     // size of the frame in the ASM stream
	NumFrame                uint32     // number of frames x number of cards
}

// Print prints header fields
func (h *Header) Print() {
	fmt.Printf("Header (type = %#v):\n", h.HdrType.String())
	switch {
	case h.HdrType == HeaderCAL:
		fmt.Printf("   History = %v\n", h.History)
		fmt.Printf("   RunNumber = %v\n", h.RunNumber)
		fmt.Printf("   FreeField = %v\n", h.FreeField)
		fmt.Printf("   TimeStart = %v\n", time.Unix(int64(h.TimeStart), 0).Format(time.UnixDate))
		fmt.Printf("   TimeStop = %v\n", time.Unix(int64(h.TimeStop), 0).Format(time.UnixDate))
		fmt.Printf("   NoEvents = %v\n", h.NoEvents)
		fmt.Printf("   NoASMCards = %v\n", h.NoASMCards)
		fmt.Printf("   NoSamples = %v\n", h.NoSamples)
		fmt.Printf("   DataToRead = %x\n", h.DataToRead)
		fmt.Printf("   TriggerEq = %x\n", h.TriggerEq)
		fmt.Printf("   TriggerDelay = %x\n", h.TriggerDelay)
		fmt.Printf("   ChanUsedForTrig = %x\n", h.ChanUsedForTrig)
		fmt.Printf("   Threshold = %v\n", h.Threshold)
		fmt.Printf("   LowHighThres = %x\n", h.LowHighThres)
		fmt.Printf("   TrigSigShapingHighThres = %x\n", h.TrigSigShapingHighThres)
		fmt.Printf("   TrigSigShapingLowThres = %x\n", h.TrigSigShapingLowThres)
	case h.HdrType == HeaderOld:
		fmt.Printf("   Size = %v\n", h.Size)
		fmt.Printf("   NumFrame = %v\n", h.NumFrame)
	default:
		panic("error ! header type not known")
	}
}
*/

type ChanData struct {
	// Raw quantities
	ParityChanIdCtrl uint16
	Data             []uint16

	// Derived quantities
	Channel uint16
}

type HalfDRSData struct {
	Data [4]ChanData
}

func (h *HalfDRSData) SetNoSamples(n uint16) {
	for i := range h.Data {
		h.Data[i].Data = make([]uint16, n)
	}
}

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
}

func (b *Block) Integrity() error {
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
	for i := range b.Data.Data {
		if (b.Data.Data[i].ParityChanIdCtrl & 0xff) != ctrl0xfd {
			return fmt.Errorf("asm: missing %x magic\n", ctrl0xfd)
		}
	}
	if b.CRC != ctrl0xCRC {
		return fmt.Errorf("asm: missing %x magic\n", ctrl0xCRC)
	}
	if (b.ParityFEIdCtrl2 & 0xff) != ctrl0xfb {
		return fmt.Errorf("asm: missing %x magic\n", ctrl0xfb)
	}
	return nil
}

func (b *Block) Print(s string) {
	fmt.Printf(" Printing block:\n")
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
			fmt.Printf("   -> Data[0] = %x\n", data.Data[0])
			fmt.Printf("   -> Data[1] = %x\n", data.Data[1])
			fmt.Printf("   -> Data[2] = %x\n", data.Data[2])
			fmt.Printf("   -> Data[3] = %x\n", data.Data[3])
			fmt.Printf("   ->    ...\n")
			fmt.Printf("   -> Data[1008] = %x\n", data.Data[1008])
			fmt.Printf("   -> Data[1009] = %x\n", data.Data[1009])
			fmt.Printf("   -> Data[1010] = %x\n", data.Data[1010])
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
			fmt.Printf("   -> Datas = %x\n", data.Data)
		}
	}

}

type TypeOfFrame byte

const (
	FirstFrameOfCluster TypeOfFrame = iota
	SecondFrameOfCluster
)

// Frame is a single frame in an ASM stream
type Frame struct {
	ID          uint32 // id of the frame in the ASM stream
	Block       Block  // data payload for this frame
	typeOfFrame TypeOfFrame
}

func (f *Frame) Print(s string) {
	fmt.Printf("Printing frame ID = %v\n", f.ID)
	f.Block.Print(s)
}
