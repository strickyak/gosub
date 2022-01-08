package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"strings"
)

func FindTypeByName(list []NameTV, name string) (TypeValue, bool) {
	log.Printf("Finding %q in list of len=%d", name, len(list))
	for _, ntv := range list {
		stuff := "?ftbn?"
		switch {
		case ntv.TV != nil:
			stuff = ntv.TV.String()
			/*
				case ntv.Expr != nil:
					stuff = ntv.Expr.String()
			*/
		}

		log.Printf("?find %q? { %q ; %s }", name, ntv.Name, stuff)
		if ntv.Name == name {
			log.Printf("YES")
			return ntv.TV, true
		}
	}
	log.Printf("NO")
	return nil, false
}

///////////

type Options struct {
	LibDir      string
	SkipBuiltin bool
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
	VisitFunction(*FunctionX) Value
}

type Expr interface {
	String() string
	VisitExpr(ExprVisitor) Value
}

/*
type Lval interface {
	String() string
	VisitLVal(LvalVisitor) Value
}
*/

//

type PointerTX struct {
	E NameTX
}
type SliceTX struct {
	E NameTX
}
type MapTX struct {
	K NameTX
	V NameTX
}
type StructTX struct {
	StructRecX *StructRecX
}
type InterfaceTX struct {
	InterfaceRecX *InterfaceRecX // nil for interface{}
}
type FunctionTX struct {
	FuncRecX *FuncRecX
}

func (n NameTX) String() string {
	return Format("NaT{%q~%s@%s}", n.Name, n.Expr.String(), n.Mod.Package)
}
func (n NameTV) String() string {
	return Format("NaT{%q~~%s}", n.Name, n.TV)
}
func (r *StructRecX) String() string {
	return Format("structX %s", r.Name)
}
func (r *InterfaceRecX) String() string {
	return Format("interfaceX %s", r.Name)
}
func (r *FuncRecX) String() string {
	s := "funcX("
	if r.IsMethod {
		s += "[meth]"
	}
	for _, e := range r.Ins {
		s += F("%s, ", e)
	}
	s += ")("
	for _, e := range r.Outs {
		s += F("%s, ", e)
	}
	if r.HasDotDotDot {
		s += "..."
	}
	s += ")"

	if r.Body != nil {
		s += "{BODY}"
	}
	return s
}

func (r *StructRec) String() string {
	return Format("struct %s", r.Name)
}
func (r *InterfaceRec) String() string {
	return Format("interface %s", r.Name)
}
func (r *FuncRec) String() string {
	s := "func("
	if r.IsMethod {
		s += "[meth]"
	}
	for _, e := range r.Ins {
		s += F("%s, ", e)
	}
	s += ")("
	for _, e := range r.Outs {
		s += F("%s, ", e)
	}
	if r.HasDotDotDot {
		s += "..."
	}
	s += ")"

	if r.FuncRecX.Body != nil {
		s += "{BODY}"
	}
	return s
}

func (o *PointerTX) String() string { return Format("PointerTX(%v)", o.E) }
func (o *SliceTX) String() string   { println(o.E.String()); return Format("SliceTX(%v)", o.E) }
func (o *MapTX) String() string     { return Format("MapTX(%v=>%v)", o.K, o.V) }
func (o *StructTX) String() string {
	return "TX:" + o.StructRecX.String()
}
func (o *InterfaceTX) String() string {
	return "TX:" + o.InterfaceRecX.String()
}
func (o *FunctionTX) String() string { return Format("FunctionTX(%v)", o.FuncRecX) }

func CompileTX(v ExprVisitor, x NameTX, where Expr) NameTV {
	return NameTV{
		Name: x.Name,
		TV:   x.Mod.VisitTypeExpr(x.Expr),
	}
}

func (o *PointerTX) VisitExpr(v ExprVisitor) Value {
	z := &PointerTV{
		E: CompileTX(v, o.E, o).TV,
	}
	return &TypeVal{z}
}
func (o *SliceTX) VisitExpr(v ExprVisitor) Value {
	z := &SliceTV{
		E: CompileTX(v, o.E, o).TV,
	}
	Say(z)
	return &TypeVal{z}
}
func (o *MapTX) VisitExpr(v ExprVisitor) Value {
	z := &MapTV{
		K: CompileTX(v, o.K, o).TV,
		V: CompileTX(v, o.V, o).TV,
	}
	return &TypeVal{z}
}

func (o *StructTX) VisitExpr(v ExprVisitor) Value {
	p := o.StructRecX
	z := &StructRec{
		Name:   p.Name,
		Fields: make([]NameTV, len(p.Fields)),
		Meths:  make([]NameTV, len(p.Meths)),
	}
	for i, e := range p.Fields {
		z.Fields[i] = CompileTX(v, e, o)
	}
	for i, e := range p.Meths {
		z.Meths[i] = CompileTX(v, e, o)
	}
	Say(z)
	return &TypeVal{&StructTV{z}}
}
func (o *InterfaceTX) VisitExpr(v ExprVisitor) Value {
	p := o.InterfaceRecX
	if p == nil {
		// nil means this is _any_ "interface empty"
		return &TypeVal{AnyTO}
	}

	numMeths := len(p.Meths)
	assert(numMeths > 0)
	z := &InterfaceRec{
		Name:  p.Name,
		Meths: make([]NameTV, numMeths),
	}
	for i, e := range p.Meths {
		Say(i, e)
		z.Meths[i] = CompileTX(v, e, o)
		Say(z.Meths[i])
	}
	Say(z)
	return &TypeVal{&InterfaceTV{z}}
}
func (o *FunctionTX) VisitExpr(v ExprVisitor) Value {
	z := o.FuncRecX.VisitFuncRecX(v)
	return &TypeVal{&FunctionTV{z}}
}
func (x *FuncRecX) VisitFuncRecX(v ExprVisitor) *FuncRec {
	z := &FuncRec{
		Ins:          make([]NameTV, len(x.Ins)),
		Outs:         make([]NameTV, len(x.Outs)),
		HasDotDotDot: x.HasDotDotDot,
		IsMethod:     x.IsMethod,
		FuncRecX:     x,
	}
	for i, e := range x.Ins {
		L("250: Ins %d name %q expr %#v", i, e.Name, e.Expr)
		z.Ins[i] = NameTV{e.Name, e.Mod.VisitTypeExpr(e.Expr)}
	}
	for i, e := range x.Outs {
		L("250: Outs %d name %q expr %#v", i, e.Name, e.Expr)
		z.Outs[i] = NameTV{e.Name, e.Mod.VisitTypeExpr(e.Expr)}
	}
	return z
}

type TypeValue interface {
	String() string
	// Value
	// Intlike() bool // only on PrimTV
	CType() string
	// TypeOfHandle() (z string, ok bool)
	Assign(c string, typ TypeValue) (z string, ok bool)
	Cast(c string, typ TypeValue) (z string, ok bool)
	Equal(typ TypeValue) bool
}

type PrimTV struct {
	Name string
}
type TypeTV struct {
	Name string
}
type PointerTV struct {
	E TypeValue
}
type SliceTV struct {
	E TypeValue
}
type DotDotDotSliceTV struct {
	E TypeValue
}
type MapTV struct {
	K TypeValue
	V TypeValue
}
type StructTV struct {
	StructRec *StructRec
}
type InterfaceTV struct {
	InterfaceRec *InterfaceRec
}
type FunctionTV struct {
	FuncRec *FuncRec
}

type MultiTV struct {
	Multi []NameTV
}

const kResolveTooDeep = 16

// Type values have type TypeTV (the metatype).
func (tv *PrimTV) Type() TypeValue    { return &TypeTV{} }
func (tv *TypeTV) Type() TypeValue    { return &TypeTV{} }
func (tv *PointerTV) Type() TypeValue { return &TypeTV{} }
func (tv *SliceTV) Type() TypeValue   { return &TypeTV{} }
func (tv *MapTV) Type() TypeValue     { return &TypeTV{} }

