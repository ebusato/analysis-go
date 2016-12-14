package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"runtime/pprof"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/go-hep/csvutil"
	"github.com/go-hep/hbook"
	"github.com/go-hep/hplot"
	"github.com/gonum/plot/palette"
	"github.com/gonum/plot/plotter"
	"github.com/toqueteos/webbrowser"

	"golang.org/x/net/websocket"

	"gitlab.in2p3.fr/avirm/analysis-go/applyCorrCalib"
	"gitlab.in2p3.fr/avirm/analysis-go/dpga/dpgadetector"
	"gitlab.in2p3.fr/avirm/analysis-go/dpga/dq"
	"gitlab.in2p3.fr/avirm/analysis-go/dpga/rw"
	"gitlab.in2p3.fr/avirm/analysis-go/dpga/trees"
	"gitlab.in2p3.fr/avirm/analysis-go/pulse"
	"gitlab.in2p3.fr/avirm/analysis-go/reconstruction"
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
	hdrType      = rw.HeaderCAL
	cpuprof      = flag.String("cpuprof", "", "Name of file for CPU profiling")
	noEvents     = flag.Uint("n", 100000, "Number of events")
	outfileName  = flag.String("o", "", "Name of the output file. If not specified, setting it automatically using the following syntax: runXXX.bin (where XXX is the run number)")
	ip           = flag.String("ip", "192.168.100.11", "IP address")
	port         = flag.String("p", "1024", "Port number")
	monFreq      = flag.Uint("mf", 100, "Monitoring frequency")
	monLight     = flag.Bool("monlight", false, "If set, the program performs a light monitoring, removing some plots")
	evtFreq      = flag.Uint("ef", 100, "Event printing frequency")
	st           = flag.Bool("st", false, "If set, server start time is used rather than client's one")
	debug        = flag.Bool("d", false, "If set, debugging informations are printed")
	webad        = flag.String("webad", ":5555", "server address:port")
	nobro        = flag.Bool("nobro", false, "If set, no webbrowser are open (it's up to the user to open it with the right address)")
	sleep        = flag.Bool("s", false, "If set, sleep a bit between events")
	sigthres     = flag.Uint("sigthres", 800, "Value above which a pulse is considered to have signal")
	notree       = flag.Bool("notree", false, "If set, no root tree is produced")
	test         = flag.Bool("test", false,
		"If set, update runs_test.csv rather than the \"official\" runs.csv file and name by default the output binary file using the following scheme: runXXX_test.bin")
	refplots = flag.String("ref", os.Getenv("GOPATH")+"/src/gitlab.in2p3.fr/avirm/analysis-go/dpga/dqref/dq-run37020evtsPedReference.gob",
		"Name of the file containing reference plots. If empty, no reference plots are overlayed")
	hvMonDegrad = flag.Uint("hvmondeg", 100, "HV monitoring frequency degradation factor")
	comment     = flag.String("c", "None", "Comment to be put in runs csv file")
	distr       = flag.String("distr", "charge", "Possible values: charge (default), ampl, energy")
	ped         = flag.String("ped", "", "Name of the csv file containing pedestal constants. If not set, pedestal corrections are not applied.")
	tdo         = flag.String("tdo", "", "Name of the csv file containing time dependent offsets. If not set, time dependent offsets are not applied. Relevant only when ped!=\"\".")
	en          = flag.String("en", "", "Name of the csv file containing energy calibration constants. If not set, energy calibration is not applied.")
)

// XY is a struct used to store a couple of values
// It occupies 2*64 = 128 bits
type XY struct {
	X float64
	Y float64
}

