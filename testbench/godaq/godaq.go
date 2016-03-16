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

	"gitlab.in2p3.fr/avirm/analysis-go/event"
	"gitlab.in2p3.fr/avirm/analysis-go/pulse"
	"gitlab.in2p3.fr/avirm/analysis-go/testbench/rw"
)

var (
	datac = make(chan Data)
)

// type Data struct {
// 	X    float64      `json:"x"`
// 	Sin  float64      `json:"sin"`
// 	Cos  float64      `json:"cos"`
// 	SinF [][2]float64 `json:"sinf"`
// }

type Data struct {
	Time   float64      `json:"time"` // time at which monitoring data are taken
	Freq   float64      `json:"freq"` // number of events processed per second
	Pulse0 [][2]float64 `json:"pulse0"`
	Pulse1 [][2]float64 `json:"pulse1"`
	Pulse2 [][2]float64 `json:"pulse2"`
	Pulse3 [][2]float64 `json:"pulse3"`
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

	r, err := rw.NewReader(bufio.NewReader(tcp))
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

func GetMonData(pulse pulse.Pulse) [][2]float64 {
	data := make([][2]float64, pulse.NoSamples())
	for i := range data {
		data[i] = [2]float64{float64(pulse.Samples[i].Index), pulse.Samples[i].Amplitude}
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
		//iEvent := nFrames / 12
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
						//fmt.Println("data:", time, noEventsForMon, duration, freq)
						pulse0 := GetMonData(event.Clusters[0].Pulses[0])
						pulse1 := GetMonData(event.Clusters[0].Pulses[1])
						pulse2 := GetMonData(event.Clusters[0].Pulses[2])
						pulse3 := GetMonData(event.Clusters[0].Pulses[3])
						datac <- Data{Time: time, Freq: freq, Pulse0: pulse0, Pulse1: pulse1, Pulse2: pulse2, Pulse3: pulse3}
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
		<title>Test bench monitoring</title>
		<script src="//cdnjs.cloudflare.com/ajax/libs/jquery/2.0.3/jquery.min.js"></script>
		<script src="//cdnjs.cloudflare.com/ajax/libs/flot/0.8.3/jquery.flot.min.js"></script>
		<script type="text/javascript">
		var sock = null;
		var freqplot = {
			label: "frequency (Hz)",
			data: [],
		};
		var pulse0plot = {
			label: "channel 0 (ampl. vs sample index)",
			data: [],
		};
		var pulse1plot = {
			label: "channel 1 (ampl. vs sample index)",
			data: [],
		};
		var pulse2plot = {
			label: "channel 2 (ampl. vs sample index)",
			data: [],
		};
		var pulse3plot = {
			label: "channel 3 (ampl. vs sample index)",
			data: [],
		};
		var optionspulse0 = {
			series: {
				stack: true,
			},
			yaxis: { min: 0, max: 4096 },
			xaxis: { tickDecimals: 0 }
		};
		optionspulse0.colors = ['blue'];
		var optionspulse1 = {
			series: {
				stack: true,
			},
			yaxis: { min: 0, max: 4096 },
			xaxis: { tickDecimals: 0 }
		};
		optionspulse1.colors = ['red'];
		var optionspulse2 = {
			series: {
				stack: true,
			},
			yaxis: { min: 0, max: 4096 },
			xaxis: { tickDecimals: 0 }
		};
		optionspulse2.colors = ['black'];
		var optionspulse3 = {
			series: {
				stack: true,
			},
			yaxis: { min: 0, max: 4096 },
			xaxis: { tickDecimals: 0 }
		};
		optionspulse3.colors = ['green'];
		function update() {
			var freq = $.plot("#my-freq-plot", [freqplot]);
			freq.setupGrid(); // needed as x-axis changes
			freq.draw();
			var p0 = $.plot("#my-pulse0-plot", [pulse0plot], optionspulse0);
			p0.setupGrid();
			p0.draw();
			var p1 = $.plot("#my-pulse1-plot", [pulse1plot], optionspulse1);
			p1.setupGrid();
			p1.draw();
			var p2 = $.plot("#my-pulse2-plot", [pulse2plot], optionspulse2);
			p2.setupGrid();
			p2.draw();
			var p3 = $.plot("#my-pulse3-plot", [pulse3plot], optionspulse3);
			p3.setupGrid();
			p3.draw();
		};

		window.onload = function() {
			sock = new WebSocket("ws://localhost:5555/data");

			sock.onmessage = function(event) {
				var data = JSON.parse(event.data);
				console.log("data: "+JSON.stringify(data));
				if (data.freq != 0) {	
					freqplot.data.push([data.time, data.freq])
				}
				for (var i = 0; i < 999; i += 1) {
					pulse0plot.data.push([data.pulse0[i][0], data.pulse0[i][1]]);
					pulse1plot.data.push([data.pulse1[i][0], data.pulse1[i][1]]);
					pulse2plot.data.push([data.pulse2[i][0], data.pulse2[i][1]]);
					pulse3plot.data.push([data.pulse3[i][0], data.pulse3[i][1]]);
				}
				update();
				pulse0plot.data = []
				pulse1plot.data = []
				pulse2plot.data = []
				pulse3plot.data = []
			};
		};

		</script>

		<style>
		.my-plot-style {
			width: 400px;
			height: 200px;
			font-size: 14px;
			line-height: 1.2em;
		}
		</style>
	</head>

	<body>
		<div id="header">
			<h2>Test bench monitoring</h2>
		</div>
		<div id="my-freq-plot" class="my-plot-style"></div>
		<table>
			<tr>
				<td>
				<div id="my-pulse0-plot" class="my-plot-style"></div>
				</td>
				<td>
				<div id="my-pulse1-plot" class="my-plot-style"></div>
				</td>
			</tr>
			<tr>
				<td>
				<div id="my-pulse2-plot" class="my-plot-style"></div>
				</td>
				<td>
				<div id="my-pulse3-plot" class="my-plot-style"></div>
				</td>
			</tr>
		</table>
	</body>
</html>
`
