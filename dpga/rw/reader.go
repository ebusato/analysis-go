package rw

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"

	"gitlab.in2p3.fr/avirm/analysis-go/dpga/dpgadetector"
	"gitlab.in2p3.fr/avirm/analysis-go/event"
	"gitlab.in2p3.fr/avirm/analysis-go/pulse"
)

// Reader wraps an io.Reader and reads avirm data files
type Reader struct {
	r     io.Reader
	err   error
	hdr   Header
	Debug bool
}

// Err return the reader error
func (r *Reader) Err() error {
	return r.err
}

// Header returns the ASM-stream header
func (r *Reader) Header() Header {
	return r.hdr
}

// NewReader returns a new ASM stream in read mode
func NewReader(r io.Reader) (*Reader, error) {
	rr := &Reader{
		r: r,
	}
	rr.readHeader(&rr.hdr)
	return rr, rr.err
}

// Read implements io.Reader
func (r *Reader) Read(data []byte) (int, error) {
	return r.r.Read(data)
}

// Frame reads a single frame from the underlying io.Reader.
//
// Frame returns io.EOF when there are no more frame to read.
func (r *Reader) Frame() (*Frame, error) {
	if r.err != nil {
		return &Frame{}, r.err
	}
	f := &Frame{
		Block: Block{
			Data: make([]uint32, numSamples),
		},
	}
	r.readFrame(f)
	return f, r.err
}

func (r *Reader) read(v interface{}) {
	if r.err != nil {
		return
	}
	r.err = binary.Read(r.r, binary.BigEndian, v)
	if r.Debug {
		fmt.Printf("word = %x\n", *(v.(*uint32)))
	}
}

func (r *Reader) ReadFrame(f *Frame) {
	r.readFrame(f)
}

func (r *Reader) readFrame(f *Frame) {
	if r.Debug {
		fmt.Printf("rw: start reading frame\n")
	}
	r.read(&f.ID)
	//fmt.Printf("rw: frame id = %v\n", f.ID)
	if f.ID == lastFrame {
		r.err = io.EOF
		return
	}
	r.readBlock(&f.Block)
	if r.err != nil {
		if r.err != io.EOF {
			log.Fatalf("error loading frame: %v\n", r.err)
		}
		if f.ID != lastFrame {
			log.Fatalf("invalid last frame id. got=%x. want=%x", f.ID, lastFrame)
		}
	}
}

func (r *Reader) readHeader(hdr *Header) {
	r.read(&hdr.Size)
	r.read(&hdr.NumFrame)
	//fmt.Printf("rw: reading header %v %v\n", hdr.Size, hdr.NumFrame)
}

func (r *Reader) readBlock(blk *Block) {
	r.readBlockHeader(blk)
	r.readBlockData(blk)
	r.readBlockTrailer(blk)
}

func (r *Reader) readBlockHeader(blk *Block) {
	if r.err != nil {
		return
	}
	r.read(&blk.Evt)
	r.read(&blk.ID)
	var ctrl uint32
	r.read(&ctrl)
	if ctrl != blockHeader && r.err == nil {
		r.err = fmt.Errorf("asm: missing 0xCAFEDECA magic")
	}
	//fmt.Printf("rw: reading block header %v %v %x\n", blk.Evt, blk.ID, ctrl)
}

func (r *Reader) readBlockData(blk *Block) {
	if r.err != nil {
		return
	}
	//for i := range blk.Data {
	//	r.read(&blk.Data[i])
	//}
	r.read(&blk.Data)
	r.read(&blk.SRout)
	//fmt.Printf("rw: srout = %v\n", blk.SRout)
	for i := range blk.Counters {
		r.read(&blk.Counters[i])
	}
}

func (r *Reader) readBlockTrailer(blk *Block) {
	var ctrl uint32
	r.read(&ctrl)
	//fmt.Printf("rw: block trailer = %x\n", ctrl)
	if (ctrl>>4) != blockTrailer && r.err == nil {
		r.err = fmt.Errorf("asm: missing 0xBADCAFEF magic")
	}
}

