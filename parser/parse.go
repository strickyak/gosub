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

func (o *Parser) ParsePrim() Tree {
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

func (o *Parser) ParseProduct() Tree {
	a := o.ParsePrim()
	for o.Word == "*" || o.Word == "/" || o.Word == "%" || o.Word == "<<" || o.Word == ">>" || o.Word == "&" || o.Word == "&^" {
		o.Next()
		b := o.ParseProduct()
		a = &T_BinOp{a, o.Word, b}
	}
	return a
}

func (o *Parser) ParseSum() Tree {
	a := o.ParseProduct()
	for o.Word == "+" || o.Word == "-" || o.Word == "|" || o.Word == "^" {
		o.Next()
		b := o.ParseSum()
		a = &T_BinOp{a, o.Word, b}
	}
	return a
}

func (o *Parser) ParseRelational() Tree {
	a := o.ParseSum()
	for o.Word == "==" || o.Word == "!=" || o.Word == "<" || o.Word == ">" || o.Word == "<=" || o.Word == ">=" {
		o.Next()
		b := o.ParseRelational()
		a = &T_BinOp{a, o.Word, b}
	}
	return a
}

func (o *Parser) ParseAnd() Tree {
	a := o.ParseRelational()
	for o.Word == "&&" {
		o.Next()
		b := o.ParseAnd()
		a = &T_BinOp{a, o.Word, b}
	}
	return a
}

func (o *Parser) ParseOr() Tree {
	a := o.ParseAnd()
	for o.Word == "||" {
		o.Next()
		b := o.ParseOr()
		a = &T_BinOp{a, o.Word, b}
	}
	return a
}

func (o *Parser) ParseList() Tree {
	a := o.ParseOr()
	if o.Word == "," {
		v := []Tree{a}
		for o.Word == "," {
			o.Next()
			b := o.ParseOr()
			v = append(v, b)
		}
		return &T_List{v}
	}
	return a
}

func (o *Parser) ParseAssignment() Tree {
	a := o.ParseList()
	if o.Word == "=" {
		o.Next()
		b := o.ParseList()
		a = &T_BinOp{a, o.Word, b}
	}
	return a
}
