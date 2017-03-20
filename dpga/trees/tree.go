package trees

import (
	"github.com/go-hep/croot"
	//"gitlab.in2p3.fr/avirm/analysis-go/dpga/dpgadetector"
	"gitlab.in2p3.fr/avirm/analysis-go/dpga/dpgadetector"
	"gitlab.in2p3.fr/avirm/analysis-go/dpga/rw"
	"gitlab.in2p3.fr/avirm/analysis-go/event"
	"gitlab.in2p3.fr/avirm/analysis-go/pulse"
	"gitlab.in2p3.fr/avirm/analysis-go/reconstruction"
	"gitlab.in2p3.fr/avirm/analysis-go/utils"
)

const NoPulsesMax = 240

type ROOTData struct {
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

	NoPulses  int32
	TestSlice [NoPulsesMax]float64

	IChanAbs240         [NoPulsesMax]uint16
	IQuartetAbs60       [NoPulsesMax]uint8
	ILineAbs12          [NoPulsesMax]uint8
	E                   [NoPulsesMax]float64
	Ampl                [NoPulsesMax]float64
	Sat                 [NoPulsesMax]uint8
	Charge              [NoPulsesMax]float64
	T10                 [NoPulsesMax]float64
	T20                 [NoPulsesMax]float64
	T30                 [NoPulsesMax]float64
	T80                 [NoPulsesMax]float64
	T90                 [NoPulsesMax]float64
	Tf20                [NoPulsesMax]float64
	NoLocMaxRisingFront [NoPulsesMax]uint16
	SampleTimes         [999]float64
	Pulse               [NoPulsesMax][999]float64
	X                   [NoPulsesMax]float64
	Y                   [NoPulsesMax]float64
	Z                   [NoPulsesMax]float64
	Xmaa                float64
	Ymaa                float64
	Zmaa                float64
	TRF                 float64
}

type Tree struct {
	data ROOTData
	file croot.File
	tree croot.Tree
}

func NewTree(outrootfileName string) *Tree {
	f, err := croot.OpenFile(outrootfileName, "recreate", "ROOT file with event information", 1, 0)
	if err != nil {
		panic(err)
	}
	t := Tree{file: f, tree: croot.NewTree("tree", "tree", 32)}
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

	// 	_, err = t.tree.Branch2("TestSize", &t.data.TestSize, "TestSize/L", bufsiz)
	// 	_, err = t.tree.Branch2("TestSlice", &t.data.TestSlice, "TestSlice[TestSize]/D", bufsiz)
	_, err = t.tree.Branch2("NoPulses", &t.data.NoPulses, "NoPulses/I", bufsiz)
	_, err = t.tree.Branch2("IChanAbs240", &t.data.IChanAbs240, "IChanAbs240[NoPulses]/s", bufsiz)
	_, err = t.tree.Branch2("IQuartetAbs60", &t.data.IQuartetAbs60, "IQuartetAbs60[NoPulses]/b", bufsiz)
	_, err = t.tree.Branch2("ILineAbs12", &t.data.ILineAbs12, "ILineAbs12[NoPulses]/b", bufsiz)
	_, err = t.tree.Branch2("E", &t.data.E, "E[NoPulses]/D", bufsiz)
	_, err = t.tree.Branch2("Ampl", &t.data.Ampl, "Ampl[NoPulses]/D", bufsiz)
	_, err = t.tree.Branch2("Sat", &t.data.Sat, "Sat[NoPulses]/b", bufsiz)
	_, err = t.tree.Branch2("Charge", &t.data.Charge, "Charge[NoPulses]/D", bufsiz)
	_, err = t.tree.Branch2("T10", &t.data.T10, "T10[NoPulses]/D", bufsiz)
	_, err = t.tree.Branch2("T20", &t.data.T20, "T20[NoPulses]/D", bufsiz)
	_, err = t.tree.Branch2("T30", &t.data.T30, "T30[NoPulses]/D", bufsiz)
	_, err = t.tree.Branch2("T80", &t.data.T80, "T80[NoPulses]/D", bufsiz)
	_, err = t.tree.Branch2("T90", &t.data.T90, "T90[NoPulses]/D", bufsiz)
	_, err = t.tree.Branch2("Tf20", &t.data.Tf20, "Tf20[NoPulses]/D", bufsiz)
	_, err = t.tree.Branch2("NoLocMaxRisingFront", &t.data.NoLocMaxRisingFront, "NoLocMaxRisingFront[NoPulses]/s", bufsiz)
	_, err = t.tree.Branch2("SampleTimes", &t.data.SampleTimes, "SampleTimes[999]/D", bufsiz)
	_, err = t.tree.Branch2("Pulse", &t.data.Pulse, "Pulse[NoPulses][999]/D", bufsiz)
	_, err = t.tree.Branch2("X", &t.data.X, "X[NoPulses]/D", bufsiz)
	_, err = t.tree.Branch2("Y", &t.data.Y, "Y[NoPulses]/D", bufsiz)
	_, err = t.tree.Branch2("Z", &t.data.Z, "Z[NoPulses]/D", bufsiz)
	_, err = t.tree.Branch2("Xmaa", &t.data.Xmaa, "Xmaa/D", bufsiz)
	_, err = t.tree.Branch2("Ymaa", &t.data.Ymaa, "Ymaa/D", bufsiz)
	_, err = t.tree.Branch2("Zmaa", &t.data.Zmaa, "Zmaa/D", bufsiz)
	_, err = t.tree.Branch2("TRF", &t.data.TRF, "TRF/D", bufsiz)

	//t.data.Pulse[0] = make([]float64, dpgadetector.Det.NoSamples())
	//t.data.Pulse[1] = make([]float64, dpgadetector.Det.NoSamples())
	return &t
}

