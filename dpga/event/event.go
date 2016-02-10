package event

import (
	"fmt"
	"log"

	"gitlab.in2p3.fr/avirm/analysis-go/dpga/dpgadetector"
	"gitlab.in2p3.fr/avirm/analysis-go/pulse"
)

type Event struct {
	Clusters []pulse.Cluster
	ID       uint
}

func NewEvent() *Event {
	return &Event{
		Clusters: make([]pulse.Cluster, dpgadetector.Det.NoClusters()),
		ID:       0,
	}
}

func (e *Event) Copy() *Event {
	newevent := NewEvent()
	newevent.ID = e.ID
	for i := range e.Clusters {
		oldPulses := e.Clusters[i].Pulses
		newevent.Clusters[i].Pulses = [4]pulse.Pulse{
			*oldPulses[0].Copy(),
			*oldPulses[1].Copy(),
			*oldPulses[2].Copy(),
			*oldPulses[3].Copy()}
		newevent.Clusters[i].ID = e.Clusters[i].ID
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

func (e *Event) Multiplicity() uint8 {
	var mult uint8 = 0
	for i := range e.Clusters {
		mult += uint8(len(e.Clusters[i].PulsesWithSignal()))
	}
	return mult
}

func (e *Event) PlotPulses(x pulse.XaxisType, pedestalRange bool) {
	for i := range e.Clusters {
		e.Clusters[i].PlotPulses(e.ID, x, pedestalRange)
	}
}
