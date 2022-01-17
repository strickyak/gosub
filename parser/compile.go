package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"reflect"
	"regexp"
	"strconv"
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

		log.Printf("?find %q? { %q ; %s }", name, ntv.name, stuff)
		if ntv.name == name {
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
	return Format("NaT{%q~%s@%s}", n.name, n.Expr.String(), n.Mod.Package)
}
func (n NameTV) String() string {
	return Format("NaT{%q~~%s}", n.name, n.TV)
}
func (r *StructRecX) String() string {
	return Format("structX %s", r.name)
}
func (r *InterfaceRecX) String() string {
	if r == nil {
		panic("L103")
	}
	return Format("interfaceX %s", r.name)
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
	return Format("struct<%s>", r.cname)
}
func (r *InterfaceRec) String() string {
	return Format("interface<%s>", r.name)
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
	if o.StructRecX == nil {
		return "StructTX[nil]"
	}
	return F("StructTX[%v]", o.StructRecX.String())
}
func (o *InterfaceTX) String() string {
	if o.InterfaceRecX == nil {
		return "InterfaceTX[nil]"
	}
	return F("InterfaceTX[%v]", o.InterfaceRecX.String())
}
func (o *FunctionTX) String() string { return Format("FunctionTX(%v)", o.FuncRecX) }

func CompileTX(v ExprVisitor, x NameTX, where Expr) NameTV {
	return NameTV{
		name: x.name,
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
		name:   p.name,
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
		// nil in InterfaceRecX means this is _any_ "interface empty"
		// We will not use nil in InterfaceRec for that.
		// Instead we use a synthetic Prim AnyTO.
		return &TypeVal{AnyTO}
	}

	numMeths := len(p.Meths)
	assert(numMeths > 0)
	z := &InterfaceRec{
		name:  p.name,
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
		L("250: Ins %d name %q expr %#v", i, e.name, e.Expr)
		z.Ins[i] = NameTV{e.name, e.Mod.VisitTypeExpr(e.Expr)}
	}
	for i, e := range x.Outs {
		L("250: Outs %d name %q expr %#v", i, e.name, e.Expr)
		z.Outs[i] = NameTV{e.name, e.Mod.VisitTypeExpr(e.Expr)}
	}
	return z
}

type TypeValue interface {
	String() string
	// Value
	// Intlike() bool // only on PrimTV
	CType() string
	// TypeOfHandle() (z string, ok bool)
	//< Assign(c string, typ TypeValue) (z string, ok bool)
	Cast(c string, typ TypeValue) (z string, ok bool)
	Equals(typ TypeValue) bool
	TypeCode() string
}

type PrimTV struct {
	name     string
	typecode string
}
type TypeTV struct {
	name string
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

func (tv *PrimTV) TypeCode() string      { return tv.typecode }
func (tv *TypeTV) TypeCode() string      { return "t" }
func (tv *PointerTV) TypeCode() string   { return "P" }
func (tv *SliceTV) TypeCode() string     { return "S" }
func (tv *MapTV) TypeCode() string       { return "M" }
func (tv *StructTV) TypeCode() string    { return "R" }
func (tv *InterfaceTV) TypeCode() string { return "I" }
func (tv *FunctionTV) TypeCode() string  { return "F" }
func (tv *MultiTV) TypeCode() string     { return "?" }

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
	return strings.Title(tv.name)
}
func (tv *TypeTV) ToC() string {
	return Format("ZType(%s)", tv.name)
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
	return Format("ZStruct(%s)", tv.StructRec.cname)
}
func (tv *InterfaceTV) ToC() string {
	return Format("ZInterface(%s/%d)", tv.InterfaceRec.name, len(tv.InterfaceRec.Meths))
}
func (tv *FunctionTV) ToC() string {
	return Format("ZFunction(%s)", tv.FuncRec.SignatureStr("(*)"))
}

func (tv *MultiTV) ToC() string {
	return Format("ZMulti(...)")
}

func (o *PrimTV) Intlike() bool {
	switch o.name {
	case "byte", "int", "uint":
		return true
	}
	return false
}

func (o *FunctionTV) Equals(typ TypeValue) bool {
	switch t := typ.(type) {
	case *FunctionTV:
		return reflect.DeepEqual(o, t)
	}
	return false
}
func (o *TypeTV) Equals(typ TypeValue) bool {
	switch t := typ.(type) {
	case *TypeTV:
		return o.name == t.name
	}
	return false
}
func (o *PrimTV) Equals(typ TypeValue) bool {
	switch t := typ.(type) {
	case *PrimTV:
		return o.name == t.name
	}
	return false
}
func (o *SliceTV) Equals(typ TypeValue) bool {
	switch t := typ.(type) {
	case *SliceTV:
		return o.E.Equals(t.E)
	}
	return false
}
func (o *PointerTV) Equals(typ TypeValue) bool {
	switch t := typ.(type) {
	case *PointerTV:
		return o.E.Equals(t.E)
	}
	return false
}
func (o *MapTV) Equals(typ TypeValue) bool {
	switch t := typ.(type) {
	case *MapTV:
		return o.K.Equals(t.K) && o.V.Equals(t.V)
	}
	return false
}
func (o *StructTV) Equals(typ TypeValue) bool {
	switch t := typ.(type) {
	case *StructTV:
		return o.StructRec.cname == t.StructRec.cname
	}
	return false
}
func (o *InterfaceTV) Equals(typ TypeValue) bool {
	switch t := typ.(type) {
	case *InterfaceTV:
		return o.InterfaceRec.name == t.InterfaceRec.name
	}
	return false
}
func (o *MultiTV) Equals(typ TypeValue) bool {
	panic("cannot compare MultiTV")
}

func (o *PointerTV) TypeOfHandle() (z string, ok bool) {
	if st, ok := o.E.(*StructTV); ok {
		return st.StructRec.cname, true
	}
	return "", false
}

func (o *PrimTV) CType() string      { return "P_" + o.name }
func (o *SliceTV) CType() string     { return F("Slice_(%s)", o.E.CType()) }
func (o *MapTV) CType() string       { return F("Map_(%s,%s)", o.K.CType(), o.V.CType()) }
func (o *StructTV) CType() string    { return F("struct %s", o.StructRec.cname) }
func (o *PointerTV) CType() string   { return F("%s*", o.E.CType()) }
func (o *InterfaceTV) CType() string { return F("Interface_(%s)", o.InterfaceRec.name) }
func (o *TypeTV) CType() string      { return "Type" }
func (o *MultiTV) CType() string     { return "Multi" }

func (o *FunctionTV) CType() string { return o.FuncRec.PtrTypedef }

func (co *Compiler) ConvertTo(from Value, to Value) {
	if from.Type().Equals(to.Type()) {
		// Same type, just assign.
		co.P("&%s = %s; // L451", to.ToC(), from.ToC())
		return
	}

	if from.Type() == ConstIntTO {
		switch to.Type() {
		case ByteTO:
			co.P("%s = (P_byte)(%s);", to.ToC(), from.ToC())
			return
		case IntTO:
			co.P("%s = (P_int)(%s);", to.ToC(), from.ToC())
			return
		case UintTO:
			co.P("%s = (P_uint)(%s);", to.ToC(), from.ToC())
			return
		}
	}

	// Case of assigning to interface{}.
	if to.Type() == AnyTO {

		if from.Type() == ConstIntTO {
			// Cannot take address of integer literal, so create a tmp var for ConstInt case.
			ser := Serial("constint")
			from = co.DefineLocalTemp(ser, IntTO, from.ToC())
		}

		dest := to.ToC()
		co.P("%s.pointer = &%s; // L458", dest, from.ToC())
		co.P("%s.typecode = %q; // L459", dest, from.Type().TypeCode())
		return
	}
	panic(F("Cannot assign: %v := %v", to, from))
}

