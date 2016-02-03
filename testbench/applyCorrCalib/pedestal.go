// maybe remove it not used anymore

package applyCorrCalib

import "gitlab.in2p3.fr/avirm/analysis-go/testbench/event"

func RemovePedestal(e *event.Event) *event.Event {
	newevent := e.Copy()
	for i := range newevent.Cluster.Pulses {
		pulse := &newevent.Cluster.Pulses[i]
		pulse.SubtractPedestal()
	}
	return newevent
}
