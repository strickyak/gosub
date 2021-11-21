package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
)

var Format = fmt.Sprintf
var P = fmt.Fprintf

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

func FullName(a string, b string) string {
	return a + "__" + b
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
	VisitFunction(*FunctionX) Value
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
	FuncRec *FuncRec
}
type ImportTV struct {
	BaseTV
}
type ForwardTV struct {
	BaseTV
	Package string
	Name    string
}

func (tv *PrimTV) Type() TypeValue      { return &TypeTV{} }
func (tv *TypeTV) Type() TypeValue      { return &TypeTV{} }
func (tv *PointerTV) Type() TypeValue   { return &TypeTV{} }
func (tv *SliceTV) Type() TypeValue     { return &TypeTV{} }
func (tv *MapTV) Type() TypeValue       { return &TypeTV{} }
func (tv *ForwardTV) Type() TypeValue   { return &TypeTV{} }
func (tv *StructTV) Type() TypeValue    { return &TypeTV{} }
func (tv *InterfaceTV) Type() TypeValue { return &TypeTV{} }
func (tv *FunctionTV) Type() TypeValue  { return &TypeTV{} }
func (tv *ImportTV) Type() TypeValue    { return &TypeTV{} }

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
func (tv *ForwardTV) String() string   { return Format("%T(%#v)", *tv, *tv) }
func (tv *StructTV) String() string    { return Format("%T(%#v)", *tv, *tv) }
func (tv *InterfaceTV) String() string { return Format("%T(%#v)", *tv, *tv) }
func (tv *FunctionTV) String() string  { return Format("%T(%#v)", *tv, *tv) }
func (tv *ImportTV) String() string    { return Format("%T(%#v)", *tv, *tv) }

func (tv *PrimTV) VisitExpr(v ExprVisitor) Value      { return tv }
func (tv *TypeTV) VisitExpr(v ExprVisitor) Value      { return tv }
func (tv *PointerTV) VisitExpr(v ExprVisitor) Value   { return tv }
func (tv *SliceTV) VisitExpr(v ExprVisitor) Value     { return tv }
func (tv *MapTV) VisitExpr(v ExprVisitor) Value       { return tv }
func (tv *ForwardTV) VisitExpr(v ExprVisitor) Value   { return tv }
func (tv *StructTV) VisitExpr(v ExprVisitor) Value    { return tv }
func (tv *InterfaceTV) VisitExpr(v ExprVisitor) Value { return tv }
func (tv *FunctionTV) VisitExpr(v ExprVisitor) Value  { return tv }
func (tv *ImportTV) VisitExpr(v ExprVisitor) Value    { return tv }

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
	X        string
	Outer    *CMod // Outer scope where defined -- but the IdentX may or may not be global.
	Resolved bool  // if we looked for the *GDef.
	GDef     *GDef // cache the *GDef if resolved.
}

func (o *IdentX) FindGDef() *GDef {
	if !o.Resolved {
		if gdef, ok := o.Outer.GDefs[o.X]; ok {
			o.GDef = gdef
			o.Resolved = true
		} else if gdef, ok := o.Outer.CGen.BuiltinMod.GDefs[o.X]; ok {
			o.GDef = gdef
			o.Resolved = true
		}
	}
	return o.GDef
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
	FuncRec *FuncRec
}

