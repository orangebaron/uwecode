package std_opts

import "../../../core"
import "../../../optimize"

type ObjList struct {
	Objs *[]core.Obj
}

func (f ObjList) Call(a core.Obj) core.Obj {
	if len(*f.Objs) == 0 {
		return core.Function{0, core.ReturnVal{0}}
	} else {
		return core.Function{0, core.Called{a, core.Function{0, core.Called{core.Called{core.ReturnVal{0}, (*f.Objs)[0]}, MakeObjList((*f.Objs)[1:])}}}} //TODO: replace stuff?
	}
}
func (f ObjList) Simplify(depth uint) core.Obj {
	//TODO: this is slightly unsafe/nonfunctional, go ahead with it?
	if depth > 0 {
		for i, v := range *f.Objs {
			(*f.Objs)[i] = v.Simplify(depth - 1)
		}
	}
	return f
}
func (f ObjList) Replace(n uint, a core.Obj) core.Obj {
	newObjs := make([]core.Obj, len(*f.Objs))
	for i, v := range *f.Objs {
		newObjs[i] = v.Replace(n, a)
	}
	return MakeObjList(newObjs)
}
func (f ObjList) GetUnboundVars(bound map[uint]bool, unbound map[uint]bool) {
	for _, v := range *f.Objs {
		v.GetUnboundVars(bound, unbound)
	}
}
func (f ObjList) GetAllVars(vars map[uint]bool) {
	for _, v := range *f.Objs {
		v.GetAllVars(vars)
	}
}
func (f ObjList) ReplaceBindings(toReplace map[uint]bool) core.Obj {
	newObjs := make([]core.Obj, len(*f.Objs))
	for i, v := range *f.Objs {
		newObjs[i] = v.ReplaceBindings(toReplace)
	}
	return MakeObjList(newObjs)
}

func MakeObjList(objs []core.Obj) core.Obj {
	return ObjList{&objs}
}

type OperationType byte

const (
	Inc OperationType = iota
	Plus
	Mult
)

type NumOpt struct {
	OperationType
}

func (f NumOpt) Call(x core.Obj) (returnVal core.Obj) {
	defer func() {
		if recover() != nil {
			var fReal core.Obj
			switch f.OperationType {
			case Inc:
				fReal = core.Function{0, core.Function{1, core.Function{2, core.Called{core.Called{core.ReturnVal{0}, core.ReturnVal{1}}, core.Called{core.ReturnVal{1}, core.ReturnVal{2}}}}}}
			case Plus:
				fReal = core.Function{0, core.Called{core.ReturnVal{0}, NumOpt{Inc}}}
			case Mult:
				fReal = core.Function{0, core.Function{1, core.Called{core.Called{core.ReturnVal{1}, core.Called{NumOpt{Plus}, core.ReturnVal{0}}}, core.ChurchNum{0}}}}
			default:
				panic("unknown OperationType")
			}
			returnVal = fReal.Call(x)
		}
	}()
	switch f.OperationType {
	case Inc:
		return core.ChurchNum{core.ObjToInt(x) + 1}
	default:
		return NumOpt2{f.OperationType, core.ObjToInt(x)}
	}
}
func (f NumOpt) Simplify(_ uint) core.Obj                        { return f }
func (f NumOpt) Replace(_ uint, _ core.Obj) core.Obj             { return f }
func (f NumOpt) GetUnboundVars(_ map[uint]bool, _ map[uint]bool) {}
func (f NumOpt) GetAllVars(_ map[uint]bool)                      {}
func (f NumOpt) ReplaceBindings(_ map[uint]bool) core.Obj        { return f }

type NumOpt2 struct {
	OperationType
	Num uint
}

func (f NumOpt2) Call(x core.Obj) (returnVal core.Obj) {
	defer func() {
		if recover() != nil { // TODO: check specific error message
			var fReal core.Obj
			switch f.OperationType {
			case Plus:
				fReal = core.ChurchNum{f.Num}.Call(NumOpt{Inc})
			case Mult:
				fReal = core.Function{0, core.Called{core.Called{core.ReturnVal{0}, core.Called{NumOpt{Plus}, core.ChurchNum{f.Num}}}, core.ChurchNum{0}}}
			default:
				panic("unrecognized OperationType")
			}
			returnVal = fReal.Call(x)
		}
	}()
	switch f.OperationType {
	case Plus:
		return core.ChurchNum{f.Num + core.ObjToInt(x)}
	case Mult:
		return core.ChurchNum{f.Num * core.ObjToInt(x)}
	default:
		panic("unrecognized OperationType")
	}
}
func (f NumOpt2) Simplify(_ uint) core.Obj                        { return f } // TODO: maybe make a "simplify into normal object form" function
func (f NumOpt2) Replace(_ uint, _ core.Obj) core.Obj             { return f }
func (f NumOpt2) GetUnboundVars(_ map[uint]bool, _ map[uint]bool) {}
func (f NumOpt2) GetAllVars(_ map[uint]bool)                      {} // TODO: f -> _?
func (f NumOpt2) ReplaceBindings(_ map[uint]bool) core.Obj        { return f }

