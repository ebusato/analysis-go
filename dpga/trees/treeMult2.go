package trees

import (
	"github.com/go-hep/croot"
	//"gitlab.in2p3.fr/avirm/analysis-go/dpga/dpgadetector"
	"gitlab.in2p3.fr/avirm/analysis-go/dpga/dpgadetector"
	"gitlab.in2p3.fr/avirm/analysis-go/pulse"
)

type ROOTDataMult2 struct {
	Run                 uint32
	Evt                 uint32
	IChanAbs240         [2]uint16
	IQuartetAbs60       [2]uint8
	E                   [2]float64
	Ampl                [2]float64
	Charge              [2]float64
	T10                 [2]float64
	T20                 [2]float64
	T30                 [2]float64
	T80                 [2]float64
	T90                 [2]float64
	NoLocMaxRisingFront [2]uint16
	SampleTimes         [999]float64
	Pulse               [2][999]float64
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
	_, err = t.tree.Branch2("IChanAbs240", &t.data.IChanAbs240, "IChanAbs240[2]/s", bufsiz)
	_, err = t.tree.Branch2("IQuartetAbs60", &t.data.IQuartetAbs60, "IQuartetAbs60[2]/b", bufsiz)
	_, err = t.tree.Branch2("E", &t.data.E, "E[2]/D", bufsiz)
	_, err = t.tree.Branch2("Ampl", &t.data.Ampl, "Ampl[2]/D", bufsiz)
	_, err = t.tree.Branch2("Charge", &t.data.Charge, "Charge[2]/D", bufsiz)
	_, err = t.tree.Branch2("T10", &t.data.T10, "T10[2]/D", bufsiz)
	_, err = t.tree.Branch2("T20", &t.data.T20, "T20[2]/D", bufsiz)
	_, err = t.tree.Branch2("T30", &t.data.T30, "T30[2]/D", bufsiz)
	_, err = t.tree.Branch2("T80", &t.data.T80, "T80[2]/D", bufsiz)
	_, err = t.tree.Branch2("T90", &t.data.T90, "T90[2]/D", bufsiz)
	_, err = t.tree.Branch2("NoLocMaxRisingFront", &t.data.NoLocMaxRisingFront, "NoLocMaxRisingFront[2]/s", bufsiz)
	_, err = t.tree.Branch2("SampleTimes", &t.data.SampleTimes, "SampleTimes[999]/D", bufsiz)
	_, err = t.tree.Branch2("Pulse", &t.data.Pulse, "Pulse[2][999]/D", bufsiz)

	//t.data.Pulse[0] = make([]float64, dpgadetector.Det.NoSamples())
	//t.data.Pulse[1] = make([]float64, dpgadetector.Det.NoSamples())
	return &t
}

func (t *TreeMult2) Fill(run uint32, ievent uint32, pulse0 *pulse.Pulse, pulse1 *pulse.Pulse) {
	t.data.Run = run
	t.data.Evt = ievent
	t.data.IChanAbs240[0] = uint16(pulse0.Channel.AbsID240())
	t.data.IChanAbs240[1] = uint16(pulse1.Channel.AbsID240())
	t.data.IQuartetAbs60[0] = dpgadetector.FifoID144ToQuartetAbsIdx60(pulse0.Channel.FifoID144(), true)
	t.data.IQuartetAbs60[1] = dpgadetector.FifoID144ToQuartetAbsIdx60(pulse1.Channel.FifoID144(), true)
	t.data.E[0] = pulse0.E
	t.data.E[1] = pulse1.E
	t.data.Ampl[0] = pulse0.Ampl
	t.data.Ampl[1] = pulse1.Ampl
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
	t.data.NoLocMaxRisingFront[0] = uint16(pulse0.NoLocMaxRisingFront)
	t.data.NoLocMaxRisingFront[1] = uint16(pulse1.NoLocMaxRisingFront)
	for i := range pulse0.Samples {
		t.data.SampleTimes[i] = pulse0.Samples[i].Time
		t.data.Pulse[0][i] = pulse0.Samples[i].Amplitude
		t.data.Pulse[1][i] = pulse1.Samples[i].Amplitude
	}
	_, err := t.tree.Fill()
	if err != nil {
		panic(err)
	}
}

func (t *TreeMult2) Close() {
	t.file.Write("", 0, 0)
	t.file.Close("")
}
