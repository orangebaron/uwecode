package core

import "sync"

type GlobalState struct {
	*sync.WaitGroup
	Stop chan struct{}
}

func MakeGlobalState() GlobalState {
	return GlobalState{&sync.WaitGroup{}, make(chan struct{})}
}

type SimplifyState struct {
	GlobalState
	*sync.Mutex
	AlreadySimplified map[Obj]Obj
	TryingToSimplify  map[Obj]chan struct{}
	SimplifyStack     []Obj
}

func MakeSimplifyState() SimplifyState {
	return SimplifyState{MakeGlobalState(), &sync.Mutex{}, make(map[Obj]Obj), make(map[Obj]chan struct{}), make([]Obj, 0)}
}

type Obj interface {
	Call(Obj) Obj
	Simplify(SimplifyState) Obj
	Replace(uint, Obj) Obj
	GetUnboundVars(func(uint) bool, chan uint)
	GetAllVars(chan uint)
	ReplaceBindings(map[uint]bool) Obj
}

// the last y in x->y->y
type ReturnVal struct {
	N uint
}

func (f ReturnVal) Call(x Obj) Obj               { return Called{f, x} }
func (f ReturnVal) Simplify(_ SimplifyState) Obj { return f }
func (f ReturnVal) Replace(n uint, x Obj) Obj {
	if n == f.N {
		return x
	} else {
		return f
	}
}
func (f ReturnVal) GetUnboundVars(bound func(uint) bool, unbound chan uint) {
	if !bound(f.N) {
		unbound <- f.N
	}
}
func (f ReturnVal) GetAllVars(vars chan uint)           { vars <- f.N }
func (f ReturnVal) ReplaceBindings(_ map[uint]bool) Obj { return f }

// N -> x
type Function struct {
	N uint
	X Obj
}

func (f Function) Call(a Obj) Obj {
	unbound := make(chan uint)
	go func() {
		a.GetUnboundVars(func(_ uint) bool { return false }, unbound)
		close(unbound)
	}()
	unboundDict := make(map[uint]bool)
	for {
		val, isGood := <-unbound
		if isGood {
			unboundDict[val] = true
		} else {
			break
		}
	}
	f = f.ReplaceBindings(unboundDict).(Function)
	return f.X.Replace(f.N, a)
}
func (f Function) Simplify(state SimplifyState) Obj {
	select {
	case <-state.Stop:
		return f
	default:
		return Function{f.N, f.X.Simplify(state)}
	}
}
func (f Function) Replace(n uint, x Obj) Obj {
	if n == f.N {
		return f
	} else {
		return Function{f.N, f.X.Replace(n, x)}
	}
}
func (f Function) GetUnboundVars(bound func(uint) bool, unbound chan uint) {
	f.X.GetUnboundVars(func(n uint) bool { return n == f.N || bound(n) }, unbound)
}
func (f Function) GetAllVars(vars chan uint) {
	vars <- f.N
	f.X.GetAllVars(vars)
}
func (f Function) ReplaceBindings(toReplace map[uint]bool) Obj {
	if toReplace[f.N] {
		allVars := make(chan uint)
		go func() {
			f.X.GetAllVars(allVars)
			close(allVars)
		}()
		allVarsDict := make(map[uint]bool)
		for k, v := range toReplace {
			allVarsDict[k] = v
		}
		for {
			val, isGood := <-allVars
			if isGood {
				allVarsDict[val] = true
			} else {
				break
			}
		}
		var newN uint
		for newN = uint(0); allVarsDict[newN]; newN++ {
		}
		f.X = f.X.Replace(f.N, ReturnVal{newN})
		f.N = newN
	}
	f.X = f.X.ReplaceBindings(toReplace)
	return f
}

// a -> (x y) a
type Called struct {
	X Obj
	Y Obj
}

