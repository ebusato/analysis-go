package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

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
	filew, err := os.Open(*fileName)
	if err != nil {
		log.Fatalf("could not create data file: %v\n", err)
	}
	defer filew.Close()
	r, err := rw.NewReader(bufio.NewReader(filew))
	if err != nil {
		log.Fatalf("could not open stream: %v\n", err)
	}

	switch *con {
	case "tcp":
		r.FrameT = rw.UDPorTCP16bits
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
		nFrames := uint(0)
		for {
			fmt.Printf("frame %v\n", nFrames)
			frame, err := r.Frame()
			if err != nil {
				if err != io.EOF {
					log.Fatalf("error loading frame: %v\n", err)
				}
				break
			}
			//frame.Print("medium")
			err = w.Frame(frame)
			if err != nil {
				log.Fatalf("error writing frame: %v\n", err)
			}
			nFrames++
		}
	case "udp":
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

		var word uint16
		nFrames := 0
		for {
			fmt.Printf("frame %v\n", nFrames)
			var frameBuffer []byte //:= make([]byte, 8230)
			for i := 0; i < 4115; i++ {
				r.ReadU16(&word)
				var tempBuf [2]byte
				binary.BigEndian.PutUint16(tempBuf[:], word)
				frameBuffer = append(frameBuffer, tempBuf[:]...)
			}
			// 			fmt.Printf("frameBuffer =")
			// 			for j := 0; j < len(frameBuffer)/2; j += 1 {
			// 				fmt.Printf("  %v: %x%x\n", j, frameBuffer[2*j], frameBuffer[2*j+1])
			// 			}
			conn.WriteToUDP(frameBuffer, addrClient)
			nFrames++
			time.Sleep(100000 * time.Microsecond)
		}
	}
}
