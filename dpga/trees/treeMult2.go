package trees

import (
	"math"

	"github.com/go-hep/croot"
	//"gitlab.in2p3.fr/avirm/analysis-go/dpga/dpgadetector"
	"gitlab.in2p3.fr/avirm/analysis-go/dpga/dpgadetector"
	"gitlab.in2p3.fr/avirm/analysis-go/dpga/rw"
	"gitlab.in2p3.fr/avirm/analysis-go/event"
	"gitlab.in2p3.fr/avirm/analysis-go/pulse"
	"gitlab.in2p3.fr/avirm/analysis-go/reconstruction"
	"gitlab.in2p3.fr/avirm/analysis-go/utils"
)

type ROOTDataMult2 struct {
	Run         uint32
	Evt         uint32
	T0          uint32
	TimeStamp   uint64
	RateBoard1  float64
	RateBoard2  float64
	RateBoard3  float64
	RateBoard4  float64
	RateBoard5  float64
	RateBoard6  float64
	RateBoard7  float64
	RateBoard8  float64
	RateBoard9  float64
	RateBoard10 float64
	RateBoard11 float64
	RateBoard12 float64
	RateLvsR1   float64
	RateLvsR2   float64
	RateLvsR3   float64
	RateLvsR4   float64
	RateLvsR5   float64
	RateLvsR6   float64
	RateLvsR7   float64
	RateLvs3L1  float64
	RateLvs3L2  float64
	RateLvs3L3  float64
	RateLvs3L4  float64
	RateLvs3L5  float64
	RateLvs3L6  float64
	RateLvs3L7  float64
	RateLvsL1   float64
	RateLvsL2   float64
	RateLvsL3   float64
	RateLvsL4   float64
	RateLvsL5   float64
	RateLvsL6   float64
	RateLvsL7   float64

	NoPulses            int32
	IChanAbs240         [2]uint16
	IQuartetAbs60       [2]uint8
	ILineAbs12          [2]uint8
	IHemi               [2]uint8
	E                   [2]float64
	Ampl                [2]float64
	Sat                 [2]uint8
	Charge              [2]float64
	T10                 [2]float64
	T20                 [2]float64
	T30                 [2]float64
	T80                 [2]float64
	T90                 [2]float64
	Tf20                [2]float64
	NoLocMaxRisingFront [2]uint16
	SampleTimes         [999]float64
	Pulse               [2][999]float64
	PulseRF             [999]float64
	X                   [2]float64
	Y                   [2]float64
	Z                   [2]float64
	Xc                  [2]float64
	Yc                  [2]float64
	Zc                  [2]float64
	Xmar                float64
	Ymar                float64
	Zmar                float64
	Rmar                float64
	TRF                 float64
}

type TreeMult2 struct {
	data ROOTDataMult2
	file croot.File
	tree croot.Tree
}

