package std_optimize

import "../../core"
import "fmt"

type Optimization struct {
	Form core.Obj
	ConversionFunc func(core.Obj,func(core.Obj)string)string
	ExtraText string
}

func OptimizeObj(opts []*Optimization, optsUsed map[*Optimization]bool, obj core.Obj) string {
	for _, opt := range opts {
		if ObjMatchesForm(obj, opt.Form) {
			optsUsed[opt] = true
			return opt.ConversionFunc(obj, func(a core.Obj) string { return OptimizeObj(opts, optsUsed, a) } )
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
