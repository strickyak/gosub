package parser

import (
	"io"
)

type Parser struct {
	*Lex
}

func NewParser(r io.Reader, filename string) *Parser {
	return &Parser{
		Lex: NewLex(r, filename),
	}
}

func (o *Parser) ParsePrim() TExpr {
	if o.Kind == L_Int {
		z := &T_Int{o.Num}
		o.Next()
		return z
	}
	if o.Kind == L_String {
		z := &T_String{o.Word}
		o.Next()
		return z
	}
	if o.Kind == L_Char {
		z := &T_Char{o.Word[0]}
		o.Next()
		return z
	}
	if o.Kind == L_Ident {
		z := &T_Ident{o.Word}
		o.Next()
		return z
	}
	panic("bad ParsePrim")
}

func (o *Parser) ParseProduct() TExpr {
	a := o.ParsePrim()
	op := o.Word
	for op == "*" || op == "/" || op == "%" || op == "<<" || op == ">>" || op == "&" || op == "&^" {
		o.Next()
		b := o.ParseProduct()
		a = &T_BinOp{a, op, b}
		op = o.Word
	}
	return a
}

func (o *Parser) ParseSum() TExpr {
	a := o.ParseProduct()
	op := o.Word
	for op == "+" || op == "-" || op == "|" || op == "^" {
		o.Next()
		b := o.ParseSum()
		a = &T_BinOp{a, op, b}
		op = o.Word
	}
	return a
}

func (o *Parser) ParseRelational() TExpr {
	a := o.ParseSum()
	op := o.Word
	for o.Word == "==" || o.Word == "!=" || o.Word == "<" || o.Word == ">" || o.Word == "<=" || o.Word == ">=" {
		o.Next()
		b := o.ParseRelational()
		a = &T_BinOp{a, op, b}
		op = o.Word
	}
	return a
}

func (o *Parser) ParseAnd() TExpr {
	a := o.ParseRelational()
	for o.Word == "&&" {
		o.Next()
		b := o.ParseAnd()
		a = &T_BinOp{a, "&&", b}
	}
	return a
}

func (o *Parser) ParseOr() TExpr {
	a := o.ParseAnd()
	for o.Word == "||" {
		o.Next()
		b := o.ParseOr()
		a = &T_BinOp{a, "||", b}
	}
	return a
}

func (o *Parser) ParseList() TExpr {
	a := o.ParseOr()
	if o.Word == "," {
		v := []TExpr{a}
		for o.Word == "," {
			o.Next()
			b := o.ParseOr()
			v = append(v, b)
		}
		return &T_List{v}
	}
	return a
}

func (o *Parser) ParseAssignment() TStmt {
	a := o.ParseList()
	op := o.Word
	if op == "=" || len(op) == 2 && op[1] == '=' {
		o.Next()
		b := o.ParseList()
		return &T_Assign{a, op, b}
	}
	return &T_Assign{nil, "", a}
}
