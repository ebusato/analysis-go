package rw

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"

	"gitlab.in2p3.fr/avirm/analysis-go/detector"
	"gitlab.in2p3.fr/avirm/analysis-go/event"
	"gitlab.in2p3.fr/avirm/analysis-go/pulse"
	"gitlab.in2p3.fr/avirm/analysis-go/testbench/tbdetector"
)

var MissCAFEDECA = errors.New("missing 0xCAFEDECA")
var MissBADCAFEi = errors.New("missing 0xBADCAFEi")
var MissD0D0D0D0 = errors.New("missing 0xd0d0d0d0")

// Reader wraps an io.Reader and reads avirm data files
type Reader struct {
	r                 io.Reader
	err               error
	hdr               Header
	noSamples         uint16
	evtIDPrevFrame    uint32
	firstFrameOfEvent *Frame
	SigThreshold      uint
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

// readHeader reads the header of the binary files
func (r *Reader) readHeader(hdr *Header) {
	switch {
	case r.hdr.HdrType == HeaderGANIL:
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
		r.readU32(&hdr.LowThresholdCluster)
		r.readU32(&hdr.LowHighThres)
		r.readU32(&hdr.TrigSigShapingHighThres)
		r.readU32(&hdr.TrigSigShapingLowThres)
		r.readU32(&hdr.ClusterSigThresShaping)
		r.readU32(&hdr.FirmwareASM)
		r.readU32(&hdr.FirmwareBlond)
		// When setting the number of samples to 1000 it's actually 999
		// hence the -1 subtraction
		//r.noSamples = uint16(hdr.NoSamples) - 1
		r.noSamples = uint16(hdr.NoSamples)
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
	tbdetector.Det.SetNoSamples(int(r.noSamples))
}

func (r *Reader) ReadFrame(f *Frame) {
	r.readFrame(f)
}

func (r *Reader) readFrame(f *Frame) {
	if r.Debug {
		fmt.Printf("rw: start reading frame\n")
	}
	r.readU32(&f.ID)
	if f.ID == lastFrame {
		r.err = io.EOF
		return
	}
	r.readBlock(&f.Block)
	if r.err != nil {
		if r.err == MissCAFEDECA || r.err == MissBADCAFEi {
			return
		}
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
	if ctrl != blockHeader { // && r.err == nil {
		fmt.Println("warning: missing 0xCAFEDECA")
		//r.err = fmt.Errorf("missing 0xCAFEDECA")
		r.err = MissCAFEDECA
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
	// here start
	var ctrl uint32
	r.readU32(&ctrl)
	//if ctrl != ctrld0d0d0d0 { // && r.err == nil {
	//	fmt.Println("warning: missing 0xd0d0d0d0", ctrl)
	//	r.err = MissD0D0D0D0
	//}
	// here end
	r.readU32(&blk.SRout)
	r.read(&blk.Counters)
	//fmt.Printf("rw: srout = %v\n", blk.SRout)
	//for i := range blk.Counters {
	//	r.readU32(&blk.Counters[i])
	//}
}

func (r *Reader) readBlockTrailer(blk *Block) {
	var ctrl uint32
	r.readU32(&ctrl)
	//fmt.Printf("rw: block trailer = %x\n", ctrl)
	if (ctrl >> 4) != blockTrailer { //&& r.err == nil {
		fmt.Println("warning: missing 0xBADCAFEi")
		//r.err = fmt.Errorf("missing 0xBADCAFEi")
		r.err = MissBADCAFEi
	}
}

func MakePulses(f *Frame, iCluster uint8, sigThreshold uint) (*pulse.Pulse, *pulse.Pulse) {
	iChannel_1 := uint8(2 * f.Block.ID)
	iChannel_2 := uint8(iChannel_1 + 1)

	if iChannel_1 >= 24 || iChannel_2 >= 24 {
		log.Fatalf("reader: iChannel_1 >= 24 || iChannel_2 >= 24 (iChannel_1 = %v, iChannel_2 = %v)\n", iChannel_1, iChannel_2)
	}

	//fmt.Printf("iChannel_1=%v iChannel_2=%v\n", iChannel_1, iChannel_2)

	iDRS := uint8(0)
	iQuartet := uint8(0)
	if iChannel_1 >= 4 && iChannel_1 <= 7 {
		iChannel_1 -= 4
		iChannel_2 -= 4
		iQuartet = 1
	} else if iChannel_1 >= 8 && iChannel_1 <= 11 {
		iChannel_1 -= 8
		iChannel_2 -= 8
		iQuartet = 0
		iDRS = 1
	} else if iChannel_1 >= 12 && iChannel_1 <= 15 {
		iChannel_1 -= 12
		iChannel_2 -= 12
		iQuartet = 1
		iDRS = 1
	} else if iChannel_1 >= 16 && iChannel_1 <= 19 {
		iChannel_1 -= 16
		iChannel_2 -= 16
		iQuartet = 0
		iDRS = 2
	} else if iChannel_1 >= 20 && iChannel_1 <= 23 {
		iChannel_1 -= 20
		iChannel_2 -= 20
		iQuartet = 1
		iDRS = 2
	}

	detChannel1 := tbdetector.Det.Channel(iDRS, iQuartet, iChannel_1)
	detChannel2 := tbdetector.Det.Channel(iDRS, iQuartet, iChannel_2)

	/////////////////////////////////////////////////////////////////////
	// To be used when running with SendToSocket with DPGA binary file
	//detChannel1 := tbdetector.Det.Channel(0, 0, 0)
	//detChannel2 := tbdetector.Det.Channel(0, 0, 1)
	//////////////////////////////////////////////////////////////////////

	pulse1 := pulse.NewPulse(detChannel1)
	pulse2 := pulse.NewPulse(detChannel2)

	b := &f.Block
	srout := b.SRout & 0xFFFF
	pulse1.SRout = uint16(srout)
	pulse2.SRout = uint16(srout)

	//fmt.Println(f.Block.ID, iDRS, iQuartet)
	for i := range b.Data {
		word := b.Data[i]

		ampl2 := float64(word & 0xFFF)
		ampl1 := float64(word >> 16)

		sample1 := pulse.NewSample(ampl1, uint16(i), float64(i)*tbdetector.Det.SamplingFreq())
		sample2 := pulse.NewSample(ampl2, uint16(i), float64(i)*tbdetector.Det.SamplingFreq())

		var capa1 *detector.Capacitor
		var capa2 *detector.Capacitor
		if pulse1.SRout <= 1023 {
			capaIndex1 := sample1.CapaIndex(pulse1.SRout)
			capa1 = tbdetector.Det.Capacitor(iDRS, iQuartet, pulse1.Channel.ID(), capaIndex1)
		} else {
			log.Printf("pulse1.SRout = %v (>1023)\n", pulse1.SRout)
		}
		if pulse2.SRout <= 1023 {
			capaIndex2 := sample2.CapaIndex(pulse2.SRout)
			capa2 = tbdetector.Det.Capacitor(iDRS, iQuartet, pulse2.Channel.ID(), capaIndex2)
		} else {
			log.Printf("pulse2.SRout = %v (>1023)\n", pulse2.SRout)
		}
		//fmt.Println(" -> ", i, pulse1.SRout, pulse2.SRout, pulse1.Channel.ID(), pulse2.Channel.ID(), capaIndex1, capaIndex2)
		pulse1.AddSample(sample1, capa1, float64(sigThreshold))
		pulse2.AddSample(sample2, capa2, float64(sigThreshold))
	}

	return pulse1, pulse2
}

func (r *Reader) ReadNextEvent() (*event.Event, bool) {
	event := event.NewEvent(tbdetector.Det.NoClusters())
	for iCluster := uint8(0); iCluster < uint8(event.NoClusters()); iCluster++ {
		frame1, err := r.Frame()
		if err != nil {
			switch {
			case err == io.EOF:
				return nil, false
			case err == MissCAFEDECA || err == MissBADCAFEi:
				event.IsCorrupted = true
			default:
				log.Fatal("error not nil")
			}
		}
		frame1.typeOfFrame = FirstFrameOfCluster
		frame2, err := r.Frame()
		if err != nil {
			switch {
			case err == io.EOF:
				return nil, false
			case err == MissCAFEDECA || err == MissBADCAFEi:
				event.IsCorrupted = true
			default:
				log.Fatal("error not nil")
			}
		}
		frame2.typeOfFrame = SecondFrameOfCluster

		evtID := uint(frame1.Block.Evt)
		if evtID != uint(frame2.Block.Evt) {
			log.Fatal("event IDs of two consecutive frames differ")
		}
		switch iCluster == 0 {
		case true:
			event.ID = evtID
		case false:
			if evtID != event.ID {
				fmt.Printf("Error: \n")
				fmt.Printf("  - evtID = %v, event.ID = %v\n", evtID, event.ID)
				fmt.Printf("  - iCluster = %v\n", iCluster)
				log.Fatal(" => switched to next event")
			}
		}

		pulse0, pulse1 := MakePulses(frame1, iCluster, r.SigThreshold)
		pulse2, pulse3 := MakePulses(frame2, iCluster, r.SigThreshold)

		event.Clusters[iCluster] = *pulse.NewCluster(iCluster, [4]pulse.Pulse{*pulse0, *pulse1, *pulse2, *pulse3})

		// Start counter treatment
		event.Clusters[iCluster].CountersFifo1 = make([]uint32, numCounters)
		event.Clusters[iCluster].CountersFifo2 = make([]uint32, numCounters)
		// here start
		//fmt.Println("Cluster", iCluster)
		for i := uint8(0); i < numCounters; i++ {
			counterf1 := frame1.Block.Counters[i]
			counterf2 := frame2.Block.Counters[i]
			//fmt.Println(i, ": counterf1, counterf2 = ", counterf1, counterf2)
			event.Clusters[iCluster].CountersFifo1[i] = counterf1
			event.Clusters[iCluster].CountersFifo2[i] = counterf2
		}
		// here stop

	}

	return event, true
}
