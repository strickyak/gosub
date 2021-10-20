package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"sort"
)

var Format = fmt.Sprintf

// FullName creates the mangled name with kind and package.
func FullName(kind string, pkg string, name string) string {
	return Format("%s_%s__%s", kind, pkg, name)
}

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

var ErrorTV = &InterfaceTV{
	InterfaceRec: &InterfaceRec{
		Name:   "error",
		Fields: []NameAndType{
			// TODO: Error func() string
		},
	},
}
var BuiltinTypeIdentifiers = map[string]TypeValue{
	"bool":   &PrimTV{BoolType},
	"byte":   &PrimTV{ByteType},
	"int":    &PrimTV{IntType},
	"uint":   &PrimTV{UintType},
	"string": &PrimTV{StringType},
	"type":   &PrimTV{TypeType},
	"error":  ErrorTV,
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
		if o.Word == "*" {
			o.Next()
			elemX := o.ParseExpr().(TypeValue)
			return &PointerTV{elemX}
		}
		if o.Word == "[" {
			o.Next()
			if o.Word != "]" {
				Panicf("for slice type, after [ expected ], got %v", o.Word)
			}
			o.Next()
			elemX := o.ParseExpr().(TypeValue)
			return &SliceTV{elemX}
		}
		if o.Word == "&" {
			o.Next()
			handleClass := o.TakeIdent()
			return o.ParseConstructor(handleClass)
		}
	}
	panic("bad ParsePrim")
}

func (o *Parser) ParseConstructor(handleClass string) Expr {
	o.TakePunc("{")
	ctor := &ConstructorX{
		Name: handleClass,
	}
LOOP:
	for {
		switch o.Kind {
		case L_Ident:
			fieldName := o.TakeIdent()
			fieldType := o.ParseType()
			ctor.Fields = append(ctor.Fields, NameAndType{fieldName, fieldType})
		case L_EOL:
			o.Next()
		case L_Punc:
			if o.Word == "," {
				o.Next()
				continue LOOP
			}
			if o.Word == "}" {
				break LOOP
			}
			panic(Format("Expected identifier or `}` but got %q", o.Word))
		default:
			panic(Format("Expected identifier or `}` but got %q", o.Word))
		}
	}
	o.TakePunc("}")
	return ctor
}