func (o *TypeTV) XXX_Assign(c string, typ TypeValue) (z string, ok bool) {
	panic("cannot XXX_Assign to _type_")
}
func (o *FunctionTV) XXX_Assign(c string, typ TypeValue) (z string, ok bool) {
	panic("cannot XXX_Assign to func")
}
func (o *MultiTV) XXX_Assign(c string, typ TypeValue) (z string, ok bool) {
	panic("cannot XXX_Assign to Multi")
}

func (o *PrimTV) XXX_Assign(c string, typ TypeValue) (z string, ok bool) {
	if o.Equals(typ) {
		return c, true
	} else {
		return "", false
	}
}
func (o *SliceTV) XXX_Assign(c string, typ TypeValue) (z string, ok bool) {
	if o.Equals(typ) {
		return c, true
	} else {
		return "", false
	}
}
func (o *MapTV) XXX_Assign(c string, typ TypeValue) (z string, ok bool) {
	if o.Equals(typ) {
		return c, true
	} else {
		return "", false
	}
}
func (o *StructTV) XXX_Assign(c string, typ TypeValue) (z string, ok bool) {
	if o.Equals(typ) {
		return c, true
	} else {
		return "", false
	}
}
func (o *PointerTV) XXX_Assign(c string, typ TypeValue) (z string, ok bool) {
	if o.Equals(typ) {
		return c, true
	} else {
		return "", false
	}
}
func (o *InterfaceTV) XXX_Assign(c string, typ TypeValue) (z string, ok bool) {
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
	if o.Equals(typ) {
		return c, true
	} else {
		return "", false
	}
}
func (o *MapTV) Cast(c string, typ TypeValue) (z string, ok bool) {
	if o.Equals(typ) {
		return c, true
	} else {
		return "", false
	}
}
func (o *StructTV) Cast(c string, typ TypeValue) (z string, ok bool) {
	if o.Equals(typ) {
		return c, true
	} else {
		return "", false
	}
}
func (o *PointerTV) Cast(c string, typ TypeValue) (z string, ok bool) {
	if o.Equals(typ) {
		return c, true
	} else {
		return "", false
	}
}
func (o *InterfaceTV) Cast(c string, typ TypeValue) (z string, ok bool) {
	if o.Equals(typ) {
		return c, true
	} else {
		return "", false
	}
}

