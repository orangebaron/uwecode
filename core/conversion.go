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

// assumes that the given Obj is actually a maybe
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

func splitTupleList(fs []Obj) []Obj {
	returnVal := make([]Obj, len(fs)*2)
	for i, f := range fs {
		returnVal[i*2], returnVal[i*2+1] = ObjToTuple(f)
	}
	return returnVal
}

func ObjToByte(f Obj) byte {
	bools := splitTupleList(splitTupleList(splitTupleList([]Obj{f})))
	returnVal := byte(0)
	for i, f := range bools {
		if ObjToBool(f) {
			returnVal = returnVal | (1 << uint(7-i))
		}
	}
	return returnVal
}

func ObjToList(f Obj) []Obj {
	isSomething, val := ObjToMaybe(f)
	if isSomething {
		head, tail := ObjToTuple(val)
		return append([]Obj{head}, ObjToList(tail)...)
	} else {
		return []Obj{}
	}
}

func ObjToString(f Obj) string {
	list := ObjToList(f)
	returnVal := make([]byte, len(list))
	for i, v := range list {
		returnVal[i] = ObjToByte(v)
	}
	return string(returnVal)
}

func ObjToIO(f Obj) IO {
	isFork, val := ObjToEither(f)
	if isFork {
		isFork, val2 := ObjToMaybe(val)
		if isFork {
			objA, objB := ObjToTuple(val2)
			return ForkIO{objA, objB}
		} else {
			return NullIO{}
		}
	} else {
		isOutput, val2 := ObjToEither(val)
		if isOutput {
			str, obj := ObjToTuple(val2)
			return OutputIO{str, obj}
		} else {
			return InputIO{val2}
		}
	}
}
