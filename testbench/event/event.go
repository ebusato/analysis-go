package event

import (
	"fmt"
	"log"

	"gitlab.in2p3.fr/AVIRM/Analysis-go/pulse"
)

type Event struct {
	Cluster pulse.Cluster
	ID      uint
}

func NewEvent(frame1 *Frame, frame2 *Frame, evtID uint) *Event {
	pulse1, pulse2 := frame1.MakePulses()
	pulse3, pulse4 := frame2.MakePulses()
	event := &Event{
		Cluster: *pulse.NewCluster(0, [4]pulse.Pulse{*pulse1, *pulse2, *pulse3, *pulse4}),
	}
	event.ID = evtID
	event.CheckIntegrity()
	return event
}

func NewEventFromID(id uint) *Event {
	return &Event{
		ID: id,
	}
}

func (e *Event) Copy() *Event {
	newevent := NewEventFromID(e.ID)
	//newevent.Pulses = make([]pulse.Pulse, len(e.Pulses))
	oldPulses := e.Cluster.Pulses
	newevent.Cluster.Pulses = [4]pulse.Pulse{
		*oldPulses[0].Copy(),
		*oldPulses[1].Copy(),
		*oldPulses[2].Copy(),
		*oldPulses[3].Copy()}
	return newevent
}

func (e *Event) PulseFromName(name string) *pulse.Pulse {
	for i := range e.Cluster.Pulses {
		if e.Cluster.Pulses[i].Channel == nil {
			log.Fatal("no detector channel associated to pulse")
		}
		if e.Cluster.Pulses[i].Channel.Name() == name {
			return &e.Cluster.Pulses[i]
		}
	}
	return nil
}

func (e *Event) CheckIntegrity() {
	if len(e.Cluster.Pulses) == 0 {
		panic("event has no pulse")
	}
	noSamples := e.NoSamples(0)
	for i := 1; i < len(e.Cluster.Pulses); i++ {
		n := e.NoSamples(i)
		if n != noSamples {
			panic("not all pulses have the same number of samples")
		}
		pulse := &e.Cluster.Pulses[i]
		if pulse.Channel == nil {
			log.Fatal("pulse has not associated channel")
		}
	}
}

func (e *Event) NoPulses() int {
	return len(e.Cluster.Pulses)
}

func (e *Event) NoSamples(pulseID int) int {
	return len(e.Cluster.Pulses[pulseID].Samples)
}

func (e *Event) Print(detailed bool) {
	fmt.Println("-> Printing event", e.ID)
	fmt.Println("    o number of pulses =", len(e.Cluster.Pulses))
	if detailed {
		for i, p := range e.Cluster.Pulses {
			fmt.Printf("     - pulse %v (SRout = %v):\n", i, p.SRout)
			for _, s := range p.Samples {
				fmt.Printf("      * sample %v: %v\n", s.Index, s.Amplitude)
			}
		}
	}
}

func (e *Event) PlotPulses(x pulse.XaxisType, pedestalRange bool) string {
	return e.Cluster.PlotPulses(e.ID, x, pedestalRange)
}

func (e *Event) GlobalCorrelation(name1 string, name2 string) float64 {
	pulse1 := e.PulseFromName(name1)
	pulse2 := e.PulseFromName(name2)
	if pulse1 == nil || pulse2 == nil {
		panic("either pulse1 or pulse2 are nil")
	}
	return pulse1.Correlation(pulse2)
}
