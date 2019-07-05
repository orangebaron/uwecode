package core

type Type interface {
	MatchesType(Obj, uint) bool
	ToExampleObj() Obj
}

func (t ArbiraryVal) MatchesType(a Obj, _ uint) bool { return a.SimplifyFully() == t }
func (t ArbitraryVal) ToExampleObj() Obj { return t }

func (t ArbitraryMethod) MatchesType(a Obj, itersLeft uint) bool {
	if itersLeft == 0 {
		return true
	} else {
		return t.Otp.MatchesType(a.Call(t.Inp.ToExampleObj()), itersLeft - 1)
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
