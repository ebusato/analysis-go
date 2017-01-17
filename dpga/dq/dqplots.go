// Package dq implements structures and functions to be used for data quality assessment.
package dq

import (
	"encoding/gob"
	"fmt"
	"image/color"
	"log"
	"os"
	"strconv"

	"github.com/go-hep/hbook"
	"github.com/go-hep/hplot"
	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/plotutil"
	"github.com/gonum/plot/vg"
	"github.com/gonum/plot/vg/draw"
	"gitlab.in2p3.fr/avirm/analysis-go/dpga/dpgadetector"
	"gitlab.in2p3.fr/avirm/analysis-go/event"
	"gitlab.in2p3.fr/avirm/analysis-go/utils"
)

type DQPlot struct {
	Nevents          uint
	HFrequency       *hbook.H1D
	HSatFrequency    *hbook.H1D
	HMultiplicity    *hbook.H1D
	HSatMultiplicity *hbook.H1D
	HCharge          [][]hbook.H1D
	HAmplitude       [][]hbook.H1D
	HEnergy          [][]hbook.H1D
	HMinRecX         *hbook.H1D
	HMinRecY         *hbook.H1D
	HMinRecZ         *hbook.H1D
	DeltaT30         *hbook.H1D
	AmplCorrelation  *hbook.H2D
	HitQuartets      *hbook.H2D

	HV [4][16]plotter.XYs // first index refers to HV card (there are 4 cards), second index refers to channels (there are 16 channels per card)

	DQPlotRef *DQPlot
}

func NewDQPlot() *DQPlot {
	const N = 4
	NoClusters := dpgadetector.Det.NoClusters()
	dqp := &DQPlot{
		HFrequency:       hbook.NewH1D(240, 0, 240),
		HSatFrequency:    hbook.NewH1D(240, 0, 240),
		HMultiplicity:    hbook.NewH1D(8, -0.5, 7.5),
		HSatMultiplicity: hbook.NewH1D(8, -0.5, 7.5),
		HCharge:          make([][]hbook.H1D, NoClusters),
		HAmplitude:       make([][]hbook.H1D, NoClusters),
		HEnergy:          make([][]hbook.H1D, NoClusters),
		HMinRecX:         hbook.NewH1D(200, -50, 50),
		HMinRecY:         hbook.NewH1D(240, -60, 60),
		HMinRecZ:         hbook.NewH1D(300, -150, 150),
		DeltaT30:         hbook.NewH1D(300, -30, 30),
		// 		AmplCorrelation: hbook.NewH2D(50, 0, 0.5, 50, 0, 0.5),
		AmplCorrelation: hbook.NewH2D(50, 0, 4095, 50, 0, 4095),
		HitQuartets:     hbook.NewH2D(30, 0, 30, 30, 30, 60),
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
	return dqp
}

func NewDQPlotFromGob(fileName string) *DQPlot {
	fmt.Printf("Reading gob file %s\n", fileName)
	f, err := os.Open(fileName)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	dec := gob.NewDecoder(f)

	dqplot := NewDQPlot()

	err = dec.Decode(dqplot)
	if err != nil {
		panic(err)
	}

	err = f.Close()
	if err != nil {
		panic(err)
	}
	return dqplot
}

func (d *DQPlot) FillHistos(event *event.Event) {
	d.Nevents++

	var mult uint8 = 0
	var satmult uint8 = 0
	var counter float64 = 0

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
			}
			counter++
		}
	}

	d.HMultiplicity.Fill(float64(mult), 1)
	d.HSatMultiplicity.Fill(float64(satmult), 1)
}

// AddHVPoint adds a point to the HV curve.
// abscissa is whatever you think is more relevant in your case.
func (d *DQPlot) AddHVPoint(idCard int, idChannel int, abscissa float64, val float64) {
	d.HV[idCard][idChannel] = append(d.HV[idCard][idChannel], struct{ X, Y float64 }{X: abscissa, Y: val})
}