func (o *FunctionX) String() string {
	return fmt.Sprintf("Function(%s)", o.FuncRec)
}
func (o *FunctionX) VisitExpr(v ExprVisitor) Value {
	return v.VisitFunction(o)
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

// A callable view of a node in a parse tree,
// e.g. global func, lambda, bound method,
// ... any expr of Func kind.
type FuncRec struct {
	Receiver     *NameAndType // nil if no Receiver.
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
	Locals  []NameAndType
	Stmts   []Stmt
	Parent  *Block
	FuncRec *FuncRec
}

func (o *Block) VisitStmt(v StmtVisitor) {
	v.VisitBlock(o)
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
	Package string
	CMod    *CMod

	// For lookup by local name.
	ImportsMap map[string]*GDef
	ConstsMap  map[string]*GDef
	VarsMap    map[string]*GDef
	TypesMap   map[string]*GDef
	FuncsMap   map[string]*GDef

	// For ordered traversal:
	Imports []*GDef
	Consts  []*GDef
	Vars    []*GDef
	Types   []*GDef
	Funcs   []*GDef
}

func NewParser(r io.Reader, filename string) *Parser {
	return &Parser{
		Lex:        NewLex(r, filename),
		ImportsMap: make(map[string]*GDef),
		ConstsMap:  make(map[string]*GDef),
		VarsMap:    make(map[string]*GDef),
		TypesMap:   make(map[string]*GDef),
		FuncsMap:   make(map[string]*GDef),
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
		z := &IdentX{o.Word, o.CMod, false, nil}
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
	// It starts with a Prim.
	a := o.ParsePrim()
LOOP:
	for {
		// Then it may be followed by something like x(...), x[...], x.f.
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
		return &ForwardTV{BaseTV{}, o.Package, t.X}
	case *DotX:
		switch tx := t.X.(type) {
		case *IdentX:
			// TODO: when tx.X is not an import name.
			return &ForwardTV{BaseTV{}, tx.X, t.Member}
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
			sig := &FuncRec{}
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
		yes := o.ParseBlock(b.FuncRec)
		var no *Block
		if o.Word == "else" {
			no = o.ParseBlock(b.FuncRec)
		}
		return &IfS{pred, yes, no}
	case "for":
		o.Next()
		var pred Expr
		if o.Word != "{" {
			pred = o.ParseExpr()
		}
		b2 := o.ParseBlock(b.FuncRec)
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
				bare := o.ParseBareBlock(b.FuncRec)
				sws.Cases = append(sws.Cases, &Case{matches, bare})
			case "default":
				o.TakePunc(":")
				bare := o.ParseBareBlock(b.FuncRec)
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
func (o *Parser) ParseBlock(fn *FuncRec) *Block {
	o.TakePunc("{")
	b := o.ParseBareBlock(fn)
	o.TakePunc("}")
	return b
}
func (o *Parser) ParseBareBlock(fn *FuncRec) *Block {
	b := &Block{FuncRec: fn}
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

func (o *Parser) ParseFunctionSignature(fn *FuncRec) {
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
	// this will have to be fixed to parse types
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
func (o *Parser) ParseFunc() *FuncRec {
	fn := &FuncRec{}
	o.ParseFunctionSignature(fn)
	if o.Kind != L_EOL {
		fn.Body = o.ParseBlock(fn)
	}
	return fn
}

func (o *Parser) ParseModule(cm *CMod, cg *CGen) {
	o.CMod = cm
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
				o.Package = w
			case "import":
				if o.Kind != L_String {
					Panicf("after import, expected string, got %v", o.Word)
				}
				w := o.Word
				o.Next()
				gd := &GDef{
					Name: w,
				}
				o.Imports = append(o.Imports, gd)
				o.ImportsMap[w] = gd
			case "const":
				w := o.TakeIdent()
				o.TakePunc("=")
				x := o.ParseExpr()
				gd := &GDef{
					Package: o.Package,
					Name:    w,
					Init:    x,
				}
				o.Consts = append(o.Consts, gd)
				o.ConstsMap[w] = gd
			case "var":
				w := o.TakeIdent()
				var t TypeValue
				if o.Word != "=" {
					t = o.ParseType()
				}
				var i Expr
				if o.Word == "=" {
					o.Next()
					i = o.ParseExpr()
				}
				gd := &GDef{
					Package: o.Package,
					Name:    w,
					Type:    t,
					Init:    i,
				}
				o.Vars = append(o.Vars, gd)
				o.VarsMap[w] = gd
			case "type":
				w := o.TakeIdent()
				var tv TypeValue
				if o.Word == "interface" {
					o.Next()
					tv = &InterfaceTV{BaseTV{w}, o.ParseInterfaceType(w)}
				} else if o.Word == "struct" {
					o.Next()
					tv = &StructTV{BaseTV{w}, o.ParseStructType(w)}
				} else if o.Word == "func" {
					panic("todo")
				} else {
					tv = o.ParseType()
				}
				gd := &GDef{
					Package: o.Package,
					Name:    w,
					Init:    tv,
				}
				o.Types = append(o.Types, gd)
				o.TypesMap[w] = gd
			case "func":
				var receiver *NameAndType
				switch o.Kind {
				case L_Punc:
					// Distinguished Receiver:
					o.TakePunc("(")
					rName := o.TakeIdent() // limitation: name required.
					rType := o.ParseType()
					o.TakePunc(")")
					receiver = &NameAndType{rName, rType}
				}
				name := o.TakeIdent()
				fn := o.ParseFunc()
				fn.Receiver = receiver // may be nil
				gd := &GDef{
					Package: o.Package,
					Name:    name,
					Init:    &FunctionX{fn},
				}
				o.Funcs = append(o.Funcs, gd)
				o.FuncsMap[name] = gd
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
	Value
	LToC() string
}

type SimpleValue struct {
	C string // C language expression
	T TypeValue
}

type SimpleLValue struct {
	C  string // C language expression
	T  TypeValue
	LC string // C language expression, pointer to the LValue.
}

func (val *SimpleValue) ToC() string {
	return val.C
}
func (val *SimpleValue) Type() TypeValue {
	return val.T
}

func (lval *SimpleLValue) ToC() string {
	return lval.C
}
func (lval *SimpleLValue) LToC() string {
	return lval.LC
}
func (lval *SimpleLValue) Type() TypeValue {
	return lval.T
}

type ImportValue struct {
	Name string
}

func (val *ImportValue) ToC() string {
	panic(Format("cannot use import %q as a value", val.Name))
}
func (val *ImportValue) Type() TypeValue {
	return ImportTO
}

func (cg *CGen) LoadModule(name string) *CMod {
	log.Printf("LoadModule: << %q", name)
	if already, ok := cg.Mods[name]; ok {
		log.Printf("LoadModule: already loaded: %q", name)
		return already
	}

	filename := cg.Options.LibDir + "/" + name + ".go"
	r, err := os.Open(filename)
	if err != nil {
		panic(Format("cannot open %q: %v", filename, err))
	}
	defer r.Close()

	cm := NewCMod(name, cg, cg.W)
	cg.Mods[name] = cm

	log.Printf("LoadModule: Parser")
	p := NewParser(r, filename)
	p.ParseModule(cm, cg)
	log.Printf("LoadModule: VisitGlobals")
	cm.VisitGlobals(p)
	log.Printf("LoadModule: VisitGlobals Done")

	if name == "builtin" {
		cg.BuiltinMod = cm
	}
	return cm
}

func CompileToC(r io.Reader, sourceName string, w io.Writer, opt *Options) {
	cg, cm := NewCGenAndMainCMod(opt, w)
	cm.P("#include <stdio.h>")
	cm.P("#include \"runt.h\"")
	cg.LoadModule("builtin")
	p := NewParser(r, sourceName)
	p.ParseModule(cm, cg)

	cm.VisitGlobals(p)
}

func (cm *CMod) mustNotExistYet(s string) {
	if _, ok := cm.GDefs[s]; ok {
		Panicf("redefined global name: %s", s)
	}
}

func (cm *CMod) FirstSlotGlobals(p *Parser) {
	// first visit: Slot the globals.
	slot := func(g *GDef) {
		g.Package = cm.Package
		g.FullName = FullName(g.Package, g.Name)
		cm.mustNotExistYet(g.FullName)
	}
	for _, g := range p.Imports {
		slot(g)
	}
	for _, g := range p.Types {
		slot(g)
	}
	for _, g := range p.Consts {
		slot(g)
	}
	for _, g := range p.Vars {
		slot(g)
	}
	for _, g := range p.Funcs {
		slot(g)
	}
}

func (cm *CMod) SecondBuildGlobals(p *Parser) {
	for _, g := range p.Imports {
		cm.CGen.LoadModule(g.Name)
		g.Value = &ImportValue{g.Name}
		// If we care to do imports in order,
		// this is a good place to rememer it.
		cm.CGen.ModsInOrder = append(cm.CGen.ModsInOrder, g.Name)
	}
	for _, g := range p.Types {
		g.Value = g.Init.VisitExpr(cm)
	}
	for _, g := range p.Types {
		_ = g
		// CHECK g.Value for completeness?
	}
	for _, g := range p.Consts {
		// not allowing g.Type on constants.
		g.Value = g.Init.VisitExpr(cm)
	}
	for _, g := range p.Vars {
		typeValue := g.Type.VisitExpr(cm).(TypeValue)
		g.Value = &SimpleValue{
			C: g.FullName,
			T: typeValue,
		}
		if g.Init != nil {
			// We are writing the global init() function.
			initS := &AssignS{
				A:  []Expr{&IdentX{g.Name, cm, true, g}},
				Op: "=",
				B:  []Expr{g.Init},
			}
			initS.VisitStmt(cm)
		}
	}
	for _, g := range p.Funcs {
		g.Value = g.Init.VisitExpr(cm)
	}
}

func (cm *CMod) ThirdDefineGlobals(p *Parser) {
	// Third: Define the globals, except for functions.
	say := func(how string, g *GDef) {
		cm.P("// Third == %s %s ==", how, g.FullName)
	}
	_ = say
	for _, g := range p.Types {
		say("type", g)
		cm.P("typedef %s %s;", g.Value.(TypeValue).CType(), g.FullName)
	}
	for _, g := range p.Vars {
		say("var", g)
		cm.P("%s %s;", g.Value.Type().CType(), g.FullName)
	}
	for _, g := range p.Funcs {
		say("func", g)
		cm.P("extern %s %s;", "FUNC" /*g.Value.Type().CType()*/, g.FullName)
	}
}

func (cm *CMod) FourthInitGlobals(p *Parser) {
	// Fourth: Initialize the global vars.
	cm.P("void GLOBAL_init() {")
	say := func(how string, g *GDef) {
		cm.P("// Fourth == %s %s ==", how, g.FullName)
	}
	for _, g := range p.Vars {
		say("var", g)
		if g.Init != nil {
			initS := &AssignS{
				A:  []Expr{&IdentX{g.Name, cm, true, g}},
				Op: "=",
				B:  []Expr{g.Init},
			}
			// Emit initialization of var into init() function.
			initS.VisitStmt(cm)
		}
	}
	for _, g := range p.Funcs {
		say("func TODO: Inline init functions:", g)
	}
	cm.P("} // GLOBAL_init()")
}
func (cm *CMod) FifthPrintFunctions(p *Parser) {
	for _, g := range p.Funcs {
		cm.P("extern %s %s;", "FUNC" /*g.Value.Type().CType()*/, g.FullName)
	}
}

func (cm *CMod) VisitGlobals(p *Parser) {
	/*
	   PLAN: visit all the "def" first.
	   Then visit all the "init".
	   Then visit all the functions.
	   So only one output path is needed.
	*/
	cm.FirstSlotGlobals(p)
	cm.SecondBuildGlobals(p)
	cm.ThirdDefineGlobals(p)
	cm.FourthInitGlobals(p)
	cm.FifthPrintFunctions(p)
	cm.Flush()
}

/*
func (pre *cPreMod) PreVisitDefFunc(def *GDef) {
	fn := def.Expr.(*FunctionX).FuncRec
	pre.mustNotExistYet(def.Name)
	pre.cm.GDefs[def.Name] = def
	pre.cm.CGen.GDefs[GlobalName(pre.cm.Package, def.Name)] = def
	if pre.cm.Package == "builtin" {
		pre.cm.CGen.GDefs[def.Name] = def
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
*/

type GDef struct {
	CGen     *CGen
	Package  string
	Name     string
	FullName string
	Value    Value // Next resolve global names to Values.
	Init     Expr  // for Const or Var or Type
	Type     Expr  // for Var, not yet for Const.
	Active   bool
}

type CMod struct {
	defsBuf    bytes.Buffer
	initBuf    bytes.Buffer
	D          *bufio.Writer // Non-executable Declarations
	W          *bufio.Writer
	Package    string
	GDefs      map[string]*GDef // by short name
	BreakTo    string
	ContinueTo string
	CGen       *CGen
	/*
		Structs    []*StructRec
		Interfaces []*InterfaceRec
		Funcs      []*FuncRec
	*/
}
type CGen struct {
	Mods        map[string]*CMod // by module name
	BuiltinMod  *CMod
	ModsInOrder []string         // reverse definition order
	GDefs       map[string]*GDef // by full name
	Options     *Options
	W           *bufio.Writer
}

func NewCMod(name string, cg *CGen, w io.Writer) *CMod {
	z := &CMod{
		W:       bufio.NewWriter(w),
		Package: name,
		GDefs:   make(map[string]*GDef),
		CGen:    cg,
	}
	z.D = bufio.NewWriter(&z.defsBuf)
	z.D = z.W
	return z
}
func NewCGenAndMainCMod(opt *Options, w io.Writer) (*CGen, *CMod) {
	mainMod := NewCMod("main", nil, w)
	cg := &CGen{
		Mods:    map[string]*CMod{"main": mainMod},
		W:       mainMod.W,
		Options: opt,
		GDefs:   make(map[string]*GDef),
	}
	mainMod.CGen = cg
	return cg, mainMod
}

/*
func (cm *CMod) Onto(w io.Writer, fn func()) {
    saved := cm.W
    cm.W = w
    defer func() { cm.W = saved }
    fn()
}
*/

func (cm *CMod) P(format string, args ...interface{}) {
	// log.Printf("<<<<< %q >>>>> %q", format, fmt.Sprintf(format, args...))
	fmt.Fprintf(cm.W, format+"\n", args...)
}
func (cm *CMod) Flush() {
	cm.W.Flush()
}

func AssignNewVar(in NameAndType, out NameAndType) string {
	ctype := in.TV.CType()
	z, ok := in.TV.Assign(in.Name, in.TV)
	if ok {
		return Format("%s %s = %s;", ctype, out.Name, z)
	}
	panic(Format("Cannot assign from %s (type %s) to %s (type %s)", in.Name, in.TV, out.Name, out.TV))
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
	if gd, ok := cm.GDefs[x.X]; ok {
		return gd.Value
	}
	if gd, ok := cm.CGen.GDefs[x.X]; ok {
		return gd.Value
	}
	// Else, assume it is a local variable.
	return &SimpleValue{C: "v_TODO_" + x.X, T: IntTO}
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
func (cm *CMod) VisitFunction(x *FunctionX) Value {
	return nil // TODO
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
	_ = funcX // TODO

	/* TODO

		log.Printf("funcX: %#v", funcX)
		funcRec := funcX.(*DefFunc).FuncRec
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
	    TODO */
	panic("TODO=1733")
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
		println("GD", otherMod.GDefs)
		_, ok := otherMod.GDefs[dot.Member]
		if !ok {
			panic(Format("cannot find member %s in module %s", dot.Member, modName))
		}
		return otherMod.VisitIdent(&IdentX{X: modName})
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
	_ = lenB // TODO

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

		/* TODO
				funcRec := visited.(*DefFunc).FuncRec

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
		        TODO */
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
	for i, e := range a.Stmts {
		log.Printf("VisitBlock[%d]", i)
		e.VisitStmt(cm)
	}
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

/* TODO
func (cm *CMod) VisitDefFunc(def *DefFunc) {
	fn := def.FuncRec
	//fn.Body = &Block{FuncRec: fn}
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
		b.P("); //NATIVE//\n")
	}
}
    TODO */

var SerialNum uint

func Serial(prefix string) string {
	SerialNum++
	return Format("%s_%d", prefix, SerialNum)
}

///////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////

// Keep track of reachable Globals.
type ActiveTracker struct {
	List  []*GDef    // Reverse Definition Order.
	Stack complex128 // XXX ?
}

/*
func (o *ActiveTracker) activate(g *GDef) {
    if g.Active {
        return // Already been there.
    }
    g.Active = true
    o.List = append(o.List, g)
    if g.Expr !=nil {
        g.Expr.VisitExpr(o)
    }
}

func (o *ActiveTracker) VisitLvalIdent(x *IdentX) LValue {
    nando
	value := cm.VisitIdent(x)
	return &SimpleLValue{LC: Format("&(%s)", value.ToC()), T: value.Type()}
}
func (o *ActiveTracker) VisitLValSub(x *SubX) LValue {
	value := cm.VisitSub(x)
	return &SimpleLValue{LC: Format("TODO_LValue(%s)", value.ToC()), T: value.Type()}
}
func (o *ActiveTracker) VisitLvalDot(x *DotX) LValue {
	value := cm.VisitDot(x)
	return &SimpleLValue{LC: Format("&(%s)", value.ToC()), T: value.Type()}
}

func (o *ActiveTracker) VisitLitInt(x *LitIntX) Value {
	return &SimpleValue{
		C: Format("%d", x.X),
		T: ConstIntTO,
	}
}
func (o *ActiveTracker) VisitLitString(x *LitStringX) Value {
	return &SimpleValue{
		C: Format("%q", x.X),
		T: StringTO,
	}
}
func (o *ActiveTracker) VisitIdent(x *IdentX) Value {
	log.Printf("VisitIdent <= %v", x)
	z := cm._VisitIdent_(x)
	log.Printf("VisitIdent => %#v", z)
	return z
}
func (o *ActiveTracker) _VisitIdent_(x *IdentX) Value {
	if gd, ok := cm.GDefs[x.X]; ok {
		return gd.Value
	}
	if gd, ok := cm.CGen.GDefs[x.X]; ok {
		return gd.Value
	}
	// Else, assume it is a local variable.
	return &SimpleValue{C: "v_TODO_" + x.X, T: IntTO}
}
func (o *ActiveTracker) VisitBinOp(x *BinOpX) Value {
	a := x.A.VisitExpr(cm)
	b := x.B.VisitExpr(cm)
	return &SimpleValue{
		C: Format("(%s) %s (%s)", a.ToC(), x.Op, b.ToC()),
		T: IntTO,
	}
}
func (o *ActiveTracker) VisitConstructor(x *ConstructorX) Value {
	return &SimpleValue{
		C: Format("(%s) alloc(C_%s)", x.Name, x.Name),
		T: &PointerTV{BaseTV{}, &StructTV{BaseTV{x.Name}, nil}},
	}
}
func (o *ActiveTracker) VisitFunction(x *FunctionX) Value {
	return nil // TODO
}

func (o *ActiveTracker) VisitCall(x *CallX) Value {
	ser := Serial("call")
	cm.P("// %s: Calling Func: %#v", ser, x.Func)
	for i, a := range x.Args {
		cm.P("// %s: Calling with Arg [%d]: %#v", ser, i, a)
	}

	cm.P("{")
	log.Printf("x.Func: %#v", x.Func)
	funcX := x.Func.VisitExpr(cm)
	_ = funcX // TODO

	/# TODO

		log.Printf("funcX: %#v", funcX)
		funcRec := funcX.(*DefFunc).FuncRec
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
	    TODO #/
	panic("TODO=1733")
}

func (o *ActiveTracker) VisitSub(x *SubX) Value {
	return &SimpleValue{
		C: Format("SubXXX(%v)", x),
		T: IntTO,
	}
}
func (o *ActiveTracker) VisitDot(dot *DotX) Value {
	log.Printf("VisitDot: <------ %#v", dot)
	val := dot.X.VisitExpr(cm)
	log.Printf("VisitDot: val---- %#v", val)
	if val.Type() == ImportTO {
		modName := val.ToC() // is there a better way?
		println("DOT", modName, dot.Member)
		otherMod := cm.CGen.Mods[modName] // TODO: import aliases.
		println("OM", otherMod)
		println("GD", otherMod.GDefs)
		_, ok := otherMod.GDefs[dot.Member]
		if !ok {
			panic(Format("cannot find member %s in module %s", dot.Member, modName))
		}
		return otherMod.VisitIdent(&IdentX{X: modName})
	}

	z := &SimpleValue{
		C: Format("DotXXX(%v)", dot),
		T: IntTO,
	}
	log.Printf("VisitDot: Not Import: ----> %v", z)
	return z
}
func (o *ActiveTracker) VisitAssign(ass *AssignS) {
	cm.P("//## assign..... %v   %v   %v", ass.A, ass.Op, ass.B)
	lenA, lenB := len(ass.A), len(ass.B)
	_ = lenA
	_ = lenB // TODO

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

		/# TODO
				funcRec := visited.(*DefFunc).FuncRec

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
		        TODO #/
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
func (o *ActiveTracker) VisitReturn(ret *ReturnS) {
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
func (o *ActiveTracker) VisitWhile(wh *WhileS) {
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
func (o *ActiveTracker) VisitBreak(sws *BreakS) {
	if cm.BreakTo == "" {
		Panicf("cannot break from here")
	}
	cm.P("goto %s;", cm.BreakTo)
}
func (o *ActiveTracker) VisitContinue(sws *ContinueS) {
	if cm.ContinueTo == "" {
		Panicf("cannot continue from here")
	}
	cm.P("goto %s;", cm.ContinueTo)
}
func (o *ActiveTracker) VisitIf(ifs *IfS) {
	cm.P("  { t_bool _if_ = %s;", ifs.Pred.VisitExpr(cm).ToC())
	cm.P("  if( _if_ ) {")
	ifs.Yes.VisitStmt(cm)
	if ifs.No != nil {
		cm.P("  } else {")
		ifs.No.VisitStmt(cm)
	}
	cm.P("  }}")
}
func (o *ActiveTracker) VisitSwitch(sws *SwitchS) {
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
func (o *ActiveTracker) VisitBlock(a *Block) {
	if a == nil {
		panic(8881)
	}
	for i, e := range a.Stmts {
		log.Printf("VisitBlock[%d]", i)
		e.VisitStmt(cm)
	}
}
*/
