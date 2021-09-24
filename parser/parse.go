package parser

import (
	"bufio"
	"bytes"
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

func (o *Parser) ParseList() []Expr {
	xx := []Expr{o.ParseExpr()}
	for o.Word == "," {
		o.Next()
		xx = append(xx, o.ParseExpr())
	}
	return xx
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
			case "for":
				o.Next()
				var pred Expr
				if o.Word != "{" {
					pred = o.ParseExpr()
				}
				b2 := &Block{Func: b.Func}
				o.ParseBlock(b2)
				b.Stmts = append(b.Stmts, &WhileS{pred, b2})
			case "return":
				o.Next()
				var xx []Expr
				if o.Kind != L_EOL {
					xx = o.ParseList()
				}
				b.Stmts = append(b.Stmts, &ReturnS{xx})
			default:
				a := o.ParseAssignment()
				b.Stmts = append(b.Stmts, a)
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
	b := &Block{Func: fn}
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

type Value interface {
	Type() Type
	ToC() string
}

type VSimple struct {
	C string // C language expression
	T Type
}

func (val *VSimple) ToC() string {
	return val.C
}
func (val *VSimple) Type() Type {
	return val.T
}

func CompileToC(r io.Reader, sourceName string, w io.Writer) {
	p := NewParser(r, sourceName)
	p.ParseTop()
	cg := NewCGen(w)

	cg.Globals["println"] = NameAndType{"F_BUILTIN_println", nil}

	cg.P("#include <stdio.h>")
	cg.P("#include \"runtime_c.h\"")

	cg.pre.VisitDefPackage(p.Package)
	for _, i := range p.Imports {
		cg.pre.VisitDefImport(i)
	}
	for _, c := range p.Consts {
		cg.pre.VisitDefConst(c)
	}
	for _, t := range p.Types {
		cg.pre.VisitDefType(t)
	}
	for _, v := range p.Vars {
		cg.pre.VisitDefVar(v)
	}
	for _, f := range p.Funcs {
		cg.pre.VisitDefFunc(f)
	}

	cg.VisitDefPackage(p.Package)
	cg.P("// ..... Imports .....")
	for _, i := range p.Imports {
		cg.VisitDefImport(i)
	}
	cg.P("// ..... Consts .....")
	for _, c := range p.Consts {
		cg.VisitDefConst(c)
	}
	cg.P("// ..... Types .....")
	for _, t := range p.Types {
		cg.VisitDefType(t)
	}
	cg.P("// ..... Vars .....")
	for _, v := range p.Vars {
		cg.VisitDefVar(v)
	}
	cg.P("// ..... Funcs .....")
	for _, f := range p.Funcs {
		cg.VisitDefFunc(f)
	}
	cg.P("// ..... Done .....")

	cg.Flush()
}

type cPreGen struct {
	cg *CGen
}

/*
func (cg *cPreGen) VisitLitInt(x *LitIntX) Value {}
func (cg *cPreGen) VisitLitString(x *LitStringX) Value {}
func (cg *cPreGen) VisitIdent(x *IdentX) Value {}
func (cg *cPreGen) VisitBinOp(x *BinOpX) Value {}
func (cg *cPreGen) VisitList(x *ListX) Value {}
func (cg *cPreGen) VisitCall(x *CallX) Value {}
func (cg *cPreGen) VisitAssign(ass *AssignS) {}
func (cg *cPreGen) VisitReturn(ret *ReturnS) {}
func (cg *cPreGen) VisitWhile(ret *ReturnS) {}
func (cg *cPreGen) VisitBlock(a *Block) {}
func (cg *cPreGen) VisitIntType(*IntType) {}
*/
func (pre *cPreGen) VisitDefPackage(def *DefPackage) {
	pre.cg.Package = def.Name
}
func (pre *cPreGen) VisitDefImport(def *DefImport) {
	pre.cg.Globals[def.Name] = NameAndType{"I_" + pre.cg.Package + "__" + def.Name, nil}
}
func (pre *cPreGen) VisitDefConst(def *DefConst) {
	pre.cg.Globals[def.Name] = NameAndType{"C_" + pre.cg.Package + "__" + def.Name, ConstInt}
}
func (pre *cPreGen) VisitDefVar(def *DefVar) {
	pre.cg.Globals[def.Name] = NameAndType{"V_" + pre.cg.Package + "__" + def.Name, Int}
}
func (pre *cPreGen) VisitDefType(def *DefType) {
	pre.cg.Globals[def.Name] = NameAndType{"T_" + pre.cg.Package + "__" + def.Name, nil}
}
func (pre *cPreGen) VisitDefFunc(def *DefFunc) {
	pre.cg.Globals[def.Name] = NameAndType{"F_" + pre.cg.Package + "__" + def.Name, nil}

	// TODO -- dedup
	var b Buf
	cfunc := fmt.Sprintf("F_%s__%s", pre.cg.Package, def.Name)
	crettype := "void"
	if len(def.Outs) > 0 {
		if len(def.Outs) > 1 {
			panic("multi")
		}
		crettype = def.Outs[0].Type.TypeNameInC("")
	}
	b.P("%s %s(", crettype, cfunc)
	if len(def.Ins) > 0 {
		firstTime := true
		for _, name_and_type := range def.Ins {
			if !firstTime {
				b.P(", ")
			}
			b.P("%s", name_and_type.Type.TypeNameInC("v_"+name_and_type.Name))
			firstTime = false
		}
	}
	b.P(");\n")
	pre.cg.P(b.String())
}

type CGen struct {
	pre     *cPreGen
	W       *bufio.Writer
	Package string
	Globals map[string]NameAndType
}

func NewCGen(w io.Writer) *CGen {
	cg := &CGen{
		pre:     new(cPreGen),
		W:       bufio.NewWriter(w),
		Globals: make(map[string]NameAndType),
	}
	cg.pre.cg = cg
	return cg
}
func (cg *CGen) P(format string, args ...interface{}) {
	fmt.Fprintf(cg.W, format+"\n", args...)
}
func (cg *CGen) Flush() {
	cg.W.Flush()
}
func (cg *CGen) VisitLitInt(x *LitIntX) Value {
	return &VSimple{
		C: fmt.Sprintf("%d", x.X),
		T: Int,
	}
}
func (cg *CGen) VisitLitString(x *LitStringX) Value {
	return &VSimple{
		C: fmt.Sprintf("%q", x.X),
		T: Int,
	}
}
func (cg *CGen) VisitIdent(x *IdentX) Value {
	if gl, ok := cg.Globals[x.X]; ok {
		return &VSimple{C: gl.Name, T: /*TODO*/ Int}
	}
    // Assume it is a local variable.
	return &VSimple{C: "v_" + x.X, T: Int}
}
func (cg *CGen) VisitBinOp(x *BinOpX) Value {
	a := x.A.VisitExpr(cg)
	b := x.B.VisitExpr(cg)
	return &VSimple{
		C: fmt.Sprintf("(%s) %s (%s)", a.ToC(), x.Op, b.ToC()),
		T: Int,
	}
}
func (cg *CGen) VisitList(x *ListX) Value {
	return &VSimple{
		C: "PROBLEM:VisitList",
		T: Int,
	}
}
func (cg *CGen) VisitCall(x *CallX) Value {
	cargs := ""
	firstTime := true
	for _, e := range x.Args {
		if !firstTime {
			cargs += ", "
		}
		cargs += e.VisitExpr(cg).ToC()
		firstTime = false
	}
	cfunc := x.Func.VisitExpr(cg).ToC()
	ccall := fmt.Sprintf("(%s(%s))", cfunc, cargs)
	return &VSimple{
		C: ccall,
		T: Int,
	}
}
func (cg *CGen) VisitAssign(ass *AssignS) {
	log.Printf("assign..... %v ; %v ; %v", ass.A, ass.Op, ass.B)
	var values []Value
	for _, e := range ass.B {
		values = append(values, e.VisitExpr(cg))
	}
	if ass.A == nil {
		for _, val := range values {
			cg.P("  (void)(%s);", val.ToC())
		}
	} else {
		if len(ass.A) != len(ass.B) {
			log.Panicf("wrong number of values in assign")
		}
		for i, val := range values {
			lhs := ass.A[i]
			switch t := lhs.(type) {
			case *IdentX:
				// TODO -- check that variable t.X has the right type.
				switch ass.Op {
				case "=":
					// TODO check Globals
					cvar := "v_" + t.X
					cg.P("  %s = (%s)(%s);", cvar, val.Type().TypeNameInC(""), val.ToC())
				case ":=":
					// TODO check Globals
					cvar := val.Type().TypeNameInC("v_" + t.X)
					cg.P("  %s = (%s)(%s);", cvar, val.Type().TypeNameInC(""), val.ToC())
				}
			default:
				log.Fatal("bad VisitAssign LHS: %#v", ass.A)
			}
		}
	}
}
func (cg *CGen) VisitReturn(ret *ReturnS) {
	log.Printf("return..... %v", ret.X)
	switch len(ret.X) {
	case 0:
		cg.P("  return;")
	case 1:
		val := ret.X[0].VisitExpr(cg)
		log.Printf("return..... val=%v", val)
		cg.P("  return %s;", val.ToC())
	default:
		log.Panicf("multi-return not imp: %v", ret)
	}
}
func (cg *CGen) VisitWhile(wh *WhileS) {
	cg.P("  while(1) {")
	cg.P("    t_bool _while_ = (t_bool)(%s);", wh.Pred.VisitExpr(cg).ToC())
	cg.P("    if (!_while_) break;")
	wh.Body.VisitStmt(cg)
	cg.P("  }")
}
func (cg *CGen) VisitBlock(a *Block) {
	for i, e := range a.Stmts {
		log.Printf("VisitBlock[%d]", i)
		e.VisitStmt(cg)
	}
}
func (cg *CGen) VisitDefPackage(def *DefPackage) {
	cg.P("// package %s", def.Name)
	cg.Package = def.Name
}
func (cg *CGen) VisitDefImport(def *DefImport) {
	cg.P("// import %s", def.Name)
}
func (cg *CGen) VisitDefConst(def *DefConst) {
	cg.P("// const %s", def.Name)
}
func (cg *CGen) VisitDefVar(def *DefVar) {
	cg.P("// var %s", def.Name)
}
func (cg *CGen) VisitDefType(def *DefType) {
	cg.P("// type %s", def.Name)
}

type Buf struct {
	W bytes.Buffer
}

func (buf *Buf) P(format string, args ...interface{}) {
	fmt.Fprintf(&buf.W, format, args...)
}
func (buf *Buf) String() string {
	return buf.W.String()
}
func (cg *CGen) VisitDefFunc(def *DefFunc) {
	log.Printf("// func %s: %#v", def.Name, def)
	var b Buf
	cfunc := fmt.Sprintf("F_%s__%s", cg.Package, def.Name)
	crettype := "void"
	if len(def.Outs) > 0 {
		if len(def.Outs) > 1 {
			panic("multi")
		}
		crettype = def.Outs[0].Type.TypeNameInC("")
	}
	b.P("%s %s(", crettype, cfunc)
	if len(def.Ins) > 0 {
		firstTime := true
		for _, name_and_type := range def.Ins {
			if !firstTime {
				b.P(", ")
			}
			b.P("%s", name_and_type.Type.TypeNameInC("v_"+name_and_type.Name))
			firstTime = false
		}
	}
	b.P(") {\n")
	cg.P(b.String())
	def.Body.VisitStmt(cg)
	cg.P("}\n")
}

/*
func (cg *CGen) VisitIntType(*IntType) {
}
*/
