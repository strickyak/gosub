package parser

import (
	"bytes"
	"fmt"
)

//////// Expr

type ExprVisitor interface {
	VisitLitInt(*LitIntX) Value
	VisitLitString(*LitStringX) Value
	VisitIdent(*IdentX) Value
	VisitBinOp(*BinOpX) Value
	VisitList(*ListX) Value
	VisitCall(*CallX) Value
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

/////////// Stmt

type StmtVisitor interface {
	VisitAssign(*AssignS)
	VisitReturn(*ReturnS)
	VisitBlock(*Block)
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

type ReturnS struct {
	X []Expr
}

func (o *ReturnS) String() string {
	return fmt.Sprintf("\nReturn(%v)\n", o.X)
}

func (o *AssignS) VisitStmt(v StmtVisitor) {
	v.VisitAssign(o)
}

func (o *ReturnS) VisitStmt(v StmtVisitor) {
	v.VisitReturn(o)
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

type TypeVisitor interface {
	VisitIntType(*IntType)
}
type Type interface {
	TypeNameInC(string) string
}

type IntType struct {
	Size   int
	Signed bool
}

func (o *IntType) VisitType(v TypeVisitor) {
	v.VisitIntType(o)
}

var Byte = &IntType{Size: 1, Signed: false}
var Int = &IntType{Size: 2, Signed: true}
var UInt = &IntType{Size: 2, Signed: false}

// With Size:0, a ConstInt represents const number that has infinite size.
var ConstInt = &IntType{Size: 0, Signed: true}

func (o IntType) TypeNameInC(v string) string {
	if o.Signed {
		switch o.Size {
		case 1:
			return "signed char " + v
		case 2:
			return "int " + v
		}
	} else {
		switch o.Size {
		case 1:
			return "unsigned char " + v
		case 2:
			return "unsigned int " + v
		}
	}
	panic("bad")
}
