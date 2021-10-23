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

// #################################################

func assert(b bool) {
	if !b {
		panic("assert fails")
	}
}

func Panicf(format string, args ...interface{}) string {
	s := Format(format, args...)
	panic(s)
}

///////////

type Options struct {
	LibDir string
}

//////// Expr

type ExprVisitor interface {
	VisitLitInt(*LitIntX) Value
	VisitLitString(*LitStringX) Value
	VisitIdent(*IdentX) Value
	VisitBinOp(*BinOpX) Value
	VisitCall(*CallX) Value
	VisitSub(*SubX) Value
	VisitDot(*DotX) Value
	VisitConstructor(*ConstructorX) Value
}
type LvalVisitor interface {
	VisitLvalIdent(*IdentX) LValue
	VisitLvalSub(*SubX) LValue
	VisitLvalDot(*DotX) LValue
}

type Expr interface {
	String() string
	VisitExpr(ExprVisitor) Value
}

type Lval interface {
	String() string
	VisitLVal(LvalVisitor) Value
}

//

type WrapperTypeX struct {
	TV TypeValue
}

func (o *WrapperTypeX) VisitExpr(v ExprVisitor) Value {
	return o.TV
}
func (o *WrapperTypeX) String() string {
	return Format("Wrap{%v}", o.TV)
}

//

type TypeValue interface {
	Expr
	Value
	Intlike() bool
	CType() string
	HandleType() (z string, ok bool)
	Assign(c string, typ TypeValue) (z string, ok bool)
	Cast(c string, typ TypeValue) (z string, ok bool)
	Equal(typ TypeValue) bool
}

type BaseTV struct {
	Name string
}

type PrimTV struct {
	BaseTV
}
type TypeTV struct {
	BaseTV
}
type PointerTV struct {
	BaseTV
	E TypeValue
}
type SliceTV struct {
	BaseTV
	E TypeValue
}
type DotDotDotSliceTV struct {
	BaseTV
	E TypeValue
}
type MapTV struct {
	BaseTV
	K TypeValue
	V TypeValue
}
type IndirectTV struct {
	BaseTV
	Name string
}
type StructTV struct {
	BaseTV
	StructRec *StructRec
}
type InterfaceTV struct {
	BaseTV
	InterfaceRec *InterfaceRec
}
type FunctionTV struct {
	BaseTV
	FunctionRec *FunctionRec
}

func (tv *PrimTV) Type() TypeValue      { return &TypeTV{} }
func (tv *TypeTV) Type() TypeValue      { return &TypeTV{} }
func (tv *PointerTV) Type() TypeValue   { return &TypeTV{} }
func (tv *SliceTV) Type() TypeValue     { return &TypeTV{} }
func (tv *MapTV) Type() TypeValue       { return &TypeTV{} }
func (tv *IndirectTV) Type() TypeValue  { return &TypeTV{} }
func (tv *StructTV) Type() TypeValue    { return &TypeTV{} }
func (tv *InterfaceTV) Type() TypeValue { return &TypeTV{} }
func (tv *FunctionTV) Type() TypeValue  { return &TypeTV{} }

func (tv *BaseTV) ToC() string { panic("cant") }

func (o *BaseTV) Intlike() bool { return false }
func (o *PrimTV) Intlike() bool {
	switch o.Name {
	case "byte", "int", "uint":
		return true
	}
	return false
}

func (o *BaseTV) Equal(typ TypeValue) bool { panic("todo") }
func (o *PrimTV) Equal(typ TypeValue) bool {
	switch t := typ.(type) {
	case *PrimTV:
		return o.Name == t.Name
	}
	return false
}
func (o *SliceTV) Equal(typ TypeValue) bool {
	switch t := typ.(type) {
	case *SliceTV:
		return o.E.Equal(t.E)
	}
	return false
}
func (o *PointerTV) Equal(typ TypeValue) bool {
	switch t := typ.(type) {
	case *PointerTV:
		return o.E.Equal(t.E)
	}
	return false
}
func (o *MapTV) Equal(typ TypeValue) bool {
	switch t := typ.(type) {
	case *MapTV:
		return o.K.Equal(t.K) && o.V.Equal(t.V)
	}
	return false
}
func (o *StructTV) Equal(typ TypeValue) bool {
	switch t := typ.(type) {
	case *StructTV:
		return o.Name == t.Name
	}
	return false
}

