package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"text/template"
	"time"

	"github.com/go-hep/csvutil"
	"github.com/go-hep/hbook"
	"github.com/toqueteos/webbrowser"

	"golang.org/x/net/websocket"

	"gitlab.in2p3.fr/avirm/analysis-go/dpga/dpgadetector"
	"gitlab.in2p3.fr/avirm/analysis-go/dpga/dq"
	"gitlab.in2p3.fr/avirm/analysis-go/dpga/rw"
	"gitlab.in2p3.fr/avirm/analysis-go/event"
	"gitlab.in2p3.fr/avirm/analysis-go/pulse"
	"gitlab.in2p3.fr/avirm/analysis-go/utils"
)

var (
	datac       = make(chan Data, 10)
	hdrType     = rw.HeaderCAL
	noEvents    = flag.Uint("n", 100000, "Number of events")
	outfileName = flag.String("o", "out.bin", "Name of the output file")
	ip          = flag.String("ip", "192.168.100.11", "IP address")
	port        = flag.String("p", "1024", "Port number")
	monFreq     = flag.Uint("mf", 50, "Monitoring frequency")
	evtFreq     = flag.Uint("ef", 100, "Event printing frequency")
	st          = flag.Bool("st", false, "If set, server start time is used rather than client's one")
	debug       = flag.Bool("d", false, "If set, debugging informations are printed")
	webad       = flag.String("webad", ":5555", "server address:port")
	nobro       = flag.Bool("nobro", false, "If set, no webbrowser are open (it's up to the user to open it with the right address)")
	sleep       = flag.Bool("s", false, "If set, sleep a bit between events")
	runcsvtest  = flag.Bool("runcsvtest", false, "If set, update runs_test.csv rather than the \"official\" runs.csv file")
)

type XY struct {
	X float64
	Y float64
}

type Pulse []XY

type Quartet [4]Pulse

type Quartets [60]Quartet

type H1D []XY // X is the bin center and Y the bin content

func NewH1D(h *hbook.H1D) H1D {
	var hist H1D
	nbins := h.Len()
	for i := 0; i < nbins; i++ {
		x, y := h.XY(i)
		hist = append(hist, XY{X: x, Y: y})
	}
	return hist
}

