package std_optimize

import "../../core"
import "fmt"

type Optimization struct {
	Form           func(uint, uint) (bool, uint)
	ConversionFunc func(core.Obj, func(core.Obj) string) string
	ExtraText      string
}

func OptimizeObj(opts []*Optimization, optsUsed map[*Optimization]bool, obj core.Obj) string {
	for _, opt := range opts {
		if ObjMatchesForm(obj, opt.Form) {
			optsUsed[opt] = true
			return opt.ConversionFunc(obj, func(a core.Obj) string { return OptimizeObj(opts, optsUsed, a) })
		}
	}
	switch v := obj.(type) {
	case core.Function:
		return fmt.Sprintf("core.Function{%d,%s}", v.N, OptimizeObj(opts, optsUsed, v.X))
	case core.Called:
		return fmt.Sprintf("core.Called{%s,%s}", OptimizeObj(opts, optsUsed, v.X), OptimizeObj(opts, optsUsed, v.Y))
	case core.CalledChurchNum:
		return fmt.Sprintf("core.CalledChurchNum{%d,%s}", v.Num, OptimizeObj(opts, optsUsed, v.X))
	default:
		return fmt.Sprintf("%#v", obj)
	}
}

type FormTestArbitraryVal struct {
	Val  uint
	Form func(uint, uint) (bool, uint)
}

func (f FormTestArbitraryVal) Call(a core.Obj) core.Obj {
	arb, isArb := a.(FormTestArbitraryVal)
	if isArb {
		b, val := f.Form(f.Val, arb.Val)
		if b {
			return FormTestArbitraryVal{val, f.Form}
		}
	}
	return core.Called{f, a}
}
func (f FormTestArbitraryVal) Simplify(_ uint) core.Obj                        { return f }
func (f FormTestArbitraryVal) Replace(_ uint, _ core.Obj) core.Obj             { return f }
func (f FormTestArbitraryVal) GetUnboundVars(_ map[uint]bool, _ map[uint]bool) {}
func (f FormTestArbitraryVal) GetAllVars(_ map[uint]bool)                      {}
func (f FormTestArbitraryVal) ReplaceBindings(_ map[uint]bool) core.Obj        { return f }

func ObjMatchesFormHelper(obj core.Obj) (bool, interface{}) {
	arb, isArb := obj.(FormTestArbitraryVal)
	return isArb, arb.Val == 0
}

func ObjMatchesForm(obj core.Obj, form func(uint, uint) (bool, uint)) bool {
	defer func() { recover() }()
	for i := uint(1); ; i++ {
		obj = obj.Call(FormTestArbitraryVal{i, form}) // TODO make a new absoluteval type
		b, val := core.SimplifyUntilNoPanic(ObjMatchesFormHelper, obj)
		if b {
			return val.(bool)
		} else {
			obj = val.(core.Obj)
		}
	}
}