func incOptHelper(f core.Obj) (bool, interface{}) {
	called1, isCalled1 := f.(core.Called)
	if !isCalled1 {
		return false, nil
	}
	called2, isCalled2 := called1.X.(core.Called)
	if !isCalled2 {
		return false, nil
	}
	arb1, isArb1 := called2.X.(core.ArbitraryVal)
	if !isArb1 || arb1.ID != 0 {
		return false, nil
	}
	arb2, isArb2 := called2.Y.(core.ArbitraryVal)
	if !isArb2 || arb2.ID != 1 {
		return false, nil
	}
	called3, isCalled3 := called1.Y.(core.Called)
	if !isCalled3 {
		return false, nil
	}
	arb3, isArb3 := called3.X.(core.ArbitraryVal)
	if !isArb3 || arb3.ID != 1 {
		return false, nil
	}
	arb4, isArb4 := called3.Y.(core.ArbitraryVal)
	if !isArb4 || arb4.ID != 2 {
		return false, nil
	}
	return true, nil
}

func plusOptHelper(f core.Obj) (bool, interface{}) {
	called, isCalled := f.(core.Called)
	if !isCalled {
		return false, nil
	}
	arb, isArb := called.X.(core.ArbitraryVal)
	if !isArb || arb.ID != 3 {
		return false, nil
	}
	defer func() { recover() }()
	core.SimplifyUntil(incOptHelper, called.Y.Call(core.ArbitraryVal{0}).Call(core.ArbitraryVal{1}).Call(core.ArbitraryVal{2}))
	return true, nil
}

func multOptHelper(f core.Obj) (bool, interface{}) {
	defer func() { recover() }()
	called1, isCalled1 := f.(core.Called)
	if !isCalled1 || core.ObjToInt(called1.Y) != 0 {
		return false, nil
	}
	called2, isCalled2 := called1.X.(core.Called)
	if !isCalled2 {
		return false, nil
	}
	arb1, isArb1 := called2.X.(core.ArbitraryVal)
	if !isArb1 || arb1.ID != 5 {
		return false, nil
	}
	called3, isCalled3 := called2.Y.(core.Called)
	if !isCalled3 {
		return false, nil
	}
	arb2, isArb2 := called3.Y.(core.ArbitraryVal)
	if !isArb2 || arb2.ID != 4 {
		return false, nil
	}
	core.SimplifyUntil(plusOptHelper, called3.X.Call(core.ArbitraryVal{3}))
	return true, nil
}

var listOpt optimize.Optimization = optimize.Optimization{
	func(f core.Obj, convert func(core.Obj) string) string {
		defer func() { recover() }()
		str := "std_opts.MakeObjList([]core.Obj{"
		for _, obj := range core.ObjToList(f) {
			str += convert(obj) + ","
		}
		str += "})"
		return str
	},
	"./.uwe/opts/std_opts",
}

var incOpt optimize.Optimization = optimize.Optimization{
	func(f core.Obj, convert func(core.Obj) string) string {
		defer func() { recover() }()
		core.SimplifyUntil(incOptHelper, f.Call(core.ArbitraryVal{0}).Call(core.ArbitraryVal{1}).Call(core.ArbitraryVal{2}))
		return "std_opts.NumOpt{std_opts.Inc}"
	},
	"./.uwe/opts/std_opts",
}

var plusOpt optimize.Optimization = optimize.Optimization{
	func(f core.Obj, convert func(core.Obj) string) string {
		defer func() { recover() }()
		core.SimplifyUntil(plusOptHelper, f.Call(core.ArbitraryVal{3}))
		return "std_opts.NumOpt{std_opts.Plus}"
	},
	"./.uwe/opts/std_opts",
}

var multOpt optimize.Optimization = optimize.Optimization{
	func(f core.Obj, convert func(core.Obj) string) string {
		defer func() { recover() }()
		core.SimplifyUntil(multOptHelper, f.Call(core.ArbitraryVal{4}).Call(core.ArbitraryVal{5}))
		return "std_opts.NumOpt{std_opts.Mult}"
	},
	"./.uwe/opts/std_opts",
}

var OptsList []*optimize.Optimization = []*optimize.Optimization{&listOpt, &incOpt, &plusOpt, &multOpt}
