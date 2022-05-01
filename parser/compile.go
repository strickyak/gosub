package parser

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
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
	VisitSubSlice(*SubSliceX) Value
	VisitDot(*DotX) Value
	VisitTypeAssert(*TypeAssertX) Value
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
func (o *SliceTX) String() string   { return Format("SliceTX(%v)", o.E) }
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
	CType() string
	Zero() string
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
func (tv *PointerTV) TypeCode() string   { return "P" + tv.E.TypeCode() }
func (tv *SliceTV) TypeCode() string     { return "S" + tv.E.TypeCode() }
func (tv *MapTV) TypeCode() string       { return "M" + tv.K.TypeCode() + tv.V.TypeCode() }
func (tv *StructTV) TypeCode() string    { return "R0" }
func (tv *InterfaceTV) TypeCode() string { return "I0" }
func (tv *MultiTV) TypeCode() string     { return "?" }
func (tv *FunctionTV) TypeCode() string {
	return tv.FuncRec.BuildTypeCode(false)
}

func (rec *FuncRec) BuildTypeCode(omitFirst bool) string {
	var buf bytes.Buffer
	if omitFirst {
		buf.WriteByte('F') // D for meth
	} else {
		buf.WriteByte('F') // F for func
	}
	for i, e := range rec.Ins {
		if omitFirst && i == 0 {
			continue
		}
		buf.WriteString(e.TV.TypeCode())
	}
	if rec.HasDotDotDot {
		buf.WriteByte('X')
	}
	buf.WriteByte('_')
	for _, e := range rec.Outs {
		buf.WriteString(e.TV.TypeCode())
	}
	buf.WriteByte('_')
	return buf.String()
}

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
	return Format("ZFunction(%s)", tv.FuncRec.SignatureStr("(*)", false))
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

func (o *PrimTV) Zero() string {
	switch o.typecode[0] {
	case 'z', 'b', 'k', 'i', 'u', 'p':
		return "0"
	case 's':
		return "{0, 0, 0}"
	case 'a':
		return "{0, 0}"
	}
	panic("Zero " + o.name)
}

// Zero

func (o *SliceTV) Zero() string     { return "{0, 0, 0}" }
func (o *MapTV) Zero() string       { return "(void*)0" }
func (o *StructTV) Zero() string    { panic("Zero Struct") }
func (o *PointerTV) Zero() string   { return "(void*)0" }
func (o *InterfaceTV) Zero() string { return "(void*)0" }
func (o *TypeTV) Zero() string      { panic("Zero Type") }
func (o *MultiTV) Zero() string     { panic("Zero Multi") }
func (o *FunctionTV) Zero() string  { return "(void*)0" }

// CType

func (o *PrimTV) CType() string      { return "P_" + o.name }
func (o *SliceTV) CType() string     { return F("Slice_(%s)", o.E.CType()) }
func (o *MapTV) CType() string       { return F("Map_(%s,%s)", o.K.CType(), o.V.CType()) }
func (o *StructTV) CType() string    { return F("struct %s", o.StructRec.cname) }
func (o *PointerTV) CType() string   { return F("%s*", o.E.CType()) }
func (o *InterfaceTV) CType() string { return F("Interface_(%s)", o.InterfaceRec.name) }
func (o *TypeTV) CType() string      { return "Type" }
func (o *MultiTV) CType() string     { return "Multi" }

func (o *FunctionTV) CType() string { return o.FuncRec.TODO_PtrTypedef }