func (f Called) Call(a Obj) Obj { return Called{f.X.Call(f.Y), a} }
func (f Called) Simplify(state SimplifyState) Obj {
	select {
	case <-state.Stop:
		return f
	default:
	}
	state.Mutex.Lock()
	// TODO: do this stuff in the background
	if v, exists := state.AlreadySimplified[f]; exists {
		state.Mutex.Unlock()
		return v
	}
	for _, v := range state.SimplifyStack {
		if v == f {
			state.Mutex.Unlock()
			return f
		}
	}
	for {
		if channel, exists := state.TryingToSimplify[f]; exists {
			state.Mutex.Unlock()
			<-channel
			state.Mutex.Lock()
			v, exists := state.AlreadySimplified[f]
			if exists {
				state.Mutex.Unlock()
				return v
			}
		} else {
			break
		}
	}
	state.TryingToSimplify[f] = make(chan struct{})
	state.Mutex.Unlock()
	returnVal := make(chan Obj, 3)
	otherSimplifiedVal := make(chan Obj, 1)
	newState := state
	newState.GlobalState.Stop = make(chan struct{})
	newState.SimplifyStack = append(newState.SimplifyStack, f)
	otherSimplifiedValMutex := &sync.Mutex{}
	state.WaitGroup.Add(3)
	go func() {
		called := f.X.Call(f.Y)
		if called != f {
			returnVal <- called.Simplify(newState)
		}
		state.WaitGroup.Done()
	}()
	go func() {
		newX := f.X.Simplify(newState)
		otherSimplifiedValMutex.Lock()
		if len(otherSimplifiedVal) == 0 {
			otherSimplifiedVal <- newX
			otherSimplifiedValMutex.Unlock()
			called := newX.Call(f.Y)
			if called != f {
				returnVal <- called.Simplify(newState)
			}
		} else {
			other := <-otherSimplifiedVal
			otherSimplifiedValMutex.Unlock()
			called := newX.Call(other)
			if called != f {
				returnVal <- called.Simplify(newState)
			} else {
				returnVal <- called
			}
		}
		state.WaitGroup.Done()
	}()
	go func() {
		newY := f.Y.Simplify(newState)
		otherSimplifiedValMutex.Lock()
		if len(otherSimplifiedVal) == 0 {
			otherSimplifiedVal <- newY
			otherSimplifiedValMutex.Unlock()
			called := f.X.Call(newY)
			if called != f {
				returnVal <- called.Simplify(newState)
			}
		} else {
			other := <-otherSimplifiedVal
			otherSimplifiedValMutex.Unlock()
			called := other.Call(newY)
			if called != f {
				returnVal <- called.Simplify(newState)
			} else {
				returnVal <- called
			}
		}
		state.WaitGroup.Done()
	}()
	select {
	case v := <-returnVal:
		close(newState.GlobalState.Stop)
		state.Mutex.Lock()
		state.AlreadySimplified[f] = v
		close(state.TryingToSimplify[f])
		delete(state.TryingToSimplify, f)
		state.Mutex.Unlock()
		return v
	case <-state.GlobalState.Stop:
		close(newState.GlobalState.Stop)
		state.Mutex.Lock()
		close(state.TryingToSimplify[f])
		delete(state.TryingToSimplify, f)
		state.Mutex.Unlock()
		return f
	}
}
func (f Called) Replace(n uint, x Obj) Obj {
	a := make(chan Obj)
	go func() {
		a <- f.X.Replace(n, x)
	}()
	b := f.Y.Replace(n, x)
	return Called{<-a, b}
}
func (f Called) GetUnboundVars(bound func(uint) bool, unbound chan uint) {
	done := make(chan bool)
	go func() {
		f.X.GetUnboundVars(bound, unbound)
		close(done)
	}()
	f.Y.GetUnboundVars(bound, unbound)
	<-done
}
func (f Called) GetAllVars(vars chan uint) {
	done := make(chan struct{})
	go func() {
		f.X.GetAllVars(vars)
		close(done)
	}()
	f.Y.GetAllVars(vars)
	<-done
}
func (f Called) ReplaceBindings(toReplace map[uint]bool) Obj {
	return Called{f.X.ReplaceBindings(toReplace), f.Y.ReplaceBindings(toReplace)}
}

// a -> b -> a (a (a ...Num times... (a b))))
type ChurchNum struct {
	Num uint
}

