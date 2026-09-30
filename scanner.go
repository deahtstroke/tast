package tast

import (
	"errors"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"
)

const (
	NanoSecondPrecision = 9
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

var timeTerminators map[byte]struct{} = map[byte]struct{}{
	'\n': {},
	'\t': {},
	' ':  {},
	'#':  {},
	']':  {},
	'}':  {},
}

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
		s.nextToken()
	}

	s.eof()

	if len(s.errors) > 0 {
		errs := make([]error, len(s.errors))
		for i, e := range s.errors {
			errs[i] = &e
		}
		return s.tokens, errors.Join(errs...)
	}

	return s.tokens, nil
}

func (s *scanner) eof() {
	s.tokens = append(s.tokens, token{Type: eof, Lexeme: "", Line: s.line})
}

func (s *scanner) nextToken() {
	curr := s.next()
	switch curr {
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
	case '{':
		s.addToken(leftBrace, "{")
	case '}':
		s.addToken(rightBracket, "}")
	case 'i':
		if s.matchSequence("nf") {
			s.addToken(infinity, math.Inf(1))
		} else {
			s.scanKeys()
		}
	case 'n':
		if s.matchSequence("an") {
			s.addToken(nan, math.NaN())
		} else {
			s.scanKeys()
		}
	case 't':
		if s.matchSequence("rue") {
			s.addToken(boolean, true)
		} else {
			s.scanKeys()
		}
	case 'f':
		if s.matchSequence("alse") {
			s.addToken(boolean, false)
		} else {
			s.scanKeys()
		}
	case '+':
		s.addToken(plus, "+")
	case '-':
		s.addToken(minus, "-")
	case '\t', ' ', '\r': // ignore tabs, spaces and carriage returns
		break
	default:
		if isDigit(curr) {
			s.scanNumerals()
			return
		}

		if isKey(curr) {
			s.scanKeys()
			return
		}

		s.addError("unexpected character " + string(curr))
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

// Advance to the next token
func (s *scanner) next() byte {
	curr := s.source[s.current]
	s.column++
	s.current++
	return curr
}

// Advance N times to the next token
func (s *scanner) nextN(n int) {
	if n > 0 {
		for range n {
			s.next()
		}
	}
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

// Adds a token to the scanner's list of consumed tokens
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
		s.next()
	}
	commentValue := s.source[s.start:s.current]

	// make up for finding a newline character
	s.line++
	s.addToken(comment, commentValue)
}

// scanNumerals() encapsulates numerical-specific behavior when scanning tokens
// in a TOML file such as branching out to Local Time or Local Date respectively
// depending on the layout of the tokens scanned or returning an integer or float
// token
func (s *scanner) scanNumerals() {
	var hasUnderscores bool

	for !s.isAtEnd() {
		isUnderscore := s.isValidUnderscore()

		if isUnderscore {
			hasUnderscores = true
		}

		if !hasUnderscores {
			if s.current-s.start == 2 && s.peek() == ':' {
				s.scanTime()
				return
			} else if s.current-s.start == 4 && s.peek() == '-' {
				s.scanDate()
				return
			}
		}

		if !isDigit(s.peek()) && !isUnderscore {
			break
		}

		s.next()
	}

	var isFloatingPoint bool
	if s.peek() == '.' && isDigit(s.peekNext()) {

		isFloatingPoint = true
		s.next()

		for isDigit(s.peek()) {
			s.next()
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

// scanTime() tries to parse the time portion of RFC 3339 as specified in the
// TOML v1.1 spec. It makes no assumptions regarding timezone or offset
func (s *scanner) scanTime() {
	s.next()
	if s.isAtEnd() || !(isDigit(s.peek()) && isDigit(s.peekNext())) {
		msg := "Unable to parse local time token: Minutes are malformed"
		s.addError(msg)
		return
	}

	s.nextN(2)

	// Unknown seconds are assumed to be :00, therefore we just save the token
	_, isTerminator := timeTerminators[s.peek()]
	if s.isAtEnd() || isTerminator {
		union := string(s.source[s.start:s.current]) + ":00"
		t, err := time.Parse(time.TimeOnly, union)
		if err != nil {
			s.addError(fmt.Sprintf("Unable to parse time: %v", err))
			return
		}

		s.addToken(localTime, t)
		return
	} else if s.peek() != ':' {
		msg := "Unable to parse local time token: Minutes are malformed"
		s.addError(msg)
		return
	}

	// ':' for seconds
	s.next()

	if s.isAtEnd() || !(isDigit(s.peek()) && isDigit(s.peekNext())) {
		msg := "Unable to parse local time: Seconds are malformed"
		s.addError(msg)
		return
	}

	s.nextN(2)

	// Local time that stops at seconds, no millisecond precision
	_, isTerminator = timeTerminators[s.peek()]
	illegalTermination := !isTerminator && s.peek() != '.'
	switch {
	case isTerminator, s.isAtEnd():
		t, err := time.Parse("15:04:05", string(s.source[s.start:s.current]))
		if err != nil {
			s.addError(fmt.Sprintf("Unable to parse time: %v", err))
			return
		}

		s.addToken(localTime, t)
		return
	case illegalTermination:
		msg := "Unable to parse local time: Seconds are malformed"
		s.addError(msg)
		return
	default:
	}

	// '.' for milliseconds
	s.next()

	c := 0
	for !s.isAtEnd() || c > NanoSecondPrecision {
		s.next()
	}

	t, err := time.Parse("15:04:05.999999", string(s.source[s.start:s.current]))
	if err != nil {
		s.addError(fmt.Sprintf("Unable to parse time: %v", err))
		return
	}

	s.addToken(localTime, t)
}

func (s *scanner) scanDate() {
	// consume initial hyphen
	s.next()

	if s.isAtEnd() || !(isDigit(s.peek()) && isDigit(s.peekNext())) {
		msg := "Unable to parse local date token: Month is malformed"
		s.addError(msg)
		return
	}

	s.nextN(2)

	if s.peek() != '-' {
		msg := "Unable to parse date token: expected hyphen, got %s"
		s.addError(fmt.Sprintf(msg, string(s.peek())))
		return
	}

	s.next()

	if s.isAtEnd() || !(isDigit(s.peek()) && isDigit(s.peekNext())) {
		msg := "Unable to parse local date token: Day is malformed"
		s.addError(msg)
		return
	}

	s.nextN(2)

	t, err := time.Parse(time.DateOnly, string(s.source[s.start:s.current]))
	if err != nil {
		s.addError(fmt.Sprintf("Unable to parse time: %v", err))
		return
	}

	s.addToken(localDate, t)
}

func (s *scanner) scanKeys() {
	for !s.isAtEnd() && isKey(s.peek()) {
		s.next()
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
		s.next()
	}

	// Unterminated multilne string
	if s.isAtEnd() {
		s.addError("Unterminated multiline basic string")
		return
	}

	s.next() // Trim first '"'
	s.next() // Trim second '"'
	s.next() // Trim third '"'

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
		s.next()
	}

	if s.isAtEnd() {
		s.addError("Unterminated basic string")
		return
	}

	s.next()

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

// isDigit checks if the passed in byte is a digit, e.g., in between '0' and '9'
// or 48 <= b <= 57
func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func isAlphanumeric(b byte) bool {
	return isDigit(b) || (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

func isKey(b byte) bool {
	return isAlphanumeric(b) || b == '_' || b == '-'
}