type Data struct {
	Time    float64  `json:"time"` // time at which monitoring data are taken
	Freq    float64  `json:"freq"` // number of events processed per second
	Qs      Quartets `json:"quartets"`
	Mult    H1D      `json:"mult"`    // multiplicity of pulses
	FreqH   string   `json:"freqh"`   // frequency histogram
	ChargeL string   `json:"chargel"` // charge histograms for left hemisphere
	ChargeR string   `json:"charger"` // charge histograms for right hemisphere
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

	// Writer for binary file
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

	// determine run number
	var runCSVFileName string
	switch *runcsvtest {
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

	// web address handling
	webadSlice := strings.Split(*webad, ":")
	if webadSlice[0] == "" {
		webadSlice[0] = getHostIP()
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

	terminateStream := make(chan bool)
	commandIsEnded := make(chan bool)
	cevent := make(chan event.Event)

	if *debug {
		r.Debug = true
	}

	iEvent := uint(0)
	go control(terminateStream, commandIsEnded)
	go stream(terminateStream, cevent, r, w, &iEvent, &wg)
	go command(commandIsEnded)
	go webserver()
	//go monitoring(cevent)

	wg.Wait()

	// Update header
	//bufiow.Flush()
	timeStop := uint32(time.Now().Unix())
	noEvents := uint32(iEvent)
	updateHeader(filew, 16, timeStop)
	updateHeader(filew, 20, noEvents)

	// Dump run info in csv. Only relevant when ran on DAQ PC, where the csv file is present.
	updateRunsCSV(runCSVFileName, currentRunNumber, timeStop, noEvents, hdr)
	updateHeader(filew, 4, currentRunNumber)

	err = os.Rename(*outfileName, "run"+strconv.FormatUint(uint64(currentRunNumber), 10)+".bin")
}

func updateHeader(f *os.File, offset int64, val uint32) {
	var buf [4]byte
	binary.BigEndian.PutUint32(buf[:], val)
	f.WriteAt(buf[:], offset)
}

type RunsCSV struct {
	RunNumber uint32
	NoEvents  uint32
	ExecDir   string
	StartTime string
	StopTime  string
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

func updateRunsCSV(fileName string, runNumber uint32, timeStop uint32, noEvents uint32, hdr *rw.Header) {
	// Determine working directory
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Open csv file in append mode
	tbl, err := csvutil.Append(fileName)
	if err != nil {
		log.Fatalf("could not create dpgageom.csv: %v\n", err)
	}
	defer tbl.Close()
	tbl.Writer.Comma = ' '

	data := RunsCSV{
		RunNumber: runNumber,
		NoEvents:  noEvents,
		ExecDir:   pwd,
		StartTime: time.Unix(int64(hdr.TimeStart), 0).Format(time.UnixDate),
		StopTime:  time.Unix(int64(timeStop), 0).Format(time.UnixDate),
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

func control(terminateStream chan bool, commandIsEnded chan bool) {
	for {
		select {
		case <-commandIsEnded:
			fmt.Printf("command is ended, terminating stream.\n")
			terminateStream <- true
		default:
			// do nothing
		}
	}
}

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

func stream(terminateStream chan bool, cevent chan event.Event, r *rw.Reader, w *rw.Writer, iEvent *uint, wg *sync.WaitGroup) {
	defer wg.Done()
	//nFrames := uint(0)
	noEventsForMon := uint64(0)
	hMult := hbook.NewH1D(8, -0.5, 7.5)
	dqplots := dq.NewDQPlot()
	start := time.Now()
	startabs := start
	for {
		select {
		case <-terminateStream:
			*noEvents = *iEvent + 1
			fmt.Printf("terminating stream for total number of events = %v.\n", *noEvents)
		default:
			switch *iEvent < *noEvents {
			case true:
				if *iEvent%*evtFreq == 0 {
					fmt.Printf("event %v\n", *iEvent)
				}
				event, status := r.ReadNextEvent()
				if status == false {
					panic("error: status is false\n")
				}
				switch event.IsCorrupted {
				case false:
					//event.Print(true, false)
					w.Event(event)
					hMult.Fill(float64(event.Multiplicity()), 1)
					dqplots.FillHistos(event)
					if *iEvent%*monFreq == 0 {
						//cevent <- *event
						// Webserver data
						stop := time.Now()
						duration := stop.Sub(start).Seconds()
						start = stop
						time := stop.Sub(startabs).Seconds()
						freq := float64(noEventsForMon) / duration
						if *iEvent == 0 {
							freq = 0
						}

						var qs Quartets
						for iq := 0; iq < len(qs); iq++ {
							qs[iq][0] = GetMonData(r.NoSamples(), event.Clusters[iq].Pulses[0])
							qs[iq][1] = GetMonData(r.NoSamples(), event.Clusters[iq].Pulses[1])
							qs[iq][2] = GetMonData(r.NoSamples(), event.Clusters[iq].Pulses[2])
							qs[iq][3] = GetMonData(r.NoSamples(), event.Clusters[iq].Pulses[3])
						}

						//fmt.Println("data:", time, noEventsForMon, duration, freq)

						// Make frequency histo plot
						tpfreq := dqplots.MakeFreqTiledPlot()
						freqhsvg := utils.RenderSVG(tpfreq, 50, 10)

						// Make charge distrib histo plot
						tpchargeL := dqplots.MakeChargeAmplTiledPlot(dq.Charge, dpgadetector.Left)
						tpchargeR := dqplots.MakeChargeAmplTiledPlot(dq.Charge, dpgadetector.Right)
						chargeLsvg := utils.RenderSVG(tpchargeL, 25, 25)
						chargeRsvg := utils.RenderSVG(tpchargeR, 25, 25)
						// send to channel
						datac <- Data{
							Time:    time,
							Freq:    freq,
							Qs:      qs,
							Mult:    NewH1D(hMult),
							FreqH:   freqhsvg,
							ChargeL: chargeLsvg,
							ChargeR: chargeRsvg,
						}
						noEventsForMon = 0
					}
					*iEvent++
					noEventsForMon++
					if *sleep {
						time.Sleep(1 * time.Second)
					}
				case true:
					fmt.Println("warning, event is corrupted and therefore not written to output file.")
					log.Fatalf(" -> quitting")
				}
			case false:
				fmt.Println("reached specified number of events, stopping.")
				return
			}
		}
	}
}

func command(commandIsEnded chan bool) {
	for {
		in := bufio.NewReader(os.Stdin) // to be replaced by Scanner
		word, _ := in.ReadString('\n')
		word = strings.Replace(word, "\n", "", -1)
		switch word {
		default:
			fmt.Println("command not known, what do you mean ?", word)
		case "stop":
			fmt.Println("stopping run")
			commandIsEnded <- true
		}
	}
}

func monitoring(cevent chan event.Event) {
	for {
		//fmt.Println("receiving from cframe1")
		event := <-cevent
		event.Clusters[0].PlotPulses(0, pulse.XaxisIndex, false)
	}
}

func dataHandler(ws *websocket.Conn) {
	for data := range datac {
		err := websocket.JSON.Send(ws, data)
		if err != nil {
			log.Printf("error sending data: %v\n", err)
			return
		}
	}
}

func getHostIP() string {
	host, err := os.Hostname()
	if err != nil {
		log.Fatalf("could not retrieve hostname: %v\n", err)
	}

	addrs, err := net.LookupIP(host)
	if err != nil {
		log.Fatalf("could not lookup hostname IP: %v\n", err)
	}

	for _, addr := range addrs {
		ipv4 := addr.To4()
		if ipv4 == nil {
			continue
		}
		return ipv4.String()
	}

	log.Fatalf("could not infer host IP")
	return ""
}
