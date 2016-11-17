// Questions:
//   - is it better to have bufio or not ?

package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"runtime/pprof"
	"strconv"

	"gitlab.in2p3.fr/avirm/analysis-go/dpgatca/rw"
)

var (
	ip         = flag.String("ip", "192.168.100.11", "IP address")
	port       = flag.String("p", "1024", "Port number")
	frameFreq  = flag.Uint("ff", 1000, "Frame printing frequency")
	nFramesTot = flag.Uint("n", 100000, "Number of frames to process")
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

	f, err := os.Create("perf.prof")
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

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
	conn.Write([]byte("Hello from client"))
	r, _ = rw.NewReader(bufio.NewReader(NewReader(conn)))
	r.ReadMode = rw.UDPHalfDRS
	nframes := uint(0)
	AMCFrameCounterPrev := uint32(0)
	//ASMFrameCounterPrev := uint64(0)
	buf := make([]byte, 8238)
	for nframes < *nFramesTot {
		if nframes%*frameFreq == 0 {
			fmt.Printf("reading frame %v\n", nframes)
		}
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
		/////////////////////////////////////
		nframes++
	}
}
