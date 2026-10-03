package tast

import (
	"errors"
	"fmt"
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
		s.emitToken(equal, "=")
	case '\n':
		s.emitToken(newLine, "\n")
		s.column = 0
		s.line++
	case '.':
		s.emitToken(dot, ".")
	case '[':
		s.emitToken(leftBracket, "[")
	case ']':
		s.emitToken(rightBracket, "]")
	case '{':
		s.emitToken(leftBrace, "{")
	case '}':
		s.emitToken(rightBracket, "}")
	case 'i':
		if s.matchSequence("nf") {
			s.emitToken(infinity, math.Inf(1))
		} else {
			s.scanKeys()
		}
	case 'n':
		if s.matchSequence("an") {
			s.emitToken(nan, math.NaN())
		} else {
			s.scanKeys()
		}
	case 't':
		if s.matchSequence("rue") {
			s.emitToken(boolean, true)
		} else {
			s.scanKeys()
		}
	case 'f':
		if s.matchSequence("alse") {
			s.emitToken(boolean, false)
		} else {
			s.scanKeys()
		}
	case '+':
		s.emitToken(plus, "+")
	case '-':
		s.emitToken(minus, "-")
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
func (s *scanner) emitToken(tokenType tokenType, literal any) {
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
	s.emitToken(comment, commentValue)
}

// scanDigits loops through the characters and will only return an error
// if we find an invalid underscore that's not between two digits.
// Finding something other than a digit will still return with no issues
// since errors are supposed to be handled downstream for unexpected
// characters
func (s *scanner) scanDigits() bool {
	for !s.isAtEnd() {
		switch c := s.peek(); {
		case isDigit(c):
			s.next()
		case c == '_' && isDigit(s.peekNext()):
			s.nextN(2)
		case c == '_':
			s.next()
			s.addError("Numeral: Undescore must between digits")
			return false
		default:
			return true
		}
	}
	return true
}

// scanNumerals encapsulates numerical-specific behavior when scanning tokens
// in a TOML file such as branching out to LocalTime or LocalDate depending on
// the layout of the tokens scanned or returning an integer or float or handling
// appropriately formatted floats and integers
func (s *scanner) scanNumerals() {
	if !s.scanDigits() {
		return
	}

	lexeme := string(s.source[s.start:s.current])
	hasUnderscores := strings.Contains(lexeme, "_")

	if !hasUnderscores {
		switch {
		case s.current-s.start == 2 && s.peek() == ':':
			s.scanClock()
			return
		case s.current-s.start == 4 && s.peek() == '-':
			s.scanDate()
			return
		}
	}

	isFloat := false
	if s.peek() == '.' && isDigit(s.peekNext()) {

		isFloat = true
		s.next() // '.'

		if !s.scanDigits() {
			return
		}

		lexeme = string(s.source[s.start:s.current])
	}

	cleaned := strings.ReplaceAll(string(lexeme), "_", "")

	if isFloat {
		v, err := strconv.ParseFloat(cleaned, 64)
		if err != nil {
			s.addError(fmt.Sprintf("Numericals: invalid float %v", lexeme))
			return
		}
		s.emitToken(floatPoint, v)
		return
	}

	v, err := strconv.ParseInt(cleaned, 10, 64)
	if err != nil {
		s.addError(fmt.Sprintf("Numericals invalid integer %v", lexeme))
		return
	}

	s.emitToken(integer, v)
}

// scanClock() tries to parse the time portion of RFC 3339 as specified in the
// TOML v1.1 spec. It makes no assumptions regarding timezone or offset
func (s *scanner) scanClock() {
	layoutFragment, err := s.consumeClock()
	if err != nil {
		s.addError("Unable to parse time: " + err.Error())
	}

	lexeme := string(s.source[s.start:s.current])
	t, err := time.Parse("15"+layoutFragment, lexeme)
	if err != nil {
		s.addError(fmt.Sprintf("Unable to parse time: %v", err))
	}

	s.emitToken(localTime, t)
}

// TODO: Need to go back to scanning the date portion since scanClock assumes
// already-consumed hours upstream, need to check and consume those hours
// in this method accoridngly
// Need to double check whether the layouts being hardcoded to time._
// constands is the right way since my scanClock method allows for time without
// seconds which RFC3339 does not recognize as valid
// Same thing for time.dateTime: Is of the form 2026-01-02 15:04:05 which is trouble
// if I scan clock without seconds
// TBH! It would honestly just be easier if I just append the ':00' seconds portion
// to the clock whenever they're missing
func (s *scanner) scanDate() {
	tokenType := localDate
	layout := time.DateOnly

	if err := s.consumeDate(); err != nil {
		s.addError(err.Error())
		return
	}

	switch s.peek() {
	case ' ', 'T', 't':
		s.next() // ' ' | 'T' | 't'

		tokenType = localDateTime
		layout = time.DateTime

		if _, err := s.consumeClock(); err != nil {
			s.addError(err.Error())
			return
		}

		switch s.peek() {
		case '-', '+', 'Z':
			c := s.next() // '-' | '+' | 'Z'

			tokenType = offsetDateTime
			layout = time.RFC3339

			if c == 'Z' {
				break
			}

			if err := s.consumeTimeOffset(); err != nil {
				s.addError(err.Error())
				return
			}
		}
	}

	// Branch to localdatetime or offsetdatetime
	t, err := time.Parse(layout, string(s.source[s.start:s.current]))
	if err != nil {
		s.addError(fmt.Sprintf("Unable to parse time: %v", err))
		return
	}

	s.emitToken(tokenType, t)
}

func (s *scanner) consumeTimeOffset() error {
}

func (s *scanner) consumeDate() error {
	s.next() // '-'

	if err := s.consumeTwoDigits(Months); err != nil {
		return err
	}

	if s.peek() != '-' {
		return fmt.Errorf("Unable to parse date token: expected hyphen, got %s", string(s.peek()))
	}

	s.next() // '-'

	if err := s.consumeTwoDigits(Days); err != nil {
		return err
	}

	return nil
}

// consumeClock has the core logic for scanning-only a clock value
// of the form :MM[:ss[.frac]], where 'frac' allows up to nanosecond precision
//
// # It returns the matching layout format fragment
//
// For example, if the clock portion only had hours and minutes then we know the
// layout for this would be of the form `15:04` since we dont have seconds or fractional
// seconds to scan, but since we already consumed the hour portion of the time then we
// just need to make sure that we return the layout with either minutes, e.g, 15:04
// or with minutes and seconds 15:04:05
func (s *scanner) consumeClock() (string, error) {
	s.next() // ':'

	if err := s.consumeTwoDigits(Minutes); err != nil {
		return "", err
	}

	// Short circuit if we don't find seconds
	if s.peek() != ':' {
		return ":04", nil
	}

	s.next() // ':'

	if err := s.consumeTwoDigits(Seconds); err != nil {
		return "", err
	}

	if s.peek() == '.' {
		s.next()

		if err := s.consumeFraction(); err != nil {
			return "", err
		}
	}

	return ":04:05", nil
}

// consumeFraction will scan digits until it finds a non-digit rune
// It will only consume up-to nanosecond precision which is nine digits
func (s *scanner) consumeFraction() error {
	digits := 0
	for isDigit(s.peek()) || digits > NanoSecondPrecision {
		s.next()
		digits++
	}

	switch digits {
	case 0:
		return errors.New("Unable to parse time: fractional seconds have no digits")
	default:
		// Just fall through
	}

	return nil
}

func (s *scanner) consumeTwoDigits(unit TemporalUnit) error {
	if !isDigit(s.peek()) || !isDigit(s.peekNext()) {
		return fmt.Errorf("Unable to parse local time: %s are malformed", unit)
	}
	s.nextN(2)
	return nil
}

func (s *scanner) scanKeys() {
	for !s.isAtEnd() && isKey(s.peek()) {
		s.next()
	}

	lexeme := s.source[s.start:s.current]
	s.emitToken(bareKey, string(lexeme))
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

	s.emitToken(multilineBasicString, string(strValue))
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

	s.emitToken(basicString, string(strValue))
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
