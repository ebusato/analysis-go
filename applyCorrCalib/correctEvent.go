// Package applyCorrCalib is intended to implement functions applying the various corrections necessary for data analysis.
// For the moment, only the pedestal correction is implemented.
package applyCorrCalib

import "gitlab.in2p3.fr/avirm/analysis-go/event"

func CorrectEvent(e *event.Event, doPedestal bool, doTimeDepOffset bool, doEnergyCalib bool) *event.Event {
	if doPedestal == false {
		if doTimeDepOffset == true {
			panic("doTimeDepOffset == true && doPedestal == false")
		}
		// do nothing, return the original event
		return e
	}
	newevent := e.Copy()
	for i := range newevent.Clusters {
		cluster := &newevent.Clusters[i]
		for j := range cluster.Pulses {
			if doPedestal {
				cluster.Pulses[j].SubtractPedestal()
				if doTimeDepOffset {
					cluster.Pulses[j].SubtractTimeDepOffsets()
					if doEnergyCalib {
						//fmt.Println("pointer", i, j, cluster.Pulses[j].Channel)
						cluster.Pulses[j].Amplitude()
						if cluster.Pulses[j].Channel != nil {
							cluster.Pulses[j].E = cluster.Pulses[j].Channel.EnergyCalib.Y(cluster.Pulses[j].Ampl)
							//fmt.Println("E=", cluster.Pulses[j].Ampl, cluster.Pulses[j].E)
						}
					}
				}
			}
		}
	}
	return newevent
}
