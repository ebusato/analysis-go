package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
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
	/*
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
	*/
	addr, err := net.ResolveUDPAddr("udp4", *ip+":"+*port) // maybe change to udp4
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	var buf [2]byte
	fmt.Println("Server: before ReadFromUDP")
	_, addrClient, _ := conn.ReadFromUDP(buf[:])
	fmt.Println("UDP client address: ", addrClient)
	fmt.Printf("buf[:] from client= %v\n", buf[:])

	// Start writing stream to TCP
	// 	hdr := r.Header()
	// 	hdr.Print()
	//
	// 	err = w.Header(hdr, false)
	// 	if err != nil {
	// 		log.Fatalf("error writing header: %v\n", err)
	// 	}

	/*
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
	*/

	nWords := 0
	var word uint16
	/*
		for {
			r.ReadU16(&word)
			binary.BigEndian.PutUint16(buf[:], word)
			//fmt.Printf("buf[:] = %x %x\n", word, buf[:])
			conn.WriteToUDP(buf[:], addrClient)
			//conn.ReadFromUDP(buf[:])
			//fmt.Println("In loop ReadFromUDP =", buf)
			time.Sleep(1000 * time.Microsecond)
			nWords++
		}
	*/

	var frameBuffer []byte //:= make([]byte, 8230)
	for {
		for i := 0; i < 4115; i++ {
			r.ReadU16(&word)
			var tempBuf [2]byte
			binary.BigEndian.PutUint16(tempBuf[:], word)
			frameBuffer = append(frameBuffer, tempBuf[:]...)
		}
		fmt.Printf("frameBuffer = %x\n", frameBuffer)
		conn.WriteToUDP(frameBuffer, addrClient)
		nWords++
	}
}
