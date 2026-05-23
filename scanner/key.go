package scanner

func (s *Scanner) key() {
	for !s.isAtEnd() && isKey(s.peek()) {
		s.advance()
	}

	text := string(s.Source[s.start:s.current])
	token, ok := reserved[text]
	if ok {
		s.addToken(token)
		return
	}

	lexeme := s.Source[s.start:s.current]
	s.addTokenValue(BARE_KEY, string(lexeme))
}

func isKey(b byte) bool {
	return isAlphanumeric(b) || b == '_' || b == '-'
}
