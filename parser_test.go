package tast

import (
	"fmt"
	"math"
	"reflect"
	"testing"
)

func Test_TomlTables(t *testing.T) {
	tests := map[string]struct {
		tokens          []Token
		expectedLiteral string
		expectedNodes   int
		shouldErr       bool
		errorCount      int
		errorCodes      []ParseErrorCode
	}{
		"Table with basic string key": {
			tokens: []Token{
				{
					Type:    LEFT_BRACKET,
					Lexeme:  "[",
					Literal: string("["),
				},
				{
					Type:    BARE_KEY,
					Lexeme:  "HelloWorld",
					Literal: string("HelloWorld"),
				},
				{
					Type:    RIGHT_BRACKET,
					Lexeme:  "]",
					Literal: string("]"),
				},
				{
					Type: EOF,
				},
			},
			expectedLiteral: "[HelloWorld]",
			expectedNodes:   1,
			shouldErr:       false,
		},
		"Table with basic dotted string key": {
			tokens: []Token{
				{
					Type:    LEFT_BRACKET,
					Lexeme:  "[",
					Literal: string("["),
				},
				{
					Type:    BASIC_STRING,
					Lexeme:  "\"Hello.World\"",
					Literal: string("\"Hello.World\""),
				},
				{
					Type:    RIGHT_BRACKET,
					Lexeme:  "]",
					Literal: string("]"),
				},
				{
					Type: EOF,
				},
			},
			expectedLiteral: "[\"Hello.World\"]",
			expectedNodes:   1,
			shouldErr:       false,
		},
		"Table with bare dotted key": {
			tokens: []Token{
				{
					Type:    LEFT_BRACKET,
					Lexeme:  "[",
					Literal: string("["),
				},
				{
					Type:    BARE_KEY,
					Lexeme:  "hello",
					Literal: string("hello"),
				},
				{
					Type:    DOT,
					Lexeme:  ".",
					Literal: string("."),
				},
				{
					Type:    BARE_KEY,
					Lexeme:  "world",
					Literal: string("world"),
				},
				{
					Type:    RIGHT_BRACKET,
					Lexeme:  "]",
					Literal: string("]"),
				},
				{
					Type: EOF,
				},
			},
			expectedLiteral: "[hello.world]",
			expectedNodes:   1,
			shouldErr:       false,
		},
		"Table with bare dotted key and basic string": {
			tokens: []Token{
				{
					Type:    LEFT_BRACKET,
					Lexeme:  "[",
					Literal: string("["),
				},
				{
					Type:    BASIC_STRING,
					Lexeme:  "\"hello.world\"",
					Literal: string("\"hello.world\""),
				},
				{
					Type:    DOT,
					Lexeme:  ".",
					Literal: string("."),
				},
				{
					Type:    BARE_KEY,
					Lexeme:  "bar",
					Literal: string("bar"),
				},
				{
					Type:    RIGHT_BRACKET,
					Lexeme:  "]",
					Literal: string("]"),
				},
				{
					Type: EOF,
				},
			},
			expectedLiteral: "[\"hello.world\".bar]",
			expectedNodes:   1,
			shouldErr:       false,
		},
		"Should error on KeyValue node with no assignment after key": {
			tokens: []Token{
				{
					Type:    BARE_KEY,
					Lexeme:  "hello",
					Literal: string("hello"),
				},
				{
					Type:    FLOAT,
					Lexeme:  "3.14",
					Literal: float64(3.14),
				},
				{
					Type: EOF,
				},
			},
			shouldErr:  true,
			errorCount: 1,
			errorCodes: []ParseErrorCode{ErrMissingAssignmentAfterKey},
		},
		"Should error on KeyValue node with missing key after dot": {
			tokens: []Token{
				{
					Type:    BARE_KEY,
					Lexeme:  "hello",
					Literal: string("hello"),
				},
				{
					Type:    DOT,
					Lexeme:  ".",
					Literal: string("."),
				},
				{
					Type:    EQUAL,
					Lexeme:  "=",
					Literal: string("="),
				},
				{
					Type:    FLOAT,
					Lexeme:  "3.14",
					Literal: float64(3.14),
				},
				{
					Type: EOF,
				},
			},
			shouldErr:  true,
			errorCount: 1,
			errorCodes: []ParseErrorCode{ErrNoKeyAfterDot},
		},
		"Should report only missing key after dot even with no assignment": {
			tokens: []Token{
				{
					Type:    BARE_KEY,
					Lexeme:  "hello",
					Literal: string("hello"),
				},
				{
					Type:    DOT,
					Lexeme:  ".",
					Literal: string("."),
				},
				{
					Type:    FLOAT,
					Lexeme:  "3.14",
					Literal: float64(3.14),
				},
				{
					Type: EOF,
				},
			},
			shouldErr:  true,
			errorCount: 1,
			errorCodes: []ParseErrorCode{ErrNoKeyAfterDot},
		},
		"Should report error when key is malformed": {
			tokens: []Token{
				{
					Type:    LEFT_BRACKET,
					Lexeme:  "[",
					Literal: string("["),
				},
				{
					Type:    EQUAL,
					Lexeme:  "=",
					Literal: string("="),
				},
				{
					Type: EOF,
				},
			},
			shouldErr:  true,
			errorCount: 1,
			errorCodes: []ParseErrorCode{ErrMalformedTableKey},
		},
		"Should report both key is malformed and no key after dot": {
			tokens: []Token{
				{
					Type:    LEFT_BRACKET,
					Lexeme:  "[",
					Literal: string("["),
				},
				{
					Type:    EQUAL,
					Lexeme:  "=",
					Literal: string("="),
				},
				{
					Type:    NEW_LINE,
					Lexeme:  "\n",
					Literal: string("\n"),
				},
				{
					Type:    BARE_KEY,
					Lexeme:  "hello",
					Literal: string("hello"),
				},
				{
					Type:    DOT,
					Lexeme:  ".",
					Literal: string("."),
				},
				{
					Type:    FLOAT,
					Lexeme:  "3.14",
					Literal: float64(3.14),
				},
				{
					Type: EOF,
				},
			},
			shouldErr:  true,
			errorCount: 2,
			errorCodes: []ParseErrorCode{ErrMalformedTableKey, ErrNoKeyAfterDot},
		},
	}

	for test, params := range tests {
		t.Run(test, func(t *testing.T) {
			parser := NewParser(params.tokens)
			doc, errs := parser.parse()
			if params.shouldErr {
				if len(errs) != params.errorCount {
					t.Fatalf("Expecting %d errors, found %d: %v", params.errorCount, len(errs), errs)
				}

				for _, code := range params.errorCodes {
					if !containsErrorCode(errs, code) {
						t.Fatalf("Expeced error code %v but was not found in %v", code, errs)
					}
				}
				return
			}

			if len(errs) != 0 {
				t.Fatalf("Incorrect parse tree: %+v", parser.errors)
			}

			length := len(doc.Content)
			if length != params.expectedNodes {
				t.Errorf("Incorrect length of nodes for root document node: expected: %d, got: %d", params.expectedNodes, length)
			}

			tokenLiteral := doc.Content[0].NodeLexeme()
			if tokenLiteral != params.expectedLiteral {
				t.Errorf("Incorrect token literal. Expected: %s. Got: %s", params.expectedLiteral, tokenLiteral)
			}
		})
	}
}

