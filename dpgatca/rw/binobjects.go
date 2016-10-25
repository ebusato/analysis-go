package rw

import "fmt"

const (
	numCounters uint8  = 12
	firstWord   uint16 = 0x1230
	ctrl1       uint8  = 0xfe
	ctrl2       uint16 = 0xCAFE
	ctrl3       uint16 = 0xDECA
	ctrl4       uint16 = 0x9876
	lastWord    uint8  = 0xfb
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

// Block is a single data frame in an ASM stream
// Each block is associated to one fifo
type Block struct {
	FrameIdBeg     uint16
	FrameIdEnd     uint16
	ParityIdCtrl   uint16
	TriggerMode    uint16
	Trigger        uint16
	Counters       [numCounters]uint16
	TimeStamp1     uint16
	TimeStamp2     uint16
	TimeStamp3     uint16
	TimeStamp4     uint16
	NoSamples      uint16
	ParityChanCtrl uint16
	// Below quantities are to be changed (because repeated 4 times successively)
	Data  []uint16
	SRout uint32
}

func (b *Block) Print(s string) {
	fmt.Printf(" Printing block FrameIdBeg, FrameIdEnd = %v, %v\n", b.FrameIdBeg, b.FrameIdEnd)
	fmt.Printf("   -> ParityIdCtrl = %x\n", b.ParityIdCtrl)
	fmt.Printf("   -> TriggerMode = %x\n", b.TriggerMode)
	fmt.Printf("   -> Trigger = %x\n", b.Trigger)
	fmt.Printf("   -> Counters = %x\n", b.Counters)
	fmt.Printf("   -> TimeStamp1 = %x\n", b.TimeStamp1)
	fmt.Printf("   -> TimeStamp2 = %x\n", b.TimeStamp2)
	fmt.Printf("   -> TimeStamp3 = %x\n", b.TimeStamp3)
	fmt.Printf("   -> TimeStamp4 = %x\n", b.TimeStamp4)
	fmt.Printf("   -> NoSamples = %x\n", b.NoSamples)
	fmt.Printf("   -> ParityChanCtrl = %x\n", b.ParityChanCtrl)
	/*
		switch s {
		case "short":
			// do nothing
		case "medium":
			fmt.Printf("  Data %v = %x\n", 0, b.Data[0])
			fmt.Printf("  Data %v = %x\n", 1, b.Data[1])
			fmt.Println("\t.\n\t.")
			fmt.Printf("  Data %v = %x\n", len(b.Data)-1, b.Data[len(b.Data)-1])
			fmt.Printf("  SRout = %v\n", b.SRout)
			fmt.Printf("  Counter %v = %v\n", 0, b.Counters[0])
			fmt.Println("\t.\n\t.")
			fmt.Printf("  Counter %v = %v\n", len(b.Counters)-1, b.Counters[len(b.Counters)-1])
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
		case "full":
			for i := range b.Data {
				fmt.Printf("  Data %v = %x\n", i, b.Data[i])
			}
			fmt.Printf("  SRout = %v\n", b.SRout)
			for i := range b.Counters {
				fmt.Printf("  Counter %v = %v\n", i, b.Counters[i])
			}
		}
	*/
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
