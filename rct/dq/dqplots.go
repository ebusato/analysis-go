// Package dq implements structures and functions to be used for data quality assessment.
package dq

import (
	"encoding/gob"
	"fmt"
	"image/color"
	"math"
	"os"
	"strconv"

	"gitlab.in2p3.fr/avirm/analysis-go/dpga/dpgadetector"
	"gitlab.in2p3.fr/avirm/analysis-go/event"
	"gitlab.in2p3.fr/avirm/analysis-go/utils"
	"go-hep.org/x/hep/hbook"
	"go-hep.org/x/hep/hplot"
	"gonum.org/v1/plot"
	"gonum.org/v1/plot/palette/brewer"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
	"gonum.org/v1/plot/vg/draw"
)

type DQPlot struct {
	Nevents             uint
	HFrequency          *hbook.H1D
	HSatFrequency       *hbook.H1D
	HMultiplicity       *hbook.H1D
	HSatMultiplicity    *hbook.H1D
	HLORMult            *hbook.H1D
	HCharge             [][]hbook.H1D
	HAmplitude          [][]hbook.H1D
	HEnergy             [][]hbook.H1D
	HMinRecX            *hbook.H1D
	HMinRecY            *hbook.H1D
	HMinRecZ            *hbook.H1D
	DeltaT30            *hbook.H1D
	HEnergyAll          *hbook.H1D
	AmplCorrelation     *hbook.H2D
	EnergyCorrelation   *hbook.H2D
	HitQuartets         *hbook.H2D
	HEnergyVsDeltaTggRF *hbook.H2D

	SRout [3]hbook.H1D // One value for each DRS
}

func NewDQPlot() *DQPlot {
	const N = 4
	NoClusters := uint8(5)
	dqp := &DQPlot{
		HFrequency:       hbook.NewH1D(24, 0, 24),
		HSatFrequency:    hbook.NewH1D(24, 0, 24),
		HMultiplicity:    hbook.NewH1D(8, -0.5, 7.5),
		HSatMultiplicity: hbook.NewH1D(8, -0.5, 7.5),
		HLORMult:         hbook.NewH1D(20, 0, 20),
		HCharge:          make([][]hbook.H1D, NoClusters),
		HAmplitude:       make([][]hbook.H1D, NoClusters),
		HEnergy:          make([][]hbook.H1D, NoClusters),
		HMinRecX:         hbook.NewH1D(200, -50, 50),
		HMinRecY:         hbook.NewH1D(200, -50, 50),
		// 		HMinRecZ:         hbook.NewH1D(1000, -150, 150),
		HMinRecZ:   hbook.NewH1D(61, -97.5-3.25/2., 97.5+3.25/2.),
		DeltaT30:   hbook.NewH1D(300, -30, 30),
		HEnergyAll: hbook.NewH1D(200, 0, 1022),
		// 		AmplCorrelation: hbook.NewH2D(50, 0, 0.5, 50, 0, 0.5),
		AmplCorrelation:     hbook.NewH2D(50, 0, 4095, 50, 0, 4095),
		EnergyCorrelation:   hbook.NewH2D(50, 0, 1000, 50, 0, 1000),
		HitQuartets:         hbook.NewH2D(30, 0, 30, 30, 30, 60),
		HEnergyVsDeltaTggRF: hbook.NewH2D(50, 0, 40, 50, 0, 1050),
	}
	for i := uint8(0); i < NoClusters; i++ {
		dqp.HCharge[i] = make([]hbook.H1D, N)
		dqp.HAmplitude[i] = make([]hbook.H1D, N)
		dqp.HEnergy[i] = make([]hbook.H1D, N)
		for j := 0; j < N; j++ {
			dqp.HCharge[i][j] = *hbook.NewH1D(100, 0, 0.5)
			dqp.HAmplitude[i][j] = *hbook.NewH1D(150, 0, 4095)
			dqp.HEnergy[i][j] = *hbook.NewH1D(200, 0, 1022)
		}
	}
	for k := 0; k < 3; k++ {
		dqp.SRout[k] = *hbook.NewH1D(1024, 0, 1024)
	}
	return dqp
}