func containsErrorCode(errs []ParseError, code ParseErrorCode) bool {
	for _, err := range errs {
		if err.Code == code {
			return true
		}
	}
	return false
}

func Test_ParseKeyValue(t *testing.T) {
	keyForms := []struct {
		tokens      []Token
		expectedStr string
	}{
		{
			tokens: []Token{
				{
					Type:    BARE_KEY,
					Literal: string("foo"),
					Lexeme:  "foo",
				},
			},
			expectedStr: "foo",
		},
		{
			tokens: []Token{
				{
					Type:    BASIC_STRING,
					Literal: string("\"foo\""),
					Lexeme:  "\"foo\"",
				},
			},
			expectedStr: "\"foo\"",
		},
		{
			tokens: []Token{
				{
					Type:    BASIC_STRING,
					Literal: string("\"foo\""),
					Lexeme:  "\"foo\"",
				},
				{
					Type: DOT,
				},
				{
					Type:    BASIC_STRING,
					Literal: string("\"bar\""),
					Lexeme:  "\"bar\"",
				},
			},
			expectedStr: "\"foo\".\"bar\"",
		},
		{
			tokens: []Token{
				{
					Type:    BASIC_STRING,
					Literal: string("\"foo\""),
					Lexeme:  "\"foo\"",
				},
				{
					Type: DOT,
				},
				{
					Type:    BARE_KEY,
					Literal: string("bar"),
					Lexeme:  "bar",
				},
			},
			expectedStr: "\"foo\".bar",
		},
	}

	valueForms := []struct {
		token       Token
		expectedStr string
	}{
		{
			token: Token{
				Type:    INTEGER,
				Literal: int64(314),
				Lexeme:  "314",
			},
			expectedStr: "314",
		},
		{
			token: Token{
				Type:    INTEGER,
				Literal: int64(-314),
				Lexeme:  "-314",
			},
			expectedStr: "-314",
		},
		{
			token: Token{
				Type:    FLOAT,
				Literal: float64(3.14),
				Lexeme:  "3.14",
			},
			expectedStr: "3.14",
		},
		{
			token: Token{
				Type:    FLOAT,
				Literal: float64(-3.14),
				Lexeme:  "-3.14",
			},
			expectedStr: "-3.14",
		},
		{
			token: Token{
				Type:    BASIC_STRING,
				Literal: string("\"Roses are red, Violets are blue\""),
				Lexeme:  "\"Roses are red, Violets are blue\"",
			},
			expectedStr: "\"Roses are red, Violets are blue\"",
		},
		{
			token: Token{
				Type:    TRUE,
				Literal: bool(true),
				Lexeme:  "true",
			},
			expectedStr: "true",
		},
		{
			token: Token{
				Type:    FALSE,
				Literal: bool(false),
				Lexeme:  "false",
			},
			expectedStr: "false",
		},
	}

	for _, key := range keyForms {
		for _, value := range valueForms {
			testName := fmt.Sprintf("%s = %s", key.expectedStr, value.expectedStr)
			t.Run(testName, func(t *testing.T) {
				tokens := []Token{}
				tokens = append(tokens, key.tokens...)
				tokens = append(tokens, Token{Type: EQUAL})
				tokens = append(tokens, value.token)
				tokens = append(tokens, Token{Type: EOF})

				parser := NewParser(tokens)
				doc, errs := parser.parse()
				if len(errs) != 0 {
					t.Fatalf("Not expecting errors, got: %+v", errs)
				}

				actual, ok := doc.Content[0].(*KeyValueNode)
				if !ok {
					t.Fatalf("Not a KeyValueNode instance")
				}

				expected := fmt.Sprintf("%s = %s", key.expectedStr, value.expectedStr)
				if actual.NodeLexeme() != expected {
					t.Fatalf("Non matching. Expected: %s. Got: %s", expected, actual.NodeLexeme())
				}
			})
		}
	}
}

