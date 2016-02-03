package reader

import (
	"bufio"
	"io"
	"os"
	"reflect"
	"testing"
)

func TestReader(t *testing.T) {
	f, err := os.Open("testdata/data-03-frames.bin")
	if err != nil {
		t.Fatalf("could not open data file: %v\n", err)
	}
	defer f.Close()

	r, err := NewReader(bufio.NewReader(f))
	if err != nil {
		t.Fatalf("could not open asm file: %v\n", err)
	}

	hdr := r.Header()
	const (
		hdrSize    uint32 = 1007
		hdrNFrames uint32 = 3
	)
	if hdr.Size != hdrSize {
		t.Fatalf("invalid header.Size. got=%d. want=%d", hdr.Size, hdrSize)
	}
	if hdr.NumFrame != hdrNFrames {
		t.Fatalf("invalid header.NumFrame. got=%d. want=%d",
			hdr.NumFrame,
			hdrNFrames,
		)
	}

	nframes := 0
	for {
		frame, err := r.Frame()
		if err != nil {
			if err != io.EOF {
				t.Fatalf("error loading frame: %v\n", err)
			}
			if frame.ID != lastFrame {
				t.Fatalf("invalid last frame id. got=%d. want=%d",
					frame.ID,
					lastFrame,
				)
			}
			break
		}
		if frame.ID != uint32(nframes) {
			t.Fatalf(
				"frame.id differ. got=%d want=%d\n",
				frame.ID,
				uint32(nframes),
			)
		}
		//frame.Print()
		nframes++
	}
	if nframes != 3 {
		t.Fatalf("got %d frames. want 3\n", nframes)
	}
}

func testReadWrite(t *testing.T) {
	const (
		rname = "testdata/data-03-frames.bin"
		wname = "testdata/wdata-03-frames.bin"
	)

	rhdr, rframes := testRead(t, rname)

	f, err := os.Create(wname)
	if err != nil {
		t.Fatalf("could not create data file: %v\n", err)
	}
	defer f.Close()
	defer os.Remove(wname)

	w := NewWriter(bufio.NewWriter(f))
	if err != nil {
		t.Fatalf("could not open asm file: %v\n", err)
	}

	err = w.Header(rhdr)
	if err != nil {
		t.Fatalf("error: %v\n", err)
	}

	for _, frame := range rframes {
		err = w.Frame(frame)
		if err != nil {
			if err == io.EOF {
				break
			}
			t.Fatalf("error: %v\n", err)
		}
	}

	err = w.Close()
	if err != nil {
		t.Fatalf("error closing asm-stream: %v\n", err)
	}

	err = f.Close()
	if err != nil {
		t.Fatalf("error closing output file: %v\n", err)
	}

	whdr, wframes := testRead(t, wname)

	if !reflect.DeepEqual(rhdr, whdr) {
		t.Fatalf("headers differ.\ngot= %#v\nwant=%v\n", whdr, rhdr)
	}

	if !reflect.DeepEqual(rframes, wframes) {
		t.Fatalf("frames differ.\ngot= %#v\nwant=%v\n", wframes, rframes)
	}
}

func testRead(t *testing.T, name string) (Header, []Frame) {
	f, err := os.Open(name)
	if err != nil {
		t.Fatalf("could not open data file [%s]: %v\n", name, err)
	}
	defer f.Close()

	r, err := NewReader(bufio.NewReader(f))
	if err != nil {
		t.Fatalf("could not open asm file [%s]: %v\n", name, err)
	}

	hdr := r.Header()
	frames := []Frame{}

	for {
		frame, err := r.Frame()
		frames = append(frames, frame)
		if err != nil {
			if err != io.EOF {
				t.Fatalf("[%s]: error loading frame: %v\n", name, err)
			}
			break
		}
	}
	return hdr, frames
}
