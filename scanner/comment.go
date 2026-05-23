package scanner

func (s *Scanner) comment() {
	for s.peek() != '\n' && !s.isAtEnd() {
		s.advance()
	}
	commentValue := s.Source[s.start:s.current]

	// make up for finding a newline character
	s.line++
	s.addTokenValue(COMMENT, commentValue)
}
