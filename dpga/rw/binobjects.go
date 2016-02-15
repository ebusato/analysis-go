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
	fmt.Printf(" Size = %v\n", h.Size)
	fmt.Printf(" NumFrame = %v\n", h.NumFrame)
}

// Block is a single data frame in an ASM stream
// Each block is associated to one fifo
type Block struct {
	Evt uint32 // event id
	ID  uint32 // ID of fifo (0 -> 143)

	Data     []uint32
	SRout    uint32
	Counters [17]uint32
}

func (b *Block) Print() {
	fmt.Println("Printing block:")
	fmt.Printf(" Evt = %v\n", b.Evt)
	fmt.Printf(" ID = %v\n", b.ID)
	for i := range b.Data {
		fmt.Printf(" Data %v = %x\n", i, b.Data[i])
	}
	fmt.Printf(" SRout = %v\n", b.SRout)
	for i := range b.Counters {
		fmt.Printf(" Counter %v = %v\n", i, b.Counters[i])
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

func (f *Frame) Print() {
	fmt.Println("Printing frame:")
	fmt.Printf(" ID = %v\n", f.ID)
	f.Block.Print()
}