func (f ChurchNum) Call(a Obj) Obj                                { return CalledChurchNum{f.Num, a} }
func (f ChurchNum) Simplify(_ SimplifyState) Obj                  { return f }
func (f ChurchNum) Replace(_ uint, _ Obj) Obj                     { return f }
func (f ChurchNum) GetUnboundVars(_ func(uint) bool, _ chan uint) {}
func (f ChurchNum) GetAllVars(_ chan uint)                        {}
func (f ChurchNum) ReplaceBindings(_ map[uint]bool) Obj           { return f }

// a -> X (x (x ...Num times... (x a))))
type CalledChurchNum struct {
	Num uint
	X   Obj
}

func (f CalledChurchNum) Call(a Obj) Obj {
	for i := uint(0); i < f.Num; i++ {
		a = f.X.Call(a)
	}
	return a
}
func (f CalledChurchNum) Simplify(_ SimplifyState) Obj { return f }
func (f CalledChurchNum) Replace(n uint, x Obj) Obj    { return CalledChurchNum{f.Num, f.X.Replace(n, x)} }
func (f CalledChurchNum) GetUnboundVars(bound func(uint) bool, unbound chan uint) {
	f.X.GetUnboundVars(bound, unbound)
}
func (f CalledChurchNum) GetAllVars(vars chan uint) {
	f.X.GetAllVars(vars)
}
func (f CalledChurchNum) ReplaceBindings(toReplace map[uint]bool) Obj {
	return CalledChurchNum{f.Num, f.X.ReplaceBindings(toReplace)}
}

type ChurchTupleChar struct {
	Char byte
}

var falseObj Obj = Function{0, Function{1, ReturnVal{1}}}
var trueObj Obj = Function{0, Function{1, ReturnVal{0}}}

func tuple(a Obj, b Obj) Obj {
	return Function{0, Called{Called{ReturnVal{0}, a}, b}}
}
func (f ChurchTupleChar) ToNormalObj() Obj {
	bools := make([]Obj, 8)
	for i := uint(0); i < 8; i++ {
		if (f.Char>>i)&byte(1) == 0 {
			bools[7-i] = falseObj
		} else {
			bools[7-i] = trueObj
		}
	}
	return tuple(tuple(tuple(bools[0], bools[1]), tuple(bools[2], bools[3])), tuple(tuple(bools[4], bools[5]), tuple(bools[6], bools[7])))
}
func (f ChurchTupleChar) Call(a Obj) Obj                                { return Called{f.ToNormalObj(), a} }
func (f ChurchTupleChar) Simplify(_ SimplifyState) Obj                  { return f }
func (f ChurchTupleChar) Replace(_ uint, _ Obj) Obj                     { return f }
func (f ChurchTupleChar) GetUnboundVars(_ func(uint) bool, _ chan uint) {}
func (f ChurchTupleChar) GetAllVars(_ chan uint)                        {}
func (f ChurchTupleChar) ReplaceBindings(_ map[uint]bool) Obj           { return f }

type ChurchTupleCharString struct {
	Str string
}

func just(a Obj) Obj {
	return Function{0, Function{1, Called{ReturnVal{0}, a}}}
}
func (f ChurchTupleCharString) ToNormalObj() Obj {
	if len(f.Str) == 0 {
		return falseObj
	} else {
		return just(tuple(ChurchTupleChar{f.Str[0]}, ChurchTupleCharString{f.Str[1:]}))
	}
}
func (f ChurchTupleCharString) Call(a Obj) Obj                                { return Called{f.ToNormalObj(), a} }
func (f ChurchTupleCharString) Simplify(_ SimplifyState) Obj                  { return f }
func (f ChurchTupleCharString) Replace(_ uint, _ Obj) Obj                     { return f }
func (f ChurchTupleCharString) GetUnboundVars(_ func(uint) bool, _ chan uint) {}
func (f ChurchTupleCharString) GetAllVars(_ chan uint)                        {}
func (f ChurchTupleCharString) ReplaceBindings(_ map[uint]bool) Obj           { return f }
