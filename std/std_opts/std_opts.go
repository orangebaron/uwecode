package std_opts

import "../../../core"
import "../../../optimize"

var listOpt optimize.Optimization = optimize.Optimization{
	func(f core.Obj, convert func(core.Obj) string) string {
		defer func() { recover() }()
		str := "stdOpts.ObjList{[]core.Obj{"
		for _, obj := range core.ObjToList(f) {
			str += convert(obj) + ","
		}
		str += "}}"
		return str
	},
	"./.uwe/opts/std_opts",
}
var OptsList []*optimize.Optimization = []*optimize.Optimization{&listOpt}
