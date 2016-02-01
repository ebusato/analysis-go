package event

import (
	"bufio"
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
		oldPulses := e.Clusters.Pulses
		newevent.Clusters[i].Pulses = [4]pulse.Pulse{
			*oldPulses[0].Copy(),
			*oldPulses[1].Copy(),
			*oldPulses[2].Copy(),
			*oldPulses[3].Copy()}
	}
}

func (e *Event) CheckIntegrity() {
	for i := range e.Clusters {
		cluster := &e.Clusters[i]
		if len(cluster.Pulses) == 0 {
			log.Fatalf("cluster %v has no pulse", i)
		}
		noSamples := e.Cluster[i].NoSamples(0)
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
}

func (e *Event) Print(detailed bool) {

}

func (e *Event) PrintPulsesToFile(w *bufio.Writer) {

}

func (e *Event) PrintGlobalVarsToFile(w *bufio.Writer) {

}

func (e *Event) PlotPulses(x pulse.XaxisType, pedestalRange bool) string {

}
