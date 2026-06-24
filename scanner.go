package tast

import (
	"errors"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
)

type scanner struct {
	source []byte
	tokens []token
	errors []scanError

	// Internal reading state
	current int
	line    int
	column  int
	start   int
}

type tokenType uint32

const (
	_ tokenType = iota
	comment
	leftBracket
	rightBracket
	comma
	dot
	minus
	plus
	slash
	star
	equal
	newLine

	basicString
	multilineBasicString

	literalString
	multilineLiteralString

	floatPoint
	integer

	bareKey
	// Reserved keywords
	boolean
	infinity
	nan

	eof
)

type token struct {
	Type    tokenType
	Lexeme  string
	Literal any
	Line    int
	Column  int
}

func newScanner(src []byte) *scanner {
	return &scanner{
		source: src,
		tokens: []token{},
		errors: []scanError{},
	}
}

func (s *scanner) scan() ([]token, error) {
	for !s.isAtEnd() {
		s.start = s.current
		s.scanNext()
	}
	s.tokens = append(s.tokens, token{Type: eof, Lexeme: "", Line: s.line})

	if len(s.errors) > 0 {
		errs := make([]error, len(s.errors))
		for i, e := range s.errors {
			errs[i] = &e
		}
		return s.tokens, errors.Join(errs...)
	}

	return s.tokens, nil
}

func (s *scanner) scanNext() {
	currentChar := s.advance()
	switch currentChar {
	case '#':
		s.comment()
	case '"':
		if s.isMultlineStart() {
			s.multilineBasicString()
		} else {
			s.basicString()
		}
	case '=':
		s.addToken(equal, "=")
	case '\n':
		s.addToken(newLine, "\n")
		s.column = 0
		s.line++
	case '.':
		s.addToken(dot, ".")
	case '[':
		s.addToken(leftBracket, "[")
	case ']':
		s.addToken(rightBracket, "]")
	case 'i':
		if s.matchSequence("nf") {
			s.addToken(infinity, math.Inf(1))
		} else {
			s.key()
		}
	case 'n':
		if s.matchSequence("an") {
			s.addToken(nan, math.NaN())
		} else {
			s.key()
		}
	case 't':
		if s.matchSequence("rue") {
			s.addToken(boolean, true)
			return
		} else {
			s.key()
		}
	case 'f':
		if s.matchSequence("alse") {
			s.addToken(boolean, false)
			return
		} else {
			s.key()
		}
	case '+':
		s.addToken(plus, "+")
	case '-':
		s.addToken(minus, "-")
	case '\t', ' ', '\r':
		break
	default:
		if isDigit(currentChar) {
			s.number()
			return
		}

		if isKey(currentChar) {
			s.key()
			return
		}

		s.addError(fmt.Sprintf("unexpected character %q", currentChar))
	}
}

func (s *scanner) matchSequence(expected string) bool {
	for i, c := range expected {
		if s.current+i >= len(s.source) {
			return false
		}
		if rune(s.source[s.current+i]) != c {
			return false
		}
	}
	s.current += len(expected)
	return true
}

func (s *scanner) advance() byte {
	curr := s.source[s.current]
	s.column++
	s.current++
	return curr
}

// Looks at the value of the source at the current index
// without consuming it
//
// Alias for peekAt0
func (s *scanner) peek() byte {
	return s.peekAt(0)
}

// Looks at the value of the source at the current index + 1
// without consuming it
//
// Alias for peekAt 1
func (s *scanner) peekNext() byte {
	return s.peekAt(1)
}

// Looks at the value of the source at the current index + an
// arbitrary offset value without consuming it
func (s *scanner) peekAt(offset int) byte {
	if s.current+offset >= len(s.source) {
		return 0
	}

	return s.source[s.current+offset]
}

func (s *scanner) addToken(tokenType tokenType, literal any) {
	lexeme := string(s.source[s.start:s.current])
	t := token{
		Line:    s.line,
		Column:  s.column,
		Lexeme:  lexeme,
		Literal: literal,
		Type:    tokenType,
	}
	s.tokens = append(s.tokens, t)
}

