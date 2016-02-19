package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"

	"gitlab.in2p3.fr/avirm/analysis-go/dpga/rw"
)

func main() {
	log.SetFlags(log.Llongfile | log.LstdFlags)

	var (
		noEvents    = flag.Uint("n", 1000, "Number of events")
		outfileName = flag.String("o", "out.bin", "Name of the output file")
		port        = flag.String("p", "1024", "Port number")
	)

	flag.Parse()

	// Reader
	laddr, err := net.ResolveTCPAddr("tcp", "192.168.100.11:"+*port)
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
	const N = 2
	var wg sync.WaitGroup
	wg.Add(N)
	go stream(r, w, noEvents, &wg)
	go command(&wg)
	wg.Wait()
}

func stream(r *rw.Reader, w *rw.Writer, noEvents *uint, wg *sync.WaitGroup) {
	defer wg.Done()
	nFrames := uint(0)
	for nFrames/120 < *noEvents {
		//start := time.Now()
		frame, err := r.Frame()
		//duration := time.Since(start)
// 		time.Sleep(1 * time.Millisecond)
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
		nFrames++
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
			fmt.Println("do nothing !", word)
		case "q":
			fmt.Println("quit !")
			return
		}
	}
}

/*
func command(e interface{}) {
	for {
		switch e := e.(type) {
		case key.Event:
			switch e.Code {
			case key.CodeEscape, key.CodeQ:
				if e.Direction == key.DirPress {
					return
				}
			}
		}
	}
}
*/

/*
// OLD STUFF, TO BE REMOVED
	word := make([]byte, 4)
	n, err := tcp.Read(word)
	wordu32 := binary.BigEndian.Uint32(word)
	fmt.Printf("Word from server: %v %x\n", n, wordu32)
	if err != nil {
		if err != io.EOF {
			log.Fatalf("error reading word")
		}
	}
*/
