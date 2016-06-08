package event

import (
	"fmt"
	"log"

	"gitlab.in2p3.fr/avirm/analysis-go/pulse"
)

type Event struct {
	Clusters    []pulse.Cluster
	ID          uint
	IsCorrupted bool
}

func NewEvent(noClusters uint8) *Event {
	return &Event{
		Clusters:    make([]pulse.Cluster, noClusters),
		ID:          0,
		IsCorrupted: false,
	}
}

func (e *Event) Copy() *Event {
	newevent := NewEvent(uint8(e.NoClusters()))
	newevent.ID = e.ID
	for i := range e.Clusters {
		oldPulses := e.Clusters[i].Pulses
		newevent.Clusters[i].Pulses = [4]pulse.Pulse{
			*oldPulses[0].Copy(),
			*oldPulses[1].Copy(),
			*oldPulses[2].Copy(),
			*oldPulses[3].Copy()}
		newevent.Clusters[i].ID = e.Clusters[i].ID
		newevent.Clusters[i].Counters = make([]uint32, len(e.Clusters[i].Counters))
		copy(newevent.Clusters[i].Counters, e.Clusters[i].Counters)
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

func (e *Event) NoClusters() int {
	return len(e.Clusters)
}

func (e *Event) Print(printClusters bool, printClusterDetails bool) {
	fmt.Println("\n-> Printing event", e.ID)
	fmt.Println("    o number of clusters =", len(e.Clusters))
	if printClusters {
		for i := range e.Clusters {
			cluster := e.Clusters[i]
			cluster.Print(printClusterDetails)
		}
	}
}

func (e *Event) Multiplicity() (uint8, []*pulse.Pulse) {
	var mult uint8 = 0
	var pulsesWSig []*pulse.Pulse
	for i := range e.Clusters {
		pulsesWSigInCluster := e.Clusters[i].PulsesWithSignal()
		pulsesWSig = append(pulsesWSig, pulsesWSigInCluster...)
		mult += uint8(len(pulsesWSigInCluster))
	}
	return mult, pulsesWSig
}

func (e *Event) PlotPulses(x pulse.XaxisType, pedestalRange bool) {
	for i := range e.Clusters {
		cluster := &e.Clusters[i]
		if len(cluster.PulsesWithSignal()) > 0 {
			cluster.PlotPulses(e.ID, x, pedestalRange)
		}
	}
}

func (e *Event) PushPedestalSamples() {
	for iCluster := range e.Clusters {
		cluster := &e.Clusters[iCluster]
		for iPulse := range cluster.Pulses {
			pulse := &cluster.Pulses[iPulse]
			if pulse.HasSignal {
				continue
			}
			for iSample := range pulse.Samples {
				sample := &pulse.Samples[iSample]
				capacitor := sample.Capacitor
				noSamples := capacitor.NoPedestalSamples()
				if e.ID == 0 && noSamples != 0 {
					log.Fatal("noSamples != 0!")
				}
				capacitor.AddPedestalSample(sample.Amplitude)
			}
		}
	}
}
