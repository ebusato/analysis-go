package event

import (
	"fmt"
	"log"
	"math"

	"gitlab.in2p3.fr/avirm/analysis-go/dpga/dpgadetector"
	"gitlab.in2p3.fr/avirm/analysis-go/pulse"
	"gitlab.in2p3.fr/avirm/analysis-go/reconstruction"
	"gitlab.in2p3.fr/avirm/analysis-go/utils"
)

type LOR struct {
	Pulses [2]*pulse.Pulse
	Idx1   int
	Idx2   int
	TMean  float64
	TRF    float64
	Xmar   float64
	Ymar   float64
	Zmar   float64
	Rmar   float64
}

func NewLOR(pulse1, pulse2 *pulse.Pulse, idx1, idx2 int, Xmar, Ymar, Zmar, Rmar float64) *LOR {
	l := &LOR{}
	l.Pulses[0] = pulse1
	l.Pulses[1] = pulse2
	l.Idx1 = idx1
	l.Idx2 = idx2
	l.TMean = (pulse1.Time30 + pulse2.Time30) / 2.
	l.TRF = -1
	l.Xmar = Xmar
	l.Ymar = Ymar
	l.Zmar = Zmar
	l.Rmar = Rmar
	return l
}

func (l *LOR) Copy() *LOR {
	return NewLOR(l.Pulses[0], l.Pulses[1], l.Idx1, l.Idx2, l.Xmar, l.Ymar, l.Zmar, l.Rmar)
}

func (l *LOR) CalcTRF(timesRF []float64) {
	if len(timesRF) == 0 {
		//log.Printf("len(timesRF) == 0 -> setting TRF to 0.\n")
		return
	}
	if l.TMean <= timesRF[0] {
		//fmt.Println("here ", t.data.LORTMean[i], timesRF[0])
		l.TRF = timesRF[0] - 1/24.85e6*1e9 // 24.85 MHz is the HF frequency
	} else if l.TMean >= timesRF[len(timesRF)-1] {
		l.TRF = timesRF[len(timesRF)-1]
	} else {
		for j := range timesRF {
			if j < len(timesRF)-1 {
				if l.TMean > timesRF[j] && l.TMean < timesRF[j+1] {
					l.TRF = timesRF[j]
					break
				}
			} else {
				fmt.Println(timesRF)
				log.Printf("This should not happen, tMean=%v\n", l.TMean)
			}
		}
	}
	if l.TMean-l.TRF > 1/24.85e6*1e9+3 {
		fmt.Println(timesRF)
		fmt.Println(l.TMean, l.TRF)
		log.Printf("ERROR\n")
	}
}

type Event struct {
	Clusters        []pulse.Cluster
	ClustersWoData  [12]pulse.Cluster // These are the 12 clusters corresponding to the 12*4 channels unused for data at the end of each ASM board
	ID              uint
	TimeStamp       uint64
	Counters        []uint32
	LORs            []LOR
	IsCorrupted     bool
	UDPPayloadSizes []int // number of octets for each frame making this event (FrameSize has NoFrames components)
}

func NewEvent(noClusters uint8) *Event {
	return &Event{
		Clusters:    make([]pulse.Cluster, noClusters),
		ID:          0,
		TimeStamp:   0,
		IsCorrupted: false,
	}
}

