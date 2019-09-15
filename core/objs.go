package core

type Obj interface {
	Call(Obj) Obj
	Simplify(uint) Obj
	Replace(uint, Obj) Obj
	GetUnboundVars(func(uint) bool, chan uint)
	GetAllVars(chan uint)
	ReplaceBindings(map[uint]bool) Obj
}

// the last y in x->y->y
type ReturnVal struct {
	N uint
}

func (f ReturnVal) Call(x Obj) Obj      { return Called{f, x} }
func (f ReturnVal) Simplify(_ uint) Obj { return f }
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
		select {
		case n := <-unbound:
			unboundDict[n] = true
		default:
			break
		}
	}
	f = f.ReplaceBindings(unboundDict).(Function)
	return f.X.Replace(f.N, a)
}
func (f Function) Simplify(depth uint) Obj {
	if depth == 0 {
		return f
	}
	return Function{f.N, f.X.Simplify(depth - 1)}
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
			select {
			case n := <-allVars:
				allVarsDict[n] = true
			default:
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
func (f Called) Simplify(depth uint) Obj {
	if depth == 0 {
		return f
	}
	v := f.X.Call(f.Y)
	if v != f {
		v = v.Simplify(depth - 1)
	}
	return v
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
	done := make(chan bool)
	go func() {
		f.X.GetAllVars(vars)
		close(done)
	}()
	f.Y.GetAllVars(vars)
}
func (f Called) ReplaceBindings(toReplace map[uint]bool) Obj {
	return Called{f.X.ReplaceBindings(toReplace), f.Y.ReplaceBindings(toReplace)}
}

// a -> b -> a (a (a ...Num times... (a b))))
type ChurchNum struct {
	Num uint
}

func (f ChurchNum) Call(a Obj) Obj                                { return CalledChurchNum{f.Num, a} }
func (f ChurchNum) Simplify(_ uint) Obj                           { return f }
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
func (f CalledChurchNum) Simplify(_ uint) Obj       { return f }
func (f CalledChurchNum) Replace(n uint, x Obj) Obj { return CalledChurchNum{f.Num, f.X.Replace(n, x)} }
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
func (f ChurchTupleChar) Simplify(_ uint) Obj                           { return f }
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
func (f ChurchTupleCharString) Simplify(_ uint) Obj                           { return f }
func (f ChurchTupleCharString) Replace(_ uint, _ Obj) Obj                     { return f }
func (f ChurchTupleCharString) GetUnboundVars(_ func(uint) bool, _ chan uint) {}
func (f ChurchTupleCharString) GetAllVars(_ chan uint)                        {}
func (f ChurchTupleCharString) ReplaceBindings(_ map[uint]bool) Obj           { return f }
