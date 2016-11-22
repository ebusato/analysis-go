// Remains to be done:
//   - manage runs.csv and runs_test.csv
//   - add header to binary files and code that fills and updates it properly on godaq side
//   - put same stuff on monitoring as there was in the VME version (multiplicity, minimal reconstruction, hv, ...)
//   - reintroduce in a proper way user commands to pause/resume/stop run and monitoring
//   - add creation of root tree, as in dpga/godaq

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
	"github.com/go-hep/hplot"
	"github.com/toqueteos/webbrowser"

	"golang.org/x/net/websocket"

	"gitlab.in2p3.fr/avirm/analysis-go/applyCorrCalib"
	"gitlab.in2p3.fr/avirm/analysis-go/dpga/dpgadetector"
	"gitlab.in2p3.fr/avirm/analysis-go/dpga/dq"
	"gitlab.in2p3.fr/avirm/analysis-go/dpgatca/rw"
	"gitlab.in2p3.fr/avirm/analysis-go/event"
	"gitlab.in2p3.fr/avirm/analysis-go/pulse"
	"gitlab.in2p3.fr/avirm/analysis-go/utils"
)

type frameSliceType []*rw.Frame

var (
	datacsize int = 10
	datac         = make(chan Data, datacsize)

	frameSliceChan = make(chan struct {
		uint    // event index
		float64 // event processing frequency
		frameSliceType
	}, 10)
	evtChan = make(chan struct {
		float64
		*event.Event
	})
	timeStamps      [20]uint64 // set of last 20 timestamps
	allClosedEvents []uint64   // set of timestamps of all closed events

	noEventsForMon uint

	cpuprof     = flag.String("cpuprof", "", "Name of file for CPU profiling")
	noEventsTot = flag.Uint("n", 100000, "Number of events")
	outfileName = flag.String("o", "", "Name of the output file. If not specified, setting it automatically using the following syntax: runXXX.bin (where XXX is the run number)")
	ip          = flag.String("ip", "192.168.100.11", "IP address")
	port        = flag.String("p", "6000", "Port number")
	monFreq     = flag.Int("mf", 1, "Monitoring frequency")
	monLight    = flag.Bool("monlight", false, "If set, the program performs a light monitoring, removing some plots")
	frameFreq   = flag.Int("ff", 1000, "Frame printing frequency")
	evtFreq     = flag.Uint("ef", 1000, "Event printing frequency")
	st          = flag.Bool("st", false, "If set, server start time is used rather than client's one")
	debug       = flag.Bool("d", false, "If set, debugging informations are printed")
	webad       = flag.String("webad", ":5555", "server address:port")
	nobro       = flag.Bool("nobro", false, "If set, no webbrowser are open (it's up to the user to open it with the right address)")
	sigthres    = flag.Uint("sigthres", 800, "Value above which a pulse is considered to have signal")
	notree      = flag.Bool("notree", false, "If set, no root tree is produced")
	test        = flag.Bool("test", false,
		"If set, update runs_test.csv rather than the \"official\" runs.csv file and name by default the output binary file using the following scheme: runXXX_test.bin")
	refplots = flag.String("ref", "",
		"Name of the file containing reference plots. If empty, no reference plots are overlayed")
	hvMonDegrad = flag.Uint("hvmondeg", 100, "HV monitoring frequency degradation factor")
	comment     = flag.String("c", "None", "Comment to be put in runs csv file")
	distr       = flag.String("distr", "charge", "Possible values: charge (default), ampl, energy")
	ped         = flag.String("ped", "", "Name of the csv file containing pedestal constants. If not set, pedestal corrections are not applied.")
	tdo         = flag.String("tdo", "", "Name of the csv file containing time dependent offsets. If not set, time dependent offsets are not applied. Relevant only when ped!=\"\".")
	en          = flag.String("en", "", "Name of the csv file containing energy calibration constants. If not set, energy calibration is not applied.")
	con         = flag.String("con", "none", "Connection type. Possible values: udp, tcp, none. If none, infileName is read")
	infileName  = flag.String("i", "", "Name of the input file.")
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

/*
func NewH1D(h *hbook.H1D) H1D {
	var hist H1D
	nbins := h.Len()
	for i := 0; i < nbins; i++ {
		x, y := h.XY(i)
		hist = append(hist, XY{X: x, Y: y})
	}
	return hist
}
*/

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
	EvtID           uint     `json:"evt"`             // event id (64 bits a priori)
	MonBufSize      int      `json:"monbufsize"`      // monitoring channel buffer size
	Freq            float64  `json:"freq"`            // number of events processed per second (64 bits)
	UDPPayloadSizes []int    `json:"udppayloadsizes"` // UDP frame payload sizes in octets for all frames making events
	Qs              Quartets `json:"quartets"`        // (30689280 bits)
	Mult            string   `json:"mult"`            // multiplicity of pulses
	FreqH           string   `json:"freqh"`           // frequency histogram
	ChargeL         string   `json:"chargel"`         // charge histograms for left hemisphere
	ChargeR         string   `json:"charger"`         // charge histograms for right hemisphere
	DeltaT30        string   `json:"deltat30"`        // distribution of the difference of T30
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

/*
func UDPConn(p *string) *net.UDPConn {
	fmt.Println("addr", *ip+":"+*p)
	// 	conn, err := net.Dial("tcp", *ip+":"+*p)

	RemoteAddr, err := net.ResolveUDPAddr("udp", *ip+":"+*p)
	fmt.Println("client RemoteAddr =", RemoteAddr.IP, RemoteAddr.Port, RemoteAddr.Zone)
	conn, err := net.DialUDP("udp", nil, RemoteAddr)
	if err != nil {
		return nil
	}
	return conn
}*/

func UDPConn(p *string) *net.UDPConn {
	fmt.Println("addr", *ip+":"+*p)
	locAddr, err := net.ResolveUDPAddr("udp", *ip+":"+*p)
	conn, err := net.ListenUDP("udp", locAddr)
	if err != nil {
		return nil
	}
	return conn
}

type Reader struct {
	conn *net.UDPConn
}

func NewReader(conn *net.UDPConn) *Reader {
	return &Reader{conn: conn}
}

func (r *Reader) Read(p []byte) (n int, err error) {
	//n, _, err = r.conn.ReadFromUDP(p)
	n, err = r.conn.Read(p)
	return
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

	// 	var conn net.Conn
	// 	conn := UDPConn(port)
	// 	conn := TCPConn(port)

	var r *rw.Reader
	var err error
	switch *con {
	case "udp":
		conn := UDPConn(port)
		for i := 0; conn == nil; i++ {
			newportu, err := strconv.ParseUint(*port, 10, 64)
			if err != nil {
				panic(err)
			}
			newportu += 1
			newport := strconv.FormatUint(newportu, 10)
			fmt.Printf("Port %v not responding, trying %v\n", *port, newport)
			*port = newport
			conn = UDPConn(port)
			if i >= 5 {
				log.Fatalf("Cannot find port to connect to server")
			}
		}
		conn.SetReadBuffer(216320) // not sure what is the unit of the argument
		//conn.SetReadBuffer(216320)
		//conn.Write([]byte("Hello from client"))
		r, err = rw.NewReader(bufio.NewReader(NewReader(conn)))
		r.ReadMode = rw.UDPHalfDRS
	case "tcp":
		conn := TCPConn(port)
		for i := 0; conn == nil; i++ {
			newportu, err := strconv.ParseUint(*port, 10, 64)
			if err != nil {
				panic(err)
			}
			newportu += 1
			newport := strconv.FormatUint(newportu, 10)
			fmt.Printf("Port %v not responding, trying %v\n", *port, newport)
			*port = newport
			conn = TCPConn(port)
			if i >= 5 {
				log.Fatalf("Cannot find port to connect to server")
			}
		}
		r, err = rw.NewReader(bufio.NewReader(conn))
		r.ReadMode = rw.Default
	case "none":
		file, err := os.Open(*infileName)
		if err != nil {
			log.Fatalf("could not create data file: %v\n", err)
		}
		defer file.Close()
		r, err = rw.NewReader(file)
		if err != nil {
			log.Fatalf("could not open stream: %v\n", err)
		}
		r.ReadMode = rw.Default
	default:
		log.Fatalf("Connection type not known")
	}

	//r, err := rw.NewReader(bufio.NewReader(*conn)) // tcp
	//r, err := rw.NewReader(bufio.NewReader(NewReader(conn))) // udp
	//r, err := rw.NewReader(NewReader(conn)) // udp
	if err != nil {
		log.Fatalf("could not open stream: %v\n", err)
	}
	r.SigThreshold = *sigthres

	var w *rw.Writer
	if *con != "none" {
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
		w = rw.NewWriter(bufiow)
		defer w.Close()
	}

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
		"WebAd": *webad,
		/*
			"RunNumber": currentRunNumber,
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
		*/
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

	go readFrames(r, w, &wg)
	go reconstructEvent(r)
	go monitor(r, w)
	go webserver()
	go command()

	wg.Wait()
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

func webserver() {
	if !*nobro {
		webbrowser.Open("http://" + *webad)
	}
	fmt.Println("Starting webserver")
	//http.HandleFunc("/", plotHandle)
	http.Handle("/", http.FileServer(http.Dir("./root-fs")))
	//http.Handle("/", http.FileServer(assetFS()))
	http.Handle("/data", websocket.Handler(dataHandler))
	err := http.ListenAndServe(*webad, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func dataHandler(ws *websocket.Conn) {
	fmt.Println("Starting dataHandler")
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

// func command(terminateRun, pauseRun chan bool) {
func command() {
	in := bufio.NewReader(os.Stdin) // to be replaced by Scanner
	for {
		word, _ := in.ReadString('\n')
		word = strings.Replace(word, "\n", "", -1)
		switch word {
		default:
			fmt.Println("command not known, what do you mean ?", word)
		}
	}
}

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
			//fmt.Println("debug GetMonData:", i, counter, x, samp.Amplitude)
			data[counter] = XY{X: x, Y: samp.Amplitude}
			counter++
		}
	}
	return data
}

func monitor(r *rw.Reader, w *rw.Writer) {
	if *ped != "" {
		dpgadetector.Det.ReadPedestalsFile(*ped)
	}
	if *tdo != "" {
		dpgadetector.Det.ReadTimeDepOffsetsFile(*tdo)
	}
	if *en != "" {
		dpgadetector.Det.ReadEnergyCalibFile(*en)
	}
	dqplots := dq.NewDQPlot()
	if *refplots != "" {
		dqplots.DQPlotRef = dq.NewDQPlotFromGob(*refplots)
	}

	for {
		evt := <-evtChan
		event := evt.Event

		switch event.IsCorrupted {
		case false:
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

			// Make multiplicity plot
			pMult, err := hplot.New()
			if err != nil {
				panic(err)
			}
			pMult.X.Label.Text = "Pulse multiplicity"
			pMult.Y.Label.Text = "No entries"
			pMult.X.Tick.Marker = &hplot.FreqTicks{N: 17, Freq: 1}
			hMult, err := hplot.NewH1D(dqplots.HMultiplicity)
			if err != nil {
				panic(err)
			}
			pMult.Add(hMult)
			pMult.Add(hplot.NewGrid())
			Multsvg := utils.RenderSVG(pMult, 15, 7)

			// send to channel
			if float64(len(datac)) >= 0.6*float64(datacsize) {
				fmt.Printf("Warning: monitoring buffer filled at more than 60 percent (len(datac) = %v, datacsize = %v)\n", len(datac), datacsize)
			}
			datac <- Data{
				EvtID:           event.ID,
				MonBufSize:      len(datac),
				Freq:            evt.float64,
				UDPPayloadSizes: event.UDPPayloadSizes,
				Qs:              qs,
				Mult:            Multsvg,
				FreqH:           freqhsvg,
				ChargeL:         chargeLsvg,
				ChargeR:         chargeRsvg,
				DeltaT30:        DeltaT30svg,
			}
		case true:
			fmt.Println("warning, event is corrupted and therefore not written to output file.")
			log.Fatalf(" -> quitting")
		}
	} // event loop
}

func makePulses(f *rw.Frame, sigThreshold uint) [4]*pulse.Pulse {
	var pulses [len(f.Data.Data)]*pulse.Pulse
	for i := range f.Data.Data {
		chanData := &f.Data.Data[i]
		channelId023 := chanData.Channel
		iChannel := uint8(channelId023 % 4)
		iHemi, iASM, iDRS, iQuartet := dpgadetector.QuartetAbsIdx60ToRelIdx(f.QuartetAbsIdx60)
		detChannel := dpgadetector.Det.Channel(iHemi, iASM, iDRS, iQuartet, iChannel)
		pul := pulse.NewPulse(detChannel)
		for j := range chanData.Amplitudes {
			ampl := float64(chanData.Amplitudes[j])
			sample := pulse.NewSample(ampl, uint16(j), float64(j)*dpgadetector.Det.SamplingFreq())
			pul.AddSample(sample, dpgadetector.Det.Capacitor(iHemi, iASM, iDRS, iQuartet, iChannel, 0), float64(sigThreshold))
		}
		pulses[i] = pul
	}
	return pulses
}

func closedEvents(framesMap map[uint64][]*rw.Frame) []uint64 {
	var closedEvts []uint64
	for ts1, _ := range framesMap {
		canBeClosed := true
		for _, ts2 := range timeStamps {
			if ts1 == ts2 {
				// event can't be closed
				canBeClosed = false
				break
			}
		}
		if canBeClosed {
			closedEvts = append(closedEvts, ts1)
		}
	}
	return closedEvts
}

func eventAlreadyClosed(timestamp uint64) bool {
	for i := len(allClosedEvents) - 1; i >= 0; i-- {
		if timestamp == allClosedEvents[i] {
			return true
		}
	}
	return false
}

func readFrames(r *rw.Reader, w *rw.Writer, wg *sync.WaitGroup) {
	defer wg.Done()
	var nframes int
	AMCFrameCounterPrev := uint32(0)
	ASMFrameCounterPrev := uint64(0)
	framesMap := make(map[uint64][]*rw.Frame)
	var noEvents uint
	start := time.Now()
	for {
		if nframes%*frameFreq == 0 {
			fmt.Printf("reading frame %v\n", nframes)
		}
		// 		if noEvents%*evtFreq == 0 {
		// 			fmt.Printf("reading event %v\n", noEvents)
		// 		}
		frame, _ := r.Frame()
		timeStamps[nframes%len(timeStamps)] = frame.TimeStamp
		framesMap[frame.TimeStamp] = append(framesMap[frame.TimeStamp], frame)

		/////////////////////////////////////////////////////////////////////////////////////
		// Sanity checks
		// 		fmt.Println("AMCFrameCounter =", frame.AMCFrameCounter)
		// 		fmt.Println("ASMFrameCounter =", frame.ASMFrameCounter)
		if nframes > 0 {
			if frame.AMCFrameCounter != AMCFrameCounterPrev+1 {
				fmt.Printf("frame.AMCFrameCounter != AMCFrameCounterPrev + 1\n")
			}
			if frame.ASMFrameCounter != ASMFrameCounterPrev+1 {
				fmt.Printf("frame.ASMFrameCounter != ASMFrameCounterPrev + 1\n")
			}
		}
		AMCFrameCounterPrev = frame.AMCFrameCounter
		ASMFrameCounterPrev = frame.ASMFrameCounter

		if r.ReadMode == rw.UDPHalfDRS && frame.UDPPayloadSize < 8230 {
			log.Printf("frame.UDPPayloadSize = %v\n", frame.UDPPayloadSize)
		}
		/////////////////////////////////////////////////////////////////////////////////////

		/////////////////////////////////////////////////////////////////////////////////////
		// Write to disk
		// 		for i := 0; i < frame.UDPPayloadSize; i++ {
		// 			w.writeByte(r.UDPHalfDRSBuffer[i])
		// 		}
		/////////////////////////////////////////////////////////////////////////////////////

		/////////////////////////////////////////////////////////////////////////////////////
		// The following check is possibly time consuming, consider removing it
		if eventAlreadyClosed(frame.TimeStamp) {
			log.Fatalf("Timestamp %v already sent to reconstruction\n", frame.TimeStamp)
		}
		/////////////////////////////////////////////////////////////////////////////////////

		if nframes%*monFreq == 0 {
			closedEvts := closedEvents(framesMap)
			noClosedEvts := len(closedEvts)
			noEvents += uint(noClosedEvts)
			noEventsForMon += uint(noClosedEvts)

			if noClosedEvts >= 1 {
				switch noEvents < *noEventsTot {
				case true:
					// 					if noClosedEvts >= 2 {
					// 						log.Fatalf("len(closedEvts) > 1 (len(closedEvts) = %v)\n", noClosedEvts)
					// 					}
					tsToMonitor := closedEvts[noClosedEvts-1]
					if len(framesMap[tsToMonitor]) < 2 {
						fmt.Printf("len(framesMap[tsToMonitor] < 2)\n")
					}
					stop := time.Now()
					duration := stop.Sub(start).Seconds()
					start = stop
					freq := float64(noEventsForMon) / duration
					frameSliceChan <- struct {
						uint
						float64
						frameSliceType
					}{noEvents, freq, framesMap[tsToMonitor]}
					noEventsForMon = 0

				case false:
					fmt.Println("reached specified number of events, stopping.")
					return
				}
			}
			for _, ts := range closedEvts {
				delete(framesMap, ts)
			}

			allClosedEvents = append(allClosedEvents, closedEvts...)
		}

		nframes++
	} //frame loop
}

func reconstructEvent(r *rw.Reader) {
	for {
		frameSlice := <-frameSliceChan
		firstFrame := true
		start := time.Now()
		evt := event.NewEvent(dpgadetector.Det.NoClusters())
		evt.ID = frameSlice.uint
		// build event from slice of frames
		for _, frame := range frameSlice.frameSliceType {
			if !firstFrame && frame.TimeStamp != evt.TimeStamp {
				log.Fatalf("Time stamps are not all equal to the same value. This should never happen !\n")
			}
			evt.TimeStamp = frame.TimeStamp
			pulses := makePulses(frame, r.SigThreshold)
			evt.Clusters[frame.QuartetAbsIdx60].Pulses[0] = *pulses[0]
			evt.Clusters[frame.QuartetAbsIdx60].Pulses[1] = *pulses[1]
			evt.Clusters[frame.QuartetAbsIdx60].Pulses[2] = *pulses[2]
			evt.Clusters[frame.QuartetAbsIdx60].Pulses[3] = *pulses[3]
			evt.UDPPayloadSizes = append(evt.UDPPayloadSizes, frame.UDPPayloadSize)
			firstFrame = false
		}
		stop := time.Now()
		duration := stop.Sub(start).Seconds()
		fmt.Printf("Time for event making = %v (no frames = %v)\n", duration, len(frameSlice.frameSliceType))
		evtChan <- struct {
			float64
			*event.Event
		}{frameSlice.float64, evt}
	}
}
