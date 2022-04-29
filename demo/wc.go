//go:build foo

package main

import "io"
import "os"
import "log"

func main() {
	wasWhite = true
	onebyte := make([]byte, 1)
	for {
		n, err := os.Stdin.Read(onebyte)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatalf("cannot read stdin: %v", err)
		}
		if n != 1 {
			log.Fatalf("expected to read 1 byte, got %d", n)
		}
		Count(onebyte[0])
	}
	println(Lines, Words, Bytes)
	log.Printf("%d %d %d", Lines, Words, Bytes)
}

var Bytes int
var Words int
var Lines int
var wasWhite bool

func Count(ch byte) {
	Bytes++
	switch ch {
	case '\n', '\r':
		Lines++
		wasWhite = true
	case ' ', '\t':
		wasWhite = true
	default:
		if wasWhite {
			Words++
		}
		wasWhite = false
	}
}