// XYZ is a struct used to store a triplet of values
// It occupies 3*64 = 192 bits
type XYZ struct {
	X float64
	Y float64
	Z float64
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

// H1D is a local struct representing a histogram
// The length of the slice is the number of bins
// X is the bin center and Y the bin content
// By default, the number of bins is 8
// A value of type H1D therefore occupies by default 8*128 = 1024 bits
type H1D []XY

func NewH1D(h *hbook.H1D) H1D {
	var hist H1D
	nbins := h.Len()
	for i := 0; i < nbins; i++ {
		x, y := h.XY(i)
		hist = append(hist, XY{X: x, Y: y})
	}
	return hist
}

type HVexec struct {
	execName string
}

func NewHVexec(execName string, coefDir string) *HVexec {
	if !utils.Exists(execName) {
		fmt.Printf("could not find executable %v\n", execName)
		return nil
	}
	if !utils.Exists(coefDir) {
		fmt.Printf("could not find directory %v\n", coefDir)
		return nil
		if !utils.Exists(coefDir+"/Coef_poly_C001.txt") ||
			!utils.Exists(coefDir+"/Coef_poly_C002.txt") ||
			!utils.Exists(coefDir+"/Coef_poly_C003.txt") ||
			!utils.Exists(coefDir+"/Coef_poly_C004.txt") {
			fmt.Printf("could not find at least one of the Coef_poly_C00?.txt file\n")
			return nil
		}
	}
	_, linkName := path.Split(coefDir)
	if !utils.Exists(linkName) {
		fmt.Println("Link to coeff directory not existing, making it")
		err := os.Symlink(coefDir, linkName)
		if err != nil {
			panic(err)
		}
	}
	return &HVexec{
		execName: execName,
	}
}

type HVvalue struct {
	Raw float64
	HV  float64
}

type HVvalues [4][16]HVvalue // first index refers to HV card (there are 4 cards), second index refers to channels (there are 16 channels per card)

// NewHVvalues creates a new HVvalues object from a HVexec.
func NewHVvalues(hvex *HVexec) *HVvalues {
	hvvals := &HVvalues{}
	for iHVcard := int64(1); iHVcard <= 4; iHVcard++ {
		cmd := exec.Command(hvex.execName, "--serial", strconv.FormatInt(iHVcard, 10), "--display")
		//fmt.Println(cmd.Args)
		cmdReader, err := cmd.StdoutPipe()
		if err != nil {
			log.Fatalf("error executing %v\n", hvex.execName)
		}
		scanner := bufio.NewScanner(cmdReader)

		err = cmd.Start()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error starting Cmd", err)
			os.Exit(1)
		}
		for scanner.Scan() {
			line := scanner.Text()
			fields := strings.Split(line, " ")
			if fields[0] == "Read" {
				//fmt.Println("debug: ", fields[3], fields[8], fields[11])
				channelIdx, err := strconv.ParseInt(fields[3], 10, 64)
				if err != nil {
					panic(err)
				}
				raw, err := strconv.ParseFloat(fields[8], 64)
				if err != nil {
					panic(err)
				}
				svalwithunwantedchar := fields[11]
				sval := strings.TrimRight(svalwithunwantedchar, " ")
				val, err := strconv.ParseFloat(sval, 64)
				if err != nil {
					panic(err)
				}
				(*hvvals)[iHVcard-1][channelIdx] = HVvalue{Raw: raw, HV: val}
			}
		}
		err = cmd.Wait()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error waiting for Cmd", err)
			os.Exit(1)
		}
	} // end of loop over HV cards
	return hvvals
}

