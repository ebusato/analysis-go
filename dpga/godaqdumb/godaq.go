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

	// 	r.Debug = true

	// Start reading TCP stream
	hdr := r.Header()
	//hdr.Print()

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

	go control(terminateStream, commandIsEnded)
	go stream(terminateStream, r, w, noEvents, &wg)
	go command(commandIsEnded)
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

func stream(terminateStream chan bool, r *rw.Reader, w *rw.Writer, noEvents *uint, wg *sync.WaitGroup) {
	defer wg.Done()
	nFrames := uint(0)

	frame := &rw.Frame{
		Block: rw.Block{
			Data: make([]uint32, rw.NumSamples()),
		},
	}

	for {
		iEvent := nFrames / 120
		select {
		case <-terminateStream:
			*noEvents = iEvent + 1
			fmt.Printf("terminating stream for total number of events = %v.\n", *noEvents)
		default:
			switch iEvent < *noEvents {
			case true:
				if math.Mod(float64(nFrames)/120., 100) == 0 {
					fmt.Printf("event %v\n", iEvent)
				}

				r.ReadFrame(frame)
				err := r.Err()
				//start := time.Now()
				//frame, err := r.Frame()
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
				err = w.Frame(frame)
				if err != nil {
					log.Fatalf("error writing frame: %v\n", err)
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
