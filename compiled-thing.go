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

func (f called) call(a obj) obj           { return called{f.x.call(f.y), a} }
func (f called) simplify() obj            { return f.x.call(f.y) }
func (f called) simplifyFully() obj       { return f.x.call(f.y).simplifyFully() }
func (f called) replace(n int, x obj) obj { return called{f.x.replace(n, x), f.y.replace(n, x)} }

func main() {
	two := function{0, function{1, called{returnVal{0}, called{returnVal{0}, returnVal{1}}}}}                                     // f -> x -> f (f x)
	incr := function{0, function{1, function{2, called{called{returnVal{0}, returnVal{1}}, called{returnVal{1}, returnVal{2}}}}}} // n -> f -> x -> n f (f x)
	three := incr.call(two)                                                                                                       // f -> x -> f (f (f x))
	aba := function{0, function{1, returnVal{0}}}                                                                                 // a -> b -> a
	nine := two.call(three)                                                                                                       // don't feel like typing this out
	veryCool := nine.call(aba).call(aba)                                                                                          // f1 -> f2 -> ... -> f9 -> a -> b -> a
	fmt.Println(veryCool.simplify())
	fmt.Println(veryCool.simplifyFully())
}
