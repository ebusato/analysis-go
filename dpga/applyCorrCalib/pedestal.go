// maybe remove it not used anymore

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