func (d *DQPlot) Finalize() {
	d.HFrequency.Scale(1 / float64(d.Nevents))
	d.HSatFrequency.Scale(1 / float64(d.Nevents))
	d.HMultiplicity.Scale(1 / d.HMultiplicity.Integral())
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
	tp, err := hplot.NewTiledPlot(draw.Tiles{Cols: 1, Rows: 2, PadY: 1 * vg.Centimeter})
	if err != nil {
		panic(err)
	}

	p1 := tp.Plot(0, 0)
	p1.X.Min = 0
	p1.X.Max = 240
	p1.Y.Min = 0
	p1.X.Label.Text = "channel"
	p1.Y.Label.Text = "No pulses"
	p1.X.Tick.Marker = &hplot.FreqTicks{N: 241, Freq: 4}
	p1.Add(hplot.NewGrid())
	hplotfreq, err := hplot.NewH1D(d.HFrequency)
	if err != nil {
		panic(err)
	}
	hplotfreq.FillColor = color.RGBA{R: 255, G: 204, B: 153, A: 255}
	hplotfreq.Color = plotutil.Color(3)
	p1.Add(hplotfreq)
	if err != nil {
		panic(err)
	}
	//p1.Title.Text = fmt.Sprintf("Number of pulses vs channel\n")

	p2 := tp.Plot(1, 0)
	p2.X.Min = 0
	p2.X.Max = 240
	p2.Y.Min = 0
	p2.X.Label.Text = "channel"
	p2.Y.Label.Text = "No sat. pulses"
	p2.X.Tick.Marker = &hplot.FreqTicks{N: 241, Freq: 4}
	p2.Add(hplot.NewGrid())
	hplotsatfreq, err := hplot.NewH1D(d.HSatFrequency)
	if err != nil {
		panic(err)
	}
	hplotsatfreq.FillColor = color.RGBA{R: 255, G: 204, B: 153, A: 255}
	hplotsatfreq.Color = plotutil.Color(3)
	p2.Add(hplotsatfreq)
	if err != nil {
		log.Fatalf("error creating histogram \n")
	}
	//p2.Title.Text = fmt.Sprintf("Number of saturating pulses vs channel\n")
	return tp
}

func (d *DQPlot) MakeMinRec1DTiledPlot() *hplot.TiledPlot {
	tp, err := hplot.NewTiledPlot(draw.Tiles{Cols: 1, Rows: 3, PadY: 0.2 * vg.Centimeter})
	if err != nil {
		panic(err)
	}

	p1 := tp.Plot(0, 0)
	p1.X.Min = -50
	p1.X.Max = 50
	p1.X.Label.Text = "X (mm)"
	p1.Y.Label.Text = "No entries"
	p1.X.Tick.Marker = &hplot.FreqTicks{N: 101, Freq: 5}
	p1.Add(hplot.NewGrid())
	hplotX, err := hplot.NewH1D(d.HMinRecX)
	if err != nil {
		panic(err)
	}
	hplotX.FillColor = color.RGBA{R: 255, G: 204, B: 153, A: 255}
	hplotX.Color = plotutil.Color(3)
	p1.Add(hplotX)
	if err != nil {
		panic(err)
	}
	//p1.Title.Text = fmt.Sprintf("Distribution of minimal reconstruction X (mm)\n")

	p2 := tp.Plot(1, 0)
	p2.X.Min = -60
	p2.X.Max = 60
	p2.X.Label.Text = "Y (mm)"
	p2.Y.Label.Text = "No entries"
	p2.X.Tick.Marker = &hplot.FreqTicks{N: 121, Freq: 5}
	p2.Add(hplot.NewGrid())
	hplotY, err := hplot.NewH1D(d.HMinRecY)
	if err != nil {
		panic(err)
	}
	hplotY.FillColor = color.RGBA{R: 255, G: 204, B: 153, A: 255}
	hplotY.Color = plotutil.Color(3)
	p2.Add(hplotY)
	if err != nil {
		log.Fatalf("error creating histogram \n")
	}
	//p2.Title.Text = fmt.Sprintf("Distribution of minimal reconstruction Y (mm)\n")

	p3 := tp.Plot(2, 0)
	p3.X.Min = -150
	p3.X.Max = 150
	p3.X.Label.Text = "Z (mm)"
	p3.Y.Label.Text = "No entries"
	p3.X.Tick.Marker = &hplot.FreqTicks{N: 151, Freq: 10}
	p3.Add(hplot.NewGrid())
	hplotZ, err := hplot.NewH1D(d.HMinRecZ)
	if err != nil {
		panic(err)
	}
	hplotZ.FillColor = color.RGBA{R: 255, G: 204, B: 153, A: 255}
	hplotZ.Color = plotutil.Color(3)
	p3.Add(hplotZ)
	if err != nil {
		log.Fatalf("error creating histogram \n")
	}
	//p3.Title.Text = fmt.Sprintf("Distribution of minimal reconstruction Z (mm)\n")

	return tp
}

