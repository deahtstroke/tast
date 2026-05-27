package tast

import (
	"testing"
)

func Test_ParseTable(t *testing.T) {
	tests := map[string]struct {
		tokens           []Token
		expectedDocument *Document
		shouldErr        bool
		errorCount       int
		errorCodes       []ParseErrorCode
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
			expectedDocument: &Document{
				Content: []Node{
					&TableNode{
						Key: &KeyNode{
							Segments: []string{"HelloWorld"},
						},
						Children: []Node{},
					},
				},
			},
			shouldErr: false,
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
			expectedDocument: &Document{},
			shouldErr:        false,
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
			shouldErr: false,
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
			shouldErr: false,
		},
		"Should parse leading comments as part of table": {
			tokens: []Token{
				{
					Type:    COMMENT,
					Lexeme:  "# This is a useful comment",
					Literal: string("# This is a useful comment"),
				},
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

			if len(doc.Content) != len(params.expectedDocument.Content) {
				t.Fatalf("expected %d nodes in the document, got %d", len(params.expectedDocument.Content), len(doc.Content))
			}

			for i := range len(doc.Content) {
				assertNode(t, params.expectedDocument.Content[i], doc.Content[i])
			}
		})
	}
}

func Test_KeyNode_Segments(t *testing.T) {
	keyForms := map[string]struct {
		keyTokens []Token
		expected  []string
	}{
		"bare key": {
			keyTokens: []Token{
				{
					Type:    BARE_KEY,
					Literal: string("hello_world"),
					Lexeme:  "hello_world",
				},
			},
			expected: []string{
				"hello_world",
			},
		},
	}

	for test, tt := range keyForms {
		t.Run(test, func(t *testing.T) {
			parser := NewParser(tt.keyTokens)
			keyNode := parser.Key()

			if keyNode == nil {
				t.Fatal("skmething went wrong when parsing the key")
			}

			if len(keyNode.Segments) != len(tt.expected) {
				t.Fatalf("expected %d segments, got: %d", len(tt.expected), len(keyNode.Segments))
			}

			for i := range len(keyNode.Segments) {
				if keyNode.Segments[i] != tt.expected[i] {
					t.Fatalf("expected %s, got: %s", keyNode.Segments[i], tt.expected[i])
				}
			}
		})
	}
}

func assertNode(t *testing.T, expected, got Node) {
	t.Helper()

	switch e := expected.(type) {
	case *TableNode:
		g, ok := got.(*TableNode)
		if !ok {
			t.Fatalf("Expected *TableNode, got %T", got)
		}
		assertKeyNode(t, e.Key, g.Key)

		if len(e.Children) != len(g.Children) {
			t.Fatalf("expected %d children, got: %d", len(e.Children), len(g.Children))
		}
		for i := range len(g.Children) {
			assertNode(t, e.Children[i], g.Children[i])
		}
	case *KeyValueNode:
		g, ok := got.(*KeyValueNode)
		if !ok {
			t.Fatalf("expected *KeyValueNode, got %T", got)
		}
		assertKeyNode(t, e.Key, g.Key)
		assertNode(t, e.Value, g.Value)
	case *StringNode:
		g, ok := got.(*StringNode)
		if !ok {
			t.Fatalf("expected *StringNode, got %T", got)
		}

		if g.Value != e.Value {
			t.Fatalf("expected value %s, got %s", e.Value, g.Value)
		}
	case *IntegerNode:
		g, ok := got.(*IntegerNode)
		if !ok {
			t.Fatalf("expected *IntegerNode, got %T", got)
		}

		if g.Value != e.Value {
			t.Fatalf("expected value %d, got %d", e.Value, g.Value)
		}
	case *FloatNode:
		g, ok := got.(*FloatNode)
		if !ok {
			t.Fatalf("expected *FloatNode, got %T", got)
		}

		if g.Value != e.Value {
			t.Fatalf("expected value %f, got %f", e.Value, g.Value)
		}
	case *BooleanNode:
		g, ok := got.(*BooleanNode)
		if !ok {
			t.Fatalf("expected *BooleanNode, got %T", got)
		}

		if g.Value != e.Value {
			t.Fatalf("expected value %t, got %t", e.Value, g.Value)
		}
	}
}

func assertKeyNode(t *testing.T, expected, got *KeyNode) {
	t.Helper()

	if len(expected.Segments) != len(got.Segments) {
		t.Fatalf("expected %d segments, got %d", len(expected.Segments), len(got.Segments))
	}

	for i := range len(got.Segments) {
		if got.Segments[i] != expected.Segments[i] {
			t.Fatalf("expected segment %s at index %d to be equal, got: %s", expected.Segments[i], i, got.Segments[i])
		}
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
