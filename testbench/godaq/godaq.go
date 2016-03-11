package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"

	"golang.org/x/net/websocket"

	"gitlab.in2p3.fr/avirm/analysis-go/pulse"
	"gitlab.in2p3.fr/avirm/analysis-go/testbench/rw"
)

var (
	datac = make(chan Data)
)

// type Data struct {
// 	X   float64 `json:"x"`
// 	Sin float64 `json:"sin"`
// 	Cos float64 `json:"cos"`
// }

type Data struct {
	N     int       `json:"n"`
	Ampl1 []float64 `json:"ampl1"`
	Ampl2 []float64 `json:"ampl2"`
}

func main() {
	log.SetFlags(log.Llongfile | log.LstdFlags)

	var (
		noEvents    = flag.Uint("n", 100000, "Number of events")
		outfileName = flag.String("o", "out.bin", "Name of the output file")
		ip          = flag.String("ip", "192.168.100.11", "IP address")
		port        = flag.String("p", "1024", "Port number")
		monFreq     = flag.Uint("mf", 1500, "Monitoring frequency")
		evtFreq     = flag.Uint("ef", 100, "Event printing frequency")
		debug       = flag.Bool("d", false, "If set, debugging informations are printed")
		//webad       = flag.String("webad", ":5555", "server address:port")
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
	cframe1 := make(chan rw.Frame)
	cframe2 := make(chan rw.Frame)

	if *debug {
		r.Debug = true
	}

	go control(terminateStream, commandIsEnded)
	go stream(terminateStream, cframe1, cframe2, r, w, noEvents, monFreq, evtFreq, &wg)
	go command(commandIsEnded)
	go monitoring(cframe1, cframe2)

	// web server
	// 	http.HandleFunc("/", plotHandle)
	// 	http.Handle("/data", websocket.Handler(dataHandler))
	// 	err = http.ListenAndServe(*webad, nil)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	wg.Wait()
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

func stream(terminateStream chan bool, cframe1 chan rw.Frame, cframe2 chan rw.Frame, r *rw.Reader, w *rw.Writer, noEvents *uint, monFreq *uint, evtFreq *uint, wg *sync.WaitGroup) {
	defer wg.Done()
	nFrames := uint(0)
	for {
		iEvent := nFrames / 2
		select {
		case <-terminateStream:
			*noEvents = iEvent + 1
			fmt.Printf("terminating stream for total number of events = %v.\n", *noEvents)
		default:
			switch iEvent < *noEvents {
			case true:
				if math.Mod(float64(nFrames)/2., float64(*evtFreq)) == 0 {
					fmt.Printf("event %v\n", iEvent)
				}
				//start := time.Now()
				frame, err := r.Frame()
				//frame.Print("short")
				//duration := time.Since(start)
				// 				//time.Sleep(1 * time.Millisecond)S
				if err != nil {
					if err != io.EOF {
						log.Fatalf("error loading frame: %v\n", err)
					}
					if frame.ID != rw.LastFrame() {
						log.Fatalf("invalid last frame id. got=%d. want=%d", frame.ID, rw.LastFrame())
					}
					break
				}
				err = w.Frame(*frame)
				if err != nil {
					log.Fatalf("error writing frame: %v\n", err)
				}
				// monitoring
				if iEvent%*monFreq == 0 {
					if nFrames%2 == 0 {
						//fmt.Printf("sending to cframe1; iEvent = %v, nFrames=%v\n", iEvent, nFrames)
						cframe1 <- *frame
						//fmt.Println("sent to cframe1", iEvent)
					} else {
						//fmt.Printf("sending to cframe2; iEvent = %v, nFrames=%v\n", iEvent, nFrames)
						cframe2 <- *frame
						//fmt.Println("sent to cframe2", iEvent)
					}

					// webserver data
					//datac <- Data{float64(frame.ID), math.Sin(float64(frame.ID)), math.Cos(float64(frame.ID))}
					// 					data := Data{N: len(frame.Block.Data)}
					// 					ampl1 := make([]float64, data.N)
					// 					ampl2 := make([]float64, data.N)
					// 					for i := range frame.Block.Data {
					// 						ampl1[i] = float64(frame.Block.Data[i] & 0xFFF)
					// 						ampl2[i] = float64(frame.Block.Data[i] >> 16)
					// 					}
					// 					data.Ampl1 = ampl1
					// 					data.Ampl2 = ampl2
					// 					datac <- data
				}
				nFrames++
			case false:
				fmt.Println("reaching specified number of events, stopping.")
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

func monitoring(cframe1 chan rw.Frame, cframe2 chan rw.Frame) {
	for {
		//fmt.Println("receiving from cframe1")
		frame1 := <-cframe1
		//fmt.Println("receiving from cframe2")
		frame2 := <-cframe2
		//fmt.Println("received everything from frames")

		event := rw.MakeEventFromFrames(&frame1, &frame2)
		event.PlotPulses(pulse.XaxisTime, false)
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
		<title>Plotting stuff with Flot</title>
		<script src="//cdnjs.cloudflare.com/ajax/libs/jquery/2.0.3/jquery.min.js"></script>
		<script src="//cdnjs.cloudflare.com/ajax/libs/flot/0.8.3/jquery.flot.min.js"></script>
		<script type="text/javascript">
		var sock = null;
		var sinplot = {
			label: "sin(x)",
			data: [],
		};
		var cosplot = {
			label: "cos(x)",
			data: [],
		};

		function update() {
			var p1 = $.plot("#my-sin-plot", [sinplot]);
			p1.setupGrid(); // needed as x-axis changes
			p1.draw();

			var cos = $.plot("#my-cos-plot", [cosplot]);
			cos.setupGrid();
			cos.draw();
		};

		window.onload = function() {
			sock = new WebSocket("ws://localhost:5555/data");

			sock.onmessage = function(event) {
				var data = JSON.parse(event.data);
				console.log("data: "+JSON.stringify(data));
				for (var i = 0; i < data.n; i += 1) {
					sinplot.data.push([i, data.ampl1(i)])
				}
				cosplot.data.push([data.x, data.cos]);
				update();
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
			<h2>My plot</h2>
		</div>

		<div id="content">
			<div id="my-sin-plot" class="my-plot-style"></div>
			<br>
			<div id="my-cos-plot" class="my-plot-style"></div>
		</div>
	</body>
</html>
`

//sinplot.data.push([data.x, data.sin]);

/*
// google chart
const page = `
<html>
	<head>
		<title>Plotting stuff with Google-Charts</title>
		<script type="text/javascript" src="https://www.gstatic.com/charts/loader.js"></script>
		<script type="text/javascript">
			google.charts.load('current', {packages: ['corechart', 'line']});
			google.charts.setOnLoadCallback(initDrawSin);
			google.charts.setOnLoadCallback(initDrawCos);

			var sindata = null;
			var sinplot = null;
			function initDrawSin() {
				sindata = new google.visualization.DataTable();
				sindata.addColumn("number", "time");
				sindata.addColumn("number", "sine");

				sinplot = new google.charts.Line(document.getElementById("my-sin-plot"));
				drawSin();
			};

			var cosdata = null;
			var cosplot = null;
			function initDrawCos() {
				cosdata = new google.visualization.DataTable();
				cosdata.addColumn("number", "time");
				cosdata.addColumn("number", "cosine");

				cosplot = new google.charts.Line(document.getElementById("my-cos-plot"));
				drawCos();
			};

			function drawSin() {
				sinplot.draw(sindata, {
					hAxis: {
						title: "Time",
					},
					vAxis: {
						title: "Sine",
					},
					legend: {
						position: "none",
					},
					chart: {
						title: "Sine",
					},
				});
			};

			function drawCos() {
				cosplot.draw(cosdata, {
					hAxis: {
						title: "Time",
					},
					vAxis: {
						title: "Cosine",
					},
					legend: {
						position: "none",
					},
					chart: {
						title: "Cosine",
					},
				});
			};


			var sock = null;

			function update() {
				drawSin();
				drawCos();
			};

			window.onload = function() {
				sock = new WebSocket("ws://localhost:5555/data");

				sock.onmessage = function(event) {
					var data = JSON.parse(event.data);
					console.log("data: "+JSON.stringify(data));
					sindata.addRows([[data.x, data.sin]]);
					cosdata.addRows([[data.x, data.cos]]);
					update();
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
			<h2>My plot</h2>
		</div>

		<div id="content">
			<div id="my-sin-plot" class="my-plot-style"></div>
			<br>
			<div id="my-cos-plot" class="my-plot-style"></div>
		</div>
	</body>
</html>
`
*/
