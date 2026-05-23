package scanner

import (
	"strconv"
	"strings"
)

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

	lexeme := s.Source[s.start:s.current]

	// Cleanup any underscores
	cleaned := strings.ReplaceAll(string(lexeme), "_", "")

	if isFloatingPoint {
		floatVal, _ := strconv.ParseFloat(cleaned, 64)
		s.addTokenValue(FLOAT, floatVal)
	} else {
		intVal, _ := strconv.Atoi(cleaned)
		s.addTokenValue(INTEGER, intVal)
	}
}
