package core

import "sync"

const maxDepth = uint(10000)

func min(a uint, b uint) uint {
	if a > b {
		return b
	} else {
		return a
	}
}
func SimplifyUntilNoPanic(f func(Obj) (bool, interface{}), obj Obj) (bool, interface{}) {
	state := SimplifyState{GlobalState{&sync.WaitGroup{}, make(chan struct{})}, &sync.Mutex{}, make(map[Obj]Obj), make(map[Obj]chan struct{}), make([]Obj, 0)}
	newObj := obj.Simplify(state)
	for {
		isGood, val := f(newObj)
		if isGood {
			return true, val
		} else {
			obj = newObj
			newObj = obj.Simplify(state)
			if newObj == obj {
				break
			}
		}
	}
	state.GlobalState.WaitGroup.Wait()
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
