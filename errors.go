package tast

import (
	"fmt"
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
	ErrDuplicateKey
	ErrDuplicateTable
	ErrUnspecifiedValueForKey
	ErrMissingNewLine
)

type ParseError struct {
	Token   token
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
	case ErrDuplicateKey:
		return "ErrDuplicateKey"
	case ErrUnspecifiedValueForKey:
		return "ErrUnspecifiedValueForKey"
	default:
		return "Unknown error code"
	}
}

func (e ParseError) Error() string {
	return fmt.Sprintf("tast [%s]: parse error at %d:%d: %s",
		e.Code.String(), e.Token.Line, e.Token.Column, e.Message)
}

type ScanError struct {
	Line    int
	Column  int
	Offset  int
	Message string
}

func (e ScanError) Error() string {
	return fmt.Sprintf("tast: scan error at %d:%d (offset %d): %s",
		e.Line, e.Column, e.Offset, e.Message)
}
