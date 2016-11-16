// Questions:
//   - is it better to have bufio or not ?

package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net"
	"strconv"

	"gitlab.in2p3.fr/avirm/analysis-go/dpgatca/rw"
)

var (
	ip        = flag.String("ip", "192.168.100.11", "IP address")
	port      = flag.String("p", "1024", "Port number")
	frameFreq = flag.Uint("ff", 1000, "Frame printing frequency")
)

func UDPConn(p *string) *net.UDPConn {
	fmt.Println("addr", *ip+":"+*p)
	// 	conn, err := net.Dial("tcp", *ip+":"+*p)

	RemoteAddr, err := net.ResolveUDPAddr("udp", *ip+":"+*p)
	fmt.Println("client RemoteAddr =", RemoteAddr.IP, RemoteAddr.Port, RemoteAddr.Zone)
	conn, err := net.DialUDP("udp", nil, RemoteAddr)
	if err != nil {
		return nil
	}
	return conn
}

type Reader struct {
	conn *net.UDPConn
}

func NewReader(conn *net.UDPConn) *Reader {
	return &Reader{conn: conn}
}

func (r *Reader) Read(p []byte) (n int, err error) {
	n, _, err = r.conn.ReadFromUDP(p)
	return
}

func main() {
	log.SetFlags(log.Llongfile | log.LstdFlags)
	flag.Parse()
	var r *rw.Reader
	conn := UDPConn(port)
	for i := 0; conn == nil; i++ {
		newportu, err := strconv.ParseUint(*port, 10, 64)
		if err != nil {
			panic(err)
		}
		newportu += 1
		newport := strconv.FormatUint(newportu, 10)
		fmt.Printf("Port %v not responding, trying %v\n", *port, newport)
		*port = newport
		conn = UDPConn(port)
		if i >= 5 {
			log.Fatalf("Cannot find port to connect to server")
		}
	}
	conn.SetReadBuffer(216320) // not sure what is the unit of the argument
	//conn.SetReadBuffer(216320)
	conn.Write([]byte("Hello from client"))
	r, _ = rw.NewReader(bufio.NewReader(NewReader(conn)))
	r.ReadMode = rw.UDPHalfDRS
	nframes := uint(0)
	AMCFrameCounterPrev := uint32(0)
	//ASMFrameCounterPrev := uint64(0)
	data := make([]byte, 8238)
	//amplitudes := make([]uint16, 1022)
	for {
		if nframes%*frameFreq == 0 {
			fmt.Printf("reading frame %v\n", nframes)
		}
		/////////////////////////////////////
		// Option 1
		/*
			frame, _ := r.Frame()
			if nframes > 0 {
				if frame.Block.AMCFrameCounter != AMCFrameCounterPrev+1 {
					fmt.Printf("frame.Block.AMCFrameCounter != AMCFrameCounterPrev + 1\n")
				}
				if frame.Block.ASMFrameCounter != ASMFrameCounterPrev+1 {
					fmt.Printf("frame.Block.ASMFrameCounter != ASMFrameCounterPrev + 1\n")
				}
			}
			AMCFrameCounterPrev = frame.Block.AMCFrameCounter
			ASMFrameCounterPrev = frame.Block.ASMFrameCounter
			if r.ReadMode == rw.UDPHalfDRS && frame.Block.UDPPayloadSize < 8230 {
				log.Printf("frame.Block.UDPPayloadSize = %v\n", frame.Block.UDPPayloadSize)
			}
		*/
		/////////////////////////////////////

		/////////////////////////////////////
		// Option 2
		conn.ReadFromUDP(data)
		/*
			for i := 0; i < 4; i++ {
				for j := range amplitudes {
					amplitudes[j] = binary.BigEndian.Uint16(data[44+2*j+i*2*1023 : 46+2*j+i*2*1023])
				}
			}*/

		//fmt.Println(data)
		AMCFrameCounter0 := binary.BigEndian.Uint16(data[2:4])
		AMCFrameCounter1 := binary.BigEndian.Uint16(data[4:6])
		AMCFrameCounter := (uint32(AMCFrameCounter0) << 16) + uint32(AMCFrameCounter1)
		if nframes > 0 {
			if AMCFrameCounter != AMCFrameCounterPrev+1 {
				fmt.Printf("AMCFrameCounter != AMCFrameCounterPrev+1\n")
			}
		}
		AMCFrameCounterPrev = AMCFrameCounter

		/////////////////////////////////////

		nframes++
	}
}
