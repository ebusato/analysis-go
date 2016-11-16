package rw

import (
	"bufio"
	"encoding/binary"
	"io"
)

// Writer wraps an io.Writer and writes an ASM stream.
type Writer struct {
	w   io.Writer
	err error
	//hdr          *Header
	//frameCounter uint32
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
	//w.writeU32(lastFrame)
	if ww, ok := w.w.(*bufio.Writer); ok {
		ww.Flush()
	}
	if w.err != nil && w.err != io.EOF {
		return w.err
	}
	return nil
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

func (w *Writer) writeByte(v byte) {
	if w.err != nil {
		return
	}
	var buf []byte
	buf = append(buf, v)
	_, w.err = w.w.Write(buf[:])
}

func (w *Writer) writeU16(v uint16) {
	if w.err != nil {
		return
	}
	var buf [2]byte
	binary.BigEndian.PutUint16(buf[:], v)
	_, w.err = w.w.Write(buf[:])
}

func (w *Writer) WriteU16(v uint16) {
	w.writeU16(v)
}

func (w *Writer) writeFrame(f *Frame) {
	if w.err != nil {
		return
	}
	w.writeHeader(f)
	w.writeData(f)
	w.writeTrailer(f)
}

func (w *Writer) writeHeader(f *Frame) {
	if w.err != nil {
		return
	}
	//fmt.Printf("rw: writing block header: %v %v %x\n", f.Evt, f.ID, blockHeader)
	w.writeU16(f.FirstBlockWord)
	w.write(f.AMCFrameCounters)
	w.writeU16(f.ParityFEIdCtrl)
	w.writeU16(f.TriggerMode)
	w.writeU16(f.Trigger)
	w.write(f.ASMFrameCounters)
	w.writeU16(f.Cafe)
	w.writeU16(f.Deca)
	w.write(f.Counters)
	w.write(f.TimeStamps)
	w.writeU16(f.NoSamples)
}

func (w *Writer) writeData(f *Frame) {
	if w.err != nil {
		return
	}
	for i := range f.Data.Data {
		data := &f.Data.Data[i]
		w.writeU16(data.ParityChanIdCtrl)
		w.write(data.Amplitudes)
		if f.Err == ErrorCode1 {
			//fmt.Println("ErrorCode1, add extra word")
			w.writeU16(uint16(0))
		}
	}
}

func (w *Writer) writeTrailer(f *Frame) {
	w.writeU16(f.CRC)
	w.writeU16(f.ParityFEIdCtrl2)
}

/*
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

		if pulses[0].NoSamples() != 0 || pulses[1].NoSamples() != 0 {
			frame := w.MakeFrame(iCluster, uint32(event.ID), &pulses[0], &pulses[1])
			if uint8(len(cluster.CountersFifo1)) != numCounters {
				log.Fatalf("rw: len(cluster.CountersFifo1) = %v, numCounters = %v", len(cluster.CountersFifo1), numCounters)
			}
			for j := 0; j < int(numCounters); j++ {
				frame.Block.Counters[j] = cluster.CounterFifo1(j)
			}
			w.Frame(frame)
		}
		if pulses[2].NoSamples() != 0 || pulses[3].NoSamples() != 0 {
			frame := w.MakeFrame(iCluster, uint32(event.ID), &pulses[2], &pulses[3])
			if uint8(len(cluster.CountersFifo2)) != numCounters {
				log.Fatalf("rw: len(cluster.CountersFifo2) = %v, numCounters = %v", len(cluster.CountersFifo2), numCounters)
			}
			for j := 0; j < int(numCounters); j++ {
				frame.Block.Counters[j] = cluster.CounterFifo2(j)
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

		if uint8(len(cluster.CountersFifo1)) != numCounters {
			log.Fatalf("rw: len(cluster.CountersFifo1) = %v, numCounters = %v", len(cluster.CountersFifo1), numCounters)
		}
		if uint8(len(cluster.CountersFifo2)) != numCounters {
			log.Fatalf("rw: len(cluster.CountersFifo2) = %v, numCounters = %v", len(cluster.CountersFifo2), numCounters)
		}
		for j := 0; j < int(numCounters); j++ {
			block1.Counters[j] = cluster.CounterFifo1(j)
			block2.Counters[j] = cluster.CounterFifo2(j)
		}

		w.Frame(&frame1)
		w.Frame(&frame2)
		iframe += 2
	}
}
*/
