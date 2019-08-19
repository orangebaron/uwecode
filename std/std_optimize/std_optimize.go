package std_optimize

import "../../core"
import "fmt"

type Form struct {
	CallTable func(uint, uint) (bool, uint)
	AcceptibleVals func(uint) bool
}
type Optimization struct {
	Form
	ConversionFunc func(core.Obj, func(core.Obj) string) string
	ExtraText      string
	Imports []string
}

func OptimizeObjHelper(opts []*Optimization, optsUsed map[*Optimization]bool, obj core.Obj) string {
	for _, opt := range opts {
		if ObjMatchesForm(obj, opt.Form) {
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

func OptimizeObj(opts []*Optimization, obj core.Obj) (string, string) {
	optsUsed := make(map[*Optimization]bool)
	mainString := OptimizeObjHelper(opts, optsUsed, obj)
	headerString := ""
	importsUsed := make(map[string]bool)
	for opt, wasUsed := range optsUsed {
		if wasUsed {
			headerString += opt.ExtraText + ";"
			for _, imp := range opt.Imports {
				importsUsed[imp] = true
			}
		}
	}
	for imp, _ := range importsUsed {
		headerString = fmt.Sprintf("import \"%s\";", imp) + headerString
	}
	return headerString, mainString
}

type FormTestArbitraryVal struct {
	Val  uint
	*Form
}

func (f FormTestArbitraryVal) Call(a core.Obj) core.Obj {
	arb, isArb := a.(FormTestArbitraryVal)
	if isArb {
		b, val := f.Form.CallTable(f.Val, arb.Val)
		if b {
			return FormTestArbitraryVal{val, f.Form}
		} else {
			panic("No value in calltable")
		}
	}
	return core.Called{f, a}
}
func (f FormTestArbitraryVal) Simplify(_ uint) core.Obj                        { return f }
func (f FormTestArbitraryVal) Replace(_ uint, _ core.Obj) core.Obj             { return f }
func (f FormTestArbitraryVal) GetUnboundVars(_ map[uint]bool, _ map[uint]bool) {}
func (f FormTestArbitraryVal) GetAllVars(_ map[uint]bool)                      {}
func (f FormTestArbitraryVal) ReplaceBindings(_ map[uint]bool) core.Obj        { return f }

func IsJustArbsCalledsAndReturnVals(obj core.Obj) bool {
	switch v := obj.(type) {
	case core.Called:
		return IsJustArbsCalledsAndReturnVals(v.X) && IsJustArbsCalledsAndReturnVals(v.Y)
	case core.ReturnVal:
		return true
	case FormTestArbitraryVal:
		return true
	default:
		return false
	}
}

func ObjMatchesFormHelper(obj core.Obj) (bool, interface{}) {
	arb, isArb := obj.(FormTestArbitraryVal)
	if isArb {
		return true, arb.Form.AcceptibleVals(arb.Val)
	} else if IsJustArbsCalledsAndReturnVals(obj) {
		return true, false
	} else {
		return false, nil
	}
}

func ObjMatchesForm(obj core.Obj, form Form) bool {
	defer func() {
		recover()
	}()
	for i := uint(1); ; i++ {
		obj = obj.Call(FormTestArbitraryVal{i, &form})
		b, val := core.SimplifyUntilNoPanic(ObjMatchesFormHelper, obj)
		if b {
			return val.(bool)
		} else {
			obj = val.(core.Obj)
		}
	}
}

func ObjToForm(obj core.Obj) Form {
	head, tail := core.ObjToTuple(obj)
	return Form{ func(a uint, b uint) (bool, uint) {
		defer func() { recover() }()
		return true, core.ObjToInt(head.Call(core.ChurchNum{a}).Call(core.ChurchNum{b}))
	},
	func(a uint) bool {
		defer func() {recover()}()
		return core.ObjToBool(tail.Call(core.ChurchNum{a}))
	}}
}

type ObjToStringObj struct {
	F func(core.Obj) string
}

func (f ObjToStringObj) Call(a core.Obj) core.Obj {
	return core.ChurchTupleCharString{f.F(a)}
}
func (f ObjToStringObj) Simplify(_ uint) core.Obj                        { return f }
func (f ObjToStringObj) Replace(_ uint, _ core.Obj) core.Obj             { return f }
func (f ObjToStringObj) GetUnboundVars(_ map[uint]bool, _ map[uint]bool) {}
func (f ObjToStringObj) GetAllVars(_ map[uint]bool)                      {}
func (f ObjToStringObj) ReplaceBindings(_ map[uint]bool) core.Obj        { return f }

func ObjToConversion(obj core.Obj) func(core.Obj, func(core.Obj) string) string {
	return func(a core.Obj, f func(core.Obj) string) string {
		return core.ObjToString(obj.Call(a).Call(ObjToStringObj{f}))
	}
}

func ObjToOptimization(obj core.Obj) *Optimization {
	formObj, tail := core.ObjToTuple(obj)
	conversionObj, tail2 := core.ObjToTuple(tail)
	extraTextObj, importsObj := core.ObjToTuple(tail2)
	importsObjList := core.ObjToList(importsObj)
	imports := make([]string, len(importsObjList))
	for i, strObj := range importsObjList {
		imports[i] = core.ObjToString(strObj)
	}
	returnVal := Optimization{ObjToForm(formObj), ObjToConversion(conversionObj), core.ObjToString(extraTextObj), imports}
	return &returnVal
}

func ObjToOptimizationList(obj core.Obj) []*Optimization {
	list := core.ObjToList(obj)
	returnVal := make([]*Optimization, len(list))
	for k, v := range list {
		returnVal[k] = ObjToOptimization(v)
	}
	return returnVal
}
