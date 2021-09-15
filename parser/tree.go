package parser

import (
	"bytes"
	"fmt"
)

//////// Expr

type ExprVisitor interface {
    VisitInt(*IntX)
    VisitString(*StringX)
    VisitIdent(*IdentX)
    VisitBinOp(*BinOpX)
    VisitList(*ListX)
}

type Expr interface {
	String() string
    VisitExpr(ExprVisitor)
}

type IntX struct {
	X int
}

func (o *IntX) String() string {
	return fmt.Sprintf("Int(%d)", o.X)
}
func (o *IntX) VisitExpr(v ExprVisitor) {
    v.VisitInt(o)
}

type StringX struct {
	X string
}

func (o *StringX) String() string {
	return fmt.Sprintf("String(%q)", o.X)
}
func (o *StringX) VisitExpr(v ExprVisitor) {
    v.VisitString(o)
}

type IdentX struct {
	X string
}

func (o *IdentX) String() string {
	return fmt.Sprintf("Ident(%s)", o.X)
}
func (o *IdentX) VisitExpr(v ExprVisitor) {
    v.VisitIdent(o)
}

type BinOpX struct {
	A  Expr
	Op string
	B  Expr
}

func (o *BinOpX) String() string {
	return fmt.Sprintf("Bin(%v %q %v)", o.A, o.Op, o.B)
}
func (o *BinOpX) VisitExpr(v ExprVisitor) {
    v.VisitBinOp(o)
}

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
func (o *ListX) VisitExpr(v ExprVisitor) {
    v.VisitList(o)
}

/////////// Stmt

type Stmt interface {
	String() string
}

type AssignS struct {
	A  Expr
	Op string
	B  Expr
}

func (o *AssignS) String() string {
	return fmt.Sprintf("\nAssign(%v %q %v)\n", o.A, o.Op, o.B)
}

type ReturnS struct {
	X Expr
}

func (o *ReturnS) String() string {
	return fmt.Sprintf("\nReturn(%v)\n", o.X)
}

////////////////////////

type TDef interface {
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
	Body   []Stmt
	Parent *Block
	Fn     *DefFunc
}
type DefFunc struct {
	Name string
	Type Type
	Ins  []NameAndType
	Outs []NameAndType
	Body *Block
}

type Type interface {
}

type IntType struct {
	Size   int
	Signed bool
}

var Byte = &IntType{Size: 1, Signed: false}
var Int = &IntType{Size: 2, Signed: true}
var UInt = &IntType{Size: 2, Signed: false}
