package trees

import (
	"fmt"
	"log"

	"github.com/go-hep/croot"
	//"gitlab.in2p3.fr/avirm/analysis-go/dpga/dpgadetector"
	"gitlab.in2p3.fr/avirm/analysis-go/dpga/dpgadetector"
	"gitlab.in2p3.fr/avirm/analysis-go/dpga/rw"
	"gitlab.in2p3.fr/avirm/analysis-go/event"
	"gitlab.in2p3.fr/avirm/analysis-go/pulse"
	"gitlab.in2p3.fr/avirm/analysis-go/utils"
)

const NoPulsesInLORMax = 2 * NoLORsMax

type ROOTDataLOR struct {
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

	NoPulses int32

	IChanAbs240         [NoPulsesInLORMax]uint16
	IQuartetAbs60       [NoPulsesInLORMax]uint8
	ILineAbs12          [NoPulsesInLORMax]uint8
	IHemi               [NoPulsesInLORMax]uint8
	E                   [NoPulsesInLORMax]float64
	Ampl                [NoPulsesInLORMax]float64
	Sat                 [NoPulsesInLORMax]uint8
	Charge              [NoPulsesInLORMax]float64
	T10                 [NoPulsesInLORMax]float64
	T20                 [NoPulsesInLORMax]float64
	T30                 [NoPulsesInLORMax]float64
	T80                 [NoPulsesInLORMax]float64
	T90                 [NoPulsesInLORMax]float64
	Tf20                [NoPulsesInLORMax]float64
	NoLocMaxRisingFront [NoPulsesInLORMax]uint16
	SampleTimes         [999]float64
	Pulse               [NoPulsesInLORMax][999]float64
	PulseRF             [999]float64
	X                   [NoPulsesInLORMax]float64
	Y                   [NoPulsesInLORMax]float64
	Z                   [NoPulsesInLORMax]float64
	Xc                  [NoPulsesInLORMax]float64
	Yc                  [NoPulsesInLORMax]float64
	Zc                  [NoPulsesInLORMax]float64
	NoLORs              int32
	LORIdx1             [NoLORsMax]int32
	LORIdx2             [NoLORsMax]int32
	LORTMean            [NoLORsMax]float64
	LORTRF              [NoLORsMax]float64
	LORXmar             [NoLORsMax]float64
	LORYmar             [NoLORsMax]float64
	LORZmar             [NoLORsMax]float64
	LORRmar             [NoLORsMax]float64
}

type TreeLOR struct {
	data ROOTDataLOR
	file croot.File
	tree croot.Tree
}

func NewTreeLOR(outrootfileName string) *TreeLOR {
	f, err := croot.OpenFile(outrootfileName, "recreate", "ROOT file with event information", 1, 0)
	if err != nil {
		panic(err)
	}
	t := TreeLOR{file: f, tree: croot.NewTree("tree", "tree", 32)}
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
	_, err = t.tree.Branch2("SampleTimes", &t.data.SampleTimes, "SampleTimes[999]/D", bufsiz)
	_, err = t.tree.Branch2("Pulse", &t.data.Pulse, "Pulse[NoPulses][999]/D", bufsiz)
	_, err = t.tree.Branch2("PulseRF", &t.data.PulseRF, "PulseRF[999]/D", bufsiz)
	_, err = t.tree.Branch2("X", &t.data.X, "X[NoPulses]/D", bufsiz)
	_, err = t.tree.Branch2("Y", &t.data.Y, "Y[NoPulses]/D", bufsiz)
	_, err = t.tree.Branch2("Z", &t.data.Z, "Z[NoPulses]/D", bufsiz)
	_, err = t.tree.Branch2("Xc", &t.data.Xc, "Xc[NoPulses]/D", bufsiz)
	_, err = t.tree.Branch2("Yc", &t.data.Yc, "Yc[NoPulses]/D", bufsiz)
	_, err = t.tree.Branch2("Zc", &t.data.Zc, "Zc[NoPulses]/D", bufsiz)
	_, err = t.tree.Branch2("NoLORs", &t.data.NoLORs, "NoLORs/I", bufsiz)
	_, err = t.tree.Branch2("LORIdx1", &t.data.LORIdx1, "LORIdx1[NoLORs]/I", bufsiz)
	_, err = t.tree.Branch2("LORIdx2", &t.data.LORIdx2, "LORIdx2[NoLORs]/I", bufsiz)
	_, err = t.tree.Branch2("LORTMean", &t.data.LORTMean, "LORTMean[NoLORs]/D", bufsiz)
	_, err = t.tree.Branch2("LORTRF", &t.data.LORTRF, "LORTRF[NoLORs]/D", bufsiz)
	_, err = t.tree.Branch2("LORXmar", &t.data.LORXmar, "LORXmar[NoLORs]/D", bufsiz)
	_, err = t.tree.Branch2("LORYmar", &t.data.LORYmar, "LORYmar[NoLORs]/D", bufsiz)
	_, err = t.tree.Branch2("LORZmar", &t.data.LORZmar, "LORZmar[NoLORs]/D", bufsiz)
	_, err = t.tree.Branch2("LORRmar", &t.data.LORRmar, "LORRmar[NoLORs]/D", bufsiz)

	return &t
}