// Data is the struct that is sent via the websocket to the web client.
type Data struct {
	EvtID             uint     `json:"evt"`               // event id (64 bits a priori)
	Time              float64  `json:"time"`              // time at which monitoring data are taken (64 bits)
	Counters          []uint32 `json:"counters"`          // counters provided by Thor
	MonBufSize        int      `json:"monbufsize"`        // monitoring channel buffer size
	Freq              float64  `json:"freq"`              // number of events processed per second (64 bits)
	Qs                Quartets `json:"quartets"`          // (30689280 bits)
	Mult              H1D      `json:"mult"`              // multiplicity of pulses (1024 bits)
	FreqH             string   `json:"freqh"`             // frequency histogram
	ChargeL           string   `json:"chargel"`           // charge histograms for left hemisphere
	ChargeR           string   `json:"charger"`           // charge histograms for right hemisphere
	HVvals            string   `json:"hv"`                // hv values
	MinRec            []XYZ    `json:"minrec"`            // outcome of the minimal reconstruction algorithm
	MinRec1DDistrs    string   `json:"minrec1Ddistrs"`    // minimal reconstruction X, Y, Z distributions
	DeltaT30          string   `json:"deltat30"`          // distribution of the difference of T30
	ChargeCorrelation string   `json:"chargecorrelation"` // charge correlation for events with multiplicity=2
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

	flag.Var(&hdrType, "h", "Type of header: HeaderCAL or HeaderOld")
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
	r, err := rw.NewReader(bufio.NewReader(tcp), hdrType)
	if err != nil {
		log.Fatalf("could not open stream: %v\n", err)
	}
	r.SigThreshold = *sigthres

	// determine run number
	var runCSVFileName string
	switch *test {
	case true:
		runCSVFileName = os.Getenv("HOME") + "/godaq/runs/runs_test.csv"
	case false:
		runCSVFileName = os.Getenv("HOME") + "/godaq/runs/runs.csv"
	}
	if !utils.Exists(runCSVFileName) {
		fmt.Printf("could not open %v -> nothing will be written to it.\n", runCSVFileName)
		return
	}
	prevRunNumber := getPreviousRunNumber(runCSVFileName)
	currentRunNumber := prevRunNumber + 1
	fmt.Printf("Previous run number is %v -> setting current run number to %v\n", prevRunNumber, currentRunNumber)

	// Writer for binary file
	if *outfileName == "" {
		*outfileName = "run" + strconv.FormatUint(uint64(currentRunNumber), 10)
		if *test {
			*outfileName += "_test"
		}
		*outfileName += ".bin"
	}
	filew, err := os.Create(*outfileName)
	if err != nil {
		log.Fatalf("could not create data file: %v\n", err)
	}
	defer filew.Close()

	bufiow := bufio.NewWriter(filew)
	w := rw.NewWriter(bufiow)
	defer w.Close()

	// Start reading TCP stream
	hdr := r.Header()

	err = w.Header(hdr, !*st)
	if err != nil {
		log.Fatalf("error writing header: %v\n", err)
	}
	hdr.Print()

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
		"WebAd":                   *webad,
		"RunNumber":               currentRunNumber,
		"TimeStart":               time.Unix(int64(hdr.TimeStart), 0).Format(time.UnixDate),
		"TimeStop":                time.Unix(int64(hdr.TimeStop), 0).Format(time.UnixDate),
		"NoEvents":                strconv.FormatUint(uint64(hdr.NoEvents), 10),
		"NoASMCards":              strconv.FormatUint(uint64(hdr.NoASMCards), 10),
		"NoSamples":               strconv.FormatUint(uint64(r.NoSamples()), 10),
		"DataToRead":              "0x" + strconv.FormatUint(uint64(hdr.DataToRead), 16),
		"TriggerEq":               "0x" + strconv.FormatUint(uint64(hdr.TriggerEq), 16),
		"TriggerDelay":            "0x" + strconv.FormatUint(uint64(hdr.TriggerDelay), 16),
		"ChanUsedForTrig":         "0x" + strconv.FormatUint(uint64(hdr.ChanUsedForTrig), 16),
		"Threshold":               strconv.FormatUint(uint64(hdr.Threshold), 10),
		"LowHighThres":            "0x" + strconv.FormatUint(uint64(hdr.LowHighThres), 16),
		"TrigSigShapingHighThres": "0x" + strconv.FormatUint(uint64(hdr.TrigSigShapingHighThres), 16),
		"TrigSigShapingLowThres":  "0x" + strconv.FormatUint(uint64(hdr.TrigSigShapingLowThres), 16),
	})
	if err != nil {
		panic(err)
	}

	// Start goroutines
	const N = 1
	var wg sync.WaitGroup
	wg.Add(N)

	if *debug {
		r.Debug = true
	}

	iEvent := uint(0)
	// 	go control(terminateStream, terminateRun)
	// 	go stream(terminateStream, r, w, &iEvent, &wg)
	// 	go command(terminateRun, pauseRun)
	//go control()
	go stream(currentRunNumber, r, w, &iEvent, &wg)
	go command()
	go webserver()

	wg.Wait()

	// Update header
	//bufiow.Flush()
	timeStop := uint32(time.Now().Unix())
	noEvents := uint32(iEvent)
	updateHeader(filew, 16, timeStop)
	updateHeader(filew, 20, noEvents)

	// Dump run info in csv. Only relevant when ran on DAQ PC, where the csv file is present.
	updateRunsCSV(runCSVFileName, currentRunNumber, timeStop, noEvents, *outfileName, hdr)
	updateHeader(filew, 4, currentRunNumber)
}

func updateHeader(f *os.File, offset int64, val uint32) {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], val)
	f.WriteAt(buf[:], offset)
}

type RunsCSV struct {
	RunNumber   uint32
	NoEvents    uint32
	Threshold   uint32
	OutFileName string
	ExecDir     string
	StartTime   string
	StopTime    string
	Comment     string
}