func (e *Event) Copy() *Event {
	newevent := NewEvent(uint8(e.NoClusters()))
	newevent.ID = e.ID
	newevent.TimeStamp = e.TimeStamp
	newevent.UDPPayloadSizes = e.UDPPayloadSizes
	newevent.Counters = make([]uint32, len(e.Counters))
	for i := range e.Counters {
		newevent.Counters[i] = e.Counters[i]
	}
	newevent.IsCorrupted = e.IsCorrupted
	for i := range e.Clusters {
		oldPulses := e.Clusters[i].Pulses
		newevent.Clusters[i].Pulses = [4]pulse.Pulse{
			*oldPulses[0].Copy(),
			*oldPulses[1].Copy(),
			*oldPulses[2].Copy(),
			*oldPulses[3].Copy()}
		newevent.Clusters[i].ID = e.Clusters[i].ID
	}
	for i := range e.ClustersWoData {
		oldPulses := e.ClustersWoData[i].Pulses
		newevent.ClustersWoData[i].Pulses = [4]pulse.Pulse{
			*oldPulses[0].Copy(),
			*oldPulses[1].Copy(),
			*oldPulses[2].Copy(),
			*oldPulses[3].Copy()}
		newevent.ClustersWoData[i].ID = e.ClustersWoData[i].ID
	}
	return newevent
}

func (e *Event) CheckIntegrity() {
	for i := range e.Clusters {
		cluster := &e.Clusters[i]
		noSamples := e.Clusters[i].NoSamples()
		for j := 1; j < len(e.Clusters[i].Pulses); j++ {
			n := cluster.NoSamples()
			if n != noSamples {
				log.Fatal("not all pulses have the same number of samples")
			}
			pulse := &cluster.Pulses[j]
			if pulse.Channel == nil {
				log.Fatal("pulse has not associated channel")
			}
		}
	}
}

func (e *Event) NoClusters() int {
	return len(e.Clusters)
}

func (e *Event) Print(printClusters bool, printClusterDetails bool) {
	fmt.Println("\n-> Printing event", e.ID)
	fmt.Println("    o Number of clusters =", len(e.Clusters))
	fmt.Println("    o ID =", e.ID)
	fmt.Println("    o TimeStamp =", e.TimeStamp)
	fmt.Println("    o UDPPayloadSizes =", e.UDPPayloadSizes)
	if printClusters {
		for i := range e.Clusters {
			cluster := e.Clusters[i]
			cluster.Print(printClusterDetails)
		}
		for i := range e.ClustersWoData {
			cluster := e.ClustersWoData[i]
			cluster.Print(printClusterDetails)
		}
	}
}

// idxFirstPulseLeftSide is the index within the []*pulse.Pulse slice of the
// first pulse on the left side of the DPGA
func (e *Event) Multiplicity() (uint8, []*pulse.Pulse, int) {
	var mult uint8 = 0
	var pulsesWSig []*pulse.Pulse
	var idxFirstPulseLeftSide int
	for i := range e.Clusters {
		pulsesWSigInCluster := e.Clusters[i].PulsesWithSignal()
		if i == 30 { // start processing clusters on left side of DPGA
			idxFirstPulseLeftSide = len(pulsesWSig)
		}
		pulsesWSig = append(pulsesWSig, pulsesWSigInCluster...)
		mult += uint8(len(pulsesWSigInCluster))
	}
	return mult, pulsesWSig, idxFirstPulseLeftSide
}

func (e *Event) PlotPulses(x pulse.XaxisType, onlyClustersWithSig bool, yrange pulse.YRange, xRangeZoomAround500 bool) {
	for i := range e.Clusters {
		cluster := &e.Clusters[i]
		if (onlyClustersWithSig == true && len(cluster.PulsesWithSignal()) > 0) || !onlyClustersWithSig {
			cluster.PlotPulses(e.ID, x, yrange, xRangeZoomAround500)
		}
	}
}

func (e *Event) pushPedestalCluster(cluster *pulse.Cluster) {
	for iPulse := range cluster.Pulses {
		pulse := &cluster.Pulses[iPulse]
		if pulse.HasSignal {
			continue
		}
		//fmt.Println("pushing pedestal sample for pulse", iPulse)
		for iSample := range pulse.Samples {
			sample := &pulse.Samples[iSample]
			capacitor := sample.Capacitor
			if capacitor != nil {
				noSamples := capacitor.NoPedestalSamples()
				if e.ID == 0 && noSamples != 0 {
					log.Fatal("noSamples != 0!")
				}
				capacitor.AddPedestalSample(sample.Amplitude)
			} else {
				log.Fatalf("no capacitor\n")
			}
		}
	}
}

