package main

import (
	. "github.com/strickyak/gosub/parser"

	"log"
	"os"
)

func main() {
	p := NewParser(os.Stdin, "stdin")
	for p.Kind != L_EOF {
		t := p.ParseAssignment()
		log.Printf(":::: %s", t)
		p.Next()
	}
	log.Printf("DONE")
}