func (t *Tree) Fill(run uint32, hdr *rw.Header, event *event.Event) {
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
	///////////////////////////////////////////
	// TRF calculation
	//utils.FindIntersections(event.ClustersWoData[11].Pulses[3], min, max)
	///////////////////////////////////////////
	noPulses, pulses := event.Multiplicity()
	t.data.NoPulses = int32(noPulses)
	for i := range pulses {
		pulse := pulses[i]
		pulse.CalcRisingFront(true)
		pulse.CalcFallingFront(false)
		// 		fmt.Println("i=", i)
		t.data.IChanAbs240[i] = uint16(pulse.Channel.AbsID240())
		t.data.IQuartetAbs60[i] = dpgadetector.FifoID144ToQuartetAbsIdx60(pulse.Channel.FifoID144(), true)
		t.data.ILineAbs12[i] = dpgadetector.QuartetAbsIdx60ToLineAbsIdx12(t.data.IQuartetAbs60[i])
		t.data.E[i] = pulse.E
		t.data.Ampl[i] = pulse.Ampl
		t.data.Sat[i] = utils.BoolToUint8(pulse.HasSatSignal)
		t.data.Charge[i] = pulse.Charg
		t.data.T10[i] = pulse.Time10
		t.data.T20[i] = pulse.Time20
		t.data.T30[i] = pulse.Time30
		t.data.T80[i] = pulse.Time80
		t.data.T90[i] = pulse.Time90
		t.data.Tf20[i] = pulse.TimeFall20
		t.data.NoLocMaxRisingFront[i] = uint16(pulse.NoLocMaxRisingFront)
		for j := range pulse.Samples {
			t.data.SampleTimes[j] = pulse.Samples[j].Time
			t.data.Pulse[i][j] = pulse.Samples[j].Amplitude
		}
		t.data.X[i] = pulse.Channel.X
		t.data.Y[i] = pulse.Channel.Y
		t.data.Z[i] = pulse.Channel.Z
	}
	if noPulses == 2 {
		if len(pulses) != 2 {
			panic("mult == 2 but len(pulsesWithSignal) != 2: this should NEVER happen !")
		}
		doMinRec := true
		if hdr.TriggerEq == 3 {
			// In case TriggerEq = 3 (pulser), one has to check that the two pulses are
			// on different hemispheres, otherwise the minimal reconstruction is not well
			// defined
			if pulse.SameHemi(pulses[0], pulses[1]) {
				doMinRec = false
			}
		}
		if doMinRec {
			xbeam, ybeam := 0., 0.
			ch0 := pulses[0].Channel
			ch1 := pulses[1].Channel
			x, y, z := reconstruction.Minimal(ch0, ch1, xbeam, ybeam)
			t.data.Xmaa = x
			t.data.Ymaa = y
			t.data.Zmaa = z
		}
	}
	_, err := t.tree.Fill()
	if err != nil {
		panic(err)
	}
}

func (t *Tree) Close() {
	t.file.Write("", 0, 0)
	t.file.Close("")
}
