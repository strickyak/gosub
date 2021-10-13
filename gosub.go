package main

import (
	. "github.com/strickyak/gosub/parser"

	"flag"
	"log"
	"os"
)

var LibDir = flag.String("libdir", "lib", "where to import libs from")

func main() {
	flag.Parse()
	CompileToC(&Options{
		LibDir: *LibDir,
	}, os.Stdin, "stdin", os.Stdout)
	log.Printf("DONE")
}
