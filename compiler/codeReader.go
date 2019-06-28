package compiler

import "strings"

type CharacterReader func(byte, interface{}, []Declaration) (interface{}, []Declaration, CharacterReader)

func IsWhitespace(b byte) bool {
	return contains(" \t\r\n", b)
}

type WhitespaceReaderState struct {
	State interface{}
	CharacterReader
}

func WhitespaceReader(b byte, state interface{}, decls []Declaration) (interface{}, []Declaration, CharacterReader) {
	if IsWhitespace(b) {
		return state, decls, WhitespaceReader
	} else {
		convertedState := state.(WhitespaceReaderState)
		return convertedState.CharacterReader(b, convertedState.State, decls)
	}
}

type CommentReaderState struct {
	Escaped bool
	State   interface{}
	CharacterReader
}

func CommentReader(b byte, state interface{}, decls []Declaration) (interface{}, []Declaration, CharacterReader) {
	convertedState := state.(CommentReaderState)
	if b == ']' && !convertedState.Escaped {
		return convertedState.CharacterReader(b, convertedState.State, decls)
	} else {
		convertedState.Escaped = false
		return convertedState, decls, CommentReader
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

func NormalReader(b byte, state interface{}, decls []Declaration) (interface{}, []Declaration, CharacterReader) {
	convertedState := state.(NormalReaderState)
	if IsWhitespace(b) {
		if convertedState.CurrentWord == "=" {
			convertedState.NormalDeclaration.Expression = convertedState.LastExpression
			return WhitespaceReaderState{NormalReaderState{NullExpression{}, NullExpression{}, nil, "", "", NormalDeclaration{convertedState.LastWord, NullExpression{}}},NormalReader}, append(decls, convertedState.NormalDeclaration), WhitespaceReader
		} else {
			if convertedState.InParentheses == nil {
				convertedState.LastExpression = convertedState.Expression
				convertedState.Expression = convertedState.Expression.AddWordToEnd(convertedState.CurrentWord)
				convertedState.LastWord = convertedState.CurrentWord
				convertedState.CurrentWord = ""
			} else {
				newEnclosedWhiteState, _, _ := NormalReader(b, *convertedState.InParentheses, decls)
newEnclosedNormState := newEnclosedWhiteState.(WhitespaceReaderState).State.(NormalReaderState)
				convertedState.InParentheses = &newEnclosedNormState
			}
			return WhitespaceReaderState{convertedState, NormalReader}, decls, WhitespaceReader
		}
	} else if b == '(' {
		// assuming that theres a space before open paren
		// TODO: don't assume that
		newEnclosedNormState := NormalReaderState{NullExpression{}, NullExpression{}, nil, "", "", NormalDeclaration{"", NullExpression{}}}
		convertedState.InParentheses = &newEnclosedNormState
		return WhitespaceReaderState{convertedState, NormalReader}, decls, WhitespaceReader
	} else if b == ')' {
		// assuming that there's a space before closing paren
		// TODO: don't assume that
		// also assuming that InParentheses != nil, ie something didnt go very badly wrong
		// TODO: don't assume that either
		if convertedState.InParentheses.InParentheses == nil {
			convertedState.LastExpression = convertedState.Expression
			convertedState.Expression = convertedState.Expression.AddExpressionToEnd(ParenExpression{convertedState.InParentheses.Expression})
		} else {
			newEnclosedWhiteState, _, _ := NormalReader(b, *convertedState.InParentheses, decls)
newEnclosedNormState := newEnclosedWhiteState.(WhitespaceReaderState).State.(NormalReaderState)
			convertedState.InParentheses = &newEnclosedNormState
		}
		return WhitespaceReaderState{convertedState, NormalReader}, decls, WhitespaceReader
	} else {
		if convertedState.InParentheses == nil {
			convertedState.CurrentWord = strings.Append(convertedState.CurrentWord, b)
		} else {
			newEnclosedWhiteState, _, _ := NormalReader(b, *convertedState.InParentheses, decls)
newEnclosedNormState := newEnclosedWhiteState.(WhitespaceReaderState).State.(NormalReaderState)
			convertedState.InParentheses = &newEnclosedNormState
		}
		return convertedState, decls, NormalReader
	}
}