func (e *Event) PushPedestalSamples() {
	for iCluster := range e.Clusters {
		cluster := &e.Clusters[iCluster]
		e.pushPedestalCluster(cluster)
		// 		for iPulse := range cluster.Pulses {
		// 			pulse := &cluster.Pulses[iPulse]
		// 			if pulse.HasSignal {
		// 				continue
		// 			}
		// 			for iSample := range pulse.Samples {
		// 				sample := &pulse.Samples[iSample]
		// 				capacitor := sample.Capacitor
		// 				if capacitor != nil {
		// 					noSamples := capacitor.NoPedestalSamples()
		// 					if e.ID == 0 && noSamples != 0 {
		// 						log.Fatal("noSamples != 0!")
		// 					}
		// 					capacitor.AddPedestalSample(sample.Amplitude)
		// 				}
		// 			}
		// 		}
	}
	for iCluster := range e.ClustersWoData {
		//fmt.Println("pushing pedestal sample for clusterWoData", iCluster)
		cluster := &e.ClustersWoData[iCluster]
		e.pushPedestalCluster(cluster)
	}
}

func (e *Event) pushTimeDepOffsetCluster(cluster *pulse.Cluster) {
	for iPulse := range cluster.Pulses {
		pulse := &cluster.Pulses[iPulse]
		if pulse.HasSignal {
			continue
		}
		ch := pulse.Channel
		if ch == nil {
			panic("pulse has no channel associated to it.")
		}
		ch.IncrementNoTimeDepOffsetSamples()
		for iSample := range pulse.Samples {
			sample := &pulse.Samples[iSample]
			ch.AddTimeDepOffsetSample(iSample, sample.Amplitude)
		}
	}
}

func (e *Event) PushTimeDepOffsetSamples() {
	for iCluster := range e.Clusters {
		cluster := &e.Clusters[iCluster]
		e.pushTimeDepOffsetCluster(cluster)
		// 		for iPulse := range cluster.Pulses {
		// 			pulse := &cluster.Pulses[iPulse]
		// 			if pulse.HasSignal {
		// 				continue
		// 			}
		// 			ch := pulse.Channel
		// 			if ch == nil {
		// 				panic("pulse has no channel associated to it.")
		// 			}
		// 			ch.IncrementNoTimeDepOffsetSamples()
		// 			for iSample := range pulse.Samples {
		// 				sample := &pulse.Samples[iSample]
		// 				ch.AddTimeDepOffsetSample(iSample, sample.Amplitude)
		// 			}
		// 		}
	}
	for iCluster := range e.ClustersWoData {
		cluster := &e.ClustersWoData[iCluster]
		e.pushTimeDepOffsetCluster(cluster)
	}
}

// center is the center of the intervall
// sig is the energy resolution
// n is the number of standard deviations considered
func (e *Event) PulsesInEnergyWindow(center, n, sig float64) []*pulse.Pulse {
	var selectedPulses []*pulse.Pulse
	for c := range e.Clusters {
		cluster := &e.Clusters[c]
		for p := range cluster.Pulses {
			pulse := &cluster.Pulses[p]
			if pulse.E > center-n*sig && pulse.E < center+n*sig {
				selectedPulses = append(selectedPulses, pulse)
			}
		}
	}
	return selectedPulses
}

