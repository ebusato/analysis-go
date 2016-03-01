// Package applyCorrCalib is intended to implement functions applying the various corrections necessary for data analysis.
// For the moment, only the pedestal correction is implemented.
package applyCorrCalib

import "gitlab.in2p3.fr/avirm/analysis-go/dpga/event"

func RemovePedestal(e *event.Event) *event.Event {
	newevent := e.Copy()
	for i := range newevent.Clusters {
		cluster := &newevent.Clusters[i]
		for j := range cluster.Pulses {
			cluster.Pulses[j].SubtractPedestal()
		}
	}
	return newevent
}
