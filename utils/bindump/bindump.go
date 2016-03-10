package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

func main() {
	log.SetFlags(log.Llongfile | log.LstdFlags)

	var (
		infileName = flag.String("i", "", "Name of the input file")
	)

	flag.Parse()

	f, err := os.Open(*infileName)
	if err != nil {
		log.Fatalf("could not open data file: %v\n", err)
	}
	defer f.Close()

	r := bufio.NewReader(f)

	for i := 0; ; i++ {
		var word uint32
		err := binary.Read(r, binary.BigEndian, &word)
		if err != nil {
			if err == io.EOF {
				fmt.Printf("reached end of file, stopping.\n")
				return
			}
			log.Fatalf("error while reading word %v\n", word)
		}
		if word == 0xCAFEDECA || word>>4 == 0xBADCAFE {
			fmt.Printf("\n")
		}
		fmt.Printf("%x ", word)
		if (i != 0 && i%15 == 0) || word == 0xCAFEDECA || word>>4 == 0xBADCAFE {
			fmt.Printf("\n")
		}
	}
}
