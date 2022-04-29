package parser

type Uses struct {
	Locals map[string]TypeValue
	User   *GDef // current
}

func (o *Uses) VisitLitInt(x *LitIntX) Value {
	return nil
}
func (o *Uses) VisitLitString(x *LitStringX) Value {
	return nil
}
func (o *Uses) VisitIdent(x *IdentX) Value {
	return nil
}
func (o *Uses) _VisitIdent_(x *IdentX) Value {
	return nil
}
func (o *Uses) VisitBinOp(x *BinOpX) Value {
	return nil
}
func (o *Uses) VisitConstructor(x *ConstructorX) Value {
	return nil
}
func (o *Uses) VisitFunction(x *FunctionX) Value {
	return nil
}
func (o *Uses) VisitCall(x *CallX) Value {
	return nil
}
func (o *Uses) VisitSub(x *SubX) Value {
	return nil
}
func (o *Uses) VisitDot(dotx *DotX) Value {
	return nil
}

func (o *Uses) VisitAssign(ass *AssignS) {
}
func (o *Uses) VisitReturn(ret *ReturnS) {
}
func (o *Uses) VisitWhile(wh *WhileS) {
}
func (o *Uses) VisitFor(fors *ForS) {
}
func (o *Uses) VisitBreak(sws *BreakS) {
}
func (o *Uses) VisitContinue(sws *ContinueS) {
}
func (o *Uses) VisitIf(ifs *IfS) {
}
func (o *Uses) VisitSwitch(sws *SwitchS) {
}
func (o *Uses) VisitBlock(a *Block) {
}