func (d *DQPlot) FillHistos(event *event.Event, RFcutMean, RFcutWidth float64) {
	d.Nevents++

	var mult uint8 = 0
	var satmult uint8 = 0
	var counter float64 = 0

	// Used to make sure one fills only once the SRout histograms
	var SRoutBool [3]bool

	for i := range event.Clusters {
		cluster := &event.Clusters[i]
		mult += uint8(len(cluster.PulsesWithSignal()))
		satmult += uint8(len(cluster.PulsesWithSatSignal()))
		for j := range cluster.Pulses {
			pulse := &cluster.Pulses[j]
			if pulse.HasSignal {
				d.HFrequency.Fill(counter, 1)
			}
			if pulse.HasSatSignal {
				d.HSatFrequency.Fill(counter, 1)
			}
			if pulse.HasSignal {
				d.HCharge[i][j].Fill(float64(pulse.Charge()/1e6), 1)
				_, ampl := pulse.Amplitude()
				d.HAmplitude[i][j].Fill(ampl, 1)
				d.HEnergy[i][j].Fill(pulse.E, 1)
				d.HEnergyAll.Fill(pulse.E, 1)
			}
			counter++
		}

		if cluster.Quartet != nil { // for rct, this selects only cluster for which one has data
			// 			fmt.Printf("Quartet %p\n", cluster.Quartet)
			// 			fmt.Printf("DRS %p\n", cluster.Quartet.DRS)
			iDRS := cluster.Quartet.DRS.ID()
			// 			fmt.Println("LALA: ", iDRS, iASM, hemi.Which())
			if SRoutBool[iDRS] == false {
				d.SRout[iDRS].Fill(float64(cluster.SRout), 1)
				SRoutBool[iDRS] = true
			}
		}
	}

	// 	fmt.Println("SRout entries =", d.SRout[0][0][2].Entries())

	d.HMultiplicity.Fill(float64(mult), 1)
	d.HSatMultiplicity.Fill(float64(satmult), 1)

	d.HLORMult.Fill(float64(len(event.LORs)), 1)
	// 						fmt.Println(len(event.LORs))
	if len(event.LORs) == 1 {
		lor := &event.LORs[0]
		d.AmplCorrelation.Fill(lor.Pulses[0].Ampl, lor.Pulses[1].Ampl, 1)
		d.EnergyCorrelation.Fill(lor.Pulses[0].E, lor.Pulses[1].E, 1)
		quartet0 := float64(dpgadetector.FifoID144ToQuartetAbsIdx60(lor.Pulses[0].Channel.FifoID144(), true))
		quartet1 := float64(dpgadetector.FifoID144ToQuartetAbsIdx60(lor.Pulses[1].Channel.FifoID144(), true))
		d.HitQuartets.Fill(quartet0, quartet1, 1)

		timeDiff := lor.TMean - lor.TRF

		d.HEnergyVsDeltaTggRF.Fill(timeDiff, lor.Pulses[0].E, 1)
		d.HEnergyVsDeltaTggRF.Fill(timeDiff, lor.Pulses[1].E, 1)

		if math.Abs(timeDiff-RFcutMean) > RFcutWidth {
			d.HMinRecX.Fill(lor.Xmar, 1)
			d.HMinRecY.Fill(lor.Ymar, 1)
			d.HMinRecZ.Fill(lor.Zmar, 1)
		}
	}

	lors := event.FindLORsLose(0, 0)
	for i := range lors {
		lor := &lors[i]
		d.DeltaT30.Fill(lor.Pulses[0].Time30-lor.Pulses[1].Time30, 1)
	}
}

func (d *DQPlot) Finalize() {
	d.HFrequency.Scale(1 / float64(d.Nevents))
	d.HSatFrequency.Scale(1 / float64(d.Nevents))
	d.HMultiplicity.Scale(1 / d.HMultiplicity.Integral())
	d.HLORMult.Scale(1 / d.HLORMult.Integral())
	d.HSatMultiplicity.Scale(1 / d.HSatMultiplicity.Integral())
	// Take len of HCharge and HCharge[0] as it should be the same for all other
	// objects used here
	for i := 0; i < len(d.HCharge); i++ {
		for j := 0; j < len(d.HCharge[0]); j++ {
			d.HCharge[i][j].Scale(1 / d.HCharge[i][j].Integral())
			d.HAmplitude[i][j].Scale(1 / d.HAmplitude[i][j].Integral())
		}
	}
}

