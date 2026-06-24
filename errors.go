package tast

import (
	"fmt"
)

type parserErrorCode int

const (
	errMissingAssignmentAfterKey parserErrorCode = iota
	errMalformedTableKey
	errMissingClosingBracket
	errNoKeyAfterDot
	errUnrecognizedToken
	errParsingString
	errParsingInt
	errParsingFloat
	errParsingBool
	errDuplicateKey
	errDuplicateTable
	errUnspecifiedValueForKey
	errMissingNewLine
)

type parseError struct {
	Token   token
	Message string
	Code    parserErrorCode
}

func (c parserErrorCode) String() string {
	switch c {
	case errMissingAssignmentAfterKey:
		return "ErrMissingAssignmentAfterKey"
	case errMalformedTableKey:
		return "ErrMalformedTableKey"
	case errMissingClosingBracket:
		return "ErrMissingClosingBracket"
	case errNoKeyAfterDot:
		return "ErrNoKeyAfterDot"
	case errUnrecognizedToken:
		return "ErrUnrecognizedToken"
	case errParsingString:
		return "ErrParsingString"
	case errParsingInt:
		return "ErrParsingInt"
	case errParsingFloat:
		return "ErrParsingFloat"
	case errDuplicateKey:
		return "ErrDuplicateKey"
	case errUnspecifiedValueForKey:
		return "ErrUnspecifiedValueForKey"
	default:
		return "Unknown error code"
	}
}

func (e parseError) Error() string {
	return fmt.Sprintf("tast [%s]: parse error at %d:%d: %s",
		e.Code.String(), e.Token.Line, e.Token.Column, e.Message)
}

type scanError struct {
	Line    int
	Column  int
	Offset  int
	Message string
}

func (e scanError) Error() string {
	return fmt.Sprintf("tast: scan error at %d:%d (offset %d): %s",
		e.Line, e.Column, e.Offset, e.Message)
}
