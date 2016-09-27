package trees

import (
	"github.com/go-hep/croot"
	"gitlab.in2p3.fr/avirm/analysis-go/pulse"
)

type ROOTDataMult2 struct {
	IChanAbs240 [2]uint16
	Ampl        [2]float64
	Charge      [2]float64
	T30         [2]float64
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
	_, err = t.tree.Branch("data", &t.data, bufsiz, 0)
	if err != nil {
		panic(err)
	}
	return &t
}

func (t *TreeMult2) Fill(pulse0 *pulse.Pulse, pulse1 *pulse.Pulse) {
	t.data.IChanAbs240[0] = uint16(pulse0.Channel.AbsID240())
	t.data.IChanAbs240[1] = uint16(pulse1.Channel.AbsID240())
	t.data.Ampl[0] = pulse0.Ampl
	t.data.Ampl[1] = pulse1.Ampl
	t.data.Charge[0] = pulse0.Charg
	t.data.Charge[1] = pulse1.Charg
	t.data.T30[0] = pulse0.Time30
	t.data.T30[1] = pulse1.Time30
	_, err := t.tree.Fill()
	if err != nil {
		panic(err)
	}
}

func (t *TreeMult2) Close() {
	t.file.Write("", 0, 0)
	t.file.Close("")
}