package main

import (
	. "github.com/strickyak/gosub/parser"

	"flag"
	"log"
	"os"
)

var flagX = flag.Bool("x", false, "parse expression")

func main() {
	flag.Parse()

	p := NewParser(os.Stdin, "stdin")
	for p.Kind != L_EOF {
		if *flagX {
			t := p.ParseAssignment()
			log.Printf(":::: %s", t)
			p.Next()
		} else {
			p.ParseTop()
			log.Printf(":::: %#v", p)
			break
		}
	}
	log.Printf("DONE")
}