func (o *BaseTV) HandleType() (z string, ok bool) { return "", false }
func (o *PointerTV) HandleType() (z string, ok bool) {
	if st, ok := o.E.(*StructTV); ok {
		return st.Name, true
	}
	return "", false
}

func (o *BaseTV) CType() string      { return "TODO" }
func (o *PrimTV) CType() string      { return PrimTypeCMap[o.Name] }
func (o *SliceTV) CType() string     { return "Slice" }
func (o *MapTV) CType() string       { return "Map" }
func (o *StructTV) CType() string    { return "Struct" }
func (o *PointerTV) CType() string   { return "Pointer" }
func (o *InterfaceTV) CType() string { return "Interface" }

func (o *BaseTV) Assign(c string, typ TypeValue) (z string, ok bool) { panic("todo") }
func (o *PrimTV) Assign(c string, typ TypeValue) (z string, ok bool) {
	if o.Equal(typ) {
		return c, true
	} else {
		return "", false
	}
}
func (o *SliceTV) Assign(c string, typ TypeValue) (z string, ok bool) {
	if o.Equal(typ) {
		return c, true
	} else {
		return "", false
	}
}
func (o *MapTV) Assign(c string, typ TypeValue) (z string, ok bool) {
	if o.Equal(typ) {
		return c, true
	} else {
		return "", false
	}
}
func (o *StructTV) Assign(c string, typ TypeValue) (z string, ok bool) {
	if o.Equal(typ) {
		return c, true
	} else {
		return "", false
	}
}
func (o *PointerTV) Assign(c string, typ TypeValue) (z string, ok bool) {
	if o.Equal(typ) {
		return c, true
	} else {
		return "", false
	}
}
func (o *InterfaceTV) Assign(c string, typ TypeValue) (z string, ok bool) {
	if _, ok := typ.HandleType(); ok {
		// TODO: check compat
		return Format("HandleToInterface(%s)", c), true
	}
	if _, ok := typ.(*InterfaceTV); ok {
		return c, true
	}
	return "", false
}

func (o *BaseTV) Cast(c string, typ TypeValue) (z string, ok bool) { panic("todo") }
func (o *SliceTV) Cast(c string, typ TypeValue) (z string, ok bool) {
	if o.Equal(typ) {
		return c, true
	} else {
		return "", false
	}
}
func (o *MapTV) Cast(c string, typ TypeValue) (z string, ok bool) {
	if o.Equal(typ) {
		return c, true
	} else {
		return "", false
	}
}
func (o *StructTV) Cast(c string, typ TypeValue) (z string, ok bool) {
	if o.Equal(typ) {
		return c, true
	} else {
		return "", false
	}
}
func (o *PointerTV) Cast(c string, typ TypeValue) (z string, ok bool) {
	if o.Equal(typ) {
		return c, true
	} else {
		return "", false
	}
}
func (o *InterfaceTV) Cast(c string, typ TypeValue) (z string, ok bool) {
	if o.Equal(typ) {
		return c, true
	} else {
		return "", false
	}
}

func (o *PrimTV) Cast(c string, typ TypeValue) (z string, ok bool) {
	if o.Equal(typ) {
		return c, true
	}
	if o.Intlike() && typ.Intlike() {
		return Format("(%s)(%s)", PrimTypeCMap[o.Name], c), true
	}
	return "", false
}

func (tv *PrimTV) String() string      { return Format("%T(%#v)", *tv, *tv) }
func (tv *TypeTV) String() string      { return Format("%T()", *tv) }
func (tv *PointerTV) String() string   { return Format("%T(%#v)", *tv, *tv) }
func (tv *SliceTV) String() string     { return Format("%T(%#v)", *tv, *tv) }
func (tv *MapTV) String() string       { return Format("%T(%#v)", *tv, *tv) }
func (tv *IndirectTV) String() string  { return Format("%T(%#v)", *tv, *tv) }
func (tv *StructTV) String() string    { return Format("%T(%#v)", *tv, *tv) }
func (tv *InterfaceTV) String() string { return Format("%T(%#v)", *tv, *tv) }
func (tv *FunctionTV) String() string  { return Format("%T(%#v)", *tv, *tv) }

