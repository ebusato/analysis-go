package rw

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
)

// Writer wraps an io.Writer and writes an ASM stream.
type Writer struct {
	w   io.Writer
	err error
	hdr Header
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

	// FIXME(sbinet): frame-id seems to always be 0...
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
func (w *Writer) WriteNextEvent(event *event.Event) {
	// Build frames for this event
	// One event is made of 120 frames
	var frames []Frame
	for i := range event.Clusters {
		cluster := &event.Clusters[i]

		frame1 := &Frame{ID: 0}
		frame2 := &Frame{ID: 0}

		block1 := &frame1.Block
		block2 := &frame2.Block

		block1.Evt = event.ID
		block2.Evt = event.ID

		frames = append(frames, *frame1, *frame2)
	}

	// write frames
	if len(frames) != 120 {
		log.Fatalf("rw: number of frames not correct, expect 120, get %v", len(frames))
	}
}
*/
