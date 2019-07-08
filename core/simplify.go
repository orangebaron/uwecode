package core

const maxDepth = uint(10000)

func min(a uint, b uint) uint {
	if a > b {
		return b
	} else {
		return a
	}
}
func SimplifyUntil(f func(Obj) (bool, interface{}), obj Obj) interface{} {
	for depth, newObj := uint(2), obj.Simplify(1); ; depth = min(depth+1, maxDepth) {
		isGood, val := f(newObj)
		if isGood {
			return val
		} else {
			obj = newObj
			newObj = obj.Simplify(depth)
			if newObj == obj && depth == maxDepth {
				break
			}
		}
	}
	panic("Failed to simplify to expected value")
}
