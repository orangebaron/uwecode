package main

import "fmt"

type obj interface {
	call(obj) obj
	simplify() obj
	simplifyFully() obj
	replace(int, obj) obj
}

// the last y in x->y->y
type returnVal struct {
	n int
}

func (f returnVal) call(x obj) obj     { return called{f, x} }
func (f returnVal) simplify() obj      { return f }
func (f returnVal) simplifyFully() obj { return f }
func (f returnVal) replace(n int, x obj) obj {
	switch n {
	case f.n:
		return x
	default:
		return f
	}
}

// n -> x
type function struct {
	n int
	x obj
}

func (f function) call(a obj) obj     { return f.x.replace(f.n, a) }
func (f function) simplify() obj      { return function{f.n, f.x.simplify()} }
func (f function) simplifyFully() obj { return function{f.n, f.x.simplifyFully()} }
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

func (f called) call(a obj) obj { return called{f.x.call(f.y), a} }
func (f called) simplify() obj  { return f.x.call(f.y) }
func (f called) simplifyFully() obj {
	switch f.x.(type) {
	case returnVal:
		return called{f.x, f.y.simplifyFully()}
	default:
		return f.x.call(f.y).simplifyFully()
	}
}
func (f called) replace(n int, x obj) obj { return called{f.x.replace(n, x), f.y.replace(n, x)} }

// a -> b -> a (a (a ...num times... (a b))))
type churchNum struct {
	num int
}

func (f churchNum) call(a obj) obj           { return calledChurchNum{f.num, a} } // TODO: if a is a churchNum, return churchNum{a.num ^ f.num}
func (f churchNum) simplify() obj            { return f }
func (f churchNum) simplifyFully() obj       { return f }
func (f churchNum) replace(_ int, _ obj) obj { return f }

// a -> x (x (x ...num times... (x a))))
type calledChurchNum struct {
	num int
	x   obj
}

func (f calledChurchNum) call(a obj) obj {
	return calledCalledChurchNum{f.num, f.x, a}
}
func (f calledChurchNum) simplify() obj            { return f }
func (f calledChurchNum) simplifyFully() obj       { return f }
func (f calledChurchNum) replace(_ int, _ obj) obj { return f }

// a -> (x (x (x ... num times ... (x y)))) a
type calledCalledChurchNum struct {
	num int
	x   obj
	y   obj
}

func (f calledCalledChurchNum) call(a obj) obj { return called{f.simplify(), a} }
func (f calledCalledChurchNum) simplify() obj {
	switch f.num {
	case 0:
		return f.y
	default:
		return calledCalledChurchNum{f.num - 1, f.x.call(f.y), f.y}
	}
}
func (f calledCalledChurchNum) simplifyFully() obj {
	for i := 0; i < f.num; i++ {
		f.y = f.x.call(f.y)
	}
	return f.y
}
func (f calledCalledChurchNum) replace(n int, x obj) obj {
	return calledCalledChurchNum{f.num, f.x.replace(n, x), f.y.replace(n, x)}
}

//
type incrFunction struct {}

var incrFunctionNormalForm = function{0, function{1, function{2, called{called{returnVal{0}, returnVal{1}}, called{returnVal{1}, returnVal{2}}}}}}

func (f incrFunction) call(a obj) obj {
	switch at := a.(type) {
	case churchNum:
		return churchNum{at.num + 1}
	default:
		return incrFunctionNormalForm.call(a)
	}
}
func (f incrFunction) simplify() obj { return f }
func (f incrFunction) simplifyFully() obj { return f }
func (f incrFunction) replace(n int, x obj) obj { return f }

func main() {
	two := churchNum{2}
	three := incrFunction{}.call(two)
	aba := function{0, function{1, returnVal{0}}}
	nine := two.call(three)
	veryCool := nine.call(aba).call(aba)
	fmt.Printf("%+v\n", veryCool.simplifyFully())
	fmt.Println(churchNum{3}.call(incrFunction{}).call(churchNum{0}).simplifyFully())
}
