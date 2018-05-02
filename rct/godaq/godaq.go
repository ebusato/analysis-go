package main // package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"runtime/pprof"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/toqueteos/webbrowser"

	"golang.org/x/net/websocket"

	"gitlab.in2p3.fr/avirm/analysis-go/applyCorrCalib"
	"gitlab.in2p3.fr/avirm/analysis-go/dpga/calib/selectCalib"
	"gitlab.in2p3.fr/avirm/analysis-go/pulse"
	"gitlab.in2p3.fr/avirm/analysis-go/rct/dq"
	"gitlab.in2p3.fr/avirm/analysis-go/rct/rw"
	"gitlab.in2p3.fr/avirm/analysis-go/rct/trees"
	"gitlab.in2p3.fr/avirm/analysis-go/utils"
)

var (
	datacsize int = 10
	datac         = make(chan Data, datacsize)
	//terminateStream = make(chan bool)
	terminateRun = make(chan bool)
	pauseRun     = make(chan bool)
	resumeRun    = make(chan bool)
	pauseMonBool bool
	cpuprof      = flag.String("cpuprof", "", "Name of file for CPU profiling")
	noEvents     = flag.Uint("n", 100000, "Number of events")
	inFileName   = flag.String("i", "", "Name of input file (if non empty, use it rather than input stream from DAQ")
	outFileName  = flag.String("o", "", "Name of the output file. If not specified, setting it automatically using the following syntax: runXXX.bin (where XXX is the run number)")
	treeFileName = flag.String("ot", "", "Name of the TFile containing the output tree")
	ip           = flag.String("ip", "192.168.100.11", "IP address")
	port         = flag.String("p", "1024", "Port number")
	monFreq      = flag.Uint("mf", 100, "Monitoring frequency")
	monLight     = flag.Bool("monlight", false, "If set, the program performs a light monitoring, removing some plots")
	evtFreq      = flag.Uint("ef", 100, "Event printing frequency")
	st           = flag.Bool("st", false, "If set, server start time is used rather than client's one")
	debug        = flag.Bool("d", false, "If set, debugging informations are printed")
	webad        = flag.String("webad", ":5555", "server address:port")
	bro          = flag.Bool("bro", false, "If set, webbrowser is open (if not, it's up to the user to open it with the right address)")
	sleep        = flag.Int64("s", 0, "If set, sleeps for the time given as argument (in milliseconds) between two events")
	sigthres     = flag.Uint("sigthres", 800, "Value above which a pulse is considered to have signal")
	notree       = flag.Bool("notree", false, "If set, no root tree is produced")
	test         = flag.Bool("test", false,
		"If set, update runs_test.csv rather than the \"official\" runs.csv file and name by default the output binary file using the following scheme: runXXX_test.bin")
	refplots                  = flag.String("ref", "", "Name of the file containing reference plots. If empty, no reference plots are overlayed")
	comment                   = flag.String("c", "None", "Comment to be put in runs csv file")
	distr                     = flag.String("distr", "ampl", "Possible values: ampl (default), charge, energy")
	calib                     = flag.String("calib", "", "String indicating which calib to use (e.g. A1 for period A, version 1)")
	noped                     = flag.Bool("noped", false, "If specified, no pedestal correction applied")
	notdo                     = flag.Bool("notdo", false, "If specified, no time dependent offset correction applied")
	noen                      = flag.Bool("noen", false, "If specified, no energy calibration applied.")
	rfcutmean                 = flag.Float64("rfcutmean", 7, "Mean used to apply RF selection cut.")
	rfcutwidth                = flag.Float64("rfcutwidth", 5, "Width used to apply RF selection cut.")
	nopanic                   = flag.Bool("nopanic", false, "If set, the program won't panic when errors are not nil")
	printWarningMonBufferSize = flag.Bool("nowarning", false, "If set, the program won't print the warning related to the size of the monitoring buffer")
)

