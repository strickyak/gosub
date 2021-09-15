package parser

import (
	"bytes"
	"fmt"
)

type TExpr interface {
	String() string
}

type T_Int struct {
	X int
}

func (o *T_Int) String() string {
	return fmt.Sprintf("Int(%d)", o.X)
}

type T_String struct {
	X string
}

func (o *T_String) String() string {
	return fmt.Sprintf("String(%q)", o.X)
}

type T_Char struct {
	X byte
}

func (o *T_Char) String() string {
	return fmt.Sprintf("Char(%q)", string(o.X))
}

type T_Ident struct {
	X string
}

func (o *T_Ident) String() string {
	return fmt.Sprintf("Ident(%s)", o.X)
}

type T_BinOp struct {
	A  TExpr
	Op string
	B  TExpr
}

func (o *T_BinOp) String() string {
	return fmt.Sprintf("Bin(%v %q %v)", o.A, o.Op, o.B)
}

type T_List struct {
	V []TExpr
}

func (o *T_List) String() string {
	var buf bytes.Buffer
	buf.WriteString("List(")
	for i, e := range o.V {
		buf.WriteString(fmt.Sprintf("[%d] %v ", i, e))
	}
	buf.WriteString(")")
	return buf.String()
}

type TStmt interface {
	String() string
}

type T_Assign struct {
	A  TExpr
	Op string
	B  TExpr
}

func (o *T_Assign) String() string {
	return fmt.Sprintf("\nAssign(%v %q %v)\n", o.A, o.Op, o.B)
}

////////////////////////

type TDef interface {
}

type D_Package struct {
	Name string
}
type D_Import struct {
	Name string
	Path string
}
type D_Const struct {
	Name  string
	Value interface{}
}
type D_Var struct {
	Name string
	Type string
}
type D_Type struct {
	Name string
	Type string
}
type D_Func struct {
	Name    string
	Type    string
	Args    []string
	Results []string
	Body    []TStmt
}
