package parser

import (
	"bufio"
	"fmt"
	"io"
	"log"
)

const (
	L_EOF    = 0
	L_EOL    = 1
	L_Int    = 2
	L_String = 3
	L_Char   = 4
	L_Ident  = 5
	L_Punc   = 6
)

type Lex struct {
	Line     int
	Col      int
	Filename string

	Pending  byte // holds UnReadChar
	PrevLine int
	PrevCol  int

	Kind int
	Num  int
	Word string
	R    *bufio.Reader
}

func NewLex(r io.Reader, filename string) *Lex {
	z := &Lex{
		Line:     1,
		Col:      0,
		Filename: filename,
		R:        bufio.NewReader(r),
	}
	z.Next()
	return z
}

func (o *Lex) UnReadChar(ch byte) {
	o.Pending = ch
	o.Line, o.Col = o.PrevLine, o.PrevCol
}

func (o *Lex) ReadChar() byte {
	if o.Pending > 0 {
		z := o.Pending
		o.Pending = 0
		// log.Printf("=>> %q", string(z))
		return z
	}
	ch, err := o.R.ReadByte()
	// log.Printf("== ReadByte %d %v", ch, err)
	if err == io.EOF {
		// log.Printf("==> 0 (EOF)")
		return 0
	}
	if err != nil {
		panic(err)
	}
	o.PrevLine, o.PrevCol = o.Line, o.Col
	if ch == '\n' {
		o.Line++
		o.Col = 1
	} else {
		o.Col++
	}
	// log.Printf("==> %q", string(ch))
	return ch
}

func (o *Lex) Next() {
	o._Next_()
	log.Printf("--------- (%d)  %d  %q", o.Kind, o.Num, o.Word)
}
func (o *Lex) _Next_() {
	o.Num, o.Word = 0, ""
	c := o.ReadChar()
	for 0 < c && c <= 32 {
		if c == '\n' {
			o.Kind, o.Word = L_EOL, ";;"
			return
		}
		c = o.ReadChar()
	}
	if c == 0 {
		o.Kind, o.Word = L_EOF, "<EOF>"
		return
	}
	if c == '/' {
		c2 := o.ReadChar()
		if c2 == '/' {
			for {
				z := o.ReadChar()
				if z < 32 {
					o.Kind, o.Word = L_EOL, "//EOL//"
					return
				}
			}
		} else {
			o.UnReadChar(c2)
		}
	}
	/*
		neg := false
		if c == '-' {
			prevC := c
			c = o.ReadChar()
			if '0' <= c && c <= '9' {
				neg = true
			} else {
				o.UnReadChar(c)
				c = prevC
			}
		}
	*/
	if '0' <= c && c <= '9' {
		x := int(c - '0')
		c = o.ReadChar()
		for '0' <= c && c <= '9' {
			if x == 0 {
				panic("no octal")
			}
			x = 10*x + int(c-'0')
			c = o.ReadChar()
		}
		o.UnReadChar(c)
		/*
			if neg {
				o.Kind, o.Num = L_Int, -x
			} else {
		*/
		o.Kind, o.Num, o.Word = L_Int, x, fmt.Sprintf("%d", x)
		/*
			}
		*/
		return
	}
	if 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z' || c == '_' {
		var s []byte
		s = append(s, c)
		c = o.ReadChar()
		for '0' <= c && c <= '9' || 'A' <= c && c <= 'Z' || 'a' <= c && c <= 'z' || c == '_' {
			s = append(s, c)
			c = o.ReadChar()
		}
		o.UnReadChar(c)
		o.Kind, o.Word = L_Ident, string(s)
		return
	}
	if c == '"' {
		var s []byte
		c = o.ReadChar()
		for c != '"' {
			if c == '\\' {
				c = o.ReadChar()
				if c == 'n' {
					c = '\n'
				} else if '0' <= c && c <= '7' {
					c2 := o.ReadChar()
					c3 := o.ReadChar()
					if !('0' <= c2 && c2 <= '7') || !('0' <= c3 && c3 <= '7') {
						panic("bad octal in str")
					}
					c = byte(64*int(c-'0') + 8*int(c2-'0') + int(c3-'0'))
				}
			}
			s = append(s, c)
			c = o.ReadChar()
		}
		o.Kind, o.Word = L_String, string(s)
		return
	}
	if c == '\'' {
		var s []byte
		c = o.ReadChar()
		for c != '\'' {
			if c == '\\' {
				c = o.ReadChar()
				if c == 'n' {
					c = '\n'
				} else if '0' <= c && c <= '7' {
					c2 := o.ReadChar()
					c3 := o.ReadChar()
					if !('0' <= c2 && c2 <= '7') || !('0' <= c3 && c3 <= '7') {
						panic("bad octal in char literal")
					}
					c = byte(64*int(c-'0') + 8*int(c2-'0') + int(c3-'0'))
				}
			}
			s = append(s, c)
			c = o.ReadChar()
		}
		if len(s) != 1 {
			log.Panicf("bad char literal: %q", s)
		}
		o.Kind, o.Word = L_Char, string(s)
		return
	}

	d := o.ReadChar()
	for _, digraph := range []string{
		"..", "++", "--", ":=", "<=", "<<", ">=", ">>", "==", "!=", "+=", "-=", "*="} {
		if c == digraph[0] && d == digraph[1] {
			o.Kind, o.Word = L_Punc, digraph
			return
		}
	}
	o.UnReadChar(d)

	o.Kind, o.Word = L_Punc, string(c)
	return
}