func (tv *StructTV) Type() TypeValue    { return &TypeTV{} }
func (tv *InterfaceTV) Type() TypeValue { return &TypeTV{} }
func (tv *FunctionTV) Type() TypeValue  { return &TypeTV{} }

func (tv *MultiTV) Type() TypeValue { return &TypeTV{} }

func ToC(v Value) string {
	if v == nil {
		return "<nil>"
	}
	return v.ToC()
}

func (tv *PrimTV) ToC() string {
	return strings.Title(tv.Name)
}
func (tv *TypeTV) ToC() string {
	return Format("ZType(%s)", tv.Name)
}
func (tv *PointerTV) ToC() string {
	return Format("ZPointer(%s)", tv.E)
}
func (tv *SliceTV) ToC() string {
	return Format("ZSlice(%s)", tv.E)
}
func (tv *MapTV) ToC() string {
	return Format("ZMap(%s, %s)", tv.K, tv.V)
}

func (tv *StructTV) ToC() string {
	return Format("ZStruct(%s)", tv.StructRec.Name)
}
func (tv *InterfaceTV) ToC() string {
	return Format("ZInterface(%s/%d)", tv.InterfaceRec.Name, len(tv.InterfaceRec.Meths))
}
func (tv *FunctionTV) ToC() string {
	return Format("ZFunction(%s)", tv.FuncRec.SignatureStr("(*)"))
}

func (tv *MultiTV) ToC() string {
	return Format("ZMulti(...)")
}

func (o *PrimTV) Intlike() bool {
	switch o.Name {
	case "byte", "int", "uint":
		return true
	}
	return false
}

func (o *FunctionTV) Equal(typ TypeValue) bool {
	switch t := typ.(type) {
	case *FunctionTV:
		return reflect.DeepEqual(o, t)
	}
	return false
}
func (o *TypeTV) Equal(typ TypeValue) bool {
	switch t := typ.(type) {
	case *TypeTV:
		return o.Name == t.Name
	}
	return false
}
func (o *PrimTV) Equal(typ TypeValue) bool {
	switch t := typ.(type) {
	case *PrimTV:
		return o.Name == t.Name
	}
	return false
}
func (o *SliceTV) Equal(typ TypeValue) bool {
	switch t := typ.(type) {
	case *SliceTV:
		return o.E.Equal(t.E)
	}
	return false
}
func (o *PointerTV) Equal(typ TypeValue) bool {
	switch t := typ.(type) {
	case *PointerTV:
		return o.E.Equal(t.E)
	}
	return false
}
func (o *MapTV) Equal(typ TypeValue) bool {
	switch t := typ.(type) {
	case *MapTV:
		return o.K.Equal(t.K) && o.V.Equal(t.V)
	}
	return false
}
func (o *StructTV) Equal(typ TypeValue) bool {
	switch t := typ.(type) {
	case *StructTV:
		return o.StructRec.Name == t.StructRec.Name
	}
	return false
}
func (o *InterfaceTV) Equal(typ TypeValue) bool {
	switch t := typ.(type) {
	case *InterfaceTV:
		return o.InterfaceRec.Name == t.InterfaceRec.Name
	}
	return false
}
func (o *MultiTV) Equal(typ TypeValue) bool {
	panic("cannot compare MultiTV")
}

func (o *PointerTV) TypeOfHandle() (z string, ok bool) {
	if st, ok := o.E.(*StructTV); ok {
		return st.StructRec.Name, true
	}
	return "", false
}

func (o *PrimTV) CType() string      { return "P_" + o.Name }
func (o *SliceTV) CType() string     { return F("Slice_(%s)", o.E.CType()) }
func (o *MapTV) CType() string       { return F("Map_(%s,%s)", o.K.CType(), o.V.CType()) }
func (o *StructTV) CType() string    { return F("Struct_(%s)", o.StructRec.Name) }
func (o *PointerTV) CType() string   { return F("Pointer_(%s)", o.E.CType()) }
func (o *InterfaceTV) CType() string { return F("Interface_(%s)", o.InterfaceRec.Name) }
func (o *TypeTV) CType() string      { return "Type" }
func (o *MultiTV) CType() string     { return "Multi" }

func (o *FunctionTV) CType() string { return o.FuncRec.PtrTypedef }

func (o *TypeTV) Assign(c string, typ TypeValue) (z string, ok bool) {
	panic("cannot Assign to _type_")
}
func (o *FunctionTV) Assign(c string, typ TypeValue) (z string, ok bool) {
	panic("cannot Assign to func")
}
func (o *MultiTV) Assign(c string, typ TypeValue) (z string, ok bool) {
	panic("cannot Assign to Multi")
}

func (o *PrimTV) Assign(c string, typ TypeValue) (z string, ok bool) {
	if o.Equal(typ) {
		return c, true
	} else {
		return "", false
	}
}
func (o *SliceTV) Assign(c string, typ TypeValue) (z string, ok bool) {
	if o.Equal(typ) {
		return c, true
	} else {
		return "", false
	}
}
func (o *MapTV) Assign(c string, typ TypeValue) (z string, ok bool) {
	if o.Equal(typ) {
		return c, true
	} else {
		return "", false
	}
}
func (o *StructTV) Assign(c string, typ TypeValue) (z string, ok bool) {
	if o.Equal(typ) {
		return c, true
	} else {
		return "", false
	}
}
func (o *PointerTV) Assign(c string, typ TypeValue) (z string, ok bool) {
	if o.Equal(typ) {
		return c, true
	} else {
		return "", false
	}
}
func (o *InterfaceTV) Assign(c string, typ TypeValue) (z string, ok bool) {
	if ptr, ok := typ.(*PointerTV); ok {
		if _, ok := ptr.TypeOfHandle(); ok {
			// TODO: check compat
			return Format("HandleToInterface(%s)", c), true
		}
	}
	if _, ok := typ.(*InterfaceTV); ok {
		return c, true
	}
	return "", false
}

func (o *SliceTV) Cast(c string, typ TypeValue) (z string, ok bool) {
	if o.Equal(typ) {
		return c, true
	} else {
		return "", false
	}
}
func (o *MapTV) Cast(c string, typ TypeValue) (z string, ok bool) {
	if o.Equal(typ) {
		return c, true
	} else {
		return "", false
	}
}
func (o *StructTV) Cast(c string, typ TypeValue) (z string, ok bool) {
	if o.Equal(typ) {
		return c, true
	} else {
		return "", false
	}
}
func (o *PointerTV) Cast(c string, typ TypeValue) (z string, ok bool) {
	if o.Equal(typ) {
		return c, true
	} else {
		return "", false
	}
}
func (o *InterfaceTV) Cast(c string, typ TypeValue) (z string, ok bool) {
	if o.Equal(typ) {
		return c, true
	} else {
		return "", false
	}
}

func (o *PrimTV) Cast(c string, typ TypeValue) (z string, ok bool) {
	if o.Equal(typ) {
		return c, true
	}
	if other, ok := typ.(*PrimTV); ok {
		if o.Intlike() && other.Intlike() {
			return Format("(P_%s)(%s)", o.Name, c), true
		}
	}
	return "", false
}
func (o *TypeTV) Cast(c string, typ TypeValue) (z string, ok bool) {
	panic("TypeTV cannot Cast")
}
func (o *FunctionTV) Cast(c string, typ TypeValue) (z string, ok bool) {
	panic("FunctionTV cannot Cast (yet)")
}
func (o *MultiTV) Cast(c string, typ TypeValue) (z string, ok bool) {
	panic("MultiTV cannot Cast (yet)")
}

