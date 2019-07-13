package core

type Type interface {
	MatchesType(Obj, uint) bool
	IsToObj(Obj) bool
	ToObj() Obj
}

func (t ArbitraryVal) MatchesType(a Obj, _ uint) (returned bool) {
	defer func() {
		if recover() != nil {
			returned = false
		}
	}()
	SimplifyUntil(func(f Obj) (bool, interface{}) { return f == t, nil }, a)
	return true
}
func (t ArbitraryVal) IsToObj(a Obj) bool { return a == t }
func (t ArbitraryVal) ToObj() Obj         { return t }

type MethodType struct {
	Inp Type
	Otp Type
}

func (t MethodType) MatchesType(a Obj, itersLeft uint) bool {
	if itersLeft == 0 {
		return true
	} else if _, isArb := a.(ArbitraryVal); isArb {
		//		panic("lol")
		return false
	} else {
		return t.Otp.MatchesType(a.Call(t.Inp.ToObj()), itersLeft-1)
	}
}
func (t MethodType) IsToObj(a Obj) bool {
	// TODO: that assumes its simplified fully (fair assumption) and also that == works right (bad assumption)
	if methodObj, isMethodObj := a.(MethodObj); isMethodObj {
		return methodObj.Inp == t.Inp && t.Otp.IsToObj(methodObj.Otp)
	} else {
		return false
	}
}
func (t MethodType) ToObj() Obj { return MethodObj{t.Inp, t.Otp.ToObj()} }

type MethodObj struct {
	Inp Type
	Otp Obj
}

func (f MethodObj) Call(a Obj) Obj {
	if f.Inp.MatchesType(a, 1000) { // TODO number other than 1000 lol
		return f.Otp.ReplaceF(func(x Obj) bool { return f.Inp.IsToObj(x) }, a)
	} else {
		return Called{f, a}
	}
}
func (f MethodObj) Simplify(_ uint) Obj       { return f }
func (f MethodObj) Replace(_ uint, _ Obj) Obj { return f }
func (f MethodObj) ReplaceF(fun func(Obj) bool, x Obj) Obj {
	if fun(f) {
		return x
	} else {
		return f
	}
}
func (f MethodObj) GetUnboundVars(_ map[uint]bool, _ map[uint]bool) {}
func (f MethodObj) GetAllVars(_ map[uint]bool)                      {}
func (f MethodObj) ReplaceBindings(_ map[uint]bool) Obj             { return f }

type TBDType struct {
	Obj
}

func (t TBDType) MatchesType(a Obj, itersLeft uint) bool {
	return ObjToType(t.Obj).MatchesType(a, itersLeft)
}
func (t TBDType) IsToObj(a Obj) bool { return ObjToType(t.Obj).IsToObj(a) }
func (t TBDType) ToObj() Obj {
	return ObjToType(t.Obj).ToObj()
}
