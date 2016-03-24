package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/toqueteos/webbrowser"

	"golang.org/x/net/websocket"

	"gitlab.in2p3.fr/avirm/analysis-go/dpga/rw"
	"gitlab.in2p3.fr/avirm/analysis-go/event"
	"gitlab.in2p3.fr/avirm/analysis-go/pulse"
)

var (
	datac = make(chan Data)
)

type XY struct {
	X float64
	Y float64
}

type Pulse []XY

type Quartet [4]Pulse

type Quartets [60]Quartet

type Data struct {
	Time float64  `json:"time"` // time at which monitoring data are taken
	Freq float64  `json:"freq"` // number of events processed per second
	Qs   Quartets `json:"quartets"`
}

func main() {
	log.SetFlags(log.Llongfile | log.LstdFlags)

	var (
		noEvents    = flag.Uint("n", 100000, "Number of events")
		outfileName = flag.String("o", "out.bin", "Name of the output file")
		ip          = flag.String("ip", "192.168.100.11", "IP address")
		port        = flag.String("p", "1024", "Port number")
		monFreq     = flag.Uint("mf", 50, "Monitoring frequency")
		evtFreq     = flag.Uint("ef", 100, "Event printing frequency")
		debug       = flag.Bool("d", false, "If set, debugging informations are printed")
		webad       = flag.String("webad", "localhost:5555", "server address:port")
	)

	flag.Parse()

	// Reader
	laddr, err := net.ResolveTCPAddr("tcp", *ip+":"+*port)
	if err != nil {
		log.Fatal(err)
	}
	tcp, err := net.DialTCP("tcp", nil, laddr)
	if err != nil {
		log.Fatal(err)
	}

	r, err := rw.NewReader(bufio.NewReader(tcp), rw.HeaderCAL)
	if err != nil {
		log.Fatalf("could not open stream: %v\n", err)
	}

	// Writer
	filew, err := os.Create(*outfileName)
	if err != nil {
		log.Fatalf("could not create data file: %v\n", err)
	}
	defer filew.Close()

	w := rw.NewWriter(bufio.NewWriter(filew))
	if err != nil {
		log.Fatalf("could not open file: %v\n", err)
	}
	defer w.Close()

	// Start reading TCP stream
	hdr := r.Header()
	hdr.Print()

	err = w.Header(hdr)
	if err != nil {
		log.Fatalf("error writing header: %v\n", err)
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

	go control(terminateStream, commandIsEnded)
	go stream(terminateStream, cevent, r, w, noEvents, monFreq, evtFreq, &wg)
	go command(commandIsEnded)
	go webserver(webad)
	//go monitoring(cevent)

	wg.Wait()
}

func webserver(webad *string) {
	webbrowser.Open("http://" + *webad)
	http.HandleFunc("/", plotHandle)
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

func GetMonData(pulse pulse.Pulse) []XY {
	data := make([]XY, pulse.NoSamples())
	for i := range data {
		data[i] = XY{
			X: float64(pulse.Samples[i].Index),
			Y: pulse.Samples[i].Amplitude,
		}
	}
	return data
}

func stream(terminateStream chan bool, cevent chan event.Event, r *rw.Reader, w *rw.Writer, noEvents *uint, monFreq *uint, evtFreq *uint, wg *sync.WaitGroup) {
	defer wg.Done()
	//nFrames := uint(0)
	iEvent := uint(0)
	noEventsForMon := uint64(0)
	start := time.Now()
	startabs := start
	for {
		select {
		case <-terminateStream:
			*noEvents = iEvent + 1
			fmt.Printf("terminating stream for total number of events = %v.\n", *noEvents)
		default:
			switch iEvent < *noEvents {
			case true:
				if iEvent%*evtFreq == 0 {
					fmt.Printf("event %v\n", iEvent)
				}
				event, status := r.ReadNextEvent()
				if status == false {
					panic("error: status is false\n")
				}
				switch event.IsCorrupted {
				case false:
					w.Event(event)
					if iEvent%*monFreq == 0 {
						//cevent <- *event
						// Webserver data
						stop := time.Now()
						duration := stop.Sub(start).Seconds()
						start = stop
						time := stop.Sub(startabs).Seconds()
						freq := float64(noEventsForMon) / duration
						if iEvent == 0 {
							freq = 0
						}

						var qs Quartets
						for iq := 0; iq < len(qs); iq++ {
							qs[iq][0] = GetMonData(event.Clusters[iq].Pulses[0])
							qs[iq][1] = GetMonData(event.Clusters[iq].Pulses[1])
							qs[iq][2] = GetMonData(event.Clusters[iq].Pulses[2])
							qs[iq][3] = GetMonData(event.Clusters[iq].Pulses[3])
						}

						//fmt.Println("data:", time, noEventsForMon, duration, freq)
						datac <- Data{
							Time: time,
							Freq: freq,
							Qs:   qs,
						}
						noEventsForMon = 0
					}
					iEvent++
					noEventsForMon++
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

func plotHandle(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, page)
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

const page = `
<html>
	<head>
		<title>DPGA monitoring</title>
		<script src="//cdnjs.cloudflare.com/ajax/libs/jquery/2.0.3/jquery.min.js"></script>
		<script src="//cdnjs.cloudflare.com/ajax/libs/flot/0.8.3/jquery.flot.min.js"></script>
		<script type="text/javascript">
		var sock = null;
		
		function options(bkgcolor) {
			this.series = {stack: true};
			//this.yaxis = {min: 0, max: 4096};
			this.xaxis = {tickDecimals: 0};
			this.grid = {backgroundColor: bkgcolor}
		}
		
		function plot(label, data, color) {
			this.label = label;
			this.data = data;
			this.color = color;
		}
		
		var freqplot = new plot("frequency (Hz)", [], 'orange')
		
		var colors = ['blue', 'red', 'black', 'green']
		
		var Nquartets = 60
		var quartetplots = []
		var Nplots = 4
			
		for (var iq = 0; iq < Nquartets; iq += 1) {
			var pulseplots = []
			for (var ip = 0; ip < Nplots; ip += 1) {
				//pulseplots.push(new plot("q"+iq+"c"+ip, [], colors[ip]))
				pulseplots.push(new plot("", [], colors[ip]))
			}
			quartetplots.push(pulseplots)
		}
		
		function clearplots() {
			for (var iq = 0; iq < Nquartets; iq += 1) {
				for (var ip = 0; ip < Nplots; ip += 1) {
					quartetplots[iq][ip].data = []
				}
			}
		}
		
		function update() {
			var freq = $.plot("#my-freq-plot", [freqplot]);
			freq.setupGrid(); // needed as x-axis changes
			freq.draw();
			for (var i = 0; i < Nquartets; i++) {
				
				if (i < Nquartets/2) {
					var p = $.plot("#my-q"+i+"-plot", quartetplots[i], new options('yellow'));
					p.setupGrid();
					p.draw();
				}
				else {
					var p = $.plot("#my-q"+i+"-plot", quartetplots[i], new options('#ADD8E6'));
					p.setupGrid();
					p.draw();
				}	
			}
		};

		window.onload = function() {
			sock = new WebSocket("ws://localhost:5555/data");

			sock.onmessage = function(event) {
				var data = JSON.parse(event.data);
				console.log("data: "+JSON.stringify(data));
				if (data.freq != 0) {	
					freqplot.data.push([data.time, data.freq])
				}
				for (var iq = 0; iq < Nquartets; iq += 1) {
					for (var ip = 0; ip < Nplots; ip += 1) {
						for (var is = 0; is < 999; is += 1) {
							quartetplots[iq][ip].data.push([data.quartets[iq][ip][is].X, data.quartets[iq][ip][is].Y]);
						}
					}
				}
				update();
				clearplots();
			};
		};
		
		</script>

		<style>
		.my-plot-stylefreq {
			width: 250px;
			height: 200px;
			font-size: 14px;
			line-height: 1.2em;
		}
		.my-plot-style {
			width: 250px;
			height: 75px;
			font-size: 14px;
			line-height: 1.2em;
		}
		</style>
	</head>

	<body>
		<div id="header">
			<h2>DPGA monitoring</h2>
		</div>
		<div id="my-freq-plot" class="my-plot-stylefreq"></div>
		<hr>
		<table>
			<tr>
				<td><div id="my-q0-plot" class="my-plot-style"></div></td>
				<td><div id="my-q1-plot" class="my-plot-style"></div></td>
				<td><div id="my-q2-plot" class="my-plot-style"></div></td>
				<td><div id="my-q3-plot" class="my-plot-style"></div></td>
				<td><div id="my-q4-plot" class="my-plot-style"></div></td>
			</tr>
			<tr>
				<td><div id="my-q5-plot" class="my-plot-style"></div></td>
				<td><div id="my-q6-plot" class="my-plot-style"></div></td>
				<td><div id="my-q7-plot" class="my-plot-style"></div></td>
				<td><div id="my-q8-plot" class="my-plot-style"></div></td>
				<td><div id="my-q9-plot" class="my-plot-style"></div></td>
			</tr>
			<tr>
				<td><div id="my-q10-plot" class="my-plot-style"></div></td>
				<td><div id="my-q11-plot" class="my-plot-style"></div></td>
				<td><div id="my-q12-plot" class="my-plot-style"></div></td>
				<td><div id="my-q13-plot" class="my-plot-style"></div></td>
				<td><div id="my-q14-plot" class="my-plot-style"></div></td>
			</tr>
			<tr>
				<td><div id="my-q15-plot" class="my-plot-style"></div></td>
				<td><div id="my-q16-plot" class="my-plot-style"></div></td>
				<td><div id="my-q17-plot" class="my-plot-style"></div></td>
				<td><div id="my-q18-plot" class="my-plot-style"></div></td>
				<td><div id="my-q19-plot" class="my-plot-style"></div></td>
			</tr>
			<tr>
				<td><div id="my-q20-plot" class="my-plot-style"></div></td>
				<td><div id="my-q21-plot" class="my-plot-style"></div></td>
				<td><div id="my-q22-plot" class="my-plot-style"></div></td>
				<td><div id="my-q23-plot" class="my-plot-style"></div></td>
				<td><div id="my-q24-plot" class="my-plot-style"></div></td>
			</tr>
			<tr>
				<td><div id="my-q25-plot" class="my-plot-style"></div></td>
				<td><div id="my-q26-plot" class="my-plot-style"></div></td>
				<td><div id="my-q27-plot" class="my-plot-style"></div></td>
				<td><div id="my-q28-plot" class="my-plot-style"></div></td>
				<td><div id="my-q29-plot" class="my-plot-style"></div></td>
			</tr>
			<tr>
				<td><div id="my-q30-plot" class="my-plot-style"></div></td>
				<td><div id="my-q31-plot" class="my-plot-style"></div></td>
				<td><div id="my-q32-plot" class="my-plot-style"></div></td>
				<td><div id="my-q33-plot" class="my-plot-style"></div></td>
				<td><div id="my-q34-plot" class="my-plot-style"></div></td>
			</tr>
			<tr>
				<td><div id="my-q35-plot" class="my-plot-style"></div></td>
				<td><div id="my-q36-plot" class="my-plot-style"></div></td>
				<td><div id="my-q37-plot" class="my-plot-style"></div></td>
				<td><div id="my-q38-plot" class="my-plot-style"></div></td>
				<td><div id="my-q39-plot" class="my-plot-style"></div></td>
			</tr>
			<tr>
				<td><div id="my-q40-plot" class="my-plot-style"></div></td>
				<td><div id="my-q41-plot" class="my-plot-style"></div></td>
				<td><div id="my-q42-plot" class="my-plot-style"></div></td>
				<td><div id="my-q43-plot" class="my-plot-style"></div></td>
				<td><div id="my-q44-plot" class="my-plot-style"></div></td>
			</tr>
			<tr>
				<td><div id="my-q45-plot" class="my-plot-style"></div></td>
				<td><div id="my-q46-plot" class="my-plot-style"></div></td>
				<td><div id="my-q47-plot" class="my-plot-style"></div></td>
				<td><div id="my-q48-plot" class="my-plot-style"></div></td>
				<td><div id="my-q49-plot" class="my-plot-style"></div></td>
			</tr>
			<tr>
				<td><div id="my-q50-plot" class="my-plot-style"></div></td>
				<td><div id="my-q51-plot" class="my-plot-style"></div></td>
				<td><div id="my-q52-plot" class="my-plot-style"></div></td>
				<td><div id="my-q53-plot" class="my-plot-style"></div></td>
				<td><div id="my-q54-plot" class="my-plot-style"></div></td>
			</tr>
			<tr>
				<td><div id="my-q55-plot" class="my-plot-style"></div></td>
				<td><div id="my-q56-plot" class="my-plot-style"></div></td>
				<td><div id="my-q57-plot" class="my-plot-style"></div></td>
				<td><div id="my-q58-plot" class="my-plot-style"></div></td>
				<td><div id="my-q59-plot" class="my-plot-style"></div></td>
			</tr>
		</table>
	</body>
</html>
`
