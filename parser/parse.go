package parser

import (
	"io"
	"log"
)

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
	Meths   []*GDef
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
			// Only {} should follow.
			o.Next()
			o.TakePunc("{")
			o.TakePunc("}")
			return &InterfaceTX{nil}
		}
		if o.Word == "struct" {
			panic("Keyword `struct` not expected, except after global `type`")
		}
		z := &IdentX{o.Word, o.CMod}
		o.Next()
		return z
	}
	if o.Kind == L_Punc {
		if o.Word == "-" {
			o.Next()
			x := o.ParsePrim()
			return &BinOpX{&LitIntX{0}, "-", x}
		}
		if o.Word == "*" {
			o.Next()
			elemX := o.ParseType()
			return &PointerTX{o.ExprToNameTX(elemX)}
		}
		if o.Word == "[" {
			o.Next()
			if o.Word != "]" {
				panic(F("for slice type, after [ expected ], got %v", o.Word))
			}
			o.Next()
			elemX := o.ParseType()
			return &SliceTX{o.ExprToNameTX(elemX)}
		}
		if o.Word == "&" {
			o.Next()
			typeX := o.ParseType()
			return o.ParseConstructor(typeX)
		}
	}
	panic("bad ParsePrim")
}

func (o *Parser) ParseConstructor(typeX Expr) Expr {
	o.TakePunc("{")
	ctor := &ConstructorX{
		typeX: typeX,
	}
LOOP:
	for {
		switch o.Kind {
		case L_Ident:
			fieldName := o.TakeIdent()
			o.TakePunc(":")
			fieldX := o.ParseExpr()
			ctor.inits = append(ctor.inits, NameAndExpr{fieldName, fieldX, o.CMod})
		case L_EOL:
			o.Next()
		case L_Punc:
			if o.Word == "," {
				o.Next()
			} else if o.Word == "}" {
				break LOOP
			} else {
				panic(F("Expected identifier or `}` but got %q", o.Word))
			}
		default:
			panic(F("Expected identifier or `,` or `}` but got %q", o.Word))
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
			if o.Word == ":" {
				o.Next()
				if o.Word == "]" {
					a = &SubSliceX{a, nil, nil}
				} else {
					b := o.ParseExpr()
					a = &SubSliceX{a, nil, b}
				}
			} else {
				sub := o.ParseExpr()
				if o.Word == ":" {
					o.Next()
					if o.Word == "]" {
						a = &SubSliceX{a, sub, nil}
					} else {
						b := o.ParseExpr()
						a = &SubSliceX{a, sub, b}
					}
				} else {
					a = &SubX{a, sub}
				}
			}
			o.TakePunc("]")
		case ".":
			o.TakePunc(".")
			if o.Word == "(" {
				o.TakePunc("(")
				b := o.ParseType()
				o.TakePunc(")")
				return &RuntimeCastX{a, b}
			} else {
				member := o.TakeIdent()
				a = &DotX{a, member}
			}
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

func (o *Parser) ParseType() Expr {
	return o.ParseExpr() // ParseType is now ParseExpr.
}

func (o *Parser) ParseStructType(name string) *StructRecX {
	o.TakePunc("{")
	rec := &StructRecX{
		name: name,
	}
LOOP:
	for {
		switch o.Kind {
		case L_Ident:
			fieldName := o.TakeIdent()
			fieldType := o.ParseType()
			rec.Fields = append(rec.Fields, NameTX{fieldName, fieldType, o.CMod})
		case L_EOL:
			o.Next()
		case L_Punc:
			if o.Word == "}" {
				break LOOP
			}
			panic(F("Expected identifier or `}` but got %q", o.Word))
		default:
			panic(F("Expected identifier or `}` but got %q", o.Word))
		}
	}
	o.TakePunc("}")
	return rec
}
func (o *Parser) ParseInterfaceType(name string) *InterfaceRecX {
	o.TakePunc("{")
	rec := &InterfaceRecX{
		name: name,
	}
LOOP:
	for {
		switch o.Kind {
		case L_Ident:
			fieldName := o.TakeIdent()
			sigx := &FuncRecX{}
			o.ParseFunctionSignature(sigx)
			// RegisterFuncRec(sigx)
			fieldType := &FunctionTX{sigx}
			rec.Meths = append(rec.Meths, NameTX{fieldName, fieldType, o.CMod})
		case L_EOL:
			o.Next()
		case L_Punc:
			if o.Word == "}" {
				break LOOP
			}
			panic(F("Expected identifier or `}` but got %q", o.Word))
		default:
			panic(F("Expected identifier or `}` but got %q", o.Word))
		}
	}
	o.TakePunc("}")
	return rec
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
	isRange := false
	a := o.ParseList()
	op := o.Word
	if op == "=" || op == ":=" {
		o.Next()
		if o.Word == "range" {
			o.Next()
			isRange = true
		}
		b := o.ParseList()
		return &AssignS{a, op, b, isRange}
	} else if op == "++" {
		o.Next()
		return &AssignS{a, op, nil, isRange}
	} else if op == "--" {
		o.Next()
		return &AssignS{a, op, nil, isRange}
	} else if o.Kind == L_EOL || o.Word == "{" {
		// Result not assigned.
		return &AssignS{nil, "", a, isRange}
	} else {
		panic(F("Unexpected token after statement: %v", o.Word))
	}
}

func (o *Parser) TakePunc(s string) {
	if o.Kind != L_Punc || s != o.Word {
		panic(F("expected %q, got (%d) %q", s, o.Kind, o.Word))
	}
	o.Next()
}

func (o *Parser) TakeIdent() string {
	if o.Kind != L_Ident {
		panic(F("expected Ident, got (%d) %q", o.Kind, o.Word))
	}
	w := o.Word
	o.Next()
	return w
}

func (o *Parser) TakeEOL() {
	if o.Kind != L_EOL {
		panic(F("expected EOL, got (%d) %q", o.Kind, o.Word))
	}
	o.Next()
}

func (o *Parser) ParseStmt(b *Block) Stmt {
	switch o.Word {
	case "var":
		o.Next()
		varIdent := o.TakeIdent()
		varType := o.ParseType()
		return &VarStmt{varIdent, varType}
	case "if":
		o.Next()
		pred := o.ParseExpr()
		yes := o.ParseBlock()
		var no *Block
		if o.Word == "else" {
			o.Next()
			if o.Word == "if" {
				noStmt := o.ParseStmt(b)
				no = &Block{
					why:      "Parser:elseIf", // TODO why in Parser?
					locals:   make(map[string]*GDef),
					stmts:    []Stmt{noStmt},
					parent:   b,
					compiler: b.compiler,
				}
			} else {
				no = o.ParseBlock()
			}
		}
		return &IfS{pred, yes, no}
	case "for":
		o.Next()

		forscope := &Block{
			why:      "Parser:forscope", // TODO why in Parser?
			locals:   make(map[string]*GDef),
			parent:   b,
			compiler: b.compiler,
		}

		var one Stmt
		if o.Word != "{" {
			one = o.ParseStmt(forscope)
		}
		var two Expr
		var three Stmt
		if o.Word != "{" {
			o.TakePunc(";")
			two = o.ParseExpr()
			o.TakePunc(";")
			three = o.ParseStmt(forscope)
		}
		body := o.ParseBlock()

		if two == nil {
			switch t := one.(type) {
			case nil:
				return &WhileS{nil, nil, nil, body} // for ever
			case (*AssignS):

				if t.IsRange {
					if len(t.A) == 1 && len(t.B) == 1 {
						return &ForS{t.A[0], nil, t.B[0], body}
					} else if len(t.A) == 2 && len(t.B) == 1 {
						return &ForS{t.A[0], t.A[1], t.B[0], body}
					} else {
						panic(F("bad range assignment after `for`; got %v", one))
					}
				} else {
					if len(t.A) == 0 && len(t.B) == 1 {
						pred := t.B[0]
						return &WhileS{nil, pred, nil, body}
					} else {
						panic(F("expected predicate expr after `for`; got %v", one))
					}
				}
			}
		}
		return &WhileS{one, two, three, body}

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
				bare := o.ParseBareBlock()
				sws.Cases = append(sws.Cases, &Case{matches, bare})
			case "default":
				o.TakePunc(":")
				bare := o.ParseBareBlock()
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
func (o *Parser) ParseBlock() *Block {
	o.TakePunc("{")
	b := o.ParseBareBlock()
	o.TakePunc("}")
	return b
}
func (o *Parser) ParseBareBlock() *Block {
	b := &Block{
		locals: make(map[string]*GDef),
	}
	for o.Word != "}" && o.Word != "case" && o.Word != "default" {
		if o.Kind == L_EOL {
			o.TakeEOL()
		} else {
			stmt := o.ParseStmt(b)
			if stmt != nil {
				b.stmts = append(b.stmts, stmt)
			}
			o.TakeEOL()
		}
	}
	return b
}

func (o *Parser) ParseFunctionSignature(fn *FuncRecX) {
	o.TakePunc("(")
	for o.Word != ")" {
		s := o.TakeIdent()
		if o.Word == ".." {
			o.TakePunc("..")
			o.TakePunc(".")
			fn.HasDotDotDot = true
		}
		t := o.ParseType()
		Say(t)
		fn.Ins = append(fn.Ins, NameTX{s, t, o.CMod})
		if o.Word == "," {
			o.TakePunc(",")
		} else if o.Word != ")" {
			panic(F("expected `,` or `)` but got %q", o.Word))
		}
		if fn.HasDotDotDot {
			if o.Word != ")" {
				panic(F("Expected `)` after ellipsis arg, but got `%v`", o.Word))
			}
			numIns := len(fn.Ins)
			last := fn.Ins[numIns-1]
			Say(fn.Ins[numIns-1])
			elementNat := NameTX{"", last.Expr, o.CMod}
			wrapWithSliceTX := NameTX{last.name, &SliceTX{E: elementNat}, o.CMod}
			fn.Ins[numIns-1] = wrapWithSliceTX //- &SliceTV{BaseTV{}, fn.Ins[numIns-1].TV}
			Say(fn.Ins[numIns-1])
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
				fn.Outs = append(fn.Outs, NameTX{s, t, o.CMod})
				if o.Word == "," {
					o.TakePunc(",")
				} else if o.Word != ")" {
					panic(F("expected `,` or `)` but got %q", o.Word))
				}
			}
			o.TakePunc(")")
		} else {
			t := o.ParseType()
			fn.Outs = append(fn.Outs, NameTX{"", t, o.CMod})
		}
	}
}
func (o *Parser) ParseFunc(receiver *NameTX) *FuncRecX {
	fn := &FuncRecX{}
	if receiver != nil {
		fn.Ins = append(fn.Ins, *receiver)
		fn.IsMethod = true
	}
	o.ParseFunctionSignature(fn)
	// RegisterFuncRec(fn)
	if o.Kind != L_EOL {
		fn.Body = o.ParseBlock()
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
					log.Printf("WARNING: Expected package %s, got %s", cm.Package, w)
				}
				o.Package = w
			case "import":
				if o.Kind != L_String {
					panic(F("after import, expected string, got %v", o.Word))
				}
				w := o.Word
				o.Next()
				gd := &GDef{
					name:   w,
					typeof: ImportTO,
				}
				o.Imports = append(o.Imports, gd)
				o.ImportsMap[w] = gd
			case "const":
				w := o.TakeIdent()
				var tx Expr
				if o.Word != "=" {
					tx = o.ParseType()
				}
				o.TakePunc("=")
				x := o.ParseExpr()
				gd := &GDef{
					Package: o.Package,
					name:    w,
					initx:   x,
					typex:   tx,
				}
				o.Consts = append(o.Consts, gd)
				o.ConstsMap[w] = gd
			case "var":
				w := o.TakeIdent()
				var tx Expr
				if o.Word != "=" {
					tx = o.ParseType()
				}
				var i Expr
				if o.Word == "=" {
					o.Next()
					i = o.ParseExpr()
				}
				gd := &GDef{
					Package: o.Package,
					name:    w,
					typex:   tx,
					initx:   i,
				}
				o.Vars = append(o.Vars, gd)
				o.VarsMap[w] = gd
			case "type":
				w := o.TakeIdent()
				var tx Expr
				if o.Word == "interface" {
					o.Next()
					tx = &InterfaceTX{o.ParseInterfaceType(w)}
				} else if o.Word == "struct" {
					o.Next()
					tx = &StructTX{o.ParseStructType(w)}
				} else if o.Word == "func" {
					panic("todo")
				} else {
					tx = o.ParseType()
				}
				gd := &GDef{
					Package: o.Package,
					name:    w,
					initx:   tx,
					typeof:  TypeTO,
				}
				o.Types = append(o.Types, gd)
				o.TypesMap[w] = gd
			case "func":
				var receiver *NameTX
				if o.Word == "(" {
					// Distinguished Receiver:
					o.Next()
					rName := "_"

					if o.Kind == L_Ident {
						// The receiver is named.
						rName = o.Word
						o.Next()
					}

					if o.Word != "*" {
						panic(F("Got %q but expected '*': Method receiver type must be pointer to struct", o.Word))
					}
					rType := o.ParseExpr()
					o.TakePunc(")")
					receiver = &NameTX{rName, rType, o.CMod}
				}
				name := o.TakeIdent()
				fn := o.ParseFunc(receiver)
				gd := &GDef{
					Package: o.Package,
					name:    name,
					initx:   &FunctionX{fn},
				}
				if receiver == nil {
					o.Funcs = append(o.Funcs, gd)
					o.FuncsMap[name] = gd
				} else {
					// Receiver TypeValue is not resolved yet,
					// so save it for later.
					o.Meths = append(o.Meths, gd)
				}
			default:
				panic(F("Expected top level decl, got %q", d))
			}
			o.TakeEOL()
		case L_EOL:
			o.TakeEOL()
			continue LOOP
		case L_EOF:
			break LOOP
		default:
			panic(F("expected toplevel decl; got (%d) %q", o.Kind, o.Word))
		}
	}
}

//////////////////////////////////////////////////
//////////////////////////////////////////////////

func (o *Parser) ExprToNameTX(e Expr) NameTX {
	return NameTX{Expr: e, Mod: o.CMod}
}
