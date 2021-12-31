package parser

import (
	//"bufio"
	//"bytes"
	"fmt"
	//"io"
	"log"
	//"os"
	"path/filepath"
	//"reflect"
	"runtime/debug"
	"strings"
)

var Format = fmt.Sprintf
var F = fmt.Sprintf
var P = fmt.Fprintf
var L = log.Printf

// #################################################

func assert(b bool) {
	if !b {
		panic("assert fails")
	}
}

const CantString = "--?--"

func TryString(a fmt.Stringer) (z string) {
	defer func() {
		r := recover()
		if r != nil {
			log.Printf("CantString: %#v :: %v", a, r)
			z = CantString
		}
	}()
	return a.String()
}

func Say(args ...interface{}) {
	// Log a line with the code location of the caller.
	bb := debug.Stack()
	ww := strings.Split(string(bb), "\n")
	log.Printf("%s", filepath.Base(ww[6]))

	for i, a := range args {
		if str, ok := a.(fmt.Stringer); ok {
			z := TryString(str)
			if z == CantString {
				log.Printf("-=Say[%d]: *PANIC*IN*STRING* %#v", i, a)
			} else {
				log.Printf("- Say[%d]: %s", i, str.String())
			}
		} else {
			log.Printf("--Say[%d]: %#v", i, a)
		}
	}
}

func Panicf(format string, args ...interface{}) string {
	s := Format(format, args...)
	panic(s)
}

func FullName(a string, b string) string {
	return a + "__" + b
}

var _serial_prev uint = 100

func Serial(prefix string) string {
	_serial_prev++
	return Format("%s_%d", prefix, _serial_prev)
}
func SerialIfEmpty(s string) string {
	if len(s) > 0 {
		return s
	}
	return Serial("empty")
}
