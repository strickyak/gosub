package main

import (
	. "github.com/strickyak/gosub/parser"

	"flag"
	"log"
	"os"
)

var LibDir = flag.String("libdir", "lib", "where to import libs from")
var SkipBuiltin = flag.Bool("skip_builtin", false, "Don't automatically import `builtin` library")

func main() {
	log.SetFlags(0)
	log.SetPrefix("## ")
	flag.Parse()
	CompileToC(os.Stdin, "stdin", os.Stdout, &Options{
		LibDir:      *LibDir,
		SkipBuiltin: *SkipBuiltin,
	})
	log.Printf("DONE")
}