func (d *DQPlot) MakeFreqTiledPlot() *hplot.TiledPlot {
	tp := hplot.NewTiledPlot(draw.Tiles{Cols: 1, Rows: 2, PadY: 1 * vg.Centimeter})

	p1 := tp.Plot(0, 0)
	p1.X.Min = 0
	p1.X.Max = 24
	p1.Y.Min = 0
	p1.X.Label.Text = "channel"
	p1.Y.Label.Text = "No pulses"
	p1.X.Tick.Marker = &hplot.FreqTicks{N: 25, Freq: 4}
	p1.Add(hplot.NewGrid())
	hplotfreq := hplot.NewH1D(d.HFrequency)
	hplotfreq.FillColor = color.RGBA{R: 255, G: 204, B: 153, A: 255}
	hplotfreq.Color = plotutil.Color(3)
	p1.Add(hplotfreq)
	//p1.Title.Text = fmt.Sprintf("Number of pulses vs channel\n")

	p2 := tp.Plot(1, 0)
	p2.X.Min = 0
	p2.X.Max = 24
	p2.Y.Min = 0
	p2.X.Label.Text = "channel"
	p2.Y.Label.Text = "No sat. pulses"
	p2.X.Tick.Marker = &hplot.FreqTicks{N: 25, Freq: 4}
	p2.Add(hplot.NewGrid())
	hplotsatfreq := hplot.NewH1D(d.HSatFrequency)
	hplotsatfreq.FillColor = color.RGBA{R: 255, G: 204, B: 153, A: 255}
	hplotsatfreq.Color = plotutil.Color(3)
	p2.Add(hplotsatfreq)
	//p2.Title.Text = fmt.Sprintf("Number of saturating pulses vs channel\n")
	return tp
}

func (d *DQPlot) MakeSRoutTiledPlot() *hplot.TiledPlot {
	tp := hplot.NewTiledPlot(draw.Tiles{Cols: 3, Rows: 1, PadY: 1 * vg.Centimeter})

	for k := 0; k < 3; k++ {
		p1 := tp.Plot(0, k)
		p1.X.Min = 0
		p1.X.Max = 240
		p1.Y.Min = 0
		p1.X.Label.Text = "SRout"
		p1.Y.Label.Text = "Entries"
		p1.X.Tick.Marker = &hplot.FreqTicks{N: 21, Freq: 5}
		p1.Add(hplot.NewGrid())
		hplot := hplot.NewH1D(&d.SRout[k])
		hplot.FillColor = color.RGBA{R: 255, G: 204, B: 153, A: 255}
		hplot.Color = plotutil.Color(3)
		p1.Add(hplot)

	}

	return tp
}

func (d *DQPlot) MakeMinRecXYDistrs() *hplot.TiledPlot {
	tp := hplot.NewTiledPlot(draw.Tiles{Cols: 1, Rows: 2, PadY: 0.2 * vg.Centimeter})
	p1 := tp.Plot(0, 0)
	p1.X.Min = -50
	p1.X.Max = 50
	p1.X.Label.Text = "X (mm)"
	p1.Y.Label.Text = "No entries"
	p1.X.Tick.Marker = &hplot.FreqTicks{N: 101, Freq: 5}
	p1.Add(hplot.NewGrid())
	hplotX := hplot.NewH1D(d.HMinRecX)
	hplotX.FillColor = color.RGBA{R: 255, G: 204, B: 153, A: 255}
	hplotX.Color = plotutil.Color(3)
	p1.Add(hplotX)
	p1.BackgroundColor = color.RGBA{R: 230, G: 247, B: 255, A: 255}
	//p1.Title.Text = fmt.Sprintf("Distribution of minimal reconstruction X (mm)\n")

	p2 := tp.Plot(1, 0)
	p2.X.Min = -50
	p2.X.Max = 50
	p2.X.Label.Text = "Y (mm)"
	p2.Y.Label.Text = "No entries"
	p2.X.Tick.Marker = &hplot.FreqTicks{N: 101, Freq: 5}
	p2.Add(hplot.NewGrid())
	hplotY := hplot.NewH1D(d.HMinRecY)
	hplotY.FillColor = color.RGBA{R: 255, G: 204, B: 153, A: 255}
	hplotY.Color = plotutil.Color(3)
	p2.Add(hplotY)
	p2.BackgroundColor = color.RGBA{R: 230, G: 247, B: 255, A: 255}
	//p2.Title.Text = fmt.Sprintf("Distribution of minimal reconstruction Y (mm)\n")

	return tp
}