// Checks to see if the current pointer is off bounds from the
// length of the source byte array
func (s *scanner) isAtEnd() bool {
	return s.current >= len(s.source)
}

func (s *scanner) comment() {
	for s.peek() != '\n' && !s.isAtEnd() {
		s.advance()
	}
	commentValue := s.source[s.start:s.current]

	// make up for finding a newline character
	s.line++
	s.addToken(comment, commentValue)
}

func (s *scanner) number() {
	for !s.isAtEnd() && (isDigit(s.peek()) || s.isValidUnderscore()) {
		s.advance()
	}

	var isFloatingPoint bool
	if s.peek() == '.' && isDigit(s.peekNext()) {

		isFloatingPoint = true
		s.advance()

		for isDigit(s.peek()) {
			s.advance()
		}
	}

	lexeme := s.source[s.start:s.current]

	// Cleanup any underscores
	cleaned := strings.ReplaceAll(string(lexeme), "_", "")

	if isFloatingPoint {
		floatVal, _ := strconv.ParseFloat(cleaned, 64)
		s.addToken(floatPoint, floatVal)
	} else {
		intVal, err := strconv.ParseInt(cleaned, 10, 64)
		if err != nil {
			log.Printf("err: %v", err)
		}
		s.addToken(integer, intVal)
	}
}

func (s *scanner) key() {
	for !s.isAtEnd() && isKey(s.peek()) {
		s.advance()
	}

	lexeme := s.source[s.start:s.current]
	s.addToken(bareKey, string(lexeme))
}

func (s *scanner) multilineBasicString() {
	for !s.isAtEnd() && !s.isMultilineClosing() {
		if s.peek() == '\n' {
			s.line++
			s.column = 0
		}
		s.advance()
	}

	// Unterminated multilne string
	if s.isAtEnd() {
		s.addError("Unterminated multiline basic string")
		return
	}

	s.advance() // Trim first '"'
	s.advance() // Trim second '"'
	s.advance() // Trim third '"'

	strValue := s.source[s.start+3 : s.current-3]

	// trim initial newline value as per the TOML spec
	if len(strValue) > 0 {
		if strValue[0] == '\n' {
			strValue = strValue[1:]
		} else if strValue[0] == '\r' && len(strValue) > 1 && strValue[1] == '\n' {
			strValue = strValue[2:]
		}
	}

	s.addToken(multilineBasicString, string(strValue))
}

func (s *scanner) addError(msg string) {
	s.errors = append(s.errors, scanError{
		Line:    s.line,
		Column:  s.column,
		Offset:  s.current,
		Message: msg,
	})
}

func (s *scanner) basicString() {
	for !s.isAtEnd() && s.peek() != '"' {
		if s.peek() == '\n' {
			s.line++
		}
		s.advance()
	}

	if s.isAtEnd() {
		s.addError("Unterminated basic string")
		return
	}

	s.advance()

	// lexeme = s.Source[s.start:s.current] → "hello" (with quotes, handled by addTokenValue)
	// literal = just the content between the quotes
	strValue := s.source[s.start+1 : s.current-1]

	s.addToken(basicString, string(strValue))
}

func (s *scanner) isMultlineStart() bool {
	if s.isAtEnd() {
		return false
	}

	return s.peek() == '"' && s.peekNext() == '"'
}

func (s *scanner) isMultilineClosing() bool {
	if s.isAtEnd() {
		return false
	}

	return s.peek() == '"' && s.peekNext() == '"' && s.peekAt(2) == '"'
}

// Valid underscore means that it should be proceded by another digit value
// otherwise is not valid
func (s *scanner) isValidUnderscore() bool {
	return s.peek() == '_' && isDigit(s.peekNext())
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func isAlphanumeric(b byte) bool {
	return isDigit(b) || (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

func isKey(b byte) bool {
	return isAlphanumeric(b) || b == '_' || b == '-'
}
