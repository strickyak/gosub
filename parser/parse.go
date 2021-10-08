package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
)

var Format = fmt.Sprintf

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
	if o.Kind == L_Punc {
		if o.Word == "[" {
			return &TypeX{o.ParseType()}
		}
	}
	panic("bad ParsePrim")
}

func (o *Parser) ParsePrimEtc() Expr {
	a := o.ParsePrim()
LOOP:
	for {
		switch o.Word {
		case "(":
			o.TakePunc("(")
			if o.Word != ")" {
				args := o.ParseList()
				a = &CallX{a, args}
			}
			o.TakePunc(")")
		case "[":
			o.TakePunc("[")
			sub := o.ParseExpr()
			o.TakePunc("]")
			a = &SubX{a, sub}
		case ".":
			o.TakePunc(".")
			member := o.TakeIdent()
			a = &DotX{a, member}
		default:
			break LOOP
		}
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
	switch o.Kind {
	case L_Ident:
		w := o.TakeIdent()
		switch w {
		case "bool":
			return BoolType
		case "byte":
			return ByteType
		case "int":
			return IntType
		case "uint":
			return UintType
		}
		Panicf("expected a type, got %q", w)
	case L_Punc:
		if o.Word == "[" {
			o.Next()
			if o.Word != "]" {
				Panicf("for slice type, after [ expected ], got %v", o.Word)
			}
			o.Next()
			memberType := o.ParseType()
			return Type(Format(SliceForm, memberType))
		}
	}
	Panicf("not a type: starts with %v", o.Word)
	panic("notreached")
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
	} else if op == "++" {
		o.Next()
		return &AssignS{a, op, nil}
	} else if op == "--" {
		o.Next()
		return &AssignS{a, op, nil}
	}
	return &AssignS{nil, "", a}
}

func (o *Parser) TakePunc(s string) {
	if o.Kind != L_Punc || s != o.Word {
		Panicf("expected %q, got (%d) %q", s, o.Kind, o.Word)
	}
	o.Next()
}

func (o *Parser) TakeIdent() string {
	if o.Kind != L_Ident {
		Panicf("expected Ident, got (%d) %q", o.Kind, o.Word)
	}
	w := o.Word
	o.Next()
	return w
}

func (o *Parser) TakeEOL() {
	if o.Kind != L_EOL {
		Panicf("expected EOL, got (%d) %q", o.Kind, o.Word)
	}
	o.Next()
}

