package optimize

import "../core"
import "fmt"

type Optimization struct {
	ConversionFunc func(core.Obj, func(core.Obj) string) string
	Import         string
}

func OptimizeObjHelper(opts []*Optimization, optsUsed map[*Optimization]bool, obj core.Obj) string {
	for _, opt := range opts {
		str := opt.ConversionFunc(obj, func(a core.Obj) string { return OptimizeObjHelper(opts, optsUsed, a) })
		if str != "" {
			optsUsed[opt] = true
			return str
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

func OptimizeObj(opts []*Optimization, obj core.Obj) (string, string) {
	optsUsed := make(map[*Optimization]bool)
	mainString := OptimizeObjHelper(opts, optsUsed, obj)
	headerString := ""
	importsUsed := make(map[string]bool)
	for opt, wasUsed := range optsUsed {
		if wasUsed {
			importsUsed[opt.Import] = true
		}
	}
	for imp, _ := range importsUsed {
		if imp != "" {
			headerString += fmt.Sprintf("import \"%s\"\n", imp)
		}
	}
	return headerString, mainString
}
