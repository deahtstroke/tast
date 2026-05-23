package scanner

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
	HASHTAG
	NEW_LINE

	QUOTE

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

type Token struct {
	Type    TokenType
	Lexeme  string
	Literal any
	Line    int64
}
