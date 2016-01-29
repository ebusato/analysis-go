package event

import (
	"bufio"
	"log"

	"gitlab.in2p3.fr/AVIRM/Analysis-go/pulse"
)

type Event struct {
	Cluster pulse.Cluster
	ID      uint
}

func (e *Event) Copy() *Event {
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

func (e *Event) Print(detailed bool) {

}

func (e *Event) PrintPulsesToFile(w *bufio.Writer) {

}

func (e *Event) PrintGlobalVarsToFile(w *bufio.Writer) {

}

func (e *Event) PlotPulses(x pulse.XaxisType, pedestalRange bool) string {

}
