package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	src := flag.String("in", "", "Name of the source to parse")
	flag.Parse()

	fh, err := os.Open(*src)
	if err != nil {
		log.Fatalf("cannot open %s: %s", *src, err)
	}
	p := Parser{br: Reader{R: fh}}
	lnum := 0
	for {
		l, err := p.Next()
		if err != nil {
			break
		}
		fmt.Println("line", lnum, ":", l)
		lnum++
	}
	fh.Close()
}
