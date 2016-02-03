package event

import "gitlab.in2p3.fr/avirm/analysis-go/pulse"

type Data []Event

func (d *Data) CheckIntegrity() {
	for i := range *d {
		(*d)[i].CheckIntegrity()
	}
}

// func HbookToGonum(histo ...hbook.H1D) []plotter.Histogram {
// 	output := make([]plotter.Histogram, len(histo))
// 	for i, h := range histo {
// 		h, err := plotter.NewHistogram(&h, h.Axis().Bins())
// 		if err != nil {
// 			panic(err)
// 		}
// 		fmt.Printf("histo: %+v\n", h)
// 		h.FillColor = nil //plotutil.Color(i)
// 		h.Color = plotutil.Color(i)
// 		output[i] = *h
// 	}
// 	return output
// }
//
// func MakePlot(xTitle string, yTitle string, outFile string, histo ...hbook.H1D) {
// 	p, err := plot.New()
// 	if err != nil {
// 		panic(err)
// 	}
//
// 	p.X.Label.Text = xTitle
// 	p.Y.Label.Text = yTitle
//
// 	hGonum := HbookToGonum(histo...)
// 	for i := range hGonum {
// 		p.Add(&hGonum[i])
// 	}
//
// 	if err := p.Save(4*vg.Inch, 4*vg.Inch, outFile); err != nil {
// 		panic(err)
// 	}
// }
//
// func (d *Data) PlotDistribs() {
// 	const N = 4
// 	hCharge := make([]hbook.H1D, N)
// 	hHasSignal := make([]hbook.H1D, N)
// 	hSRout := *hbook.NewH1D(1024, 0, 1023)
//
// 	for i := 0; i < N; i++ {
// 		hCharge[i] = *hbook.NewH1D(100, -100e3, 600e3)
// 		hHasSignal[i] = *hbook.NewH1D(2, 0, 2)
// 	}
// 	for _, event := range *d {
// 		for i := 0; i < N; i++ {
// 			pulse := event.Cluster.Pulses[i]
// 			hCharge[i].Fill(float64(pulse.Charge()), 1)
// 			hasSig := 0.
// 			switch pulse.HasSignal {
// 			case true:
// 				hasSig = 1
// 			case false:
// 				hasSig = 0
// 			}
// 			hHasSignal[i].Fill(hasSig, 1)
// 			if i == 0 {
// 				hSRout.Fill(float64(pulse.SRout), 1)
// 			}
// 		}
// 	}
//
// 	PrintHbookH1D(hHasSignal[0])
// 	MakePlot("Charge", "Entries (A. U.)", "output/distribCharge.png", hCharge...)
// 	MakePlot("HasSignal", "Entries (A. U.)", "output/distribHasSignal.png", hHasSignal...)
// 	MakePlot("SRout", "Entries (A. U.)", "output/distribSRout.png", hSRout)
// }

func (d *Data) PlotPulses(xaxis pulse.XaxisType, pedestalRange bool, savePulses bool) {
	for i := range *d {
		(*d)[i].PlotPulses(xaxis, pedestalRange)
		if i >= 4 {
			break
		}
	}
}

// func (d *Data) Plot() {
// 	d.PlotDistribs()
// 	d.PlotPulses(pulse.XaxisTime, false, true)
// }
