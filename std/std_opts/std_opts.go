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
var OptsList []*optimize.Optimization = []*optimize.Optimization{&listOpt}
