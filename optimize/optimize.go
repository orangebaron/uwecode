package optimize

import "./core"
import "fmt"

type Optimization struct {
	Form func(core.Obj) bool
	ConversionFunc func(core.Obj, func(core.Obj) string) string
	ExtraText      string
	Imports []string
}

func OptimizeObjHelper(opts []*Optimization, optsUsed map[*Optimization]bool, obj core.Obj) string {
	for _, opt := range opts {
		if opt.Form(obj) {
			optsUsed[opt] = true
			return opt.ConversionFunc(obj, func(a core.Obj) string { return OptimizeObjHelper(opts, optsUsed, a) })
		}
	}
	switch v := obj.(type) {
	case core.Function:
		return fmt.Sprintf("core.Function{%d,%s}", v.N, OptimizeObjHelper(opts, optsUsed, v.X))
	case core.Called:
		return fmt.Sprintf("core.Called{%s,%s}", OptimizeObjHelper(opts, optsUsed, v.X), OptimizeObjHelper(opts, optsUsed, v.Y))
	case core.CalledChurchNum:
		return fmt.Sprintf("core.CalledChurchNum{%d,%s}", v.Num, OptimizeObjHelper(opts, optsUsed, v.X))
	default:
		return fmt.Sprintf("%#v", obj)
	}
}
