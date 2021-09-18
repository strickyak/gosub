package parser

import (
	"bufio"
	"fmt"
	"io"
	"log"
)

type Parser struct {
	*Lex
	Package *DefPackage
	Imports map[string]*DefImport
	Consts  map[string]*DefConst
	Vars    map[string]*DefVar
	Types   map[string]*DefType
	Funcs   map[string]*DefFunc
}

func NewParser(r io.Reader, filename string) *Parser {
	return &Parser{
		Lex:     NewLex(r, filename),
		Imports: make(map[string]*DefImport),
		Consts:  make(map[string]*DefConst),
		Vars:    make(map[string]*DefVar),
		Types:   make(map[string]*DefType),
		Funcs:   make(map[string]*DefFunc),
	}
}

func (o *Parser) ParsePrim() Expr {
	if o.Kind == L_Int {
		z := &LitIntX{o.Num}
		o.Next()
		return z
	}
	if o.Kind == L_String {
		z := &LitStringX{o.Word}
		o.Next()
		return z
	}
	if o.Kind == L_Char {
		z := &LitIntX{int(o.Word[0])}
		o.Next()
		return z
	}
	if o.Kind == L_Ident {
		z := &IdentX{o.Word}
		o.Next()
		return z
	}
	panic("bad ParsePrim")
}

func (o *Parser) ParsePrimEtc() Expr {
	a := o.ParsePrim()
	if o.Word == "(" {
		o.TakePunc("(")
		if o.Word != ")" {
			args := o.ParseList()
			a = &CallX{a, args}
		}
		o.TakePunc(")")
	}
	return a
}

func (o *Parser) ParseProduct() Expr {
	a := o.ParsePrimEtc()
	op := o.Word
	for op == "*" || op == "/" || op == "%" || op == "<<" || op == ">>" || op == "&" || op == "&^" {
		o.Next()
		b := o.ParsePrimEtc()
		a = &BinOpX{a, op, b}
		op = o.Word
	}
	return a
}

func (o *Parser) ParseSum() Expr {
	a := o.ParseProduct()
	op := o.Word
	for op == "+" || op == "-" || op == "|" || op == "^" {
		o.Next()
		b := o.ParseSum()
		a = &BinOpX{a, op, b}
		op = o.Word
	}
	return a
}

func (o *Parser) ParseRelational() Expr {
	a := o.ParseSum()
	op := o.Word
	for o.Word == "==" || o.Word == "!=" || o.Word == "<" || o.Word == ">" || o.Word == "<=" || o.Word == ">=" {
		o.Next()
		b := o.ParseRelational()
		a = &BinOpX{a, op, b}
		op = o.Word
	}
	return a
}

func (o *Parser) ParseAnd() Expr {
	a := o.ParseRelational()
	for o.Word == "&&" {
		o.Next()
		b := o.ParseAnd()
		a = &BinOpX{a, "&&", b}
	}
	return a
}

func (o *Parser) ParseOr() Expr {
	a := o.ParseAnd()
	for o.Word == "||" {
		o.Next()
		b := o.ParseOr()
		a = &BinOpX{a, "||", b}
	}
	return a
}

func (o *Parser) ParseExpr() Expr {
	return o.ParseOr()
}

func (o *Parser) ParseType() Type {
	w := o.TakeIdent()
	switch w {
	case "byte":
		return Byte
	case "int":
		return Int
	case "uint":
		return UInt
	}
	log.Panicf("expected a type, got %q", w)
	return nil
}

func (o *Parser) ParseList() Expr {
	a := o.ParseExpr()
	if o.Word == "," {
		v := []Expr{a}
		for o.Word == "," {
			o.Next()
			b := o.ParseExpr()
			v = append(v, b)
		}
		return &ListX{v}
	}
	return a
}

func (o *Parser) ParseAssignment() Stmt {
	a := o.ParseList()
	op := o.Word
	if op == "=" || len(op) == 2 && op[1] == '=' {
		o.Next()
		b := o.ParseList()
		return &AssignS{a, op, b}
	}
	return &AssignS{nil, "", a}
}

func (o *Parser) TakePunc(s string) {
	if o.Kind != L_Punc || s != o.Word {
		log.Panicf("expected %q, got (%d) %q", s, o.Kind, o.Word)
	}
	o.Next()
}

func (o *Parser) TakeIdent() string {
	if o.Kind != L_Ident {
		log.Panicf("expected Ident, got (%d) %q", o.Kind, o.Word)
	}
	w := o.Word
	o.Next()
	return w
}

func (o *Parser) TakeEOL() {
	if o.Kind != L_EOL {
		log.Panicf("expected EOL, got (%d) %q", o.Kind, o.Word)
	}
	o.Next()
}

func (o *Parser) ParseBlock(b *Block) {
	o.TakePunc("{")
BLOCK:
	for o.Word != "}" {
		switch o.Kind {
		case L_EOL:
			o.TakeEOL()
			continue BLOCK
		case L_Ident:
			switch o.Word {
			case "var":
				o.Next()
				s := o.TakeIdent()
				t := o.ParseType()
				b.Locals = append(b.Locals, NameAndType{s, t})
			case "return":
				o.Next()
				var xx Expr
				if o.Kind != L_EOL {
					xx = o.ParseList()
				}
				b.Body = append(b.Body, &ReturnS{xx})
			default:
				a := o.ParseAssignment()
				b.Body = append(b.Body, a)
			}
			o.TakeEOL()
		}
	}
	o.TakePunc("}")
}

