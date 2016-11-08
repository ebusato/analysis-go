package rw

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"sync"

	"gitlab.in2p3.fr/avirm/analysis-go/dpga/dpgadetector"
	"gitlab.in2p3.fr/avirm/analysis-go/event"
	"gitlab.in2p3.fr/avirm/analysis-go/pulse"
)

type FrameType byte

const (
	UDPorTCP16bits FrameType = iota
	UDPHalfDRS
)

// Reader wraps an io.Reader and reads avirm data files
type Reader struct {
	r   io.Reader
	err error
	//hdr               Header
	noSamples uint16
	eventMap  map[uint64]*event.Event
	//evtIDPrevFrame    uint32
	//firstFrameOfEvent *Frame
	SigThreshold uint
	Debug        bool
	FrameT       FrameType
	FrameBuffer  []byte // relevant only when reading from UDP with packet = half DRS
}

// NoSamples returns the number of samples
func (r *Reader) NoSamples() uint16 {
	return r.noSamples
}

// Err return the reader error
func (r *Reader) Err() error {
	return r.err
}

// EventMapKeys returns a slice of keys stored in the reader's event map
func (r *Reader) EventMapKeys() []uint64 {
	var keys []uint64
	for k := range r.eventMap {
		keys = append(keys, k)
	}
	return keys
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
		r:        r,
		eventMap: make(map[uint64]*event.Event),
		//evtIDPrevFrame: 0,
		SigThreshold: 800,
		FrameT:       UDPHalfDRS,
		FrameBuffer:  make([]byte, 8230),
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
	/*
		Block: Block{
			Data: make([]uint16, r.noSamples),
		},
	*/

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

func (r *Reader) ReadU16(v *uint16) {
	r.readU16(v)
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
	switch r.FrameT {
	case UDPHalfDRS:
		r.r.Read(r.FrameBuffer)
	case UDPorTCP16bits:
		// do nothing
	}
	r.readBlockHeader(blk)
	r.readBlockData(blk)
	r.readBlockTrailer(blk)
	//blk.Print("medium")
	r.err = blk.Integrity()
	if r.err != nil {
		fmt.Println("Integrity check failed")
		blk.Print("medium")
		return
	}
}

func (r *Reader) readBlockHeader(blk *Block) {
	switch r.FrameT {
	case UDPHalfDRS:
		/*
			fmt.Printf("frameBuffer =")
			for j := range frameBuffer {
				fmt.Printf("  %v: %x\n", j, frameBuffer[j])
			}
		*/
		blk.FirstBlockWord = binary.BigEndian.Uint16(r.FrameBuffer[0:2])
		blk.AMCFrameCounters[0] = binary.BigEndian.Uint16(r.FrameBuffer[2:4])
		blk.AMCFrameCounters[1] = binary.BigEndian.Uint16(r.FrameBuffer[4:6])
		blk.ParityFEIdCtrl = binary.BigEndian.Uint16(r.FrameBuffer[6:8])
		blk.TriggerMode = binary.BigEndian.Uint16(r.FrameBuffer[8:10])
		blk.Trigger = binary.BigEndian.Uint16(r.FrameBuffer[10:12])
		blk.ASMFrameCounters[0] = binary.BigEndian.Uint16(r.FrameBuffer[12:14])
		blk.ASMFrameCounters[1] = binary.BigEndian.Uint16(r.FrameBuffer[14:16])
		blk.ASMFrameCounters[2] = binary.BigEndian.Uint16(r.FrameBuffer[16:18])
		blk.ASMFrameCounters[3] = binary.BigEndian.Uint16(r.FrameBuffer[18:20])
		blk.Cafe = binary.BigEndian.Uint16(r.FrameBuffer[20:22])
		blk.Deca = binary.BigEndian.Uint16(r.FrameBuffer[22:24])
		blk.Counters[0] = binary.BigEndian.Uint16(r.FrameBuffer[24:26])
		blk.Counters[1] = binary.BigEndian.Uint16(r.FrameBuffer[26:28])
		blk.Counters[2] = binary.BigEndian.Uint16(r.FrameBuffer[28:30])
		blk.Counters[3] = binary.BigEndian.Uint16(r.FrameBuffer[30:32])
		blk.TimeStamps[0] = binary.BigEndian.Uint16(r.FrameBuffer[32:34])
		blk.TimeStamps[1] = binary.BigEndian.Uint16(r.FrameBuffer[34:36])
		blk.TimeStamps[2] = binary.BigEndian.Uint16(r.FrameBuffer[36:38])
		blk.TimeStamps[3] = binary.BigEndian.Uint16(r.FrameBuffer[38:40])
		blk.NoSamples = binary.BigEndian.Uint16(r.FrameBuffer[40:42])
	case UDPorTCP16bits:
		r.readU16(&blk.FirstBlockWord)
		r.read(&blk.AMCFrameCounters)
		r.readU16(&blk.ParityFEIdCtrl)
		r.readU16(&blk.TriggerMode)
		r.readU16(&blk.Trigger)
		r.read(&blk.ASMFrameCounters)
		r.readU16(&blk.Cafe)
		r.readU16(&blk.Deca)
		r.read(&blk.Counters)
		r.read(&blk.TimeStamps)
		r.readU16(&blk.NoSamples)
	}
	blk.AMCFrameCounter = (uint32(blk.AMCFrameCounters[0]) << 16) + uint32(blk.AMCFrameCounters[1])
	blk.FrontEndId = (blk.ParityFEIdCtrl & 0x7fff) >> 8
	blk.ASMFrameCounter = (uint64(blk.ASMFrameCounters[0]) << 48) + (uint64(blk.ASMFrameCounters[1]) << 32) + (uint64(blk.ASMFrameCounters[2]) << 16) + uint64(blk.ASMFrameCounters[3])
	temp := (uint64(blk.TimeStamps[0]) << 16) | uint64(blk.TimeStamps[1])
	temp = (temp << 32)
	temp1 := (uint64(blk.TimeStamps[2]) << 16) | uint64(blk.TimeStamps[3])
	// 	temp |= temp1
	blk.TimeStamp = temp | temp1
	///////////////////////////////////////////////////////////////////////
	// This +11 is necessary but currently not really understood
	// 11 clock periods are generated by "machine d'etat" in ASM firmware
	// These additionnal 11 samples should currently be considered junk
	blk.Data.SetNoSamples(blk.NoSamples + 11)
	///////////////////////////////////////////////////////////////////////
}

var (
	noAttempts         int
	QuartetAbsIdx60old uint8
)

// readParityChanIdCtrl is a temporary fix, until we understand where the additionnal 16 bits words come from
func (r *Reader) readParityChanIdCtrl(blk *Block, i int) bool {
	data := &blk.Data.Data[i]
	switch r.FrameT {
	case UDPHalfDRS:
		data.ParityChanIdCtrl = binary.BigEndian.Uint16(r.FrameBuffer[42+i*2*1023 : 44+i*2*1023])
	case UDPorTCP16bits:
		r.readU16(&data.ParityChanIdCtrl)
	}

	data.Channel = (data.ParityChanIdCtrl & 0x7f00) >> 8
	blk.QuartetAbsIdx60 = dpgadetector.FEIdAndChanIdToQuartetAbsIdx60(blk.FrontEndId, data.Channel)

	//fmt.Printf("%v, %v, %x, %v, %v, %v\n", noAttempts, i, data.ParityChanIdCtrl, data.Channel, blk.QuartetAbsIdx60, QuartetAbsIdx60old)
	if (data.ParityChanIdCtrl & 0xff) != ctrl0xfd {
		//panic("(data.ParityChanIdCtrl & 0xff) != ctrl0xfd")
		return true
	}
	if i > 0 && blk.QuartetAbsIdx60 != QuartetAbsIdx60old {
		//panic("i > 0 && blk.QuartetAbsIdx60 != QuartetAbsIdx60old")
		return true
	}
	if data.Channel != blk.Data.Data[0].Channel+uint16(i) {
		//panic("reader.readParityChanIdCtrl: data.Channel != blk.Data.Data[0].Channel+uint16(i)")
		return true
	}
	QuartetAbsIdx60old = blk.QuartetAbsIdx60
	return false
}

func (r *Reader) readBlockData(blk *Block) {
	if r.err != nil {
		return
	}
	//blk.Print("short")
	for i := range blk.Data.Data {
		data := &blk.Data.Data[i]
		for r.readParityChanIdCtrl(blk, i) {
			noAttempts++
			if noAttempts >= 2 {
				log.Fatalf("reader.readParityChanIdCtrl: noAttempts >= 2\n")
			}
		}
		noAttempts = 0
		//fmt.Printf("data.ParityChanIdCtrl = %x\n", data.ParityChanIdCtrl)
		switch r.FrameT {
		case UDPorTCP16bits:
			r.read(&data.Amplitudes)
		case UDPHalfDRS:
			for j := range data.Amplitudes {
				data.Amplitudes[j] = binary.BigEndian.Uint16(r.FrameBuffer[44+2*j+i*2*1023 : 46+2*j+i*2*1023])
			}
		}
		// 		for j := range data.Amplitudes {
		// 			fmt.Printf("data.Amplitudes[%v] = %x\n", j, data.Amplitudes[j])
		// 		}
	}
}

func (r *Reader) readBlockTrailer(blk *Block) {
	switch r.FrameT {
	case UDPorTCP16bits:
		r.readU16(&blk.CRC)
		// Temporary fix, until we understand where these additionnal 16 bits come from
		if blk.CRC != ctrl0xCRC {
			//fmt.Printf("CRC = %x (should be %x)\n", blk.CRC, ctrl0xCRC)
			r.readU16(&blk.CRC)
			//fmt.Printf("new CRC = %x\n", blk.CRC)
		}
		// End of temporary fix
		r.readU16(&blk.ParityFEIdCtrl2)
	case UDPHalfDRS:
		blk.CRC = binary.BigEndian.Uint16(r.FrameBuffer[len(r.FrameBuffer)-4 : len(r.FrameBuffer)-2])
		blk.ParityFEIdCtrl2 = binary.BigEndian.Uint16(r.FrameBuffer[len(r.FrameBuffer)-2 : len(r.FrameBuffer)])
	}
}

func MakePulses(f *Frame, sigThreshold uint) [4]*pulse.Pulse {
	var pulses [len(f.Block.Data.Data)]*pulse.Pulse
	for i := range f.Block.Data.Data {
		chanData := &f.Block.Data.Data[i]
		channelId023 := chanData.Channel
		iChannel := uint8(channelId023 % 4)
		iHemi, iASM, iDRS, iQuartet := dpgadetector.QuartetAbsIdx60ToRelIdx(f.Block.QuartetAbsIdx60)
		detChannel := dpgadetector.Det.Channel(iHemi, iASM, iDRS, iQuartet, iChannel)
		pul := pulse.NewPulse(detChannel)
		for j := range chanData.Amplitudes {
			ampl := float64(chanData.Amplitudes[j])
			sample := pulse.NewSample(ampl, uint16(j), float64(j)*dpgadetector.Det.SamplingFreq())
			pul.AddSample(sample, dpgadetector.Det.Capacitor(iHemi, iASM, iDRS, iQuartet, iChannel, 0), float64(sigThreshold))
		}
		pulses[i] = pul
	}
	return pulses
}

func EventsNotUpdatedForLongTime(timestamps []uint64, eventmapkeys []uint64) []uint64 {
	var eventsToFlush []uint64
	for _, evttimestamp := range eventmapkeys {
		noFramesSinceLastUpdate := 0
		j := len(timestamps) - 1
		for timestamps[j] != evttimestamp {
			noFramesSinceLastUpdate++
			j--
		}
		if noFramesSinceLastUpdate > 20 {
			eventsToFlush = append(eventsToFlush, evttimestamp)
		}
	}
	return eventsToFlush
}

func EventAlreadyFlushed(timestamp uint64, flushedEvents []uint64) bool {
	for _, ts := range flushedEvents {
		if timestamp == ts {
			return true
		}
	}
	return false
}

func (r *Reader) ReadFrames(evtChan chan *event.Event, w *Writer, wg *sync.WaitGroup) {
	defer wg.Done()
	nframes := 0
	var timestamps []uint64
	var allFlushedEvents []uint64
	for {
		//fmt.Printf("reading frame %v\n", nframes)
		frame, _ := r.Frame()
		//frame.Print("medium")
		//w.Frame(frame)
		//frame.Print("medium")
		if EventAlreadyFlushed(frame.Block.TimeStamp, allFlushedEvents) {
			log.Fatalf("Event with timestamp=%v already flushed\n", frame.Block.TimeStamp)
		}
		timestamps = append(timestamps, frame.Block.TimeStamp)
		nframes++
		_, ok := r.eventMap[frame.Block.TimeStamp]
		switch ok {
		case false:
			r.eventMap[frame.Block.TimeStamp] = event.NewEvent(dpgadetector.Det.NoClusters())
		default:
			// event already present in map
		}
		evt := r.eventMap[frame.Block.TimeStamp]
		evt.TimeStamp = frame.Block.TimeStamp
		pulses := MakePulses(frame, r.SigThreshold)
		evt.Clusters[frame.Block.QuartetAbsIdx60].Pulses[0] = *pulses[0]
		evt.Clusters[frame.Block.QuartetAbsIdx60].Pulses[1] = *pulses[1]
		evt.Clusters[frame.Block.QuartetAbsIdx60].Pulses[2] = *pulses[2]
		evt.Clusters[frame.Block.QuartetAbsIdx60].Pulses[3] = *pulses[3]
		//fmt.Println("\nEvent map keys: ", r.EventMapKeys())
		//fmt.Println("\nTimeStamps: ", timestamps)

		// Determine which events to flush
		eventsToFlush := EventsNotUpdatedForLongTime(timestamps, r.EventMapKeys())
		//fmt.Println("\nEvents to flush: ", eventsToFlush)
		allFlushedEvents = append(allFlushedEvents, eventsToFlush...)
		//fmt.Println("\nAll Flushed events: ", allFlushedEvents)

		// Flush events to channel

		for _, ts := range eventsToFlush {
			//fmt.Println("About to send event with TS =", r.eventMap[ts].TimeStamp)
			evtChan <- r.eventMap[ts]
			//fmt.Println("Sent event with TS =", r.eventMap[ts].TimeStamp)
		}

		// Remove flushed events from reader's event map
		for _, ts := range eventsToFlush {
			delete(r.eventMap, ts)
		}
	}
}
