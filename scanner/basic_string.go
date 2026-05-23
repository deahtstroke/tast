package scanner

func (s *Scanner) multilineBasicString() {
	for !s.isAtEnd() && !s.isMultilineClosing() {
		if s.peek() == '\n' {
			s.line++
		}
		s.advance()
	}

	if s.isAtEnd() {
		panic("Unterminated multiline basic string")
	}

	s.advance() // Trim first '"'
	s.advance() // Trim second '"'
	s.advance() // Trim third '"'

	strValue := s.Source[s.start+3 : s.current-3]

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

func (s *Scanner) basicString() {
	for !s.isAtEnd() && s.peek() != '"' {
		if s.peek() == '\n' {
			s.line++
		}
		s.advance()
	}

	if s.isAtEnd() {
		panic("Unterminated basic string")
	}

	s.advance()

	// lexeme = s.Source[s.start:s.current] → "hello" (with quotes, handled by addTokenValue)
	// literal = just the content between the quotes
	strValue := s.Source[s.start+1 : s.current-1]

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
