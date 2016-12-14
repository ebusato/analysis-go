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
	r                 io.Reader
	err               error
	hdr               Header
	noSamples         uint16
	evtIDPrevFrame    uint32
	firstFrameOfEvent *Frame
	SigThreshold      uint
	Counters          [NumCounters]uint32
	Debug             bool
}

// NoSamples returns the number of samples
func (r *Reader) NoSamples() uint16 {
	return r.noSamples
}

// Err return the reader error
func (r *Reader) Err() error {
	return r.err
}

// Header returns the ASM-stream header
func (r *Reader) Header() *Header {
	return &r.hdr
}

// NewReader returns a new ASM stream in read mode
func NewReader(r io.Reader, ht HeaderType) (*Reader, error) {
	rr := &Reader{
		r:              r,
		evtIDPrevFrame: 0,
		SigThreshold:   800,
	}
	rr.hdr.HdrType = ht
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
	//fmt.Println("making block with", r.noSamples, " samples")
	f := &Frame{
		Block: Block{
			Data: make([]uint32, r.noSamples),
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
		case *[NumCounters]uint32:
			for _, vv := range *v {
				fmt.Printf("word = %x\n", vv)
			}
		}
		//fmt.Printf("word = %x\n", *(v.(*uint32)))
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

func (r *Reader) ReadFrame(f *Frame) {
	r.readFrame(f)
}

func (r *Reader) readFrame(f *Frame) {
	if r.Debug {
		fmt.Printf("rw: start reading frame\n")
	}
	r.readU32(&f.ID)
	if f.ID == FirstEventWord {
		// First frame of event, read counters first
		f.FirstOfEvent = true
		r.read(&r.Counters)
		r.readU32(&f.ID)
	}
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

func (r *Reader) readBlock(blk *Block) {
	r.readBlockHeader(blk)
	r.readBlockData(blk)
	r.readBlockTrailer(blk)
}

func (r *Reader) readBlockHeader(blk *Block) {
	if r.err != nil {
		return
	}
	r.readU32(&blk.Evt)
	r.readU32(&blk.ID)
	var ctrl uint32
	r.readU32(&ctrl)
	//fmt.Printf("rw: reading block header %v %v %x\n", blk.Evt, blk.ID, ctrl)
	if ctrl != blockHeader && r.err == nil {
		r.err = fmt.Errorf("asm: missing 0xCAFEDECA magic")
	}
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
	//r.read(&blk.Counters)
	//for i := range blk.Counters {
	//	r.readU32(&blk.Counters[i])
	//}
}

func (r *Reader) readBlockTrailer(blk *Block) {
	var ctrl uint32
	// This is an extra word that we have to read coming from socket
	r.readU32(&ctrl)
	r.readU32(&ctrl)
	//fmt.Printf("rw: ctrl = %x\n", ctrl)
	//fmt.Printf("rw: block trailer = %x\n", ctrl)
	if (ctrl>>4) != blockTrailer && r.err == nil {
		r.err = fmt.Errorf("asm: missing 0xBADCAFEF magic (found %v)\n", ctrl)
	}
}

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

func (r *Reader) ReadNextEvent() (*event.Event, bool) {
	event := event.NewEvent(dpgadetector.Det.NoClusters())
	event.Counters = make([]uint32, NumCounters)
	for i := range event.Counters {
		event.Counters[i] = r.Counters[i]
	}
	firstPass := true
	for { // loop over frames
		var frame *Frame = nil
		if r.firstFrameOfEvent != nil { // enter this only for first frame of event
			frame = r.firstFrameOfEvent
			if r.err != nil {
				log.Println("error not nil", r.err)
				if r.err == io.EOF {
					return nil, false
				}
			}
			r.firstFrameOfEvent = nil
		} else { // enter this for all frames but the first one of the event
			frametemp, err := r.Frame()
			if err != nil && err != io.EOF {
				log.Fatal("error not nil", err)
			}
			frame = frametemp
		}
		evtID := frame.Block.Evt
		//fmt.Println("evtID =", evtID)
		if firstPass || evtID == r.evtIDPrevFrame { // fill event
			if firstPass {
				event.ID = uint(evtID)
			}
			firstPass = false
			fifoID144 := uint16(frame.Block.ID)
			iCluster := dpgadetector.FifoID144ToQuartetAbsIdx60(fifoID144, true)
			if iCluster >= 60 {
				log.Fatalf("error ! iCluster=%v (>= 60)\n", iCluster)
			}
			//fmt.Printf("fifoID144=%v, iCluster = %v\n", fifoID144, iCluster)
			switch fifoID144 % 2 {
			case 0:
				frame.typeOfFrame = FirstFrameOfCluster
			case 1:
				frame.typeOfFrame = SecondFrameOfCluster
			}
			pulse0, pulse1 := MakePulses(frame, iCluster, r.SigThreshold)
			event.Clusters[iCluster].ID = iCluster
			switch frame.typeOfFrame {
			case FirstFrameOfCluster:
				event.Clusters[iCluster].Pulses[0] = *pulse0
				event.Clusters[iCluster].Pulses[1] = *pulse1
			case SecondFrameOfCluster:
				event.Clusters[iCluster].Pulses[2] = *pulse0
				event.Clusters[iCluster].Pulses[3] = *pulse1
			}
		} else { // switched to next event
			r.firstFrameOfEvent = frame
			return event, true
		}
		r.evtIDPrevFrame = evtID
	} // end of loop over frames
	log.Fatalf("error ! you should never end up here")
	return nil, false
}

func (r *Reader) ReadNextEventFull() (*event.Event, bool) {
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

		pulse0, pulse1 := MakePulses(frame1, iCluster, r.SigThreshold)
		pulse2, pulse3 := MakePulses(frame2, iCluster, r.SigThreshold)

		event.Clusters[iCluster] = *pulse.NewCluster(iCluster, [4]pulse.Pulse{*pulse0, *pulse1, *pulse2, *pulse3})
	}

	return event, true
}