func (o *Parser) ParseStmt(b *Block) Stmt {
	switch o.Word {
	case "if":
		o.Next()
		pred := o.ParseExpr()
		yes := o.ParseBlock(b.Func)
		var no *Block
		if o.Word == "else" {
			no = o.ParseBlock(b.Func)
		}
		return &IfS{pred, yes, no}
	case "for":
		o.Next()
		var pred Expr
		if o.Word != "{" {
			pred = o.ParseExpr()
		}
		b2 := o.ParseBlock(b.Func)
		return &WhileS{pred, b2}
	case "switch":
		o.Next()
		var subject Expr
		if o.Word != "{" {
			subject = o.ParseExpr()
		}
		o.TakePunc("{")
		sws := &SwitchS{subject, nil}
		for o.Word != "}" {
			for o.Word == ";;" {
				o.Next()
			}
			cOrD := o.TakeIdent()
			switch cOrD {
			case "case":
				exprs := o.ParseList()
				o.TakePunc(":")
				bare := o.ParseBareBlock(b.Func)
				sws.Entries = append(sws.Entries, &SwitchEntry{exprs, bare})
			case "default":
				o.TakePunc(":")
				bare := o.ParseBareBlock(b.Func)
				sws.Entries = append(sws.Entries, &SwitchEntry{nil, bare})
			default:
				panic(cOrD)
			}
		}
		o.TakePunc("}")
		return sws
	case "return":
		o.Next()
		var xx []Expr
		if o.Kind != L_EOL {
			xx = o.ParseList()
		}
		return &ReturnS{xx}
	case "break":
		o.Next()
		break_to := ""
		if o.Kind == L_Ident {
			break_to = o.Word
			o.Next()
		}
		return &BreakS{break_to}
	case "continue":
		o.Next()
		continue_to := ""
		if o.Kind == L_Ident {
			continue_to = o.Word
			o.Next()
		}
		return &ContinueS{continue_to}
	default:
		a := o.ParseAssignment()
		return a
	}
}
func (o *Parser) ParseBlock(fn *DefFunc) *Block {
	o.TakePunc("{")
	b := o.ParseBareBlock(fn)
	o.TakePunc("}")
	return b
}
func (o *Parser) ParseBareBlock(fn *DefFunc) *Block {
	b := &Block{Func: fn}
	for o.Word != "}" && o.Word != "case" && o.Word != "default" {
		switch o.Kind {
		case L_EOL:
			o.TakeEOL()
		case L_Ident:
			switch o.Word {
			case "var":
				o.Next()
				s := o.TakeIdent()
				t := o.ParseType()
				b.Locals = append(b.Locals, NameAndType{s, t})
			default:
				stmt := o.ParseStmt(b)
				if stmt != nil {
					b.Stmts = append(b.Stmts, stmt)
				}
				o.TakeEOL()
			}
		}
	}
	return b
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
			Panicf("expected `,` or `)` but got %q", o.Word)
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
					Panicf("expected `,` or `)` but got %q", o.Word)
				}
			}
			o.TakePunc(")")
		} else {
			t := o.ParseType()
			fn.Outs = append(fn.Outs, NameAndType{"", t})
		}
	}
	b := o.ParseBlock(fn)
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
				if o.Kind != L_String {
					Panicf("after import, expected string, got %v", o.Word)
				}
				w := o.Word
				o.Next()
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
				Panicf("Expected top level decl, got %q", d)
			}
			o.TakeEOL()
		case L_EOL:
			o.TakeEOL()
			continue LOOP
		case L_EOF:
			break LOOP
		default:
			Panicf("expected toplevel decl; got (%d) %q", o.Kind, o.Word)
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
	cm := cg.Mods["main"]

	cm.Globals["println"] = NameAndType{"F_BUILTIN_println", ""}

	cm.P("#include <stdio.h>")
	cm.P("#include \"runt.h\"")

	cm.pre.VisitDefPackage(p.Package)
	for _, i := range p.Imports {
		cm.pre.VisitDefImport(i)
	}
	for _, c := range p.Consts {
		cm.pre.VisitDefConst(c)
	}
	for _, t := range p.Types {
		cm.pre.VisitDefType(t)
	}
	for _, v := range p.Vars {
		cm.pre.VisitDefVar(v)
	}
	for _, f := range p.Funcs {
		cm.pre.VisitDefFunc(f)
	}

	cm.VisitDefPackage(p.Package)
	cm.P("// ..... Imports .....")
	for _, i := range p.Imports {
		cm.VisitDefImport(i)
	}
	cm.P("// ..... Consts .....")
	for _, c := range p.Consts {
		cm.VisitDefConst(c)
	}
	cm.P("// ..... Types .....")
	for _, t := range p.Types {
		cm.VisitDefType(t)
	}
	cm.P("// ..... Vars .....")
	for _, v := range p.Vars {
		cm.VisitDefVar(v)
	}
	cm.P("// ..... Funcs .....")
	for _, f := range p.Funcs {
		cm.VisitDefFunc(f)
	}
	cm.P("// ..... Done .....")

	cm.Flush()
}

type cPreMod struct {
	cm *CMod
	// cg *CGen
}

func (pre *cPreMod) VisitDefPackage(def *DefPackage) {
	pre.cm.Package = def.Name
}
func (pre *cPreMod) VisitDefImport(def *DefImport) {
	pre.cm.Globals[def.Name] = NameAndType{"I_" + pre.cm.Package + "__" + def.Name, ""}
}
func (pre *cPreMod) VisitDefConst(def *DefConst) {
	pre.cm.Globals[def.Name] = NameAndType{"C_" + pre.cm.Package + "__" + def.Name, ConstIntType}
}
func (pre *cPreMod) VisitDefVar(def *DefVar) {
	pre.cm.Globals[def.Name] = NameAndType{"V_" + pre.cm.Package + "__" + def.Name, IntType}
}
func (pre *cPreMod) VisitDefType(def *DefType) {
	pre.cm.Globals[def.Name] = NameAndType{"T_" + pre.cm.Package + "__" + def.Name, ""}
}
func (pre *cPreMod) VisitDefFunc(def *DefFunc) {
	pre.cm.Globals[def.Name] = NameAndType{"F_" + pre.cm.Package + "__" + def.Name, ""}

	// TODO -- dedup
	var b Buf
	cfunc := Format("F_%s__%s", pre.cm.Package, def.Name)
	crettype := "void"
	if len(def.Outs) > 0 {
		if len(def.Outs) > 1 {
			panic("multi")
		}
		crettype = TypeNameInC(def.Outs[0].Type)
	}
	b.P("%s %s(", crettype, cfunc)
	if len(def.Ins) > 0 {
		firstTime := true
		for _, name_and_type := range def.Ins {
			if !firstTime {
				b.P(", ")
			}
			b.P("%s %s", TypeNameInC(name_and_type.Type), "v_"+name_and_type.Name)
			firstTime = false
		}
	}
	b.P(");\n")
	pre.cm.P(b.String())
}

