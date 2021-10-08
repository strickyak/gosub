package parser

import (
	"bytes"
	"fmt"
)

func Panicf(format string, args ...interface{}) string {
	s := Format(format, args...)
	panic(s)
}

//////// Expr

type ExprVisitor interface {
	VisitLitInt(*LitIntX) Value
	VisitLitString(*LitStringX) Value
	VisitIdent(*IdentX) Value
	VisitBinOp(*BinOpX) Value
	VisitList(*ListX) Value
	VisitCall(*CallX) Value
	VisitType(*TypeX) Value
	VisitSub(*SubX) Value
	VisitDot(*DotX) Value
}

type Expr interface {
	String() string
	VisitExpr(ExprVisitor) Value
}

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

// Maybe ListX should not be an Expr, but a different sort of thing, entirely.
type ListX struct {
	V []Expr
}

func (o *ListX) String() string {
	var buf bytes.Buffer
	buf.WriteString("List(")
	for i, e := range o.V {
		buf.WriteString(fmt.Sprintf("[%d] %v ", i, e))
	}
	buf.WriteString(")")
	return buf.String()
}
func (o *ListX) VisitExpr(v ExprVisitor) Value {
	return v.VisitList(o)
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

type TypeX struct {
	T Type
}

func (o *TypeX) String() string {
	return fmt.Sprintf("TypeX(%s)", o.T)
}
func (o *TypeX) VisitExpr(v ExprVisitor) Value {
	return v.VisitType(o)
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

type SwitchEntry struct {
	Cases []Expr
	Body  *Block
}
type SwitchS struct {
	Pred    Expr
	Entries []*SwitchEntry
}

func (o *SwitchS) String() string {
	return fmt.Sprintf("\nSwitch(%v)\n", o.Pred)
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

type Def interface {
	VisitDef(DefVisitor)
}
type DefVisitor interface {
	VisitDefPackage(*DefPackage)
	VisitDefImport(*DefImport)
	VisitDefConst(*DefConst)
	VisitDefVar(*DefVar)
	VisitDefType(*DefType)
	VisitDefFunc(*DefFunc)
}

type DefPackage struct {
	Name string
}
type DefImport struct {
	Name string
	Path string
}
type DefConst struct {
	Name string
	Expr Expr
}
type DefVar struct {
	Name string
	Type Type
}
type DefType struct {
	Name string
	Type Type
}
type NameAndType struct {
	Name string
	Type Type
}
type Block struct {
	Locals []NameAndType
	Stmts  []Stmt
	Parent *Block
	Func   *DefFunc
}

func (o *Block) VisitStmt(v StmtVisitor) {
	v.VisitBlock(o)
}

type DefFunc struct {
	Name string
	Type Type
	Ins  []NameAndType
	Outs []NameAndType
	Body *Block
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
type TypeVisitor interface {
	VisitIntType(*IntType)
	VisitSliceType(*SliceType)
}
type Type interface {
	TypeNameInC(string) string
}

type IntType struct {
	Size   int
	Signed bool
}
type SliceType struct {
	MemberType Type
}

func (o *IntType) VisitType(v TypeVisitor) {
	v.VisitIntType(o)
}
func (o *SliceType) VisitType(v TypeVisitor) {
	v.VisitSliceType(o)
}

var Bool = &IntType{Size: 1, Signed: false}
var Byte = &IntType{Size: 1, Signed: false}
var Int = &IntType{Size: 2, Signed: true}
var UInt = &IntType{Size: 2, Signed: false}

// With Size:0, a ConstInt represents const number that has infinite size.
var ConstInt = &IntType{Size: 0, Signed: true}

func CondString(pred bool, yes string, no string) string {
	if pred {
		return yes
	}
	return no
}

func (o *IntType) TypeNameInC(v string) string {
	if o.Size == 0 {
		return "t_int2 " + v
	}
	return fmt.Sprintf("t_%sint%d %s", CondString(o.Signed, "", "u"), o.Size, v)
}

func (o *SliceType) TypeNameInC(v string) string {
	memberType := o.MemberType.TypeNameInC("")
	return "_SLICE_X_" + memberType + "_Y_ " + v
}
*/

type Type string

func TypeNameInC(type_ Type) string {
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
		return "string"

	case SlicePre:
		return "slice"
	case MapPre:
		return "map"
	case ChanPre:
		return "chan"
	case FuncPre:
		return "func"
	case HandlePre:
		return "oop"
	case InterfacePre:
		return "interface"
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

const BoolPre = 'a'
const BytePre = 'b'
const UintPre = 'u'
const IntPre = 'i'
const ConstIntPre = 'c'
const StringPre = 's'
const TypePre = 't'

const SlicePre = 'S'
const MapPre = 'M'
const ChanPre = 'C'
const FuncPre = 'F'
const HandlePre = 'H'
const InterfacePre = 'I'

const SliceForm = "S%s"
const MapForm = "M%s%s"
const ChanForm = "C%s"
const FuncForm = "F(%s;%s)"
const HandleForm = "H{%s}"
const InterfaceForm = "I{%s}"
