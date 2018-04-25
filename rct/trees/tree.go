package trees

import (
	"github.com/go-hep/croot"
	//"gitlab.in2p3.fr/avirm/analysis-go/dpga/dpgadetector"
	"gitlab.in2p3.fr/avirm/analysis-go/dpga/dpgadetector"
	"gitlab.in2p3.fr/avirm/analysis-go/event"
	"gitlab.in2p3.fr/avirm/analysis-go/pulse"
	"gitlab.in2p3.fr/avirm/analysis-go/utils"
)

const NoPulsesMax = 24
const NoSamplesMax = 500

type ROOTData struct {
	Run uint32
	Evt uint32
	T0  uint32

	NoPulses            int32
	IChanAbs240         [NoPulsesMax]uint16
	IQuartetAbs60       [NoPulsesMax]uint8
	ILineAbs12          [NoPulsesMax]uint8
	IHemi               [NoPulsesMax]uint8
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
	SRout               [NoPulsesMax]uint16

	NoSamples     int32
	SampleTimes   [NoSamplesMax]float64
	SampleIndices [NoSamplesMax]uint16
	Pulse         [NoPulsesMax][NoSamplesMax]float64
	CapaId        [NoPulsesMax][NoSamplesMax]uint16
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

	_, err = t.tree.Branch2("NoPulses", &t.data.NoPulses, "NoPulses/I", bufsiz)
	_, err = t.tree.Branch2("IChanAbs240", &t.data.IChanAbs240, "IChanAbs240[NoPulses]/s", bufsiz)
	_, err = t.tree.Branch2("IQuartetAbs60", &t.data.IQuartetAbs60, "IQuartetAbs60[NoPulses]/b", bufsiz)
	_, err = t.tree.Branch2("ILineAbs12", &t.data.ILineAbs12, "ILineAbs12[NoPulses]/b", bufsiz)
	_, err = t.tree.Branch2("IHemi", &t.data.IHemi, "IHemi[NoPulses]/b", bufsiz)
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
	_, err = t.tree.Branch2("SRout", &t.data.SRout, "SRout[NoPulses]/s", bufsiz)

	_, err = t.tree.Branch2("NoSamples", &t.data.NoSamples, "NoSamples/I", bufsiz)
	_, err = t.tree.Branch2("SampleTimes", &t.data.SampleTimes, "SampleTimes[NoSamples]/D", bufsiz)
	_, err = t.tree.Branch2("SampleIndices", &t.data.SampleIndices, "SampleIndices[NoSamples]/s", bufsiz)
	_, err = t.tree.Branch2("Pulse", &t.data.Pulse, "Pulse[NoPulses][1023]/D", bufsiz)
	_, err = t.tree.Branch2("CapaId", &t.data.CapaId, "CapaId[NoPulses][1023]/s", bufsiz)
	// 	_, err = t.tree.Branch2("Pulse", &t.data.Pulse, "Pulse[NoPulses][848]/D", bufsiz)
	// 	_, err = t.tree.Branch2("CapaId", &t.data.CapaId, "CapaId[NoPulses][848]/s", bufsiz)

	return &t
}

func (t *Tree) Fill(run uint32, event *event.Event) {
	t.data.Run = run
	t.data.Evt = uint32(event.ID)
	t.data.T0 = 0

	noPulses, pulses, _, noPulsesWoData, pulsesWoData, _ := event.Multiplicity()
	t.data.NoPulses = int32(noPulses + noPulsesWoData)
	// 	fmt.Println("noPulses=", t.data.NoPulses)
	if noPulses > 0 {
		t.data.NoSamples = int32(len(pulses[0].Samples))
		// 		fmt.Println("NoSamples=", t.data.NoSamples)
	}

	pulsesFull := make([]*pulse.Pulse, 0)
	pulsesFull = append(pulsesFull, pulses...)
	pulsesFull = append(pulsesFull, pulsesWoData...)

	// 	fmt.Println("debugging: ", len(pulses), len(pulsesWoData), len(pulsesFull))

	for i := range pulsesFull {
		pulse := pulsesFull[i]
		pulse.CalcRisingFront(true)
		pulse.CalcFallingFront(false)
		// 		fmt.Println("i=", i)
		t.data.IChanAbs240[i] = uint16(pulse.Channel.AbsID240())
		t.data.IQuartetAbs60[i] = dpgadetector.FifoID144ToQuartetAbsIdx60(pulse.Channel.FifoID144(), true)
		t.data.ILineAbs12[i] = dpgadetector.QuartetAbsIdx60ToLineAbsIdx12(t.data.IQuartetAbs60[i])
		t.data.IHemi[i] = uint8(pulse.Hemi())
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
		t.data.SRout[i] = pulse.SRout

		for j := range pulse.Samples {
			if i == 0 {
				t.data.SampleTimes[j] = pulse.Samples[j].Time
				t.data.SampleIndices[j] = pulse.Samples[j].Index
			}
			t.data.Pulse[i][j] = pulse.Samples[j].Amplitude
			t.data.CapaId[i][j] = pulse.Samples[j].Capacitor.ID()
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