func AlreadyIn(pulses []*pulse.Pulse, p *pulse.Pulse) (bool, int) {
	for i := range pulses {
		if pulses[i] == p {
			return true, i
		}
	}
	return false, -1
}

func (t *TreeLOR) Fill(run uint32, hdr *rw.Header, event *event.Event) {
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

	// TRF stuff
	ampSlice := event.ClustersWoData[0].Pulses[0].MakeAmpSlice()
	var timesRF []float64
	if len(ampSlice) != 0 { // can compute TRF
		timesRF = utils.FindIntersections(event.ID, event.ClustersWoData[0].Pulses[0].MakeAmpSlice(), event.ClustersWoData[0].Pulses[0].MakeTimeSlice())
	}

	event.FindLORs(0, 0, 38., 3*1.2, 0, 1000) //511+3*28.3)
	// 	fmt.Println("no lors: ", len(event.LORs))
	t.data.NoLORs = int32(len(event.LORs))
	if t.data.NoLORs < NoLORsMax {
		var pulsesInLOR []*pulse.Pulse
		for i := range event.LORs {
			lor := &event.LORs[i]

			In1, Idx1 := AlreadyIn(pulsesInLOR, lor.Pulses[0])
			In2, Idx2 := AlreadyIn(pulsesInLOR, lor.Pulses[1])
			if !In1 {
				pulsesInLOR = append(pulsesInLOR, lor.Pulses[0])
				Idx1 = len(pulsesInLOR) - 1
			}
			if !In2 {
				pulsesInLOR = append(pulsesInLOR, lor.Pulses[1])
				Idx2 = len(pulsesInLOR) - 1
			}

			t.data.LORIdx1[i] = int32(Idx1)
			t.data.LORIdx2[i] = int32(Idx2)
			t.data.LORTMean[i] = lor.TMean
			t.data.LORXmar[i] = lor.Xmar
			t.data.LORYmar[i] = lor.Ymar
			t.data.LORZmar[i] = lor.Zmar
			t.data.LORRmar[i] = lor.Rmar

			///////////////////////////////////////////
			// TRF calculation
			// 	for i := range event.ClustersWoData {
			// 		cluster := &event.ClustersWoData[i]
			// 		for j := range cluster.Pulses {
			// 			pulse := &cluster.Pulses[j]
			// 			fmt.Println(i, j, len(pulse.Samples))
			// 		}
			// 	}
			if len(timesRF) > 0 {
				if t.data.LORTMean[i] <= timesRF[0] {
					//fmt.Println("here ", t.data.LORTMean[i], timesRF[0])
					t.data.LORTRF[i] = timesRF[0] - 1/24.85e6*1e9 // 24.85 MHz is the HF frequency
				} else if t.data.LORTMean[i] >= timesRF[len(timesRF)-1] {
					t.data.LORTRF[i] = timesRF[len(timesRF)-1]
				} else {
					for j := range timesRF {
						if j < len(timesRF)-1 {
							if t.data.LORTMean[i] > timesRF[j] && t.data.LORTMean[i] < timesRF[j+1] {
								t.data.LORTRF[i] = timesRF[j]
								break
							}
						} else {
							fmt.Println(timesRF)
							log.Fatalf("This should not happen, tMean=%v\n", t.data.LORTMean[i])
						}
					}
				}
				if t.data.LORTMean[i]-t.data.LORTRF[i] > 1/24.85e6*1e9+3 {
					fmt.Println(timesRF)
					fmt.Println(t.data.LORTMean[i], t.data.LORTRF[i])
					log.Fatalf("ERROR\n")
				}
			}
			///////////////////////////////////////////
		}

		t.data.NoPulses = int32(len(pulsesInLOR))
		for i := range pulsesInLOR {
			pulse := pulsesInLOR[i]
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
			for j := range pulse.Samples {
				t.data.SampleTimes[j] = pulse.Samples[j].Time
				t.data.Pulse[i][j] = pulse.Samples[j].Amplitude
				if len(event.ClustersWoData[0].Pulses[0].Samples) > 0 {
					t.data.PulseRF[j] = event.ClustersWoData[0].Pulses[0].Samples[j].Amplitude
				}
			}
			t.data.X[i] = pulse.Channel.X
			t.data.Y[i] = pulse.Channel.Y
			t.data.Z[i] = pulse.Channel.Z
			t.data.Xc[i] = pulse.Channel.CrystCenter.X
			t.data.Yc[i] = pulse.Channel.CrystCenter.Y
			t.data.Zc[i] = pulse.Channel.CrystCenter.Z
		}
	}

	_, err := t.tree.Fill()
	if err != nil {
		panic(err)
	}
}

func (t *TreeLOR) Close() {
	t.file.Write("", 0, 0)
	t.file.Close("")
}
