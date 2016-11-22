// Questions:
//   - is it better to have bufio or not ?

package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"runtime/pprof"
	"strconv"
)

var (
	ip         = flag.String("ip", "localhost", "IP address")
	port       = flag.String("p", "6000", "Port number")
	frameFreq  = flag.Uint("ff", 1000, "Frame printing frequency")
	nFramesTot = flag.Uint("n", 300000, "Number of frames to process")
)

func UDPConn(p *string) *net.UDPConn {
	fmt.Println("addr", *ip+":"+*p)
	locAddr, err := net.ResolveUDPAddr("udp", *ip+":"+*p)
	conn, err := net.ListenUDP("udp", locAddr)

	// 	conn, err := net.Dial("tcp", *ip+":"+*p)

	// 	RemoteAddr, err := net.ResolveUDPAddr("udp", *ip+":"+*p)
	// 	fmt.Println("RemoteAddr =", RemoteAddr.IP, RemoteAddr.Port, RemoteAddr.Zone)
	//conn, err := net.DialUDP("udp", nil, RemoteAddr)
	//LocAddr, err := net.ResolveUDPAddr("udp", *ip+":5557")
	//conn, err := net.DialUDP("udp", LocAddr, RemoteAddr)
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

	f, err := os.Create("perf.prof")
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

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
	// 	conn.SetReadBuffer(200)
	conn.SetReadBuffer(216320) // not sure what is the unit of the argument
	//conn.Write([]byte("Hello from client"))

	nframes := uint(0)
	AMCFrameCounterPrev := uint32(0)
	buf := make([]byte, 8238)
	for nframes < *nFramesTot {
		if nframes%*frameFreq == 0 {
			fmt.Printf("reading frame %v\n", nframes)
		}
		///////////////////////////////////////////////////////
		// Option 1 : fastest
		conn.ReadFromUDP(buf)
		AMCFrameCounters0 := binary.BigEndian.Uint16(buf[2:4])
		AMCFrameCounters1 := binary.BigEndian.Uint16(buf[4:6])
		AMCFrameCounter := (uint32(AMCFrameCounters0) << 16) + uint32(AMCFrameCounters1)
		if nframes > 0 {
			if AMCFrameCounter != AMCFrameCounterPrev+1 {
				fmt.Printf("AMCFrameCounter != AMCFrameCounterPrev+1\n")
			}
		}
		AMCFrameCounterPrev = AMCFrameCounter
		///////////////////////////////////////////////////////

		/*
			//////////////////////////////////////////////////////
			// Option 2 : a bit more refined
			  n, _, err := conn.ReadFromUDP(buf)
				frame := rw.NewFrame(n) // <- here
				//fmt.Println("payload =", n)
				frame.FillHeader(buf) // <- here
				err = frame.IntegrityHeader()
				if err != nil {
					panic(err)
				}
				frame.FillData(buf)
				err = frame.IntegrityData()
				if err != nil {
					panic(err)
				}
				if nframes > 0 {
					if frame.AMCFrameCounter != AMCFrameCounterPrev+1 {
						fmt.Printf("frame.AMCFrameCounter != AMCFrameCounterPrev+1\n")
					}
				}
				AMCFrameCounterPrev = frame.AMCFrameCounter
			//////////////////////////////////////////////////////
		*/
		nframes++
	}
}