// XY is a struct used to store a couple of values
// It occupies 2*64 = 128 bits
type XY struct {
	X float64
	Y float64
}

// Pulse is a slice of XY
// A value of type Pulse occupies N*128, where N is the length of the slice
// N is by default equal to 999, so a value of type Pulse occupies 999*128 = 127872 bits
type Pulse []XY

// Quartet is an array of 4 pulses
// A value of type Quartet occupies 4*N*128 = 511488 bits (taking N=999)
type Quartet [4]Pulse

// Quartets is an array of 5 Quartet
type Quartets [5]Quartet

// QuartetsWoData is an array of 12 Quartets without data (the last ones of each ASM board)
type QuartetsWoData [1]Quartet

// Data is the struct that is sent via the websocket to the web client.
type Data struct {
	EvtID                 uint           `json:"evt"`                   // event id
	Time                  float64        `json:"time"`                  // time at which monitoring data are taken
	TimeStamp             uint64         `json:"timestamp"`             // absolute timestamp from 120 MHz clock (8.33 ns)
	MonBufSize            int            `json:"monbufsize"`            // monitoring channel buffer size
	Freq                  float64        `json:"freq"`                  // number of events processed by godaq per second
	FreqDaq               float64        `json:"freqdaq"`               // number of events generated by DAQ per second
	NoPacketsPerEvent     int            `json:"nopacketsperevent"`     // number of packets per event
	Qs                    Quartets       `json:"quartets"`              // quartets
	QsWoData              QuartetsWoData `json:"quartetswodata"`        // quartets without data
	FreqH                 string         `json:"freqh"`                 // frequency histogram
	Charge                string         `json:"charge"`                // charge histograms
	PulseMult             string         `json:"pulsemult"`             // pulse with signal multiplicity
	MinRecZDistr          string         `json:"minreczdistrs"`         // minimal reconstruction Z distribution
	DeltaT30              string         `json:"deltat30"`              // distribution of the difference of T30
	EnergyAll             string         `json:"energyall"`             // distribution of energy (inclusive)
	AmplEnergyCorrelation string         `json:"amplenergycorrelation"` // amplitude or energy correlation for events with multiplicity=2
	HitQuartets           string         `json:"hitquartets"`           // 2D plot displaying quartets that are hit for events with multiplicity=2
	RFplotALaArnaud       string         `json:"rfplotalaarnaud"`       // 2D RF plot "a la Arnaud"
	SRoutTiled            string         `json:"srouttiled"`            // SRout distributions for the 36 DRS's
	LORMult               string         `json:"lormult"`               // LOR multiplicity
}

func (d *Data) Print() {
	fmt.Println("\n-> Printing monitoring data:")
	fmt.Println("   o EvtID =", d.EvtID)
	fmt.Println("   o Time =", d.Time)
	fmt.Println("   o TimeStamp =", d.TimeStamp)
	fmt.Println("   o Freq =", d.Freq)
	fmt.Println("   o MonBufSize =", d.MonBufSize)
}

func TCPConn(p *string) *net.TCPConn {
	laddr, err := net.ResolveTCPAddr("tcp", *ip+":"+*p)
	if err != nil {
		log.Fatal(err)
	}
	tcp, err := net.DialTCP("tcp", nil, laddr)
	if err != nil {
		return nil
	}
	return tcp
}

