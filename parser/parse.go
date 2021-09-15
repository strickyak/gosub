package parser

import (
	"io"
	"log"
)

type Parser struct {
	*Lex
	Package *D_Package
	Imports map[string]*D_Import
	Consts  map[string]*D_Const
	Vars    map[string]*D_Var
	Types   map[string]*D_Type
	Funcs   map[string]*D_Func
}

func NewParser(r io.Reader, filename string) *Parser {
	return &Parser{
		Lex:     NewLex(r, filename),
		Imports: make(map[string]*D_Import),
		Consts:  make(map[string]*D_Const),
		Vars:    make(map[string]*D_Var),
		Types:   make(map[string]*D_Type),
		Funcs:   make(map[string]*D_Func),
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

func (o *Parser) TakeIdent() string {
	if o.Kind != L_Ident {
		log.Panicf("expected Ident, got (%d) %q", o.Kind, o.Word)
	}
	s := o.Word
	o.Next()
	return s
}

func (o *Parser) TakeEOL() {
	if o.Kind != L_EOL {
		log.Panicf("expected EOL, got (%d) %q", o.Kind, o.Word)
	}
	o.Next()
}

func (o *Parser) ParseTop() {
LOOP:
	for {
		switch o.Kind {
		case L_Ident:
			d := o.TakeIdent()
			switch d {
			case "package":
				s := o.TakeIdent()
				o.Package = &D_Package{Name: s}
			case "import":
				s := o.TakeIdent()
				o.Imports[s] = &D_Import{Name: s}
			case "const":
				s := o.TakeIdent()
				o.Consts[s] = &D_Const{Name: s}
			case "var":
				s := o.TakeIdent()
				o.Vars[s] = &D_Var{Name: s}
			case "type":
				s := o.TakeIdent()
				o.Types[s] = &D_Type{Name: s}
			case "func":
				s := o.TakeIdent()
				o.Funcs[s] = &D_Func{Name: s}
			}
			o.TakeEOL()
		case L_EOL:
			o.TakeEOL()
			continue LOOP
		case L_EOF:
			break LOOP
		default:
			log.Panicf("expected toplevel decl; got (%d) %q", o.Kind, o.Word)
		}
	}
}