func (tv *PrimTV) VisitExpr(v ExprVisitor) Value      { return tv }
func (tv *TypeTV) VisitExpr(v ExprVisitor) Value      { return tv }
func (tv *PointerTV) VisitExpr(v ExprVisitor) Value   { return tv }
func (tv *SliceTV) VisitExpr(v ExprVisitor) Value     { return tv }
func (tv *MapTV) VisitExpr(v ExprVisitor) Value       { return tv }
func (tv *IndirectTV) VisitExpr(v ExprVisitor) Value  { return tv }
func (tv *StructTV) VisitExpr(v ExprVisitor) Value    { return tv }
func (tv *InterfaceTV) VisitExpr(v ExprVisitor) Value { return tv }
func (tv *FunctionTV) VisitExpr(v ExprVisitor) Value  { return tv }

type LitIntX struct {
	X int
}

func (o *LitIntX) String() string {
	return fmt.Sprintf("Int(%d)", o.X)
}
func (o *LitIntX) VisitExpr(v ExprVisitor) Value {
	return v.VisitLitInt(o)
}

type LitStringX struct {
	X string
}

func (o *LitStringX) String() string {
	return fmt.Sprintf("String(%q)", o.X)
}
func (o *LitStringX) VisitExpr(v ExprVisitor) Value {
	return v.VisitLitString(o)
}

type IdentX struct {
	X string
}

func (o *IdentX) String() string {
	return fmt.Sprintf("Ident(%s)", o.X)
}
func (o *IdentX) VisitExpr(v ExprVisitor) Value {
	return v.VisitIdent(o)
}

func (o *IdentX) VisitLval(v LvalVisitor) LValue {
	return v.VisitLvalIdent(o)
}

type BinOpX struct {
	A  Expr
	Op string
	B  Expr
}

func (o *BinOpX) String() string {
	return fmt.Sprintf("Bin(%v %q %v)", o.A, o.Op, o.B)
}
func (o *BinOpX) VisitExpr(v ExprVisitor) Value {
	return v.VisitBinOp(o)
}

type ConstructorX struct {
	Name   string
	Fields []NameAndType
}

func (o *ConstructorX) String() string {
	return fmt.Sprintf("Ctor(%q [[[%v]]])", o.Name, o.Fields)
}
func (o *ConstructorX) VisitExpr(v ExprVisitor) Value {
	return v.VisitConstructor(o)
}

type FunctionX struct {
	FunctionRec *FunctionRec
}
type CallX struct {
	Func Expr
	Args []Expr
}

func (o *CallX) String() string {
	return fmt.Sprintf("Call(%s; %s)", o.Func, o.Args)
}
func (o *CallX) VisitExpr(v ExprVisitor) Value {
	return v.VisitCall(o)
}

type DotX struct {
	X      Expr
	Member string
}

func (o *DotX) String() string {
	return fmt.Sprintf("Dot(%s; %s)", o.X, o.Member)
}
func (o *DotX) VisitExpr(v ExprVisitor) Value {
	return v.VisitDot(o)
}
func (o *DotX) VisitLval(v LvalVisitor) LValue {
	return v.VisitLvalDot(o)
}

type SubX struct {
	X         Expr
	Subscript Expr
}

func (o *SubX) String() string {
	return fmt.Sprintf("Sub(%s; %s)", o.X, o.Subscript)
}
func (o *SubX) VisitExpr(v ExprVisitor) Value {
	return v.VisitSub(o)
}

func (o *SubX) VisitLval(v LvalVisitor) LValue {
	return v.VisitLvalSub(o)
}

/////////// Stmt

type StmtVisitor interface {
	VisitAssign(*AssignS)
	VisitWhile(*WhileS)
	VisitSwitch(*SwitchS)
	VisitIf(*IfS)
	VisitReturn(*ReturnS)
	VisitBlock(*Block)
	VisitBreak(*BreakS)
	VisitContinue(*ContinueS)
}

type Stmt interface {
	String() string
	VisitStmt(StmtVisitor)
}

type AssignS struct {
	A  []Expr
	Op string
	B  []Expr
}

func (o *AssignS) String() string {
	return fmt.Sprintf("\nAssign(%v <%q> %v)\n", o.A, o.Op, o.B)
}

func (o *AssignS) VisitStmt(v StmtVisitor) {
	v.VisitAssign(o)
}

type ReturnS struct {
	X []Expr
}
type BreakS struct {
	Label string
}
type ContinueS struct {
	Label string
}

func (o *ReturnS) String() string {
	return fmt.Sprintf("\nReturn(%v)\n", o.X)
}

func (o *ReturnS) VisitStmt(v StmtVisitor) {
	v.VisitReturn(o)
}

func (o *BreakS) String() string {
	return fmt.Sprintf("\nBreak(%v)\n", o.Label)
}

