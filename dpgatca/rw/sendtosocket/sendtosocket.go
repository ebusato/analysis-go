package main

import (
	"bufio"
	"flag"
	"fmt"
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
		con      = flag.String("con", "udp", "Connection type (possible values: udp, tcp)")
		//freq     = flag.Uint("freq", 100, "Event number printing frequency")
	)
	flag.Parse()

	// Reader
	file, err := os.Open(*fileName)
	if err != nil {
		log.Fatalf("could not create data file: %v\n", err)
	}
	defer file.Close()
	r, err := rw.NewReader(file)
	if err != nil {
		log.Fatalf("could not open stream: %v\n", err)
	}
	r.ReadMode = rw.Default
	switch *con {
	case "tcp":
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
		nFrames := 0
		for {
			if nFrames%1 == 0 {
				fmt.Printf("frame %v\n", nFrames)
			}
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
	case "udp":
		addr, err := net.ResolveUDPAddr("udp", *ip+":"+*port) // maybe change to udp4
		conn, err := net.ListenUDP("udp", addr)
		if err != nil {
			log.Fatal(err)
		}
		defer conn.Close()

		var buf [2]byte
		fmt.Println("Server: before ReadFromUDP")
		// 		_, addrClient, _ := conn.ReadFromUDP(buf[:])
		// 		fmt.Println("UDP client address: ", addrClient)
		addrClient, err := net.ResolveUDPAddr("udp", *ip+":6000")
		fmt.Printf("buf[:] from client= %v\n", buf[:])

		AMCFrameCounterPrev := uint32(0)
		ASMFrameCounterPrev := uint64(0)

		nFrames := 0
		for {
			if nFrames%1000 == 0 {
				fmt.Printf("frame %v\n", nFrames)
			}
			frame, err := r.Frame()
			if nFrames > 0 {
				if frame.AMCFrameCounter != AMCFrameCounterPrev+1 {
					fmt.Printf("frame.Block.AMCFrameCounter != AMCFrameCounterPrev + 1\n")
				}
				if frame.ASMFrameCounter != ASMFrameCounterPrev+1 {
					fmt.Printf("frame.Block.ASMFrameCounter != ASMFrameCounterPrev + 1\n")
				}
			}
			AMCFrameCounterPrev = frame.AMCFrameCounter
			ASMFrameCounterPrev = frame.ASMFrameCounter
			nFrames++
			// 			fmt.Println("AMCFrameCounter =", frame.Block.AMCFrameCounter)
			// 			fmt.Println("ASMFrameCounter =", frame.Block.ASMFrameCounter)

			if err != nil {
				if err != io.EOF {
					log.Fatalf("error loading frame: %v\n", err)
				}
				break
			}

			// These two lines are slower than the TCP equivalent
			//    -> try to understand why
			udpBuf := frame.Buffer()
			//conn.WriteToUDP(udpBuf, addrClient)
			conn.WriteToUDP(udpBuf, addrClient)

			//time.Sleep(100000 * time.Microsecond)
		}
	}
}