func getPreviousRunNumber(fileName string) uint32 {
	tbl, err := csvutil.Open(fileName)
	if err != nil {
		log.Fatalf("could not open runs.csv.\n")
	}
	defer tbl.Close()
	tbl.Reader.Comma = ' '
	tbl.Reader.Comment = '#'

	nLines, err := utils.LineCounter(fileName)
	if err != nil {
		log.Fatalf("error reading the number of lines in runs.csv\n", err)
	}

	// the -2 is because there are two lines of header at the beginning
	rows, err := tbl.ReadRows(int64(nLines-2-1), -1)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	var data RunsCSV
	rows.Next()
	err = rows.Scan(&data)
	if err != nil {
		log.Fatalf("error reading row: %v\n", err)
	}
	return data.RunNumber
}

func updateRunsCSV(csvFileName string, runNumber uint32, timeStop uint32, noEvents uint32, outfileName string, hdr *rw.Header) {
	// Determine working directory
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Open csv file in append mode
	tbl, err := csvutil.Append(csvFileName)
	if err != nil {
		log.Fatalf("could not create dpgageom.csv: %v\n", err)
	}
	defer tbl.Close()
	tbl.Writer.Comma = ' '

	data := RunsCSV{
		RunNumber:   runNumber,
		NoEvents:    noEvents,
		Threshold:   hdr.Threshold,
		OutFileName: outfileName,
		ExecDir:     pwd,
		StartTime:   time.Unix(int64(hdr.TimeStart), 0).Format(time.UnixDate),
		StopTime:    time.Unix(int64(timeStop), 0).Format(time.UnixDate),
		Comment:     *comment,
	}
	err = tbl.WriteRow(data)
	if err != nil {
		log.Fatalf("error writing row: %v\n", err)
	}

	// implement git commit
}

