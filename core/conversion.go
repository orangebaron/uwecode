package core

// assumes that the given Obj is actually a number
func ObjToInt(f Obj) uint {
	var n uint = 0
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

func ObjToBool(f Obj) bool {
	c := f.Call(ArbitraryVal{0}).Call(ArbitraryVal{1}).SimplifyFully()
	return c.(ArbitraryVal).id == 0
}
