package tast

import (
	"errors"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
)

type Scanner struct {
	source []byte
	tokens []token
	errors []ScanError

	// Internal reading state
	current int
	line    int
	column  int
	start   int
}

type TokenType uint32

const (
	_ TokenType = iota
	COMMENT
	LEFT_BRACKET
	RIGHT_BRACKET
	COMMA
	DOT
	MINUS
	PLUS
	SLASH
	STAR
	EQUAL
	NEW_LINE

	BASIC_STRING
	MULTILINE_BASIC_STRING

	LITERAL_STRING
	MULTILINE_LITERAL_STRING

	FLOAT
	INTEGER

	BARE_KEY

	// Reserved keywords
	FALSE
	TRUE
	INF
	NAN

	EOF
)

type token struct {
	Type    TokenType
	Lexeme  string
	Literal any
	Line    int
	Column  int
}

func NewScanner(src []byte) *Scanner {
	return &Scanner{
		source: src,
		tokens: []token{},
		errors: []ScanError{},
	}
}

func (s *Scanner) Scan() ([]token, error) {
	for !s.isAtEnd() {
		s.start = s.current
		s.scanNext()
	}
	s.tokens = append(s.tokens, token{Type: EOF, Lexeme: "", Line: s.line})

	if len(s.errors) > 0 {
		errs := make([]error, len(s.errors))
		for i, e := range s.errors {
			errs[i] = &e
		}
		return s.tokens, errors.Join(errs...)
	}

	return s.tokens, nil
}

func (s *Scanner) scanNext() {
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
		s.addToken("=", "=", EQUAL)
	case '\n':
		s.addToken("\n", "\n", NEW_LINE)
		s.column = 0
		s.line++
	case '.':
		s.addToken(".", ".", DOT)
	case '[':
		s.addToken("[", "[", LEFT_BRACKET)
	case ']':
		s.addToken("]", "]", RIGHT_BRACKET)
	case 'i':
		if s.matchSequence("nf") {
			s.addToken("inf", math.Inf(1), INF)
		}
		s.key()
	case 'n':
		if s.matchSequence("an") {
			s.addToken("nan", math.NaN(), NAN)
		}
		s.key()
	case 't':
		if s.matchSequence("rue") {
			s.addToken("true", true, TRUE)
			return
		}
		s.key()
	case 'f':
		if s.matchSequence("alse") {
			s.addToken("false", false, FALSE)
			return
		}
		s.key()
	case '+':
		s.addToken("+", "+", PLUS)
	case '-':
		s.addToken("-", "-", MINUS)
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

		s.AddError(fmt.Sprintf("unexpected character %q", currentChar))
	}
}

func (s *Scanner) matchSequence(expected string) bool {
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

func (s *Scanner) advance() byte {
	curr := s.source[s.current]
	s.column++
	s.current++
	return curr
}

// Looks at the value of the source at the current index
// without consuming it
//
// Alias for peekAt0
func (s *Scanner) peek() byte {
	return s.peekAt(0)
}

// Looks at the value of the source at the current index + 1
// without consuming it
//
// Alias for peekAt 1
func (s *Scanner) peekNext() byte {
	return s.peekAt(1)
}

// Looks at the value of the source at the current index + an
// arbitrary offset value without consuming it
func (s *Scanner) peekAt(offset int) byte {
	if s.current+offset >= len(s.source) {
		return 0
	}

	return s.source[s.current+offset]
}

func (s *Scanner) addToken(lexeme string, literal any, tokenType TokenType) {
	t := token{
		Line:    s.line,
		Column:  s.column,
		Lexeme:  lexeme,
		Literal: literal,
		Type:    tokenType,
	}
	s.tokens = append(s.tokens, t)
}

func (s *Scanner) addTokenValue(t TokenType, literal any) {
	lexeme := string(s.source[s.start:s.current])
	s.addToken(lexeme, literal, t)
}

// Checks to see if the current pointer is off bounds from the
// length of the source byte array
func (s *Scanner) isAtEnd() bool {
	return s.current >= len(s.source)
}

func (s *Scanner) comment() {
	for s.peek() != '\n' && !s.isAtEnd() {
		s.advance()
	}
	commentValue := s.source[s.start:s.current]

	// make up for finding a newline character
	s.line++
	s.addTokenValue(COMMENT, commentValue)
}

func (s *Scanner) number() {
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
		s.addTokenValue(FLOAT, floatVal)
	} else {
		intVal, err := strconv.ParseInt(cleaned, 10, 64)
		if err != nil {
			log.Printf("err: %v", err)
		}
		s.addTokenValue(INTEGER, intVal)
	}
}

func (s *Scanner) key() {
	for !s.isAtEnd() && isKey(s.peek()) {
		s.advance()
	}

	lexeme := s.source[s.start:s.current]
	s.addTokenValue(BARE_KEY, string(lexeme))
}

func (s *Scanner) multilineBasicString() {
	for !s.isAtEnd() && !s.isMultilineClosing() {
		if s.peek() == '\n' {
			s.line++
			s.column = 0
		}
		s.advance()
	}

	// Unterminated multilne string
	if s.isAtEnd() {
		s.AddError("Unterminated multiline basic string")
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

	s.addTokenValue(MULTILINE_BASIC_STRING, string(strValue))
}

func (s *Scanner) AddError(msg string) {
	s.errors = append(s.errors, ScanError{
		Line:    s.line,
		Column:  s.column,
		Offset:  s.current,
		Message: msg,
	})
}

func (s *Scanner) basicString() {
	for !s.isAtEnd() && s.peek() != '"' {
		if s.peek() == '\n' {
			s.line++
		}
		s.advance()
	}

	if s.isAtEnd() {
		s.AddError("Unterminated basic string")
	}

	s.advance()

	// lexeme = s.Source[s.start:s.current] → "hello" (with quotes, handled by addTokenValue)
	// literal = just the content between the quotes
	strValue := s.source[s.start+1 : s.current-1]

	s.addTokenValue(BASIC_STRING, string(strValue))
}

func (s *Scanner) isMultlineStart() bool {
	if s.isAtEnd() {
		return false
	}

	return s.peek() == '"' && s.peekNext() == '"'
}

func (s *Scanner) isMultilineClosing() bool {
	if s.isAtEnd() {
		return false
	}

	return s.peek() == '"' && s.peekNext() == '"' && s.peekAt(2) == '"'
}

// Valid underscore means that it should be proceded by another digit value
// otherwise is not valid
func (s *Scanner) isValidUnderscore() bool {
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
