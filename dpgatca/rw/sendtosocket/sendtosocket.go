package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"net"
	"os"

	"gitlab.in2p3.fr/avirm/analysis-go/dpgatca/rw"
)

func main() {
	log.SetFlags(log.Llongfile | log.LstdFlags)

	var (
		fileName = flag.String("i", "", "Input file name")
		ip       = flag.String("ip", "localhost", "IP address")
		port     = flag.String("p", "5556", "Port number")
		//freq     = flag.Uint("freq", 100, "Event number printing frequency")
	)
	flag.Parse()

	// Reader
	filew, err := os.Open(*fileName)
	if err != nil {
		log.Fatalf("could not create data file: %v\n", err)
	}
	defer filew.Close()
	r, err := rw.NewReader(bufio.NewReader(filew))
	if err != nil {
		log.Fatalf("could not open stream: %v\n", err)
	}

	// Writer
	ln, err := net.Listen("tcp", *ip+":"+*port)
	if err != nil {
		log.Fatal(err)
	}
	conn, err := ln.Accept()

	w := rw.NewWriter(bufio.NewWriter(conn))
	if err != nil {
		log.Fatalf("could not open file: %v\n", err)
	}
	defer w.Close()

	// Start writing stream to TCP
	// 	hdr := r.Header()
	// 	hdr.Print()
	//
	// 	err = w.Header(hdr, false)
	// 	if err != nil {
	// 		log.Fatalf("error writing header: %v\n", err)
	// 	}

	nFrames := uint(0)
	for {
		frame, err := r.Frame()

		if err != nil {
			if err != io.EOF {
				log.Fatalf("error loading frame: %v\n", err)
			}
			break
		}
		err = w.Frame(frame)
		if err != nil {
			log.Fatalf("error writing frame: %v\n", err)
		}
		nFrames++
	}

}
