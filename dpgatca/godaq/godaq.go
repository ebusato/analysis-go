package main

import (
	"bufio"
	"encoding/json"
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
	"gitlab.in2p3.fr/avirm/analysis-go/dpga/dpgadetector"
	// 	"gitlab.in2p3.fr/avirm/analysis-go/dpga/trees"
	"gitlab.in2p3.fr/avirm/analysis-go/dpgatca/dq"
	"gitlab.in2p3.fr/avirm/analysis-go/dpgatca/rw"
	"gitlab.in2p3.fr/avirm/analysis-go/dpgatca/rwi"
	"gitlab.in2p3.fr/avirm/analysis-go/dpgatca/rwvme"
	"gitlab.in2p3.fr/avirm/analysis-go/pulse"
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
	ip           = flag.String("ip", "192.168.100.11", "IP address")
	port         = flag.String("p", "1024", "Port number")
	monFreq      = flag.Uint("mf", 100, "Monitoring frequency")
	monLight     = flag.Bool("monlight", false, "If set, the program performs a light monitoring, removing some plots")
	evtFreq      = flag.Uint("ef", 100, "Event printing frequency")
	st           = flag.Bool("st", false, "If set, server start time is used rather than client's one")
	debug        = flag.Bool("d", false, "If set, debugging informations are printed")
	webad        = flag.String("webad", ":5555", "server address:port")
	bro          = flag.Bool("bro", false, "If set, webbrowser is open (if not, it's up to the user to open it with the right address)")
	sleep        = flag.Bool("s", false, "If set, sleep a bit between events")
	sigthres     = flag.Uint("sigthres", 800, "Value above which a pulse is considered to have signal")
	notree       = flag.Bool("notree", false, "If set, no root tree is produced")
	test         = flag.Bool("test", false,
		"If set, update runs_test.csv rather than the \"official\" runs.csv file and name by default the output binary file using the following scheme: runXXX_test.bin")
	//refplots = flag.String("ref", os.Getenv("GOPATH")+"/src/gitlab.in2p3.fr/avirm/analysis-go/dpga/dqref/dq-run37020evtsPedReference.gob",
	//	"Name of the file containing reference plots. If empty, no reference plots are overlayed")
	refplots    = flag.String("ref", "", "Name of the file containing reference plots. If empty, no reference plots are overlayed")
	hvMonDegrad = flag.Uint("hvmondeg", 100, "HV monitoring frequency degradation factor")
	comment     = flag.String("c", "None", "Comment to be put in runs csv file")
	distr       = flag.String("distr", "ampl", "Possible values: ampl (default), charge, energy")
	calib       = flag.String("calib", "", "String indicating which calib to use (e.g. A1 for period A, version 1)")
	noped       = flag.Bool("noped", false, "If specified, no pedestal correction applied")
	notdo       = flag.Bool("notdo", false, "If specified, no time dependent offset correction applied")
	noen        = flag.Bool("noen", false, "If specified, no energy calibration applied.")
	rfcutmean   = flag.Float64("rfcutmean", 7, "Mean used to apply RF selection cut.")
	rfcutwidth  = flag.Float64("rfcutwidth", 5, "Width used to apply RF selection cut.")
	vme         = flag.Bool("vme", false, "If set, uses VME reader")
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

// Quartets is an array of 60 Quartet
// A value of type Quartets occupies 60*4*N*128 = 30689280 bits (taking N=999)
type Quartets [60]Quartet

// QuartetsWoData is an array of 12 Quartets without data (the last ones of each ASM board)
type QuartetsWoData [12]Quartet

// Data is the struct that is sent via the websocket to the web client.
type Data struct {
	EvtID                 uint           `json:"evt"`                   // event id (64 bits a priori)
	Time                  float64        `json:"time"`                  // time at which monitoring data are taken (64 bits)
	Counters              []uint32       `json:"counters"`              // counters provided by Thor
	TimeStamp             uint64         `json:"timestamp"`             // absolute timestamp from 64 MHz clock on Thor
	MonBufSize            int            `json:"monbufsize"`            // monitoring channel buffer size
	Freq                  float64        `json:"freq"`                  // number of events processed per second (64 bits)
	NoPacketsPerEvent     int            `json:"nopacketsperevent"`     // number of packets per event
	Qs                    Quartets       `json:"quartets"`              // quartets (30689280 bits)
	QsWoData              QuartetsWoData `json:"quartetswodata"`        // quartets without data
	FreqH                 string         `json:"freqh"`                 // frequency histogram
	ChargeL               string         `json:"chargel"`               // charge histograms for left hemisphere
	ChargeR               string         `json:"charger"`               // charge histograms for right hemisphere
	HVvals                string         `json:"hv"`                    // hv values
	MinRecZDistr          string         `json:"minreczdistrs"`         // minimal reconstruction Z distribution
	DeltaT30              string         `json:"deltat30"`              // distribution of the difference of T30
	EnergyAll             string         `json:"energyall"`             // distribution of energy (inclusive)
	AmplEnergyCorrelation string         `json:"amplenergycorrelation"` // amplitude or energy correlation for events with multiplicity=2
	HitQuartets           string         `json:"hitquartets"`           // 2D plot displaying quartets that are hit for events with multiplicity=2
	RFplotALaArnaud       string         `json:"rfplotalaarnaud"`       // 2D RF plot "a la Arnaud"
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

	var r rwi.Reader
	if !*vme {
		r, _ = rw.NewReader(bufio.NewReader(f))
	} else {
		r, _ = rwvme.NewReader(bufio.NewReader(f), rwvme.HeaderCAL)
	}
	r.SetDebug()
	// 	r, err := rw.NewReader(bufio.NewReader(f))
	// 	if err != nil {
	// 		log.Fatalf("could not open stream: %v\n", err)
	// 	}
	r.SetSigThreshold(*sigthres)

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
		"NoASMCards": 12,
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
	fmt.Println(noSamplesPulse, len(data))
	counter := 0
	for i := range pulse.Samples {
		if i%sampFreq == 0 {
			samp := &pulse.Samples[i]
			var x float64
			x = float64(samp.Index)
			// 			x = float64(samp.Capacitor.ID())

			// 			fmt.Println("i=", i, x, samp.Amplitude, counter)
			data[counter] = XY{X: x, Y: samp.Amplitude}
			// 			data[x] = XY{X: x, Y: samp.Amplitude}
			counter++
		}
	}
	return data
}

func stream(run uint32, r rwi.Reader, iEvent *uint, wg *sync.WaitGroup) {
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
	if *refplots != "" {
		dqplots.DQPlotRef = dq.NewDQPlotFromGob(*refplots)
	}
	/*outrootfileName := strings.Replace(*outfileName, ".bin", "LOR.root", 1)
	var treeLOR *trees.TreeLOR
	if !*notree {
		path, _ := os.Getwd()
		//fmt.Println(path)
		if strings.Contains(path, "analysis-go") {
			treeLOR = trees.NewTreeLOR(outrootfileName)
		} else {
			treeLOR = trees.NewTreeLOR(os.Getenv("HOME") + "/godaq_rootfiles/" + outrootfileName)
		}
	}*/
	minrecZsvg := ""
	start := time.Now()
	startabs := start
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
				event, err := r.ReadNextEvent()
				if *iEvent%*evtFreq == 0 {
					fmt.Printf("event %v (event.ID = %v)\n", *iEvent, event.ID)
				}
				if err != nil {
					panic(err)
				}
				if event == nil && err != nil { // EOF
					fmt.Printf("Reached EOF for iEvent = %v\n", *iEvent)
					return
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

					// 						if treeLOR != nil {
					// 							treeLOR.Fill(run, r.Header(), event, timesRF)
					// 						}
					// 						fmt.Println(" \nlength middle: ", len(event.LORs))
					dqplots.FillHistos(event, *rfcutmean, *rfcutwidth)
					// 						fmt.Println(" length after: ", len(event.LORs))
					if *iEvent%*monFreq == 0 {
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
						freqhsvg := utils.RenderSVG(tpfreq, 50, 10)

						chargeLsvg := ""
						chargeRsvg := ""
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
							tpchargeL := dqplots.MakeChargeAmplTiledPlot(whichVar, dpgadetector.Left)
							tpchargeR := dqplots.MakeChargeAmplTiledPlot(whichVar, dpgadetector.Right)
							chargeLsvg = utils.RenderSVG(tpchargeL, 45, 30)
							chargeRsvg = utils.RenderSVG(tpchargeR, 45, 30)
						}

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

						// send to channel
						if float64(len(datac)) >= 0.6*float64(datacsize) {
							fmt.Printf("Warning: monitoring buffer filled at more than 60 percent (len(datac) = %v, datacsize = %v)\n", len(datac), datacsize)
						}
						//fmt.Println(event.Counters)
						// 							var tstamp uint64 = uint64(event.Counters[3]) << 32 | uint64(event.Counters[2])
						// 							fmt.Println("tstamp = ", tstamp)

						Correlationsvg := AmplCorrelationsvg
						if *distr == "energy" {
							Correlationsvg = EnergyCorrelationsvg
						}
						// 							if event.Counters[0] != 0 {
						// 								fmt.Println(event.Counters[4], event.Counters[0], uint64(event.Counters[4])*uint64(64000000)/uint64(event.Counters[0]))
						// 							}
						dataToMonitor := Data{
							EvtID:                 event.ID,
							Time:                  time,
							Counters:              event.Counters,
							TimeStamp:             0,
							MonBufSize:            len(datac),
							Freq:                  freq,
							NoPacketsPerEvent:     int(event.NoFrames),
							Qs:                    qs,
							QsWoData:              qsWoData,
							FreqH:                 freqhsvg,
							ChargeL:               chargeLsvg,
							ChargeR:               chargeRsvg,
							MinRecZDistr:          minrecZsvg,
							DeltaT30:              DeltaT30svg,
							EnergyAll:             EnergyAllsvg,
							AmplEnergyCorrelation: Correlationsvg,
							HitQuartets:           HitQuartetssvg,
							RFplotALaArnaud:       RFplotALaArnaudsvg,
							LORMult:               LORMultsvg,
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
				if *sleep {
					fmt.Println("Warning: sleeping for 1 second")
					time.Sleep(1 * time.Second)
				}
			case false:
				fmt.Println("reached specified number of events, stopping.")
				// 				if treeLOR != nil {
				// 					treeLOR.Close()
				// 				}
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

		sb, err := json.Marshal(data)
		if err != nil {
			panic(err)
		}
		fmt.Printf("len(marshaled data) = %v bytes = %v Mbits\n", len(sb), len(sb)*8/1.e6)

		/////////////////////////////////////////////////
		err = websocket.JSON.Send(ws, data)
		if err != nil {
			log.Printf("error sending data: %v\n", err)
			return
		}
	}
}
