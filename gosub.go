package main

import (
	. "github.com/strickyak/gosub/parser"

	"flag"
	"log"
	"os"
)

func main() {
	flag.Parse()
	CompileToC(os.Stdin, "stdin", os.Stdout)
	log.Printf("DONE")
}