func MakePulses(f *Frame, iCluster uint8) (*pulse.Pulse, *pulse.Pulse) {
	iChannelAbs288_1 := uint16(2 * f.Block.ID)
	iChannelAbs288_2 := uint16(iChannelAbs288_1 + 1)

	if iChannelAbs288_1 >= 288 || iChannelAbs288_2 >= 288 {
		panic("reader: iChannelAbs288_1 >= 288 || iChannelAbs288_2 >= 288")
	}

	detChannel1 := dpgadetector.Det.ChannelFromIdAbs288(iChannelAbs288_1)
	detChannel2 := dpgadetector.Det.ChannelFromIdAbs288(iChannelAbs288_2)

	///////////////////////////////////////////////////////////////////////////////////
	// Sanity check
	iHemi, iASM, iDRS, iQuartet := dpgadetector.QuartetAbsIdx60ToRelIdx(iCluster)
	var iChannel1 uint8
	var iChannel2 uint8
	switch f.typeOfFrame {
	case FirstFrameOfCluster:
		iChannel1 = 0
		iChannel2 = 1
	case SecondFrameOfCluster:
		iChannel1 = 2
		iChannel2 = 3
	}

	detChannel1debug := dpgadetector.Det.Channel(iHemi, iASM, iDRS, iQuartet, iChannel1)
	detChannel2debug := dpgadetector.Det.Channel(iHemi, iASM, iDRS, iQuartet, iChannel2)

	//fmt.Printf(" %p %p\n", detChannel1debug, detChannel1)
	//fmt.Printf(" %p %p\n", detChannel2debug, detChannel2)

	if detChannel1debug != detChannel1 {
		panic("reader: detChannel1debug != detChannel1")
	}
	if detChannel2debug != detChannel2 {
		panic("reader: detChannel2debug != detChannel2")
	}

	absid1 := detChannel1debug.AbsID288()
	absid2 := detChannel2debug.AbsID288()

	if iChannelAbs288_1 != absid1 {
		panic("reader: iChannelAbs1 != absid1")
	}
	if iChannelAbs288_2 != absid2 {
		panic("reader: iChannelAbs2 != absid2")
	}
	// Enf of sanity check
	////////////////////////////////////////////////////////////////////////////////////

	pulse1 := pulse.NewPulse(detChannel1)
	pulse2 := pulse.NewPulse(detChannel2)

	b := &f.Block
	pulse1.SRout = uint16(b.SRout)
	pulse2.SRout = uint16(b.SRout)

	for i := range b.Data {
		word := b.Data[i]

		ampl2 := float64(word & 0xFFF)
		ampl1 := float64(word >> 16)

		sample1 := pulse.NewSample(ampl1, uint16(i), float64(i)*dpgadetector.Det.SamplingFreq())
		sample2 := pulse.NewSample(ampl2, uint16(i), float64(i)*dpgadetector.Det.SamplingFreq())

		pulse1.AddSample(sample1, dpgadetector.Det.Capacitor(iHemi, iASM, iDRS, iQuartet, iChannel1, sample1.CapaIndex(pulse1.SRout)))
		pulse2.AddSample(sample2, dpgadetector.Det.Capacitor(iHemi, iASM, iDRS, iQuartet, iChannel2, sample2.CapaIndex(pulse2.SRout)))
	}

	return pulse1, pulse2
}

func (r *Reader) ReadNextEvent() (*event.Event, bool) {
	event := event.NewEvent(dpgadetector.Det.NoClusters())
	for iCluster := uint8(0); iCluster < uint8(event.NoClusters()); iCluster++ {
		frame1, err := r.Frame()
		if err != nil {
			if err == io.EOF {
				return nil, false
			}
			log.Fatal("error not nil", err)
		}
		frame1.typeOfFrame = FirstFrameOfCluster
		frame2, err := r.Frame()
		if err != nil {
			log.Fatal("error not nil")
		}
		frame2.typeOfFrame = SecondFrameOfCluster

		// 		frame1.Print()
		// 		frame2.Print()

		evtID := uint(frame1.Block.Evt)
		if evtID != uint(frame2.Block.Evt) {
			log.Fatal("event IDs of two consecutive frames differ")
		}
		switch iCluster == 0 {
		case true:
			event.ID = evtID
		case false:
			if evtID != event.ID {
				log.Fatal("error: switched to next event")
			}
		}

		pulse0, pulse1 := MakePulses(frame1, iCluster)
		pulse2, pulse3 := MakePulses(frame2, iCluster)

		event.Clusters[iCluster] = *pulse.NewCluster(iCluster, [4]pulse.Pulse{*pulse0, *pulse1, *pulse2, *pulse3})
		event.Clusters[iCluster].Counters = make([]uint32, numCounters)
		for i := uint8(0); i < numCounters; i++ {
			counterf1 := frame1.Block.Counters[i]
			counterf2 := frame2.Block.Counters[i]
			if counterf1 != counterf2 {
				log.Fatalf("rw: countersf1 != countersf2")
			}
			event.Clusters[iCluster].Counters[i] = counterf1
		}
	}

	return event, true
}