type WhichVar byte

const (
	Charge WhichVar = iota
	Amplitude
	Energy
)

func (d *DQPlot) MakeChargeAmplTiledPlot(whichV WhichVar, whichH dpgadetector.HemisphereType) *hplot.TiledPlot {
	tp, err := hplot.NewTiledPlot(draw.Tiles{Cols: 5, Rows: 6, PadY: 0.5 * vg.Centimeter})
	if err != nil {
		panic(err)
	}

	histos := make([]hbook.H1D, len(d.HCharge[0]))
	var histosref []hbook.H1D
	if d.DQPlotRef != nil {
		histosref = make([]hbook.H1D, len(d.HCharge[0]))
	}

	iCluster := 0
	irowBeg := 0
	irowEnd := tp.Tiles.Rows
	icolBeg := tp.Tiles.Cols - 1
	icolEnd := -1

	if whichH == dpgadetector.Left {
		iCluster = 30
		irowBeg = tp.Tiles.Rows - 1
		irowEnd = -1
		icolBeg = 0
		icolEnd = tp.Tiles.Cols
	}

	for irow := irowBeg; irow != irowEnd; {
		for icol := icolBeg; icol != icolEnd; {
			switch whichV {
			case Charge:
				histos = d.HCharge[iCluster]
				if len(histosref) != 0 {
					histosref = d.DQPlotRef.HCharge[iCluster]
				}
			case Amplitude:
				histos = d.HAmplitude[iCluster]
				if len(histosref) != 0 {
					histosref = d.DQPlotRef.HAmplitude[iCluster]
				}
			case Energy:
				histos = d.HEnergy[iCluster]
				if len(histosref) != 0 {
					histosref = d.DQPlotRef.HEnergy[iCluster]
				}
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
			hplot0, err := hplot.NewH1D(&histos[0])
			if err != nil {
				panic(err)
			}
			hplot1, err := hplot.NewH1D(&histos[1])
			if err != nil {
				panic(err)
			}
			hplot2, err := hplot.NewH1D(&histos[2])
			if err != nil {
				panic(err)
			}
			hplot3, err := hplot.NewH1D(&histos[3])
			if err != nil {
				panic(err)
			}
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
			if len(histosref) != 0 {
				if histos[0].Integral() != 0 && histos[1].Integral() != 0 && histos[2].Integral() != 0 && histos[3].Integral() != 0 {
					histosref[0].Scale(histos[0].Integral() / histosref[0].Integral())
					histosref[1].Scale(histos[1].Integral() / histosref[1].Integral())
					histosref[2].Scale(histos[2].Integral() / histosref[2].Integral())
					histosref[3].Scale(histos[3].Integral() / histosref[3].Integral())
				}

				hplot0ref, err := hplot.NewH1D(&histosref[0])
				if err != nil {
					panic(err)
				}
				hplot1ref, err := hplot.NewH1D(&histosref[1])
				if err != nil {
					panic(err)
				}
				hplot2ref, err := hplot.NewH1D(&histosref[2])
				if err != nil {
					panic(err)
				}
				hplot3ref, err := hplot.NewH1D(&histosref[3])
				if err != nil {
					panic(err)
				}
				hplot0ref.Color = color.RGBA{R: 238, G: 46, B: 47, A: 255}  // red
				hplot1ref.Color = color.RGBA{R: 0, G: 140, B: 72, A: 255}   // green
				hplot2ref.Color = color.RGBA{R: 24, G: 90, B: 169, A: 255}  // blue
				hplot3ref.Color = color.RGBA{R: 250, G: 88, B: 244, A: 255} // pink
				hplot0ref.FillColor = color.NRGBA{R: 238, G: 46, B: 47, A: 80}
				hplot1ref.FillColor = color.NRGBA{R: 0, G: 140, B: 72, A: 80}
				hplot2ref.FillColor = color.NRGBA{R: 24, G: 90, B: 169, A: 80}
				hplot3ref.FillColor = color.NRGBA{R: 250, G: 88, B: 244, A: 80}
				hplot0ref.LineStyle.Width = 2
				hplot1ref.LineStyle.Width = 2
				hplot2ref.LineStyle.Width = 2
				hplot3ref.LineStyle.Width = 2
				hplot0ref.LineStyle.Dashes = []vg.Length{vg.Points(5), vg.Points(1)}
				hplot1ref.LineStyle.Dashes = []vg.Length{vg.Points(5), vg.Points(1)}
				hplot2ref.LineStyle.Dashes = []vg.Length{vg.Points(5), vg.Points(1)}
				hplot3ref.LineStyle.Dashes = []vg.Length{vg.Points(5), vg.Points(1)}
				p.Add(hplot0ref, hplot1ref, hplot2ref, hplot3ref)
			}
			iCluster++
			switch whichH {
			case dpgadetector.Left:
				p.BackgroundColor = color.RGBA{R: 224, G: 242, B: 247, A: 255} // blue
				icol++
			case dpgadetector.Right:
				p.BackgroundColor = color.RGBA{R: 255, G: 255, B: 0, A: 255} // yellow
				icol--
			}
		}
		switch whichH {
		case dpgadetector.Left:
			irow--
		case dpgadetector.Right:
			irow++
		}
	}
	return tp
}

func (d *DQPlot) MakeHVTiledPlot() *hplot.TiledPlot {
	tp, err := hplot.NewTiledPlot(draw.Tiles{Cols: 2, Rows: 6, PadX: 3.5 * vg.Centimeter, PadY: 1 * vg.Centimeter})
	if err != nil {
		panic(err)
	}
	//var TextsAndXYs []interface{}

	//color := color.RGBA{R: 224, G: 242, B: 247, A: 255}

	var irow int
	var icol int
	for iHV := uint(0); iHV < 60; iHV++ {
		if iHV >= 0 && iHV <= 4 { // right hemisphere of DPGA
			irow = 0
			icol = 1
		} else if iHV >= 10 && iHV <= 14 {
			irow = 1
			icol = 1
		} else if iHV >= 20 && iHV <= 24 {
			irow = 2
			icol = 1
		} else if iHV >= 30 && iHV <= 34 {
			irow = 3
			icol = 1
		} else if iHV >= 40 && iHV <= 44 {
			irow = 4
			icol = 1
		} else if iHV >= 50 && iHV <= 54 {
			irow = 5
			icol = 1
		} else if iHV >= 5 && iHV <= 9 { // left hemisphere of DPGA
			irow = 0
			icol = 0
		} else if iHV >= 15 && iHV <= 19 {
			irow = 1
			icol = 0
		} else if iHV >= 25 && iHV <= 29 {
			irow = 2
			icol = 0
		} else if iHV >= 35 && iHV <= 39 {
			irow = 3
			icol = 0
		} else if iHV >= 45 && iHV <= 49 {
			irow = 4
			icol = 0
		} else if iHV >= 55 && iHV <= 59 {
			irow = 5
			icol = 0
		}

		p := tp.Plot(irow, icol)
		grid := hplot.NewGrid()
		grid.Vertical.Width = 0
		grid.Horizontal.Dashes = plotutil.Dashes(1)
		p.Add(grid)
		//p.Y.Min = 800
		//p.Y.Max = 1200
		p.X.Label.Text = "event"
		p.Y.Label.Text = "HV"
		//p.Add(plot.NewGrid())
		p.Plot.X.LineStyle.Width = 2
		p.Plot.Y.LineStyle.Width = 2
		p.Plot.X.Tick.LineStyle.Width = 2
		p.Plot.Y.Tick.LineStyle.Width = 2
		//TextsAndXYs = append(TextsAndXYs, "toto")
		hvserialchan := dpgadetector.HVmap[iHV]
		ser := hvserialchan.SerialNumber
		ch := hvserialchan.ChannelNumber

		// test
		var ps []plot.Plotter
		l, s, err := plotter.NewLinePoints(d.HV[ser-1][ch])
		iii := int(iHV % 5)
		l.Color = plotutil.Color(iii)
		//l.Dashes = plotutil.Dashes(iii)
		l.LineStyle.Width = 0.1 * vg.Centimeter
		s.Color = plotutil.Color(iii)
		s.Shape = draw.CircleGlyph{} //plotutil.Shape(iii)
		s.GlyphStyle.Radius = 0.1 * vg.Centimeter
		ps = append(ps, l, s)
		p.Add(ps...)
		p.Legend.Add("HV"+strconv.FormatUint(uint64(iHV), 10)+"("+strconv.FormatUint(uint64(ser), 10)+", "+strconv.FormatUint(uint64(ch), 10)+") ", l, s)
		p.Legend.XOffs = 3.2 * vg.Centimeter
		// end test

		//err = plotutil.AddLinePoints(&p.Plot, "HV"+strconv.FormatUint(uint64(iHV), 10)+" ("+strconv.FormatUint(uint64(ser), 10)+", "+strconv.FormatUint(uint64(ch), 10)+") ", d.HV[ser-1][ch])
		if err != nil {
			panic(err)
		}
	}
	return tp
}

func (d *DQPlot) MakeAmplCorrelationPlot() *plot.Plot {
	pAmplCorrelation, err := plot.New()
	if err != nil {
		panic(err)
	}
	pAmplCorrelation.X.Label.Text = "Amplitude pulse 0 (ADC counts)"
	pAmplCorrelation.Y.Label.Text = "Amplitude pulse 1 (ADC counts)"
	pAmplCorrelation.X.Tick.Marker = &hplot.FreqTicks{N: 11, Freq: 2}
	pAmplCorrelation.Y.Tick.Marker = &hplot.FreqTicks{N: 11, Freq: 2}
	pAmplCorrelation.X.Min = d.AmplCorrelation.XMin()
	pAmplCorrelation.Y.Min = d.AmplCorrelation.YMin()
	pAmplCorrelation.X.Max = d.AmplCorrelation.XMax()
	pAmplCorrelation.Y.Max = d.AmplCorrelation.YMax()
	pAmplCorrelation.Add(hplot.NewH2D(d.AmplCorrelation, nil))
	//pAmplCorrelation.Add(plotter.NewGrid())
	return pAmplCorrelation
}

func (d *DQPlot) MakeHitQuartetsPlot() *plot.Plot {
	pHitQuartets, err := plot.New()
	if err != nil {
		panic(err)
	}
	pHitQuartets.X.Label.Text = "Quartet Id (right hemisphere)"
	pHitQuartets.Y.Label.Text = "Quartet Id (left hemisphere)"
	pHitQuartets.X.Tick.Marker = &hplot.FreqTicks{N: 31, Freq: 1}
	pHitQuartets.Y.Tick.Marker = &hplot.FreqTicks{N: 31, Freq: 1}
	pHitQuartets.X.Min = d.HitQuartets.XMin()
	pHitQuartets.Y.Min = d.HitQuartets.YMin()
	pHitQuartets.X.Max = d.HitQuartets.XMax()
	pHitQuartets.Y.Max = d.HitQuartets.YMax()
	pHitQuartets.Add(hplot.NewH2D(d.HitQuartets, nil))
	//pHitQuartets.Add(plotter.NewGrid())
	return pHitQuartets
}

// SaveHistos saves histograms on disk.
// If d.DQPlotRef is not nil, current histograms
// are overlaid with the provided reference histograms.
func (d *DQPlot) SaveHistos() {
	doplot := utils.MakeHPl
	// 	doplot := utils.MakeGonumPlot

	dqplotref := &DQPlot{}

	if d.DQPlotRef != nil {
		dqplotref = d.DQPlotRef
	}

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
