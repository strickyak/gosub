package main

import (
	. "github.com/strickyak/gosub/parser"

	"log"
	"os"
)

func main() {
	/*
		for lex := NewLex(os.Stdin, "stdin"); lex.Kind != L_EOF; lex.Next() {
			log.Printf("%d:%d: (%d) %d %q",
				lex.Line, lex.Col, lex.Kind, lex.Num, lex.Word)
		}
	*/
	p := NewParser(os.Stdin, "stdin")
	for p.Kind != L_EOF {
		if p.Kind != L_EOL {
			t := p.ParseAssignment()
			log.Printf(":::: %s", t)
		}
		p.Next()
	}
	log.Printf("DONE")
}