func (co *Compiler) CastToType(from Value, toType TypeValue) Value {
	L("// CastToType: from %v to %v", from, toType)
	// Quick and Dirty int casts
	switch from.Type().TypeCode()[0] {
	case 'b', 'i', 'u', 'k', 'p':
		L("// CASE#A")
		switch toType.TypeCode()[0] {
		case 'b', 'i', 'u', 'k', 'p':
			L("// CASE#Z")
			z := co.DefineLocalTempC("cast", toType, "")
			L("// CastToType: from %v to %v", from, toType)
			co.P("%s = (%s)(%s); // L488 CastTo", z.CName, toType.CType(), from.ToC())
			return z
		}
	case 'S':
		L("// CASE#B")
		sliceT, ok := from.Type().(*SliceTV)
		assert(ok)
		if sliceT.E == ByteTO && toType.TypeCode() == "s" {
			// Convert []byte to string
			ser := Serial("from_bytes_to_string")
			str := co.DefineLocalTempC(ser, StringTO, "")
			co.P("%s = FromBytesToString(%s);", str.ToC(), from.ToC())
			return str
		}
	}
	panic(F("cannot Cast (yet?): %v TO TYPE %v", from, toType))
}
func (co *Compiler) ConvertTo(from Value, to Value) {
	co.ConvertToCNameType(from, to.ToC(), to.Type())
}
func (co *Compiler) ConvertToCNameType(from Value, toCName string, toType TypeValue) {
	if from.Type().Equals(toType) {
		// Same type, just assign.
		co.P("%s = %s; // L451", toCName, from.ToC())
		return
	}

	if from.Type() == ConstIntTO {
		switch toType {
		case ByteTO:
			co.P("%s = (P_byte)(%s);", toCName, from.ToC())
			return
		case IntTO:
			co.P("%s = (P_int)(%s);", toCName, from.ToC())
			return
		case UintTO:
			co.P("%s = (P_uint)(%s);", toCName, from.ToC())
			return
		}
	}

	// Case of assigning to interface{}.
	if toType == AnyTO {
		if from.Type() == ConstIntTO {
			// Cannot take address of integer literal, so create a tmp var for ConstInt case.
			ser := Serial("constint")
			from = co.DefineLocalTempC(ser, IntTO, from.ToC())
		}

		if from.Type() == AnyTO {
			// Skip the reify!
			dest := toCName
			co.P("%s.pointer = (%s).pointer; // L568", dest, from.ToC())
			co.P("%s.typecode = (%s).typecode; // L569", dest, from.ToC())
			return
		}

		rfrom := co.Reify(from)
		dest := toCName
		co.P("%s.pointer = &%s; // L458", dest, rfrom.ToC())
		co.P("%s.typecode = %q; // L459", dest, rfrom.Type().TypeCode())
		return
	}

	// Case of assigning to interface non-empty.
	if _, ok := toType.(*InterfaceTV); ok {
		if _, ok2 := from.Type().(*PointerTV); ok2 {
			// TODO -- check actual compatibility.
			co.P("%s = %s; // L501 [pointer to face]", toCName, from.ToC())
			return
		}
		if from.Type() == NilTO {
			// TODO -- check actual compatibility.
			co.P("%s = (void*)0; // L507 [nil to face]", toCName)
			return
		}
	}
	panic(F("Cannot assign: (%v :: %v) = %v", toCName, toType, from))
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
	Func         Expr
	Args         []Expr
	HasDotDotDot bool
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

type TypeAssertX struct {
	X Expr
	T Expr
}

func (o *TypeAssertX) String() string {
	return fmt.Sprintf("TypeAssertX(%s; %s)", o.X, o.T)
}
func (o *TypeAssertX) VisitExpr(v ExprVisitor) Value {
	return v.VisitTypeAssert(o)
}

type SubX struct {
	container Expr
	subscript Expr
}

func (o *SubX) String() string {
	return fmt.Sprintf("Sub(%s; %s)", o.container, o.subscript)
}
func (o *SubX) VisitExpr(v ExprVisitor) Value {
	return v.VisitSub(o)
}

type SubSliceX struct {
	container Expr
	a         Expr
	b         Expr
}

func (o *SubSliceX) String() string {
	return fmt.Sprintf("SubSlice(%s; %s; %s)", o.container, o.a, o.b)
}
func (o *SubSliceX) VisitExpr(v ExprVisitor) Value {
	return v.VisitSubSlice(o)
}

/////////// Stmt

type StmtVisitor interface {
	VisitAssign(*AssignS)
	VisitVar(*VarStmt)
	VisitWhile(*WhileS)
	VisitFor(*ForS)
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
	A       []Expr
	Op      string
	B       []Expr
	IsRange bool
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
	First Stmt
	Pred  Expr
	Next  Stmt
	Body  *Block
}

func (o *WhileS) String() string {
	return fmt.Sprintf("\nWhile(%#v)\n", o)
}

func (o *WhileS) VisitStmt(v StmtVisitor) {
	v.VisitWhile(o)
}

type ForS struct {
	Key   Expr
	Value Expr
	Coll  Expr
	Body  *Block
}

func (o *ForS) String() string {
	return fmt.Sprintf("\nFor(%#v)\n", o)
}

func (o *ForS) VisitStmt(v StmtVisitor) {
	v.VisitFor(o)
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
	Ins             []NameTX
	Outs            []NameTX
	HasDotDotDot    bool
	IsMethod        bool
	Body            *Block
	TODO_PtrTypedef string // global typedef of a pointer to this function type.
}

type FuncRec struct {
	Ins             []NameTV
	Outs            []NameTV
	HasDotDotDot    bool
	IsMethod        bool
	TODO_PtrTypedef string // global typedef of a pointer to this function type.
	FuncRecX        *FuncRecX
	gdef            *GDef // needed for dispatching method
}

func (r *FuncRec) SignatureStr(daFunc string, addReceiver bool) string {
	var b bytes.Buffer
	if len(r.Outs) == 1 {
		P(&b, "%s ", r.Outs[0].TV.CType())
	} else {
		P(&b, "void ")
	}
	started := false
	P(&b, "%s(", daFunc)
	if addReceiver {
		b.WriteString("void* receiver ")
		started = true
	}
	for _, nat := range r.Ins {
		if started {
			b.WriteByte(',')
			b.WriteByte(' ')
		}
		Say(nat)
		Say(nat.TV.CType())
		P(&b, "%s in_%s", nat.TV.CType(), SerialIfEmpty(nat.name))
		started = true
	}
	if len(r.Outs) != 1 {
		for i, nat := range r.Outs {
			if started {
				b.WriteByte(',')
				b.WriteByte(' ')
			}
			L("out [%d]: %s", i, nat.TV.CType())
			P(&b, "%s *out_%s", nat.TV.CType(), SerialIfEmpty(nat.name))
			started = true
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
	why      string           // debug name
	locals   map[string]*GDef // not really G
	stmts    []Stmt
	parent   *Block
	compiler *Compiler
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
	typecode string
	isFace   bool
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
	switch val.container.Type().TypeCode()[0] {
	case 's':
		return ByteTO
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
		tmp := coHack.DefineLocalTempC(ser, t.E, "")
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
	switch val.container.Type().TypeCode()[0] {
	case 's':
		nth, ok := ResolveAsIntStr(val.subscript)
		if !ok {
			panic(F("slice subscript must be integer; got %v", val.subscript))
		}
		tmp := coHack.DefineLocalTempC(ser, ByteTO, "")
		// void StringGet(String a, int nth, byte* value);
		coHack.P("StringGet(%s, %s, &%s); //L1262",
			val.container.ToC(),
			nth,
			tmp.ToC())
		return F("(%s /*L1267*/)", tmp.ToC())
	}
	panic(F("L1188: cannot index into %v", val.container))
}
func (val *BoundMethodVal) ToC() string {
	panic(1243) // It's not that simple.
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

func (cg *CGen) LoadModule(name string, pr printer) {
	log.Printf("LoadModule: << %q", name)
	if _, ok := cg.Mods[name]; ok {
		log.Printf("LoadModule: already loaded: %q", name)
		return
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
	return
}

type printer func(format string, args ...interface{})

func CompileToC(r io.Reader, sourceName string, w io.Writer, opt *Options) {
	pr := func(format string, args ...interface{}) {
		z := fmt.Sprintf(format, args...)
		log.Print("[[[[[  " + z + "  ]]]]]")
		fmt.Fprintf(w, "%s\n", z)
	}

	cg, cm := NewCGenAndMainCMod(opt, w)
	pr(`#include "runtime/runt.h"`)
	pr(``)
	if !opt.SkipBuiltin {
		cg.LoadModule("builtin", pr)
		cg.LoadModule("low", pr)
		cg.LoadModule("io", pr) // TODO: archive os__File__Read
		// cg.LoadModule("os", pr)
	}
	p := NewParser(r, sourceName)
	p.ParseModule(cm, cg)

	cm.VisitGlobals(p, pr)

	for _, line := range cg.dynamicDefs {
		pr("%s", line)
	}

	dmap := make(map[string][]*FuncRec)
	for _, gst := range cg.structs {
		srec := gst.istype.(*StructTV).StructRec
		for _, me := range srec.Meths {
			frec := me.TV.(*FunctionTV).FuncRec
			tc := frec.BuildTypeCode(true /*omitFirst*/)

			dspec := CName(me.name, tc)
			pr("// struct %q meth %q dspec %q", gst.CName, gst.name, dspec)

			if _, ok := cg.dmeths[dspec]; ok {
				pr("// YES %q ::: %v", dspec, gst.CName)
				dmap_dmeth, _ := dmap[dspec]
				dmap[dspec] = append(dmap_dmeth, frec)
				pr("// DMAP ::: %v", dmap)
			}

		}
	}
	for dspec, recs := range dmap {
		pr("// EmitDispatch ::: %q", dspec)
		cg.EmitDispatch(dspec, recs)
	}
}

func (cg *CGen) EmitDispatch(dspec string, recs []*FuncRec) {
	rt := F("rt_Dispatch__%s", dspec)
	s := F(`
#include "___.defs.h"
%s Dispatch__%s(void* p) {
  byte cls = ocls((word)p);
	switch (cls) {
`, rt, dspec)
	for _, rec := range recs {
		rcvr := rec.Ins[0]
		srec := rcvr.TV.(*PointerTV).E.(*StructTV).StructRec
		cnum, ok := cg.classNums[srec.cname]
		if !ok {
			L("WARNING: cannot find classNum for %q", srec.cname)
		}

		s += F("case %d: return (%s)%s;\n", cnum, rt, rec.gdef.CName)

	}
	s += F("}\n")
	s += F("panic_s(\"bad dispatch for %s\");", dspec)
	s += F("return 0;")
	s += F("}\n")

	filename := F("___.Dispatch.%s.c", dspec)
	err := ioutil.WriteFile(filename, []byte(s), 0777)
	if err != nil {
		panic(F("cannot WriteFile %q: %v", filename, err))
	}
}

func (cm *CMod) defineOnce(g *GDef) {
	if g.name == "init" {
		ser := Serial("init__mod_")
		g.name = ser
	}

	if _, ok := cm.Members[g.name]; ok {
		Panicf("module %s: redefined name: %s", cm.Package, g.name)
	}
	cm.Members[g.name] = g
	g.Package = cm.Package

	g.CName = CName(g.Package, g.name)
	if strings.HasPrefix(g.name, "init__mod__") {
		cm.initFuncs = append(cm.initFuncs, g.CName)
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
		if !SliceContainsString(cm.CGen.ModsInOrder, g.name) {
			// Only add it once.
			cm.CGen.ModsInOrder = append(cm.CGen.ModsInOrder, g.name)
		}
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
		L("L1465: 2V: %#v", g)
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

		switch tt := g.istype.(type) {
		case *InterfaceTV:
			cm.CGen.faces[g.CName] = g

		case *StructTV:
			cm.CGen.structs[g.CName] = g

			pr("#ifndef DEFINED_%s", g.CName)
			pr("struct %s {", g.CName)
			for _, field := range tt.StructRec.Fields {
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
		funcRec.gdef = g
		g.typeof = &FunctionTV{funcRec}
	}
	for _, g := range p.Meths {
		Say("Third Meths: " + g.Package + " " + g.name)
		pr("// Meth %q initx: %#v", g.CName, g.initx)
		co := cm.QuickCompiler(g)
		funcX := g.initx.(*FunctionX)
		g.typex = funcX
		funcRec := funcX.FuncRecX.VisitFuncRecX(co)
		funcRec.gdef = g
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

			// We are writing the global init() function.
			initS := &AssignS{
				A:  []Expr{&IdentX{g.name, cm}},
				Op: "=",
				B:  []Expr{g.initx},
			}
			funcX := &FunctionX{
				&FuncRecX{
					Body: &Block{
						why:    "initvar:" + g.CName,
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
		if g.initx == nil {
			continue // No need to generate empty .c files.
		}
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
		if g.initx == nil {
			continue // No need to generate empty .c files.
		}
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
		for _, funcName := range cm.initFuncs {
			pr("extern void %s();", funcName)
		}
		pr("void initmod_%s() {", cm.Package)

		for _, funcName := range cm.initFuncs {
			pr("%s();", funcName)
		}

		pr("}")
		fp.Close()
	}

	if cm.Package == "main" {
		fp := NewFilePrinter("___.initmods.c")
		pr := fp.GetPrinter()
		pr(`#include "___.defs.h"`)
		for _, mod := range cm.CGen.ModsInOrder {
			pr("extern void initmod_%s();", mod)
		}
		pr("extern void initmod_main();")

		pr("void initmods() {")

		for _, mod := range cm.CGen.ModsInOrder {
			pr("initmod_%s();", mod)
		}
		pr("initmod_main();")

		pr("}")
		fp.Close()
	}
}

func TryADozenTimes(fn func()) {
	var r interface{}
	done := false
	for i := 0; i < 12; i++ {
		func() {
			defer func() {
				r = recover()
			}()
			fn()
			done = true
		}()
		if done {
			return
		}

	}
	if r != nil {
		log.Panicf("TryADozenTimes: %v", r)
	} else {
		log.Panicf("TryADozenTimes: Not done, but did not figure out what the panic was")
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
	CGen    *CGen
	Package string
	name    string
	CName   string

	UsedBy *GDef

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

	structs map[string]*GDef
	faces   map[string]*GDef

	classes            []string
	classNums          map[string]int
	dmeths             map[string][]string // dsig -> unique interfaces that dispatch it.
	dynamicDefs        []string            // late Dynamic declarations.
	dispatcherTypedefs map[string]bool
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

		structs: make(map[string]*GDef),
		faces:   make(map[string]*GDef),

		classes: []string{
			"_FREE_", "_BYTES_", "_HANDLES_",
		},
		classNums:          make(map[string]int),
		dmeths:             make(map[string][]string),
		dispatcherTypedefs: make(map[string]bool),
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
			UsedBy: nil,
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
	if gdef, ok := val.(*GDef); ok {
		if gdef.istype != nil {
			return gdef.istype
		} else {
			log.Panicf("Either global def %q isn't a type, or (LIMITATION) you used it too soon, before it was defined.", gdef.CName)
		}
	}
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
	results      []NameTV // remembers return variables, if >1
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
	z := &CVal{
		c: Format("%d", x.X),
		t: ConstIntTO,
	}
	return z
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
			case *InterfaceTV, *PointerTV:
				return &CVal{
					c: Format("(/*L1810*/(%s) %s (%s))", a.ToC(), op, b.ToC()),
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
			L("nando ptr to %T", b.Type())
			switch t := b.Type().(type) {
			case *InterfaceTV, *PointerTV:
				return &CVal{
					c: Format("(/*L1833*/(%s) %s (%s))", a.ToC(), op, b.ToC()),
					t: BoolTO,
				}
			case *PrimTV:
				switch t.typecode {
				case "n":
					return &CVal{
						c: Format("(/*L1994*/(%s) %s (void*)0)", a.ToC(), op),
						t: BoolTO,
					}
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
		case "b", "i", "u", "p":
			switch op {
			case "+", "-", "*", "/", "%":
				resultType = a.Type()
			case "==", "!=", "<", "<=", ">", ">=":
				resultType = BoolTO
			}
			if resultType != nil {
				return &CVal{
					c: Format("(%s)(/*L1920*/(%s) %s (%s))", resultType.CType(), a.ToC(), op, b.ToC()),
					t: resultType,
				}
			}
		}
	}
	if a.Type() == BoolTO && b.Type() == BoolTO {
		switch op {
		case "==", "!=", "<", "<=", ">", ">=":
			resultType = BoolTO
		}
		if resultType != nil {
			return &CVal{
				c: Format("(%s)(/*L1920*/(%s) %s (%s))", resultType.CType(), a.ToC(), op, b.ToC()),
				t: resultType,
			}
		}
	}
	if a.Type() == StringTO && b.Type() == StringTO {
		var cfunc string
		switch op {
		case "+":
			resultType = StringTO
			cfunc = "StringAdd"
		case "==":
			resultType = BoolTO
			cfunc = "StringEQ"
		case "!=":
			resultType = BoolTO
			cfunc = "StringNE"
		case "<":
			resultType = BoolTO
			cfunc = "StringLT"
		case "<=":
			resultType = BoolTO
			cfunc = "StringLE"
		case ">":
			resultType = BoolTO
			cfunc = "StringGT"
		case ">=":
			resultType = BoolTO
			cfunc = "StringGE"
		}
		if resultType != nil {
			return &CVal{
				c: Format("(/*L1990*/%s(%s, %s))", cfunc, a.ToC(), b.ToC()),
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
	return &CVal{F("%d", x), ConstIntTO}
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
	inst := co.DefineLocalTempC(ser, pointerTV, creation)

	for i, e := range ctorX.inits {
		val := e.expr.VisitExpr(co)
		co.P("  %s->f_%s = %s; //#%d L2018", inst.CName, e.name, val.ToC(), i)
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

var IDENTIFIER = regexp.MustCompile("^[A-Za-z_][A-Za-z0-9_]*$")

// Jot writes debugging messages to `/tmp/jot`.
func Jot(format string, args ...interface{}) {
	f, err := os.OpenFile("/tmp/jot", os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	text := fmt.Sprintf(format+"\n\n", args...)
	if _, err = f.WriteString(text); err != nil {
		panic(err)
	}
}

func (co *Compiler) Reify(x Value) Value {
	Jot("// DDT Reify(x): %v", x)
	return co.ReifyAs(x, x.Type()) // as its own type
}
func (co *Compiler) ReifyAs(x Value, as TypeValue) Value {
	Jot("// DDT ReifyAs(x, as): %v ; %v", x, as)
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
		gd := co.DefineLocalTempC(ser, x.Type(), cVal)
		return gd
	}

	// CASE: types are different.
	// First, reify as the input type.
	Jot("// DDT Reify(x,as) BEFORE: %v", x)
	reifiedX := co.Reify(x)
	Jot("// DDT Reify(x,as) AFTER: %v", reifiedX)

	// Then convert.
	ser := Serial("reify_as")
	y := co.DefineLocalTempC(ser, as, "")
	co.ConvertTo(reifiedX, y)
	Jot("// DDT Reify(x,as) ConvertTO: %v ; %v", reifiedX, y)
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
	slice := args[0].VisitExpr(co)
	slicec := slice.ToC()
	slice_t := slice.Type().(*SliceTV)

	// TODO: move extra processing up to CallX somehow?
	var items []Value
	var extras Value
	if hasDotDotDot {
		n := len(args)
		for _, e := range args[1 : n-1] {
			items = append(items, e.VisitExpr(co))
		}
		extras = args[n-1].VisitExpr(co)
	} else {
		for _, e := range args[1:] {
			items = append(items, e.VisitExpr(co))
		}
	}

	for _, e := range items {
		// TODO: avoid extra Reify.
		r := co.ReifyAs(e, slice.Type().(*SliceTV).E)
		co.P("%s = SliceAppend(%s, &%s, sizeof(%s), 1 /*TODO: base_cls L2174*/);", slicec, slicec, r.ToC(), r.Type().CType())
	}

	if extras != nil {
		co.P("{int n = (%s).len;//L2227", extras.ToC())
		co.P(" %s x;//L2228", slice_t.E.CType())
		co.P(" for (int i:=0; i < n; i++) {")
		co.P("  SliceGet(%s, &x, sizeof x, i);//L2230", extras.ToC())
		co.P("  %s = SliceAppend(%s, &x, sizeof x, 1 /*TODO cls*/);//L2231", slicec, slicec)
		co.P(" }")
		co.P("}")
	}

	return slice
}
func (co *Compiler) VisitLen(args []Expr) Value {
	assert(len(args) == 1)
	a := args[0].VisitExpr(co)

	switch t := a.Type().(type) {
	case *PrimTV:
		switch a.Type().TypeCode()[0] {
		case 's': // string
			return &CVal{c: F("(%s).len", a.ToC()), t: IntTO}
		}

	case *SliceTV:
		// TODO: avoid runtime division.
		return &CVal{c: F("((%s).len / sizeof(%s))", a.ToC(), t.E.CType()), t: IntTO}

	case *MapTV:
		panic(2194)
	}
	panic(2195)
}
func (co *Compiler) VisitPanic(args []Expr) {
	assert(len(args) == 1)
	val := args[0].VisitExpr(co)
	s := fmt.Sprintf("%v", val)
	s = fmt.Sprintf("%q", s)
	s = s[1 : len(s)-1]
	co.P(`fprintf(stderr, "\nPANIC: %s :: %s\n"); // L2197`, s, val.Type().CType())
	co.P(`fprintf(stderr, "PANIC at line %%d file %%s\n", __LINE__, __FILE__);`)
	co.P("exit(63); // L2199")
}

func (co *Compiler) VisitCall(callx *CallX) Value {
	if identx, ok := callx.Func.(*IdentX); ok {
		// Handle really special methods.
		switch identx.X {
		case "make":
			assert(!callx.HasDotDotDot)
			return co.VisitMake(callx.Args)

		case "append":
			// TODO dotdotdot
			return co.VisitAppend(callx.Args, callx.HasDotDotDot)

		case "len":
			assert(!callx.HasDotDotDot)
			return co.VisitLen(callx.Args)

		case "panic":
			assert(!callx.HasDotDotDot)
			co.VisitPanic(callx.Args)
			return &CVal{"/*void L2200*/", VoidTO}
		}
	}

	// type CallX struct { Func Expr; Args []Expr; HasDotDotDot bool }
	ser := Serial("call")

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

	switch funcValType_t := funcValType.(type) {
	case *PrimTV:
		if funcValType_t.typecode == "t" {
			// Casting to a type.
			// TODO: []byte(s)
			co.P("//is funcVal a type? /// %v /// %v", callx.Func, funcVal)
			targetType, ok := funcVal.ResolveAsTypeValue()
			co.P("// target type/// %v /// %v", targetType, ok)
			assert(ok)
			assert(targetType != nil)
			assert(len(callx.Args) == 1)
			subject := callx.Args[0].VisitExpr(co)
			return co.CastToType(subject, targetType)
		}
	case *FunctionTV:
		{
			funcRec := funcValType_t.FuncRec
			L("funcRec = %v", funcRec)
			fins := funcRec.Ins
			fouts := funcRec.Outs
			co.P("// IsMethod = %v", funcRec.IsMethod)
			co.P("// HasDotDotDot = %v", funcRec.HasDotDotDot)
			var argc []string

			var callme string
			var bm *BoundMethodVal

			var argVals []Value
			if bm, _ = funcVal.(*BoundMethodVal); bm != nil {
				// To avoid double-evaling the receiver (once as receiver,
				// once as first argument), reify it here.
				bmReceiver := co.Reify(bm.receiver)

				if bm.isFace {
					// "bm" gets sent to FormatCall.
					// We could also add "dm" for dispatched method.
					// VisitDot should register all method dispatched from interfaces.
					// Or do that here.
					// Then a final pass across all structs, for that method.
					// Then build a dispatch table.
					// cmeth is just the simple name, in this case.

					// Prepend receiver as first arg.
					//< argVals = append(argVals, bm.receiver)
					argc = append(argc, bmReceiver.ToC())

					callme = co.RegisterDispatchReturnCaller(bm, bmReceiver)
				} else {
					// Prepend receiver as first arg.
					argVals = append(argVals, bmReceiver)
					callme = bm.cmeth
				}
			} else {
				callme = funcVal.ToC()
			}
			for _, e := range callx.Args {
				argVals = append(argVals, e.VisitExpr(co))
			}

			co.P("// Func is V %#v", callme)
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

			// For the non-DotDotDot arguments
			for i, fin := range fins {
				temp := CName(ser, "in", D(i), fin.name)
				L("argVals[%d] :: %T = %q", i, argVals[i], argVals[i].ToC())
				gd := co.DefineLocalTempV(temp, fin.TV, argVals[i])
				// co.ConvertTo(argVals[i], gd)
				argc = append(argc, gd.CName)
			}

			if funcRec.HasDotDotDot {
				if callx.HasDotDotDot {
					if numExtras == 1 && argVals[numNormal].Type().Equals(extraSliceType) {
						argc = append(argc, argVals[numNormal].ToC())
					} else {
						L("numExtras=%d", numExtras)
						L("1=%v", argVals[numNormal].Type())
						L("2=%v", extraSliceType)
						panic("calling with ... not implemented for this case")
					}
				} else {

					sliceName := CName(ser, "in", "extras")
					sliceVar := co.DefineLocalTempC(sliceName, extraSliceType, "NilSlice /*MakeSlice L1949*/")

					for i := 0; i < numExtras; i++ {
						y := co.ReifyAs(argVals[numNormal+i], extraSliceType.E).ToC()

						co.P("%s = SliceAppend(%s, &%s, sizeof(%s), 1/*TODO base_cls*/); // L1954: For extra input #%d",
							sliceVar.CName, sliceVar.CName, y, y, i)
					}

					argc = append(argc, sliceVar.CName)
				}
			}

			if len(fouts) != 1 {
				var multi []NameTV
				for j, out := range fouts {
					rj := Format("_multi_%s_%d", ser, j)
					vj := NameTV{rj, out.TV}
					multi = append(multi, vj)
					gd := co.DefineLocalTempC(rj, out.TV, "")
					argc = append(argc, F("&%s", gd.CName))
				}
				c := co.FormatCall(callme, argc)
				return &CVal{c: c, t: &MultiTV{multi}}
			} else {
				c := co.FormatCall(callme, argc)
				t := fouts[0].TV
				return &CVal{c: c, t: t}
			}
		} // end case *FunctionTV
	} // end switch
	panic(F("cannot use as a function: %v", funcVal))
}

func AppendUnique(slice []string, a string) []string {
	for _, e := range slice {
		if e == a {
			return slice
		}
	}
	return append(slice, a)
}

// Returns dispatch function expression.
func (co *Compiler) RegisterDispatchReturnCaller(bm *BoundMethodVal, bmReceiver Value) (callme string) {
	assert(bm.isFace)
	face, ok := bmReceiver.Type().(*InterfaceTV)
	assert(ok)

	dsig := CName(bm.cmeth, bm.typecode)
	co.CGen.dmeths[dsig] = AppendUnique(
		co.CGen.dmeths[dsig], face.InterfaceRec.name)
	dispatcher := CName("Dispatch", dsig)

	irec := face.InterfaceRec
	var ftv *FunctionTV
	for _, meth := range irec.Meths {
		if meth.name == bm.cmeth {
			ftv = meth.TV.(*FunctionTV)
			break
		}
	}

	if ftv == nil {
		panic(F("cannot find method %s in interface %s", bm.cmeth, irec.name))
	}

	dd := func(line string) {
		co.CGen.dynamicDefs = append(co.CGen.dynamicDefs, line)
	}

	retCType := "rt_" + dispatcher
	if _, exists := co.CGen.dispatcherTypedefs[dispatcher]; !exists {
		dd(F("typedef %s; //L2451", ftv.FuncRec.SignatureStr("(*"+retCType+")", true /*addReceiver*/)))
		co.CGen.dispatcherTypedefs[dispatcher] = true
	}
	dd(F("extern %s %s(void* p); // L2452", retCType, dispatcher))

	return F("(%s(%s))", dispatcher, bmReceiver.ToC())
}

func (co *Compiler) FormatCall(callme string, argc []string) string {
	return Format("(%s(%s)/*L1870*/)", callme, strings.Join(argc, ", "))

}

func (co *Compiler) VisitSub(subx *SubX) Value {
	con := subx.container.VisitExpr(co)
	sub := subx.subscript.VisitExpr(co)

	return &SubVal{
		container: con,
		subscript: sub,
	}
}

func (co *Compiler) VisitSubSlice(ssx *SubSliceX) Value {
	ser := Serial("subslice")
	con := ssx.container.VisitExpr(co)
	conc := con.ToC()
	z := co.DefineLocalTempC(ser, con.Type(), "")
	zc := z.ToC()

	co.P("// DEBUG SLICE: con.Type() == %#v", con.Type())
	co.P("// DEBUG *: con.Type().(SliceTV*) == %#v", con.Type().(*SliceTV))
	co.P("// DEBUG E: con.Type().(SliceTV*).E == %#v", con.Type().(*SliceTV).E)
	co.P("// DEBUG ToC: con.Type().(SliceTV*).E == %#v", con.Type().(*SliceTV).E.CType())
	elementCType := con.Type().(*SliceTV).E.CType()

	var a, b Value
	var ac, bc string
	if ssx.a != nil {
		a = ssx.a.VisitExpr(co)
		ac = fmt.Sprintf("(%s * sizeof(%s))", a.ToC(), elementCType)
	}
	if ssx.b != nil {
		b = ssx.b.VisitExpr(co)
		bc = fmt.Sprintf("(%s * sizeof(%s))", b.ToC(), elementCType)
	}

	if ssx.a == nil {
		if ssx.b == nil {
			co.P("(%s).base = (%s).base; // L2409", zc, conc)
			co.P("(%s).offset = (%s).offset;", zc, conc)
			co.P("(%s).len = (%s).len;", zc, conc)
		} else {
			co.P("assert((%s) >= 0);", bc)
			co.P("assert((%s) <= (%s).len);", bc, conc)
			co.P("(%s).base = (%s).base; // L2409", zc, conc)
			co.P("(%s).offset = (%s).offset;", zc, conc)
			co.P("(%s).len = (%s);", zc, bc)
		}
	} else {
		if ssx.b == nil {
			co.P("assert((%s) >= 0);", ac)
			co.P("assert((%s) <= (%s).len);", ac, conc)
			co.P("(%s).base = (%s).base; // L2409", zc, conc)
			co.P("(%s).offset = (%s).offset + (%s);", zc, conc, ac)
			co.P("(%s).len = (%s).len - (%s);", zc, conc, ac)
		} else {
			co.P("assert((%s) >= 0);", ac)
			co.P("assert((%s) >= 0);", bc)
			co.P("assert((%s) <= (%s).len);", ac, conc)
			co.P("assert((%s) <= (%s).len);", bc, conc)
			co.P("assert((%s) <= (%s));", ac, bc)
			co.P("(%s).base = (%s).base; // L2409", zc, conc)
			co.P("(%s).offset = (%s).offset + (%s);", zc, conc, ac)
			co.P("(%s).len = (%s) - (%s);", zc, bc, ac)
		}
	}

	return z
}

func (co *Compiler) VisitTypeAssert(tass *TypeAssertX) Value {
	x := tass.X.VisitExpr(co)
	tx := x.Type()
	xc := x.ToC()
	castV := tass.T.VisitExpr(co)
	castTV, ok := castV.ResolveAsTypeValue()

	// Trivial assertion.
	if tx.Equals(castTV) {
		return x
	}

	// Any to Concrete.
	if tx.TypeCode() == "a" { // if x is an Any
		co.P("assert(!strcmp(%q, (%s).typecode));",
			castTV.TypeCode(), xc)

		return &CVal{
			c: F("(*(%s*)(%s).pointer)", castTV.CType(), xc),
			t: castTV,
		}
	}

	// Interfaces and Pointers.
	switch tx.(type) {
	case *InterfaceTV, *PointerTV:
		// ok
	default:
		panic(F("cannot runtime cast: %v", x))
	}

	// Don't have runtime magic yet, so be liberal,
	// allow any pointers and interfaces.

	if !ok {
		panic(F("must type-assert to a type, not %v", castV))
	}
	switch castTV.(type) {
	case *InterfaceTV, *PointerTV:
	// ok
	default:
		panic(F("cannot runtime cast to type %v", castTV))
	}

	ser := Serial("runtimeCast")
	g := co.DefineLocalTempC(ser, castTV, "")

	co.P("%s = (void*)(%s); // L2402", g.ToC(), x.ToC())
	return g
}

func (co *Compiler) VisitDot(dotx *DotX) Value {
	log.Printf("VisitDot: <------ %v", dotx)
	val := dotx.X.VisitExpr(co)
	log.Printf("VisitDot: val-- %T ---- %v", val, val)

	switch t := val.Type().(type) {
	case *PrimTV:
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

	case *InterfaceTV:
		{
			faceType := t
			Say("faceType", faceType)
			rec := faceType.InterfaceRec
			Say("rec", rec)
			Say("rec", F("%#v", rec))
			if mtype, ok := FindTypeByName(rec.Meths, dotx.Member); ok {
				ftv := mtype.(*FunctionTV)
				frec := ftv.FuncRec
				bm := &BoundMethodVal{
					receiver: val,
					cmeth:    dotx.Member,
					mtype:    mtype,
					// face already omits receiver from FuncRec
					typecode: frec.BuildTypeCode(false),
					isFace:   true,
				}
				L("VisitDot returns bound meth: %#v", bm)
				return bm
			}
		}
	case *PointerTV:
		if structType, ok := t.E.(*StructTV); ok {
			Say("structType", structType)
			rec := structType.StructRec
			Say("rec", rec)
			Say("rec", F("%#v", rec))
			if ftype, ok := FindTypeByName(rec.Fields, dotx.Member); ok {
				z := &CVal{
					c: Format("(%s)->f_%s", val.ToC(), dotx.Member),
					t: ftype,
				}
				L("VisitDot returns Field: %#v", z)
				return z
			}
			if mtype, ok := FindTypeByName(rec.Meths, dotx.Member); ok {
				ftv := mtype.(*FunctionTV)
				frec := ftv.FuncRec
				bm := &BoundMethodVal{
					receiver: val,
					cmeth:    CName(rec.cname, dotx.Member),
					mtype:    mtype,
					typecode: frec.BuildTypeCode(false),
					isFace:   false,
				}
				L("VisitDot returns bound meth: %#v", bm)
				return bm
			}
		}
	}

	panic("DotXXX L2595")
}

func (co *Compiler) VisitVar(v *VarStmt) {
	debug := co.DefineLocal("v", v.name, co.CMod.VisitTypeExpr(v.tx))
	L("debug VisitVar: %#v ==> %#v", *v, *debug)
}
func (co *Compiler) AssignSingle(left Value, right Value) {
	switch lt := left.(type) {
	case *SubVal:
		switch lt.container.Type().(type) {
		case *SliceTV:
			// slice, size, nth, value
			nth, ok := ResolveAsIntStr(lt.subscript)
			if !ok {
				panic(F("slice subscript must be integer; got %v", lt.subscript))
			}
			rright := co.ReifyAs(right, left.Type())

			co.P(" SlicePut(%s, sizeof(%s), %s, &%s); // L2071",
				lt.container.ToC(),
				right.Type().CType(),
				nth,
				rright.ToC())
			return

		case *MapTV:
			panic(F("TODO MapTV L2070 (%v :: %v) = %v", left, lt, right))
		}
		panic(F("todo SubVal L1835: (%v :: %v) = %v", left, lt, right))
	default:
		co.P("%s = %s; // L2447", left.ToC(), right.ToC())
		return
	}
	panic(F("L2450: cannot (%v :: %v) = %v", left, right))
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

	// Create local vars, if `:=` is the op.
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
				co.P("// L2484: Defined Local %v =%q= %v => %v", id, name, lclType, gd)
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
	log.Printf("co.Subject = %v", co.Subject)
	rec := co.Subject.Type().(*FunctionTV).FuncRec
	log.Printf("rec = %v", rec)
	outs := rec.Outs
	log.Printf("outs = %v", outs)
	log.Printf("co.results = %v", co.results)

	switch len(ret.X) {
	case 0:
		co.P("  return;")
	case 1:
		if len(outs) != 1 {
			panic(F("L2516: Got 1 return value, but needs %d", len(outs)))
		}
		val := ret.X[0].VisitExpr(co)
		log.Printf("return..... val=%v", val)
		co.P("  return %s;", val.ToC())
	default:
		if len(outs) != len(ret.X) {
			panic(F("L2523: Got %d return values, but needs %d", len(ret.X), len(outs)))
		}
		for i, rx := range ret.X {
			// TODO convert
			vx := rx.VisitExpr(co)
			//< co.P("  *%s = %s; // L2527: return multi #%d", co.results[i].name, vx.ToC(), i)
			r := co.results[i]
			co.ConvertToCNameType(vx, "*"+r.name, r.TV)
		}
		co.P("  return; // L2529: multi")
	}
}
func (co *Compiler) VisitFor(fors *ForS) {
	// FOR NOW, assume slice of byte.  TODO: string, map.
	label := Serial("for")
	co.StartScope("VisitFor")

	collV := fors.Coll.VisitExpr(co)

	switch coll_t := collV.Type().(type) {
	case *SliceTV:
		{
			slice := co.Reify(collV)
			index := co.DefineLocalTempC("index_"+label, IntTO, "-1")
			limit := co.DefineLocalTempC("limit_"+label, IntTO, F("((%s).len / sizeof(%s))", slice.ToC(), coll_t.E.CType()))
			var key *GDef
			switch k := fors.Key.(type) {
			case nil:
				{
				}
			case (*IdentX):
				if k.X != "_" && k.X != "" {
					key = co.DefineLocal("v", k.X, IntTO)
					L("VisitFor: SliceTV: key: %v", key)
				}
			}

			var value *GDef
			switch v := fors.Value.(type) {
			case nil:
				{
				}
			case (*IdentX):
				if v.X != "_" && v.X != "" {
					value = co.DefineLocal("v", v.X, coll_t.E)
					L("VisitFor: SliceTV: value: %v", value)
				}
			}

			co.P("while(1) { Cont_%s: {}", label)

			co.P("%s++; // L2629", index.CName)
			co.P("if (%s >= %s) break; // L2630", index.CName, limit.CName)

			if key != nil {
				co.P("%s = %s; // L2816:key", key.CName, index.CName)
			}
			if value != nil {
				co.P("SliceGet(%s, sizeof(%s), %s, &%s); //L2645", slice.ToC(), coll_t.E.CType(), index.CName, value.CName)
			}
		}
	case *PrimTV:
		if coll_t.typecode == "s" {
			str := co.Reify(collV)
			index := co.DefineLocalTempC("index_"+label, IntTO, "-1")
			limit := co.DefineLocalTempC("limit_"+label, IntTO, F("(%s).len", str.ToC()))
			var key *GDef
			switch k := fors.Key.(type) {
			case nil:
				{
				}
			case (*IdentX):
				if k.X != "_" && k.X != "" {
					key = co.DefineLocal("v", k.X, IntTO) // string keys are ints
					L("VisitFor: SliceTV: key: %v", key)
				}
			}

			var value *GDef
			switch v := fors.Value.(type) {
			case nil:
				{
				}
			case (*IdentX):
				if v.X != "_" && v.X != "" {
					value = co.DefineLocal("v", v.X, ByteTO) // string values are bytes
					L("VisitFor: SliceTV: value: %v", key)
				}
			}

			co.P("while(1) { Cont_%s: {}", label)

			co.P("%s++; // L2629", index.CName)
			co.P("if (%s >= %s) break; // L2630", index.CName, limit.CName)

			if key != nil {
				co.P("%s = %s;", key.CName, index.CName)
			}
			if value != nil {
				co.P("StringGet(%s, %s, &%s); //L2645", str.ToC(), index.CName, value.CName)
			}
		}
	} // end switch collV type

	savedB, savedC := co.BreakTo, co.ContinueTo
	co.BreakTo, co.ContinueTo = "Break_"+label, "Cont_"+label
	fors.Body.VisitStmt(co)
	co.P("  }")
	co.P("Break_%s: {}", label)
	co.BreakTo, co.ContinueTo = savedB, savedC
	co.FinishScope()
}

func (co *Compiler) VisitWhile(wh *WhileS) {
	label := Serial("while")
	co.StartScope("VisitWhile")
	if wh.First != nil {
		co.P("// First: %q", V(wh.First))
		wh.First.VisitStmt(co)
		co.P("// L2672: End First")
	}
	co.P("while(1) { // L2674: %s", label)
	if wh.Pred != nil {
		pred := wh.Pred.VisitExpr(co)
		assert(pred.Type() == BoolTO)
		co.P("    int _while_ = %s;", pred.ToC())
		co.P("    if (!_while_) break; // L2679")
	}
	savedB, savedC := co.BreakTo, co.ContinueTo
	co.BreakTo, co.ContinueTo = "Break_"+label, "Cont_"+label
	wh.Body.VisitStmt(co)
	co.P("Cont_%s: {}", label)
	if wh.Next != nil {
		wh.Next.VisitStmt(co)
	}
	co.P("  } // L2688")
	co.P("Break_%s: {}", label)
	co.BreakTo, co.ContinueTo = savedB, savedC
	co.FinishScope()
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
	co.StartScope("VisitIf")
	co.P("  { bool _if_ = %s;", ifs.Pred.VisitExpr(co).ToC())
	co.P("  if( _if_ ) {")
	ifs.Yes.VisitStmt(co)
	if ifs.No != nil {
		co.P("  } else {")
		ifs.No.VisitStmt(co)
	}
	co.P("  }}")
	co.FinishScope()
}
func (co *Compiler) VisitSwitch(sws *SwitchS) {
	co.StartScope("VisitSwitch")
	co.P("  { int _switch_ = %s;", sws.Switch.VisitExpr(co).ToC())
	for _, c := range sws.Cases {
		co.StartScope("VisitCase")
		co.P("  if (")
		for _, m := range c.Matches {
			co.P("_switch_ == %s ||", m.VisitExpr(co).ToC())
		}
		co.P("      0 ) {")
		c.Body.VisitStmt(co)
		co.P("  } else ")
		co.FinishScope()
	}
	co.P("  {")
	if sws.Default != nil {
		sws.Default.VisitStmt(co)
	}
	co.P("  }")
	co.P("  }")
	co.FinishScope()
}
func (co *Compiler) VisitBlock(a *Block) {
	co.StartScope("VisitBlock")
	if a == nil {
		panic("L2058")
	}
	for i, e := range a.stmts {
		ser := Serial("block")
		co.P("// @@ VisitBlock[%s,%d] <= %q", ser, i, F("%v", e))
		e.VisitStmt(co)
		log.Printf("VisitBlock[%d] ==>\n<<<\n%s\n>>>", i, co.Buf.String())
	}
	co.FinishScope()
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
	p := co.CurrentBlock
	for p != nil {
		q := p
		for q != nil {
			L("nando q: %q [locals %v] %s", name, q.locals, q.why)
			q = q.parent
		}
		return p.Find(name)
	}
	L("nando p==nil: %q [cmod @%q]", name, co.CMod.Package)
	return co.CMod.Find(name)
}

func (co *Compiler) DefineLocalTempC(tempName string, tempType TypeValue, initC string) *GDef {
	gd := co.DefineLocal("tmp", tempName, tempType)
	if initC == "" {
		// TODO: zero it?
	} else {
		co.P("%s = %s; // L2653", gd.CName, initC)
	}
	return gd
}
func (co *Compiler) DefineLocalTempV(tempName string, tempType TypeValue, from Value) *GDef {
	gd := co.DefineLocal("tmp", tempName, tempType)
	co.ConvertToCNameType(from, gd.CName, gd.typeof)
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
	}
	co.slots[local.CName] = local
	return local
}
func (co *Compiler) FinishScope() {
	p := co.CurrentBlock
	co.P("// Finishing Scope: %s }}}", p.why)
	co.CurrentBlock = p.parent
}
func (co *Compiler) StartScope(why string) *Block {
	p := co.CurrentBlock
	ser := Serial("scope")
	if p == nil {
		why = F("%s (%s)", why, ser)
	} else {
		why = F("%s (%s) => %s", why, ser, p.why)
	}
	co.P("// Starting Scope: %q {{{", why)
	block := &Block{
		why:      why,
		locals:   make(map[string]*GDef),
		parent:   p,
		compiler: co,
	}
	co.CurrentBlock = block
	return block
}
func (co *Compiler) EmitFunc(gd *GDef, justDeclare bool) {
	co.StartScope("EmitFunc")
	rec := gd.typeof.(*FunctionTV).FuncRec
	co.P(rec.SignatureStr(gd.CName, false))
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
		if rec.HasDotDotDot && i == len(rec.Ins)-1 {
			co.DefineLocal("in", name, in.TV)
			L("%q DDT_IN #%d name=%v in.TV=%v (DotDotDot)", gd.CName, i, name, in.TV)
		} else {
			co.DefineLocal("in", name, in.TV)
			L("%q DDT_IN #%d name=%v in.TV=%v", gd.CName, i, name, in.TV)
		}
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
			out := co.DefineLocal("out", name, out.TV)
			co.results = append(co.results, NameTV{out.CName, out.typeof})
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
		co.P("auto %v %v = %s; // DEF LOCAL L2145 Type=%#v", e.typeof.CType(), e.CName, e.typeof.Zero(), e.typeof)
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

/////////////////////////////////////////////////
/////////////////////////////////////////////////

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
