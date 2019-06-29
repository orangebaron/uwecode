package compiler

import "io"
import "fmt"
import "../core"

type EOFFunction func(interface{}, []Declaration) []Declaration
type CharacterReader func(byte, interface{}, []Declaration) (interface{}, []Declaration, CharacterReader, EOFFunction)

func IsWhitespace(b byte) bool {
	return contains(" \t\r\n", b)
}

func ErrorEOFFunction(_ interface{}, _ []Declaration) []Declaration {
	panic("Unexpected EOF")
}

type WhitespaceReaderState struct {
	State interface{}
	CharacterReader
	EOFFunction
}

func WhitespaceReader(b byte, state interface{}, decls []Declaration) (interface{}, []Declaration, CharacterReader, EOFFunction) {
	convertedState := state.(WhitespaceReaderState)
	if IsWhitespace(b) {
		return state, decls, WhitespaceReader, convertedState.EOFFunction
	} else {
		return convertedState.CharacterReader(b, convertedState.State, decls)
	}
}

type CommentReaderState struct {
	Escaped bool
	State   interface{}
	CharacterReader
	EOFFunction
}

func CommentReader(b byte, state interface{}, decls []Declaration) (interface{}, []Declaration, CharacterReader, EOFFunction) {
	// TODO: tell when a comment starts
	convertedState := state.(CommentReaderState)
	if b == ']' && !convertedState.Escaped {
		return convertedState.State, decls, convertedState.CharacterReader, convertedState.EOFFunction
	} else if b == '\\' && !convertedState.Escaped {
		convertedState.Escaped = true
		return convertedState, decls, CommentReader, ErrorEOFFunction
	} else {
		convertedState.Escaped = false
		return convertedState, decls, CommentReader, ErrorEOFFunction
	}
}

type NormalReaderState struct {
	Expression
	LastExpression    Expression
	InParentheses     *NormalReaderState // if in parentheses
	CurrentWord       string             // if not in parentheses
	LastWord          string             // only for making new declarations, relevant only on non-paren-enclosed level
	NormalDeclaration                    // relevant only on non-paren-enclosed level
}

func NormalEOFFunction(state interface{}, decls []Declaration) []Declaration {
	convertedState, isAlreadyNormal := state.(NormalReaderState)
	if !isAlreadyNormal {
		convertedWhitespaceState := state.(WhitespaceReaderState)
		convertedState = convertedWhitespaceState.State.(NormalReaderState)
	}
	if convertedState.InParentheses != nil {
		panic("Unexpected EOF")
	}
	if convertedState.CurrentWord != "" {
		convertedState.Expression = convertedState.Expression.AddWordToEnd(convertedState.CurrentWord)
	}
	convertedState.NormalDeclaration.Expression = convertedState.Expression
	return append(decls, convertedState.NormalDeclaration)
}

func NormalCallWithinParens(b byte, inParens NormalReaderState) *NormalReaderState {
	state, _, _, _ := NormalReader(b, inParens, []Declaration{})
	convertedState, isNormal := state.(NormalReaderState)
	if !isNormal {
		convertedState = state.(WhitespaceReaderState).State.(NormalReaderState)
	}
	return &convertedState
}

func NormalReader(b byte, state interface{}, decls []Declaration) (interface{}, []Declaration, CharacterReader, EOFFunction) {
	convertedState := state.(NormalReaderState)
	if IsWhitespace(b) {
		if convertedState.CurrentWord == "=" {
			convertedState.NormalDeclaration.Expression = convertedState.LastExpression
			newState := WhitespaceReaderState{NormalReaderState{NullExpression{}, NullExpression{}, nil, "", "", NormalDeclaration{convertedState.LastWord, NullExpression{}}}, NormalReader, NormalEOFFunction}
			if convertedState.NormalDeclaration.Name != "" {
				decls = append(decls, convertedState.NormalDeclaration)
			}
			return newState, decls, WhitespaceReader, ErrorEOFFunction
		} else {
			if convertedState.InParentheses == nil {
				convertedState.LastExpression = convertedState.Expression
				convertedState.Expression = convertedState.Expression.AddWordToEnd(convertedState.CurrentWord)
				convertedState.LastWord = convertedState.CurrentWord
				convertedState.CurrentWord = ""
			} else {
				convertedState.InParentheses = NormalCallWithinParens(b, *convertedState.InParentheses)
			}
			return WhitespaceReaderState{convertedState, NormalReader, NormalEOFFunction}, decls, WhitespaceReader, NormalEOFFunction
		}
	} else if b == '(' {
		// assuming that theres a space before open paren
		// TODO: don't assume that
		if convertedState.InParentheses == nil {
			newEnclosedNormState := NormalReaderState{NullExpression{}, NullExpression{}, nil, "", "", NormalDeclaration{"", NullExpression{}}}
			convertedState.InParentheses = &newEnclosedNormState
		} else {
			convertedState.InParentheses = NormalCallWithinParens(b, *convertedState.InParentheses)
		}
		return WhitespaceReaderState{convertedState, NormalReader, NormalEOFFunction}, decls, WhitespaceReader, ErrorEOFFunction
	} else if b == ')' {
		// assuming that there's a space before closing paren
		// TODO: don't assume that
		// also assuming that InParentheses != nil, ie something didnt go very badly wrong
		// TODO: don't assume that either
		fmt.Println(convertedState.InParentheses == nil)
		if convertedState.InParentheses.InParentheses == nil {
			convertedState.LastExpression = convertedState.Expression
			convertedState.Expression = convertedState.Expression.AddExpressionToEnd(ParenExpression{convertedState.InParentheses.Expression})
			convertedState.InParentheses = nil
		} else {
			convertedState.InParentheses = NormalCallWithinParens(b, *convertedState.InParentheses)
		}
		return WhitespaceReaderState{convertedState, NormalReader, NormalEOFFunction}, decls, WhitespaceReader, NormalEOFFunction
	} else {
		if convertedState.InParentheses == nil {
			convertedState.CurrentWord = convertedState.CurrentWord + string(b)
		} else {
			convertedState.InParentheses = NormalCallWithinParens(b, *convertedState.InParentheses)
		}
		return convertedState, decls, NormalReader, NormalEOFFunction
	}
}

const readLength = 20 // arbitrary
func ReadCode(reader io.Reader, dict map[string]core.Obj) {
	bytes := make([]byte, readLength)
	var state interface{} = NormalReaderState{NullExpression{}, NullExpression{}, nil, "", "", NormalDeclaration{"", NumExpression{}}}
	charReader := NormalReader
	decls := []Declaration{}
	eofFunc := ErrorEOFFunction
	for {
		n, err := reader.Read(bytes)
		if err != nil {
			break
		}
		for i := 0; i < n; i++ {
			state, decls, charReader, eofFunc = charReader(bytes[i], state, decls)
		}
	}
	decls = eofFunc(state, decls)
	for _, decl := range decls {
		decl.Apply(dict)
	}
}
