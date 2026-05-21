package parser

import (
	"fmt"
	"math"
	"reflect"
	"testing"

	"github.com/deahtstroke/toml-ast/scanner"
)

func Test_TomlTables(t *testing.T) {
	tests := map[string]struct {
		tokens          []scanner.Token
		expectedLiteral string
		expectedNodes   int
		shouldErr       bool
		errorCount      int
		errorCodes      []ParseErrorCode
	}{
		"Table with basic string key": {
			tokens: []scanner.Token{
				{
					Type:    scanner.LEFT_BRACKET,
					Lexeme:  "[",
					Literal: string("["),
				},
				{
					Type:    scanner.BARE_KEY,
					Lexeme:  "HelloWorld",
					Literal: string("HelloWorld"),
				},
				{
					Type:    scanner.RIGHT_BRACKET,
					Lexeme:  "]",
					Literal: string("]"),
				},
				{
					Type: scanner.EOF,
				},
			},
			expectedLiteral: "[HelloWorld]",
			expectedNodes:   1,
			shouldErr:       false,
		},
		"Table with basic dotted string key": {
			tokens: []scanner.Token{
				{
					Type:    scanner.LEFT_BRACKET,
					Lexeme:  "[",
					Literal: string("["),
				},
				{
					Type:    scanner.BASIC_STRING,
					Lexeme:  "\"Hello.World\"",
					Literal: string("\"Hello.World\""),
				},
				{
					Type:    scanner.RIGHT_BRACKET,
					Lexeme:  "]",
					Literal: string("]"),
				},
				{
					Type: scanner.EOF,
				},
			},
			expectedLiteral: "[\"Hello.World\"]",
			expectedNodes:   1,
			shouldErr:       false,
		},
		"Table with bare dotted key": {
			tokens: []scanner.Token{
				{
					Type:    scanner.LEFT_BRACKET,
					Lexeme:  "[",
					Literal: string("["),
				},
				{
					Type:    scanner.BARE_KEY,
					Lexeme:  "hello",
					Literal: string("hello"),
				},
				{
					Type:    scanner.DOT,
					Lexeme:  ".",
					Literal: string("."),
				},
				{
					Type:    scanner.BARE_KEY,
					Lexeme:  "world",
					Literal: string("world"),
				},
				{
					Type:    scanner.RIGHT_BRACKET,
					Lexeme:  "]",
					Literal: string("]"),
				},
				{
					Type: scanner.EOF,
				},
			},
			expectedLiteral: "[hello.world]",
			expectedNodes:   1,
			shouldErr:       false,
		},
		"Table with bare dotted key and basic string": {
			tokens: []scanner.Token{
				{
					Type:    scanner.LEFT_BRACKET,
					Lexeme:  "[",
					Literal: string("["),
				},
				{
					Type:    scanner.BASIC_STRING,
					Lexeme:  "\"hello.world\"",
					Literal: string("\"hello.world\""),
				},
				{
					Type:    scanner.DOT,
					Lexeme:  ".",
					Literal: string("."),
				},
				{
					Type:    scanner.BARE_KEY,
					Lexeme:  "bar",
					Literal: string("bar"),
				},
				{
					Type:    scanner.RIGHT_BRACKET,
					Lexeme:  "]",
					Literal: string("]"),
				},
				{
					Type: scanner.EOF,
				},
			},
			expectedLiteral: "[\"hello.world\".bar]",
			expectedNodes:   1,
			shouldErr:       false,
		},
		"Should error on KeyValue node with no assignment after key": {
			tokens: []scanner.Token{
				{
					Type:    scanner.BARE_KEY,
					Lexeme:  "hello",
					Literal: string("hello"),
				},
				{
					Type:    scanner.FLOAT,
					Lexeme:  "3.14",
					Literal: float64(3.14),
				},
				{
					Type: scanner.EOF,
				},
			},
			shouldErr:  true,
			errorCount: 1,
			errorCodes: []ParseErrorCode{ErrMissingAssignmentAfterKey},
		},
		"Should error on KeyValue node with missing key after dot": {
			tokens: []scanner.Token{
				{
					Type:    scanner.BARE_KEY,
					Lexeme:  "hello",
					Literal: string("hello"),
				},
				{
					Type:    scanner.DOT,
					Lexeme:  ".",
					Literal: string("."),
				},
				{
					Type:    scanner.EQUAL,
					Lexeme:  "=",
					Literal: string("="),
				},
				{
					Type:    scanner.FLOAT,
					Lexeme:  "3.14",
					Literal: float64(3.14),
				},
				{
					Type: scanner.EOF,
				},
			},
			shouldErr:  true,
			errorCount: 1,
			errorCodes: []ParseErrorCode{ErrNoKeyAfterDot},
		},
		"Should report only missing key after dot even with no assignment": {
			tokens: []scanner.Token{
				{
					Type:    scanner.BARE_KEY,
					Lexeme:  "hello",
					Literal: string("hello"),
				},
				{
					Type:    scanner.DOT,
					Lexeme:  ".",
					Literal: string("."),
				},
				{
					Type:    scanner.FLOAT,
					Lexeme:  "3.14",
					Literal: float64(3.14),
				},
				{
					Type: scanner.EOF,
				},
			},
			shouldErr:  true,
			errorCount: 1,
			errorCodes: []ParseErrorCode{ErrNoKeyAfterDot},
		},
		"Should report error when key is malformed": {
			tokens: []scanner.Token{
				{
					Type:    scanner.LEFT_BRACKET,
					Lexeme:  "[",
					Literal: string("["),
				},
				{
					Type:    scanner.EQUAL,
					Lexeme:  "=",
					Literal: string("="),
				},
				{
					Type: scanner.EOF,
				},
			},
			shouldErr:  true,
			errorCount: 1,
			errorCodes: []ParseErrorCode{ErrMalformedTableKey},
		},
		"Should report both key is malformed and no key after dot": {
			tokens: []scanner.Token{
				{
					Type:    scanner.LEFT_BRACKET,
					Lexeme:  "[",
					Literal: string("["),
				},
				{
					Type:    scanner.EQUAL,
					Lexeme:  "=",
					Literal: string("="),
				},
				{
					Type:    scanner.NEW_LINE,
					Lexeme:  "\n",
					Literal: string("\n"),
				},
				{
					Type:    scanner.BARE_KEY,
					Lexeme:  "hello",
					Literal: string("hello"),
				},
				{
					Type:    scanner.DOT,
					Lexeme:  ".",
					Literal: string("."),
				},
				{
					Type:    scanner.FLOAT,
					Lexeme:  "3.14",
					Literal: float64(3.14),
				},
				{
					Type: scanner.EOF,
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
			doc, errs := parser.Parse()
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
		tokens      []scanner.Token
		expectedStr string
	}{
		{
			tokens: []scanner.Token{
				{
					Type:    scanner.BARE_KEY,
					Literal: string("foo"),
					Lexeme:  "foo",
				},
			},
			expectedStr: "foo",
		},
		{
			tokens: []scanner.Token{
				{
					Type:    scanner.BASIC_STRING,
					Literal: string("\"foo\""),
					Lexeme:  "\"foo\"",
				},
			},
			expectedStr: "\"foo\"",
		},
		{
			tokens: []scanner.Token{
				{
					Type:    scanner.BASIC_STRING,
					Literal: string("\"foo\""),
					Lexeme:  "\"foo\"",
				},
				{
					Type: scanner.DOT,
				},
				{
					Type:    scanner.BASIC_STRING,
					Literal: string("\"bar\""),
					Lexeme:  "\"bar\"",
				},
			},
			expectedStr: "\"foo\".\"bar\"",
		},
		{
			tokens: []scanner.Token{
				{
					Type:    scanner.BASIC_STRING,
					Literal: string("\"foo\""),
					Lexeme:  "\"foo\"",
				},
				{
					Type: scanner.DOT,
				},
				{
					Type:    scanner.BARE_KEY,
					Literal: string("bar"),
					Lexeme:  "bar",
				},
			},
			expectedStr: "\"foo\".bar",
		},
	}

	valueForms := []struct {
		token       scanner.Token
		expectedStr string
	}{
		{
			token: scanner.Token{
				Type:    scanner.INTEGER,
				Literal: int64(314),
				Lexeme:  "314",
			},
			expectedStr: "314",
		},
		{
			token: scanner.Token{
				Type:    scanner.INTEGER,
				Literal: int64(-314),
				Lexeme:  "-314",
			},
			expectedStr: "-314",
		},
		{
			token: scanner.Token{
				Type:    scanner.FLOAT,
				Literal: float64(3.14),
				Lexeme:  "3.14",
			},
			expectedStr: "3.14",
		},
		{
			token: scanner.Token{
				Type:    scanner.FLOAT,
				Literal: float64(-3.14),
				Lexeme:  "-3.14",
			},
			expectedStr: "-3.14",
		},
		{
			token: scanner.Token{
				Type:    scanner.BASIC_STRING,
				Literal: string("\"Roses are red, Violets are blue\""),
				Lexeme:  "\"Roses are red, Violets are blue\"",
			},
			expectedStr: "\"Roses are red, Violets are blue\"",
		},
		{
			token: scanner.Token{
				Type:    scanner.TRUE,
				Literal: bool(true),
				Lexeme:  "true",
			},
			expectedStr: "true",
		},
		{
			token: scanner.Token{
				Type:    scanner.FALSE,
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
				tokens := []scanner.Token{}
				tokens = append(tokens, key.tokens...)
				tokens = append(tokens, scanner.Token{Type: scanner.EQUAL})
				tokens = append(tokens, value.token)
				tokens = append(tokens, scanner.Token{Type: scanner.EOF})

				parser := NewParser(tokens)
				doc, errs := parser.Parse()
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
	tokens := []scanner.Token{
		{
			Type:    scanner.BASIC_STRING,
			Lexeme:  "HelloWorld",
			Literal: "HelloWorld",
			Line:    0,
		},
		{
			Type:    scanner.RIGHT_BRACKET,
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
		tokens      []scanner.Token
		expNodeType any
		expValue    any
	}{
		"negative integer": {
			tokens: []scanner.Token{
				{
					Type:    scanner.MINUS,
					Lexeme:  "-",
					Literal: "-",
					Line:    1,
				},
				{
					Type:    scanner.INTEGER,
					Lexeme:  "1234",
					Literal: int64(1234),
					Line:    1,
				},
			},
			expNodeType: &IntegerNode{},
			expValue:    int64(-1234),
		},
		"positive integer": {
			tokens: []scanner.Token{
				{
					Type:    scanner.PLUS,
					Lexeme:  "+",
					Literal: "+",
					Line:    1,
				},
				{
					Type:    scanner.INTEGER,
					Lexeme:  "12341",
					Literal: int64(12341),
					Line:    1,
				},
			},
			expNodeType: &IntegerNode{},
			expValue:    int64(12341),
		},
		"unsigned integer": {
			tokens: []scanner.Token{
				{
					Type:    scanner.INTEGER,
					Lexeme:  "12341",
					Literal: int64(12341),
					Line:    1,
				},
			},
			expNodeType: &IntegerNode{},
			expValue:    int64(12341),
		},
		"negative floating point": {
			tokens: []scanner.Token{
				{
					Type:    scanner.MINUS,
					Lexeme:  "-",
					Literal: "-",
					Line:    0,
				},
				{
					Type:    scanner.FLOAT,
					Lexeme:  "3.12451",
					Literal: float64(3.12451),
					Line:    0,
				},
			},
			expNodeType: &FloatNode{},
			expValue:    float64(-3.12451),
		},
		"positive floating point": {
			tokens: []scanner.Token{
				{
					Type:    scanner.PLUS,
					Lexeme:  "+",
					Literal: "+",
					Line:    0,
				},
				{
					Type:    scanner.FLOAT,
					Lexeme:  "3.12451",
					Literal: float64(3.12451),
					Line:    0,
				},
			},
			expNodeType: &FloatNode{},
			expValue:    float64(3.12451),
		},
		"unsigned floating point": {
			tokens: []scanner.Token{
				{
					Type:    scanner.FLOAT,
					Lexeme:  "3.12451",
					Literal: float64(3.12451),
					Line:    0,
				},
			},
			expNodeType: &FloatNode{},
			expValue:    float64(3.12451),
		},
		"negative infinity": {
			tokens: []scanner.Token{
				{
					Type:    scanner.MINUS,
					Lexeme:  "-",
					Literal: "-",
					Line:    0,
				},
				{
					Type:    scanner.INF,
					Lexeme:  "inf",
					Literal: nil,
					Line:    0,
				},
			},
			expNodeType: &IntegerNode{},
			expValue:    -int64(math.MaxInt64),
		},
		"positive infinity": {
			tokens: []scanner.Token{
				{
					Type:    scanner.PLUS,
					Lexeme:  "+",
					Literal: "+",
					Line:    0,
				},
				{
					Type:    scanner.INF,
					Lexeme:  "inf",
					Literal: nil,
					Line:    0,
				},
			},
			expNodeType: &IntegerNode{},
			expValue:    int64(math.MaxInt64),
		},
		"unsigned infinity": {
			tokens: []scanner.Token{
				{
					Type:    scanner.INF,
					Lexeme:  "inf",
					Literal: nil,
					Line:    0,
				},
			},
			expNodeType: &IntegerNode{},
			expValue:    int64(math.MaxInt64),
		},
		"false": {
			tokens: []scanner.Token{
				{
					Type:   scanner.FALSE,
					Lexeme: "false",
					Line:   0,
				},
			},
			expNodeType: &BooleanNode{},
			expValue:    false,
		},
		"true": {
			tokens: []scanner.Token{
				{
					Type:   scanner.TRUE,
					Lexeme: "true",
					Line:   0,
				},
			},
			expNodeType: &BooleanNode{},
			expValue:    true,
		},
		"basic string": {
			tokens: []scanner.Token{
				{
					Type:    scanner.BASIC_STRING,
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
