// Package reader implements structures and functions to read test bench data.
package reader

import (
	"bufio"
	"fmt"
	"log"
	"strconv"
	"strings"

	"gitlab.in2p3.fr/avirm/analysis-go/event"

	//"io"
)

// InputType describes the type (decimal/ASCII, hex/ASCII, binary) of an input file.
type InputType int

func (it *InputType) String() string {
	switch *it {
	case HexInput:
		return "Hex"
	case DecInput:
		return "Dec"
	case BinInput:
		return "Bin"
	default:
		return fmt.Sprintf("InputType(%d)", int(*it))
	}
}

// Set is the method to set the flag value.
func (it *InputType) Set(value string) error {
	switch value {
	case "Hex":
		*it = HexInput
	case "Dec":
		*it = DecInput
	case "Bin":
		*it = BinInput
	default:
		return fmt.Errorf("ascii: invalid input type value %q", value)
	}
	return nil
}

// HexInput, DecInput and BinInput correspond to the various input types
// TestBench-go has to deal with.
const (
	HexInput InputType = iota
	DecInput
	BinInput
)

type Scanner struct {
	s         *bufio.Scanner
	inputType InputType
	evtID     uint
}

func NewScanner(s *bufio.Scanner) *Scanner {
	ss := &Scanner{
		s:     s,
		evtID: 0,
	}
	return ss
}

func (s *Scanner) SetInputType(inputType InputType) {
	s.inputType = inputType
}

func (s *Scanner) readNextFrame(frameType TypeOfFrame) (*Frame, bool) {
	var lines []string
	status := true
	for status = s.s.Scan(); status; status = s.s.Scan() {
		text := s.s.Text()
		switch s.inputType {
		case HexInput:
			// no-op
		case DecInput: // decimal input
			textDec, err := strconv.ParseUint(text, 10, 64)
			if err != nil {
				log.Fatalf("error parsing %q: %v\n", text, err)
			}

			text = strconv.FormatUint(textDec, 16)
		case BinInput:
			panic("ascii: binary mode not implemented !")
		}
		if strings.Contains(text, "BADCAFE") || strings.Contains(text, "badcafe") {
			// Read the last two lines with 0xFFFFFFFF
			s.s.Scan()
			s.s.Scan()
			break
		}
		lines = append(lines, text)
	}
	frame := NewFrame(lines, frameType)
	//frame.PrintWoHeadersCounters()
	return frame, status
}

func (s *Scanner) ReadNextEvent() (*event.Event, bool) {
	frame, _ := s.readNextFrame(FirstFrameOfEvent)
	frameNext, status := s.readNextFrame(SecondFrameOfEvent)
	if status == false {
		return &event.Event{}, false
	}
	event := NewEvent(frame, frameNext, s.evtID)
	s.evtID++
	return event, status
}
