package parser

import (
	"bytes"
	"testing"
	//"fmt"
	"regexp"
)

var crunchLines = regexp.MustCompile(`(?m)^[#].*`)
var commentLines = regexp.MustCompile(`(?m)[/][/].*`)
var white = regexp.MustCompile(`[ \t\n\r]*`)

func Simplify(s string) string {
	s = crunchLines.ReplaceAllString(s, "")
	s = commentLines.ReplaceAllString(s, "")
	s = white.ReplaceAllString(s, "")
	return s
}

func SimplyEqual(t *testing.T, a, b string) {
	t.Logf("SimplyEqual: a = [[[[[%s]]]]]", a)
	t.Logf("SimplyEqual: b = [[[[[%s]]]]]", b)
	sa := Simplify(a)
	sb := Simplify(b)
	t.Logf("SimplyEqual: sa = %q", sa)
	t.Logf("SimplyEqual: sb = %q", sb)
	if sa != sb {
		t.Errorf("Simply not equal: %q vs %q", sa, sb)
	}
}

func Test1(t *testing.T) {
	prog := `var foo int`
	r := bytes.NewBufferString(prog)
	w := bytes.NewBufferString("")
	CompileToC(r, "Test1", w, &Options{
		LibDir:      "/none/",
		SkipBuiltin: true,
	})
	want := `P_int main__foo; void INIT() {}`
	SimplyEqual(t, w.String(), want)
}

func Test2(t *testing.T) {
	prog := `type Apple struct { Worm int }`
	r := bytes.NewBufferString(prog)
	w := bytes.NewBufferString("")
	CompileToC(r, "Test1", w, &Options{
		LibDir:      "/none/",
		SkipBuiltin: true,
	})
	want := `typedef Struct main__Apple; void INIT(){}`
	SimplyEqual(t, w.String(), want)
}

func Test3(t *testing.T) {
	prog := `type Frobber interface { Frob()string }`
	r := bytes.NewBufferString(prog)
	w := bytes.NewBufferString("")
	CompileToC(r, "Test1", w, &Options{
		LibDir:      "/none/",
		SkipBuiltin: true,
	})
	want := `typedef Interface main__Frobber; voidINIT(){}`
	SimplyEqual(t, w.String(), want)
}