func (o *PrimTV) Cast(c string, typ TypeValue) (z string, ok bool) {
	if o.Equals(typ) {
		return c, true
	}
	if other, ok := typ.(*PrimTV); ok {
		if o.Intlike() && other.Intlike() {
			return Format("(P_%s)(%s)", o.name, c), true
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

func (tv *PrimTV) String() string    { return Format("PrimTV(%q)", tv.name) }
func (tv *TypeTV) String() string    { return Format("TypeTV(%q)", tv.name) }
func (tv *PointerTV) String() string { return Format("PointerTV(%v)", tv.E) }
func (tv *SliceTV) String() string   { return Format("SliceTV(%v)", tv.E) }
func (tv *MapTV) String() string     { return Format("MapTV(%v=>%v)", tv.K, tv.V) }

func (tv *StructTV) String() string {
	return Format("StructTV(%v)", tv.StructRec.cname)
}
func (tv *InterfaceTV) String() string {
	if tv.InterfaceRec.Meths == nil {
		return Format("TV:Any")
	} else {
		return Format("InterfaceTV(%s/%d)", tv.InterfaceRec.name, len(tv.InterfaceRec.Meths))
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

type NameAndExpr struct {
	name string
	expr Expr
	cmod *CMod
}
type ConstructorX struct {
	typeX Expr
	inits []NameAndExpr
}

func (o *ConstructorX) String() string {
	return fmt.Sprintf("Ctor(%v)", o.typeX)
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
	Container Expr
	Subscript Expr
}

func (o *SubX) String() string {
	return fmt.Sprintf("Sub(%s; %s)", o.Container, o.Subscript)
}
func (o *SubX) VisitExpr(v ExprVisitor) Value {
	return v.VisitSub(o)
}

/////////// Stmt

type StmtVisitor interface {
	VisitAssign(*AssignS)
	VisitVar(*VarStmt)
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
	return fmt.Sprintf("\nAssign{=%v <%q> %v=}\n", o.A, o.Op, o.B)
}

func (o *AssignS) VisitStmt(v StmtVisitor) {
	v.VisitAssign(o)
}

type VarStmt struct {
	name string
	tx   Expr
}

func (o *VarStmt) String() string {
	return fmt.Sprintf("\nVarS{=%v %v=}\n", o.name, o.tx)
}

func (o *VarStmt) VisitStmt(v StmtVisitor) {
	v.VisitVar(o)
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
		P(&b, "%s in_%s", nat.TV.CType(), SerialIfEmpty(nat.name))
	}
	if len(r.Outs) != 1 {
		for i, nat := range r.Outs {
			if i > 0 || len(r.Ins) > 0 {
				b.WriteByte(',')
			}
			L("out [%d]: %s", i, nat.TV.CType())
			P(&b, "%s *out_%s", nat.TV.CType(), SerialIfEmpty(nat.name))
		}
	}
	b.WriteByte(')')
	sigStr := b.String()
	L("SignatureStr: %s", sigStr)
	return sigStr
}

type StructRecX struct {
	name   string
	Fields []NameTX
	Meths  []NameTX
}

type InterfaceRecX struct {
	name  string
	Meths []NameTX
}

type StructRec struct {
	name   string
	cname  string
	Fields []NameTV
	Meths  []NameTV
}

type InterfaceRec struct {
	name  string
	cname string
	Mod   *CMod
	Meths []NameTV
}

type NameTX struct {
	name string
	Expr Expr
	Mod  *CMod
}
type NameTV struct {
	name string
	TV   TypeValue
}
type XXX_NameAndType struct {
	name    string
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
const XX_BoolType = "z"
const XX_ByteType = "b"
const XX_UintType = "u"
const XX_IntType = "i"
const XX_ConstIntType = "k"
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
var BoolTO = &PrimTV{name: "bool", typecode: "z"}
var ByteTO = &PrimTV{name: "byte", typecode: "b"}
var ConstIntTO = &PrimTV{name: "_const_int_", typecode: "k"}
var IntTO = &PrimTV{name: "int", typecode: "i"}
var UintTO = &PrimTV{name: "uint", typecode: "u"}
var UintptrTO = &PrimTV{name: "uintptr", typecode: "p"}
var StringTO = &PrimTV{name: "string", typecode: "s"}
var TypeTO = &PrimTV{name: "_type_", typecode: "t"}
var ListTO = &PrimTV{name: "_list_", typecode: "?"} // i.e. Multi-Value with `,`
var VoidTO = &PrimTV{name: "_void_", typecode: "v"}
var ImportTO = &PrimTV{name: "_import_", typecode: "?"}
var AnyTO = &PrimTV{name: "_any_", typecode: "a"} // i.e. `interface{}`
var NilTO = &PrimTV{name: "_nil_", typecode: "n"}

// A list of Type Objects to be installed.
var PrimTypeObjList = []*PrimTV{
	BoolTO,
	ByteTO,
	ConstIntTO,
	IntTO,
	UintTO,
	UintptrTO,
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
}

func (o *GDef) String() string {
	return F("GDef[%s pkg=%s cn=%s t=%v]", o.name, o.Package, o.CName, o.typeof)
}
func (o *GDef) ToC() string {
	return o.CName
}
func (o *GDef) Type() TypeValue {
	return o.typeof
}

func (o *GDef) ResolveAsTypeValue() (TypeValue, bool) {
	if o.istype != nil {
		return o.istype, true
	}
	return nil, false
}
func (o *CVal) ResolveAsTypeValue() (TypeValue, bool)           { return nil, false }
func (o *SubVal) ResolveAsTypeValue() (TypeValue, bool)         { return nil, false }
func (o *ImportVal) ResolveAsTypeValue() (TypeValue, bool)      { return nil, false }
func (o *BoundMethodVal) ResolveAsTypeValue() (TypeValue, bool) { return nil, false }
func (o *TypeVal) ResolveAsTypeValue() (TypeValue, bool)        { return o.tv, true }

func ResolveAsIntStr(v Value) (string, bool) {
	switch v.Type().TypeCode() {
	case "b", "i", "u", "k", "p":
		return v.ToC(), true
	}
	return "", false
}

type CVal struct {
	c string // C language expression
	t TypeValue
}

type SubVal struct {
	container Value
	subscript Value
}

type TypeVal struct {
	tv TypeValue
}

type BoundMethodVal struct {
	receiver Value
	cmeth    string
	mtype    TypeValue
}

func (val *CVal) String() string {
	return Format("(%s:%s)", val.c, val.t)
}
func (val *TypeVal) String() string {
	return Format("(%s:TypeVal)", val.tv)
}
func (val *SubVal) String() string {
	return Format("(%s[%s])", val.container, val.subscript)
}
func (val *BoundMethodVal) String() string {
	return Format("BM(%v ; %s)", val.receiver, val.cmeth)
}

func (val *CVal) Type() TypeValue {
	return val.t
}
func (val *TypeVal) Type() TypeValue {
	return TypeTO
}
func (val *SubVal) Type() TypeValue {
	switch t := val.container.Type().(type) {
	case *SliceTV:
		return t.E
	case *MapTV:
		return t.K
	}
	panic(F("L1169: cannot index into %v", val.container))
}
func (val *BoundMethodVal) Type() TypeValue {
	return val.mtype
}

func (val *CVal) ToC() string {
	return val.c
}
func (val *TypeVal) ToC() string {
	return F("%q", val.tv.TypeCode())
}
func (val *SubVal) ToC() string {
	ser := Serial("sub")
	switch t := val.container.Type().(type) {
	case *SliceTV:
		nth, ok := ResolveAsIntStr(val.subscript)
		if !ok {
			panic(F("slice subscript must be integer; got %v", val.subscript))
		}
		tmp := coHack.DefineLocalTemp(ser, t.E, "")
		// void SliceGet(Slice a, int size, int nth, void* value);
		coHack.P("SliceGet(%s, sizeof(%s), %s, &%s); //L1187",
			val.container.ToC(),
			t.E.CType(),
			nth,
			tmp.ToC())
		return F("(%s /*L1196*/)", tmp.ToC())
	case *MapTV:
		panic(1198)
	}
	panic(F("L1188: cannot index into %v", val.container))
}
func (val *BoundMethodVal) ToC() string {
	panic(1169) // It's not that simple.
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
	pr(`#include "runt.h"`)
	pr(``)
	if !opt.SkipBuiltin {
		cg.LoadModule("builtin", pr)
		cg.LoadModule("io", pr) // TODO: archive os__File__Read
	}
	p := NewParser(r, sourceName)
	p.ParseModule(cm, cg)

	cm.VisitGlobals(p, pr)
}

func (cm *CMod) defineOnce(g *GDef) {
	if _, ok := cm.Members[g.name]; ok {
		Panicf("module %s: redefined name: %s", cm.Package, g.name)
	}
	cm.Members[g.name] = g
	g.Package = cm.Package
	if g.name == "init" {
		ser := Serial("initmod")
		g.CName = CName(g.Package, ser)
		cm.initFuncs = append(cm.initFuncs, g.CName)
	} else {
		g.CName = CName(g.Package, g.name)
	}
}

func (cm *CMod) FirstSlotGlobals(p *Parser, pr printer) {
	pr("#define USING_MODULE_%s", cm.Package)
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
		cm.CGen.LoadModule(g.name, pr)
		g.typeof = ImportTO
		// If we care to do imports in order,
		// this is a good place to remember it.
		cm.CGen.ModsInOrder = append(cm.CGen.ModsInOrder, g.name)
	}
	for _, g := range p.Types {
		Say(g.Package, g.name, "2T")
		qc := cm.QuickCompiler(g)
		tmpX := NameTX{g.name, g.initx, cm}
		tmpV := CompileTX(qc, tmpX, g.initx)
		if tmpV.TV == nil {
			panic(g.CName)
		}
		g.istype = tmpV.TV
		g.typeof = TypeTO
		// Annotate structs & interfaces with their module.
		switch t := g.istype.(type) {
		case *StructTV:
			cname := CName(cm.Package, t.StructRec.name)
			t.StructRec.cname = cname
			cg := cm.CGen
			if _, already := cg.classNums[cname]; already {
				panic(F("struct already defined: %s", cname))
			}
			num := len(cg.classes)
			cg.classNums[cname] = num
			cg.classes = append(cg.classes, cname)
			pr("#define CLASS_%s %d", cname, num)
			pr("struct %s; // L1334", cname, cname)
		case *InterfaceTV:
			t.InterfaceRec.cname = CName(cm.Package, t.InterfaceRec.name)
		}
	}
	for _, g := range p.Consts {
		Say(g.Package, g.name, "2C")
		// not allowing g.Type on constants.
		g.constval = g.initx.VisitExpr(cm.QuickCompiler(g))
	}
	for _, g := range p.Vars {
		Say(g.Package, g.name, "2V")
		val := g.typex.VisitExpr(cm.QuickCompiler(g))
		tv, ok := val.ResolveAsTypeValue()
		if !ok {
			panic(F("got %#v when we wanted a TypeValue", val))
		}
		g.typeof = tv
		g.istype = nil // to be sure
		pr("extern %s %s; // L1320", g.typeof.CType(), g.CName)
	}
}

func (cm *CMod) StructRecOfReceiverOfFuncX(funcX *FunctionX) *StructRec {
	rec := funcX.FuncRecX
	assert(rec.IsMethod)
	assert(len(rec.Ins) > 0)
	rx := rec.Ins[0]
	r := rx.Expr.VisitExpr(rx.Mod.QuickCompiler(nil))
	Say("Receiver", r)
	Say("Receiver ToC", r.ToC())
	Say("Receiver Type", r.Type())
	tv, ok := r.ResolveAsTypeValue()
	if !ok {
		panic(F("L1309: expected a type for method receiver, but got %v", r))
	}
	pointerType, ok := tv.(*PointerTV)
	if !ok {
		panic(F("L1313: Expected pointer to struct as method receiver; got %v", r.Type()))
	}
	Say("pointerType", pointerType)
	structType, ok := pointerType.E.(*StructTV)
	if !ok {
		panic(F("L1318: Expected pointer to struct as method receiver; got %v", r.Type()))
	}
	Say(structType, structType)
	return structType.StructRec
}

func (cm *CMod) ThirdDefineGlobals(p *Parser, pr printer) {
	// Third: Define the globals, except for functions.
	say := func(how string, g *GDef) {
		pr("// Third == %s %s ==", how, g.CName)
	}
	_ = say
	for _, g := range p.Types {
		Say("Third Types: " + g.Package + " " + g.name)
		Say("type", V(g))
		Say("typeof", V(g.istype))
		//< pr("typedef %s %s;", g.istype.CType(), g.CName)
		if st, ok := g.istype.(*StructTV); ok {
			L("omg struct! %v :: %d fields %d meths", st, len(st.StructRec.Fields), len(st.StructRec.Meths))

			pr("#ifndef DEFINED_%s", g.CName)
			pr("struct %s {", g.CName)
			for _, field := range st.StructRec.Fields {
				L("omg field %q :: %v", field.name, field.TV)
				pr("  %s f_%s;", field.TV.CType(), field.name)
			}
			pr("}; // struct L1366")
			pr("#define DEFINED_%s 1", g.CName)
			pr("#endif")
		}
	}
	for _, g := range p.Vars {
		Say("Third Vars: " + g.Package + " " + g.name)
		pr("extern %s %s; //1441", g.typeof.CType(), g.CName)
	}
	for _, g := range p.Funcs {
		Say("Third Funcs: " + g.Package + " " + g.name)
		pr("// Func %q initx: %#v", g.CName, g.initx)
		co := cm.QuickCompiler(g)
		funcX := g.initx.(*FunctionX)
		g.typex = funcX
		funcRec := funcX.FuncRecX.VisitFuncRecX(co)
		g.typeof = &FunctionTV{funcRec}
	}
	for _, g := range p.Meths {
		Say("Third Meths: " + g.Package + " " + g.name)
		pr("// Meth %q initx: %#v", g.CName, g.initx)
		co := cm.QuickCompiler(g)
		funcX := g.initx.(*FunctionX)
		g.typex = funcX
		funcRec := funcX.FuncRecX.VisitFuncRecX(co)
		g.typeof = &FunctionTV{funcRec}
		Say("Got Third Meths:", g)

		// Install meth on struct.
		structRec := cm.StructRecOfReceiverOfFuncX(funcX)
		structRec.Meths = append(structRec.Meths, NameTV{g.name, g.typeof})
		g.CName = CName(structRec.cname, g.name)
	}
}

func (cm *CMod) FourthInitGlobals(p *Parser, pr printer) {
	for _, g := range p.Funcs {
		Say("Fourth Func: " + g.Package + " " + g.name)
		co := cm.QuickCompiler(g)
		coHack = co
		co.P("// Start Fourth Func: " + g.Package + " " + g.name)
		co.EmitFunc(g, true /*justDeclare*/)
		pr("\n%s\n", co.Buf.String())
		co.P("// Finish Fourth Func: " + g.Package + " " + g.name)
	}

	for _, g := range p.Meths {
		Say("Fourth Meth: " + g.Package + " " + g.name)
		co := cm.QuickCompiler(g)
		coHack = co
		co.P("// Start Fourth Meth: " + g.Package + " " + g.name)
		co.EmitFunc(g, true /*justDeclare*/)
		pr("\n%s\n", co.Buf.String())
		co.P("// Finish Fourth Meth: " + g.Package + " " + g.name)
	}
}

var cmHack *CMod
var coHack *Compiler
var prHack printer

type FilePrinter struct {
	w *os.File
}

func NewFilePrinter(filename string) *FilePrinter {
	w, err := os.Create(filename)
	if err != nil {
		panic(F("cannot open %q: %v", filename, err))
	}
	return &FilePrinter{w}
}

func (fp *FilePrinter) GetPrinter() func(string, ...interface{}) {
	pr := func(format string, args ...interface{}) {
		fmt.Fprintf(fp.w, format+"\n", args...)
	}
	return pr
}
func (fp *FilePrinter) Close() {
	fp.w.Close()
}

func (cm *CMod) FifthPrintFunctions(p *Parser, pr printer) {
	cmHack = cm
	prHack = pr

	for _, g := range p.Vars {
		Say("Fifth Var " + g.Package + " " + g.name)
		fp := NewFilePrinter(F("___.var.%s.c", g.CName))
		pr := fp.GetPrinter()
		pr("// Fifth FUNC: %T %s %q;", "#", "#", g.CName)
		pr(`#include "___.defs.h"`)
		pr("%s %s; //1443", g.typeof.CType(), g.CName)
		if g.initx != nil {
			ser := Serial("initvar")
			cname := CName(g.CName, ser)
			//pr("void %s__%s() { // L1530", g.CName, ser)

			// We are writing the global init() function.
			initS := &AssignS{
				A:  []Expr{&IdentX{g.name, cm}},
				Op: "=",
				B:  []Expr{g.initx},
			}
			funcX := &FunctionX{
				&FuncRecX{
					Body: &Block{
						locals: make(map[string]*GDef),
						stmts:  []Stmt{initS},
					},
				},
			}
			co := cm.QuickCompiler(g)
			gdef := &GDef{
				name:   "initvar",
				CName:  cname,
				initx:  funcX,
				typex:  funcX,
				typeof: &FunctionTV{funcX.FuncRecX.VisitFuncRecX(co)},
			}

			coHack = co
			co.EmitFunc(gdef, false /*justDeclare*/)
			pr("\n%s\n", co.Buf.String())
			//pr("}")
		}
	}

	for _, g := range p.Funcs {
		Say("Fifth Func " + g.Package + " " + g.name)
		fp := NewFilePrinter(F("___.func.%s.c", g.CName))
		pr := fp.GetPrinter()
		pr("// Fifth FUNC: %T %s %q;", "#", "#", g.CName)
		pr(`#include "___.defs.h"`)
		if g.initx != nil {
			co := cm.QuickCompiler(g)
			coHack = co
			co.EmitFunc(g, false /*justDeclare*/)
			pr("\n%s\n", co.Buf.String())
		} else {
			pr("// Cannot print function without body -- it must be extern.")
		}
		fp.Close()
	}

	for _, g := range p.Meths {
		Say("Fifth Meth " + g.Package + " " + g.name)
		fp := NewFilePrinter(F("___.meth.%s.c", g.CName))
		pr := fp.GetPrinter()
		pr("// Fifth METH: %T %s %q;", "#", "#", g.CName)
		pr(`#include "___.defs.h"`)
		if g.initx != nil {
			co := cm.QuickCompiler(g)
			coHack = co
			co.EmitFunc(g, false /*justDeclare*/)
			pr("\n%s\n", co.Buf.String())
		} else {
			pr("// Cannot print method without body -- it must be extern.")
		}
		fp.Close()
	}

	{
		fp := NewFilePrinter(F("___.initmod.%s.c", cm.Package))
		pr := fp.GetPrinter()
		pr(`#include "___.defs.h"`)
		pr("void initmod_%s() {", cm.Package)

		for _, funcName := range cm.initFuncs {
			pr("extern void %s();", funcName)
			pr("%s();", funcName)
		}

		pr("}")
		fp.Close()
	}

	if cm.Package == "main" {
		fp := NewFilePrinter("___.initmods.c")
		pr := fp.GetPrinter()
		pr(`#include "___.defs.h"`)
		pr("void initmods() {")

		for _, mod := range cm.CGen.ModsInOrder {
			pr("extern void initmod_%s();", mod)
			pr("initmod_%s();", mod)
		}
		pr("extern void initmod_main();")
		pr("initmod_main();")

		pr("}")
		fp.Close()
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
	name    string
	CName   string

	initx Expr // for Const or Var or Type
	typex Expr // for Const or Var or Func

	istype   TypeValue
	typeof   TypeValue
	constval Value
}

type Scope interface {
	Find(string) *GDef
}

type CMod struct { // isa Scope
	Package   string
	CGen      *CGen
	Members   map[string]*GDef
	initFuncs []string
}

func (cm *CMod) Find(s string) *GDef {
	// just debug log:
	assert(cm != nil)
	L("Searching %q for %q .......", cm.Package, s)
	for debug_k, debug_v := range cm.Members {
		L("....... debug %q %v", debug_k, debug_v)
	}
	L(".......")

	if d, ok := cm.Members[s]; ok {
		return d
	}

	switch cm.Package {
	case "":
		panic(F("Cannot find %q", s))
	case "builtin":
		return cm.CGen.Prims.Find(s)
	default:
		return cm.CGen.BuiltinMod.Find(s)
	}
}

type CGen struct {
	Mods        map[string]*CMod // by module name
	BuiltinMod  *CMod
	Prims       *CMod
	ModsInOrder []string
	Options     *Options
	W           *bufio.Writer

	classes   []string
	classNums map[string]int
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
		classes: []string{
			"_FREE_", "_BYTES_", "_HANDLES_",
		},
		classNums: make(map[string]int),
	}
	cg.Prims = &CMod{
		Package: "", // Use empty package name for Prims.
		CGen:    cg,
		Members: make(map[string]*GDef),
	}
	mainMod.CGen = cg

	// Populate PrimDog
	for _, e := range PrimTypeObjList {
		cg.Prims.Members[e.name] = &GDef{
			name:   e.name,
			CName:  "P_" + e.name,
			istype: e,
			typeof: TypeTO,
			Used:   false,
		}
	}
	cg.Prims.Members["nil"] = NIL
	cg.Prims.Members["true"] = TRUE
	cg.Prims.Members["false"] = FALSE

	return cg, mainMod
}

var NIL = &GDef{
	name:   "nil",
	CName:  "P_nil",
	typeof: NilTO,
}
var TRUE = &GDef{
	name:   "true",
	CName:  "P_true",
	typeof: BoolTO,
}
var FALSE = &GDef{
	name:   "false",
	CName:  "P_false",
	typeof: BoolTO,
}

func (cm *CMod) VisitTypeExpr(x Expr) TypeValue {
	L("VisitTypeExpr: %v", x)
	val := x.VisitExpr(NewCompiler(cm, nil))
	if tv, ok := val.ResolveAsTypeValue(); ok {
		return tv
	} else {
		log.Panicf("Expected expr [ %v ] to compile to TypeValue, but it compiled to %T :: [ %v ]", x, val, val)
		panic(0)
	}
}
func (cm *CMod) VisitExpr(x Expr) Value {
	return x.VisitExpr(NewCompiler(cm, nil))
}
func (cm *CMod) QuickCompiler(gdef *GDef) *Compiler {
	return NewCompiler(cm, gdef)
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
	slots        map[string]*GDef // not really G
	classes      []string
}

func NewCompiler(cm *CMod, subject *GDef) *Compiler {
	co := &Compiler{
		CMod:    cm,
		CGen:    cm.CGen,
		Subject: subject,
		Buf:     &Buf{},
		slots:   make(map[string]*GDef),
	}
	return co
}

func (co *Compiler) P(format string, args ...interface{}) {
	co.Buf.P(format, args...)
	co.Buf.P("\n")
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
		c: Format("MakeStringFromC(%q)", x.X),
		t: StringTO,
	}
}
func (co *Compiler) VisitIdent(x *IdentX) Value {
	L("VisitIdent: %s", x.X)
	return co.FindName(x.X)
}
func (co *Compiler) VisitBinOp(x *BinOpX) Value {
	a := x.A.VisitExpr(co)
	op := x.Op
	b := x.B.VisitExpr(co)
	var resultType TypeValue

	L("BinOp: a = %#v", a)
	L("BinOp: b = %#v", b)
	L("BinOp: a.ToC = %s :: %s", a.ToC(), a.Type())
	L("BinOp: b.ToC = %s :: %s", b.ToC(), b.Type())

	switch op {
	case "==", "!=", "<", "<=", ">", ">=":
		switch ta := a.Type().(type) {
		case *InterfaceTV:
			switch tb := b.Type().(type) {
			case *InterfaceTV:
				return &CVal{
					c: Format("(/*L1810*/(%s) %s (%s))", a.ToC(), op, b.ToC()),
					t: BoolTO,
				}

			case *PointerTV:
				return &CVal{
					c: Format("(/*L1816*/(%s) %s (%s))", a.ToC(), op, b.ToC()),
					t: BoolTO,
				}
			case *PrimTV:
				switch tb.typecode {
				case "n": // nil
					return &CVal{
						c: Format("(/*L1823*/(%s) %s (void*)0)", a.ToC(), op),
						t: BoolTO,
					}

				}
			}
		case *PointerTV:
			switch b.Type().(type) {
			case *InterfaceTV:
				return &CVal{
					c: Format("(/*L1833*/(%s) %s (%s))", a.ToC(), op, b.ToC()),
					t: BoolTO,
				}
			}

		case *PrimTV:
			switch ta.typecode {
			case "n": // nil
				switch b.Type().(type) {
				case *InterfaceTV:
					return &CVal{
						c: Format("(/*L1844*/(void*)0 %s (%s))", op, b.ToC()),
						t: BoolTO,
					}
				case *PointerTV:
					return &CVal{
						c: Format("(/*L1849*/(void*)0 %s (%s))", op, b.ToC()),
						t: BoolTO,
					}
				}

			}
		}
	}

	if a.Type().TypeCode() == "k" {
		switch b.Type().TypeCode() {
		case "b", "i", "u", "p":
			a = &CVal{a.ToC(), b.Type()}
		case "k":
			// Both a and b are ConstInt: return a computed ConstInt.
			switch op {
			case "+":
				return KVal(EvalK(a) + EvalK(b))
			case "-":
				return KVal(EvalK(a) - EvalK(b))
			case "*":
				return KVal(EvalK(a) * EvalK(b))
			case "/":
				return KVal(EvalK(a) / EvalK(b))
			case "%":
				return KVal(EvalK(a) % EvalK(b))
			case "&":
				return KVal(EvalK(a) & EvalK(b))
			case "|":
				return KVal(EvalK(a) | EvalK(b))
			case "^":
				return KVal(EvalK(a) ^ EvalK(b))

			case "<<":
				return KVal(EvalK(a) << EvalK(b))
			case ">>":
				return KVal(EvalK(a) >> EvalK(b)) // TODO: signed vs unsigned

			case "==":
				return BVal(EvalK(a) == EvalK(b))
			case "!=":
				return BVal(EvalK(a) != EvalK(b))
			case "<":
				return BVal(EvalK(a) < EvalK(b))
			case ">":
				return BVal(EvalK(a) > EvalK(b))
			case "<=":
				return BVal(EvalK(a) <= EvalK(b))
			case ">=":
				return BVal(EvalK(a) >= EvalK(b))
			}
		}
	}

	if b.Type().TypeCode() == "k" {
		switch a.Type().TypeCode() {
		case "b", "i", "u", "p":
			b = &CVal{b.ToC(), a.Type()}
		}
	}

	if a.Type().Equals(b.Type()) {
		switch a.Type().TypeCode() {
		case "z", "b", "i", "u", "p":
			switch op {
			case "+", "-", "*", "/", "%":
				resultType = a.Type()
			case "==", "!=", "<", "<=", ">", ">=":
				resultType = BoolTO
			}
			return &CVal{
				c: Format("(%s)(/*L1920*/(%s) %s (%s))", resultType.CType(), a.ToC(), op, b.ToC()),
				t: resultType,
			}
		}
	}
	panic(1824)
}

func EvalK(a Value) int64 {
	s := a.ToC()
	base := 10
	if strings.HasPrefix(s, "0") {
		base = 8
	}
	if strings.HasPrefix(s, "0x") {
		base = 16
		s = s[2:]
	}
	z, err := strconv.ParseInt(s, base, 64)
	if err != nil {
		panic(F("EvalK cannot parse %q: %v", s, err))
	}
	return z
}
func KVal(x int64) Value {
	return &CVal{F("%d", x&0xFFFF), ConstIntTO}
}
func BVal(b bool) Value {
	if b {
		return TRUE
	} else {
		return FALSE
	}
}

func (co *Compiler) VisitConstructor(ctorX *ConstructorX) Value {
	tv := ctorX.typeX.VisitExpr(co)
	g, ok := tv.(*GDef)
	if !ok {
		panic(F("L1767: Constructor must be for address of struct name: %s", tv))
	}

	if g.istype == nil {
		panic(F("L1760: Constructor must be for address of struct: %s", g.CName))
	}
	structTV, ok := g.istype.(*StructTV)
	if !ok {
		panic(F("L1764: Constructor must be for address of struct: %s", g.CName))
	}

	pointerTV := &PointerTV{structTV}
	ser := Serial("ctor")
	creation := F("(struct %s*) oalloc(sizeof(struct %s), CLASS_%s)", g.CName, g.CName, g.CName)
	inst := co.DefineLocalTemp(ser, pointerTV, creation)

	for i, e := range ctorX.inits {
		val := e.expr.VisitExpr(co)
		co.P("  %s->f_%s = %s; //#%d L2001", inst.CName, e.name, val.ToC(), i)
	}

	return inst
}
func (co *Compiler) VisitFunction(funcX *FunctionX) Value {
	L("VisitFunction: FuncRecX = %#v", funcX.FuncRecX)
	funcRec := funcX.FuncRecX.VisitFuncRecX(co)
	L("VisitFunction: FuncRec = %#v", funcRec)
	t := &FunctionTV{funcRec}
	panic(2007)
	return &CVal{c: "?1702?", t: t}
}

var IDENTIFIER = regexp.MustCompile("^[A-Za-z0-9_]+$")

func (co *Compiler) Reify(x Value) Value {
	return co.ReifyAs(x, x.Type()) // as its own type
}
func (co *Compiler) ReifyAs(x Value, as TypeValue) Value {
	L("ccc x %v", x)
	L("ccc x.Type %v", x.Type())
	L("ccc as %v", as)
	if x.Type().Equals(as) {
		// CASE: types are same.  Easy.
		cVal := x.ToC()
		if match := IDENTIFIER.MatchString(cVal); match {
			return x
		}
		ser := Serial("reify")
		gd := co.DefineLocalTemp(ser, x.Type(), cVal)
		return gd
	}

	// CASE: types are different.
	// First, reify as the input type.
	reifiedX := co.Reify(x)

	// Then convert.
	ser := Serial("reify_as")
	y := co.DefineLocalTemp(ser, as, "")
	co.ConvertTo(reifiedX, y)
	return y
}

func (co *Compiler) VisitMake(args []Expr) Value {
	if len(args) < 1 || len(args) > 3 {
		panic("wrong number of args to `make`")
	}
	theType := args[0].VisitExpr(co)
	theLen := "0"
	var ok bool
	if len(args) >= 2 {
		a1 := args[1].VisitExpr(co)
		theLen, ok = ResolveAsIntStr(args[1].VisitExpr(co))
		if !ok {
			panic(F("expected integer for arg 1 of make; got %v", a1))
		}
	}
	theCap := "0"
	if len(args) >= 3 {
		a2 := args[2].VisitExpr(co)
		theCap, ok = ResolveAsIntStr(args[2].VisitExpr(co))
		if !ok {
			panic(F("expected integer for arg 2 of make; got %v", a2))
		}
	}
	tv, ok := theType.ResolveAsTypeValue()
	if !ok {
		panic("expected type as 1st arg to `make`")
	}
	switch t := tv.(type) {
	case *SliceTV:
		return &CVal{
			c: F("MakeSlice(%q, %s, %s, sizeof(%s))", t.E.TypeCode(), theLen, theCap, t.E.CType()),
			t: tv,
		}
	}
	panic(F("cannot `make` a %v", tv))
}
func (co *Compiler) VisitAppend(args []Expr, hasDotDotDot bool) Value {
	// TODO dotdotdot
	if len(args) != 2 {
		panic("append must have 2 args")
	}
	a0 := args[0].VisitExpr(co)
	a1 := args[1].VisitExpr(co)

	return &CVal{
		c: F("AppendSlice(%s, %s)", a0.ToC(), a1.ToC()),
		t: a0.Type(),
	}
}

func (co *Compiler) VisitCall(callx *CallX) Value {
	if identx, ok := callx.Func.(*IdentX); ok {
		if identx.X == "make" {
			return co.VisitMake(callx.Args)
		}
		if identx.X == "append" {
			// TODO dotdotdot
			return co.VisitAppend(callx.Args, false)
		}
	}

	// type CallX struct { Func Expr; Args []Expr; }
	ser := Serial("call")
	//< var prep []string
	if ser == "call_103" {
		panic(ser)
	}

	co.P("// This is VisitCall %s", ser)
	co.P("// This is VisitCall %s", ser)
	co.P("// Func is callx %#v", callx.Func)
	co.P("// Args is len %d", len(callx.Args))
	for i, a := range callx.Args {
		co.P("// Args[%d] is callx %#v", i, a)
	}

	L("callx = %v", callx)
	L("callx.Func = %v", callx.Func)
	funcVal := callx.Func.VisitExpr(co)
	L("funcVal = %v", funcVal)
	funcValType := funcVal.Type()
	L("funcValType = %v", funcValType)
	funcRec := funcValType.(*FunctionTV).FuncRec
	L("funcRec = %v", funcRec)
	fins := funcRec.Ins
	fouts := funcRec.Outs
	co.P("// IsMethod = %v", funcRec.IsMethod)
	co.P("// HasDotDotDot = %v", funcRec.HasDotDotDot)

	var bm *BoundMethodVal

	var argVals []Value
	if bm, _ = funcVal.(*BoundMethodVal); bm != nil {
		// Prepend receiver as first arg.
		argVals = append(argVals, bm.receiver)
	}
	for _, e := range callx.Args {
		argVals = append(argVals, e.VisitExpr(co))
	}

	co.P("// Func is V %#v", funcVal)
	for i, a := range argVals {
		co.P("// Args[%d] is V %#v", i, a)
	}

	var extraFin NameTV
	var extraSliceType *SliceTV
	var numNormal, numExtras int

	if funcRec.HasDotDotDot {
		if len(fins)-1 > len(argVals) {
			panic(F("got %d args for func call, wanted at least %d args", len(argVals), len(fins)-1))
		}
		numNormal = len(fins) - 1
		numExtras = len(argVals) - numNormal

		extraFin = fins[numNormal]
		fins = fins[:numNormal]
		extraSliceType = extraFin.TV.(*SliceTV)
	} else {
		if len(fins) != len(argVals) {
			panic(F("got %d args for func call, wanted %d args", len(argVals), len(fins)))
		}
	}

	co.P("// 1765: extraFin=%v ;; extraSliceType=%v ;; numNormal=%d ;; numExtras=%d", extraFin, extraSliceType, numNormal, numExtras)

	var argc []string

	// For the non-DotDotDot arguments
	for i, fin := range fins {
		temp := CName(ser, "in", D(i), fin.name)
		L("argVals[%d] :: %T = %q", i, argVals[i], argVals[i].ToC())
		gd := co.DefineLocalTemp(temp, fin.TV, argVals[i].ToC())
		argc = append(argc, gd.CName)
	}

	if funcRec.HasDotDotDot {
		sliceName := CName(ser, "in", "vec")
		sliceVar := co.DefineLocalTemp(sliceName, extraSliceType, "NilSlice /*MakeSlice L1949*/")

		for i := 0; i < numExtras; i++ {
			y := co.ReifyAs(argVals[numNormal+i], extraSliceType.E).ToC()

			co.P("%s = SliceAppend(%q, %s, &%s, sizeof(%s)); // L1954: For extra input #%d", sliceVar.CName, extraSliceType.E.TypeCode(), sliceVar.CName, y, y, i)
		}

		fins = append(fins, NameTV{sliceVar.CName, extraSliceType})
		argc = append(argc, sliceVar.CName)
	}

	if len(fouts) != 1 {
		var multi []NameTV
		for j, out := range fouts {
			rj := Format("_multi_%s_%d", ser, j)
			vj := NameTV{rj, out.TV}
			multi = append(multi, vj)
			gd := co.DefineLocalTemp(rj, out.TV, "")
			argc = append(argc, F("&%s", gd.CName))
		}
		c := co.FormatCall(funcVal, argc, bm)
		return &CVal{c: c, t: &MultiTV{multi}}
	} else {
		c := co.FormatCall(funcVal, argc, bm)
		t := fouts[0].TV
		return &CVal{c: c, t: t}
	}
}

func (co *Compiler) FormatCall(funcVal Value, argc []string, bm *BoundMethodVal) string {
	if bm == nil {
		return Format("(%s(%s)/*L1870*/)", funcVal.ToC(), strings.Join(argc, ", "))
	}
	return Format("(%s(%s)/*L1872*/)", bm.cmeth, strings.Join(argc, ", "))

}

func (co *Compiler) VisitSub(subx *SubX) Value {
	con := subx.Container.VisitExpr(co)
	sub := subx.Subscript.VisitExpr(co)

	return &SubVal{
		container: con,
		subscript: sub,
	}
}

func (co *Compiler) VisitDot(dotx *DotX) Value {
	log.Printf("VisitDot: <------ %v", dotx)
	// val := co.ResolveTypeOfValue(dotx.X.VisitExpr(co))
	val := dotx.X.VisitExpr(co)
	log.Printf("VisitDot: val-- %T ---- %v", val, val)

	if val.Type() == ImportTO {
		L("YES: %#v", val)
		gd := val.(*GDef)
		modName := gd.name

		if otherMod, ok := co.CGen.Mods[modName]; ok {
			log.Printf("OM %#v", otherMod)
			if x, ok := otherMod.Members[dotx.Member]; ok {
				L("member: %v", x)
				return x
			} else {
				panic(F("member not found: %q", dotx.Member))
			}
		}
	}

	if pointedType, ok := val.Type().(*PointerTV); ok {
		Say("pointedType", pointedType)
		if structType, ok := pointedType.E.(*StructTV); ok {
			Say("structType", structType)
			rec := structType.StructRec
			Say("rec", rec)
			Say("rec", F("%#v", rec))
			if ftype, ok := FindTypeByName(rec.Fields, dotx.Member); ok {
				z := &CVal{
					c: Format("(%s).%s", val.ToC(), dotx.Member),
					t: ftype,
				}
				L("VisitDot returns Field: %#v", z)
				return z
			}
			if mtype, ok := FindTypeByName(rec.Meths, dotx.Member); ok {
				if mtype == nil {
					panic(1893)
				}
				bm := &BoundMethodVal{
					receiver: val,
					cmeth:    CName(rec.cname, dotx.Member),
					mtype:    mtype,
				}
				L("VisitDot returns bound meth: %#v", bm)
				return bm
			}
		}
	}

	panic("DotXXX")
}

func (co *Compiler) VisitVar(v *VarStmt) {
	debug := co.DefineLocal("v", v.name, co.CMod.VisitTypeExpr(v.tx))
	L("debug VisitVar: %#v ==> %#v", *v, *debug)
}
func (co *Compiler) AssignSingle(left Value, right Value) {
	switch t := left.(type) {
	case *SubVal:
		switch t.container.Type().(type) {
		case *SliceTV:
			// slice, size, nth, value
			nth, ok := ResolveAsIntStr(t.subscript)
			if !ok {
				panic(F("slice subscript must be integer; got %v", t.subscript))
			}
			rleft := co.ReifyAs(right, UintTO)

			co.P(" SlicePut(%s, sizoef(%s), %s, &%s); L2071",
				t.container.ToC(),
				right.Type().CType(),
				nth,
				rleft.ToC())
			return

		case *MapTV:
			panic("TODO L2070")
		}
		panic(F("todo SubVal L1835: %v", t))
	default:
		co.P("%s = %s; // L1837", left.ToC(), right.ToC())
		return
	}
	panic("L2093")
}
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
	var newLocals []*GDef
	if ass.Op == ":=" {
		for i, a := range ass.A {
			if id, ok := a.(*IdentX); ok {
				var name string
				if id.X != "" && id.X != "_" {
					name = id.X
				} else {
					name = Serial("tmp")
				}
				var lclType TypeValue
				if len(rvalues) == 1 {
					if t0, ok := rvalues[0].Type().(*MultiTV); ok {
						if len(ass.A) == len(t0.Multi) {
							lclType = t0.Multi[i].TV
						}
					} else {
						lclType = rvalues[0].Type()
					}
				} else {
					lclType = rvalues[i].Type()
				}
				if lclType == ConstIntTO {
					lclType = IntTO
				}
				gd := co.DefineLocal("v", name, lclType)
				newLocals = append(newLocals, gd)
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
		// CASE: An lvalue followed by ++ or --.
		if len(ass.A) != 1 {
			Panicf("operator %v requires one lvalue on the left, got %v", ass.Op, ass.A)
		}
		// TODO check lvalue
		cvar := ass.A[0].VisitExpr(co).ToC()
		co.P("  (%s)%s;", cvar, ass.Op)

	case ass.A == nil && bcall != nil:
		// CASE: No assignment.  Just a function call.
		callVal := rvalues[0]
		co.P("%s; // Call with no assign: L2001", callVal.ToC())

	case ass.A == nil && bcall == nil:
		// CASE: No assignment.  Just a non-function not allowed.
		panic(Format("L2005: Lone expr is not a function call: [%v]", ass.B))

	case len(ass.A) > 1 && bcall != nil:
		// CASE From 1 call, to 2 or more assigned vars.
		callVal := rvalues[0]
		mtv, ok := callVal.Type().(*MultiTV)
		if !ok {
			panic(F("L2011: When assigning to multi vars, expected call with %d results, but got %v", len(ass.A), callVal))
		}
		if len(ass.A) != len(mtv.Multi) {
			panic(F("L2014: When assigning to multi vars, expected call with %d results, but got %d results from %v", len(ass.A), len(mtv.Multi), callVal))
		}

		co.P("%s; // Call with multi assign: L2009", callVal.ToC())
		for i, dest := range ass.A {
			destVal := dest.VisitExpr(co)
			// TODO Sub
			// TODO type of destVal
			// TODO: Why am I adding tmp_ to this multi result?
			co.P("%s = tmp_%s; // Multi result [%d].  L2020", destVal.ToC(), mtv.Multi[i].name, i)
		}

	case len(ass.A) == 1 && len(ass.B) == 1:
		// CASE: Simple 1,1
		var target Value
		if newLocals != nil {
			target = newLocals[0]
		} else {
			target = ass.A[0].VisitExpr(co)
		}
		co.AssignSingle(target, rvalues[0])

	default:
		// CASE: more than 1: same: (right) == len(left)
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
	co.P("while(1) { Cont_%s: {}", label)
	if wh.Pred != nil {
		co.P("    bool _while_ = (bool)(%s);", wh.Pred.VisitExpr(co).ToC())
		co.P("    if (!_while_) break;")
	}
	savedB, savedC := co.BreakTo, co.ContinueTo
	co.BreakTo, co.ContinueTo = "Break_"+label, "Cont_"+label
	wh.Body.VisitStmt(co)
	co.P("  }")
	co.P("Break_%s: {}", label)
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
	co.P("  { bool _if_ = %s;", ifs.Pred.VisitExpr(co).ToC())
	co.P("  if( _if_ ) {")
	ifs.Yes.VisitStmt(co)
	if ifs.No != nil {
		co.P("  } else {")
		ifs.No.VisitStmt(co)
	}
	co.P("  }}")
}
func (co *Compiler) VisitSwitch(sws *SwitchS) {
	co.P("  { int _switch_ = %s;", sws.Switch.VisitExpr(co).ToC())
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
		panic("L2058")
	}
	for i, e := range a.stmts {
		ser := Serial("block")
		co.P("// @@ VisitBlock[%s,%d] <= %q", ser, i, F("%v", e))
		e.VisitStmt(co)
		log.Printf("VisitBlock[%d] ==>\n<<<\n%s\n>>>", i, co.Buf.String())
	}
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

func (co *Compiler) FindName(name string) *GDef {
	L("nando x: %q", name)
	if co.CurrentBlock != nil {
		L("nando z: %q [block %v]", name, co.CurrentBlock.locals)
		return co.CurrentBlock.Find(name)
	}
	L("nando y: %q [cmod @%q]", name, co.CMod.Package)
	return co.CMod.Find(name)
}
func (co *Compiler) DefineLocalTemp(tempName string, tempType TypeValue, initC string) *GDef {
	gd := co.DefineLocal("tmp", tempName, tempType)
	if initC != "" {
		co.P("%s = %s; // L1632", gd.CName, initC)
	}
	return gd
}

func (co *Compiler) DefineLocal(prefix string, name string, tv TypeValue) *GDef {
	cname := Format("%s_%s", prefix, name)
	local := &GDef{
		name:   name,
		CName:  cname,
		typeof: tv,
	}
	if _, ok := co.CurrentBlock.locals[name]; ok {
		panic(F("// Local var Already defined: %s", name))
	} else {
		co.CurrentBlock.locals[name] = local
		L("bilbo %q 111 Current Block Locals: %v", name, co.CurrentBlock.locals)
	}
	co.slots[local.CName] = local
	L("bilbo %q 222 Compiler slots: %v", name, co.slots)
	return local
}
func (co *Compiler) FinishScope() {
	co.CurrentBlock = co.CurrentBlock.parent
}
func (co *Compiler) StartScope() {
	ser := Serial("scope")
	co.P("// Starting Scope: %q", ser)
	block := &Block{
		debugName: ser,
		locals:    make(map[string]*GDef),
		parent:    co.CurrentBlock,
		compiler:  co,
	}
	co.CurrentBlock = block
}
func (co *Compiler) EmitFunc(gd *GDef, justDeclare bool) {
	co.StartScope()
	rec := gd.typeof.(*FunctionTV).FuncRec
	co.P(rec.SignatureStr(gd.CName))
	if justDeclare {
		co.FinishScope()
		co.P("; //L2432: justDeclare")
		return
	}

	// Figure out the names of Func inputs, and create locals for them.
	for i, in := range rec.Ins {
		var name string
		if in.name != "" && in.name != "_" {
			name = in.name
		} else {
			name = Format("__%d", i)
		}
		co.DefineLocal("in", name, in.TV)
	}

	// Figure out the names of Func outputs, and create locals for them.
	// Unless there is only one out -- then it is a direct return.
	if len(rec.Outs) > 1 {
		for i, out := range rec.Outs {
			var name string
			if out.name != "" && out.name != "_" {
				name = out.name
			} else {
				name = Format("__%d", i)
			}
			co.DefineLocal("out", name, out.TV)
		}
	}

	if rec.FuncRecX.Body == nil {
		// Function has no body, so it should be natively-defined.
		co.P("; //EmitFunc L2438: NATIVE\n")
		return
	}

	// For the normal case of a function with a body.
	// Emit the translated body onto a buffer.
	// Then define the local variables, and initialize them,
	// and emit them before you emit the body from the buffer.

	co.P("{\n")

	prevBuf := co.Buf
	co.Buf = &Buf{}

	rec.FuncRecX.Body.compiler = co // TODO ???
	rec.FuncRecX.Body.VisitStmt(co)
	cBody := co.Buf.String()

	co.Buf = prevBuf
	co.P("// Adding LOCALS to Func:")
	for name, e := range co.slots {
		if strings.HasPrefix(e.CName, "in_") || strings.HasPrefix(e.CName, "out_") {
			// These are declared in the formal params of the C function.
			continue
		}
		co.P("// LOCAL %q IS %v", name, e)
		co.P("auto %v %v = {0}; // DEF LOCAL L2145 Type=%#v", e.typeof.CType(), e.CName, e.typeof)
	}
	co.P("// Added LOCALS to Func.")
	L("CBODY IS %q", cBody)
	co.P("\n%s\n", cBody)

	co.FinishScope()
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
func (o *Nando) VisitDot(dotx *DotX) Value {
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
