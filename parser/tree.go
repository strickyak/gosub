package parser

import (
	"bytes"
	"fmt"
)

type Tree interface {
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
	A  Tree
	Op string
	B  Tree
}

func (o *T_BinOp) String() string {
	return fmt.Sprintf("Bin(%v %s %v)", o.A, o.Op, o.B)
}

type T_List struct {
	V []Tree
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
