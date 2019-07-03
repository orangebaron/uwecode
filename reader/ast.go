package compiler

import "strconv"
import "os"
import "../core"

type DeclaredDict struct {
	Public               map[string]core.Obj
	Private              map[string]core.Obj
	WithinDecl           map[string]uint
	BiggestWithinDeclNum *uint
}

func (d DeclaredDict) AddObj(name string, obj core.Obj, public bool) {
	_, alreadyPublic := d.Public[name]
	_, alreadyPrivate := d.Private[name]
	if alreadyPublic || alreadyPrivate {
		panic("Declaration already exists")
	}
	addTo := d.Private
	if public {
		addTo = d.Public
	}
	addTo[name] = obj
}
func (d DeclaredDict) AddWithinDecl(name string) uint {
	*d.BiggestWithinDeclNum++
	old := d.WithinDecl[name]
	d.WithinDecl[name] = *d.BiggestWithinDeclNum
	return old
}
func (d DeclaredDict) RemoveWithinDecl(name string, old uint) {
	*d.BiggestWithinDeclNum--
	d.WithinDecl[name] = old
}
func (d DeclaredDict) ClearWithinDeclInfo() {
	if *d.BiggestWithinDeclNum != 0 {
		panic("BiggestWithinDeclNum ought to be 0")
	}
	for k, v := range d.WithinDecl {
		if v != 0 {
			panic("v ought to be 0")
		}
		delete(d.WithinDecl, k)
	}
}
func (d DeclaredDict) GetObj(name string) core.Obj {
	if d.Public[name] != nil {
		return d.Public[name]
	} else if d.Private[name] != nil {
		return d.Private[name]
	} else if d.WithinDecl[name] != 0 {
		return core.ReturnVal{d.WithinDecl[name]}
	} else {
		panic(name + " does not exist")
	}
}

type Expression interface {
	ToObj(DeclaredDict) core.Obj
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

func (e CallExpression) ToObj(dict DeclaredDict) core.Obj {
	return core.Called{e.F.ToObj(dict), e.X.ToObj(dict)}
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

func (e FunctionExpression) ToObj(dict DeclaredDict) core.Obj {
	old := dict.AddWithinDecl(e.ArgName)
	returnVal := core.Function{*dict.BiggestWithinDeclNum, e.Returned.ToObj(dict)}
	dict.RemoveWithinDecl(e.ArgName, old)
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

func (e NumExpression) ToObj(dict DeclaredDict) core.Obj {
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

func (e VarExpression) ToObj(dict DeclaredDict) core.Obj {
	return dict.GetObj(e.Name)
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

func (e InfixCallExpression) ToObj(dict DeclaredDict) core.Obj {
	return core.Called{core.Called{e.F.ToObj(dict), e.A.ToObj(dict)}, e.B.ToObj(dict)}
}

func (e InfixCallExpression) AddWordToEnd(word string) Expression {
	return InfixCallExpression{e.F, e.A, e.B.AddWordToEnd(word)}
}

func (e InfixCallExpression) AddExpressionToEnd(added Expression) Expression {
	return InfixCallExpression{e.F, e.A, e.B.AddExpressionToEnd(added)}
}

type NullExpression struct{}

func (e NullExpression) ToObj(_ DeclaredDict) core.Obj {
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

func (e ParenExpression) ToObj(dict DeclaredDict) core.Obj {
	return e.Expression.ToObj(dict)
}

func (e ParenExpression) AddWordToEnd(word string) Expression {
	return DummyAddWordToEnd(e, word)
}

type Declaration interface {
	Apply(DeclaredDict)
}

type NormalDeclaration struct {
	Name string
	Expression
	Public bool
}

func (d NormalDeclaration) Apply(dict DeclaredDict) {
	dict.AddObj(d.Name, d.Expression.ToObj(dict), d.Public)
	dict.ClearWithinDeclInfo()
}

type ImportDeclaration struct {
	Public   bool
	Name     string
	ToImport []string
	Aliases  map[string]string
}

func (d ImportDeclaration) Apply(dict DeclaredDict) {
	biggest := uint(0)
	newDict := DeclaredDict{make(map[string]core.Obj), make(map[string]core.Obj), make(map[string]uint), &biggest}
	f, _ := os.Open(d.Name)
	ReadCode(f, newDict)
	toImport := d.ToImport
	if len(toImport) == 0 {
		for k := range newDict.Public {
			toImport = append(toImport, k)
		}
	}
	for _, imp := range toImport {
		if newDict.Public[imp] == nil {
			panic("Tried to import " + imp + " from " + d.Name + ", which doesn't exist")
		}
		name := d.Aliases[imp]
		if name == "" {
			name = imp
		}
		name = d.Aliases[""] + name
		dict.AddObj(name, newDict.Public[imp], d.Public)
	}
}