type CMod struct {
	pre        *cPreMod
	W          *bufio.Writer
	Package    string
	Globals    map[string]NameAndType
	BreakTo    string
	ContinueTo string
}
type CGen struct {
	Mods map[string]*CMod
}

func NewCGen(w io.Writer) *CGen {
	mainMod := &CMod{
		pre:     new(cPreMod),
		W:       bufio.NewWriter(w),
		Package: "main",
		Globals: make(map[string]NameAndType),
	}
	mainMod.pre.cm = mainMod
	cg := &CGen{
		Mods: map[string]*CMod{"main": mainMod},
	}
	return cg
}
func (cm *CMod) P(format string, args ...interface{}) {
	log.Printf("<<<<< %q >>>>> %q", format, fmt.Sprintf(format, args...))
	fmt.Fprintf(cm.W, format+"\n", args...)
}
func (cm *CMod) Flush() {
	cm.W.Flush()
}
func (cm *CMod) VisitLitInt(x *LitIntX) Value {
	return &VSimple{
		C: Format("%d", x.X),
		T: IntType,
	}
}
func (cm *CMod) VisitLitString(x *LitStringX) Value {
	return &VSimple{
		C: Format("%q", x.X),
		T: IntType,
	}
}
func (cm *CMod) VisitIdent(x *IdentX) Value {
	if gl, ok := cm.Globals[x.X]; ok {
		return &VSimple{C: gl.Name, T: /*TODO*/ IntType}
	}
	// Assume it is a local variable.
	return &VSimple{C: "v_" + x.X, T: IntType}
}
func (cm *CMod) VisitBinOp(x *BinOpX) Value {
	a := x.A.VisitExpr(cm)
	b := x.B.VisitExpr(cm)
	return &VSimple{
		C: Format("(%s) %s (%s)", a.ToC(), x.Op, b.ToC()),
		T: IntType,
	}
}
func (cm *CMod) VisitList(x *ListX) Value {
	return &VSimple{
		C: "PROBLEM:VisitList",
		T: IntType,
	}
}
func (cm *CMod) VisitCall(x *CallX) Value {
	cargs := ""
	firstTime := true
	for _, e := range x.Args {
		if !firstTime {
			cargs += ", "
		}
		cargs += e.VisitExpr(cm).ToC()
		firstTime = false
	}
	cfunc := x.Func.VisitExpr(cm).ToC()
	ccall := Format("(%s(%s))", cfunc, cargs)
	return &VSimple{
		C: ccall,
		T: IntType,
	}
}
func (cm *CMod) VisitType(x *TypeX) Value {
	return &VSimple{
		C: string(x.T),
		T: x.T,
	}
}
func (cm *CMod) VisitSub(x *SubX) Value {
	return &VSimple{
		C: Format("SubXXX(%v)", x),
		T: "",
	}
}
func (cm *CMod) VisitDot(x *DotX) Value {
	return &VSimple{
		C: Format("DotXXX(%v)", x),
		T: "",
	}
}
func (cm *CMod) VisitAssign(ass *AssignS) {
	log.Printf("assign..... %v ; %v ; %v", ass.A, ass.Op, ass.B)
	var values []Value
	for _, e := range ass.B {
		values = append(values, e.VisitExpr(cm))
	}
	var bcall *CallX
	if len(ass.B) == 1 {
		switch t := ass.B[0].(type) {
		case *CallX:
			bcall = t
		}
	}
	if ass.A == nil {
		if bcall == nil {
			for _, val := range values {
				cm.P("  (void)(%s);", val.ToC())
			}
		} else {
			// call with outputs
		}
	} else if ass.B == nil {
		if len(ass.A) != 1 {
			Panicf("operator %v requires one lvalue on the left, got %v", ass.Op, ass.A)
		}
		// TODO check Globals
		lhs := ass.A[0]
		switch t := lhs.(type) {
		case *IdentX:
			cvar := "v_" + t.X
			cm.P("  %s %s;", cvar, ass.Op)
		default:
			Panicf("operator %v: lhs not supported: %v", ass.Op, lhs)
		}
	} else if len(ass.A) > 1 && bcall != nil {
		// From 1 call, to 2 or more assigned vars.
		var buf Buf
		buf.P("((%s)(", bcall.Func.VisitExpr(cm).ToC())
		for i, arg := range bcall.Args {
			if i > 0 {
				buf.P(", ")
			}
			buf.P("%s", arg.VisitExpr(cm).ToC())
		}
		for i, arg := range ass.A {
			if len(bcall.Args)+i > 0 {
				buf.P(", ")
			}
			// TODO -- VisitAddr ?
			buf.P("&(%s)", arg.VisitExpr(cm).ToC())
		}
		buf.P("))")
	} else {
		if len(ass.A) != len(ass.B) {
			Panicf("wrong number of values in assign")
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
					cm.P("  %s = (%s)(%s);", cvar, TypeNameInC(val.Type()), val.ToC())
				case ":=":
					// TODO check Globals
					cvar := Format("%s %s", TypeNameInC(val.Type()), "v_"+t.X)
					cm.P("  %s = (%s)(%s);", cvar, TypeNameInC(val.Type()), val.ToC())
				}
			default:
				log.Fatal("bad VisitAssign LHS: %#v", ass.A)
			}
		}
	}
}
func (cm *CMod) VisitReturn(ret *ReturnS) {
	log.Printf("return..... %v", ret.X)
	switch len(ret.X) {
	case 0:
		cm.P("  return;")
	case 1:
		val := ret.X[0].VisitExpr(cm)
		log.Printf("return..... val=%v", val)
		cm.P("  return %s;", val.ToC())
	default:
		Panicf("multi-return not imp: %v", ret)
	}
}
func (cm *CMod) VisitWhile(wh *WhileS) {
	label := Serial("while")
	cm.P("Break_%s:  while(1) {", label)
	if wh.Pred != nil {
		cm.P("    t_bool _while_ = (t_bool)(%s);", wh.Pred.VisitExpr(cm).ToC())
		cm.P("    if (!_while_) break;")
	}
	savedB, savedC := cm.BreakTo, cm.ContinueTo
	cm.BreakTo, cm.ContinueTo = "Break_"+label, "Cont_"+label
	wh.Body.VisitStmt(cm)
	cm.P("  }")
	cm.P("Cont_%s: {}", label)
	cm.BreakTo, cm.ContinueTo = savedB, savedC
}
func (cm *CMod) VisitBreak(sws *BreakS) {
	if cm.BreakTo == "" {
		Panicf("cannot break from here")
	}
	cm.P("goto %s;", cm.BreakTo)
}
func (cm *CMod) VisitContinue(sws *ContinueS) {
	if cm.ContinueTo == "" {
		Panicf("cannot continue from here")
	}
	cm.P("goto %s;", cm.ContinueTo)
}
func (cm *CMod) VisitIf(ifs *IfS) {
	cm.P("  { t_bool _if_ = %s;", ifs.Pred.VisitExpr(cm).ToC())
	cm.P("  if( _if_ ) {")
	ifs.Yes.VisitStmt(cm)
	if ifs.No != nil {
		cm.P("  } else {")
		ifs.No.VisitStmt(cm)
	}
	cm.P("  }}")
}
func (cm *CMod) VisitSwitch(sws *SwitchS) {
	cm.P("  { t_int _switch_ = %s;", sws.Pred.VisitExpr(cm).ToC())
	cm.P("  TODO(switch)")
	cm.P("  }")
}
func (cm *CMod) VisitBlock(a *Block) {
	if a == nil {
		panic(8881)
	}
	if a.Stmts == nil {
		panic(8882)
	}
	for i, e := range a.Stmts {
		log.Printf("VisitBlock[%d]", i)
		e.VisitStmt(cm)
	}
}
func (cm *CMod) VisitDefPackage(def *DefPackage) {
	cm.P("// package %s", def.Name)
	cm.Package = def.Name
}
func (cm *CMod) VisitDefImport(def *DefImport) {
	cm.P("// import %s", def.Name)
}
func (cm *CMod) VisitDefConst(def *DefConst) {
	cm.P("// const %s", def.Name)
}
func (cm *CMod) VisitDefVar(def *DefVar) {
	cm.P("// var %s", def.Name)
}
func (cm *CMod) VisitDefType(def *DefType) {
	cm.P("// type %s", def.Name)
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
func (cm *CMod) VisitDefFunc(def *DefFunc) {
	log.Printf("// func %s: %#v", def.Name, def)
	var b Buf
	cfunc := Format("F_%s__%s", cm.Package, def.Name)
	crettype := "void"
	if len(def.Outs) > 0 {
		if len(def.Outs) > 1 {
			panic("multi")
		}
		crettype = TypeNameInC(def.Outs[0].Type)
	}
	b.P("%s %s(", crettype, cfunc)
	if len(def.Ins) > 0 {
		firstTime := true
		for _, name_and_type := range def.Ins {
			if !firstTime {
				b.P(", ")
			}
			b.P("%s %s", TypeNameInC(name_and_type.Type), "v_"+name_and_type.Name)
			firstTime = false
		}
	}
	b.P(") {\n")
	cm.P(b.String())
	def.Body.VisitStmt(cm)
	cm.P("}\n")
}

var SerialNum uint

func Serial(prefix string) string {
	SerialNum++
	return Format("%s_%d", prefix, SerialNum)
}