func main() {
	log.SetFlags(log.Llongfile | log.LstdFlags)

	flag.Parse()

	if *cpuprof != "" {
		f, err := os.Create(*cpuprof)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	// Reader
	/*
		var tcp *net.TCPConn = nil
		tcp = TCPConn(port)
		for i := 0; tcp == nil; i++ {
			newportu, err := strconv.ParseUint(*port, 10, 64)
			if err != nil {
				panic(err)
			}
			newportu += 1
			newport := strconv.FormatUint(newportu, 10)
			fmt.Printf("Port %v not responding, trying %v\n", *port, newport)
			*port = newport
			tcp = TCPConn(port)
			if i >= 5 {
				log.Fatalf("Cannot find port to connect to server")
			}
		}

		//for i := 0; i < 4; i++ {
		//r, err := rw.NewReader(bufio.NewReader(tcp))
	*/

	// Reader
	f, err := os.Open(*inFileName)
	if err != nil {
		log.Fatalf("could not open data file: %v\n", err)
	}
	defer f.Close()

	var r *rw.Reader
	r, _ = rw.NewReader(bufio.NewReader(f))
	// 	r, err := rw.NewReader(bufio.NewReader(f))
	// 	if err != nil {
	// 		log.Fatalf("could not open stream: %v\n", err)
	// 	}
	r.SetSigThreshold(*sigthres)
	r.NoPanic = *nopanic

	// Start reading TCP stream
	// 	hdr := r.Header()
	// 	hdr.Print()

	// web address handling
	webadSlice := strings.Split(*webad, ":")
	if webadSlice[0] == "" {
		webadSlice[0] = utils.GetHostIP()
	}
	*webad = webadSlice[0] + ":" + webadSlice[1]
	fmt.Printf("Monitoring served at %v\n", *webad)

	// html template
	t := template.New("index-template.html")
	t, err = t.ParseFiles("root-fs/index-template.html")
	if err != nil {
		panic(err)
	}

	// Writer for html template file
	filehtml, err := os.Create("root-fs/index.html")
	if err != nil {
		log.Fatalf("could not create data file: %v\n", err)
	}
	defer filehtml.Close()

	htmlw := rw.NewWriter(filehtml)
	defer htmlw.Close()

	err = t.Execute(htmlw, map[string]interface{}{
		"WebAd":      *webad,
		"RunNumber":  0,
		"TimeStart":  0,
		"TimeStop":   0,
		"NoEvents":   0,
		"NoASMCards": 1,
		"NoSamples":  strconv.FormatUint(uint64(r.NoSamples()), 10),
	})
	if err != nil {
		panic(err)
	}

	// Start goroutines
	const N = 1
	var wg sync.WaitGroup
	wg.Add(N)

	if *debug {
		r.SetDebug()
	}

	iEvent := uint(0)

	var currentRunNumber uint32
	go stream(currentRunNumber, r, &iEvent, &wg)
	go command()
	go webserver()

	wg.Wait()
}

func webserver() {
	if *bro {
		webbrowser.Open("http://" + *webad)
	}
	//http.HandleFunc("/", plotHandle)
	http.Handle("/", http.FileServer(http.Dir("./root-fs")))
	//http.Handle("/", http.FileServer(assetFS()))
	http.Handle("/data", websocket.Handler(dataHandler))
	err := http.ListenAndServe(*webad, nil)
	if err != nil {
		log.Fatal(err)
	}
}

// func command(terminateRun, pauseRun chan bool) {
func command() {
	for {
		in := bufio.NewReader(os.Stdin) // to be replaced by Scanner
		word, _ := in.ReadString('\n')
		word = strings.Replace(word, "\n", "", -1)
		switch word {
		default:
			fmt.Println("command not known, what do you mean ?", word)
		case "stop":
			fmt.Println("stopping run")
			terminateRun <- true
		case "pause":
			fmt.Println("pausing run")
			pauseRun <- true
		case "resume":
			fmt.Println("resuming run")
			resumeRun <- true
		case "pause mon":
			fmt.Println("pausing monitoring")
			pauseMonBool = true
		case "resume mon":
			fmt.Println("resuming monitoring")
			pauseMonBool = false
		}
	}
}

func GetMonData(sampFreq int, pulse pulse.Pulse) []XY {
	noSamplesPulse := int(pulse.NoSamples())
	// 	data := make([]XY, noSamplesPulse/sampFreq+1)
	data := make([]XY, noSamplesPulse/sampFreq)
	// 	data := make([]XY, 1024)
	if noSamplesPulse == 0 {
		return data
	}
	// 	fmt.Println(noSamplesPulse, len(data))
	counter := 0
	for i := range pulse.Samples {
		if i%sampFreq == 0 {
			samp := &pulse.Samples[i]
			var x float64
			//x = float64(samp.Index)
			x = float64(samp.Capacitor.ID())

			// 			fmt.Println("i=", i, x, samp.Amplitude, counter)
			data[counter] = XY{X: x, Y: samp.Amplitude}
			// 			data[x] = XY{X: x, Y: samp.Amplitude}
			counter++
		}
	}
	return data
}

func stream(run uint32, r *rw.Reader, iEvent *uint, wg *sync.WaitGroup) {
	defer wg.Done()
	doPedestal := false
	doTimeDepOffset := false
	doEnergyCalib := false
	if *calib != "" {
		selectCalib.Which(*calib)
		if !*noped {
			doPedestal = true
		}
		if !*notdo {
			doTimeDepOffset = true
		}
		if !*noen {
			doEnergyCalib = true
		}
	}
	noEventsForMon := uint64(0)
	dqplots := dq.NewDQPlot()
	outRootFileName := strings.Replace(*inFileName, ".bin", ".root", 1)
	var tree *trees.Tree
	if !*notree {
		if *treeFileName != "" {
			tree = trees.NewTree(*treeFileName)
		} else {
			path, _ := os.Getwd()
			//fmt.Println(path)
			if strings.Contains(path, "analysis-go") {
				tree = trees.NewTree(outRootFileName)
			} else {
				tree = trees.NewTree(os.Getenv("HOME") + "/godaq_rootfiles/" + outRootFileName)
			}
		}
	}
	minrecZsvg := ""
	start := time.Now()
	startabs := start

	prevEvtID := uint(0)
	prevEvtTimeStamp := uint64(0)
	for {
		select {
		// 		case <-terminateStream:
		case <-terminateRun:
			*noEvents = *iEvent + 1
			fmt.Printf("terminating stream for total number of events = %v.\n", *noEvents)
		case <-pauseRun:
			<-resumeRun
		default:
			switch *iEvent < *noEvents {
			case true:
				if *iEvent%*evtFreq == 0 {
					fmt.Printf("event %v\n", *iEvent)
				}
				event, err := r.ReadNextEvent()
				// 				if *iEvent%*evtFreq == 0 {
				// 					fmt.Printf(" -> event.ID=%v\n", event.ID)
				// 				}
				if err != nil {
					// 					panic(err)
					fmt.Println(err)
				}
				if event == nil && err != nil { // EOF
					fmt.Printf("Reached EOF for iEvent = %v\n", *iEvent)
					return
				}

				if event.ID%1 == 0 {
					//fmt.Println("before")
					//event.PlotPulses(pulse.XaxisCapacitor, true, pulse.YRangeAuto, false)
					// 					event.PlotPulses(pulse.XaxisIndex, true, pulse.YRangeAuto, false)
					//fmt.Println("after")
				}

				// 					event.Print(true, false)
				noEventsForMon++
				////////////////////////////////////////////////////////////////////////////////////////////
				// Monitoring
				if !pauseMonBool {
					//////////////////////////////////////////////////////
					// Corrections
					event = applyCorrCalib.CorrectEvent(event, doPedestal, doTimeDepOffset, doEnergyCalib)
					//////////////////////////////////////////////////////
					// 						dqplots.FillHistos(event)
					// mult, pulsesWithSignal, _ := event.Multiplicity()

					if tree != nil {
						tree.Fill(run, event)
					}
					// 						fmt.Println(" \nlength middle: ", len(event.LORs))
					dqplots.FillHistos(event, *rfcutmean, *rfcutwidth)
					// 						fmt.Println(" length after: ", len(event.LORs))
					if (*iEvent+1)%*monFreq == 0 {
						// Webserver data

						var qs Quartets
						var qsWoData QuartetsWoData
						// 							sampFreq := 5
						sampFreq := 1
						if *monLight {
							sampFreq = 20
						}
						for iq := 0; iq < len(qs); iq++ {
							qs[iq][0] = GetMonData(sampFreq, event.Clusters[iq].Pulses[0])
							qs[iq][1] = GetMonData(sampFreq, event.Clusters[iq].Pulses[1])
							qs[iq][2] = GetMonData(sampFreq, event.Clusters[iq].Pulses[2])
							qs[iq][3] = GetMonData(sampFreq, event.Clusters[iq].Pulses[3])
						}
						for iq := 0; iq < len(qsWoData); iq++ {
							qsWoData[iq][0] = GetMonData(sampFreq, event.ClustersWoData[iq].Pulses[0])
							qsWoData[iq][1] = GetMonData(sampFreq, event.ClustersWoData[iq].Pulses[1])
							qsWoData[iq][2] = GetMonData(sampFreq, event.ClustersWoData[iq].Pulses[2])
							qsWoData[iq][3] = GetMonData(sampFreq, event.ClustersWoData[iq].Pulses[3])
						}

						// Make frequency histo plot
						tpfreq := dqplots.MakeFreqTiledPlot()
						freqhsvg := utils.RenderSVG(tpfreq, 25, 8)

						chargesvg := ""
						if !*monLight {
							// Make charge (or amplitude) distrib histo plot
							var whichVar dq.WhichVar
							switch *distr {
							case "charge":
								whichVar = dq.Charge
							case "ampl":
								whichVar = dq.Amplitude
							case "energy":
								whichVar = dq.Energy
							default:
								panic("String passed to option -distr not recognized (see godaq -h)")
							}
							tpcharge := dqplots.MakeChargeAmplTiledPlot(whichVar)
							chargesvg = utils.RenderSVG(tpcharge, 45, 7)
						}

						// Make mult histo plot
						tppulsemult := dqplots.MakePulseMultTiledPlot()
						pulsemultsvg := utils.RenderSVG(tppulsemult, 18, 8)

						// MA algo
						tpMinRecZ := dqplots.MakeMinRecZDistr()
						minrecZsvg = utils.RenderSVG(tpMinRecZ, 25, 6)

						stop := time.Now()
						duration := stop.Sub(start).Seconds()
						start = stop
						time := stop.Sub(startabs).Seconds()
						freq := float64(noEventsForMon) / duration
						if *iEvent == 0 {
							freq = 0
						}

						// Make DeltaT30 plot
						pDeltaT30 := dqplots.MakeDeltaT30Plot()
						DeltaT30svg := utils.RenderSVG(pDeltaT30, 12, 7.5)

						// Make inclusive energy plots
						pEnergyAll := dqplots.MakeEnergyPlot()
						EnergyAllsvg := utils.RenderSVG(pEnergyAll, 10, 7.5)

						// Make LOR multiplicity plot
						pLORMult := dqplots.MakeLORMultPlot()
						LORMultsvg := utils.RenderSVG(pLORMult, 10, 7.5)

						// Make ampl correlation plot
						pAmplCorrelation := dqplots.MakeAmplCorrelationPlot()
						AmplCorrelationsvg := ""
						if *iEvent > 0 && dqplots.AmplCorrelation.Entries() > 0 {
							AmplCorrelationsvg = utils.RenderSVG(pAmplCorrelation, 9, 9)
						}

						// Make energy correlation plot
						pEnergyCorrelation := dqplots.MakeEnergyCorrelationPlot()
						EnergyCorrelationsvg := ""
						if *iEvent > 0 && dqplots.EnergyCorrelation.Entries() > 0 {
							EnergyCorrelationsvg = utils.RenderSVG(pEnergyCorrelation, 9, 9)
						}

						// Make RF plot "a la arnaud"
						pRFplotALaArnaud := dqplots.MakeRFPlotALaArnaud()
						RFplotALaArnaudsvg := ""
						if *iEvent > 0 && dqplots.HEnergyVsDeltaTggRF.Entries() > 0 {
							RFplotALaArnaudsvg = utils.RenderSVG(pRFplotALaArnaud, 9, 9)
						}

						// Make HitQuartets plot
						pHitQuartets := dqplots.MakeHitQuartetsPlot()
						HitQuartetssvg := ""
						if *iEvent > 0 && dqplots.HitQuartets.Entries() > 0 {
							HitQuartetssvg = utils.RenderSVG(pHitQuartets, 9, 9)
						}

						// Make SRout plot
						pSRout := dqplots.MakeSRoutTiledPlot()
						SRoutsvg := ""
						if *iEvent > 0 {
							SRoutsvg = utils.RenderSVG(pSRout, 30, 8)
						}

						// send to channel
						if !*printWarningMonBufferSize && float64(len(datac)) >= 0.6*float64(datacsize) {
							fmt.Printf("Warning: monitoring buffer filled at more than 60 percent (len(datac) = %v, datacsize = %v)\n", len(datac), datacsize)
						}

						Correlationsvg := AmplCorrelationsvg
						if *distr == "energy" {
							Correlationsvg = EnergyCorrelationsvg
						}

						freqDaq := float64(0)
						// In principle it should be sufficient to skip first event, but here we skip two as first event has bad event.ID
						// (for an unknown reason) and thus leads to bad freqDaq value.
						if *iEvent > 1 {
							freqDaq = float64(120e6) * float64(event.ID-prevEvtID) / float64(event.TimeStamp-prevEvtTimeStamp) // The clock frequency is 120 MHz
							// 							fmt.Println("debug freqdaq: ", freqDaq, event.ID, prevEvtID, event.TimeStamp, prevEvtTimeStamp)

						}
						prevEvtID = event.ID
						prevEvtTimeStamp = event.TimeStamp

						dataToMonitor := Data{
							EvtID:                 event.ID,
							Time:                  time,
							TimeStamp:             event.TimeStamp,
							MonBufSize:            len(datac),
							Freq:                  freq,
							FreqDaq:               freqDaq,
							NoPacketsPerEvent:     int(event.NoFrames),
							Qs:                    qs,
							QsWoData:              qsWoData,
							FreqH:                 freqhsvg,
							Charge:                chargesvg,
							PulseMult:             pulsemultsvg,
							MinRecZDistr:          minrecZsvg,
							DeltaT30:              DeltaT30svg,
							EnergyAll:             EnergyAllsvg,
							AmplEnergyCorrelation: Correlationsvg,
							HitQuartets:           HitQuartetssvg,
							RFplotALaArnaud:       RFplotALaArnaudsvg,
							LORMult:               LORMultsvg,
							SRoutTiled:            SRoutsvg,
						}
						//dataToMonitor.Print()
						datac <- dataToMonitor
						noEventsForMon = 0
					}
				}
				// End of monitoring
				////////////////////////////////////////////////////////////////////////////////////////////
				*iEvent++
				//noEventsForMon++
				if *sleep != 0 {
					sleepingTime := time.Duration(*sleep) * time.Millisecond
					fmt.Printf("Warning: sleeping for %v\n", sleepingTime)
					time.Sleep(sleepingTime)
				}
			case false:
				fmt.Println("reached specified number of events, stopping.")
				if tree != nil {
					tree.Close()
				}
				return
			}
		}
	} // event loop
}

func dataHandler(ws *websocket.Conn) {
	for data := range datac {
		/////////////////////////////////////////////////
		// uncomment to have an estimation of the total
		// amount of data that passes through the websocket

		// 		sb, err := json.Marshal(data)
		// 		if err != nil {
		// 			panic(err)
		// 		}
		//fmt.Printf("len(marshaled data) = %v bytes = %v Mbits\n", len(sb), len(sb)*8/1.e6)

		/////////////////////////////////////////////////
		err := websocket.JSON.Send(ws, data)
		if err != nil {
			log.Printf("error sending data: %v\n", err)
			return
		}
	}
}
