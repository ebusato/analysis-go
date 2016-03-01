package event

import "gitlab.in2p3.fr/avirm/analysis-go/pulse"

type Data struct {
	Events []Event
}

func (d *Data) CheckIntegrity() {
	for i := range d.Events {
		d.Events[i].CheckIntegrity()
	}
}

func (d *Data) PlotPulses(xaxis pulse.XaxisType, pedestalRange bool) {
	for i := range d.Events {
		d.Events[i].PlotPulses(xaxis, pedestalRange)
		//if i >= 10 {
		//break
		//}
	}
}