func (o *BreakS) VisitStmt(v StmtVisitor) {
	v.VisitBreak(o)
}

func (o *ContinueS) String() string {
	return fmt.Sprintf("\nContinue(%v)\n", o.Label)
}

func (o *ContinueS) VisitStmt(v StmtVisitor) {
	v.VisitContinue(o)
}

type Case struct {
	Matches []Expr
	Body    *Block
}
type SwitchS struct {
	Switch  Expr
	Cases   []*Case
	Default *Block
}

func (o *SwitchS) String() string {
	return fmt.Sprintf("\nSwitch(switch: %v, cases: [[[ %#v ]]], default: %v )\n", o.Switch, o.Cases, o.Default)
}

func (o *SwitchS) VisitStmt(v StmtVisitor) {
	v.VisitSwitch(o)
}

type WhileS struct {
	Pred Expr
	Body *Block
}

func (o *WhileS) String() string {
	return fmt.Sprintf("\nWhile(%v)\n", o.Pred)
}

func (o *WhileS) VisitStmt(v StmtVisitor) {
	v.VisitWhile(o)
}

type IfS struct {
	Pred Expr
	Yes  *Block
	No   *Block
}

func (o *IfS) String() string {
	return fmt.Sprintf("\nIf(%v)\n", o.Pred)
}

func (o *IfS) VisitStmt(v StmtVisitor) {
	v.VisitIf(o)
}

////////////////////////

type Named interface {
	GetName() string
}
type NamedSlice []Named

func (ns NamedSlice) Len() int {
	return len(ns)
}
func (ns NamedSlice) Swap(a, b int) {
	ns[a], ns[b] = ns[b], ns[a]
}
func (ns NamedSlice) Less(a, b int) bool {
	return ns[a].GetName() < ns[b].GetName()
}

type Def interface {
	VisitDef(DefVisitor)
	Value
	LValue
}
type DefVisitor interface {
	VisitDefPackage(*DefPackage)
	VisitDefImport(*DefImport)
	VisitDefConst(*DefConst)
	VisitDefVar(*DefVar)
	VisitDefType(*DefType)
	VisitDefFunc(*DefFunc)
}

func (d *DefCommon) ToC() string      { return d.C }
func (d *DefCommon) Type() TypeValue  { return d.T }
func (d *DefCommon) LToC() string     { return "" }
func (d *DefCommon) LType() TypeValue { return d.Type() }
func (d *DefCommon) Named() string    { return d.Name }

func (d *DefVar) LToC() string     { return Format("&%s", d.ToC()) }
func (d *DefVar) LType() TypeValue { return d.T }

type DefCommon struct {
	CMod *CMod
	Name string
	C    string
	T    TypeValue
}

type DefPackage struct {
	DefCommon
}
type DefImport struct {
	DefCommon
	Path string
}
type DefConst struct {
	DefCommon
	Expr Expr
}
type DefVar struct {
	DefCommon
}
type DefType struct {
	DefCommon
	Expr Expr
	TV   TypeValue
}
type DefFunc struct {
	DefCommon
	FunctionRec *FunctionRec
}

// A callable view of a node in a parse tree,
// e.g. global func, lambda, bound method,
// ... any expr of Func kind.
type FunctionRec struct {
	Def          *DefFunc // nil, if not a global func.
	Function     NameAndType
	IsMethod     bool
	Ins          []NameAndType
	Outs         []NameAndType
	HasDotDotDot bool
	Body         *Block
}

type StructRec struct {
	Name   string
	Fields []NameAndType
}

type InterfaceRec struct {
	Name   string
	Fields []NameAndType
}

type NameAndType struct {
	Name string
	TV   TypeValue
}
type Block struct {
	Locals      []NameAndType
	Stmts       []Stmt
	Parent      *Block
	FunctionRec *FunctionRec
}

func (o *Block) VisitStmt(v StmtVisitor) {
	v.VisitBlock(o)
}

func (o *DefPackage) VisitDef(v DefVisitor) {
	v.VisitDefPackage(o)
}
func (o *DefImport) VisitDef(v DefVisitor) {
	v.VisitDefImport(o)
}
func (o *DefConst) VisitDef(v DefVisitor) {
	v.VisitDefConst(o)
}
func (o *DefVar) VisitDef(v DefVisitor) {
	v.VisitDefVar(o)
}
func (o *DefType) VisitDef(v DefVisitor) {
	v.VisitDefType(o)
}
func (o *DefFunc) VisitDef(v DefVisitor) {
	v.VisitDefFunc(o)
}

