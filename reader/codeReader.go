package reader

import "io"
import "fmt"
import "encoding/hex"

type EOFFunction func(interface{}, []Declaration) []Declaration
type CharacterReader func(byte, interface{}, []Declaration) (interface{}, []Declaration, CharacterReader, EOFFunction)

func IsWhitespace(b byte) bool {
	return contains(" \t\r\n", b)
}

func IsDeclWord(word string) bool {
	return word == "=" || word == "_="
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

type EscapedCharLocation int

const (
	Unescaped EscapedCharLocation = iota
	Escaped
	HexFirstChar
	HexSecondChar
)

type StringLiteralReaderState struct {
	NormalReaderState
	CurrentStr   string
	IsSingleChar bool
	EscapedCharLocation
	CurrentHexChar byte
}

func StringLiteralReader(b byte, state interface{}, decls []Declaration) (interface{}, []Declaration, CharacterReader, EOFFunction) {
	convertedState := state.(StringLiteralReaderState)
	if ((b == '"' && !convertedState.IsSingleChar) || (b == '\'' && convertedState.IsSingleChar)) && convertedState.EscapedCharLocation == Unescaped {
		return convertedState.NormalReaderState.AddStringLiteralToEnd(convertedState.CurrentStr, convertedState.IsSingleChar, decls)
	} else {
		switch convertedState.EscapedCharLocation {
		case Unescaped:
			if b == '\\' {
				convertedState.EscapedCharLocation = Escaped
			} else {
				convertedState.CurrentStr += string(b)
			}
		case Escaped:
			convertedState.EscapedCharLocation = Unescaped
			switch b {
			case 'n':
				convertedState.CurrentStr += "\n"
			case 'r':
				convertedState.CurrentStr += "\r"
			case 't':
				convertedState.CurrentStr += "\t"
			case 'x':
				convertedState.EscapedCharLocation = HexFirstChar
			default:
				convertedState.CurrentStr += string(b)
			}
		case HexFirstChar:
			convertedState.EscapedCharLocation = HexSecondChar
			convertedState.CurrentHexChar = b
		case HexSecondChar:
			dst := make([]byte, 1)
			_, err := hex.Decode(dst, []byte{convertedState.CurrentHexChar, b})
			if err != nil {
				panic(err)
			}
			convertedState.CurrentStr += string(dst[0])
			convertedState.EscapedCharLocation = Unescaped
		}
		return convertedState, decls, StringLiteralReader, ErrorEOFFunction
	}
}

type ImportLoc int

const (
	ImportStart ImportLoc = iota
	ImportAfterName
	ImportAfterPrefixed
	ImportAfterPrefixedName
	ImportAfterJust
	ImportAfterJustName
	ImportAfterAs
	ImportAfterAsName
)

type ImportReaderState struct {
	LocInImport ImportLoc
	Escaped     bool
	CurrentWord string
	ImportDeclaration
}

func (s ImportReaderState) EndOfWord() ImportReaderState {
	if s.CurrentWord != "" {
		if s.LocInImport == ImportAfterAsName {
			s.LocInImport = ImportAfterJust
		}
		switch s.LocInImport {
		case ImportStart:
			s.ImportDeclaration.Name = s.CurrentWord
			s.LocInImport = ImportAfterName
		case ImportAfterName:
			switch s.CurrentWord {
			case "public":
				if s.ImportDeclaration.Public {
					panic("said public twice in the same import")
				} else {
					s.ImportDeclaration.Public = true
				}
			case "prefixed":
				s.LocInImport = ImportAfterPrefixed
			case "just":
				s.LocInImport = ImportAfterJust
			default:
				panic("expected \"public\", \"prefixed\", \"just\", or end of import")
			}
		case ImportAfterPrefixed:
			s.ImportDeclaration.Aliases[""] = s.CurrentWord
			s.LocInImport = ImportAfterPrefixedName
		case ImportAfterPrefixedName:
			if s.CurrentWord == "just" {
				s.LocInImport = ImportAfterJust
			} else {
				panic("expected \"just\" or end of import")
			}
		case ImportAfterJust:
			s.ImportDeclaration.ToImport = append(s.ImportDeclaration.ToImport, s.CurrentWord)
			s.LocInImport = ImportAfterJustName
		case ImportAfterJustName:
			switch s.CurrentWord {
			case "as":
				s.LocInImport = ImportAfterAs
			case "and":
				s.LocInImport = ImportAfterJust
			default:
				panic("expected \"as\", \"and\", or end of import")
			}
		case ImportAfterAs:
			s.ImportDeclaration.Aliases[s.ImportDeclaration.ToImport[len(s.ImportDeclaration.ToImport)-1]] = s.CurrentWord
			s.LocInImport = ImportAfterAsName
		default:
			panic("unknown LocInImport")
		}
		s.CurrentWord = ""
	}
	return s
}

func (s ImportReaderState) ValidFinish() bool {
	l := s.LocInImport
	return l == ImportAfterName || l == ImportAfterPrefixedName || l == ImportAfterJustName || l == ImportAfterAsName
}

func ImportReader(b byte, state interface{}, decls []Declaration) (interface{}, []Declaration, CharacterReader, EOFFunction) {
	convertedState := state.(ImportReaderState)
	if IsWhitespace(b) {
		return WhitespaceReaderState{convertedState.EndOfWord(), ImportReader, ErrorEOFFunction}, decls, WhitespaceReader, ErrorEOFFunction
	} else if b == '\\' && !convertedState.Escaped {
		convertedState.Escaped = true
		return convertedState, decls, ImportReader, ErrorEOFFunction
	} else if b == '}' && !convertedState.Escaped {
		convertedState = convertedState.EndOfWord()
		if !convertedState.ValidFinish() {
			panic("unexpected end of import declaration")
		}
		decls = append(decls, convertedState.ImportDeclaration)
		return NormalReaderState{NullExpression{}, NullExpression{}, nil, "", "", NormalDeclaration{"", NullExpression{}, true}}, decls, NormalReader, NormalEOFFunction
	} else {
		convertedState.CurrentWord = convertedState.CurrentWord + string(b)
		return convertedState, decls, ImportReader, ErrorEOFFunction
	}
}

type CommentReaderState struct {
	Escaped bool
	State   interface{}
	CharacterReader
	EOFFunction
}

func CommentReader(b byte, state interface{}, decls []Declaration) (interface{}, []Declaration, CharacterReader, EOFFunction) {
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

func (s NormalReaderState) AddStringLiteralToEnd(str string, isSingleChar bool, decls []Declaration) (NormalReaderState, []Declaration, CharacterReader, EOFFunction) {
	if s.InParentheses != nil {
		newState, _, _, _ := s.InParentheses.AddStringLiteralToEnd(str, isSingleChar, decls)
		s.InParentheses = &newState
	} else if isSingleChar {
		s.LastExpression = s.Expression
		s.Expression = s.Expression.AddExpressionToEnd(CharLiteralExpression{str[0]}) // TODO parse n error n stuff
	} else {
		s.LastExpression = s.Expression
		s.Expression = s.Expression.AddExpressionToEnd(StringLiteralExpression{str}) // TODO parse
	}
	return s, decls, NormalReader, NormalEOFFunction
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
	if convertedState.NormalDeclaration.Name != "" {
		decls = append(decls, convertedState.NormalDeclaration)
	}
	return decls
}

func NormalReader(b byte, state interface{}, decls []Declaration) (interface{}, []Declaration, CharacterReader, EOFFunction) {
	convertedState := state.(NormalReaderState)
	// TODO: check for errors: too many )s, = within paren, delaring something as nothing: "a = 2 b = c = 5"
	if IsWhitespace(b) && IsDeclWord(convertedState.CurrentWord) {
		convertedState.NormalDeclaration.Expression = convertedState.LastExpression
		newState := WhitespaceReaderState{NormalReaderState{NullExpression{}, NullExpression{}, nil, "", "", NormalDeclaration{convertedState.LastWord, NullExpression{}, convertedState.CurrentWord != "_="}}, NormalReader, NormalEOFFunction}
		if convertedState.NormalDeclaration.Name != "" {
			decls = append(decls, convertedState.NormalDeclaration)
		}
		return newState, decls, WhitespaceReader, ErrorEOFFunction
	} else {
		isSpecial := IsWhitespace(b) || contains("()[{\"'", b)
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
				newEnclosedNormState := NormalReaderState{NullExpression{}, NullExpression{}, nil, "", "", NormalDeclaration{"", NullExpression{}, true}}
				convertedState.InParentheses = &newEnclosedNormState
			}
		} else {
			isSecondToInner := convertedState.InParentheses.InParentheses == nil
			stateEnclosed, _, _, _ := NormalReader(b, *convertedState.InParentheses, []Declaration{})
			convertedStateEnclosed, isNormal := stateEnclosed.(NormalReaderState)
			if !isNormal {
				convertedStateEnclosed = stateEnclosed.(WhitespaceReaderState).State.(NormalReaderState)
			}
			convertedState.InParentheses = &convertedStateEnclosed
			if b == ')' && isSecondToInner {
				convertedState.LastExpression = convertedState.Expression
				convertedState.Expression = convertedState.Expression.AddExpressionToEnd(ParenExpression{convertedState.InParentheses.Expression})
				convertedState.InParentheses = nil
			}
		}

		if b == '[' {
			return CommentReaderState{false, convertedState, NormalReader, NormalEOFFunction}, decls, CommentReader, ErrorEOFFunction
		} else if b == '{' {
			if convertedState.NormalDeclaration.Name != "" {
				convertedState.NormalDeclaration.Expression = convertedState.Expression
				decls = append(decls, convertedState.NormalDeclaration)
			}
			return ImportReaderState{ImportStart, false, "", ImportDeclaration{false, "", []string{}, make(map[string]string)}}, decls, ImportReader, ErrorEOFFunction
		} else if contains("\"'", b) {
			return StringLiteralReaderState{convertedState, "", b == '\'', Unescaped, byte(0)}, decls, StringLiteralReader, ErrorEOFFunction
		} else if isSpecial {
			return WhitespaceReaderState{convertedState, NormalReader, NormalEOFFunction}, decls, WhitespaceReader, NormalEOFFunction
		} else {
			return convertedState, decls, NormalReader, NormalEOFFunction
		}
	}
}