// FindLORs finds LORs and adds them the the event.LORs slice
// It returns a slice of pointers to pulses associated to the selected LORs
func (e *Event) FindLORs(xbeam, ybeam, RmarMax, DeltaTMax, Emin, Emax float64, earlyTimePulses bool) {
	if len(e.LORs) > 0 {
		// 		log.Fatalf("LORs have already been found. This should not happen.\n")
		e.LORs = nil
	}

	mult, pulses, idxFirstLeft := e.Multiplicity()
	// 	fmt.Println("idxFirstLeft = ", idxFirstLeft)
	for i := 0; i < idxFirstLeft; i++ {
		pulseRight := pulses[i]
		if pulseRight.Hemi() != dpgadetector.Right {
			log.Fatalf("pulse is on wrong hemisphere -> error, should be fixed\n")
		}
		if pulseRight.HasSatSignal {
			continue
		}
		if pulseRight.Time30 == 0 {
			pulseRight.CalcRisingFront(true)
		}
		if earlyTimePulses {
			if pulseRight.Time30 < 20 || pulseRight.Time30 > 50 {
				continue
			}
		}
		// 		if pulseRight.Time30 == 0 || pulseRight.Time30 > 170 {
		// 			continue
		// 		}
		for j := idxFirstLeft; j < int(mult); j++ {
			pulseLeft := pulses[j]
			if pulseLeft.Hemi() != dpgadetector.Left {
				log.Fatalf("pulse is on wrong hemisphere -> error, should be fixed\n")
			}
			if pulseLeft.HasSatSignal {
				continue
			}
			if pulseLeft.Time30 == 0 {
				pulseLeft.CalcRisingFront(true)
			}
			if earlyTimePulses {
				if pulseLeft.Time30 < 20 || pulseLeft.Time30 > 50 {
					continue
				}
			}
			// 			if pulseLeft.Time30 == 0 || pulseLeft.Time30 > 170 {
			// 				continue
			// 			}
			// do MAR
			ch0 := pulseRight.Channel
			ch1 := pulseLeft.Channel
			x, y, z := reconstruction.Minimal(true, ch0, ch1, xbeam, ybeam)
			r := math.Sqrt(x*x + y*y)
			// 			fmt.Println("times: ", pulseRight.Time30, pulseLeft.Time30)
			if r < RmarMax &&
				math.Abs(pulseRight.Time30-pulseLeft.Time30) < DeltaTMax &&
				pulseRight.HasSatSignal != true && pulseLeft.HasSatSignal != true &&
				pulseRight.E > Emin && pulseRight.E < Emax &&
				pulseLeft.E > Emin && pulseLeft.E < Emax {
				l := NewLOR(pulseRight, pulseLeft, i, j, x, y, z, r)
				e.LORs = append(e.LORs, *l)
			}
		}
	}
}

func (e *Event) FindLORsLose(xbeam, ybeam float64) []LOR {
	var origLORs []LOR
	if len(e.LORs) > 0 {
		origLORs = make([]LOR, len(e.LORs))
		for i := range e.LORs {
			origLORs[i] = *(e.LORs[i].Copy())
		}
	}

	e.FindLORs(xbeam, ybeam, 1e6, 1e6, 0, 1e3, true)
	lors := e.LORs

	// 	fmt.Println("lengths: ", len(origLORs), len(lors))

	e.LORs = origLORs
	return lors
}

func (e *Event) FindTimesRF() []float64 {
	// 	for i := range event.ClustersWoData {
	// 		cluster := &event.ClustersWoData[i]
	// 		for j := range cluster.Pulses {
	// 			pulse := &cluster.Pulses[j]
	// 			fmt.Println(i, j, len(pulse.Samples))
	// 		}
	// 	}
	pulse := e.ClustersWoData[0].Pulses[0]
	ampSlice := pulse.MakeAmpSlice()
	var timesRF []float64
	if len(ampSlice) != 0 { // can compute TRF
		timesRF = utils.FindIntersections(e.ID, ampSlice, pulse.MakeTimeSlice())
	}
	return timesRF
}

// type Trigger int
//
// const (
// 	LvsR Trigger = iota + 1
// 	Lvs3L
// 	LvsL
// 	None
// )
//
// func WhichTrigger(pulses []*pulse.Pulse) Trigger {
// 	var trig Trigger
// 	for i := range pulses {
//
// 	}
// 	return trig
// }
