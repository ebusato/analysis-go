package rw

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	//"log"

	//"gitlab.in2p3.fr/avirm/analysis-go/testbench/event"
)

// Writer wraps an io.Writer and writes an ASM stream.
type Writer struct {
	w            io.Writer
	err          error
	hdr          Header
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
// Close does not close the underlying io.Reader.
func (w *Writer) Close() error {
	w.write(lastFrame)
	if ww, ok := w.w.(*bufio.Writer); ok {
		ww.Flush()
	}
	if w.err != nil && w.err != io.EOF {
		return w.err
	}
	return nil
}

// Header writes the Header to the ASM stream.
func (w *Writer) Header(hdr Header) error {
	if w.err != nil {
		return w.err
	}
	if w.hdr.Size != 0 {
		return fmt.Errorf("asm: header already written")
	}
	w.hdr = hdr
	w.writeHeader(w.hdr)
	return w.err
}

// Frame writes a Frame to the ASM stream.
func (w *Writer) Frame(f Frame) error {
	if w.err != nil {
		return w.err
	}
	if w.hdr.Size == 0 {
		w.hdr.Size = uint32(len(f.Block.Data))
		w.hdr.NumFrame = 0

		w.writeHeader(w.hdr)
		if w.err != nil {
			return w.err
		}
	}

	if int(numSamples) != len(f.Block.Data) {
		return fmt.Errorf("asm: inconsistent number of samples")
	}
	w.writeFrame(&f)
	return w.err
}

func (w *Writer) write(v uint32) {
	if w.err != nil {
		return
	}
	w.err = binary.Write(w.w, binary.BigEndian, v)
}

func (w *Writer) writeHeader(hdr Header) {
	w.write(hdr.Size)
	w.write(hdr.NumFrame)
}

func (w *Writer) writeFrame(f *Frame) {
	if w.err != nil {
		return
	}
	w.write(f.ID)
	if f.ID == lastFrame {
		w.err = io.EOF
		return
	}
	w.writeBlock(&f.Block, 0)
}

func (w *Writer) writeBlock(blk *Block, fid uint32) {
	w.writeBlockHeader(blk)
	w.writeBlockData(blk)
	w.write((blockTrailer << 4) + fid)
}

func (w *Writer) writeBlockHeader(blk *Block) {
	if w.err != nil {
		return
	}
	//fmt.Printf("rw: writing block header: %v %v %x\n", blk.Evt, blk.ID, blockHeader)
	w.write(blk.Evt)
	w.write(blk.ID)
	w.write(blockHeader)
}

func (w *Writer) writeBlockData(blk *Block) {
	if w.err != nil {
		return
	}
	for _, v := range blk.Data {
		w.write(v)
	}
	w.write(blk.SRout)
	for _, v := range blk.Counters {
		w.write(v)
	}
}

/*
func (w *Writer) Event(event *event.Event) {
	iframe := 0
	for i := range event.Clusters {
		if iframe >= 2 {
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

		if cluster.NoSamples() != numSamples {
			log.Fatalf("rw: cluster.NoSamples() != numSamples")
		}

		pulses := &cluster.Pulses

		if pulses[0].Channel.FifoID() != pulses[1].Channel.FifoID() {
			log.Fatalf("pulses[0].Channel.FifoID() != pulses[1].Channel.FifoID()")
		}
		if pulses[2].Channel.FifoID() != pulses[3].Channel.FifoID() {
			log.Fatalf("pulses[2].Channel.FifoID() != pulses[3].Channel.FifoID()")
		}
		block1.ID = uint32(pulses[0].Channel.FifoID())
		block2.ID = uint32(pulses[2].Channel.FifoID())

		block1.SRout = uint32(cluster.SRout())
		block2.SRout = block1.SRout

		for j := uint16(0); j < numSamples; j++ {
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

		w.Frame(frame1)
		w.Frame(frame2)
		iframe += 2
	}
}
*/