type SyntaxError struct {
	LineNum int
	CharNum int
	Err     string
}

func (e SyntaxError) Error() string {
	return fmt.Sprintf("Syntax error: line %d character %d: %s", e.LineNum, e.CharNum, e.Err)
}

type EvalError struct {
	Declaration
	Err string
}

func (e EvalError) Error() string {
	return fmt.Sprintf("Evaluation error: %s: %s", e.Declaration.Summary(), e.Err)
}

const readLength = 20 // arbitrary
func ReadCode(reader io.Reader, dict DeclaredDict) (err error) {
	bytes := make([]byte, readLength)
	var state interface{} = NormalReaderState{NullExpression{}, NullExpression{}, nil, "", "", NormalDeclaration{"", NullExpression{}, true}}
	charReader := NormalReader
	decls := []Declaration{}
	eofFunc := ErrorEOFFunction
	charNumber := 1
	lineNumber := 1
	defer func() {
		if r := recover(); r != nil {
			s, isString := r.(string)
			if !isString {
				s = r.(error).Error()
			}
			if charNumber == -1 {
				err = EvalError{decls[0], s}
			} else {
				err = SyntaxError{lineNumber, charNumber, s}
			}
		}
	}()
	for {
		n, err := reader.Read(bytes)
		if err != nil {
			break
		}
		for i := 0; i < n; i++ {
			state, decls, charReader, eofFunc = charReader(bytes[i], state, decls)
			charNumber++
			if bytes[i] == '\n' {
				lineNumber++
				charNumber = 1
			}
		}
	}
	decls = eofFunc(state, decls)
	charNumber = -1
	for len(decls) > 0 {
		decls[0].Apply(dict)
		decls = decls[1:]
	}
	return nil
}