func webserver() {
	if !*nobro {
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

// func control(terminateStream chan bool, terminateRun chan bool) {
// func control() {
// 	<-terminateRun
// 	fmt.Printf("command is ended, terminating stream.\n")
// 	terminateStream <- true
// }

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

/*
func GetMonData(noSamples uint16, pulse pulse.Pulse) []XY {
	data := make([]XY, noSamples)
	for i := range data {
		var x float64
		var y float64
		switch pulse.NoSamples() == noSamples {
		case true:
			x = float64(pulse.Samples[i].Index)
			y = pulse.Samples[i].Amplitude
		case false:
			x = 0
			y = 0
		}
		data[i] = XY{
			X: x,
			Y: y,
		}
	}
	return data
}
*/

func GetMonData(sampFreq int, pulse pulse.Pulse) []XY {
	noSamplesPulse := int(pulse.NoSamples())
	data := make([]XY, noSamplesPulse/sampFreq+1)
	if noSamplesPulse == 0 {
		return data
	}
	counter := 0
	for i := range pulse.Samples {
		if i%sampFreq == 0 {
			samp := &pulse.Samples[i]
			var x float64
			x = float64(samp.Index)
			//x = float64(samp.Capacitor.ID())
			data[counter] = XY{X: x, Y: samp.Amplitude}
			counter++
		}
	}
	return data
}

// func stream(terminateStream chan bool, r *rw.Reader, w *rw.Writer, iEvent *uint, wg *sync.WaitGroup) {
func stream(run uint32, r *rw.Reader, w *rw.Writer, iEvent *uint, wg *sync.WaitGroup) {
	defer wg.Done()
	if *ped != "" {
		dpgadetector.Det.ReadPedestalsFile(*ped)
	}
	if *tdo != "" {
		dpgadetector.Det.ReadTimeDepOffsetsFile(*tdo)
	}
	if *en != "" {
		fmt.Println("test 00")
		dpgadetector.Det.ReadEnergyCalibFile(*en)
	}
	noEventsForMon := uint64(0)
	dqplots := dq.NewDQPlot()
	if *refplots != "" {
		dqplots.DQPlotRef = dq.NewDQPlotFromGob(*refplots)
	}
	hvexec := NewHVexec(os.Getenv("HOME")+"/Acquisition/hv/ht-caen", os.Getenv("HOME")+"/Acquisition/hv/Coeff")
	outrootfileName := strings.Replace(*outfileName, ".bin", ".root", 1)
	var treeMult2 *trees.TreeMult2
	if !*notree {
		treeMult2 = trees.NewTreeMult2(outrootfileName)
	}
	var minrec []XYZ
	minrec1Dsvg := ""
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
				if *iEvent%*evtFreq == 0 {
					fmt.Printf("event %v\n", *iEvent)
				}
				event, status := r.ReadNextEvent()
				//fmt.Println("counters:", event.Counters)
				if status == false {
					panic("error: status is false\n")
				}
				switch event.IsCorrupted {
				case false:
					//event.Print(true, false)
					w.Event(event)
					noEventsForMon++
					////////////////////////////////////////////////////////////////////////////////////////////
					// Monitoring
					if !pauseMonBool {
						//////////////////////////////////////////////////////
						// Corrections
						doPedestal := false
						doTimeDepOffset := false
						doEnergyCalib := false
						if *ped != "" {
							doPedestal = true
						}
						if *tdo != "" {
							doTimeDepOffset = true
						}
						if *en != "" {
							doEnergyCalib = true
						}
						event = applyCorrCalib.CorrectEvent(event, doPedestal, doTimeDepOffset, doEnergyCalib)
						//////////////////////////////////////////////////////
						dqplots.FillHistos(event)
						mult, pulsesWithSignal := event.Multiplicity()
						if mult == 2 {
							//dqplots.FillHistos(event)
							if len(pulsesWithSignal) != 2 {
								panic("mult == 2 but len(pulsesWithSignal) != 2: this should NEVER happen !")
							}
							ch0 := pulsesWithSignal[0].Channel
							ch1 := pulsesWithSignal[1].Channel
							doMinRec := true
							if r.Header().TriggerEq == 3 {
								// In case TriggerEq = 3 (pulser), one has to check that the two pulses are
								// on different hemispheres, otherwise the minimal reconstruction is not well
								// defined
								hemi0, ok := ch0.Quartet.DRS.ASMCard.UpStr.(*dpgadetector.Hemisphere)
								if !ok {
									panic("ch0.Quartet.DRS.ASMCard.UpStr type assertion failed")
								}
								hemi1, ok := ch1.Quartet.DRS.ASMCard.UpStr.(*dpgadetector.Hemisphere)
								if !ok {
									panic("ch0.Quartet.DRS.ASMCard.UpStr type assertion failed")
								}
								if hemi0.Which() == hemi1.Which() {
									doMinRec = false
								}
							}
							if doMinRec {
								xbeam, ybeam := 0., 0.
								x, y, z := reconstruction.Minimal(ch0, ch1, xbeam, ybeam)
								minrec = append(minrec, XYZ{X: x, Y: y, Z: z})
								dqplots.HMinRecX.Fill(x, 1)
								dqplots.HMinRecY.Fill(y, 1)
								dqplots.HMinRecZ.Fill(z, 1)

								if doPedestal {
									_, _, T30_0, _, _, _ := pulsesWithSignal[0].CalcRisingFront(true)
									_, _, T30_1, _, _, _ := pulsesWithSignal[1].CalcRisingFront(true)
									if T30_0 != 0 && T30_1 != 0 {
										dqplots.DeltaT30.Fill(T30_0-T30_1, 1)
									}
									if treeMult2 != nil {
										treeMult2.Fill(run, uint32(event.ID), pulsesWithSignal[0], pulsesWithSignal[1])
									}
								}
								dqplots.ChargeCorrelation.Fill(pulsesWithSignal[0].Charg/1e6, pulsesWithSignal[1].Charg/1e6)
							}
						}
						if *iEvent%*monFreq == 0 {
							// Webserver data

							var qs Quartets
							sampFreq := 5
							if *monLight {
								sampFreq = 20
							}
							for iq := 0; iq < len(qs); iq++ {
								qs[iq][0] = GetMonData(sampFreq, event.Clusters[iq].Pulses[0])
								qs[iq][1] = GetMonData(sampFreq, event.Clusters[iq].Pulses[1])
								qs[iq][2] = GetMonData(sampFreq, event.Clusters[iq].Pulses[2])
								qs[iq][3] = GetMonData(sampFreq, event.Clusters[iq].Pulses[3])
							}

							// Make frequency histo plot
							tpfreq := dqplots.MakeFreqTiledPlot()
							freqhsvg := utils.RenderSVG(tpfreq, 50, 10)

							chargeLsvg := ""
							chargeRsvg := ""
							hvsvg := ""
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

								// Read HV
								hvvals := &HVvalues{}
								if hvexec != nil && *iEvent%(*monFreq**hvMonDegrad) == 0 {
									hvvals = NewHVvalues(hvexec)
									for iHVCard := 0; iHVCard < 4; iHVCard++ {
										for iHVChannel := 0; iHVChannel < 16; iHVChannel++ {
											dqplots.AddHVPoint(iHVCard, iHVChannel, float64(event.ID), hvvals[iHVCard][iHVChannel].HV)
										}
									}
								}
								hvTiled := dqplots.MakeHVTiledPlot()
								hvsvg = utils.RenderSVG(hvTiled, 45, 30)

								tpMinRec1D := dqplots.MakeMinRec1DTiledPlot()
								minrec1Dsvg = utils.RenderSVG(tpMinRec1D, 30, 13)
							}

							stop := time.Now()
							duration := stop.Sub(start).Seconds()
							start = stop
							time := stop.Sub(startabs).Seconds()
							freq := float64(noEventsForMon) / duration
							if *iEvent == 0 {
								freq = 0
							}

							// Make DeltaT30 plot
							pDeltaT30, err := hplot.New()
							if err != nil {
								panic(err)
							}
							pDeltaT30.X.Label.Text = "Delta T30 (ns)"
							pDeltaT30.Y.Label.Text = "No entries"
							pDeltaT30.X.Tick.Marker = &hplot.FreqTicks{N: 61, Freq: 5}
							hpDeltaT30, err := hplot.NewH1D(dqplots.DeltaT30)
							if err != nil {
								panic(err)
							}
							pDeltaT30.Add(hpDeltaT30)
							pDeltaT30.Add(hplot.NewGrid())
							DeltaT30svg := utils.RenderSVG(pDeltaT30, 15, 7)

							// Make ChargeCorrelation plot

							pChargeCorrelation, err := hplot.New()
							if err != nil {
								panic(err)
							}
							pChargeCorrelation.X.Label.Text = "Charge 0"
							pChargeCorrelation.Y.Label.Text = "Charge 1"
							pChargeCorrelation.X.Tick.Marker = &hplot.FreqTicks{N: 11, Freq: 2}
							pChargeCorrelation.Y.Tick.Marker = &hplot.FreqTicks{N: 11, Freq: 2}
							pChargeCorrelation.X.Min = dqplots.ChargeCorrelation.LowX
							pChargeCorrelation.Y.Min = dqplots.ChargeCorrelation.LowY
							pChargeCorrelation.X.Max = dqplots.ChargeCorrelation.HighX
							pChargeCorrelation.Y.Max = dqplots.ChargeCorrelation.HighY
							hpChargeCorrelation := plotter.NewHeatMap(dqplots.ChargeCorrelation, palette.Heat(200, 1))
							if err != nil {
								panic(err)
							}
							pChargeCorrelation.Add(hpChargeCorrelation)
							pChargeCorrelation.Add(hplot.NewGrid())
							ChargeCorrelationsvg := ""
							if false { //*iEvent > 0 {
								ChargeCorrelationsvg = utils.RenderSVG(pChargeCorrelation, 13, 13)
							}

							// send to channel
							if float64(len(datac)) >= 0.6*float64(datacsize) {
								fmt.Printf("Warning: monitoring buffer filled at more than 60 percent (len(datac) = %v, datacsize = %v)\n", len(datac), datacsize)
							}
							datac <- Data{
								EvtID:             event.ID,
								Time:              time,
								Counters:          event.Counters,
								MonBufSize:        len(datac),
								Freq:              freq,
								Qs:                qs,
								Mult:              NewH1D(dqplots.HMultiplicity),
								FreqH:             freqhsvg,
								ChargeL:           chargeLsvg,
								ChargeR:           chargeRsvg,
								HVvals:            hvsvg,
								MinRec:            minrec,
								MinRec1DDistrs:    minrec1Dsvg,
								DeltaT30:          DeltaT30svg,
								ChargeCorrelation: ChargeCorrelationsvg,
							}
							noEventsForMon = 0
							minrec = nil
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
				case true:
					fmt.Println("warning, event is corrupted and therefore not written to output file.")
					log.Fatalf(" -> quitting")
				}
			case false:
				fmt.Println("reached specified number of events, stopping.")
				if treeMult2 != nil {
					treeMult2.Close()
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

		sb, err := json.Marshal(data)
		if err != nil {
			panic(err)
		}
		fmt.Printf("len(marshaled data) = %v bytes = %v bits\n", len(sb), len(sb)*8)

		/////////////////////////////////////////////////
		err = websocket.JSON.Send(ws, data)
		if err != nil {
			log.Printf("error sending data: %v\n", err)
			return
		}
	}
}
