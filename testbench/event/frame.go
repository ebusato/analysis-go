package event

import (
	"fmt"
	"log"
	"strconv"

	"gitlab.in2p3.fr/AVIRM/Analysis-go/detector"
	"gitlab.in2p3.fr/AVIRM/Analysis-go/pulse"
	"gitlab.in2p3.fr/AVIRM/Analysis-go/testbench/tbdetector"
)

type TypeOfFrame byte

const (
	FirstFrameOfEvent TypeOfFrame = iota
	SecondFrameOfEvent
)

type Frame struct {
	lines     []string
	frameType TypeOfFrame
}

func NewFrame(lines []string, frameType TypeOfFrame) *Frame {
	frame := &Frame{
		lines:     lines,
		frameType: frameType,
	}
	return frame
}

func (f *Frame) RemoveHeaderAndCounters() []string {
	linesWoHeaderCounters := f.lines[1:1000]
	return linesWoHeaderCounters
}

func (f *Frame) SRout() uint16 {
	n, err := strconv.ParseUint(f.lines[1000], 16, 16)
	if err != nil || n < 0 || n > 1023 {
		log.Fatalf("error parsing uint %q: %v\n", f.lines[1001:1001], err)
	}
	return uint16(n)
}

func (f *Frame) Print() {
	for i, line := range f.lines {
		fmt.Println(i, line)
	}
}

func (f *Frame) PrintWoHeadersCounters() {
	linesWoHeaderCounters := f.RemoveHeaderAndCounters()
	for i, line := range linesWoHeaderCounters {
		fmt.Println(i, line)
	}
}

func (f *Frame) MakePulses() (*pulse.Pulse, *pulse.Pulse) {
	var chan1 *detector.Channel
	var chan2 *detector.Channel
	switch f.frameType {
	case FirstFrameOfEvent:
		chan1 = tbdetector.Det.Channel(0)
		chan2 = tbdetector.Det.Channel(1)
	case SecondFrameOfEvent:
		chan1 = tbdetector.Det.Channel(2)
		chan2 = tbdetector.Det.Channel(3)
	default:
		panic("cannot make pulse, frame type not recognized")
	}

	pulse1 := pulse.NewPulse(chan1)
	pulse2 := pulse.NewPulse(chan2)
	pulse1.SRout = f.SRout()
	pulse2.SRout = f.SRout()
	linesWoHeaderCounters := f.RemoveHeaderAndCounters()
	for i, line := range linesWoHeaderCounters {
		lineI, err := strconv.ParseUint(line, 16, 32)
		if err != nil {
			log.Fatalf("error parsing uint %q: %v\n", line, err)
		}

		lineUI32 := uint32(lineI)

		ampl2 := float64(lineUI32 & 0xFFF)
		ampl1 := float64(lineUI32 >> 16)

		sample1 := pulse.NewSample(ampl1, uint16(i), float64(i))
		sample2 := pulse.NewSample(ampl2, uint16(i), float64(i))

		pulse1.AddSample(sample1)
		pulse2.AddSample(sample2)
	}

	//pulse1.Print()
	//pulse2.Print()
	return pulse1, pulse2
}
