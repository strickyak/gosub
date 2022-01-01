package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	//"path/filepath"
	"reflect"
	//"runtime/debug"
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

/*
type StructRecX struct {
	Name   string
	Fields []NameTX
	Meths  []NameTX
}

type InterfaceRecX struct {
	Name  string
	Meths []NameTX
}
*/
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
	z := &InterfaceRec{
		Name:  p.Name,
		Meths: make([]NameTV, len(p.Meths)),
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
	x := o.FuncRecX
	z := &FuncRec{
		Ins:          make([]NameTV, len(x.Ins)),
		Outs:         make([]NameTV, len(x.Outs)),
		HasDotDotDot: x.HasDotDotDot,
		IsMethod:     x.IsMethod,
		FuncRecX:     x,
	}
	for i, e := range x.Ins {
		z.Ins[i] = NameTV{e.Name, e.Mod.VisitTypeExpr(e.Expr)}
	}
	for i, e := range x.Outs {
		z.Outs[i] = NameTV{e.Name, e.Mod.VisitTypeExpr(e.Expr)}
	}
	return &TypeVal{&FunctionTV{z}}
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

// Needed because parser can create a TypeValue before
// compiler starts running.
type ForwardTV struct {
	GDef *GDef
}

type MultiTV struct {
	Multi []NameTV
}

const kResolveTooDeep = 16

func (tv *ForwardTV) Resolve() TypeValue {
	for i := 0; i < kResolveTooDeep; i++ {
		if fwd2, ok := tv.GDef.TV.(*ForwardTV); ok {
			tv = fwd2
		} else {
			return tv.GDef.TV
		}
	}
	panic("Resolve too deep")
}

// Type values have type TypeTV (the metatype).
func (tv *PrimTV) Type() TypeValue      { return &TypeTV{} }
func (tv *TypeTV) Type() TypeValue      { return &TypeTV{} }
func (tv *PointerTV) Type() TypeValue   { return &TypeTV{} }
func (tv *SliceTV) Type() TypeValue     { return &TypeTV{} }
func (tv *MapTV) Type() TypeValue       { return &TypeTV{} }
func (tv *ForwardTV) Type() TypeValue   { return tv.Resolve() }
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
func (tv *ForwardTV) ToC() string {
	return Format("ZForwardTV(%s.%s)", tv.GDef.Package, tv.GDef.Name)
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

func (o *ForwardTV) Equal(typ TypeValue) bool {
	return o.Resolve().Equal(typ)
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
func (o *SliceTV) CType() string     { return "Slice" }
func (o *MapTV) CType() string       { return "Map" }
func (o *StructTV) CType() string    { return "Struct" }
func (o *PointerTV) CType() string   { return "Pointer" }
func (o *InterfaceTV) CType() string { return "Interface" }
func (o *TypeTV) CType() string      { return "Type" }
func (o *MultiTV) CType() string     { return "Multi" }
func (o *ForwardTV) CType() string   { return o.Resolve().CType() }

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
func (o *ForwardTV) Assign(c string, typ TypeValue) (z string, ok bool) {
	z, ok = o.Resolve().Assign(c, typ)
	return
}

func (o *ForwardTV) Cast(c string, typ TypeValue) (z string, ok bool) {
	z, ok = o.Resolve().Cast(c, typ)
	return
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
func (tv *ForwardTV) String() string {
	return Format("ForwardTV(%v.%v)", tv.GDef.Package, tv.GDef.Name)
}
func (tv *StructTV) String() string {
	return Format("StructTV(%v)", tv.StructRec.Name)
}
func (tv *InterfaceTV) String() string {
	return Format("InterfaceTV(%s/%d)", tv.InterfaceRec.Name, len(tv.InterfaceRec.Meths))
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
	X        string
	Outer    *CMod // Outer scope where defined -- but the IdentX may or may not be global.
	Resolved bool  // if we looked for the *GDef.
	GDef     *GDef // cache the *GDef if resolved.
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
	return fmt.Sprintf("Function(%s)", o.FuncRecX)
}
func (o *FunctionX) VisitExpr(v ExprVisitor) Value {
	//log.Printf("439: FunctionX=%#v", o)
	//log.Printf("439: FuncRec=%#v", o.FuncRec)
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

/*
func UNUSED_MethRecToFuncRec(mr *FuncRec) *FuncRec {
	ins := []X_NameAndType{*mr.Receiver}
	ins = append(ins, mr.Ins...)
	fn := &FuncRec{
		Receiver:     nil,
		Ins:          ins,
		Outs:         mr.Outs,
		HasDotDotDot: mr.HasDotDotDot,
		Body:         mr.Body,
	}
	RegisterFuncRec(fn) // TODO: should this be lazy?
	return fn
}
*/

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
		P(&b, "%s ", r.Outs[0].TV)
	} else {
		P(&b, "void ")
	}
	P(&b, "%s(", daFunc)
	for i, nat := range r.Ins {
		if i > 0 {
			b.WriteByte(',')
		}
		Say(nat)
		Say(nat.TV)
		P(&b, "%s in_%s", nat.TV, SerialIfEmpty(nat.Name))
	}
	if len(r.Outs) != 1 {
		for i, nat := range r.Outs {
			if i > 0 {
				b.WriteByte(',')
			}
			L("out [%d]: %s", i, nat.TV)
			P(&b, "%s *out_%s", nat.TV, SerialIfEmpty(nat.Name))
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
	LocalXs  []NameTX
	Locals   []NameTV
	Stmts    []Stmt
	Parent   *Block
	FuncRecX *FuncRecX
}

func (o *Block) VisitStmt(v StmtVisitor) {
	v.VisitBlock(o)
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
}

type Value interface {
	String() string
	Type() TypeValue
	Prep() []string
	ToC() string
}

type CVal struct {
	prep []string
	c    string // C language expression
	t    TypeValue
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
	if val.prep == nil {
		return Format("(%s:%s)", val.c, val.t)
	} else {
		return Format("({%s}%s:%s)", strings.Join(val.prep, ";"), val.c, val.t)
	}
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
	return &PrimTV{Name: Format("(tv:TODO:Name:%s)", val.name)}
}
func (val *TypeVal) Type() TypeValue {
	return TypeTO
}
func (val *SubVal) Type() TypeValue {
	return &PrimTV{Name: Format("(tv:TODO:Sub:%s:%s)", val.container, val.sub)}
}

func (val *CVal) Prep() []string {
	return val.prep
}
func (val *NameVal) Prep() []string {
	return nil
}
func (val *TypeVal) Prep() []string {
	return nil
}
func (val *SubVal) Prep() []string {
	var z []string
	z = append(z, val.container.Prep()...)
	return append(z, val.sub.Prep()...)
}

func (val *CVal) ToC() string {
	return val.c
}
func (val *NameVal) ToC() string {
	panic(1234)
}
func (val *TypeVal) ToC() string {
	panic(1237)
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
func (val *ImportVal) Prep() []string {
	return nil
}

func (cg *CGen) LoadModule(name string) *CMod {
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

	cm := NewCMod(name, cg, cg.W)
	cg.Mods[name] = cm

	log.Printf("LoadModule: Parser")
	p := NewParser(r, filename)
	p.ParseModule(cm, cg)
	log.Printf("LoadModule: VisitGlobals")
	cm.VisitGlobals(p)
	log.Printf("LoadModule: VisitGlobals Done")

	if name == "builtin" {
		// CGen provides quick access to the builtin Mod:
		cg.BuiltinMod = cm
	}
	return cm
}

func CompileToC(r io.Reader, sourceName string, w io.Writer, opt *Options) {
	cg, cm := NewCGenAndMainCMod(opt, w)
	cm.P("#include <stdio.h>")
	cm.P("#include \"runt.h\"")
	if !opt.SkipBuiltin {
		cg.LoadModule("builtin")
	}
	p := NewParser(r, sourceName)
	p.ParseModule(cm, cg)

	cm.VisitGlobals(p)
}

func (cm *CMod) mustNotExistYet(s string) {
	if _, ok := cm.Scope.GDefs[s]; ok {
		Panicf("redefined global name: %s", s)
	}
}

func (cm *CMod) FirstSlotGlobals(p *Parser) {
	// first visit: Slot the globals.
	slot := func(g *GDef) {
		g.Package = cm.Package
		cm.mustNotExistYet(g.Name)
		cm.Scope.GDefs[g.Name] = g
		g.FullName = FullName(g.Package, g.Name)
	}
	for _, g := range p.Imports {
		slot(g)
	}
	for _, g := range p.Types {
		slot(g)
	}
	for _, g := range p.Consts {
		slot(g)
	}
	for _, g := range p.Vars {
		slot(g)
	}
	for _, g := range p.Funcs {
		slot(g)
	}
}

func (cm *CMod) SecondBuildGlobals(p *Parser) {
	for _, g := range p.Imports {
		cm.CGen.LoadModule(g.Name)
		g.Value = &ImportVal{g.Name}
		// If we care to do imports in order,
		// this is a good place to remember it.
		cm.CGen.ModsInOrder = append(cm.CGen.ModsInOrder, g.Name)
	}
	for _, _ = range p.Types {
		for _, g := range p.Types {
			Say(g.Package, g.Name, "2T")
			qc := cm.QuickCompiler(g)
			tmpX := NameTX{g.Name, g.Init, cm}
			tmpV := CompileTX(qc, tmpX, g.Init)
			if tmpV.TV == nil {
				panic(g.FullName)
			}
			g.Value = &TypeVal{tmpV.TV}
		}
	}
	for _, g := range p.Consts {
		Say(g.Package, g.Name, "2C")
		// not allowing g.Type on constants.
		g.Value = g.Init.VisitExpr(cm.QuickCompiler(g))
	}
	for _, g := range p.Vars {
		Say(g.Package, g.Name, "2V")
		typeValue := g.Type.VisitExpr(cm.QuickCompiler(g)).(*TypeVal).tv
		g.Value = &CVal{
			c: g.FullName,
			t: typeValue,
		}
		if g.Init != nil {
			// We are writing the global init() function.
			initS := &AssignS{
				A:  []Expr{&IdentX{g.Name, cm, true, g}},
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

		cm.P("//2F// %s // %s //", g.Name, g.Value)
	}
}

func (cm *CMod) ThirdDefineGlobals(p *Parser) {
	// Third: Define the globals, except for functions.
	say := func(how string, g *GDef) {
		cm.P("// Third == %s %s ==", how, g.FullName)
	}
	_ = say
	for _, g := range p.Types {
		Say("Third Types: " + g.Package + " " + g.Name)
		say("type", g)
		cm.P("typedef %s %s;", g.Value.(*TypeVal).tv.CType(), g.FullName)
	}
	for _, g := range p.Vars {
		Say("Third Vars: " + g.Package + " " + g.Name)
		say("var", g)
		Say("%s %s;", g.Value, g.FullName)
		Say("%s %s;", g.Value.Type(), g.FullName)
		Say("%s %s;", g.Value.Type().CType(), g.FullName)
		cm.P("%s %s;", g.Value.Type().CType(), g.FullName)
	}
	for _, g := range p.Funcs {
		if g.Value != nil {
			ft := g.Value.Type().(*FunctionTV)
			decl := ft.FuncRec.SignatureStr(g.FullName)
			_ = ft
			Say("Third Funcs: " + g.Package + " " + g.Name)
			say("func", g)
			Say("extern %s;", decl)
			cm.P("extern %s; //3F//", decl)
		} else {
			cm.P("extern FUNC_TODO_1355 %s.%s; //3F//", g.Package, g.Name)
		}
	}
}

func (cm *CMod) FourthInitGlobals(p *Parser) {
	// Fourth: Initialize the global vars.
	if false {
		cm.P("void INIT() {")
		say := func(how string, g *GDef) {
			cm.P("// Fourth == %s %s ==", how, g.FullName)
		}
		for _, g := range p.Vars {
			Say("Fourth(Var) " + g.Package + " " + g.Name)
			say("var", g)
			if g.Init != nil {
				initS := &AssignS{
					A:  []Expr{&IdentX{g.Name, cm, true, g}},
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
		cm.P("} // INIT()")
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
func (cm *CMod) FifthPrintFunctions(p *Parser) {
	for _, g := range p.Funcs {
		Say("Fifth " + g.Package + " " + g.Name)
		cm.P("// Fifth FUNC: %T %s %q;", g.Value.Type(), g.Value.Type().CType(), g.FullName)
		co := cm.QuickCompiler(g)
		co.EmitFunc(g)
		cm.P(co.Buf.String())
		cm.Flush()
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
		cm.P("// Fifth METH: #T #s %q;" /#g.Value.Type(), g.Value.Type().CType(),#/, g.Name)
		co := cm.QuickCompiler(g)
		co.EmitFunc(g)
		cm.P(co.Buf.String())
		cm.Flush()

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

func (cm *CMod) VisitGlobals(p *Parser) {
	cm.FirstSlotGlobals(p)
	cm.SecondBuildGlobals(p)
	cm.ThirdDefineGlobals(p)
	cm.FourthInitGlobals(p)
	cm.FifthPrintFunctions(p)
	cm.Flush()
}

type GDef struct {
	CGen     *CGen
	Package  string
	Name     string
	FullName string
	Init     Expr  // for Const or Var or Type
	Value    Value // Next resolve global names to Values.
	Type     Expr  // for Const or Var
	TV       TypeValue
	Active   bool
}

type Scope struct {
	Name   string
	GDefs  map[string]*GDef // by short name
	Parent *Scope           // upper scope
	GDef   *GDef            // if local to a function
	CMod   *CMod            // if owned by a module
	CGen   *CGen            // for finding Builtins.
}
type CMod struct {
	W       *bufio.Writer
	Package string
	CGen    *CGen
	Scope   *Scope // members of module.
}
type CGen struct {
	Prims       *Scope
	Mods        map[string]*CMod // by module name
	BuiltinMod  *CMod
	ModsInOrder []string // reverse definition order
	Options     *Options
	W           *bufio.Writer
}

func NewScope(name string, parent *Scope, gdef *GDef, cmod *CMod, cgen *CGen) *Scope {
	return &Scope{
		Name:   name,
		GDefs:  make(map[string]*GDef),
		Parent: parent,
		GDef:   gdef,
		CMod:   cmod,
		CGen:   cgen,
	}
}
func (sco *Scope) Find(s string) (*GDef, *Scope, bool) {
	for p := sco; p != nil; p = p.Parent {
		if gdef, ok := p.GDefs[s]; ok {
			return gdef, p, true
		}
	}
	bim := sco.CGen.BuiltinMod
	if bim != nil {
		if gdef, ok := bim.Scope.GDefs[s]; ok {
			return gdef, bim.Scope, true
		}
	}
	prims := sco.CGen.Prims
	if gdef, ok := prims.GDefs[s]; ok {
		return gdef, prims, true
	}
	return nil, nil, false
}

func NewCMod(name string, cg *CGen, w io.Writer) *CMod {
	mod := &CMod{
		W:       bufio.NewWriter(w),
		Package: name,
		CGen:    cg,
	}
	mod.Scope = NewScope(name, nil, nil, mod, cg)
	return mod
}
func NewCGenAndMainCMod(opt *Options, w io.Writer) (*CGen, *CMod) {
	mainMod := NewCMod("main", nil, w)
	cg := &CGen{
		Mods:    map[string]*CMod{"main": mainMod},
		W:       mainMod.W,
		Options: opt,
	}
	cg.Prims = NewScope("_prims_", nil, nil, nil, cg)
	mainMod.CGen = cg

	// Populate PrimScope
	for _, e := range PrimTypeObjList {
		cg.Prims.GDefs[e.Name] = &GDef{
			Name:     e.Name,
			FullName: "P_" + e.Name,
			Value:    &TypeVal{e},
			TV:       TypeTO, // the metatype
			Active:   true,
		}
	}

	return cg, mainMod
}

func (cm *CMod) VisitTypeExpr(x Expr) TypeValue {
	var gdef *GDef = nil
	val := x.VisitExpr(NewCompiler(cm, gdef))
	if tv, ok := val.(*TypeVal); ok {
		return tv.tv
	} else {
		log.Panicf("Expected expr [ %v ] to compile to TypeValue, but it compiled to [ %v ]", x, val)
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

func (cm *CMod) P(format string, args ...interface{}) {
	fmt.Fprintf(cm.W, format+"\n", args...)
}
func (cm *CMod) Flush() {
	cm.W.Flush()
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
	GDef         *GDef
	BreakTo      string
	ContinueTo   string
	CurrentBlock *Block
	Locals       *Scope
	Defers       []*DeferRec
	Buf          bytes.Buffer
}

func NewCompiler(cm *CMod, gdef *GDef) *Compiler {
	co := &Compiler{
		CMod: cm,
		CGen: cm.CGen,
		GDef: gdef,
	}

	Say("cm.Scope", cm.Scope)
	Say("cm.CGen", cm.CGen)
	Say("gdef", gdef)
	gloss := "locals of something"
	if gdef != nil {
		gloss = "locals of " + gdef.FullName
	}
	co.Locals = NewScope(gloss, cm.Scope, gdef, nil, cm.CGen)
	return co
}

func (co *Compiler) P(format string, args ...interface{}) {
	fmt.Fprintf(&co.Buf, format, args...)
	co.Buf.WriteByte('\n')
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
	log.Printf("VisitIdent <= %v", x)
	z := co._VisitIdent_(x)
	log.Printf("VisitIdent => %v", z)
	return z
}
func (co *Compiler) _VisitIdent_(x *IdentX) Value {
	if gdef, _, ok := co.Locals.Find(x.X); ok {
		if gdef.Value != nil {
			return gdef.Value
		}
		return &NameVal{name: x.X, dflt: nil}
	}
	log.Panicf("Identifier not found: %q in %v", x.X, co.GDef)
	return nil
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
	L("VisitFunction: %#v", x.FuncRecX)
	L("VisitFunction: %v", x.FuncRecX)
	// TODO // return &XXXSimpleValue{"TODO:1794", &FunctionTV{x.FuncRec}}
	panic("TODO:1792")
}

func (co *Compiler) VisitCall(x *CallX) Value {
	co.CMod.W.Flush()
	ser := Serial("call")
	co.P("// %s: Calling Func: %#v", ser, x.Func)
	for i, a := range x.Args {
		co.P("// %s: Calling with Arg [%d]: %#v", ser, i, a)
	}

	co.P("{")
	log.Printf("x.Func: %#v", x.Func)
	funcVal := x.Func.VisitExpr(co)
	_ = funcVal // TODO

	log.Printf("funcVal: %#v", funcVal)
	funcTV, ok := funcVal.Type().(*FunctionTV)
	if !ok {
		log.Printf("needed a value type, but got %v", funcVal.Type())
		_ = funcVal.Type().(*FunctionTV)
	}
	funcRec := funcTV.FuncRec

	var argc []string
	for i, in := range funcRec.Ins {
		value := x.Args[i].VisitExpr(co)
		if !reflect.DeepEqual(value.Type(), in.TV) {
			Say(funcVal)
			Say(value)
			Say(value.Type())
			Say(in)
			Say(in.TV)
			log.Printf("WARNING: Function %q expects type %s for arg %d named %s, but got %q with type %s", funcVal.ToC(), in.String(), i, in.Name, value.ToC(), value.Type().String())
		}
		argc = append(argc, value.ToC())
	}
	if len(funcRec.Outs) != 1 {
		var multi []NameTV
		for j, out := range funcRec.Outs {
			rj := Format("_multi_%s_%d", ser, j)
			vj := NameTV{rj, out.TV}
			multi = append(multi, vj)
			// TODO // co.Locals[rj] = out.TV
			argc = append(argc, rj)
		}
		c := Format("(%s)(%s)", funcVal.ToC(), strings.Join(argc, ", "))
		return &CVal{c: c, t: &MultiTV{multi}}
	} else {
		c := Format("(%s)(%s)", funcVal.ToC(), strings.Join(argc, ", "))
		t := funcRec.Outs[0].TV
		return &CVal{c: c, t: t}
	}
	panic("TODO=1733")
}

func (co *Compiler) VisitSub(x *SubX) Value {
	return &CVal{
		c: Format("SubXXX(%v)", x),
		t: IntTO,
	}
}

func (co *Compiler) ResolveType(tv TypeValue) TypeValue {
	if fwd, ok := tv.(*ForwardTV); ok {
		if cm, ok := co.CGen.Mods[fwd.GDef.Package]; ok {
			if gd, ok := cm.Scope.GDefs[fwd.GDef.Name]; ok {
				tv = gd.Init.(TypeValue)
			}
			/* TODO
			} else if gd, ok := co.CGen.BuiltinMod.Scope.GDefs[fwd.Name]; ok {
				tv = gd.Init.(TypeValue)
			*/
		}
	}
	return tv
}

/* ?TODO?
func (co *Compiler) ResolveTypeOfValue(val Value) Value {
	return &CVal{c: val.ToC(), t: co.ResolveType(val.Type())}
}
*/

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
			log.Printf("GD %#v", otherMod.Scope.GDefs)
			_, ok := otherMod.Scope.GDefs[dot.Member]
			if !ok {
				panic(Format("cannot find member %s in module %s", dot.Member, modName))
			}
			z := otherMod.QuickCompiler(co.GDef).VisitIdent(&IdentX{X: dot.Member})
			L("VisitDot returns Imported thing: %#v", z)
			return z
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
	co.P("//## assign..... %v   %v   %v", ass.A, ass.Op, ass.B)
	lenA, lenB := len(ass.A), len(ass.B)
	_ = lenA
	_ = lenB // TODO

	if ass.Op == ":=" {
		// bug: should wait until after RHS to define these.
		for _, a := range ass.A {
			if id, ok := a.(*IdentX); ok {
				var name string
				if id.X != "" && id.X != "_" {
					name = id.X
				} else {
					name = Serial("tmp_")
				}

				local := &GDef{
					Name:     name,
					FullName: Format("v_%s", name),
				}
				co.Locals.GDefs[id.X] = local
			} else {
				log.Panicf("Expected an identifier in LHS of `:=` but got %v", a)
			}
		}
	}

	// Evalute the rvalues.
	var rvalues []Value
	for _, e := range ass.B {
		rvalues = append(rvalues, e.VisitExpr(co))
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
		panic(Format("Lone expr is not a funciton call: [%v]", ass.B))

	case ass.A == nil && bcall != nil:
		// No assignment.  Just a function call.
		log.Printf("bcall=%#v", bcall)
		visited := bcall.VisitExpr(co)
		log.Printf("visited=%#v", visited)

		/* TODO
				funcRec := visited.(*DefFunc).FuncRec

				funcname := funcRec.Function.Name
				log.Printf("funcname=%s", funcname)

				if lenB != len(bcall.Args) {
					panic(Format("Function %s wants %d args, got %d", funcname, len(bcall.Args), lenB))
				}
				ser := Serial("call")
				co.P("{ // %s", ser)
				c := Format(" %s( fp", funcname)
				for i, in := range funcRec.Ins {
					val := ass.B[i].VisitExpr(co)
					expectedType := in.TV
					if expectedType != val.Type() {
						panic(Format("bad type: expected %s, got %s", expectedType, val.Type()))
					}
					co.P("  %s %s_in_%d = %s;", in.TV.CType(), ser, i, val.ToC())
					c += Format(", %s_in_%d", ser, i)
				}
				for i, out := range funcRec.Outs {
					co.P("  %s %s_out_%d;", out.TV.CType(), ser, i)
					c += Format(", &%s_out_%d", ser, i)
				}
				c += " );"
				co.P("  %s\n} // %s", c, ser)
		        TODO */
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
	for i, e := range a.Stmts {
		log.Printf("VisitBlock[%d]", i)
		e.VisitStmt(co)
		log.Printf("VisitBlock[%d] ==>\n<<<\n%s\n>>>", i, co.Buf.String())
	}
	co.CurrentBlock = prevBlock
}

type Buf struct {
	W bytes.Buffer
}

func (buf *Buf) A(s string) {
	buf.W.WriteString(s)
	buf.W.WriteByte('\n')
}
func (buf *Buf) P(format string, args ...interface{}) {
	fmt.Fprintf(&buf.W, format, args...)
	buf.W.WriteByte('\n')
}
func (buf *Buf) String() string {
	return buf.W.String()
}

func (co *Compiler) EmitFunc(gd *GDef) {
	scope := &Scope{
		Name:   Format("Locals of %s", gd.FullName),
		GDefs:  make(map[string]*GDef),
		GDef:   gd,
		CGen:   co.CGen,
		Parent: co.CMod.Scope,
	}
	co.Locals = scope
	rec := gd.TV.(*FunctionTV).FuncRec
	co.P(rec.SignatureStr(gd.FullName))

	for i, in := range rec.Ins {
		var name string
		if in.Name != "" && in.Name != "_" {
			name = in.Name
		} else {
			name = Format("__%d", i)
		}

		local := &GDef{
			Name:     name,
			FullName: Format("in_%s", name),
			TV:       in.TV,
		}
		co.Locals.GDefs[name] = local
	}

	for i, out := range rec.Outs {
		var name string
		if out.Name != "" && out.Name != "_" {
			name = out.Name
		} else {
			name = Format("__%d", i)
		}

		local := &GDef{
			Name:     name,
			FullName: Format("(*out_%s)", name),
			TV:       out.TV,
		}
		co.Locals.GDefs[name] = local
	}

	if rec.FuncRecX.Body != nil {
		co.P("{\n")
		rec.FuncRecX.Body.VisitStmt(co)
		co.P("\n}\n")
	} else {
		co.P("; //EmitFunc: NATIVE\n")
	}
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
