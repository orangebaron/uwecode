package core

type TypeMap struct {
	dict map[Type]Obj // type has a different == function, so we have to use that
}

func MakeTypeMap() TypeMap {
	return TypeMap{make(map[Type]Obj)}
}

func (m TypeMap) Get(t Type) Obj {
	for k, v := range m.dict {
		if t.EqualsType(k, 1000) { // TODO: 1000?
			return v
		}
	}
	return TypeObj{t, m}
}
func (m TypeMap) Set(k Type, v Obj) {
	m.dict[k] = v
}

type Type interface {
	MatchesType(Obj, uint, TypeMap) bool
	EqualsType(Type, uint) bool
	CallType(Obj, TypeMap) Obj
}

func (t ArbitraryVal) MatchesType(a Obj, _ uint, _ TypeMap) (returned bool) {
	defer func() {
		if recover() != nil {
			returned = false
		}
	}()
	SimplifyUntil(func(f Obj) (bool, interface{}) {
		to, isTo := f.(TypeObj)
		return isTo && to.Type == t, nil
	}, a)
	return true
}
func (t ArbitraryVal) EqualsType(a Type, depth uint) bool {
	if arb, isArb := a.(ArbitraryVal); isArb {
		return arb.ID == t.ID
	} else if depth == 0 {
		return true
	} else {
		return a.EqualsType(t, depth-1)
	}
}
func (t ArbitraryVal) CallType(a Obj, dict TypeMap) Obj {
	return Called{dict.Get(t), a}
}

type MethodType struct {
	Inp Type
	Otp Type
}

func (t MethodType) MatchesType(a Obj, depth uint, dict TypeMap) bool {
	if depth == 0 {
		return true
	} else {
		return t.Otp.MatchesType(a.Call(TypeObj{t.Inp, dict}), depth-1, dict)
	}
}
func (t MethodType) EqualsType(a Type, depth uint) bool {
	if a == t {
		return true
	} else if depth == 0 {
		return true
	} else if _, isArb := a.(ArbitraryVal); isArb {
		return false
	} else if meth, isMeth := a.(MethodType); isMeth {
		return t.Inp.EqualsType(meth.Inp, depth-1) && t.Otp.EqualsType(meth.Otp, depth-1)
	} else {
		return a.EqualsType(t, depth-1)
	}
}
func (t MethodType) CallType(a Obj, dict TypeMap) Obj {
	if t.Inp.MatchesType(a, 1000, dict) { // TODO 1000?
		dict.Set(t.Inp, a)
		return dict.Get(t.Otp)
	} else {
		return Called{TypeObj{t, dict}, a}
	}
}

type TBDType struct {
	Obj
}

func (t TBDType) MatchesType(a Obj, depth uint, dict TypeMap) bool {
	return ObjToType(t.Obj).MatchesType(a, depth, dict)
}
func (t TBDType) EqualsType(a Type, depth uint) bool {
	return ObjToType(t.Obj).EqualsType(a, depth)
}
func (t TBDType) CallType(a Obj, dict TypeMap) Obj {
	return ObjToType(t.Obj).CallType(a, dict)
}

type TypeObj struct {
	Type
	TypeMap
}

func (f TypeObj) Call(a Obj) Obj {
	return f.Type.CallType(a, f.TypeMap)
}
func (f TypeObj) Simplify(_ uint) Obj                             { return f }
func (f TypeObj) Replace(_ uint, _ Obj) Obj                       { return f }
func (f TypeObj) GetUnboundVars(_ map[uint]bool, _ map[uint]bool) {}
func (f TypeObj) GetAllVars(_ map[uint]bool)                      {}
func (f TypeObj) ReplaceBindings(_ map[uint]bool) Obj             { return f }