func (o *Parser) ParseFunc(fn *DefFunc) {
	o.TakePunc("(")
	for o.Word != ")" {
		s := o.TakeIdent()
		t := o.ParseType()
		fn.Ins = append(fn.Ins, NameAndType{s, t})
		if o.Word == "," {
			o.TakePunc(",")
		} else if o.Word != ")" {
			log.Panicf("expected `,` or `)` but got %q", o.Word)
		}
	}
	o.TakePunc(")")
	if o.Word != "{" {
		if o.Word == "{" {
			o.TakePunc("(")
			for o.Word != ")" {
				s := o.TakeIdent()
				t := o.ParseType()
				fn.Outs = append(fn.Outs, NameAndType{s, t})
				if o.Word == "," {
					o.TakePunc(",")
				} else if o.Word != ")" {
					log.Panicf("expected `,` or `)` but got %q", o.Word)
				}
			}
			o.TakePunc(")")
		} else {
			t := o.ParseType()
			fn.Outs = append(fn.Outs, NameAndType{"", t})
		}
	}
	b := &Block{Fn: fn}
	o.ParseBlock(b)
	fn.Body = b
}

func (o *Parser) ParseTop() {
LOOP:
	for {
		switch o.Kind {
		case L_Ident:
			d := o.TakeIdent()
			switch d {
			case "package":
				w := o.TakeIdent()
				o.Package = &DefPackage{Name: w}
			case "import":
				w := o.TakeIdent()
				o.Imports[w] = &DefImport{Name: w}
			case "const":
				w := o.TakeIdent()
				o.TakePunc("=")
				x := o.ParseExpr()
				o.Consts[w] = &DefConst{Name: w, Expr: x}
			case "var":
				w := o.TakeIdent()
				t := o.ParseType()
				o.Vars[w] = &DefVar{Name: w, Type: t}
			case "type":
				w := o.TakeIdent()
				o.Types[w] = &DefType{Name: w}
			case "func":
				w := o.TakeIdent()
				fn := &DefFunc{Name: w}
				o.ParseFunc(fn)
				o.Funcs[w] = fn
			default:
				log.Panicf("Expected top level decl, got %q", d)
			}
			o.TakeEOL()
		case L_EOL:
			o.TakeEOL()
			continue LOOP
		case L_EOF:
			break LOOP
		default:
			log.Panicf("expected toplevel decl; got (%d) %q", o.Kind, o.Word)
		}
	}
}

func CompileToC(r io.Reader, sourceName string, w io.Writer) {
	p := NewParser(r, sourceName)
	p.ParseTop()
	cg := NewCGen(w)
	cg.P("#include <stdio.h>")
	cg.P("#include \"runtime_c.h\"")
	cg.VisitDefPackage(p.Package)
	for _, i := range p.Imports {
		cg.VisitDefImport(i)
	}
	for _, c := range p.Consts {
		cg.VisitDefConst(c)
	}
	for _, t := range p.Types {
		cg.VisitDefType(t)
	}
	for _, v := range p.Vars {
		cg.VisitDefVar(v)
	}
	for _, f := range p.Funcs {
		cg.VisitDefFunc(f)
	}
	cg.Flush()
}

type CGen struct {
	W *bufio.Writer
}

func NewCGen(w io.Writer) *CGen {
	cg := &CGen{
		W: bufio.NewWriter(w),
	}
	return cg
}
func (cg *CGen) P(format string, args ...interface{}) {
	fmt.Fprintf(cg.W, format+"\n", args...)
}
func (cg *CGen) Flush() {
	cg.W.Flush()
}
func (cg *CGen) VisitLitInt(x *LitIntX) string {
	return fmt.Sprintf("%d", x.X)
}
func (cg *CGen) VisitLitString(x *LitStringX) string {
	return fmt.Sprintf("%q", x.X)
}
func (cg *CGen) VisitIdent(x *IdentX) string {
	return "v_" + x.X
}
func (cg *CGen) VisitBinOp(x *BinOpX) string {
	a := x.A.VisitExpr(cg)
	b := x.B.VisitExpr(cg)
	return fmt.Sprintf("((%s) %s (%s))", a, x.Op, b)
}
func (cg *CGen) VisitList(x *ListX) string {
	return "PROBLEM:VisitList"
}
func (cg *CGen) VisitCall(x *CallX) string {
	return " -- need to know the results from the type -- "
}
func (cg *CGen) VisitAssign(*AssignS) {
}
func (cg *CGen) VisitReturn(*ReturnS) {
}
func (cg *CGen) VisitDefPackage(def *DefPackage) {
	cg.P("// package %s: %#v", def.Name, def)
}
func (cg *CGen) VisitDefImport(def *DefImport) {
	cg.P("// import %s: %#v", def.Name, def)
}
func (cg *CGen) VisitDefConst(def *DefConst) {
	cg.P("// const %s: %#v", def.Name, def)
}
func (cg *CGen) VisitDefVar(def *DefVar) {
	cg.P("// var %s: %#v", def.Name, def)
}
func (cg *CGen) VisitDefType(def *DefType) {
	cg.P("// type %s: %#v", def.Name, def)
}
func (cg *CGen) VisitDefFunc(def *DefFunc) {
	cg.P("// func %s: %#v", def.Name, def)
}

/*
func (cg *CGen) VisitIntType(*IntType) {
}
*/
