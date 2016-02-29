package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"os"
	"strings"
	"sync"

	"gitlab.in2p3.fr/avirm/analysis-go/dpga/rw"
)

func main() {
	log.SetFlags(log.Llongfile | log.LstdFlags)

	var (
		noEvents    = flag.Uint("n", 100000, "Number of events")
		outfileName = flag.String("o", "out.bin", "Name of the output file")
		ip          = flag.String("ip", "192.168.100.11", "IP address")
		port        = flag.String("p", "1024", "Port number")
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
	const N = 3
	var wg sync.WaitGroup
	wg.Add(N)

	go control(&wg)
	go stream(r, w, noEvents, &wg)
	go command(&wg)
	wg.Wait()
}

var (
	endStream  chan bool
	endCommand chan bool
	stopAll    chan bool
)

func control(wg *sync.WaitGroup) {
	defer wg.Done()
	streamIsEnded := false
	commandIsEnded := false
	for {
		select {
		case <-endStream:
			streamIsEnded = true
		case <-endCommand:
			commandIsEnded = true
		default:
			// do nothing
		}
		if streamIsEnded || commandIsEnded {
			fmt.Printf("endStream = %v or endCommand = %v\n", streamIsEnded, commandIsEnded)
			stopAll <- true
		}

		// 		fmt.Println("here")
		// 		if <-endStream || <-endCommand {
		// 			stopAll <- true
		// 		}
	}
}

func stream(r *rw.Reader, w *rw.Writer, noEvents *uint, wg *sync.WaitGroup) {
	defer wg.Done()
	nFrames := uint(0)
	for {
		switch {
		case nFrames/120 >= *noEvents:
			fmt.Println("specified number of events reached -> stopping acquisition")
			endStream <- true
			return
		default:
			iEvent := nFrames / 120
			if math.Mod(float64(nFrames)/120., 1) == 0 {
				fmt.Printf("event %v\n", iEvent)
			}
			//start := time.Now()
			frame, err := r.Frame()
			//duration := time.Since(start)
			//time.Sleep(1 * time.Millisecond)
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
			select {
			case <-stopAll:
				fmt.Printf("got stopAll %v %v\n", iEvent, *noEvents)
				*noEvents = iEvent + 1
			default:
				// do nothing
			}
			nFrames++
		}
	}
}

func command(wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		in := bufio.NewReader(os.Stdin) //Scanner
		word, _ := in.ReadString('\n')
		word = strings.Replace(word, "\n", "", -1)
		switch word {
		default:
			fmt.Println("-> Command not known, what do you mean ?", word)
		case "stop":
			fmt.Println("-> Stopping run")
			endCommand <- true
			return
		}
	}
}

/*
func stream(r *rw.Reader, w *rw.Writer, noEvents *uint, stopRun chan bool, wg *sync.WaitGroup) {
	defer wg.Done()
	nFrames := uint(0)
	for nFrames/120 < *noEvents {
		iEvent := float64(nFrames) / 120.
		if math.Mod(iEvent, 100) == 0 {
			fmt.Printf("event %v\n", iEvent)
		}
		//start := time.Now()
		frame, err := r.Frame()
		//duration := time.Since(start)
		//time.Sleep(1 * time.Millisecond)
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
		select {
		case <-stopRun:
			*noEvents = nFrames/120 + 1
		default:
			// do nothing
		}
		nFrames++
	}
}

func command(stopRun chan bool, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		in := bufio.NewReader(os.Stdin) //Scanner
		word, _ := in.ReadString('\n')
		word = strings.Replace(word, "\n", "", -1)
		switch word {
		default:
			fmt.Println("-> Command not known, what do you mean ?", word)
		case "stop":
			fmt.Println("-> Stopping run")
			stopRun <- true
			return
		}
	}
}
*/