/*
const XX_BoolType = "a"
const XX_ByteType = "b"
const XX_UintType = "u"
const XX_IntType = "i"
const XX_ConstIntType = "c"
const XX_StringType = "s"
const XX_TypeType = "t"
const XX_ImportType = "@"
const XX_VoidType = "v"
const XX_ListType = "l"

const XX_BoolPre = 'a'
const XX_BytePre = 'b'
const XX_UintPre = 'u'
const XX_IntPre = 'i'
const XX_ConstIntPre = 'c'
const XX_StringPre = 's'
const XX_TypePre = 't'
const XX_ImportPre = '@'
const XX_VoidPre = 'v'
const XX_ListPre = 'l'

const XX_SlicePre = 'S'
const XX_DotDotDotSlicePre = 'E'
const XX_MapPre = 'M'
const XX_ChanPre = 'C'
const XX_FuncPre = 'F'
//const HandlePre = 'H'
const XX_StructPre = 'R'
const XX_InterfacePre = 'I'
const XX_PointerPre = 'P'

const XX_SliceForm = "S:%s"
const XX_DotDotDotSliceForm = "E:%s"
const XX_MapForm = "M:%s:%s"
const XX_ChanForm = "C:%s"
const XX_TypeForm = "t(%s)"
const XX_FuncForm = "F(%s;%s)"
const XX_StructForm = "R{%s}"
//const HandleForm = "H{%s}"
const XX_InterfaceForm = "I{%s}"
const XX_PointerForm = "P{%s}"
*/

// #################################################