func Test_Table(t *testing.T) {
	tokens := []Token{
		{
			Type:    BASIC_STRING,
			Lexeme:  "HelloWorld",
			Literal: "HelloWorld",
			Line:    0,
		},
		{
			Type:    RIGHT_BRACKET,
			Lexeme:  "[",
			Literal: "[",
			Line:    0,
		},
	}
	p := NewParser(tokens)
	tableNode := p.Table()
	if tableNode == nil {
		t.Fatalf("Parse tree is incorrect")
	}

	keyNode := tableNode.Key
	if keyNode.Segments[0] != "HelloWorld" {
		t.Errorf("Wrong key value. Expecting: HelloWorld. Got: %s", tableNode.Key.NodeLiteral())
	}
}

func Test_Value(t *testing.T) {
	tests := map[string]struct {
		tokens      []Token
		expNodeType any
		expValue    any
	}{
		"negative integer": {
			tokens: []Token{
				{
					Type:    MINUS,
					Lexeme:  "-",
					Literal: "-",
					Line:    1,
				},
				{
					Type:    INTEGER,
					Lexeme:  "1234",
					Literal: int64(1234),
					Line:    1,
				},
			},
			expNodeType: &IntegerNode{},
			expValue:    int64(-1234),
		},
		"positive integer": {
			tokens: []Token{
				{
					Type:    PLUS,
					Lexeme:  "+",
					Literal: "+",
					Line:    1,
				},
				{
					Type:    INTEGER,
					Lexeme:  "12341",
					Literal: int64(12341),
					Line:    1,
				},
			},
			expNodeType: &IntegerNode{},
			expValue:    int64(12341),
		},
		"unsigned integer": {
			tokens: []Token{
				{
					Type:    INTEGER,
					Lexeme:  "12341",
					Literal: int64(12341),
					Line:    1,
				},
			},
			expNodeType: &IntegerNode{},
			expValue:    int64(12341),
		},
		"negative floating point": {
			tokens: []Token{
				{
					Type:    MINUS,
					Lexeme:  "-",
					Literal: "-",
					Line:    0,
				},
				{
					Type:    FLOAT,
					Lexeme:  "3.12451",
					Literal: float64(3.12451),
					Line:    0,
				},
			},
			expNodeType: &FloatNode{},
			expValue:    float64(-3.12451),
		},
		"positive floating point": {
			tokens: []Token{
				{
					Type:    PLUS,
					Lexeme:  "+",
					Literal: "+",
					Line:    0,
				},
				{
					Type:    FLOAT,
					Lexeme:  "3.12451",
					Literal: float64(3.12451),
					Line:    0,
				},
			},
			expNodeType: &FloatNode{},
			expValue:    float64(3.12451),
		},
		"unsigned floating point": {
			tokens: []Token{
				{
					Type:    FLOAT,
					Lexeme:  "3.12451",
					Literal: float64(3.12451),
					Line:    0,
				},
			},
			expNodeType: &FloatNode{},
			expValue:    float64(3.12451),
		},
		"negative infinity": {
			tokens: []Token{
				{
					Type:    MINUS,
					Lexeme:  "-",
					Literal: "-",
					Line:    0,
				},
				{
					Type:    INF,
					Lexeme:  "inf",
					Literal: nil,
					Line:    0,
				},
			},
			expNodeType: &IntegerNode{},
			expValue:    -int64(math.MaxInt64),
		},
		"positive infinity": {
			tokens: []Token{
				{
					Type:    PLUS,
					Lexeme:  "+",
					Literal: "+",
					Line:    0,
				},
				{
					Type:    INF,
					Lexeme:  "inf",
					Literal: nil,
					Line:    0,
				},
			},
			expNodeType: &IntegerNode{},
			expValue:    int64(math.MaxInt64),
		},
		"unsigned infinity": {
			tokens: []Token{
				{
					Type:    INF,
					Lexeme:  "inf",
					Literal: nil,
					Line:    0,
				},
			},
			expNodeType: &IntegerNode{},
			expValue:    int64(math.MaxInt64),
		},
		"false": {
			tokens: []Token{
				{
					Type:   FALSE,
					Lexeme: "false",
					Line:   0,
				},
			},
			expNodeType: &BooleanNode{},
			expValue:    false,
		},
		"true": {
			tokens: []Token{
				{
					Type:   TRUE,
					Lexeme: "true",
					Line:   0,
				},
			},
			expNodeType: &BooleanNode{},
			expValue:    true,
		},
		"basic string": {
			tokens: []Token{
				{
					Type:    BASIC_STRING,
					Lexeme:  "hello world!",
					Literal: "hello world!",
					Line:    0,
				},
			},
			expNodeType: &StringNode{},
			expValue:    "hello world!",
		},
	}

	for test, tt := range tests {
		t.Run(test, func(t *testing.T) {
			parser := NewParser(tt.tokens)
			node := parser.value()
			if node == nil {
				t.Fatalf("Incorrect parse tree")
			}

			gotType := reflect.TypeOf(node)
			expType := reflect.TypeOf(tt.expNodeType)

			if gotType != expType {
				t.Fatalf("Expected node type %s, got %s", expType, gotType)
			}

			gotValue := reflect.ValueOf(node).Elem().FieldByName("Value").Interface()
			if gotValue != tt.expValue {
				t.Fatalf("expected value %v (%T), got %v (%T)", tt.expValue, tt.expValue, gotValue, gotValue)
			}
		})
	}
}
