package core

type Obj interface {
	Call(Obj) Obj
	Simplify() Obj
	SimplifyFully() Obj
	Replace(int, Obj) Obj
}

// the last y in x->y->y
type ReturnVal struct {
	N int
}

func (f ReturnVal) Call(x Obj) Obj     { return Called{f, x} }
func (f ReturnVal) Simplify() Obj      { return f }
func (f ReturnVal) SimplifyFully() Obj { return f }
func (f ReturnVal) Replace(n int, x Obj) Obj {
	switch n {
	case f.N:
		return x
	default:
		return f
	}
}

// N -> x
type Function struct {
	N int
	X Obj
}

func (f Function) Call(a Obj) Obj     { return f.X.Replace(f.N, a) }
func (f Function) Simplify() Obj      { return Function{f.N, f.X.Simplify()} }
func (f Function) SimplifyFully() Obj { return Function{f.N, f.X.SimplifyFully()} }
func (f Function) Replace(n int, x Obj) Obj {
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
	switch f.X.(type) {
	case ReturnVal:
		return Called{f.X, f.Y.SimplifyFully()}
	case ArbitraryVal:
		return Called{f.X, f.Y.SimplifyFully()}
	default:
		return f.X.Call(f.Y).SimplifyFully()
	}
}
func (f Called) Replace(n int, x Obj) Obj { return Called{f.X.Replace(n, x), f.Y.Replace(n, x)} }

// a -> b -> a (a (a ...Num times... (a b))))
type ChurchNum struct {
	Num int
}

func (f ChurchNum) Call(a Obj) Obj           { return CalledChurchNum{f.Num, a} } // TODO: if a is a ChurchNum, return ChurchNum{a.Num ^ f.num}
func (f ChurchNum) Simplify() Obj            { return f }
func (f ChurchNum) SimplifyFully() Obj       { return f }
func (f ChurchNum) Replace(_ int, _ Obj) Obj { return f }

// a -> X (x (x ...Num times... (x a))))
type CalledChurchNum struct {
	Num int
	X   Obj
}

func (f CalledChurchNum) Call(a Obj) Obj {
	return CalledCalledChurchNum{f.Num, f.X, a}
}
func (f CalledChurchNum) Simplify() Obj            { return f }
func (f CalledChurchNum) SimplifyFully() Obj       { return f }
func (f CalledChurchNum) Replace(_ int, _ Obj) Obj { return f }

// a -> (x (x (x ... Num times ... (x y)))) a
type CalledCalledChurchNum struct {
	Num int
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
	for i := 0; i < f.Num; i++ {
		f.Y = f.X.Call(f.Y)
	}
	return f.Y
}
func (f CalledCalledChurchNum) Replace(n int, x Obj) Obj {
	return CalledCalledChurchNum{f.Num, f.X.Replace(n, x), f.Y.Replace(n, x)}
}

// N -> f -> X -> (n f) (f x) aka n -> n+1
type IncrFunction struct{}

var IncrFunctionNormalForm = Function{0, Function{1, Function{2, Called{Called{ReturnVal{0}, ReturnVal{1}}, Called{ReturnVal{1}, ReturnVal{2}}}}}}

func (f IncrFunction) Call(a Obj) Obj {
	switch at := a.(type) {
	case ChurchNum:
		return ChurchNum{at.Num + 1}
	default:
		return IncrFunctionNormalForm.Call(a)
	}
}
func (f IncrFunction) Simplify() Obj            { return f }
func (f IncrFunction) SimplifyFully() Obj       { return f }
func (f IncrFunction) Replace(n int, x Obj) Obj { return f }

// arbitrary value identified by id
type ArbitraryVal struct {
	id int
}

func (f ArbitraryVal) Call(x Obj) Obj           { return Called{f, x} }
func (f ArbitraryVal) Simplify() Obj            { return f }
func (f ArbitraryVal) SimplifyFully() Obj       { return f }
func (f ArbitraryVal) Replace(n int, x Obj) Obj { return f }