func NewTreeMult2(outrootfileName string) *TreeMult2 {
	f, err := croot.OpenFile(outrootfileName, "recreate", "ROOT file with event information", 1, 0)
	if err != nil {
		panic(err)
	}
	t := TreeMult2{file: f, tree: croot.NewTree("tree", "tree", 32)}
	const bufsiz = 32000
	/*
		_, err = t.tree.Branch("data", &t.data, bufsiz, 0)
		if err != nil {
			panic(err)
		}
	*/

	_, err = t.tree.Branch2("Run", &t.data.Run, "Run/i", bufsiz)
	_, err = t.tree.Branch2("Evt", &t.data.Evt, "Evt/i", bufsiz)
	_, err = t.tree.Branch2("T0", &t.data.T0, "T0/i", bufsiz)
	_, err = t.tree.Branch2("TimeStamp", &t.data.TimeStamp, "TimeStamp/l", bufsiz)
	_, err = t.tree.Branch2("RateBoard1", &t.data.RateBoard1, "RateBoard1/D", bufsiz)
	_, err = t.tree.Branch2("RateBoard2", &t.data.RateBoard2, "RateBoard2/D", bufsiz)
	_, err = t.tree.Branch2("RateBoard3", &t.data.RateBoard3, "RateBoard3/D", bufsiz)
	_, err = t.tree.Branch2("RateBoard4", &t.data.RateBoard4, "RateBoard4/D", bufsiz)
	_, err = t.tree.Branch2("RateBoard5", &t.data.RateBoard5, "RateBoard5/D", bufsiz)
	_, err = t.tree.Branch2("RateBoard6", &t.data.RateBoard6, "RateBoard6/D", bufsiz)
	_, err = t.tree.Branch2("RateBoard7", &t.data.RateBoard7, "RateBoard7/D", bufsiz)
	_, err = t.tree.Branch2("RateBoard8", &t.data.RateBoard8, "RateBoard8/D", bufsiz)
	_, err = t.tree.Branch2("RateBoard9", &t.data.RateBoard9, "RateBoard9/D", bufsiz)
	_, err = t.tree.Branch2("RateBoard10", &t.data.RateBoard10, "RateBoard10/D", bufsiz)
	_, err = t.tree.Branch2("RateBoard11", &t.data.RateBoard11, "RateBoard11/D", bufsiz)
	_, err = t.tree.Branch2("RateBoard12", &t.data.RateBoard12, "RateBoard12/D", bufsiz)
	_, err = t.tree.Branch2("RateLvsR1", &t.data.RateLvsR1, "RateLvsR1/D", bufsiz)
	_, err = t.tree.Branch2("RateLvsR2", &t.data.RateLvsR2, "RateLvsR2/D", bufsiz)
	_, err = t.tree.Branch2("RateLvsR3", &t.data.RateLvsR3, "RateLvsR3/D", bufsiz)
	_, err = t.tree.Branch2("RateLvsR4", &t.data.RateLvsR4, "RateLvsR4/D", bufsiz)
	_, err = t.tree.Branch2("RateLvsR5", &t.data.RateLvsR5, "RateLvsR5/D", bufsiz)
	_, err = t.tree.Branch2("RateLvsR6", &t.data.RateLvsR6, "RateLvsR6/D", bufsiz)
	_, err = t.tree.Branch2("RateLvsR7", &t.data.RateLvsR7, "RateLvsR7/D", bufsiz)
	_, err = t.tree.Branch2("RateLvs3L1", &t.data.RateLvs3L1, "RateLvs3L1/D", bufsiz)
	_, err = t.tree.Branch2("RateLvs3L2", &t.data.RateLvs3L2, "RateLvs3L2/D", bufsiz)
	_, err = t.tree.Branch2("RateLvs3L3", &t.data.RateLvs3L3, "RateLvs3L3/D", bufsiz)
	_, err = t.tree.Branch2("RateLvs3L4", &t.data.RateLvs3L4, "RateLvs3L4/D", bufsiz)
	_, err = t.tree.Branch2("RateLvs3L5", &t.data.RateLvs3L5, "RateLvs3L5/D", bufsiz)
	_, err = t.tree.Branch2("RateLvs3L6", &t.data.RateLvs3L6, "RateLvs3L6/D", bufsiz)
	_, err = t.tree.Branch2("RateLvs3L7", &t.data.RateLvs3L7, "RateLvs3L7/D", bufsiz)
	_, err = t.tree.Branch2("RateLvsL1", &t.data.RateLvsL1, "RateLvsL1/D", bufsiz)
	_, err = t.tree.Branch2("RateLvsL2", &t.data.RateLvsL2, "RateLvsL2/D", bufsiz)
	_, err = t.tree.Branch2("RateLvsL3", &t.data.RateLvsL3, "RateLvsL3/D", bufsiz)
	_, err = t.tree.Branch2("RateLvsL4", &t.data.RateLvsL4, "RateLvsL4/D", bufsiz)
	_, err = t.tree.Branch2("RateLvsL5", &t.data.RateLvsL5, "RateLvsL5/D", bufsiz)
	_, err = t.tree.Branch2("RateLvsL6", &t.data.RateLvsL6, "RateLvsL6/D", bufsiz)
	_, err = t.tree.Branch2("RateLvsL7", &t.data.RateLvsL7, "RateLvsL7/D", bufsiz)

	_, err = t.tree.Branch2("NoPulses", &t.data.NoPulses, "NoPulses/I", bufsiz)
	_, err = t.tree.Branch2("IChanAbs240", &t.data.IChanAbs240, "IChanAbs240[2]/s", bufsiz)
	_, err = t.tree.Branch2("IQuartetAbs60", &t.data.IQuartetAbs60, "IQuartetAbs60[2]/b", bufsiz)
	_, err = t.tree.Branch2("ILineAbs12", &t.data.ILineAbs12, "ILineAbs12[2]/b", bufsiz)
	_, err = t.tree.Branch2("IHemi", &t.data.IHemi, "IHemi[2]/b", bufsiz)
	_, err = t.tree.Branch2("E", &t.data.E, "E[2]/D", bufsiz)
	_, err = t.tree.Branch2("Ampl", &t.data.Ampl, "Ampl[2]/D", bufsiz)
	_, err = t.tree.Branch2("Sat", &t.data.Sat, "Sat[2]/b", bufsiz)
	_, err = t.tree.Branch2("Charge", &t.data.Charge, "Charge[2]/D", bufsiz)
	_, err = t.tree.Branch2("T10", &t.data.T10, "T10[2]/D", bufsiz)
	_, err = t.tree.Branch2("T20", &t.data.T20, "T20[2]/D", bufsiz)
	_, err = t.tree.Branch2("T30", &t.data.T30, "T30[2]/D", bufsiz)
	_, err = t.tree.Branch2("T80", &t.data.T80, "T80[2]/D", bufsiz)
	_, err = t.tree.Branch2("T90", &t.data.T90, "T90[2]/D", bufsiz)
	_, err = t.tree.Branch2("Tf20", &t.data.Tf20, "Tf20[2]/D", bufsiz)
	_, err = t.tree.Branch2("NoLocMaxRisingFront", &t.data.NoLocMaxRisingFront, "NoLocMaxRisingFront[2]/s", bufsiz)
	_, err = t.tree.Branch2("SampleTimes", &t.data.SampleTimes, "SampleTimes[999]/D", bufsiz)
	_, err = t.tree.Branch2("Pulse", &t.data.Pulse, "Pulse[2][999]/D", bufsiz)
	_, err = t.tree.Branch2("PulseRF", &t.data.PulseRF, "PulseRF[999]/D", bufsiz)
	_, err = t.tree.Branch2("X", &t.data.X, "X[2]/D", bufsiz)
	_, err = t.tree.Branch2("Y", &t.data.Y, "Y[2]/D", bufsiz)
	_, err = t.tree.Branch2("Z", &t.data.Z, "Z[2]/D", bufsiz)
	_, err = t.tree.Branch2("Xc", &t.data.Xc, "Xc[2]/D", bufsiz)
	_, err = t.tree.Branch2("Yc", &t.data.Yc, "Yc[2]/D", bufsiz)
	_, err = t.tree.Branch2("Zc", &t.data.Zc, "Zc[2]/D", bufsiz)
	_, err = t.tree.Branch2("Xmar", &t.data.Xmar, "Xmar/D", bufsiz)
	_, err = t.tree.Branch2("Ymar", &t.data.Ymar, "Ymar/D", bufsiz)
	_, err = t.tree.Branch2("Zmar", &t.data.Zmar, "Zmar/D", bufsiz)
	_, err = t.tree.Branch2("Rmar", &t.data.Rmar, "Rmar/D", bufsiz)
	_, err = t.tree.Branch2("TRF", &t.data.TRF, "TRF/D", bufsiz)

	//t.data.Pulse[0] = make([]float64, dpgadetector.Det.NoSamples())
	//t.data.Pulse[1] = make([]float64, dpgadetector.Det.NoSamples())
	return &t
}

