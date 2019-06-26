package compiler

type CharacterReader func(byte, interface{}, []Declaration) (interface{}, []Declaration, CharacterReader)

func IsWhitespace(b byte) bool {
	for _, c := range " \t\r\n" {
		if c == b {
			return true
		}
	}
	return false
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
