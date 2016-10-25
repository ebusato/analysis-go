package rw

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
)

// Reader wraps an io.Reader and reads avirm data files
type Reader struct {
	r   io.Reader
	err error
	//hdr               Header
	noSamples uint16
	//evtIDPrevFrame    uint32
	//firstFrameOfEvent *Frame
	//SigThreshold      uint
	Debug bool
}

// NoSamples returns the number of samples
func (r *Reader) NoSamples() uint16 {
	return r.noSamples
}

// Err return the reader error
func (r *Reader) Err() error {
	return r.err
}

/*
// Header returns the ASM-stream header
func (r *Reader) Header() *Header {
	return &r.hdr
}
*/

// NewReader returns a new ASM stream in read mode
func NewReader(r io.Reader) (*Reader, error) {
	rr := &Reader{
		r: r,
		//evtIDPrevFrame: 0,
		//SigThreshold:   800,
	}
	//rr.readHeader(&rr.hdr)
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
			Data: make([]uint16, r.noSamples),
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

func (r *Reader) readU32(v *uint32) {
	if r.err != nil {
		return
	}
	var buf [4]byte
	_, r.err = r.r.Read(buf[:])
	if r.err != nil {
		return
	}
	*v = binary.BigEndian.Uint32(buf[:])
	if r.Debug {
		fmt.Printf("word = %x\n", *v)
	}
}

/*
func (r *Reader) readHeader(hdr *Header) {
	switch {
	case r.hdr.HdrType == HeaderCAL:
		r.readU32(&hdr.History)
		r.readU32(&hdr.RunNumber)
		r.readU32(&hdr.FreeField)
		r.readU32(&hdr.TimeStart)
		r.readU32(&hdr.TimeStop)
		r.readU32(&hdr.NoEvents)
		r.readU32(&hdr.NoASMCards)
		r.readU32(&hdr.NoSamples)
		r.readU32(&hdr.DataToRead)
		r.readU32(&hdr.TriggerEq)
		r.readU32(&hdr.TriggerDelay)
		r.readU32(&hdr.ChanUsedForTrig)
		r.readU32(&hdr.Threshold)
		r.readU32(&hdr.LowHighThres)
		r.readU32(&hdr.TrigSigShapingHighThres)
		r.readU32(&hdr.TrigSigShapingLowThres)
		// When setting the number of samples to 1000 it's actually 999
		// hence the -1 subtraction
		r.noSamples = uint16(hdr.NoSamples) - 1
	case r.hdr.HdrType == HeaderOld:
		r.readU32(&hdr.Size)
		r.readU32(&hdr.NumFrame)
		// In the case of old header, the number of samples
		// is retrieved from the header.Size field
		// When header.Size = 1007, the number of samples is 999
		// hence the -8 subtraction
		r.noSamples = uint16(hdr.Size) - 8
		//fmt.Printf("rw: reading header %v %v\n", hdr.Size, hdr.NumFrame)
	default:
		panic("error ! header type not known")
	}
	dpgadetector.Det.SetNoSamples(int(r.noSamples))
}
*/

func (r *Reader) ReadFrame(f *Frame) {
	r.readFrame(f)
}

func (r *Reader) readFrame(f *Frame) {
	if r.Debug {
		fmt.Printf("rw: start reading frame\n")
	}
	//fmt.Printf("rw: frame id = %v\n", f.ID)
	/*
		if f.ID == lastFrame {
			r.err = io.EOF
			return
		}
	*/
	r.readBlock(&f.Block)
	if r.err != nil {
		if r.err != io.EOF {
			log.Fatalf("error loading frame: %v\n", r.err)
		}
		/*
			if f.ID != lastFrame {
				log.Fatalf("invalid last frame id. got=%x. want=%x", f.ID, lastFrame)
			}
		*/

	}
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
	var ctrl uint16
	r.readU16(&ctrl)
	if ctrl != firstWord && r.err == nil {
		r.err = fmt.Errorf("asm: missing %x magic\n", firstWord)
	}
	r.readU16(&blk.FrameIdBeg)
	r.readU16(&blk.FrameIdEnd)
	r.readU16(&blk.ParityIdCtrl)
	r.readU16(&blk.TriggerMode)
	r.readU16(&blk.Trigger)
	r.read(&blk.Counters)
	r.readU16(&blk.TimeStamp1)
	r.readU16(&blk.TimeStamp2)
	r.readU16(&blk.TimeStamp3)
	r.readU16(&blk.TimeStamp4)
	r.readU16(&blk.NoSamples)
	r.readU16(&blk.ParityChanCtrl)
}

func (r *Reader) readBlockData(blk *Block) {
	if r.err != nil {
		return
	}
	r.read(&blk.Data)
	//for i := range blk.Data {
	//	r.readU32(&blk.Data[i])
	//}
	r.readU32(&blk.SRout)
	r.read(&blk.Counters)
	//fmt.Printf("rw: srout = %v\n", blk.SRout)
	//for i := range blk.Counters {
	//	r.readU32(&blk.Counters[i])
	//}
}

func (r *Reader) readBlockTrailer(blk *Block) {
	var ctrl uint16
	r.readU16(&ctrl)
	//fmt.Printf("rw: block trailer = %x\n", ctrl)
	if ctrl != ctrl4 && r.err == nil {
		r.err = fmt.Errorf("asm: missing %x magic\n", ctrl4)
	}
}

/*
func MakePulses(f *Frame, iCluster uint8, sigThreshold uint) (*pulse.Pulse, *pulse.Pulse) {
	iChannelAbs288_1 := uint16(2 * f.Block.ID)
	iChannelAbs288_2 := uint16(iChannelAbs288_1 + 1)

	if iChannelAbs288_1 >= 288 || iChannelAbs288_2 >= 288 {
		panic("reader: iChannelAbs288_1 >= 288 || iChannelAbs288_2 >= 288")
	}

	detChannel1 := dpgadetector.Det.ChannelFromIdAbs288(iChannelAbs288_1)
	detChannel2 := dpgadetector.Det.ChannelFromIdAbs288(iChannelAbs288_2)

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

	///////////////////////////////////////////////////////////////////////////////////
	// Sanity check
	iHemi, iASM, iDRS, iQuartet := dpgadetector.QuartetAbsIdx60ToRelIdx(iCluster)
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

		pulse1.AddSample(sample1, dpgadetector.Det.Capacitor(iHemi, iASM, iDRS, iQuartet, iChannel1, sample1.CapaIndex(pulse1.SRout)), float64(sigThreshold))
		pulse2.AddSample(sample2, dpgadetector.Det.Capacitor(iHemi, iASM, iDRS, iQuartet, iChannel2, sample2.CapaIndex(pulse2.SRout)), float64(sigThreshold))
	}

	return pulse1, pulse2
}
*/
