package core

// arbitrary value identified by id
type ArbitraryVal struct {
	ID uint
}

func (f ArbitraryVal) Call(x Obj) Obj                                { return Called{f, x} }
func (f ArbitraryVal) Simplify(_ SimplifyState) Obj                  { return f }
func (f ArbitraryVal) Replace(_ uint, _ Obj) Obj                     { return f }
func (f ArbitraryVal) GetUnboundVars(_ func(uint) bool, _ chan uint) {}
func (f ArbitraryVal) GetAllVars(_ chan uint)                        {}
func (f ArbitraryVal) ReplaceBindings(_ map[uint]bool) Obj           { return f }

func objToIntHelper(f Obj) (bool, interface{}) {
	n := uint(0)
	for {
		switch v := f.(type) {
		case Called:
			n++
			f = v.Y
			arb, isArb := v.X.(ArbitraryVal)
			if !isArb || arb.ID != 0 {
				return false, nil
			}
		case ArbitraryVal:
			return v.ID == 1, n
		default:
			return false, nil
		}
	}
}
func ObjToInt(f Obj, state SimplifyState) uint {
	cn, isChurchNum := f.(ChurchNum)
	if isChurchNum {
		return cn.Num
	} else {
		return SimplifyUntil(objToIntHelper, f.Call(ArbitraryVal{0}).Call(ArbitraryVal{1}), state).(uint)
	}
}

func objToBoolHelper(f Obj) (bool, interface{}) {
	arb, isArb := f.(ArbitraryVal)
	return isArb, (isArb && arb.ID == 0)
}
func ObjToBool(f Obj, state SimplifyState) bool {
	return SimplifyUntil(objToBoolHelper, f.Call(ArbitraryVal{0}).Call(ArbitraryVal{1}), state).(bool)
}

func objToTupleHelper(f Obj) (bool, interface{}) {
	called, isCalled := f.(Called)
	if !isCalled {
		return false, nil
	}
	called2, isCalled2 := called.X.(Called)
	if !isCalled2 {
		return false, nil
	}
	_, isArb := called2.X.(ArbitraryVal) // TODO: check the ID??????
	return isArb, [2]Obj{called2.Y, called.Y}
}
func ObjToTuple(f Obj, state SimplifyState) (Obj, Obj) {
	x := SimplifyUntil(objToTupleHelper, f.Call(ArbitraryVal{0}), state).([2]Obj)
	return x[0], x[1]
}

type maybeEither struct {
	b bool
	Obj
}

func objToMaybeHelper(f Obj) (bool, interface{}) {
	called, isCalled := f.(Called)
	if isCalled {
		arb, isArb := called.X.(ArbitraryVal)
		return isArb && arb.ID == 0, maybeEither{true, called.Y}
	} else {
		arb, isArb := f.(ArbitraryVal)
		return isArb && arb.ID == 1, maybeEither{false, nil}
	}
}
func ObjToMaybe(f Obj, state SimplifyState) (bool, Obj) {
	maybe := SimplifyUntil(objToMaybeHelper, f.Call(ArbitraryVal{0}).Call(ArbitraryVal{1}), state).(maybeEither)
	return maybe.b, maybe.Obj
}

func objToEitherHelper(f Obj) (bool, interface{}) {
	called, isCalled := f.(Called)
	if !isCalled {
		return false, nil
	}
	arb, isArb := called.X.(ArbitraryVal)
	return isArb, maybeEither{arb.ID == 1, called.Y}
}
func ObjToEither(f Obj, state SimplifyState) (bool, Obj) {
	either := SimplifyUntil(objToEitherHelper, f.Call(ArbitraryVal{0}).Call(ArbitraryVal{1}), state).(maybeEither)
	return either.b, either.Obj
}

func splitTupleList(fs []Obj, state SimplifyState) []Obj {
	returnVal := make([]Obj, len(fs)*2)
	for i, f := range fs {
		returnVal[i*2], returnVal[i*2+1] = ObjToTuple(f, state)
	}
	return returnVal
}

func ObjToByte(f Obj, state SimplifyState) byte {
	churchChar, isChurchChar := f.(ChurchTupleChar)
	if isChurchChar {
		return churchChar.Char
	} else {
		bools := splitTupleList(splitTupleList(splitTupleList([]Obj{f}, state), state), state)
		returnVal := byte(0)
		for i, f := range bools {
			if ObjToBool(f, state) {
				returnVal = returnVal | (1 << uint(7-i))
			}
		}
		return returnVal
	}
}

func ObjToList(f Obj, state SimplifyState) []Obj {
	isSomething, val := ObjToMaybe(f, state)
	if isSomething {
		head, tail := ObjToTuple(val, state)
		return append([]Obj{head}, ObjToList(tail, state)...)
	} else {
		return []Obj{}
	}
}

func ObjToString(f Obj, state SimplifyState) string {
	churchString, isChurchString := f.(ChurchTupleCharString)
	if isChurchString {
		return churchString.Str
	} else {
		list := ObjToList(f, state)
		returnVal := make([]byte, len(list))
		for i, v := range list {
			returnVal[i] = ObjToByte(v, state)
		}
		return string(returnVal)
	}
}

func ObjToIO(f Obj, state SimplifyState) IO {
	isRight, val := ObjToEither(f, state)
	if isRight {
		isFork, val2 := ObjToMaybe(val, state)
		if isFork {
			objA, objB := ObjToTuple(val2, state)
			return ForkIO{objA, objB}
		} else {
			return NullIO{}
		}
	} else {
		isOutput, val2 := ObjToEither(val, state)
		if isOutput {
			str, obj := ObjToTuple(val2, state)
			return OutputIO{str, obj}
		} else {
			return InputIO{val2}
		}
	}
}
