package core

type Obj interface {
	Call(Obj) Obj
	Simplify() Obj
	SimplifyFully() Obj
	Replace(uint, Obj) Obj
}

// the last y in x->y->y
type ReturnVal struct {
	N uint
}

func (f ReturnVal) Call(x Obj) Obj     { return Called{f, x} }
func (f ReturnVal) Simplify() Obj      { return f }
func (f ReturnVal) SimplifyFully() Obj { return f }
func (f ReturnVal) Replace(n uint, x Obj) Obj {
	switch n {
	case f.N:
		return x
	default:
		return f
	}
}

// N -> x
type Function struct {
	N uint
	X Obj
}

func (f Function) Call(a Obj) Obj     { return f.X.Replace(f.N, a) }
func (f Function) Simplify() Obj      { return Function{f.N, f.X.Simplify()} }
func (f Function) SimplifyFully() Obj { return Function{f.N, f.X.SimplifyFully()} }
func (f Function) Replace(n uint, x Obj) Obj {
	switch f.N {
	case n:
		return f
	default:
		return Function{f.N, f.X.Replace(n, x)}
	}
}

// a -> (x y) a
type Called struct {
	X Obj
	Y Obj
}

func (f Called) Call(a Obj) Obj { return Called{f.X.Call(f.Y), a} }
func (f Called) Simplify() Obj  { return f.X.Call(f.Y) }
func (f Called) SimplifyFully() Obj {
	v := f.X.Call(f.Y.SimplifyFully())
	_, isCalled := v.(Called)
	if !isCalled {
		v = v.SimplifyFully()
	}
	return v
}
func (f Called) Replace(n uint, x Obj) Obj { return Called{f.X.Replace(n, x), f.Y.Replace(n, x)} }

// a -> b -> a (a (a ...Num times... (a b))))
type ChurchNum struct {
	Num uint
}

func (f ChurchNum) Call(a Obj) Obj            { return CalledChurchNum{f.Num, a} } // TODO: if a is a ChurchNum, return ChurchNum{a.Num ^ f.num}
func (f ChurchNum) Simplify() Obj             { return f }
func (f ChurchNum) SimplifyFully() Obj        { return f }
func (f ChurchNum) Replace(_ uint, _ Obj) Obj { return f }

// a -> X (x (x ...Num times... (x a))))
type CalledChurchNum struct {
	Num uint
	X   Obj
}

func (f CalledChurchNum) Call(a Obj) Obj {
	return CalledCalledChurchNum{f.Num, f.X, a}
}
func (f CalledChurchNum) Simplify() Obj             { return f }
func (f CalledChurchNum) SimplifyFully() Obj        { return f }
func (f CalledChurchNum) Replace(n uint, x Obj) Obj { return CalledChurchNum{f.Num, f.X.Replace(n, x)} }

// a -> (x (x (x ... Num times ... (x y)))) a
type CalledCalledChurchNum struct {
	Num uint
	X   Obj
	Y   Obj
}

func (f CalledCalledChurchNum) Call(a Obj) Obj { return Called{f.Simplify(), a} }
func (f CalledCalledChurchNum) Simplify() Obj {
	switch f.Num {
	case 0:
		return f.Y
	default:
		return CalledCalledChurchNum{f.Num - 1, f.X.Call(f.Y), f.Y}
	}
}
func (f CalledCalledChurchNum) SimplifyFully() Obj {
	for i := uint(0); i < f.Num; i++ {
		f.Y = f.X.Call(f.Y)
	}
	return f.Y
}
func (f CalledCalledChurchNum) Replace(n uint, x Obj) Obj {
	return CalledCalledChurchNum{f.Num, f.X.Replace(n, x), f.Y.Replace(n, x)}
}
