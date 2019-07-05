package core

// arbitrary value identified by id
type ArbitraryVal struct {
	ID uint
}

func (f ArbitraryVal) Call(x Obj) Obj                                  { return Called{f, x} }
func (f ArbitraryVal) Simplify() Obj                                   { return f }
func (f ArbitraryVal) SimplifyFully() Obj                              { return f }
func (f ArbitraryVal) Replace(n uint, x Obj) Obj                       { return f }
func (f ArbitraryVal) GetUnboundVars(_ map[uint]bool, _ map[uint]bool) {}
func (f ArbitraryVal) GetAllVars(_ map[uint]bool)                      {}
func (f ArbitraryVal) ReplaceBindings(_ map[uint]bool) Obj             { return f }

// arbitrary method: only accepts Inp as input, returns Otp if given Inp
type ArbitraryMethod struct {
	Inp Obj
	Otp Obj
}

func (f ArbitraryMethod) Call(x Obj) Obj {
	if x.SimplifyFully() == f.Inp {
		return f.Otp
	} else {
		return Called{f, x}
	}
}
func (f ArbitraryMethod) Simplify() Obj                                   { return f }
func (f ArbitraryMethod) SimplifyFully() Obj                              { return f }
func (f ArbitraryMethod) Replace(n uint, x Obj) Obj                       { return f }
func (f ArbitraryMethod) GetUnboundVars(_ map[uint]bool, _ map[uint]bool) {}
func (f ArbitraryMethod) GetAllVars(_ map[uint]bool)                      {}
func (f ArbitraryMethod) ReplaceBindings(_ map[uint]bool) Obj             { return f }

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
	churchChar, isChurchChar := f.(ChurchTupleChar)
	if isChurchChar {
		return churchChar.Char
	} else {
		bools := splitTupleList(splitTupleList(splitTupleList([]Obj{f})))
		returnVal := byte(0)
		for i, f := range bools {
			if ObjToBool(f) {
				returnVal = returnVal | (1 << uint(7-i))
			}
		}
		return returnVal
	}
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
	churchString, isChurchString := f.(ChurchTupleCharString)
	if isChurchString {
		return churchString.Str
	} else {
		list := ObjToList(f)
		returnVal := make([]byte, len(list))
		for i, v := range list {
			returnVal[i] = ObjToByte(v)
		}
		return string(returnVal)
	}
}

func ObjToIO(f Obj) IO {
	isRight, val := ObjToEither(f)
	if isRight {
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

func ObjToType(f Obj) Type {
	isTup, val := ObjToEither(f)
	if isTup {
		a, b := ObjToTuple(val)
		return ArbitraryMethod{TBDType{a}, TBDType{b}}
	} else {
		return ArbitraryVal{ObjToInt(val)}
	}
}
