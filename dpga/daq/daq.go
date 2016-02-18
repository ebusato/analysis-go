package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"net"
	"os"

	"gitlab.in2p3.fr/avirm/analysis-go/dpga/rw"
)

func main() {
	log.SetFlags(log.Llongfile | log.LstdFlags)

	var (
		noEvents = flag.Uint("n", 1000, "Number of events")
		outfileName = flag.String("o", "out.bin", "Name of the output file")
	)

	flag.Parse()

	laddr, err := net.ResolveTCPAddr("tcp", "192.168.100.11:1024")
	if err != nil {
		log.Fatal(err)
	}
	tcp, err := net.DialTCP("tcp", nil, laddr)
	if err != nil {
		log.Fatal(err)
	}

	// Reader
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

	nFrames := uint(0)
	for nFrames / 120 < *noEvents {
		frame, err := r.Frame()
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