func (t *TreeMult2) Fill(run uint32, hdr *rw.Header, event *event.Event, pulse0 *pulse.Pulse, pulse1 *pulse.Pulse) {
	t.data.Run = run
	t.data.Evt = uint32(event.ID)
	t.data.TimeStamp = uint64(event.Counters[3])<<32 | uint64(event.Counters[2])
	t.data.T0 = hdr.TimeStart
	if event.Counters[0] != 0 {
		t.data.RateBoard1 = float64(event.Counters[4]) * 64e6 / float64(event.Counters[0])
		t.data.RateBoard2 = float64(event.Counters[5]) * 64e6 / float64(event.Counters[0])
		t.data.RateBoard3 = float64(event.Counters[6]) * 64e6 / float64(event.Counters[0])
		t.data.RateBoard4 = float64(event.Counters[7]) * 64e6 / float64(event.Counters[0])
		t.data.RateBoard5 = float64(event.Counters[8]) * 64e6 / float64(event.Counters[0])
		t.data.RateBoard6 = float64(event.Counters[9]) * 64e6 / float64(event.Counters[0])
		t.data.RateBoard7 = float64(event.Counters[10]) * 64e6 / float64(event.Counters[0])
		t.data.RateBoard8 = float64(event.Counters[11]) * 64e6 / float64(event.Counters[0])
		t.data.RateBoard9 = float64(event.Counters[12]) * 64e6 / float64(event.Counters[0])
		t.data.RateBoard10 = float64(event.Counters[13]) * 64e6 / float64(event.Counters[0])
		t.data.RateBoard11 = float64(event.Counters[14]) * 64e6 / float64(event.Counters[0])
		t.data.RateBoard12 = float64(event.Counters[15]) * 64e6 / float64(event.Counters[0])
		t.data.RateLvsR1 = float64(event.Counters[16]) * 64e6 / float64(event.Counters[0])
		t.data.RateLvsR2 = float64(event.Counters[17]) * 64e6 / float64(event.Counters[0])
		t.data.RateLvsR3 = float64(event.Counters[18]) * 64e6 / float64(event.Counters[0])
		t.data.RateLvsR4 = float64(event.Counters[19]) * 64e6 / float64(event.Counters[0])
		t.data.RateLvsR5 = float64(event.Counters[20]) * 64e6 / float64(event.Counters[0])
		t.data.RateLvsR6 = float64(event.Counters[21]) * 64e6 / float64(event.Counters[0])
		t.data.RateLvsR7 = float64(event.Counters[22]) * 64e6 / float64(event.Counters[0])
		t.data.RateLvs3L1 = float64(event.Counters[30]) * 64e6 / float64(event.Counters[0])
		t.data.RateLvs3L2 = float64(event.Counters[31]) * 64e6 / float64(event.Counters[0])
		t.data.RateLvs3L3 = float64(event.Counters[32]) * 64e6 / float64(event.Counters[0])
		t.data.RateLvs3L4 = float64(event.Counters[33]) * 64e6 / float64(event.Counters[0])
		t.data.RateLvs3L5 = float64(event.Counters[34]) * 64e6 / float64(event.Counters[0])
		t.data.RateLvs3L6 = float64(event.Counters[35]) * 64e6 / float64(event.Counters[0])
		t.data.RateLvs3L7 = float64(event.Counters[36]) * 64e6 / float64(event.Counters[0])
		t.data.RateLvsL1 = float64(event.Counters[23]) * 64e6 / float64(event.Counters[0])
		t.data.RateLvsL2 = float64(event.Counters[24]) * 64e6 / float64(event.Counters[0])
		t.data.RateLvsL3 = float64(event.Counters[25]) * 64e6 / float64(event.Counters[0])
		t.data.RateLvsL4 = float64(event.Counters[26]) * 64e6 / float64(event.Counters[0])
		t.data.RateLvsL5 = float64(event.Counters[27]) * 64e6 / float64(event.Counters[0])
		t.data.RateLvsL6 = float64(event.Counters[28]) * 64e6 / float64(event.Counters[0])
		t.data.RateLvsL7 = float64(event.Counters[29]) * 64e6 / float64(event.Counters[0])
	}
	t.data.NoPulses = 2
	t.data.IChanAbs240[0] = uint16(pulse0.Channel.AbsID240())
	t.data.IChanAbs240[1] = uint16(pulse1.Channel.AbsID240())
	t.data.IQuartetAbs60[0] = dpgadetector.FifoID144ToQuartetAbsIdx60(pulse0.Channel.FifoID144(), true)
	t.data.IQuartetAbs60[1] = dpgadetector.FifoID144ToQuartetAbsIdx60(pulse1.Channel.FifoID144(), true)
	t.data.ILineAbs12[0] = dpgadetector.QuartetAbsIdx60ToLineAbsIdx12(t.data.IQuartetAbs60[0])
	t.data.ILineAbs12[1] = dpgadetector.QuartetAbsIdx60ToLineAbsIdx12(t.data.IQuartetAbs60[1])
	t.data.IHemi[0] = uint8(pulse0.Hemi())
	t.data.IHemi[1] = uint8(pulse1.Hemi())
	t.data.E[0] = pulse0.E
	t.data.E[1] = pulse1.E
	t.data.Ampl[0] = pulse0.Ampl
	t.data.Ampl[1] = pulse1.Ampl
	t.data.Sat[0] = utils.BoolToUint8(pulse0.HasSatSignal)
	t.data.Sat[1] = utils.BoolToUint8(pulse1.HasSatSignal)
	t.data.Charge[0] = pulse0.Charg
	t.data.Charge[1] = pulse1.Charg
	pulse0.CalcRisingFront(true)
	pulse0.CalcFallingFront(false)
	pulse1.CalcRisingFront(true)
	pulse1.CalcFallingFront(false)
	t.data.T10[0] = pulse0.Time10
	t.data.T10[1] = pulse1.Time10
	t.data.T20[0] = pulse0.Time20
	t.data.T20[1] = pulse1.Time20
	t.data.T30[0] = pulse0.Time30
	t.data.T30[1] = pulse1.Time30
	t.data.T80[0] = pulse0.Time80
	t.data.T80[1] = pulse1.Time80
	t.data.T90[0] = pulse0.Time90
	t.data.T90[1] = pulse1.Time90
	t.data.Tf20[0] = pulse0.TimeFall20
	t.data.Tf20[1] = pulse1.TimeFall20
	t.data.NoLocMaxRisingFront[0] = uint16(pulse0.NoLocMaxRisingFront)
	t.data.NoLocMaxRisingFront[1] = uint16(pulse1.NoLocMaxRisingFront)
	for i := range pulse0.Samples {
		t.data.SampleTimes[i] = pulse0.Samples[i].Time
		t.data.Pulse[0][i] = pulse0.Samples[i].Amplitude
		t.data.Pulse[1][i] = pulse1.Samples[i].Amplitude
		if len(event.ClustersWoData[0].Pulses[0].Samples) > 0 {
			t.data.PulseRF[i] = event.ClustersWoData[0].Pulses[0].Samples[i].Amplitude
		}
	}
	t.data.X[0] = pulse0.Channel.X
	t.data.Y[0] = pulse0.Channel.Y
	t.data.Z[0] = pulse0.Channel.Z
	t.data.X[1] = pulse1.Channel.X
	t.data.Y[1] = pulse1.Channel.Y
	t.data.Z[1] = pulse1.Channel.Z
	t.data.Xc[0] = pulse0.Channel.CrystCenter.X
	t.data.Yc[0] = pulse0.Channel.CrystCenter.Y
	t.data.Zc[0] = pulse0.Channel.CrystCenter.Z
	t.data.Xc[1] = pulse1.Channel.CrystCenter.X
	t.data.Yc[1] = pulse1.Channel.CrystCenter.Y
	t.data.Zc[1] = pulse1.Channel.CrystCenter.Z
	// 	fmt.Println(pulse0.Channel.ScintCoords)

	doMinRec := true
	if hdr.TriggerEq == 3 {
		// In case TriggerEq = 3 (pulser), one has to check that the two pulses are
		// on different hemispheres, otherwise the minimal reconstruction is not well
		// defined
		if pulse.SameHemi(pulse0, pulse1) {
			doMinRec = false
		}
	}
	if doMinRec {
		xbeam, ybeam := 0., 0.
		ch0 := pulse0.Channel
		ch1 := pulse1.Channel
		x, y, z := reconstruction.Minimal(true, ch0, ch1, xbeam, ybeam)
		t.data.Xmar = x
		t.data.Ymar = y
		t.data.Zmar = z
		t.data.Rmar = math.Sqrt(x*x + y*y)
	}

	///////////////////////////////////////////
	// TRF calculation
	// 	for i := range event.ClustersWoData {
	// 		cluster := &event.ClustersWoData[i]
	// 		for j := range cluster.Pulses {
	// 			pulse := &cluster.Pulses[j]
	// 			fmt.Println(i, j, len(pulse.Samples))
	// 		}
	// 	}
	/*
		ampSlice := event.ClustersWoData[0].Pulses[0].MakeAmpSlice()
		if len(ampSlice) != 0 { // can compute TRF
			timesRF := utils.FindIntersections(event.ID, event.ClustersWoData[0].Pulses[0].MakeAmpSlice(), event.ClustersWoData[0].Pulses[0].MakeTimeSlice())
			tMean := (pulse0.Time30 + pulse1.Time30) / 2.
			if tMean < timesRF[0] {
				t.data.TRF = timesRF[0] - 1/24.85e6*1e9 // 24.85 MHz is the HF frequency
			} else if tMean > timesRF[len(timesRF)-1] {
				t.data.TRF = timesRF[len(timesRF)-1]
			} else {
				for i := range timesRF {
					if i < len(timesRF)-1 {
						if tMean > timesRF[i] && tMean < timesRF[i+1] {
							t.data.TRF = timesRF[i]
							break
						}
					} else {
						fmt.Println(timesRF)
						log.Fatalf("This should not happen, tMean=%v\n", tMean)
					}
				}
			}
			if tMean-t.data.TRF > 1/24.85e6*1e9+3 {
				log.Fatalf("ERROR\n")
			}
		}
	*/
	///////////////////////////////////////////

	_, err := t.tree.Fill()
	if err != nil {
		panic(err)
	}
}

func (t *TreeMult2) Close() {
	t.file.Write("", 0, 0)
	t.file.Close("")
}
