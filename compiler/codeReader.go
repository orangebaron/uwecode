package compiler

import "io"
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

func NormalReader(b byte, state interface{}, decls []Declaration) (interface{}, []Declaration, CharacterReader, EOFFunction) {
	convertedState := state.(NormalReaderState)
	// TODO: check for errors: too many )s, = within paren, delaring something as nothing: "a = 2 b = c = 5"
	if IsWhitespace(b) && convertedState.CurrentWord == "=" {
		convertedState.NormalDeclaration.Expression = convertedState.LastExpression
		newState := WhitespaceReaderState{NormalReaderState{NullExpression{}, NullExpression{}, nil, "", "", NormalDeclaration{convertedState.LastWord, NullExpression{}}}, NormalReader, NormalEOFFunction}
		if convertedState.NormalDeclaration.Name != "" {
			decls = append(decls, convertedState.NormalDeclaration)
		}
		return newState, decls, WhitespaceReader, ErrorEOFFunction
	} else {
		isSpecial := IsWhitespace(b) || contains("()[", b)
		if convertedState.InParentheses == nil {
			if isSpecial && convertedState.CurrentWord != "" {
				convertedState.LastExpression = convertedState.Expression
				convertedState.Expression = convertedState.Expression.AddWordToEnd(convertedState.CurrentWord)
				convertedState.LastWord = convertedState.CurrentWord
				convertedState.CurrentWord = ""
			}
			if !isSpecial {
				convertedState.CurrentWord = convertedState.CurrentWord + string(b)
			}
			if b == '(' {
				newEnclosedNormState := NormalReaderState{NullExpression{}, NullExpression{}, nil, "", "", NormalDeclaration{"", NullExpression{}}}
				convertedState.InParentheses = &newEnclosedNormState
			}
		} else {
			stateEnclosed, _, _, _ := NormalReader(b, *convertedState.InParentheses, []Declaration{})
			convertedStateEnclosed, isNormal := stateEnclosed.(NormalReaderState)
			if !isNormal {
				convertedStateEnclosed = stateEnclosed.(WhitespaceReaderState).State.(NormalReaderState)
			}
			convertedState.InParentheses = &convertedStateEnclosed
			if b == ')' && convertedState.InParentheses.InParentheses == nil {
				convertedState.LastExpression = convertedState.Expression
				convertedState.Expression = convertedState.Expression.AddExpressionToEnd(ParenExpression{convertedState.InParentheses.Expression})
				convertedState.InParentheses = nil
			}
		}

		if b == '[' {
			return CommentReaderState{false, convertedState, NormalReader, NormalEOFFunction}, decls, CommentReader, ErrorEOFFunction
		} else if isSpecial {
			return WhitespaceReaderState{convertedState, NormalReader, NormalEOFFunction}, decls, WhitespaceReader, NormalEOFFunction
		} else {
			return convertedState, decls, NormalReader, NormalEOFFunction
		}
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
