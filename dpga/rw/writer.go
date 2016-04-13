package rw

import (
	"bufio"
	"encoding/binary"
	"io"
	"log"
	"time"

	"gitlab.in2p3.fr/avirm/analysis-go/event"
	"gitlab.in2p3.fr/avirm/analysis-go/pulse"
)

// Writer wraps an io.Writer and writes an ASM stream.
type Writer struct {
	w            io.Writer
	err          error
	hdr          *Header
	frameCounter uint32
}

// NewWriter returns a new ASM stream in write mode.
func NewWriter(w io.Writer) *Writer {
	return &Writer{w: w}
}

// Write implements io.Writer.
func (w *Writer) Write(data []byte) (int, error) {
	return w.w.Write(data)
}

// Close closes the ASM stream.
//
// Close flushes any pending data.
// Close does not close the underlying io.Writer.
func (w *Writer) Close() error {
	w.writeU32(lastFrame)
	if ww, ok := w.w.(*bufio.Writer); ok {
		ww.Flush()
	}
	if w.err != nil && w.err != io.EOF {
		return w.err
	}
	return nil
}

// Header writes the Header to the ASM stream.
func (w *Writer) Header(hdr *Header, clientTime bool) error {
	if w.err != nil {
		return w.err
	}
	w.hdr = hdr
	w.writeHeader(w.hdr, clientTime)
	return w.err
}

// Frame writes a Frame to the ASM stream.
func (w *Writer) Frame(f *Frame) error {
	if w.err != nil {
		return w.err
	}
	w.writeFrame(f)
	return w.err
}

func (w *Writer) write(v interface{}) {
	if w.err != nil {
		return
	}
	w.err = binary.Write(w.w, binary.BigEndian, v)
}

func (w *Writer) writeU32(v uint32) {
	if w.err != nil {
		return
	}
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], v)
	_, w.err = w.w.Write(buf[:])
}

func (w *Writer) writeHeader(hdr *Header, clientTime bool) {
	switch {
	case hdr.HdrType == HeaderCAL:
		w.writeU32(hdr.History)
		w.writeU32(hdr.RunNumber)
		w.writeU32(hdr.FreeField)
		if clientTime {
			// Hack: set time from client clock rather than from server's
			// since the later is not correct.
			hdr.TimeStart = uint32(time.Now().Unix())
		}
		w.writeU32(hdr.TimeStart)
		w.writeU32(hdr.TimeStop)
		w.writeU32(hdr.NoEvents)
		w.writeU32(hdr.NoASMCards)
		w.writeU32(hdr.NoSamples)
		w.writeU32(hdr.DataToRead)
		w.writeU32(hdr.TriggerEq)
		w.writeU32(hdr.TriggerDelay)
		w.writeU32(hdr.ChanUsedForTrig)
		w.writeU32(hdr.LowHighThres)
		w.writeU32(hdr.TrigSigShapingHighThres)
		w.writeU32(hdr.TrigSigShapingLowThres)
	case hdr.HdrType == HeaderOld:
		w.writeU32(hdr.Size)
		w.writeU32(hdr.NumFrame)
	default:
		panic("error ! header type not known")
	}
}

func (w *Writer) writeFrame(f *Frame) {
	if w.err != nil {
		return
	}
	w.writeU32(f.ID)
	if f.ID == lastFrame {
		w.err = io.EOF
		return
	}
	w.writeBlock(&f.Block, 0)
}

func (w *Writer) writeBlock(blk *Block, fid uint32) {
	w.writeBlockHeader(blk)
	w.writeBlockData(blk)
	w.writeU32((blockTrailer << 4) + fid)
}

func (w *Writer) writeBlockHeader(blk *Block) {
	if w.err != nil {
		return
	}
	//fmt.Printf("rw: writing block header: %v %v %x\n", blk.Evt, blk.ID, blockHeader)
	w.writeU32(blk.Evt)
	w.writeU32(blk.ID)
	w.writeU32(blockHeader)
}

func (w *Writer) writeBlockData(blk *Block) {
	if w.err != nil {
		return
	}
	for _, v := range blk.Data {
		w.writeU32(v)
	}
	w.writeU32(blk.SRout)
	for _, v := range blk.Counters {
		w.writeU32(v)
	}
}

