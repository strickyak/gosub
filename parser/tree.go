package parser

import (
	//"bytes"
	"fmt"
)

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
	AsTStr() TStr
}

type PrimTV struct {
	S TStr
}
type PointerTV struct {
	E TypeValue
}
type SliceTV struct {
	E TypeValue
}
type MapTV struct {
	K TypeValue
	V TypeValue
}
type NamedTV struct {
	Name string
}
type StructTV struct {
	StructRec *StructRec
}
type InterfaceTV struct {
	InterfaceRec *InterfaceRec
}
type FunctionTV struct {
	FunctionRec *FunctionRec
}

// a
func (tv *PrimTV) AsTStr() TStr {
	return tv.S
}
func (tv *PointerTV) AsTStr() TStr {
	return TStr(Format(PointerForm, tv.E.TStr()))
}
func (tv *SliceTV) AsTStr() TStr {
	return TStr(Format(SliceForm, tv.E.TStr()))
}
func (tv *MapTV) AsTStr() TStr {
	return TStr(Format(MapForm, tv.K.TStr(), tv.V.TStr()))
}
func (tv *NamedTV) AsTStr() TStr {
	return Format("-named-%s", tv.Name) // TODO
}
func (tv *StructTV) AsTStr() TStr {
	return TStr(Format(StructForm, tv.Name))
}
func (tv *InterfaceTV) AsTStr() TStr {
	return TStr(Format(InterfaceForm, tv.Name))
}
func (tv *FunctionTV) AsTStr() TStr {
	return TStr(Format(FuncForm, "-in", "-")) // TODO
}

// b
func (tv *PrimTV) String() string {
	return Format("%T(%s)", *tv, tv.TStr())
}
func (tv *PointerTV) String() string {
	return Format("%T(%s)", *tv, tv.TStr())
}
func (tv *SliceTV) String() string {
	return Format("%T(%s)", *tv, tv.TStr())
}
func (tv *MapTV) String() string {
	return Format("%T(%s)", *tv, tv.TStr())
}
func (tv *NamedTV) String() string {
	return Format("%T(%s)", *tv, tv.TStr())
}
func (tv *StructTV) String() string {
	return Format("%T(%s)", *tv, tv.TStr())
}
func (tv *InterfaceTV) String() string {
	return Format("%T(%s)", *tv, tv.TStr())
}
func (tv *FunctionTV) String() string {
	return Format("%T(%s)", *tv, tv.TStr())
}

// c
func (tv *PrimTV) VisitExpr(v ExprVisitor) Value {
	return tv
}
func (tv *PointerTV) VisitExpr(v ExprVisitor) Value {
	return tv
}
func (tv *SliceTV) VisitExpr(v ExprVisitor) Value {
	return tv
}
func (tv *MapTV) VisitExpr(v ExprVisitor) Value {
	return tv
}
func (tv *NamedTV) VisitExpr(v ExprVisitor) Value {
	return tv
}
func (tv *StructTV) VisitExpr(v ExprVisitor) Value {
	return tv
}
func (tv *InterfaceTV) VisitExpr(v ExprVisitor) Value {
	return tv
}
func (tv *FunctionTV) VisitExpr(v ExprVisitor) Value {
	return tv
}

// d

/*
func (o *TypeX) String() string {
	return fmt.Sprintf("TypeX(%q)", o.T)
}
func (o *TypeX) VisitExpr(v ExprVisitor) Value {
	return v.VisitType(o)
}
*/

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

func (d *DefCommon) TStr() TStr { return "?common?" }

func (d *DefCommon) ToC() string      { return d.C }
func (d *DefCommon) Type() TypeValue  { return d.T }
func (d *DefCommon) LToC() string     { return "" }
func (d *DefCommon) LType() TypeValue { return "" }
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
	Def        *DefFunc // nil, if not a global func.
	Function   NameAndType
	IsMethod   bool
	Ins        []NameAndType
	Outs       []NameAndType
	IsEllipsis bool
	Body       *Block
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

type TStr string

func TypeNameInC(type_ TStr) string {
	switch type_[0] {
	case BoolPre:
		return "bool"
	case BytePre:
		return "byte"
	case UintPre:
		return "word"
	case IntPre:
		return "int"
	case ConstIntPre:
		return "int"
	case StringPre:
		return "String"

	case SlicePre:
		return "Slice"
	case MapPre:
		return "Map"
	case ChanPre:
		return "Chan"
	case FuncPre:
		return "Func"
	case HandlePre:
		return "Handle"
	case InterfacePre:
		return "InterfaceRec"
	case TypePre:
		return "type"
	default:
		Panicf("unknown type: %q", type_)
	}
	panic(0)
}

const BoolType = "a"
const ByteType = "b"
const UintType = "u"
const IntType = "i"
const ConstIntType = "c"
const StringType = "s"
const TypeType = "t"
const ImportType = "@"
const VoidType = "v"
const ListType = "l"

const BoolPre = 'a'
const BytePre = 'b'
const UintPre = 'u'
const IntPre = 'i'
const ConstIntPre = 'c'
const StringPre = 's'
const TypePre = 't'
const ImportPre = '@'
const VoidPre = 'v'
const ListPre = 'l'

const SlicePre = 'S'
const MapPre = 'M'
const ChanPre = 'C'
const FuncPre = 'F'
const HandlePre = 'H'
const StructPre = 'R'
const InterfacePre = 'I'

// const PointerPre = 'P'

const SliceForm = "S%s"
const MapForm = "M%s%s"
const ChanForm = "C%s"
const TypeForm = "t(%s)"
const FuncForm = "F(%s;%s)"
const StructForm = "R{%s}"
const HandleForm = "H{%s}"
const InterfaceForm = "I{%s}"

// const PointerForm = "P{%s}"
