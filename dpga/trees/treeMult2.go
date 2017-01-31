package trees

import (
	"github.com/go-hep/croot"
	//"gitlab.in2p3.fr/avirm/analysis-go/dpga/dpgadetector"
	"gitlab.in2p3.fr/avirm/analysis-go/dpga/dpgadetector"
	"gitlab.in2p3.fr/avirm/analysis-go/pulse"
	"gitlab.in2p3.fr/avirm/analysis-go/utils"
)

type ROOTDataMult2 struct {
	Run         uint32
	Evt         uint32
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
	// 	TestSize            uint32
	// 	TestSlice           []float64
	IChanAbs240         [2]uint16
	IQuartetAbs60       [2]uint8
	ILineAbs12          [2]uint8
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
	X                   [2]float64
	Y                   [2]float64
	Z                   [2]float64
	Xmaa                float64
	Ymaa                float64
	Zmaa                float64
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
	// 	_, err = t.tree.Branch2("TestSize", &t.data.TestSize, "TestSize/i", bufsiz)
	// 	_, err = t.tree.Branch("TestSlice", &t.data.TestSlice, bufsiz, 0)
	_, err = t.tree.Branch2("IChanAbs240", &t.data.IChanAbs240, "IChanAbs240[2]/s", bufsiz)
	_, err = t.tree.Branch2("IQuartetAbs60", &t.data.IQuartetAbs60, "IQuartetAbs60[2]/b", bufsiz)
	_, err = t.tree.Branch2("ILineAbs12", &t.data.ILineAbs12, "ILineAbs12[2]/b", bufsiz)
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
	_, err = t.tree.Branch2("X", &t.data.X, "X[2]/D", bufsiz)
	_, err = t.tree.Branch2("Y", &t.data.Y, "Y[2]/D", bufsiz)
	_, err = t.tree.Branch2("Z", &t.data.Z, "Z[2]/D", bufsiz)
	_, err = t.tree.Branch2("Xmaa", &t.data.Xmaa, "Xmaa/D", bufsiz)
	_, err = t.tree.Branch2("Ymaa", &t.data.Ymaa, "Ymaa/D", bufsiz)
	_, err = t.tree.Branch2("Zmaa", &t.data.Zmaa, "Zmaa/D", bufsiz)
	_, err = t.tree.Branch2("TRF", &t.data.TRF, "TRF/D", bufsiz)

	//t.data.Pulse[0] = make([]float64, dpgadetector.Det.NoSamples())
	//t.data.Pulse[1] = make([]float64, dpgadetector.Det.NoSamples())
	return &t
}