// GlobalName creates the global name with kind and package.
func GlobalName(pkg string, name string) string {
	return Format("%s__%s", pkg, name)
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

var DotDotDotAnyTO = &DotDotDotSliceTV{BaseTV{"?"}, InterfaceAnyTO}

var InterfaceAnyTO = &InterfaceTV{
	InterfaceRec: &InterfaceRec{
		Name:   "interface{}",
		Fields: []NameAndType{},
	},
}
var ErrorTO = &InterfaceTV{
	InterfaceRec: &InterfaceRec{
		Name:   "error",
		Fields: []NameAndType{
			// TODO: Error func() string
		},
	},
}
var BoolTO = &PrimTV{BaseTV{"bool"}}
var ByteTO = &PrimTV{BaseTV{"byte"}}
var ConstIntTO = &PrimTV{BaseTV{"_const_int_"}}
var IntTO = &PrimTV{BaseTV{"int"}}
var UintTO = &PrimTV{BaseTV{"uint"}}
var StringTO = &PrimTV{BaseTV{"string"}}
var TypeTO = &PrimTV{BaseTV{"_type_"}}
var ImportTO = &PrimTV{BaseTV{"_import_"}}
var ListTO = &PrimTV{BaseTV{"_list_"}}
var VoidTO = &PrimTV{BaseTV{"_void_"}}

var PrimTypeObjMap = map[string]TypeValue{
	"bool":        BoolTO,
	"byte":        ByteTO,
	"_const_int_": ConstIntTO,
	"int":         IntTO,
	"uint":        UintTO,
	"string":      StringTO,
	"error":       ErrorTO,
	"_type_":      TypeTO,
	"_import_":    ImportTO,
	"_list_":      ListTO,
	"_void_":      VoidTO,
}
var PrimTypeCMap = map[string]string{
	"bool":   "bool",
	"byte":   "byte",
	"int":    "int",
	"uint":   "word",
	"string": "String",
	"error":  "Interface",
	"type":   "const char*",
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
		if o.Word == "interface" {
			o.Next()
			return &InterfaceTV{BaseTV{}, o.ParseInterfaceType("")}
		}
		if o.Word == "struct" {
			o.Next()
			return &StructTV{BaseTV{}, o.ParseStructType("")}
		}
		z := &IdentX{o.Word}
		o.Next()
		return z
	}
	if o.Kind == L_Punc {
		if o.Word == "*" {
			o.Next()
			elemX := o.ParseType()
			return &PointerTV{BaseTV{}, elemX}
		}
		if o.Word == "[" {
			o.Next()
			if o.Word != "]" {
				Panicf("for slice type, after [ expected ], got %v", o.Word)
			}
			o.Next()
			elemX := o.ParseType()
			return &SliceTV{BaseTV{}, elemX}
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
	switch t := x.(type) {
	case TypeValue:
		return t
	case *IdentX:
		if typeObj, ok := PrimTypeObjMap[t.X]; ok {
			return typeObj
		}
		return &IndirectTV{BaseTV{}, t.X}
	case *DotX:
		switch tx := t.X.(type) {
		case *IdentX:
			return &IndirectTV{BaseTV{}, GlobalName(tx.X, t.Member)}
		}
	}
	panic(Format("Expected type expression; got %v", x))
}

func (o *Parser) ParseStructType(name string) *StructRec {
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
			fieldType := &FunctionTV{BaseTV{fieldName}, sig}
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
		if o.Word == ".." {
			o.TakePunc("..")
			o.TakePunc(".")
			fn.HasDotDotDot = true
		}
		t := o.ParseType()
		fn.Ins = append(fn.Ins, NameAndType{s, t})
		if o.Word == "," {
			o.TakePunc(",")
		} else if o.Word != ")" {
			Panicf("expected `,` or `)` but got %q", o.Word)
		}
		if fn.HasDotDotDot {
			if o.Word != ")" {
				panic(Format("Expected `)` after ellipsis arg, but got `%v`", o.Word))
			}
			numIns := len(fn.Ins)
			fn.Ins[numIns-1].TV = &SliceTV{BaseTV{}, fn.Ins[numIns-1].TV}
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

func (o *Parser) ParseTop(cm *CMod, cg *CGen) {
LOOP:
	for {
		switch o.Kind {
		case L_Ident:
			d := o.TakeIdent()
			switch d {
			case "package":
				w := o.TakeIdent()
				if w != cm.Package {
					panic(Format("Expected package %s, got %s", cm.Package, w))
				}
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
						T:    ImportTO,
					},
				}
			case "const":
				w := o.TakeIdent()
				o.TakePunc("=")
				x := o.ParseExpr()
				o.Consts[w] = &DefConst{
					DefCommon: DefCommon{
						Name: w,
						C:    GlobalName(o.Package.Name, w),
					},
					Expr: x,
				}
			case "var":
				w := o.TakeIdent()
				t := o.ParseType()
				o.Vars[w] = &DefVar{
					DefCommon: DefCommon{
						Name: w,
						C:    GlobalName(o.Package.Name, w),
						T:    t,
					},
				}
			case "type":
				w := o.TakeIdent()
				var tv TypeValue
				if o.Word == "interface" {
					o.Next()
					tv = &InterfaceTV{BaseTV{w}, o.ParseInterfaceType(w)}
				} else if o.Word == "struct" {
					o.Next()
					tv = &StructTV{BaseTV{w}, o.ParseStructType(w)}
				} else {
					tv = o.ParseType()
				}
				o.Types[w] = &DefType{
					DefCommon: DefCommon{
						Name: w,
						C:    GlobalName(o.Package.Name, w),
						T:    TypeTO,
					},
					Expr: tv,
					TV:   tv,
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
						C:    GlobalName(o.Package.Name, name),
						T:    &FunctionTV{BaseTV{name}, fn},
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
	Type() TypeValue
	ToC() string
}

type LValue interface {
	LToC() string
}

type SimpleValue struct {
	C string // C language expression
	T TypeValue
	// GlobalDef Def
}

type SimpleLValue struct {
	LC string // C language expression
	T  TypeValue
	// GlobalDef Def
}

func (val *SimpleValue) ToC() string {
	return val.C
}
func (val *SimpleValue) Type() TypeValue {
	return val.T
}
func (lval *SimpleLValue) LToC() string {
	return lval.LC
}
func (lval *SimpleLValue) Type() TypeValue {
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
	p.ParseTop(cm, cg)
	log.Printf("LoadModule: Visit")
	cm.BigVisit(p)
	log.Printf("LoadModule: Done")
	return cm
}

func CompileToC(opt *Options, r io.Reader, sourceName string, w io.Writer) {
	cg := NewCGen(opt, w)
	// BootstrapModules(cg)
	cm := cg.Mods["main"]
	// BootstrapBuiltins(cm)
	cg.LoadModule("builtin")
	p := NewParser(r, sourceName)
	p.ParseTop(cm, cg)

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
	pre.cm.P("\n// MARCO: %s", def.Name)
	pre.cm.CGen.LoadModule(def.Name)
	pre.cm.P("\n// POLO: %s", def.Name)
}
func (pre *cPreMod) VisitDefConst(def *DefConst) {
	pre.mustNotExistYet(def.Name)
	pre.cm.GlobalDefs[def.Name] = def
	pre.cm.CGen.GlobalDefs[GlobalName(pre.cm.Package, def.Name)] = def
    if pre.cm.Package == "builtin" {
	    pre.cm.CGen.GlobalDefs[def.Name] = def
    }
	pre.cm.P("\n// PRE VISIT %#v\n", def)
}
func (pre *cPreMod) VisitDefVar(def *DefVar) {
	pre.mustNotExistYet(def.Name)
	pre.cm.GlobalDefs[def.Name] = def
	pre.cm.CGen.GlobalDefs[GlobalName(pre.cm.Package, def.Name)] = def
    if pre.cm.Package == "builtin" {
	    pre.cm.CGen.GlobalDefs[def.Name] = def
    }
	log.Printf("pre visit DefVar: %v => %v", def, pre.cm.GlobalDefs)
	pre.cm.P("\n// PRE VISIT %#v\n", def)
}
func (pre *cPreMod) VisitDefType(def *DefType) {
	pre.mustNotExistYet(def.Name)
	pre.cm.GlobalDefs[def.Name] = def
	pre.cm.CGen.GlobalDefs[GlobalName(pre.cm.Package, def.Name)] = def
    if pre.cm.Package == "builtin" {
	    pre.cm.CGen.GlobalDefs[def.Name] = def
    }
	pre.cm.P("\n// PRE VISIT %#v\n", def)
}
func (pre *cPreMod) VisitDefFunc(def *DefFunc) {
	fn := def.FunctionRec
	pre.mustNotExistYet(def.Name)
	pre.cm.GlobalDefs[def.Name] = def
	pre.cm.CGen.GlobalDefs[GlobalName(pre.cm.Package, def.Name)] = def
    if pre.cm.Package == "builtin" {
	    pre.cm.CGen.GlobalDefs[def.Name] = def
    }

	// TODO -- dedup
	var b Buf
	b.P("void %s(", def.C)
	if len(fn.Ins) > 0 {
		firstTime := true
		for _, name_and_type := range fn.Ins {
			if !firstTime {
				b.P(", ")
			}
			b.P("%s %s", name_and_type.TV.CType(), "v_"+name_and_type.Name)
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
	Mods       map[string]*CMod
	GlobalDefs map[string]Def
	Options    *Options
	W          *bufio.Writer
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
		Mods:       map[string]*CMod{"main": mainMod},
		W:          mainMod.W,
		Options:    opt,
		GlobalDefs: make(map[string]Def),
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
	return &SimpleLValue{LC: Format("&(%s)", value.ToC()), T: value.Type()}
}
func (cm *CMod) VisitLValSub(x *SubX) LValue {
	value := cm.VisitSub(x)
	return &SimpleLValue{LC: Format("TODO_LValue(%s)", value.ToC()), T: value.Type()}
}
func (cm *CMod) VisitLvalDot(x *DotX) LValue {
	value := cm.VisitDot(x)
	return &SimpleLValue{LC: Format("&(%s)", value.ToC()), T: value.Type()}
}

func (cm *CMod) VisitLitInt(x *LitIntX) Value {
	return &SimpleValue{
		C: Format("%d", x.X),
		T: ConstIntTO,
	}
}
func (cm *CMod) VisitLitString(x *LitStringX) Value {
	return &SimpleValue{
		C: Format("%q", x.X),
		T: StringTO,
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
	if gd, ok := cm.CGen.GlobalDefs[x.X]; ok {
		return gd
	}
	// Else, assume it is a local variable.
	return &SimpleValue{C: "v_" + x.X, T: IntTO}
}
func (cm *CMod) VisitBinOp(x *BinOpX) Value {
	a := x.A.VisitExpr(cm)
	b := x.B.VisitExpr(cm)
	return &SimpleValue{
		C: Format("(%s) %s (%s)", a.ToC(), x.Op, b.ToC()),
		T: IntTO,
	}
}
func (cm *CMod) VisitConstructor(x *ConstructorX) Value {
	return &SimpleValue{
		C: Format("(%s) alloc(C_%s)", x.Name, x.Name),
		T: &PointerTV{BaseTV{}, &StructTV{BaseTV{x.Name}, nil}},
	}
}

func AssignNewVar(in NameAndType, out NameAndType) string {
	ctype := in.TV.CType()
	z, ok := in.TV.Assign(in.Name, in.TV)
	if ok {
		return Format("%s %s = %s;", ctype, out.Name, z)
	}
	panic(Format("Cannot assign from %s (type %s) to %s (type %s)", in.Name, in.TV, out.Name, out.TV))
}

func (cm *CMod) VisitCall(x *CallX) Value {
	ser := Serial("call")
	cm.P("// %s: Calling Func: %#v", ser, x.Func)
	for i, a := range x.Args {
		cm.P("// %s: Calling with Arg [%d]: %#v", ser, i, a)
	}

	cm.P("{")
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

		if funcRec.HasDotDotDot && i == len(funcRec.Ins)-1 {
			sliceT, ok := expectedType.(*SliceTV)
			assert(ok)
			// TODO DotDotDot
			elementT := sliceT.E
			c2 += Format("Slice %s_in_rest = CreateSlice();", ser)
			for j := i; j < len(x.Args); j++ {
				cm.P(AssignNewVar(
					NameAndType{val.ToC(), val.Type()},
					NameAndType{Format("%s_in_%d", ser, j), elementT}))
				c2 += Format("AppendSlice(%d_in_rest,  %s_in_%d);", ser, ser, j)
			}
			c += Format("FINISH(%s_in_rest);", ser)

		} else {
			cm.P(AssignNewVar(
				NameAndType{val.ToC(), val.Type()},
				NameAndType{Format("%s_in_%d", ser, i), expectedType}))
			//##cm.P("  %s %s_in_%d = %s;", in.CType(), ser, i, val.ToC())
			c += Format(", %s_in_%d", ser, i)
		}

	}
	for i, out := range funcRec.Outs {
		cm.P("  %s %s_out_%d;", out.TV.CType(), ser, i)
		c += Format(", &%s_out_%d", ser, i)
	}
	c += " );"
	cm.P("[[[%s]]]  %s\n} // %s", c2, c, ser)

	switch len(funcRec.Outs) {
	case 0:
		return &SimpleValue{"VOID", VoidTO}
	case 1:
		return &SimpleValue{Format("%s_out_0", ser), funcRec.Outs[0].TV}
	default:
		return &SimpleValue{ser, ListTO}
	}
}

func (cm *CMod) VisitSub(x *SubX) Value {
	return &SimpleValue{
		C: Format("SubXXX(%v)", x),
		T: IntTO,
	}
}
func (cm *CMod) VisitDot(dot *DotX) Value {
	log.Printf("VisitDot: <------ %#v", dot)
	val := dot.X.VisitExpr(cm)
	log.Printf("VisitDot: val---- %#v", val)
	if val.Type() == ImportTO {
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
		T: IntTO,
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

		if lenB != len(bcall.Args) {
			panic(Format("Function %s wants %d args, got %d", funcname, len(bcall.Args), lenB))
		}
		ser := Serial("call")
		cm.P("{ // %s", ser)
		c := Format(" %s( fp", funcname)
		for i, in := range funcRec.Ins {
			val := ass.B[i].VisitExpr(cm)
			expectedType := in.TV
			if expectedType != val.Type() {
				panic(Format("bad type: expected %s, got %s", expectedType, val.Type()))
			}
			cm.P("  %s %s_in_%d = %s;", in.TV.CType(), ser, i, val.ToC())
			c += Format(", %s_in_%d", ser, i)
		}
		for i, out := range funcRec.Outs {
			cm.P("  %s %s_out_%d;", out.TV.CType(), ser, i)
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
					cm.P("  %s = (%s)(%s);", cvar, val.Type().CType(), val.ToC())
				case ":=":
					// TODO check Globals
					cvar := Format("%s %s", val.Type().CType(), "v_"+t.X)
					cm.P("  %s = (%s)(%s);", cvar, val.Type().CType(), val.ToC())
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
	/*
		if a.Stmts == nil {
			panic(8882)
		}
	*/
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
	cm.P("%s V_%s__%s; // global var", def.T.CType(), cm.Package, def.Name)
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
	//fn.Body = &Block{FunctionRec: fn}
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
			b.P("%s %s", name_and_type.TV.CType(), "v_"+name_and_type.Name)
			firstTime = false
		}
	}
	if fn.Body != nil {
		b.P(") {\n")
		cm.P(b.String())
		fn.Body.VisitStmt(cm)
		cm.P("}\n")
	} else {
		b.P("); /*NATIVE*/\n")
	}
}

var SerialNum uint

func Serial(prefix string) string {
	SerialNum++
	return Format("%s_%d", prefix, SerialNum)
}
