// maybe remove it not used anymore

package applyCorrCalib

import "gitlab.in2p3.fr/AVIRM/Analysis-go/testbench/event"

// to be put in a separate file
var pedestalCoeffs = map[string]float64{
	"PMT1": 480,
	"PMT2": 450,
	"PMT3": 450,
	"PMT4": 450,
}

// to be updated
func RemovePedestal(e *event.Event) *event.Event {
	newevent := e.Copy()
	for i := range newevent.Cluster.Pulses {
		pulse := &newevent.Cluster.Pulses[i]
		pulse.SubtractPedestal()
	}
	return newevent
}
