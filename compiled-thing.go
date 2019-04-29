package main

type obj interface {
	call(obj) obj
	simplify() obj
	replace(int, obj) obj
}

// a -> a
type id struct{}

func (_ id) call(a obj) obj           { return a }
func (_ id) simplify() obj            { return id{} }
func (_ id) replace(_ int, _ obj) obj { return id{} }

// a -> x where x is the same regardless of a
type consObj struct {
	x obj
}

func (f consObj) call(_ obj) obj           { return f.x }
func (f consObj) simplify() obj            { return consObj{f.x.simplify()} }
func (f consObj) replace(_ int, _ obj) obj { return f }

// a -> (value of n) where value of n is the same regardless of a
type consInt struct {
	n int
}

func (_ consInt) call(_ obj) obj { panic("consInt was called (should never happen)") }
func (f consInt) simplify() obj  { return f }
func (f consInt) replace(n int, x obj) obj {
	switch n {
	case f.n:
		return x
	default:
		return consObj{x}
	}
}

// n -> x
type function struct {
	n int
	x obj
}

func (f function) call(a obj) obj { return f.x.replace(f.n, a) }
func (f function) simplify() obj  { return function{f.n, f.x.simplify()} }
func (f function) replace(n int, x obj) obj {
	switch f.n {
	case n:
		return f
	default:
		return function{f.n, f.x.replace(n, x)}
	}
}

// a -> (x y) a
type called struct {
	x obj
	y obj
}

func (f called) call(a obj) obj           { return called{f.x.call(f.y), a} }
func (f called) simplify() obj            { return f.x.call(f.y) }
func (f called) replace(n int, x obj) obj { return called{f.x.replace(n, x), f.y.replace(n, x)} }

func main() {}
