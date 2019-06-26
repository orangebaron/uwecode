package compiler

import "../core"

type Expression interface {
	ToObj(map[string]core.Obj, uint) core.Obj
}

type CallExpression struct {
	F Expression
	X Expression
}

func (e CallExpression) ToObj(dict map[string]core.Obj, biggestNum uint) core.Obj {
	return core.Called{e.F.ToObj(dict, biggestNum), e.X.ToObj(dict, biggestNum)}
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

type NumExpression struct {
	Num uint
}

func (e NumExpression) ToObj(_ map[string]core.Obj, _ uint) core.Obj {
	return core.ChurchNum{e.Num}
}

type VarExpression struct {
	Name string
}

func (e VarExpression) ToObj(dict map[string]core.Obj, _ uint) core.Obj {
	return dict[e.Name]
}

type Declaration interface {
	Apply(map[string]core.Obj)
}

type NormalDeclaration struct {
	Name string
	Expr Expression
}

func (d NormalDeclaration) Apply(dict map[string]core.Obj) {
	dict[d.Name] = d.Expr.ToObj(dict, 0)
}