func (d *DQPlot) MakeMinRecZDistr() *plot.Plot {
	p, err := plot.New()

	p.X.Min = -10
	p.X.Max = 10
	if err != nil {
		panic(err)
	}
	p.X.Label.Text = "Z (mm)"
	p.Y.Label.Text = "No entries"
	p.X.Tick.Marker = &hplot.FreqTicks{N: 124, Freq: 6}
	p.BackgroundColor = color.RGBA{R: 230, G: 247, B: 255, A: 255}

	hplotZ := hplot.NewH1D(d.HMinRecZ)
	hplotZ.FillColor = color.RGBA{R: 255, G: 204, B: 153, A: 255}
	// 	hplotZ.Color = plotutil.Color(3)
	p.Add(hplotZ)
	p.Add(hplot.NewGrid())
	return p
}

type WhichVar byte

const (
	Charge WhichVar = iota
	Amplitude
	Energy
)

func (d *DQPlot) MakeChargeAmplTiledPlot(whichV WhichVar) *hplot.TiledPlot {
	tp := hplot.NewTiledPlot(draw.Tiles{Cols: 5, Rows: 1, PadY: 0.5 * vg.Centimeter})
	histos := make([]hbook.H1D, len(d.HCharge[0]))

	iCluster := 0

	irow := 0
	for icol := 0; icol < 5; icol++ {
		switch whichV {
		case Charge:
			histos = d.HCharge[iCluster]
		case Amplitude:
			histos = d.HAmplitude[iCluster]
		case Energy:
			histos = d.HEnergy[iCluster]
		}
		p := tp.Plot(irow, icol)
		p.Title.Text = "Channel " + strconv.FormatInt(int64(iCluster*4), 10) + " -> " + strconv.FormatInt(int64(iCluster*4+3), 10)
		switch whichV {
		case Amplitude:
			p.X.Tick.Marker = &hplot.FreqTicks{N: 6, Freq: 1}
		case Charge:
			p.X.Tick.Marker = &hplot.FreqTicks{N: 4, Freq: 1}
		case Energy:
			p.X.Tick.Marker = &hplot.FreqTicks{N: 9, Freq: 1}
		}
		p.Add(hplot.NewGrid())
		p.Plot.X.LineStyle.Width = 2
		p.Plot.Y.LineStyle.Width = 2
		p.Plot.X.Tick.LineStyle.Width = 2
		p.Plot.Y.Tick.LineStyle.Width = 2
		hplot0 := hplot.NewH1D(&histos[0])
		hplot1 := hplot.NewH1D(&histos[1])
		hplot2 := hplot.NewH1D(&histos[2])
		hplot3 := hplot.NewH1D(&histos[3])
		hplot0.Color = color.RGBA{R: 238, G: 46, B: 47, A: 255}  // red
		hplot1.Color = color.RGBA{R: 0, G: 140, B: 72, A: 255}   // green
		hplot2.Color = color.RGBA{R: 24, G: 90, B: 169, A: 255}  // blue
		hplot3.Color = color.RGBA{R: 250, G: 88, B: 244, A: 255} // pink
		// 			hplot0.FillColor = color.NRGBA{R: 238, G: 46, B: 47, A: 80}
		// 			hplot1.FillColor = color.NRGBA{R: 0, G: 140, B: 72, A: 80}
		// 			hplot2.FillColor = color.NRGBA{R: 24, G: 90, B: 169, A: 80}
		// 			hplot3.FillColor = color.NRGBA{R: 250, G: 88, B: 244, A: 80}
		hplot0.FillColor = nil
		hplot1.FillColor = nil
		hplot2.FillColor = nil
		hplot3.FillColor = nil
		hplot0.LineStyle.Width = 1
		hplot1.LineStyle.Width = 1
		hplot2.LineStyle.Width = 1
		hplot3.LineStyle.Width = 1
		hplot0.FillColor = color.NRGBA{R: 238, G: 46, B: 47, A: 80}
		hplot1.FillColor = color.NRGBA{R: 0, G: 140, B: 72, A: 80}
		hplot2.FillColor = color.NRGBA{R: 24, G: 90, B: 169, A: 80}
		hplot3.FillColor = color.NRGBA{R: 250, G: 88, B: 244, A: 80}
		p.Add(hplot0, hplot1, hplot2, hplot3)
		iCluster++
	}
	return tp
}

