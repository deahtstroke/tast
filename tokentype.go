package tast

func (t tokenType) String() string {
	switch t {
	case comment:
		return "Comment"
	case leftBracket:
		return "Left Bracket"
	case rightBracket:
		return "Right Bracket"
	case leftBrace:
		return "Left Curly Brace"
	case rightBrace:
		return "Right Curly Brace"
	case comma:
		return "Comma"
	case dot:
		return "Dot"
	case minus:
		return "Minus"
	case plus:
		return "Plus"
	case slash:
		return "Slash"
	case star:
		return "Star"
	case equal:
		return "Equal"
	case newLine:
		return "New Line"
	case basicString:
		return "Basic String"
	case multilineBasicString:
		return "Multi-line Basic String"
	case literalString:
		return "Literal String"
	case multilineLiteralString:
		return "Multi-line Literal String"
	case floatPoint:
		return "Floating Point"
	case integer:
		return "Integer"
	case localDate:
		return "Local Date"
	case localTime:
		return "Local Time"
	case localDateTime:
		return "Local Date Time"
	case offsetDateTime:
		return "Offset Date Time"
	case bareKey:
		return "Bare Key"
	case boolean:
		return "Boolean"
	case infinity:
		return "Infinity"
	case nan:
		return "NaN"
	case eof:
		return "EOF"
	default:
		return ""
	}
}

type tokenType uint32

const (
	_ tokenType = iota
	comment
	leftBracket
	rightBracket
	leftBrace
	rightBrace
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

	localDate
	localTime
	localDateTime
	offsetDateTime

	bareKey
	// Reserved keywords
	boolean
	infinity
	nan

	eof
)
