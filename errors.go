package tast

import (
	"fmt"

	"github.com/deahtstroke/tast/scanner"
)

type ParseErrorCode int

const (
	ErrMissingAssignmentAfterKey ParseErrorCode = iota
	ErrMalformedTableKey
	ErrMissingClosingBracket
	ErrNoKeyAfterDot
	ErrUnrecognizedToken
	ErrParsingString
	ErrParsingInt
	ErrParsingFloat
)

type ParseError struct {
	Token   scanner.Token
	Message string
	Code    ParseErrorCode
}

func (c ParseErrorCode) String() string {
	switch c {
	case ErrMissingAssignmentAfterKey:
		return "ErrMissingAssignmentAfterKey"
	case ErrMalformedTableKey:
		return "ErrMalformedTableKey"
	case ErrMissingClosingBracket:
		return "ErrMissingClosingBracket"
	case ErrNoKeyAfterDot:
		return "ErrNoKeyAfterDot"
	case ErrUnrecognizedToken:
		return "ErrUnrecognizedToken"
	case ErrParsingString:
		return "ErrParsingString"
	case ErrParsingInt:
		return "ErrParsingInt"
	case ErrParsingFloat:
		return "ErrParsingFloat"
	default:
		return "Unknown error code"
	}
}

func (e ParseError) Error() string {
	return fmt.Sprintf("[line %d] at %q (code %s): %s", e.Token.Line, e.Token.Lexeme, e.Code, e.Message)
}
