package event

import "gitlab.in2p3.fr/avirm/analysis-go/pulse"

type Data []Event

func (d *Data) CheckIntegrity() {
	for i := range *d {
		(*d)[i].CheckIntegrity()
	}
}

func (d *Data) PlotPulses(xaxis pulse.XaxisType, pedestalRange bool) {
	for i := range *d {
		(*d)[i].PlotPulses(xaxis, pedestalRange)
		if i >= 3 {
			break
		}
	}
}
