package compiler

import "strconv"

import "../core"

type Expression interface {
	ToObj(map[string]core.Obj, uint) core.Obj
	AddWordToEnd(string) Expression
	AddExpressionToEnd(Expression) Expression
}

func contains(l string, b byte) bool {
	for _, c := range []byte(l) {
		if c == b {
			return true
		}
	}
	return false
}

func isFirstCharacter(l string, b byte) bool {
	bs := []byte(l)
	return len(bs) > 0 && bs[0] == b
}

func IsInfixCall(word string) bool {
	return isFirstCharacter(word, '`')
}

func WordToExpression(word string) Expression {
	n, err := strconv.Atoi(word)
	if err == nil && n > 0 { // what about numbers that're too big?
		return NumExpression{uint(n)}
	} else if isFirstCharacter(word, '\\') {
		return FunctionExpression{word[1:], NullExpression{}}
	} else if IsInfixCall(word) {
		return VarExpression{word[1:]}
	} else {
		return VarExpression{word}
	}
}

func DummyAddWordToEnd(e Expression, word string) Expression {
	// TODO: if the word is the start of a function literal, things won't get appended to it afterwards; throw an error to tell the user not to do that?
	exp := WordToExpression(word)
	if IsInfixCall(word) {
		return InfixCallExpression{exp, e, NullExpression{}}
	} else {
		return CallExpression{e, exp}
	}
}

func DummyAddExpressionToEnd(e Expression, added Expression) Expression {
	return CallExpression{e, added}
}

type CallExpression struct {
	F Expression
	X Expression
}

func (e CallExpression) ToObj(dict map[string]core.Obj, biggestNum uint) core.Obj {
	return core.Called{e.F.ToObj(dict, biggestNum), e.X.ToObj(dict, biggestNum)}
}

func (e CallExpression) AddWordToEnd(word string) Expression {
	return DummyAddWordToEnd(e, word)
}

func (e CallExpression) AddExpressionToEnd(added Expression) Expression {
	return DummyAddExpressionToEnd(e, added)
}

type FunctionExpression struct {
	ArgName  string
	Returned Expression
}

func (e FunctionExpression) ToObj(dict map[string]core.Obj, biggestNum uint) core.Obj {
	oldVal, valWasPresent := dict[e.ArgName]
	dict[e.ArgName] = core.ReturnVal{biggestNum}
	returnVal := core.Function{biggestNum, e.Returned.ToObj(dict, biggestNum+1)}
	if valWasPresent {
		dict[e.ArgName] = oldVal
	} else {
		delete(dict, e.ArgName)
	}
	return returnVal
}

func (e FunctionExpression) AddWordToEnd(word string) Expression {
	return FunctionExpression{e.ArgName, e.Returned.AddWordToEnd(word)}
}

func (e FunctionExpression) AddExpressionToEnd(added Expression) Expression {
	return FunctionExpression{e.ArgName, e.Returned.AddExpressionToEnd(added)}
}

type NumExpression struct {
	Num uint
}

func (e NumExpression) ToObj(_ map[string]core.Obj, _ uint) core.Obj {
	return core.ChurchNum{e.Num}
}

func (e NumExpression) AddWordToEnd(word string) Expression {
	return DummyAddWordToEnd(e, word)
}

func (e NumExpression) AddExpressionToEnd(added Expression) Expression {
	return DummyAddExpressionToEnd(e, added)
}

type VarExpression struct {
	Name string
}

func (e VarExpression) ToObj(dict map[string]core.Obj, _ uint) core.Obj {
	return dict[e.Name]
}

func (e VarExpression) AddExpressionToEnd(added Expression) Expression {
	return DummyAddExpressionToEnd(e, added)
}

func (e VarExpression) AddWordToEnd(word string) Expression {
	return DummyAddWordToEnd(e, word)
}

type InfixCallExpression struct {
	F Expression
	A Expression
	B Expression
}

func (e InfixCallExpression) ToObj(dict map[string]core.Obj, biggestNum uint) core.Obj {
	return core.Called{core.Called{e.F.ToObj(dict, biggestNum), e.A.ToObj(dict, biggestNum)}, e.B.ToObj(dict, biggestNum)}
}

func (e InfixCallExpression) AddWordToEnd(word string) Expression {
	return InfixCallExpression{e.F, e.A, e.B.AddWordToEnd(word)}
}

func (e InfixCallExpression) AddExpressionToEnd(added Expression) Expression {
	return InfixCallExpression{e.F, e.A, e.B.AddExpressionToEnd(added)}
}

type NullExpression struct{}

func (e NullExpression) ToObj(_ map[string]core.Obj, _ uint) core.Obj {
	panic("can't convert null expression to object")
}

func (e NullExpression) AddWordToEnd(word string) Expression {
	return WordToExpression(word)
}

func (e NullExpression) AddExpressionToEnd(added Expression) Expression {
	return added
}

type ParenExpression struct {
	Expression
}

func (e ParenExpression) ToObj(dict map[string]core.Obj, biggestNum uint) core.Obj {
	return e.Expression.ToObj(dict, biggestNum)
}

func (e ParenExpression) AddWordToEnd(word string) Expression {
	return DummyAddWordToEnd(e, word)
}

type Declaration interface {
	Apply(map[string]core.Obj)
}

type NormalDeclaration struct {
	Name string
	Expression
}

func (d NormalDeclaration) Apply(dict map[string]core.Obj) {
	dict[d.Name] = d.Expression.ToObj(dict, 0)
}