func (o *Parser) ParsePrimEtc() Expr {
	a := o.ParsePrim()
LOOP:
	for {
		switch o.Word {
		case "(":
			o.TakePunc("(")
			var args []Expr
			if o.Word != ")" {
				args = o.ParseList()
			}
			o.TakePunc(")")
			a = &CallX{a, args}
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

func (o *Parser) ParseType() TypeValue {
	x := o.ParseExpr()
	if t, ok := x.(TypeValue); ok {
		return t
	}
	panic(Format("Expected type expression; got %v", x))
}

func (o *Parser) ParseStructType(name string) *StructRec {
	// name := o.TakeIdent()
	o.TakePunc("{")
	def := &StructRec{
		Name: name,
	}
LOOP:
	for {
		switch o.Kind {
		case L_Ident:
			fieldName := o.TakeIdent()
			fieldType := o.ParseType()
			def.Fields = append(def.Fields, NameAndType{fieldName, fieldType})
		case L_EOL:
			o.Next()
		case L_Punc:
			if o.Word == "}" {
				break LOOP
			}
			panic(Format("Expected identifier or `}` but got %q", o.Word))
		default:
			panic(Format("Expected identifier or `}` but got %q", o.Word))
		}
	}
	o.TakePunc("}")
	return def
}
func (o *Parser) ParseInterfaceType(name string) *InterfaceRec {
	// name := o.TakeIdent()
	o.TakePunc("{")
	def := &InterfaceRec{
		Name: name,
	}
LOOP:
	for {
		switch o.Kind {
		case L_Ident:
			fieldName := o.TakeIdent()
			sig := &FunctionRec{}
			o.ParseFunctionSignature(sig)
			fieldType := &FunctionTV{sig}
			def.Fields = append(def.Fields, NameAndType{fieldName, fieldType})
		case L_EOL:
			o.Next()
		case L_Punc:
			if o.Word == "}" {
				break LOOP
			}
			panic(Format("Expected identifier or `}` but got %q", o.Word))
		default:
			panic(Format("Expected identifier or `}` but got %q", o.Word))
		}
	}
	o.TakePunc("}")
	return def
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
	} else if o.Kind == L_EOL {
		// Result not assigned.
		return &AssignS{nil, "", a}
	} else {
		panic(Format("Unexpected token after statement: %v", o.Word))
	}
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
		yes := o.ParseBlock(b.FunctionRec)
		var no *Block
		if o.Word == "else" {
			no = o.ParseBlock(b.FunctionRec)
		}
		return &IfS{pred, yes, no}
	case "for":
		o.Next()
		var pred Expr
		if o.Word != "{" {
			pred = o.ParseExpr()
		}
		b2 := o.ParseBlock(b.FunctionRec)
		return &WhileS{pred, b2}
	case "switch":
		o.Next()
		var subject Expr
		if o.Word != "{" {
			subject = o.ParseExpr()
		}
		o.TakePunc("{")
		sws := &SwitchS{subject, nil, nil}
		for o.Word != "}" {
			for o.Word == ";;" {
				o.Next()
			}
			cOrD := o.TakeIdent()
			switch cOrD {
			case "case":
				matches := o.ParseList()
				o.TakePunc(":")
				bare := o.ParseBareBlock(b.FunctionRec)
				sws.Cases = append(sws.Cases, &Case{matches, bare})
			case "default":
				o.TakePunc(":")
				bare := o.ParseBareBlock(b.FunctionRec)
				sws.Default = bare
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
func (o *Parser) ParseBlock(fn *FunctionRec) *Block {
	o.TakePunc("{")
	b := o.ParseBareBlock(fn)
	o.TakePunc("}")
	return b
}
func (o *Parser) ParseBareBlock(fn *FunctionRec) *Block {
	b := &Block{FunctionRec: fn}
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

func (o *Parser) ParseFunctionSignature(fn *FunctionRec) {
	o.TakePunc("(")
	for o.Word != ")" {
		s := o.TakeIdent()
		if o.Word == "." {
			o.TakePunc(".")
			o.TakePunc(".")
			o.TakePunc(".")
			fn.IsEllipsis = true
		}
		t := o.ParseType()
		fn.Ins = append(fn.Ins, NameAndType{s, t})
		if o.Word == "," {
			o.TakePunc(",")
		} else if o.Word != ")" {
			Panicf("expected `,` or `)` but got %q", o.Word)
		}
		if fn.IsEllipsis {
			if o.Word != ")" {
				panic(Format("Expected `)` after ellipsis arg, but got `%v`", o.Word))
			}
			numIns := len(fn.Ins)
			fn.Ins[numIns-1].TV = &SliceTV{fn.Ins[numIns-1].TV}
		}
	}
	o.TakePunc(")")
	if o.Word != "{" && o.Kind != L_EOL {
		if o.Word == "(" {
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
}
func (o *Parser) ParseFunc(df *DefFunc) {
	fn := df.FunctionRec
	o.ParseFunctionSignature(fn)
	if o.Kind != L_EOL {
		fn.Body = o.ParseBlock(fn)
	}
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
				o.Package = &DefPackage{
					DefCommon: DefCommon{
						Name: w,
					},
				}
			case "import":
				if o.Kind != L_String {
					Panicf("after import, expected string, got %v", o.Word)
				}
				w := o.Word
				o.Next()
				o.Imports[w] = &DefImport{
					DefCommon: DefCommon{
						Name: w,
						C:    w,
						T:    &PrimTV{ImportType},
					},
				}
			case "const":
				w := o.TakeIdent()
				o.TakePunc("=")
				x := o.ParseExpr()
				o.Consts[w] = &DefConst{
					DefCommon: DefCommon{
						Name: w,
						C:    FullName("C", o.Package.Name, w),
					},
					Expr: x,
				}
			case "var":
				w := o.TakeIdent()
				t := o.ParseType()
				o.Vars[w] = &DefVar{
					DefCommon: DefCommon{
						Name: w,
						C:    FullName("V", o.Package.Name, w),
						T:    t,
					},
				}
			case "type":
				w := o.TakeIdent()
				t := o.ParseType()
				o.Types[w] = &DefType{
					DefCommon: DefCommon{
						Name: w,
						C:    FullName("T", o.Package.Name, w),
						T:    &PrimTV{TypeType},
					},
					Expr: t,
					TV:   nil, // TODO
				}
			case "func":
				fn := &FunctionRec{}
				switch o.Kind {
				case L_Punc:
					o.TakePunc("(")
					receiver := o.TakeIdent()
					receiverType := o.ParseType()
					o.TakePunc(")")
					fn.Ins = append(fn.Ins, NameAndType{receiver, receiverType})
					fn.IsMethod = true
				}
				name := o.TakeIdent()
				df := &DefFunc{
					DefCommon: DefCommon{
						Name: name,
						C:    FullName("F", o.Package.Name, name),
						T:    &FunctionTV{fn},
					},
					FunctionRec: fn,
				}
				fn.Def = df
				o.ParseFunc(df)
				o.Funcs[name] = df
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
	TStr() TStr
	ToC() string
}

type LValue interface {
	TStr() TStr
	LToC() string
}

type SimpleValue struct {
	C string // C language expression
	T TStr
	// GlobalDef Def
}

type SimpleLValue struct {
	LC string // C language expression
	T  TStr
	// GlobalDef Def
}

func (val *SimpleValue) ToC() string {
	return val.C
}
func (val *SimpleValue) TStr() TStr {
	return val.T
}
func (lval *SimpleLValue) LToC() string {
	return lval.LC
}
func (lval *SimpleLValue) TStr() TStr {
	return lval.T
}

func (cg *CGen) LoadModule(name string) *CMod {
	log.Printf("LoadModule: << %q", name)
	if already, ok := cg.Mods[name]; ok {
		log.Printf("LoadModule: short return: %v", already)
		return already
	}

	filename := cg.Options.LibDir + "/" + name + ".go"
	r, err := os.Open(filename)
	if err != nil {
		panic(Format("cannot open %q: %v", filename, err))
	}
	defer r.Close()

	cm := &CMod{
		pre:        &cPreMod{},
		W:          cg.W,
		Package:    name,
		GlobalDefs: make(map[string]Def),
		CGen:       cg,
	}
	cm.pre.cm = cm
	cg.Mods[name] = cm

	log.Printf("LoadModule: Parser")
	p := NewParser(r, filename)
	p.ParseTop()
	log.Printf("LoadModule: Visit")
	cm.BigVisit(p)
	log.Printf("LoadModule: Done")
	return cm
}

func CompileToC(opt *Options, r io.Reader, sourceName string, w io.Writer) {
	p := NewParser(r, sourceName)
	p.ParseTop()
	cg := NewCGen(opt, w)
	// BootstrapModules(cg)
	cm := cg.Mods["main"]
	// BootstrapBuiltins(cm)
	cg.LoadModule("builtin")

	cm.P("#include <stdio.h>")
	cm.P("#include \"runt.h\"")
	cm.BigVisit(p)
}

func Sorted(aMap interface{}) []interface{} {
	var z []interface{}

	for _, key := range reflect.ValueOf(aMap).MapKeys() {
		value := reflect.ValueOf(aMap).MapIndex(key)
		z = append(z, value.Interface())
	}

	sort.Slice(z, func(i, j int) bool { // Less Than function
		a := reflect.ValueOf(z[i]).Elem().FieldByName("Name").Interface().(string)
		b := reflect.ValueOf(z[j]).Elem().FieldByName("Name").Interface().(string)
		return a < b
	})

	return z
}
func Sort(defs []Named) {
	sort.Sort(NamedSlice(defs))
}

func (cm *CMod) BigVisit(p *Parser) {
	cm.pre.VisitDefPackage(p.Package)
	for _, i := range Sorted(p.Imports) {
		cm.pre.VisitDefImport(i.(*DefImport))
	}
	for _, c := range Sorted(p.Consts) {
		cm.pre.VisitDefConst(c.(*DefConst))
	}
	for _, t := range Sorted(p.Types) {
		cm.pre.VisitDefType(t.(*DefType))
	}
	for _, v := range Sorted(p.Vars) {
		cm.pre.VisitDefVar(v.(*DefVar))
	}
	for _, f := range Sorted(p.Funcs) {
		cm.pre.VisitDefFunc(f.(*DefFunc))
	}

	cm.VisitDefPackage(p.Package)
	cm.P("// ..... Imports .....")
	for _, i := range Sorted(p.Imports) {
		cm.VisitDefImport(i.(*DefImport))
	}
	cm.P("// ..... Consts .....")
	for _, c := range Sorted(p.Consts) {
		cm.VisitDefConst(c.(*DefConst))
	}
	cm.P("// ..... Types .....")
	for _, t := range Sorted(p.Types) {
		cm.VisitDefType(t.(*DefType))
	}
	cm.P("// ..... Vars .....")
	for _, v := range Sorted(p.Vars) {
		cm.VisitDefVar(v.(*DefVar))
	}
	cm.P("// ..... Funcs .....")
	for _, f := range Sorted(p.Funcs) {
		cm.VisitDefFunc(f.(*DefFunc))
	}
	cm.P("// ..... Done .....")

	cm.Flush()
}

type cPreMod struct {
	cm *CMod
	// cg *CGen
}

func (pre *cPreMod) mustNotExistYet(s string) {
	if _, ok := pre.cm.GlobalDefs[s]; ok {
		Panicf("redefined global name: %s", s)
	}
}
func (pre *cPreMod) VisitDefPackage(def *DefPackage) {
	pre.mustNotExistYet(def.Name)
	pre.cm.Package = def.Name
	pre.cm.P("\n// PRE VISIT %#v\n", def)
}
func (pre *cPreMod) VisitDefImport(def *DefImport) {
	pre.mustNotExistYet(def.Name)
	pre.cm.GlobalDefs[def.Name] = def
	pre.cm.P("\n// PRE VISIT %#v\n", def)
	pre.cm.P("\n// MARCO")
	pre.cm.CGen.LoadModule(def.Name)
	pre.cm.P("\n// POLO")
}
func (pre *cPreMod) VisitDefConst(def *DefConst) {
	pre.mustNotExistYet(def.Name)
	pre.cm.GlobalDefs[def.Name] = def
	pre.cm.P("\n// PRE VISIT %#v\n", def)
}
func (pre *cPreMod) VisitDefVar(def *DefVar) {
	pre.mustNotExistYet(def.Name)
	pre.cm.GlobalDefs[def.Name] = def
	log.Printf("pre visit DefVar: %v => %v", def, pre.cm.GlobalDefs)
	pre.cm.P("\n// PRE VISIT %#v\n", def)
}
func (pre *cPreMod) VisitDefType(def *DefType) {
	pre.mustNotExistYet(def.Name)
	pre.cm.GlobalDefs[def.Name] = def
	pre.cm.P("\n// PRE VISIT %#v\n", def)
}
func (pre *cPreMod) VisitDefFunc(def *DefFunc) {
	fn := def.FunctionRec
	pre.mustNotExistYet(def.Name)
	pre.cm.GlobalDefs[def.Name] = def

	// TODO -- dedup
	var b Buf
	b.P("void %s(", def.C)
	if len(fn.Ins) > 0 {
		firstTime := true
		for _, name_and_type := range fn.Ins {
			if !firstTime {
				b.P(", ")
			}
			b.P("%s %s", TypeNameInC(name_and_type.TV.AsTStr()), "v_"+name_and_type.Name)
			firstTime = false
		}
	}
	b.P(");\n")
	pre.cm.P("\n// {{{{{ // BEGIN Pre def of func: %v\n", def)
	pre.cm.P(b.String())
	pre.cm.P("\n// }}}}} // END Pre def of func: %v\n", def)
}

type CMod struct {
	pre        *cPreMod
	W          *bufio.Writer
	Package    string
	GlobalDefs map[string]Def
	BreakTo    string
	ContinueTo string
	CGen       *CGen
	Structs    *StructRec
	Interfaces *InterfaceRec
	Functions  *FunctionRec
}
type CGen struct {
	Mods    map[string]*CMod
	Options *Options
	W       *bufio.Writer
}

func NewCGen(opt *Options, w io.Writer) *CGen {
	mainMod := &CMod{
		pre:        &cPreMod{},
		W:          bufio.NewWriter(w),
		Package:    "main",
		GlobalDefs: make(map[string]Def),
	}
	mainMod.pre.cm = mainMod
	cg := &CGen{
		Mods:    map[string]*CMod{"main": mainMod},
		W:       mainMod.W,
		Options: opt,
	}
	mainMod.CGen = cg
	return cg
}
func (cm *CMod) P(format string, args ...interface{}) {
	log.Printf("<<<<< %q >>>>> %q", format, fmt.Sprintf(format, args...))
	fmt.Fprintf(cm.W, format+"\n", args...)
}
func (cm *CMod) Flush() {
	cm.W.Flush()
}

func (cm *CMod) VisitLvalIdent(x *IdentX) LValue {
	value := cm.VisitIdent(x)
	return &SimpleLValue{LC: Format("&(%s)", value.ToC()), T: value.TStr()}
}
func (cm *CMod) VisitLValSub(x *SubX) LValue {
	value := cm.VisitSub(x)
	return &SimpleLValue{LC: Format("TODO_LValue(%s)", value.ToC()), T: value.TStr()}
}
func (cm *CMod) VisitLvalDot(x *DotX) LValue {
	value := cm.VisitDot(x)
	return &SimpleLValue{LC: Format("&(%s)", value.ToC()), T: value.TStr()}
}

func (cm *CMod) VisitLitInt(x *LitIntX) Value {
	return &SimpleValue{
		C: Format("%d", x.X),
		T: ConstIntType,
	}
}
func (cm *CMod) VisitLitString(x *LitStringX) Value {
	return &SimpleValue{
		C: Format("%q", x.X),
		T: StringType,
	}
}
func (cm *CMod) VisitIdent(x *IdentX) Value {
	log.Printf("VisitIdent <= %v", x)
	z := cm._VisitIdent_(x)
	log.Printf("VisitIdent => %#v", z)
	return z
}
func (cm *CMod) _VisitIdent_(x *IdentX) Value {
	if gd, ok := cm.GlobalDefs[x.X]; ok {
		return gd
	}
	// Else, assume it is a local variable.
	return &SimpleValue{C: "v_" + x.X, T: IntType}
}
func (cm *CMod) VisitBinOp(x *BinOpX) Value {
	a := x.A.VisitExpr(cm)
	b := x.B.VisitExpr(cm)
	return &SimpleValue{
		C: Format("(%s) %s (%s)", a.ToC(), x.Op, b.ToC()),
		T: IntType,
	}
}
func (cm *CMod) VisitConstructor(x *ConstructorX) Value {
	return &SimpleValue{
		C: Format("(%s) alloc(C_%s)", x.Name, x.Name),
		T: TStr(Format(HandleForm, x.Name)),
	}
}

func Intlike(ty TStr) bool {
	switch ty {
	case ByteType, IntType, UintType, ConstIntType:
		return true
	default:
		return false
	}
}

func CopyAndSoftConvert(in NameAndType, out NameAndType) string {
	switch out.TV.AsTStr()[0] {
	case InterfacePre: // Create an interface.
		handle, pointer := "0", "0"
		switch in.TV.AsTStr()[0] {
		case HandlePre:
			handle = in.Name
		default:
			pointer = Format("(word)(&%s)" + in.Name) // broken for int constants
		}
		return Format("Interface %s = {%s, %s, %q};", out.Name, handle, pointer, in.TV)

	default:
		outCType := TypeNameInC(out.TV.AsTStr())
		if Intlike(in.TV.AsTStr()) && Intlike(out.TV.AsTStr()) {
			return Format("%s %s = (%s)%s;", outCType, out.Name, outCType, in.Name)
		}
		if in.TV == out.TV {
			return Format("%s %s = %s;", outCType, out.Name, in.Name)
		}
	}
	return Format("((( CONVERT TO %#v FROM %#v )));", out, in)
}

func (cm *CMod) VisitCall(x *CallX) Value {
	cm.P("// Calling: %#v", x)
	cm.P("// Calling Func: %#v", x.Func)
	for i, a := range x.Args {
		cm.P("// Calling with Arg [%d]: %#v", i, a)
	}

	ser := Serial("call")
	cm.P("{ // %s", ser)
	log.Printf("x.Func: %#v", x.Func)
	funcX := x.Func.VisitExpr(cm)
	log.Printf("funcX: %#v", funcX)
	funcRec := funcX.(*DefFunc).FunctionRec
	funcname := funcRec.Function.Name
	c2 := ""
	c := Format(" %s( fp", funcname)

	for i, in := range funcRec.Ins {
		val := x.Args[i].VisitExpr(cm)
		expectedType := in.TV

		if expectedType.AsTStr()[0] == '.' {
			memberType := expectedType.AsTStr()[1:]
			sliceType := "S" + memberType
			c2 += Format("Slice %s_in_rest = CreateSlice();", ser)
			for j := i; j < len(x.Args); j++ {
				CopyAndSoftConvert(
					NameAndType{val.ToC(), val},
					NameAndType{Format("%s_in_%d", ser, j), sliceType})
				c2 += Format("AppendSlice(%d_in_rest,  %s_in_%d);", ser, ser, j)
			}
			c += Format("FINISH(%s_in_rest);", ser)

		} else {
			//##if expectedType != val.TStr() {
			//##panic(Format("bad type: expected %s, got %s", expectedType, val.TStr()))
			//##}
			CopyAndSoftConvert(
				NameAndType{val.ToC(), val},
				NameAndType{Format("%s_in_%d", ser, i), expectedType})
			//##cm.P("  %s %s_in_%d = %s;", TypeNameInC(in.TStr), ser, i, val.ToC())
			c += Format(", %s_in_%d", ser, i)
		}

	}
	for i, out := range funcRec.Outs {
		cm.P("  %s %s_out_%d;", TypeNameInC(out.TV.AsTStr()), ser, i)
		c += Format(", &%s_out_%d", ser, i)
	}
	c += " );"
	cm.P("[[[%s]]]  %s\n} // %s", c2, c, ser)

	switch len(funcRec.Outs) {
	case 0:
		return &SimpleValue{"VOID", VoidType}
	case 1:
		return &SimpleValue{Format("%s_out_0", ser), funcRec.Outs[0].TV}
	default:
		return &SimpleValue{ser, ListType}
	}
	/*
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
		return &SimpleValue{
			C: ccall,
			T: "dunno",
		}
	*/
}

/*
func (cm *CMod) VisitType(x *TypeX) Value {
	return &SimpleValue{
		C: string(x.T),
		T: TypeType,
	}
}
*/
func (cm *CMod) VisitSub(x *SubX) Value {
	return &SimpleValue{
		C: Format("SubXXX(%v)", x),
		T: "",
	}
}
func (cm *CMod) VisitDot(dot *DotX) Value {
	log.Printf("VisitDot: <------ %#v", dot)
	val := dot.X.VisitExpr(cm)
	log.Printf("VisitDot: val---- %#v", val)
	if val.TStr() == ImportType {
		modName := val.ToC() // is there a better way?
		println("DOT", modName, dot.Member)
		otherMod := cm.CGen.Mods[modName] // TODO: import aliases.
		println("OM", otherMod)
		println("GD", otherMod.GlobalDefs)
		_, ok := otherMod.GlobalDefs[dot.Member]
		if !ok {
			panic(Format("cannot find member %s in module %s", dot.Member, modName))
		}
		return otherMod.VisitIdent(&IdentX{modName})
	}

	z := &SimpleValue{
		C: Format("DotXXX(%v)", dot),
		T: "i",
	}
	log.Printf("VisitDot: Not Import: ----> %v", z)
	return z
}
func (cm *CMod) VisitAssign(ass *AssignS) {
	cm.P("//## assign..... %v   %v   %v", ass.A, ass.Op, ass.B)
	lenA, lenB := len(ass.A), len(ass.B)
	_ = lenA

	// Evalute the rvalues.
	var rvalues []Value
	for _, e := range ass.B {
		rvalues = append(rvalues, e.VisitExpr(cm))
	}

	// If there is just one thing on right, and it is a CallX, set bcall.
	var bcall *CallX
	if len(ass.B) == 1 {
		switch t := ass.B[0].(type) {
		case *CallX:
			bcall = t
		}
	}

	switch {
	case ass.B == nil:
		// An lvalue followed by ++ or --.
		if len(ass.A) != 1 {
			Panicf("operator %v requires one lvalue on the left, got %v", ass.Op, ass.A)
		}
		// TODO check lvalue
		cvar := ass.A[0].VisitExpr(cm).ToC()
		cm.P("  (%s)%s;", cvar, ass.Op)

	case ass.A == nil && bcall == nil:
		// No assignment.  Just a non-function.  Does this happen?
		panic(Format("Lone expr is not a funciton call: [%v]", ass.B))

	case ass.A == nil && bcall != nil:
		// No assignment.  Just a function call.
		log.Printf("bcall=%#v", bcall)
		visited := bcall.VisitExpr(cm)
		log.Printf("visited=%#v", visited)

		funcRec := visited.(*DefFunc).FunctionRec

		funcname := funcRec.Function.Name
		log.Printf("funcname=%s", funcname)

		// functype := fn.TStr()
		if lenB != len(bcall.Args) {
			panic(Format("Function %s wants %d args, got %d", funcname, len(bcall.Args), lenB))
		}
		ser := Serial("call")
		cm.P("{ // %s", ser)
		c := Format(" %s( fp", funcname)
		for i, in := range funcRec.Ins {
			val := ass.B[i].VisitExpr(cm)
			expectedType := in.TV
			if expectedType != val.TStr() {
				panic(Format("bad type: expected %s, got %s", expectedType, val.TStr()))
			}
			cm.P("  %s %s_in_%d = %s;", TypeNameInC(in.TV.AsTStr()), ser, i, val.ToC())
			c += Format(", %s_in_%d", ser, i)
		}
		for i, out := range funcRec.Outs {
			cm.P("  %s %s_out_%d;", TypeNameInC(out.TV.AsTStr()), ser, i)
			c += Format(", &%s_out_%d", ser, i)
		}
		c += " );"
		cm.P("  %s\n} // %s", c, ser)
	case len(ass.A) > 1 && bcall != nil:
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
	default:
		if len(ass.A) != len(ass.B) {
			Panicf("wrong number of values in assign: left has %d, right has %d", len(ass.A), len(ass.B))
		}
		for i, val := range rvalues {
			lhs := ass.A[i]
			switch t := lhs.(type) {
			case *IdentX:
				// TODO -- check that variable t.X has the right type.
				switch ass.Op {
				case "=":
					// TODO check Globals
					cvar := "v_" + t.X
					cm.P("  %s = (%s)(%s);", cvar, TypeNameInC(val.TStr()), val.ToC())
				case ":=":
					// TODO check Globals
					cvar := Format("%s %s", TypeNameInC(val.TStr()), "v_"+t.X)
					cm.P("  %s = (%s)(%s);", cvar, TypeNameInC(val.TStr()), val.ToC())
				}
			default:
				log.Fatal("bad VisitAssign LHS: %#v", ass.A)
			}
		}
	} // switch
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
	cm.P("  { t_int _switch_ = %s;", sws.Switch.VisitExpr(cm).ToC())
	for _, c := range sws.Cases {
		cm.P("  if (")
		for _, m := range c.Matches {
			cm.P("_switch_ == %s ||", m.VisitExpr(cm).ToC())
		}
		cm.P("      0 ) {")
		c.Body.VisitStmt(cm)
		cm.P("  } else ")
	}
	cm.P("  {")
	if sws.Default != nil {
		sws.Default.VisitStmt(cm)
	}
	cm.P("  }")
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
	cm.P("%s V_%s__%s; // global var", TypeNameInC(def.T), cm.Package, def.Name)
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
	fn := def.FunctionRec
	log.Printf("// func %s: %#v", def.Name, def)
	var b Buf
	cfunc := Format("F_%s__%s", cm.Package, def.Name)
	b.P("void %s(", cfunc)
	if len(fn.Ins) > 0 {
		firstTime := true
		for _, name_and_type := range fn.Ins {
			if !firstTime {
				b.P(", ")
			}
			b.P("%s %s", TypeNameInC(name_and_type.TV), "v_"+name_and_type.Name)
			firstTime = false
		}
	}
	b.P(") {\n")
	cm.P(b.String())
	fn.Body.VisitStmt(cm)
	cm.P("}\n")
}

var SerialNum uint

func Serial(prefix string) string {
	SerialNum++
	return Format("%s_%d", prefix, SerialNum)
}
