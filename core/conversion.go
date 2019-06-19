package core

// arbitrary value identified by id
type ArbitraryVal struct {
	ID uint
}

func (f ArbitraryVal) Call(x Obj) Obj            { return Called{f, x} }
func (f ArbitraryVal) Simplify() Obj             { return f }
func (f ArbitraryVal) SimplifyFully() Obj        { return f }
func (f ArbitraryVal) Replace(n uint, x Obj) Obj { return f }

// assumes that the given Obj is actually a number
func ObjToInt(f Obj) uint {
	cn, isChurchNum := f.(ChurchNum)
	if isChurchNum {
		return cn.Num
	} else {
		n := uint(0)
		c := f.Call(ArbitraryVal{0}).Call(ArbitraryVal{1}).SimplifyFully()
		for {
			switch v := c.(type) {
			case Called:
				n++
				c = v.Y
			case ArbitraryVal:
				return n
			}
		}
	}
}

func ObjToBool(f Obj) bool {
	c := f.Call(ArbitraryVal{0}).Call(ArbitraryVal{1}).SimplifyFully()
	return c.(ArbitraryVal).ID == 0
}

// assumes that the given Obj is actually a tuple
func ObjToTuple(f Obj) (Obj, Obj) {
	c := f.Call(ArbitraryVal{0}).SimplifyFully().(Called)
	return c.X.(Called).Y, c.Y
}

func ObjToMaybe(f Obj) (bool, Obj) {
	c := f.Call(ArbitraryVal{0}).Call(ArbitraryVal{1}).SimplifyFully()
	called, isCalled := c.(Called)
	return isCalled, called.Y
}

func ObjToEither(f Obj) (bool, Obj) {
	c := f.Call(ArbitraryVal{0}).Call(ArbitraryVal{1}).SimplifyFully()
	called := c.(Called)
	return called.X.(ArbitraryVal).ID == 1, called.Y
}