func (d *DQPlot) MakeAmplCorrelationPlot() *plot.Plot {
	pCorrelation, err := plot.New()
	if err != nil {
		panic(err)
	}
	pCorrelation.X.Label.Text = "Amplitude pulse 0 (ADC counts)"
	pCorrelation.Y.Label.Text = "Amplitude pulse 1 (ADC counts)"
	pCorrelation.X.Tick.Marker = &hplot.FreqTicks{N: 11, Freq: 2}
	pCorrelation.Y.Tick.Marker = &hplot.FreqTicks{N: 11, Freq: 2}
	pCorrelation.X.Min = d.AmplCorrelation.XMin()
	pCorrelation.Y.Min = d.AmplCorrelation.YMin()
	pCorrelation.X.Max = d.AmplCorrelation.XMax()
	pCorrelation.Y.Max = d.AmplCorrelation.YMax()
	pCorrelation.Add(hplot.NewH2D(d.AmplCorrelation, nil))
	//pCorrelation.Add(plotter.NewGrid())
	return pCorrelation
}

func (d *DQPlot) MakeEnergyCorrelationPlot() *plot.Plot {
	pCorrelation, err := plot.New()
	if err != nil {
		panic(err)
	}
	pCorrelation.X.Label.Text = "Energy pulse 0 (keV)"
	pCorrelation.Y.Label.Text = "Energy pulse 1 (keV)"
	pCorrelation.X.Tick.Marker = &hplot.FreqTicks{N: 11, Freq: 2}
	pCorrelation.Y.Tick.Marker = &hplot.FreqTicks{N: 11, Freq: 2}
	pCorrelation.X.Min = d.EnergyCorrelation.XMin()
	pCorrelation.Y.Min = d.EnergyCorrelation.YMin()
	pCorrelation.X.Max = d.EnergyCorrelation.XMax()
	pCorrelation.Y.Max = d.EnergyCorrelation.YMax()
	pCorrelation.Add(hplot.NewH2D(d.EnergyCorrelation, nil))
	//pCorrelation.Add(plotter.NewGrid())
	pCorrelation.BackgroundColor = color.RGBA{R: 230, G: 247, B: 255, A: 255}
	return pCorrelation
}

func (d *DQPlot) MakeRFPlotALaArnaud() *plot.Plot {
	pRF, err := plot.New()
	if err != nil {
		panic(err)
	}
	pRF.X.Label.Text = "tgg - trf (ns)"
	pRF.Y.Label.Text = "Energy (keV)"
	pRF.X.Tick.Marker = &hplot.FreqTicks{N: 11, Freq: 2}
	pRF.Y.Tick.Marker = &hplot.FreqTicks{N: 11, Freq: 2}
	pRF.X.Min = d.HEnergyVsDeltaTggRF.XMin()
	pRF.Y.Min = d.HEnergyVsDeltaTggRF.YMin()
	pRF.X.Max = d.HEnergyVsDeltaTggRF.XMax()
	pRF.Y.Max = d.HEnergyVsDeltaTggRF.YMax()
	pRF.Add(hplot.NewH2D(d.HEnergyVsDeltaTggRF, nil))
	pRF.BackgroundColor = color.RGBA{R: 230, G: 247, B: 255, A: 255}
	//pRF.Add(plotter.NewGrid())
	return pRF
}

func (d *DQPlot) MakeHitQuartetsPlot() *plot.Plot {
	pHitQuartets, err := plot.New()
	if err != nil {
		panic(err)
	}
	pHitQuartets.X.Label.Text = "Quartet Id (right hemisphere)"
	pHitQuartets.Y.Label.Text = "Quartet Id (left hemisphere)"
	pHitQuartets.X.Tick.Marker = &hplot.FreqTicks{N: 31, Freq: 2}
	pHitQuartets.Y.Tick.Marker = &hplot.FreqTicks{N: 31, Freq: 2}
	pHitQuartets.X.Min = d.HitQuartets.XMin()
	pHitQuartets.Y.Min = d.HitQuartets.YMin()
	pHitQuartets.X.Max = d.HitQuartets.XMax()
	pHitQuartets.Y.Max = d.HitQuartets.YMax()
	p, _ := brewer.GetPalette(brewer.TypeAny, "RdYlBu", 11)
	pHitQuartets.Add(hplot.NewH2D(d.HitQuartets, p))
	//pHitQuartets.Add(plotter.NewGrid())
	pHitQuartets.BackgroundColor = color.RGBA{R: 230, G: 247, B: 255, A: 255}
	return pHitQuartets
}

