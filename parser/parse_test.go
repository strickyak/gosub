package parser

import (
	"bytes"
	"regexp"
	"testing"
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

func TestVarInt(t *testing.T) {
	prog := `var foo int`
	r := bytes.NewBufferString(prog)
	w := bytes.NewBufferString("")
	CompileToC(r, "TEST", w, &Options{
		LibDir:      "/none/",
		SkipBuiltin: true,
	})
	want := `P_int main__foo;`
	SimplyEqual(t, w.String(), want)
}

func TestAppleStruct(t *testing.T) {
	prog := `type Apple struct { Worm int }`
	r := bytes.NewBufferString(prog)
	w := bytes.NewBufferString("")
	CompileToC(r, "TEST", w, &Options{
		LibDir:      "/none/",
		SkipBuiltin: true,
	})
	want := `typedef Struct main__Apple;`
	SimplyEqual(t, w.String(), want)
}

func TestFrobberInterface(t *testing.T) {
	prog := `type Frobber interface { Frob()string }`
	r := bytes.NewBufferString(prog)
	w := bytes.NewBufferString("")
	CompileToC(r, "TEST", w, &Options{
		LibDir:      "/none/",
		SkipBuiltin: true,
	})
	want := `typedef Interface main__Frobber;`
	SimplyEqual(t, w.String(), want)
}

func TestConstInt(t *testing.T) {
	prog := `const foo = 123`
	r := bytes.NewBufferString(prog)
	w := bytes.NewBufferString("")
	CompileToC(r, "TEST", w, &Options{
		LibDir:      "/none/",
		SkipBuiltin: true,
	})
	want := ``
	SimplyEqual(t, w.String(), want)
}

func TestFunc0(t *testing.T) {
	prog := `func zero(){}`
	r := bytes.NewBufferString(prog)
	w := bytes.NewBufferString("")
	CompileToC(r, "TEST", w, &Options{
		LibDir:      "/none/",
		SkipBuiltin: true,
	})
	want := `nando`
	SimplyEqual(t, w.String(), want)
}
