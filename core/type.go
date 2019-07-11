package core

type Type interface {
	MatchesType(Obj, uint) bool
	ToExampleObj() Obj
}

func (t ArbitraryVal) MatchesType(a Obj, _ uint) (returned bool) {
	defer recover()
	SimplifyUntil(func(f Obj) (bool, interface{}) { return f == t, nil }, a)
	return true
}
func (t ArbitraryVal) ToExampleObj() Obj { return t }

type TypeMethod struct {
	Inp Type
	Otp Type
}

func (t TypeMethod) MatchesType(a Obj, itersLeft uint) bool {
	if itersLeft == 0 {
		return true
	} else {
		return t.Otp.MatchesType(a.Call(t.Inp.ToExampleObj()), itersLeft-1)
	}
}
func (t TypeMethod) ToExampleObj() Obj { return t }

func (f TypeMethod) Call(a Obj) Obj {
	if f.Inp.MatchesType(a, 1000) { // TODO number other than 1000 lol
		return f.Otp.ToExampleObj()
	} else {
		return Called{f, a}
	}
}
func (f TypeMethod) Simplify(_ uint) Obj                             { return f }
func (f TypeMethod) Replace(_ uint, _ Obj) Obj                       { return f }
func (f TypeMethod) GetUnboundVars(_ map[uint]bool, _ map[uint]bool) {}
func (f TypeMethod) GetAllVars(_ map[uint]bool)                      {}
func (f TypeMethod) ReplaceBindings(_ map[uint]bool) Obj             { return f }

type TBDType struct {
	Obj
}

func (t TBDType) MatchesType(a Obj, itersLeft uint) bool {
	return ObjToType(t.Obj).MatchesType(a, itersLeft)
}
func (t TBDType) ToExampleObj() Obj {
	return ObjToType(t.Obj).ToExampleObj()
}
