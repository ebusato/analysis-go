package event

import (
	"fmt"
	"log"

	"gitlab.in2p3.fr/avirm/analysis-go/pulse"
)

type Event struct {
	Clusters        []pulse.Cluster
	ClustersWoData  [12]pulse.Cluster // These are the 12 clusters corresponding to the 12*4 channels unused for data at the end of each ASM board
	ID              uint
	TimeStamp       uint64
	Counters        []uint32
	IsCorrupted     bool
	UDPPayloadSizes []int // number of octets for each frame making this event (FrameSize has NoFrames components)
}

func NewEvent(noClusters uint8) *Event {
	return &Event{
		Clusters:    make([]pulse.Cluster, noClusters),
		ID:          0,
		TimeStamp:   0,
		IsCorrupted: false,
	}
}

func (e *Event) Copy() *Event {
	newevent := NewEvent(uint8(e.NoClusters()))
	newevent.ID = e.ID
	newevent.TimeStamp = e.TimeStamp
	newevent.UDPPayloadSizes = e.UDPPayloadSizes
	newevent.Counters = make([]uint32, len(e.Counters))
	for i := range e.Counters {
		newevent.Counters[i] = e.Counters[i]
	}
	newevent.IsCorrupted = e.IsCorrupted
	for i := range e.Clusters {
		oldPulses := e.Clusters[i].Pulses
		newevent.Clusters[i].Pulses = [4]pulse.Pulse{
			*oldPulses[0].Copy(),
			*oldPulses[1].Copy(),
			*oldPulses[2].Copy(),
			*oldPulses[3].Copy()}
		newevent.Clusters[i].ID = e.Clusters[i].ID
	}
	for i := range e.ClustersWoData {
		oldPulses := e.ClustersWoData[i].Pulses
		newevent.ClustersWoData[i].Pulses = [4]pulse.Pulse{
			*oldPulses[0].Copy(),
			*oldPulses[1].Copy(),
			*oldPulses[2].Copy(),
			*oldPulses[3].Copy()}
		newevent.ClustersWoData[i].ID = e.ClustersWoData[i].ID
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
	fmt.Println("    o ID =", e.ID)
	fmt.Println("    o TimeStamp =", e.TimeStamp)
	fmt.Println("    o UDPPayloadSizes =", e.UDPPayloadSizes)
	if printClusters {
		for i := range e.Clusters {
			cluster := e.Clusters[i]
			cluster.Print(printClusterDetails)
		}
		for i := range e.ClustersWoData {
			cluster := e.ClustersWoData[i]
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

func (e *Event) PlotPulses(x pulse.XaxisType, onlyClustersWithSig bool, yrange pulse.YRange, xRangeZoomAround500 bool) {
	for i := range e.Clusters {
		cluster := &e.Clusters[i]
		if (onlyClustersWithSig == true && len(cluster.PulsesWithSignal()) > 0) || !onlyClustersWithSig {
			cluster.PlotPulses(e.ID, x, yrange, xRangeZoomAround500)
		}
	}
}

func (e *Event) pushPedestalCluster(cluster *pulse.Cluster) {
	for iPulse := range cluster.Pulses {
		pulse := &cluster.Pulses[iPulse]
		if pulse.HasSignal {
			continue
		}
		for iSample := range pulse.Samples {
			sample := &pulse.Samples[iSample]
			capacitor := sample.Capacitor
			if capacitor != nil {
				noSamples := capacitor.NoPedestalSamples()
				if e.ID == 0 && noSamples != 0 {
					log.Fatal("noSamples != 0!")
				}
				capacitor.AddPedestalSample(sample.Amplitude)
			}
		}
	}
}

func (e *Event) PushPedestalSamples() {
	for iCluster := range e.Clusters {
		cluster := &e.Clusters[iCluster]
		e.pushPedestalCluster(cluster)
		// 		for iPulse := range cluster.Pulses {
		// 			pulse := &cluster.Pulses[iPulse]
		// 			if pulse.HasSignal {
		// 				continue
		// 			}
		// 			for iSample := range pulse.Samples {
		// 				sample := &pulse.Samples[iSample]
		// 				capacitor := sample.Capacitor
		// 				if capacitor != nil {
		// 					noSamples := capacitor.NoPedestalSamples()
		// 					if e.ID == 0 && noSamples != 0 {
		// 						log.Fatal("noSamples != 0!")
		// 					}
		// 					capacitor.AddPedestalSample(sample.Amplitude)
		// 				}
		// 			}
		// 		}
	}
	for iCluster := range e.ClustersWoData {
		cluster := &e.ClustersWoData[iCluster]
		e.pushPedestalCluster(cluster)
	}
}

func (e *Event) pushTimeDepOffsetCluster(cluster *pulse.Cluster) {
	for iPulse := range cluster.Pulses {
		pulse := &cluster.Pulses[iPulse]
		if pulse.HasSignal {
			continue
		}
		ch := pulse.Channel
		if ch == nil {
			panic("pulse has no channel associated to it.")
		}
		ch.IncrementNoTimeDepOffsetSamples()
		for iSample := range pulse.Samples {
			sample := &pulse.Samples[iSample]
			ch.AddTimeDepOffsetSample(iSample, sample.Amplitude)
		}
	}
}

func (e *Event) PushTimeDepOffsetSamples() {
	for iCluster := range e.Clusters {
		cluster := &e.Clusters[iCluster]
		e.pushTimeDepOffsetCluster(cluster)
		// 		for iPulse := range cluster.Pulses {
		// 			pulse := &cluster.Pulses[iPulse]
		// 			if pulse.HasSignal {
		// 				continue
		// 			}
		// 			ch := pulse.Channel
		// 			if ch == nil {
		// 				panic("pulse has no channel associated to it.")
		// 			}
		// 			ch.IncrementNoTimeDepOffsetSamples()
		// 			for iSample := range pulse.Samples {
		// 				sample := &pulse.Samples[iSample]
		// 				ch.AddTimeDepOffsetSample(iSample, sample.Amplitude)
		// 			}
		// 		}
	}
	for iCluster := range e.ClustersWoData {
		cluster := &e.ClustersWoData[iCluster]
		e.pushTimeDepOffsetCluster(cluster)
	}
}

// center is the center of the intervall
// sig is the energy resolution
// n is the number of standard deviations considered
func (e *Event) PulsesInEnergyWindow(center, n, sig float64) []*pulse.Pulse {
	var selectedPulses []*pulse.Pulse
	for c := range e.Clusters {
		cluster := &e.Clusters[c]
		for p := range cluster.Pulses {
			pulse := &cluster.Pulses[p]
			if pulse.E > center-n*sig && pulse.E < center+n*sig {
				selectedPulses = append(selectedPulses, pulse)
			}
		}
	}
	return selectedPulses
}

// type Trigger int
//
// const (
// 	LvsR Trigger = iota + 1
// 	Lvs3L
// 	LvsL
// 	None
// )
//
// func WhichTrigger(pulses []*pulse.Pulse) Trigger {
// 	var trig Trigger
// 	for i := range pulses {
//
// 	}
// 	return trig
// }
