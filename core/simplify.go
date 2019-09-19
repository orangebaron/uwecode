package core

const maxDepth = uint(10000)

func min(a uint, b uint) uint {
	if a > b {
		return b
	} else {
		return a
	}
}
func SimplifyUntilNoPanic(f func(Obj) (bool, interface{}), obj Obj) (bool, interface{}) {
	newObj := obj.Simplify(1, make(chan bool))
	for depth := uint(2); ; depth = min(depth+1, maxDepth) {
		isGood, val := f(newObj)
		if isGood {
			return true, val
		} else {
			obj = newObj
			newObj = obj.Simplify(depth, make(chan bool))
			if newObj == obj {
				if depth == maxDepth {
					break
				} else {
					depth += 1000
				}
			}
		}
	}
	return false, newObj
}
func SimplifyUntil(f func(Obj) (bool, interface{}), obj Obj) interface{} {
	b, val := SimplifyUntilNoPanic(f, obj)
	if !b {
		panic("Failed to simplify to expected value")
	} else {
		return val
	}
}