func (t *TreeMult2) Fill(run uint32, ievent uint32, counters []uint32, pulse0 *pulse.Pulse, pulse1 *pulse.Pulse, Xmaa, Ymaa, Zmaa float64) {
	t.data.Run = run
	t.data.Evt = ievent
	t.data.TimeStamp = uint64(counters[3])<<32 | uint64(counters[2])
	if counters[0] != 0 {
		t.data.RateBoard1 = float64(counters[4]) * 64e6 / float64(counters[0])
		t.data.RateBoard2 = float64(counters[5]) * 64e6 / float64(counters[0])
		t.data.RateBoard3 = float64(counters[6]) * 64e6 / float64(counters[0])
		t.data.RateBoard4 = float64(counters[7]) * 64e6 / float64(counters[0])
		t.data.RateBoard5 = float64(counters[8]) * 64e6 / float64(counters[0])
		t.data.RateBoard6 = float64(counters[9]) * 64e6 / float64(counters[0])
		t.data.RateBoard7 = float64(counters[10]) * 64e6 / float64(counters[0])
		t.data.RateBoard8 = float64(counters[11]) * 64e6 / float64(counters[0])
		t.data.RateBoard9 = float64(counters[12]) * 64e6 / float64(counters[0])
		t.data.RateBoard10 = float64(counters[13]) * 64e6 / float64(counters[0])
		t.data.RateBoard11 = float64(counters[14]) * 64e6 / float64(counters[0])
		t.data.RateBoard12 = float64(counters[15]) * 64e6 / float64(counters[0])
		t.data.RateLvsR1 = float64(counters[16]) * 64e6 / float64(counters[0])
		t.data.RateLvsR2 = float64(counters[17]) * 64e6 / float64(counters[0])
		t.data.RateLvsR3 = float64(counters[18]) * 64e6 / float64(counters[0])
		t.data.RateLvsR4 = float64(counters[19]) * 64e6 / float64(counters[0])
		t.data.RateLvsR5 = float64(counters[20]) * 64e6 / float64(counters[0])
		t.data.RateLvsR6 = float64(counters[21]) * 64e6 / float64(counters[0])
		t.data.RateLvsR7 = float64(counters[22]) * 64e6 / float64(counters[0])
		t.data.RateLvs3L1 = float64(counters[30]) * 64e6 / float64(counters[0])
		t.data.RateLvs3L2 = float64(counters[31]) * 64e6 / float64(counters[0])
		t.data.RateLvs3L3 = float64(counters[32]) * 64e6 / float64(counters[0])
		t.data.RateLvs3L4 = float64(counters[33]) * 64e6 / float64(counters[0])
		t.data.RateLvs3L5 = float64(counters[34]) * 64e6 / float64(counters[0])
		t.data.RateLvs3L6 = float64(counters[35]) * 64e6 / float64(counters[0])
		t.data.RateLvs3L7 = float64(counters[36]) * 64e6 / float64(counters[0])
		t.data.RateLvsL1 = float64(counters[23]) * 64e6 / float64(counters[0])
		t.data.RateLvsL2 = float64(counters[24]) * 64e6 / float64(counters[0])
		t.data.RateLvsL3 = float64(counters[25]) * 64e6 / float64(counters[0])
		t.data.RateLvsL4 = float64(counters[26]) * 64e6 / float64(counters[0])
		t.data.RateLvsL5 = float64(counters[27]) * 64e6 / float64(counters[0])
		t.data.RateLvsL6 = float64(counters[28]) * 64e6 / float64(counters[0])
		t.data.RateLvsL7 = float64(counters[29]) * 64e6 / float64(counters[0])
	}
	// 	t.data.TestSize = 1
	// 	t.data.TestSlice = make([]float64, t.data.TestSize)
	// 	t.data.TestSlice[0] = 13
	t.data.IChanAbs240[0] = uint16(pulse0.Channel.AbsID240())
	t.data.IChanAbs240[1] = uint16(pulse1.Channel.AbsID240())
	t.data.IQuartetAbs60[0] = dpgadetector.FifoID144ToQuartetAbsIdx60(pulse0.Channel.FifoID144(), true)
	t.data.IQuartetAbs60[1] = dpgadetector.FifoID144ToQuartetAbsIdx60(pulse1.Channel.FifoID144(), true)
	t.data.ILineAbs12[0] = dpgadetector.QuartetAbsIdx60ToLineAbsIdx12(t.data.IQuartetAbs60[0])
	t.data.ILineAbs12[1] = dpgadetector.QuartetAbsIdx60ToLineAbsIdx12(t.data.IQuartetAbs60[1])
	t.data.E[0] = pulse0.E
	t.data.E[1] = pulse1.E
	t.data.Ampl[0] = pulse0.Ampl
	t.data.Ampl[1] = pulse1.Ampl
	t.data.Sat[0] = utils.BoolToUint8(pulse0.HasSatSignal)
	t.data.Sat[1] = utils.BoolToUint8(pulse1.HasSatSignal)
	t.data.Charge[0] = pulse0.Charg
	t.data.Charge[1] = pulse1.Charg
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
	}
	t.data.X[0] = pulse0.Channel.X
	t.data.Y[0] = pulse0.Channel.Y
	t.data.Z[0] = pulse0.Channel.Z
	t.data.X[1] = pulse1.Channel.X
	t.data.Y[1] = pulse1.Channel.Y
	t.data.Z[1] = pulse1.Channel.Z
	t.data.Xmaa = Xmaa
	t.data.Ymaa = Ymaa
	t.data.Zmaa = Zmaa
	_, err := t.tree.Fill()
	if err != nil {
		panic(err)
	}
}

func (t *TreeMult2) Close() {
	t.file.Write("", 0, 0)
	t.file.Close("")
}