func (tv *PrimTV) String() string    { return Format("PrimTV(%q)", tv.Name) }
func (tv *TypeTV) String() string    { return Format("TypeTV(%q)", tv.Name) }
func (tv *PointerTV) String() string { return Format("PointerTV(%v)", tv.E) }
func (tv *SliceTV) String() string   { return Format("SliceTV(%v)", tv.E) }
func (tv *MapTV) String() string     { return Format("MapTV(%v=>%v)", tv.K, tv.V) }

func (tv *StructTV) String() string {
	return Format("StructTV(%v)", tv.StructRec.Name)
}
func (tv *InterfaceTV) String() string {
	if tv.InterfaceRec.Meths == nil {
		return Format("TV:Any")
	} else {
		return Format("InterfaceTV(%s/%d)", tv.InterfaceRec.Name, len(tv.InterfaceRec.Meths))
	}
}
func (tv *FunctionTV) String() string { return Format("FunctionTV(%v)", tv.FuncRec) }

func (tv *MultiTV) String() string { return Format("MultiTV(%v)", tv.Multi) }

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
	X     string
	Outer *CMod // Outer scope where defined -- but the IdentX may or may not be global.
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

type ConstructorX struct {
	Name   string
	Fields []NameTX
}

func (o *ConstructorX) String() string {
	return fmt.Sprintf("Ctor(%q [[[%v]]])", o.Name, o.Fields)
}
func (o *ConstructorX) VisitExpr(v ExprVisitor) Value {
	return v.VisitConstructor(o)
}

type FunctionX struct {
	FuncRecX *FuncRecX
}

