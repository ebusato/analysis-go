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
	r.err = blk.Integrity()
	if r.err != nil {
		blk.Print("medium")
		return
	}
}

func (r *Reader) readBlockHeader(blk *Block) {
	r.readU16(&blk.FirstBlockWord)
	r.read(&blk.AMCFrameCounters)
	blk.AMCFrameCounter = (uint32(blk.AMCFrameCounters[0]) << 16) + uint32(blk.AMCFrameCounters[1])
	r.readU16(&blk.ParityFEIdCtrl)
	blk.FrontEndId = (blk.ParityFEIdCtrl & 0x7fff) >> 8
	r.readU16(&blk.TriggerMode)
	r.readU16(&blk.Trigger)
	r.read(&blk.ASMFrameCounters)
	blk.ASMFrameCounter = (uint64(blk.ASMFrameCounters[0]) << 48) + (uint64(blk.ASMFrameCounters[1]) << 32) + (uint64(blk.ASMFrameCounters[2]) << 16) + uint64(blk.ASMFrameCounters[3])
	r.readU16(&blk.Cafe)
	r.readU16(&blk.Deca)
	r.read(&blk.Counters)
	r.read(&blk.TimeStamps)
	temp := (uint64(blk.TimeStamps[0]) << 16) | uint64(blk.TimeStamps[1])
	temp = (temp << 32)
	temp1 := (uint64(blk.TimeStamps[2]) << 16) | uint64(blk.TimeStamps[3])
	// 	temp |= temp1
	blk.TimeStamp = temp | temp1
	r.readU16(&blk.NoSamples)
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
	r.read(&data.ParityChanIdCtrl)
	data.Channel = (data.ParityChanIdCtrl & 0x7f00) >> 8
	blk.QuartetAbsIdx60 = dpgadetector.FEIdAndChanIdToQuartetAbsIdx60(blk.FrontEndId, data.Channel)

	//fmt.Printf("%v, %x, %v, %v, %v\n", noAttempts, data.ParityChanIdCtrl, data.Channel, blk.QuartetAbsIdx60, QuartetAbsIdx60old)
	if (data.ParityChanIdCtrl & 0xff) != ctrl0xfd {
		//panic("(data.ParityChanIdCtrl & 0xff) != ctrl0xfd")
		return true
	}
	if i > 0 && blk.QuartetAbsIdx60 != QuartetAbsIdx60old {
		//panic("i > 0 && blk.QuartetAbsIdx60 != QuartetAbsIdx60old")
		return true
	}
	QuartetAbsIdx60old = blk.QuartetAbsIdx60
	return false
}

func (r *Reader) readBlockData(blk *Block) {
	if r.err != nil {
		return
	}
	for i := range blk.Data.Data {
		data := &blk.Data.Data[i]
		for r.readParityChanIdCtrl(blk, i) {
			noAttempts++
			if noAttempts >= 5 {
				log.Fatalf("reader.readParityChanIdCtrl: noAttempts >= 3\n")
			}
		}
		noAttempts = 0
		r.read(&data.Amplitudes)
	}
}

func (r *Reader) readBlockTrailer(blk *Block) {
	r.readU16(&blk.CRC)
	// Temporary fix, until we understand where these additionnal 16 bits come from
	if blk.CRC != ctrl0xCRC {
		r.readU16(&blk.CRC)
	}
	// End of temporary fix
	r.readU16(&blk.ParityFEIdCtrl2)
	if (blk.ParityFEIdCtrl2&0x7fff)>>8 != blk.FrontEndId {
		log.Fatalf("Front end ids in header and trailer don't match\n")
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
		w.Frame(frame)
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

/*
func (r *Reader) ReadNextEvent() (*event.Event, bool) {

}*/