func (w *Writer) MakeFrame(iCluster int, evtID uint32, pulse0 *pulse.Pulse, pulse1 *pulse.Pulse) *Frame {
	frame := &Frame{ID: w.frameCounter}
	w.frameCounter++
	block := &frame.Block
	block.Evt = evtID
	if pulse0.Channel.FifoID144() != pulse1.Channel.FifoID144() {
		log.Fatalf("pulses[0].Channel.FifoID() != pulses[1].Channel.FifoID()")
	}
	block.ID = uint32(pulse0.Channel.FifoID144())
	if pulse0.SRout != pulse1.SRout {
		log.Fatalf("pulse0.SRout != pulse1.SRout")
	}
	block.SRout = uint32(pulse0.SRout)
	if pulse0.NoSamples() != pulse1.NoSamples() {
		log.Fatalf("pulse0.NoSamples() != pulse1.NoSamples()\n")
	}
	for j := uint16(0); j < pulse0.NoSamples(); j++ {
		amp0, amp1 := uint32(pulse0.Samples[j].Amplitude), uint32(pulse1.Samples[j].Amplitude)
		word := (amp0&0xFFF)<<16 | (amp1 & 0xFFF)
		block.Data = append(block.Data, word)
	}
	return frame
}

func (w *Writer) Event(event *event.Event) {
	for iCluster := range event.Clusters {
		cluster := &event.Clusters[iCluster]
		pulses := &cluster.Pulses

		if uint8(len(cluster.Counters)) != numCounters {
			log.Fatalf("rw: len(cluster.Counters) = %v, numCounters = %v", len(cluster.Counters), numCounters)
		}

		if pulses[0].NoSamples() != 0 || pulses[1].NoSamples() != 0 {
			frame := w.MakeFrame(iCluster, uint32(event.ID), &pulses[0], &pulses[1])
			for j := 0; j < int(numCounters); j++ {
				frame.Block.Counters[j] = cluster.Counter(j)
			}
			w.Frame(frame)
		}
		if pulses[2].NoSamples() != 0 || pulses[3].NoSamples() != 0 {
			frame := w.MakeFrame(iCluster, uint32(event.ID), &pulses[2], &pulses[3])
			for j := 0; j < int(numCounters); j++ {
				frame.Block.Counters[j] = cluster.Counter(j)
			}
			w.Frame(frame)
		}
	}
}

func (w *Writer) EventFull(event *event.Event) {
	iframe := 0
	for i := range event.Clusters {
		if iframe >= 120 {
			log.Fatalf("rw: iframe out of range")
		}

		frame1 := Frame{ID: w.frameCounter}
		w.frameCounter++
		frame2 := Frame{ID: w.frameCounter}
		w.frameCounter++

		block1 := &frame1.Block
		block2 := &frame2.Block

		block1.Evt = uint32(event.ID)
		block2.Evt = uint32(event.ID)

		cluster := &event.Clusters[i]

		pulses := &cluster.Pulses

		// 		fmt.Println("pulses[0].Channel=", pulses[0].Channel)
		// 		fmt.Println("pulses[1].Channel=", pulses[1].Channel)

		if pulses[0].Channel.FifoID144() != pulses[1].Channel.FifoID144() {
			log.Fatalf("pulses[0].Channel.FifoID() != pulses[1].Channel.FifoID()")
		}
		if pulses[2].Channel.FifoID144() != pulses[3].Channel.FifoID144() {
			log.Fatalf("pulses[2].Channel.FifoID() != pulses[3].Channel.FifoID()")
		}
		block1.ID = uint32(pulses[0].Channel.FifoID144())
		block2.ID = uint32(pulses[2].Channel.FifoID144())

		block1.SRout = uint32(cluster.SRout())
		block2.SRout = block1.SRout

		for j := uint16(0); j < cluster.NoSamples(); j++ {
			// Make block1 data from pulse[0] and pulse[1]
			amp0, amp1 := uint32(pulses[0].Samples[j].Amplitude), uint32(pulses[1].Samples[j].Amplitude)
			word := (amp0&0xFFF)<<16 | (amp1 & 0xFFF)
			block1.Data = append(block1.Data, word)
			// Make block2 data from pulse[2] and pulse[3]
			amp2, amp3 := uint32(pulses[2].Samples[j].Amplitude), uint32(pulses[3].Samples[j].Amplitude)
			word = (amp2&0xFFF)<<16 | (amp3 & 0xFFF)
			block2.Data = append(block2.Data, word)
		}

		if uint8(len(cluster.Counters)) != numCounters {
			log.Fatalf("rw: len(cluster.Counters) = %v, numCounters = %v", len(cluster.Counters), numCounters)
		}

		for j := 0; j < int(numCounters); j++ {
			block1.Counters[j] = cluster.Counter(j)
			block2.Counters[j] = cluster.Counter(j)
		}

		w.Frame(&frame1)
		w.Frame(&frame2)
		iframe += 2
	}
}