func (o *FunctionX) String() string {
	return fmt.Sprintf("FunctionX(%s)", o.FuncRecX)
}
func (o *FunctionX) VisitExpr(v ExprVisitor) Value {
	return v.VisitFunction(o)
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

// A callable view of a node in a parse tree,
// e.g. global func, lambda, bound method,
// ... any expr of Func kind.
type FuncRecX struct {
	Ins          []NameTX
	Outs         []NameTX
	HasDotDotDot bool
	IsMethod     bool
	Body         *Block
	PtrTypedef   string // global typedef of a pointer to this function type.
}

type FuncRec struct {
	Ins          []NameTV
	Outs         []NameTV
	HasDotDotDot bool
	IsMethod     bool
	PtrTypedef   string // global typedef of a pointer to this function type.
	FuncRecX     *FuncRecX
}

// Maps global typedef names to the C definition.
var FuncPtrTypedefs = make(map[string]string)

func RegisterFuncRec(fn *FuncRec) {
	// TODO: should this be lazy?
	name := Serial("funk")
	sigStr := fn.SignatureStr(Format("(*%s)", name))
	FuncPtrTypedefs[name] = sigStr
	fn.PtrTypedef = name
}

func (r *FuncRec) SignatureStr(daFunc string) string {
	var b bytes.Buffer
	if len(r.Outs) == 1 {
		P(&b, "%s ", r.Outs[0].TV.CType())
	} else {
		P(&b, "void ")
	}
	P(&b, "%s(", daFunc)
	for i, nat := range r.Ins {
		if i > 0 {
			b.WriteByte(',')
		}
		Say(nat)
		Say(nat.TV.CType())
		P(&b, "%s in_%s", nat.TV.CType(), SerialIfEmpty(nat.Name))
	}
	if len(r.Outs) != 1 {
		for i, nat := range r.Outs {
			if i > 0 {
				b.WriteByte(',')
			}
			L("out [%d]: %s", i, nat.TV.CType())
			P(&b, "%s *out_%s", nat.TV.CType(), SerialIfEmpty(nat.Name))
		}
	}
	b.WriteByte(')')
	sigStr := b.String()
	L("SignatureStr: %s", sigStr)
	return sigStr
}

type StructRecX struct {
	Name   string
	Fields []NameTX
	Meths  []NameTX
}

type InterfaceRecX struct {
	Name  string
	Meths []NameTX
}

type StructRec struct {
	Name   string
	Fields []NameTV
	Meths  []NameTV
}

type InterfaceRec struct {
	Name  string
	Meths []NameTV
}

type NameTX struct {
	Name string
	Expr Expr
	Mod  *CMod
}
type NameTV struct {
	Name string
	TV   TypeValue
}
type XXX_NameAndType struct {
	Name    string
	Expr    Expr
	TV      TypeValue
	Package string
}
type Block struct {
	debugName string
	locals    map[string]*GDef // not really G
	stmts     []Stmt
	parent    *Block
	compiler  *Compiler
}

func (b *Block) Find(name string) *GDef {
	if d, ok := b.locals[name]; ok {
		return d
	}
	if b.parent != nil {
		return b.parent.Find(name)
	} else {
		L("b.compiler= %#v", b.compiler)
		L("b.compiler.CMod= %#v", b.compiler.CMod)
		return b.compiler.CMod.Find(name)
	}
}

func (b *Block) VisitStmt(v StmtVisitor) {
	v.VisitBlock(b)
}

/*
const XX_BoolType = "a"
const XX_ByteType = "b"
const XX_UintType = "u"
const XX_IntType = "i"
const XX_ConstIntType = "c"
const XX_StringType = "s"
const XX_TypeType = "t"
const XX_ImportType = "@"
const XX_VoidType = "v"
const XX_ListType = "l"

const XX_BoolPre = 'a'
const XX_BytePre = 'b'
const XX_UintPre = 'u'
const XX_IntPre = 'i'
const XX_ConstIntPre = 'c'
const XX_StringPre = 's'
const XX_TypePre = 't'
const XX_ImportPre = '@'
const XX_VoidPre = 'v'
const XX_ListPre = 'l'

const XX_SlicePre = 'S'
const XX_DotDotDotSlicePre = 'E'
const XX_MapPre = 'M'
const XX_ChanPre = 'C'
const XX_FuncPre = 'F'
//const HandlePre = 'H'
const XX_StructPre = 'R'
const XX_InterfacePre = 'I'
const XX_PointerPre = 'P'

const XX_SliceForm = "S:%s"
const XX_DotDotDotSliceForm = "E:%s"
const XX_MapForm = "M:%s:%s"
const XX_ChanForm = "C:%s"
const XX_TypeForm = "t(%s)"
const XX_FuncForm = "F(%s;%s)"
const XX_StructForm = "R{%s}"
//const HandleForm = "H{%s}"
const XX_InterfaceForm = "I{%s}"
const XX_PointerForm = "P{%s}"
*/

// #################################################

// GlobalName creates the global name with kind and package.
func GlobalName(pkg string, name string) string {
	return Format("%s__%s", pkg, name)
}

// TODO ???
var BoolTO = &PrimTV{Name: "bool"}
var ByteTO = &PrimTV{Name: "byte"}
var ConstIntTO = &PrimTV{Name: "_const_int_"}
var IntTO = &PrimTV{Name: "int"}
var UintTO = &PrimTV{Name: "uint"}
var StringTO = &PrimTV{Name: "string"}
var TypeTO = &PrimTV{Name: "_type_"}
var ListTO = &PrimTV{Name: "_list_"} // i.e. Multi-Value with `,`
var VoidTO = &PrimTV{Name: "_void_"}
var ImportTO = &PrimTV{Name: "_import_"}
var AnyTO = &PrimTV{Name: "_any_"} // i.e. `interface{}`
var NilTO = &PrimTV{Name: "_nil_"}

// Mapping primative Go type names to Type Objects.
var PrimTypeObjList = []*PrimTV{
	BoolTO,
	ByteTO,
	ConstIntTO,
	IntTO,
	UintTO,
	StringTO,
	TypeTO,
	ListTO,
	VoidTO,
	ImportTO,
	AnyTO,
	NilTO,
}

type Value interface {
	String() string
	Type() TypeValue
	ToC() string
	ResolveAsTypeValue() (TypeValue, bool)
	ResolveAsValue() Value
}

func (o *GDef) String() string {
	return F("GDef[%s.%s=%s]", o.Name, o.Package, o.CName)
}
func (o *GDef) ToC() string {
	return o.CName
}
func (o *GDef) Type() TypeValue {
	return o.TV
}

func (o *GDef) ResolveAsValue() Value      { return o.Value }
func (o *CVal) ResolveAsValue() Value      { return o }
func (o *SubVal) ResolveAsValue() Value    { return o }
func (o *ImportVal) ResolveAsValue() Value { return o }
func (o *TypeVal) ResolveAsValue() Value   { return o }
func (o *NameVal) ResolveAsValue() Value {
	gd := o.dflt.Find(o.name)
	return gd.Value
}

func (o *GDef) ResolveAsTypeValue() (TypeValue, bool) {
	L("GDef.ResolveAsTypeValue < %s > Value: %T :: %v", o, o.Value, o.Value)
	if tval, ok := o.Value.(*TypeVal); ok {
		return tval.tv, true
	}
	return nil, false
}
func (o *CVal) ResolveAsTypeValue() (TypeValue, bool)      { return nil, false }
func (o *SubVal) ResolveAsTypeValue() (TypeValue, bool)    { return nil, false }
func (o *ImportVal) ResolveAsTypeValue() (TypeValue, bool) { return nil, false }
func (o *TypeVal) ResolveAsTypeValue() (TypeValue, bool)   { return o.tv, true }
func (o *NameVal) ResolveAsTypeValue() (TypeValue, bool) {
	//# if gd, ok := o.dflt.Members[o.name]; ok #
	gd := o.dflt.Find(o.name)
	if val, ok := gd.Value.(*TypeVal); ok {
		return val.ResolveAsTypeValue()
	}
	panic(F("wanted TypeValue for %q (in package %q), got %#v", o.name, o.dflt.Package, gd.Value))
}

type CVal struct {
	c string // C language expression
	t TypeValue
}

type NameVal struct {
	name string
	dflt *CMod
}

type SubVal struct {
	container Value
	sub       Value
}

type TypeVal struct {
	tv TypeValue
}

func (val *CVal) String() string {
	return Format("(%s:%s)", val.c, val.t)
}
func (val *NameVal) String() string {
	return Format("(%s:NameVal@%s)", val.name, val.dflt.Package)
}
func (val *TypeVal) String() string {
	return Format("(%s:TypeVal)", val.tv)
}
func (val *SubVal) String() string {
	return Format("(%s[%s])", val.container, val.sub)
}

func (val *CVal) Type() TypeValue {
	return val.t
}
func (val *NameVal) Type() TypeValue {
	gd := val.dflt.Find(val.name)
	if gd.Value == val {
		panic(F("Looping NameVal.Type for name %q dflt %q", val.name, val.dflt.Package))
	}
	if gd.Value.Type() == nil {
		panic(F("NameVal.Type for %q (@%q) has nil Value.Type(): %#v", val.name, val.dflt.Package, gd))
	}
	return gd.Value.Type()
}
func (val *TypeVal) Type() TypeValue {
	return TypeTO
}
func (val *SubVal) Type() TypeValue {
	return &PrimTV{Name: Format("(tv:TODO:Sub:%s:%s)", val.container, val.sub)}
}

func (val *CVal) ToC() string {
	return val.c
}
func (val *NameVal) ToC() string {
	return val.dflt.Find(val.name).Value.ToC()
}
func (val *TypeVal) ToC() string {
	return F("q", F("TYPE[%#v]", val.tv))
}
func (val *SubVal) ToC() string {
	panic(1240)
	// TODO: case string, slice, map.
}

type ImportVal struct {
	name string
}

func (val *ImportVal) String() string {
	return Format("ImportVal(%s)", val.name)
}
func (val *ImportVal) ToC() string {
	panic(1252)
}
func (val *ImportVal) Type() TypeValue {
	return ImportTO
}

func (cg *CGen) LoadModule(name string, pr printer) *CMod {
	log.Printf("LoadModule: << %q", name)
	if already, ok := cg.Mods[name]; ok {
		log.Printf("LoadModule: already loaded: %q", name)
		return already
	}

	filename := cg.Options.LibDir + "/" + name + ".go"
	r, err := os.Open(filename)
	if err != nil {
		panic(Format("cannot open %q: %v", filename, err))
	}
	defer r.Close()

	cm := NewCMod(name, cg)
	cg.Mods[name] = cm

	log.Printf("LoadModule: Parser")
	p := NewParser(r, filename)
	p.ParseModule(cm, cg)
	log.Printf("LoadModule: VisitGlobals")
	cm.VisitGlobals(p, pr)
	log.Printf("LoadModule: VisitGlobals Done")

	if name == "builtin" {
		// CGen provides quick access to the builtin Mod:
		cg.BuiltinMod = cm
	}
	return cm
}

type printer func(format string, args ...interface{})

func CompileToC(r io.Reader, sourceName string, w io.Writer, opt *Options) {
	pr := func(format string, args ...interface{}) {
		z := fmt.Sprintf(format, args...)
		log.Print("[[[[[  " + z + "  ]]]]]")
		fmt.Fprintf(w, "%s\n", z)
	}

	cg, cm := NewCGenAndMainCMod(opt, w)
	pr(`#include <stdio.h>`)
	pr(``)
	pr(`#include "runt.h"`)
	pr(``)
	if !opt.SkipBuiltin {
		cg.LoadModule("builtin", pr)
	}
	p := NewParser(r, sourceName)
	p.ParseModule(cm, cg)

	cm.VisitGlobals(p, pr)
}

func (cm *CMod) defineOnce(g *GDef) {
	if _, ok := cm.Members[g.Name]; ok {
		Panicf("module %s: redefined name: %s", cm.Package, g.Name)
	}
	cm.Members[g.Name] = g
	g.Package = cm.Package
	g.CName = CName(g.Package, g.Name)
}

func (cm *CMod) FirstSlotGlobals(p *Parser, pr printer) {
	// first visit: Slot the globals.
	for _, g := range p.Imports {
		cm.defineOnce(g)
	}
	for _, g := range p.Types {
		cm.defineOnce(g)
	}
	for _, g := range p.Consts {
		cm.defineOnce(g)
	}
	for _, g := range p.Vars {
		cm.defineOnce(g)
	}
	for _, g := range p.Funcs {
		cm.defineOnce(g)
	}
}

func (cm *CMod) SecondBuildGlobals(p *Parser, pr printer) {
	for _, g := range p.Imports {
		cm.CGen.LoadModule(g.Name, pr)
		g.Value = &ImportVal{g.Name}
		// If we care to do imports in order,
		// this is a good place to remember it.
		cm.CGen.ModsInOrder = append(cm.CGen.ModsInOrder, g.Name)
	}
	for _, g := range p.Types {
		Say(g.Package, g.Name, "2T")
		qc := cm.QuickCompiler(g)
		tmpX := NameTX{g.Name, g.Init, cm}
		tmpV := CompileTX(qc, tmpX, g.Init)
		if tmpV.TV == nil {
			panic(g.CName)
		}
		g.Value = &TypeVal{tmpV.TV}
	}
	for _, g := range p.Consts {
		Say(g.Package, g.Name, "2C")
		// not allowing g.Type on constants.
		g.Value = g.Init.VisitExpr(cm.QuickCompiler(g))
	}
	for _, g := range p.Vars {
		Say(g.Package, g.Name, "2V")
		val := g.Type_.VisitExpr(cm.QuickCompiler(g))
		tv, ok := val.ResolveAsTypeValue()
		if !ok {
			panic(F("got %#v when we wanted a TypeValue", val))
		}
		g.Value = &CVal{
			c: g.CName,
			t: tv,
		}
		if g.Init != nil {
			// We are writing the global init() function.
			initS := &AssignS{
				A:  []Expr{&IdentX{g.Name, cm}},
				Op: "=",
				B:  []Expr{g.Init},
			}
			initS.VisitStmt(cm.QuickCompiler(g))
		}
	}
	for _, g := range p.Funcs {
		Say(g.Package, g.Name, "2F")

		p := g.Init // .Type().(*FunctionTV).FuncRec
		log.Printf("bilbo %v", p)
		//panic(p)

		//XX g.Value = g.Init.VisitExpr(cm.QuickCompiler(g))
		//XX Say(g.Value, g.Name, "2F.")

		/*
			qc := cm.QuickCompiler(g)

			p := g.Init.Type().(*FunctionTV).FuncRec
			for i, e := range p.Ins {
				p.Ins[i] = CompileTX(qc, e, g.Init)
			}
			for i, e := range p.Outs {
				p.Outs[i] = CompileTX(qc, e, g.Init)
			}
		*/

		pr("//2F// %s // %s //", g.Name, g.Value)
	}
}

func (cm *CMod) ThirdDefineGlobals(p *Parser, pr printer) {
	// Third: Define the globals, except for functions.
	say := func(how string, g *GDef) {
		pr("// Third == %s %s ==", how, g.CName)
	}
	_ = say
	for _, g := range p.Types {
		Say("Third Types: " + g.Package + " " + g.Name)
		say("type", g)
		pr("typedef %s %s;", g.Value.(*TypeVal).tv.CType(), g.CName)
	}
	for _, g := range p.Vars {
		Say("Third Vars: " + g.Package + " " + g.Name)
		say("var", g)
		Say("1387", g.Value, g.CName)
		Say("1388", g.Value.Type(), g.CName)
		Say("1389", g.Value.Type().CType(), g.CName)
		pr("%s %s; //1390", g.Value.Type().CType(), g.CName)
	}
	for _, g := range p.Funcs {
		if g.Init == nil {
			panic(F("extern FUNC_1355 %s.%s; //3F//", g.Package, g.Name))
		}
		co := cm.QuickCompiler(g)
		pr("// Func Init: %#v", g.Init)

		fx := g.Init.(*FunctionX)
		frx := fx.FuncRecX
		fr := frx.VisitFuncRecX(co)
		g.Type_ = fx
		g.TV = &FunctionTV{fr}
		g.Value = &CVal{c: CName(cm.Package, g.Name), t: g.TV}

		pr("// third func: g.Value=%#v", g.Value)
		pr("// third func: g.Type_=%#v", g.Type_)
		pr("// third func: g.TV=%#v", g.TV)

		assert(g.Type_ != nil)
		pr("// Func Type_: %#v", g.Type_)

		_ = fx
		// decl := fx.FuncRecX.SignatureStr(g.CName)
		Say("Third Funcs: " + g.Package + " " + g.Name)
		say("func", g)
		//Say("extern %s;", decl)
		//pr("extern %s; //3F//", decl)
	}
}

func (cm *CMod) FourthInitGlobals(p *Parser, pr printer) {
	// Fourth: Initialize the global vars.
	if false {
		pr("void INIT() {")
		say := func(how string, g *GDef) {
			pr("// Fourth == %s %s ==", how, g.CName)
		}
		for _, g := range p.Vars {
			Say("Fourth(Var) " + g.Package + " " + g.Name)
			say("var", g)
			if g.Init != nil {
				initS := &AssignS{
					A:  []Expr{&IdentX{g.Name, cm}},
					Op: "=",
					B:  []Expr{g.Init},
				}
				// Emit initialization of var into init() function.
				initS.VisitStmt(cm.QuickCompiler(g))
			}
		}
		for _, g := range p.Funcs {
			Say("// Fourth(Func) TODO: Inline init functions:", g)
		}
		pr("} // INIT()")
	}

	for _, g := range p.Meths {
		_ = g
		/* SOON
		// Attach methods to their Struct.
		methRec := g.Init.(*FunctionX).FuncRec
		rcvr := *methRec.Receiver

		Say("Fourth(Meth)1 " + g.Package + " " + g.Name + " @ " + rcvr.String())
		qc := cm.CGen.Mods[rcvr.Package].QuickCompiler(g)
		rcvr = CompileTX(qc, rcvr, rcvr.Expr)
		g.Init.(*FunctionX).FuncRec.Receiver = &rcvr
		Say("Fourth(Meth)2 " + g.Package + " " + g.Name + " @ " + rcvr.String())

		// Type must be a pointer.
		if pointedType, ok := rcvr.TV.(*PointerTV); ok {
			if structType, ok := pointedType.E.(*StructTV); ok {
				rec := structType.StructRec
				meth := NameTV{ g.Name, &FunctionTV{BaseTV{}, methRec} }
				rec.Meths = append(rec.Meths, meth)

			} else {
				log.Panicf("expected *STRUCT receiver, got (not a struct) %v", rcvr)
			}
		} else {
			log.Panicf("expected *STRUCT receiver, got (not a pointer) %v", rcvr)
		}
		*/
	}
}
func (cm *CMod) FifthPrintFunctions(p *Parser, pr printer) {
	for _, g := range p.Funcs {
		Say("Fifth " + g.Package + " " + g.Name)
		pr("// Fifth FUNC: %T %s %q;", "#", "#", g.CName)
		if g.Init != nil {
			co := cm.QuickCompiler(g)
			co.EmitFunc(g)
			pr(co.Buf.String())
		} else {
			pr("// Cannot print function without body -- it must be extern.")
		}
	}

	for _, g := range p.Meths {
		_ = g
		/* SOON
		methRec := g.Init.(*FunctionX).FuncRec
		rcvr := *methRec.Receiver
		Say("Fifth(Meth) " + g.Package + " " + g.Name + " @ " + rcvr.String())

		// Type must be a pointer.
		pointedType, ok := rcvr.TV.(*PointerTV)
		if !ok {
			Panicf("To generate method for *STRUCT, expected pointer, got %v", rcvr.TV)
		}

		structType, ok := pointedType.E.(*StructTV)
		if !ok {
			Panicf("To generate method for *STRUCT, expected struct, got %v", pointedType.E)
		}

		Say("Fifth " + g.Package + " #" + structType.ToC() + "# " + g.Name)
		pr("// Fifth METH: #T #s %q;" /#g.Value.Type(), g.Value.Type().CType(),#/, g.Name)
		co := cm.QuickCompiler(g)
		co.EmitFunc(g)
		pr(co.Buf.String())

		/#
			if structType, ok := pointedType.E.(*StructTV); ok |
				rec := structType.StructRec
				methNat := X_NameAndType{
					g.Name,
					nil,
					&FunctionTV{BaseTV{}, methRec},
					g.Package,
				}
				// methNat = X_FillTV(qc, methNat, g.Init)
				rec.Meths = append(rec.Meths, methNat)
		#/
		*/
	}
}

func (cm *CMod) VisitGlobals(p *Parser, pr printer) {
	cm.FirstSlotGlobals(p, pr)
	cm.SecondBuildGlobals(p, pr)
	cm.ThirdDefineGlobals(p, pr)
	cm.FourthInitGlobals(p, pr)
	cm.FifthPrintFunctions(p, pr)
}

type GDef struct {
	Used    bool
	CGen    *CGen
	Package string
	Name    string
	CName   string

	Init  Expr // for Const or Var or Type
	Type_ Expr // for Const or Var or Func

	Value Value     // Next resolve global names to Values.
	TV    TypeValue // Only for Type: or embed in Value?
}

type Scope interface {
	Find(string) *GDef
}

type CMod struct { // isa Scope
	Package string
	CGen    *CGen
	Members map[string]*GDef
}

func (cm *CMod) Find(s string) *GDef {
	switch s {
	case "nil":
		return &GDef{
			Name:  s,
			Value: &CVal{c: "NULL", t: NilTO},
			TV:    NilTO,
		}
	case "true":
		return &GDef{
			Name:  s,
			Value: &CVal{c: "1", t: BoolTO},
			TV:    BoolTO,
		}
	case "false":
		return &GDef{
			Name:  s,
			Value: &CVal{c: "0", t: BoolTO},
			TV:    BoolTO,
		}
	}
	if cm == nil {
		// Fall back to just the prims.
		for _, e := range PrimTypeObjList {
			if e.Name == s {
				L("Fallback to PrimTypeObjList for %q", s)
				return &GDef{
					Name:  e.Name,
					Value: &TypeVal{e},
					TV:    TypeTO,
				}
			}
		}
		panic(F("Find %q, but with nil CMod", s))
	}
	if d, ok := cm.Members[s]; ok {
		return d
	}
	if cm.Package == "builtin" {
		return cm.CGen.Prims.Find(s)
	} else {
		return cm.CGen.BuiltinMod.Find(s)
	}
}

type CGen struct {
	Mods        map[string]*CMod // by module name
	BuiltinMod  *CMod
	Prims       *CMod
	ModsInOrder []string // reverse definition order
	Options     *Options
	W           *bufio.Writer
}

func NewCMod(name string, cg *CGen) *CMod {
	mod := &CMod{
		Package: name,
		CGen:    cg,
		Members: make(map[string]*GDef),
	}
	//< mod.Dog = NewDog(name, nil, nil, mod, cg)
	return mod
}
func NewCGenAndMainCMod(opt *Options, w io.Writer) (*CGen, *CMod) {
	mainMod := NewCMod("main", nil)
	cg := &CGen{
		Mods:    map[string]*CMod{"main": mainMod},
		Options: opt,
	}
	cg.Prims = &CMod{
		Package: "", // Use empty package name for Prims.
		CGen:    cg,
		Members: make(map[string]*GDef),
	}
	mainMod.CGen = cg

	// Populate PrimDog
	for _, e := range PrimTypeObjList {
		cg.Prims.Members[e.Name] = &GDef{
			Name:  e.Name,
			CName: "P_" + e.Name,
			Value: &TypeVal{e},
			TV:    TypeTO,
			Used:  false,
		}
	}

	return cg, mainMod
}

func (cm *CMod) VisitTypeExpr(x Expr) TypeValue {
	var gdef *GDef = nil
	val := x.VisitExpr(NewCompiler(cm, gdef))
	// if tv, ok := val.(*TypeVal); ok #
	if tv, ok := val.ResolveAsTypeValue(); ok {
		return tv
	} else {
		log.Panicf("Expected expr [ %v ] to compile to TypeValue, but it compiled to %T :: [ %v ]", x, val, val)
		panic(0)
	}
}
func (cm *CMod) VisitExpr(x Expr) Value {
	var gdef *GDef = nil
	return x.VisitExpr(NewCompiler(cm, gdef))
}
func (cm *CMod) QuickCompiler(gdef *GDef) *Compiler {
	return NewCompiler(cm, gdef)
}

func AssignNewVar(in NameTV, out NameTV) string {
	ctype := in.TV.CType()
	z, ok := in.TV.Assign(in.Name, in.TV)
	if ok {
		return Format("%s %s = %s;", ctype, out.Name, z)
	}
	panic(Format("Cannot assign from %s (type %s) to %s (type %s)", in.Name, in.TV, out.Name, out.TV))
}

type DeferRec struct {
	ToDo string
}

type Compiler struct {
	CMod         *CMod
	CGen         *CGen
	Subject      *GDef
	BreakTo      string
	ContinueTo   string
	CurrentBlock *Block
	Defers       []*DeferRec
	Buf          *Buf
}

func NewCompiler(cm *CMod, subject *GDef) *Compiler {
	co := &Compiler{
		CMod:    cm,
		CGen:    cm.CGen,
		Subject: subject,
		Buf:     &Buf{},
	}
	return co
}

func (co *Compiler) P(format string, args ...interface{}) {
	co.Buf.P(format, args...)
	co.Buf.P("\n")
}

func (co *Compiler) AddLocalTemp(tempName string, tempType TypeValue, initC string) {
	b := co.CurrentBlock
	if _, ok := b.locals[tempName]; ok {
		panic(F("CurrentBlock already has local with name: %q", tempName))
	}
	L("AddLocalTemp: b=%#v", b)
	b.locals[tempName] = &GDef{
		Name:  tempName,
		CName: tempName,
		Value: &CVal{c: tempName, t: tempType}, // Redundant?
		TV:    tempType,
	}
	if initC != "" {
		co.P("%s = %s; // AddLocalTemp", tempName, initC)
	} else {
		co.P("// %s uses default init // AddLocalTemp", tempName)
	}
}

// Compiler for Expressions

func (co *Compiler) VisitLitInt(x *LitIntX) Value {
	return &CVal{
		c: Format("%d", x.X),
		t: ConstIntTO,
	}
}
func (co *Compiler) VisitLitString(x *LitStringX) Value {
	return &CVal{
		c: Format("%q", x.X),
		t: StringTO,
	}
}
func (co *Compiler) VisitIdent(x *IdentX) Value {
	if co.CurrentBlock != nil {
		gd := co.CurrentBlock.Find(x.X)
		return gd
	}
	return co.CMod.Find(x.X)
	// return &NameVal{x.X, co.CMod}
}
func (co *Compiler) VisitBinOp(x *BinOpX) Value {
	a := x.A.VisitExpr(co)
	b := x.B.VisitExpr(co)
	return &CVal{
		c: Format("(%s) %s (%s)", a.ToC(), x.Op, b.ToC()),
		t: IntTO,
	}
}
func (co *Compiler) VisitConstructor(x *ConstructorX) Value {
	return &CVal{
		c: Format("(%s) alloc(C_%s)", x.Name, x.Name),
		t: &PointerTV{&StructTV{nil}}, // TODO!
	}
}
func (co *Compiler) VisitFunction(x *FunctionX) Value {
	L("VisitFunction: FuncRecX = %#v", x.FuncRecX)
	funcRec := x.FuncRecX.VisitFuncRecX(co)
	L("VisitFunction: FuncRec = %#v", funcRec)
	t := &FunctionTV{funcRec}
	return &CVal{c: "?1702?", t: t}
}

func (co *Compiler) VisitCall(x *CallX) Value {
	// type CallX struct { Func Expr; Args []Expr; }
	ser := Serial("call")
	//< var prep []string
	if ser == "call_103" {
		panic(ser)
	}

	co.P("// This is VisitCall %s", ser)
	co.P("// This is VisitCall %s", ser)
	co.P("// Func is X %#v", x.Func)
	co.P("// Args is len %d", len(x.Args))
	for i, a := range x.Args {
		co.P("// Args[%d] is X %#v", i, a)
	}

	funcVal := x.Func.VisitExpr(co)
	funcValType := funcVal.Type()
	funcRec := funcValType.(*FunctionTV).FuncRec
	fins := funcRec.Ins
	fouts := funcRec.Outs
	co.P("// IsMethod = %v", funcRec.IsMethod)
	co.P("// HasDotDotDot = %v", funcRec.HasDotDotDot)

	var argVals []Value
	for _, e := range x.Args {
		argVals = append(argVals, e.VisitExpr(co))
	}

	co.P("// Func is V %#v", funcVal)
	for i, a := range argVals {
		co.P("// Args[%d] is V %#v", i, a)
	}

	var extraFin NameTV
	var sliceType *SliceTV
	var numNormal, numExtras int

	if funcRec.HasDotDotDot {
		if len(fins)-1 > len(argVals) {
			panic(F("got %d args for func call, wanted at least %d args", len(argVals), len(fins)-1))
		}
		numNormal = len(fins) - 1
		numExtras = len(argVals) - numNormal

		extraFin = fins[numNormal]
		fins = fins[:numNormal]
		sliceType = &SliceTV{extraFin.TV}
	} else {
		if len(fins) != len(argVals) {
			panic(F("got %d args for func call, wanted %d args", len(argVals), len(fins)))
		}
	}

	co.P("// 1765: extraFin=%v ;; sliceType=%v ;; numNormal=%d ;; numExtras=%d", extraFin, sliceType, numNormal, numExtras)

	var argc []string

	// For the non-DotDotDot arguments
	for i, fin := range fins {
		temp := CName(ser, "in", D(i), fin.Name)
		co.AddLocalTemp(temp, fin.TV, argVals[i].ToC())
		argc = append(argc, temp)
	}

	if funcRec.HasDotDotDot {
		sliceName := CName(ser, "in", "vec")
		co.AddLocalTemp(sliceName, sliceType, "MakeSlice()")
		for i := 0; i < numExtras; i++ {
			co.P("%s = AppendSlice(%s, %s) // For extra input #%d", sliceName, sliceName, argVals[numNormal+i], i)
		}
		fins = append(fins, NameTV{sliceName, sliceType})
		argc = append(argc, sliceName)
	}

	if len(fouts) != 1 {
		var multi []NameTV
		for j, out := range fouts {
			rj := Format("_multi_%s_%d", ser, j)
			vj := NameTV{rj, out.TV}
			multi = append(multi, vj)
			co.AddLocalTemp(rj, out.TV, "")
			argc = append(argc, F("&%s", rj))
		}
		c := Format("(%s)(%s)", funcVal.ToC(), strings.Join(argc, ", "))
		return &CVal{c: c, t: &MultiTV{multi}}
	} else {
		c := Format("(%s)(%s)", funcVal.ToC(), strings.Join(argc, ", "))
		t := fouts[0].TV
		return &CVal{c: c, t: t}
	}
}

func (co *Compiler) VisitSub(x *SubX) Value {
	return &CVal{
		c: Format("SubXXX(%v)", x),
		t: IntTO,
	}
}

func (co *Compiler) VisitDot(dot *DotX) Value {
	log.Printf("VisitDot: <------ %v", dot)
	// val := co.ResolveTypeOfValue(dot.X.VisitExpr(co))
	val := dot.X.VisitExpr(co)
	log.Printf("VisitDot: val---- %v", val)

	if imp, ok := val.(*ImportVal); ok {
		modName := imp.name
		log.Printf("DOT %q %#v", modName, dot.Member)
		log.Printf("MODS: %#v", co.CGen.Mods)
		if otherMod, ok := co.CGen.Mods[modName]; ok {
			log.Printf("OM %#v", otherMod)
			_, ok := otherMod.Members[dot.Member]
			if !ok {
				panic(Format("cannot find member %s in module %s", dot.Member, modName))
			}
			panic(1728)
			/*
				z := otherMod.QuickCompiler(co.GDef).VisitIdent(&IdentX{X: dot.Member})
				L("VisitDot returns Imported thing: %#v", z)
				return z
			*/
		} else {
			panic(Format("imported %q but cannot find it in CGen.Mods: %#v", modName, co.CGen.Mods))
		}
	}

	/*
		if typ, ok := val.Type().(*PointerTV); ok {
			val = &XXXSimpleValue{Format("(*(%s))", val.ToC()), typ.E}
			L("VisitDot eliminating pointer: %#v of type %#v", val, val.Type())
		}
	*/

	if pointedType, ok := val.Type().(*PointerTV); ok {
		Say("pointedType", pointedType)
		if structType, ok := pointedType.E.(*StructTV); ok {
			Say("structType", structType)
			rec := structType.StructRec
			Say("rec", rec)
			if ftype, ok := FindTypeByName(rec.Fields, dot.Member); ok {
				z := &CVal{
					c: Format("(%s).%s", val.ToC(), dot.Member),
					t: ftype,
				}
				L("VisitDot returns Field: %#v", z)
				return z
			}
			if mtype, ok := FindTypeByName(rec.Meths, dot.Member); ok {
				z := &CVal{
					c: Format("METH__%s__%s@(%s)", rec.Name, dot.Member, val.ToC()),
					t: mtype,
				}
				L("VisitDot returns meth: %#v", z)
				return z
			}
		}
	}

	panic("DotXXX")
}

// Compiler for Statements
func (co *Compiler) VisitAssign(ass *AssignS) {
	L("//## assign..... %v   %v   %v", ass.A, ass.Op, ass.B)
	lenA, lenB := len(ass.A), len(ass.B)
	_ = lenA
	_ = lenB // TODO

	// Evalute the rvalues.
	var rvalues []Value
	for _, e := range ass.B {
		rvalues = append(rvalues, e.VisitExpr(co))
	}

	// Create local vars, if := is used.
	if ass.Op == ":=" {
		for _, a := range ass.A {
			if id, ok := a.(*IdentX); ok {
				var name string
				if id.X != "" && id.X != "_" {
					name = id.X
				} else {
					name = Serial("tmp")
				}

				local := &GDef{
					Name:  name,
					CName: Format("v_%s", name),
				}
				if _, ok := co.CurrentBlock.locals[id.X]; ok {
					co.P("// Already defined local: %q", id.X)
				} else {
					co.CurrentBlock.locals[id.X] = local
				}
			} else {
				log.Panicf("Expected an identifier in LHS of `:=` but got %v", a)
			}
		}
	}

	// If there is just one thing on right, and it is a CallX, set bcall.
	var bcall *CallX
	if len(ass.B) == 1 {
		switch t := ass.B[0].(type) {
		case *CallX:
			bcall = t
		}
	}

	switch {
	case ass.B == nil:
		// An lvalue followed by ++ or --.
		if len(ass.A) != 1 {
			Panicf("operator %v requires one lvalue on the left, got %v", ass.Op, ass.A)
		}
		// TODO check lvalue
		cvar := ass.A[0].VisitExpr(co).ToC()
		co.P("  (%s)%s;", cvar, ass.Op)

	case ass.A == nil && bcall == nil:
		// No assignment.  Just a non-function.  Does this happen?
		panic(Format("Lone expr is not a function call: [%v]", ass.B))

	case ass.A == nil && bcall != nil:
		// No assignment.  Just a function call.
		callVal := rvalues[0]

		_ = callVal // throw away the result.

		co.P("// @@@@ Please Call %v", callVal)
		co.P("// @@@@ Please Call %#v", callVal)

		co.P("(void) %s; // Call with no assign: 1932", callVal.ToC())

	case len(ass.A) > 1 && bcall != nil:
		// From 1 call, to 2 or more assigned vars.
		var buf Buf
		buf.P("((%s)(", bcall.Func.VisitExpr(co).ToC())
		for i, arg := range bcall.Args {
			if i > 0 {
				buf.P(", ")
			}
			buf.P("%s", arg.VisitExpr(co).ToC())
		}
		for i, arg := range ass.A {
			if len(bcall.Args)+i > 0 {
				buf.P(", ")
			}
			// TODO -- VisitAddr ?
			buf.P("&(%s)", arg.VisitExpr(co).ToC())
		}
		buf.P("))")
	default:
		if len(ass.A) != len(ass.B) {
			Panicf("wrong number of values in assign: left has %d, right has %d", len(ass.A), len(ass.B))
		}
		for i, val := range rvalues {
			lhs := ass.A[i]
			switch t := lhs.(type) {
			case *IdentX:
				// TODO -- check that variable t.X has the right type.
				switch ass.Op {
				case "=":
					// TODO check Globals
					cvar := "v_" + t.X
					L("cvar = %v", cvar)
					L("val = %v", val)
					L("val.Type() = %v", val.Type())
					L("val.Type().CType() = %v", val.Type().CType())
					L("val.ToC() = %v", val.ToC())
					co.P("  %s = (%s)(%s);", cvar, val.Type().CType(), val.ToC())
				case ":=":
					// TODO check Globals
					cvar := Format("%s %s", val.Type().CType(), "v_"+t.X)
					co.P("  %s = (%s)(%s);", cvar, val.Type().CType(), val.ToC())
				}
			default:
				log.Fatalf("bad VisitAssign LHS: %#v", ass.A)
			}
		}
	} // switch
}
func (co *Compiler) VisitReturn(ret *ReturnS) {
	log.Printf("return..... %v", ret.X)
	switch len(ret.X) {
	case 0:
		co.P("  return;")
	case 1:
		val := ret.X[0].VisitExpr(co)
		log.Printf("return..... val=%v", val)
		co.P("  return %s;", val.ToC())
	default:
		Panicf("multi-return not imp: %v", ret)
	}
}
func (co *Compiler) VisitWhile(wh *WhileS) {
	label := Serial("while")
	co.P("Break_%s:  while(1) {", label)
	if wh.Pred != nil {
		co.P("    t_bool _while_ = (t_bool)(%s);", wh.Pred.VisitExpr(co).ToC())
		co.P("    if (!_while_) break;")
	}
	savedB, savedC := co.BreakTo, co.ContinueTo
	co.BreakTo, co.ContinueTo = "Break_"+label, "Cont_"+label
	wh.Body.VisitStmt(co)
	co.P("  }")
	co.P("Cont_%s: {}", label)
	co.BreakTo, co.ContinueTo = savedB, savedC
}
func (co *Compiler) VisitBreak(sws *BreakS) {
	if co.BreakTo == "" {
		Panicf("cannot break from here")
	}
	co.P("goto %s;", co.BreakTo)
}
func (co *Compiler) VisitContinue(sws *ContinueS) {
	if co.ContinueTo == "" {
		Panicf("cannot continue from here")
	}
	co.P("goto %s;", co.ContinueTo)
}
func (co *Compiler) VisitIf(ifs *IfS) {
	co.P("  { t_bool _if_ = %s;", ifs.Pred.VisitExpr(co).ToC())
	co.P("  if( _if_ ) {")
	ifs.Yes.VisitStmt(co)
	if ifs.No != nil {
		co.P("  } else {")
		ifs.No.VisitStmt(co)
	}
	co.P("  }}")
}
func (co *Compiler) VisitSwitch(sws *SwitchS) {
	co.P("  { t_int _switch_ = %s;", sws.Switch.VisitExpr(co).ToC())
	for _, c := range sws.Cases {
		co.P("  if (")
		for _, m := range c.Matches {
			co.P("_switch_ == %s ||", m.VisitExpr(co).ToC())
		}
		co.P("      0 ) {")
		c.Body.VisitStmt(co)
		co.P("  } else ")
	}
	co.P("  {")
	if sws.Default != nil {
		sws.Default.VisitStmt(co)
	}
	co.P("  }")
	co.P("  }")
}
func (co *Compiler) VisitBlock(a *Block) {
	if a == nil {
		panic(8881)
	}
	prevBlock := co.CurrentBlock
	co.CurrentBlock = a
	a.compiler = prevBlock.compiler // TODO is this good?
	for i, e := range a.stmts {
		ser := Serial("block")
		co.P("// @@ VisitBlock[%s,%d] <= %q", ser, i, F("%v", e))
		e.VisitStmt(co)
		log.Printf("VisitBlock[%d] ==>\n<<<\n%s\n>>>", i, co.Buf.String())
	}
	co.CurrentBlock = prevBlock
}

type Buf struct {
	W bytes.Buffer
}

func (buf *Buf) XXX_A(s string) {
	buf.W.WriteString(s)
	buf.W.WriteByte('\n')
}
func (buf *Buf) P(format string, args ...interface{}) {
	z := fmt.Sprintf(format, args...)
	log.Printf("<<<[ %s ]>>>", z)
	buf.W.WriteString(z)
	buf.W.WriteByte('\n')
}
func (buf *Buf) String() string {
	return buf.W.String()
}

func (co *Compiler) EmitFunc(gd *GDef) {
	block := &Block{
		debugName: gd.CName,
		locals:    make(map[string]*GDef),
		parent:    nil,
		compiler:  co,
	}
	rec := gd.TV.(*FunctionTV).FuncRec
	co.P(rec.SignatureStr(gd.CName))

	// Figure out the names of Func inputs, and create locals for them.
	for i, in := range rec.Ins {
		var name string
		if in.Name != "" && in.Name != "_" {
			name = in.Name
		} else {
			name = Format("__%d", i)
		}

		local := &GDef{
			Name:  name,
			CName: Format("in_%s", name),
			TV:    in.TV,
		}
		block.locals[name] = local
	}

	// Figure out the names of Func outputs, and create locals for them.
	for i, out := range rec.Outs {
		var name string
		if out.Name != "" && out.Name != "_" {
			name = out.Name
		} else {
			name = Format("__%d", i)
		}

		local := &GDef{
			Name:  name,
			CName: Format("(*out_%s)", name),
			TV:    out.TV,
		}
		block.locals[name] = local
	}

	if rec.FuncRecX.Body == nil {
		// Function has no body, so it should be natively-defined.
		co.P("; //EmitFunc: NATIVE\n")
		return
	}

	// For the normal case of a function with a body.
	// Emit the translated body onto a buffer.
	// Then define the local variables, and initialize them,
	// and emit them before you emit the body from the buffer.

	co.P("{\n")

	prevBuf := co.Buf
	co.Buf = &Buf{}

	prevBlock := co.CurrentBlock
	co.CurrentBlock = rec.FuncRecX.Body
	rec.FuncRecX.Body.compiler = co
	rec.FuncRecX.Body.VisitStmt(co)
	co.CurrentBlock = prevBlock

	prevBuf.P("// Adding LOCALS to Func")
	for name, e := range rec.FuncRecX.Body.locals {
		prevBuf.P("// LOCAL %q IS %v", name, e)
		prevBuf.P("auto %v %v; // DEF LOCAL 2145", e.TV.CType(), e.CName)
	}

	prevBuf.P(co.Buf.String())
	co.Buf = prevBuf

	co.P("\n}\n")
}

///////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////

// TODO: Keep track of reachable Globals.
type ActiveTracker struct {
	List []*GDef // Reverse Definition Order.
}

///////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////
///////////////////////////////////////////////////////////

type Nando struct {
	Locals map[string]TypeValue
}

func (o *Nando) VisitLitInt(x *LitIntX) Value {
	return nil
}
func (o *Nando) VisitLitString(x *LitStringX) Value {
	return nil
}
func (o *Nando) VisitIdent(x *IdentX) Value {
	return nil
}
func (o *Nando) _VisitIdent_(x *IdentX) Value {
	return nil
}
func (o *Nando) VisitBinOp(x *BinOpX) Value {
	return nil
}
func (o *Nando) VisitConstructor(x *ConstructorX) Value {
	return nil
}
func (o *Nando) VisitFunction(x *FunctionX) Value {
	return nil
}
func (o *Nando) VisitCall(x *CallX) Value {
	return nil
}
func (o *Nando) VisitSub(x *SubX) Value {
	return nil
}
func (o *Nando) VisitDot(dot *DotX) Value {
	return nil
}

func (o *Nando) VisitAssign(ass *AssignS) {
}
func (o *Nando) VisitReturn(ret *ReturnS) {
}
func (o *Nando) VisitWhile(wh *WhileS) {
}
func (o *Nando) VisitBreak(sws *BreakS) {
}
func (o *Nando) VisitContinue(sws *ContinueS) {
}
func (o *Nando) VisitIf(ifs *IfS) {
}
func (o *Nando) VisitSwitch(sws *SwitchS) {
}
func (o *Nando) VisitBlock(a *Block) {
}

type VarStack struct {
	A []NameTV
}

func (o *VarStack) Push(n NameTV) {
	o.A = append(o.A, n)
}
func (o *VarStack) Pop() NameTV {
	z := o.A[len(o.A)-1]
	o.A = o.A[:len(o.A)-1]
	return z
}

type FuncBuilder struct {
	Handles VarStack
	Strings VarStack
	Bytes   VarStack
	Words   VarStack
}
