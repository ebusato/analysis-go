package rw

import "fmt"

const (
	numSamples   uint16 = 999
	numCounters  uint8  = 17
	blockHeader  uint32 = 0xCAFEDECA
	blockTrailer uint32 = 0xBADCAFE
	lastFrame    uint32 = 0xFFFFFFFF
)

func LastFrame() uint32 {
	return lastFrame
}

// Header holds metadata about the frames in the ASM stream
type Header struct {
	Size     uint32 // size of the frame in the ASM stream
	NumFrame uint32 // number of frames x number of cards
}

func (h *Header) Print() {
	fmt.Println("Printing header:")
	fmt.Printf("   Size = %v\n", h.Size)
	fmt.Printf("   NumFrame = %v\n", h.NumFrame)
}

// Block is a single data frame in an ASM stream
// Each block is associated to one fifo
type Block struct {
	Evt uint32 // event id
	ID  uint32 // ID of fifo (0 -> 143)

	Data     []uint32
	SRout    uint32
	Counters [numCounters]uint32
}

func (b *Block) Print(s string) {
	fmt.Printf(" Printing block Evt = %v, ID = %v\n", b.Evt, b.ID)

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
