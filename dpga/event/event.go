package event

import (
	"fmt"
	"log"

	"gitlab.in2p3.fr/AVIRM/Analysis-go/pulse"
)

type Event struct {
	Clusters [72]pulse.Cluster
	ID       uint
}

func NewEventFromID(id uint) *Event {
	return &Event{
		ID: id,
	}
}

func (e *Event) Copy() *Event {
	newevent := NewEventFromID(e.ID)
	for i := range e.Clusters {
		oldPulses := e.Clusters[i].Pulses
		newevent.Clusters[i].Pulses = [4]pulse.Pulse{
			*oldPulses[0].Copy(),
			*oldPulses[1].Copy(),
			*oldPulses[2].Copy(),
			*oldPulses[3].Copy()}
	}
	return newevent
}

func (e *Event) CheckIntegrity() {
	for i := range e.Clusters {
		cluster := &e.Clusters[i]
		noSamples := e.Clusters[i].NoSamples()
		for j := 1; j < len(e.Clusters[i].Pulses); j++ {
			n := cluster.NoSamples()
			if n != noSamples {
				log.Fatal("not all pulses have the same number of samples")
			}
			pulse := &cluster.Pulses[j]
			if pulse.Channel == nil {
				log.Fatal("pulse has not associated channel")
			}
		}
	}
}

func (e *Event) Print(detailed bool) {
	fmt.Println("-> Printing event", e.ID)
	fmt.Println("    o number of clusters =", len(e.Clusters))
	if detailed {
		for i := range e.Clusters {
			cluster := e.Clusters[i]
			cluster.Print()
		}
	}
}
