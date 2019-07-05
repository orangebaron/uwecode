package core

type Type interface {
	MatchesType(Obj, uint) bool
	ToExampleObj() Obj
}

func (t ArbitraryVal) MatchesType(a Obj, _ uint) bool { return a.SimplifyFully() == t }
func (t ArbitraryVal) ToExampleObj() Obj { return t }

func (t ArbitraryMethod) MatchesType(a Obj, itersLeft uint) bool {
	if itersLeft == 0 {
		return true
	} else {
inp, otp := ObjToType(t.Inp), ObjToType(t.Otp)
return otp.MatchesType(a.Call(inp.ToExampleObj()), itersLeft - 1)
	}
}
func (t ArbitraryMethod) ToExampleObj() Obj { return t }

type TBDType struct {
	Obj
}

func (t TBDType) MatchesType(a Obj, itersLeft uint) bool {
	return ObjToType(t.Obj).MatchesType(a, itersLeft)
}
func (t TBDType) ToExampleObj() Obj {
	return ObjToType(t.Obj).ToExampleObj()
}