func (d *DQPlot) MakeDeltaT30Plot() *hplot.Plot {
	p := hplot.New()
	p.X.Label.Text = "Delta T30 (ns)"
	p.Y.Label.Text = "No entries"
	p.X.Tick.Marker = &hplot.FreqTicks{N: 61, Freq: 5}
	hp := hplot.NewH1D(d.DeltaT30)
	hp.FillColor = color.RGBA{R: 102, G: 102, B: 255, A: 255}
	p.Add(hp)
	p.Add(hplot.NewGrid())
	p.BackgroundColor = color.RGBA{R: 230, G: 247, B: 255, A: 255}
	return p
}

func (d *DQPlot) MakeLORMultPlot() *hplot.Plot {
	p := hplot.New()
	p.X.Label.Text = "Number of LORs"
	p.Y.Label.Text = "No entries"
	p.X.Tick.Marker = &hplot.FreqTicks{N: 21, Freq: 2}
	hp := hplot.NewH1D(d.HLORMult)
	hp.FillColor = color.RGBA{R: 255, G: 255, B: 51, A: 255}
	p.Add(hp)
	p.Add(hplot.NewGrid())
	p.BackgroundColor = color.RGBA{R: 230, G: 247, B: 255, A: 255}
	return p
}

func (d *DQPlot) MakeEnergyPlot() *hplot.Plot {
	p := hplot.New()
	p.X.Label.Text = "Energy (keV)"
	p.Y.Label.Text = "No entries"
	p.X.Tick.Marker = &hplot.FreqTicks{N: 9, Freq: 1}
	var hp *hplot.H1D
	hp = hplot.NewH1D(d.HEnergyAll)
	hp.FillColor = color.RGBA{R: 77, G: 255, B: 136, A: 255}
	p.Add(hp)
	p.Add(hplot.NewGrid())
	p.BackgroundColor = color.RGBA{R: 230, G: 247, B: 255, A: 255}
	return p
}

// SaveHistos saves histograms on disk.
// If d.DQPlotRef is not nil, current histograms
// are overlaid with the provided reference histograms.
func (d *DQPlot) SaveHistos() {
	doplot := utils.MakeHPl
	// 	doplot := utils.MakeGonumPlot

	dqplotref := &DQPlot{}

	linestyle := draw.LineStyle{Width: vg.Points(2)}
	linestyleref := draw.LineStyle{Width: vg.Points(1), Dashes: []vg.Length{vg.Points(5), vg.Points(5)}}

	for i := range d.HAmplitude {
		doplot("Amplitude",
			"Entries",
			"output/distribAmplitude"+strconv.FormatInt(int64(i), 10)+".png",
			utils.H1dToHplot(linestyle, d.HAmplitude[i]...)...)
		doplot("Charge",
			"Entries",
			"output/distribCharge"+strconv.FormatInt(int64(i), 10)+".png",
			utils.H1dToHplot(linestyle, d.HCharge[i]...)...)
	}
	doplot("Channel",
		"# pulses / event",
		"output/distribFrequency.png",
		append(utils.H1dptrToHplot(linestyle, d.HFrequency),
			utils.H1dptrToHplot(linestyleref, dqplotref.HFrequency)...)...)
	doplot("Channel",
		"# pulses with saturation / event",
		"output/distribSatFrequency.png",
		append(utils.H1dptrToHplot(linestyle, d.HSatFrequency),
			utils.H1dptrToHplot(linestyleref, dqplotref.HSatFrequency)...)...)
	doplot("Multiplicity",
		"Entries",
		"output/distribMultiplicity.png",
		append(utils.H1dptrToHplot(linestyle, d.HMultiplicity),
			utils.H1dptrToHplot(linestyleref, dqplotref.HMultiplicity)...)...)
	doplot("Multiplicity of pulses with saturation",
		"Entries",
		"output/distribSatMultiplicity.png",
		append(utils.H1dptrToHplot(linestyle, d.HSatMultiplicity),
			utils.H1dptrToHplot(linestyleref, dqplotref.HSatMultiplicity)...)...)
}

func (d *DQPlot) WriteGob(fileName string) error {
	f, err := os.Create(fileName)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := gob.NewEncoder(f)
	fmt.Printf("Saving DQplot to %s\n", fileName)
	err = enc.Encode(d)
	if err != nil {
		return err
	}
	err = f.Close()
	if err != nil {
		return err
	}

	return err
}
