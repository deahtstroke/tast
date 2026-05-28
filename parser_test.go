package tast

import (
	"fmt"
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
		// Comments
		// TODO: This test is wrong because it doesn't assert comments
		"General comments with key-value pairs": {
			tokens: InitializeTokens(
				Comment("# This is a full-line comment"),
				BareKey("key"), Equal(), BasicString("value"), Comment("# This is a comment at the end of a line"),
				BareKey("another"), Equal(), BasicString("# This is not a comment"),
			),
			expectedDocument: &Document{
				Content: []Node{
					&KeyValueNode{
						LeadingComments: []Trivia{{Lexeme: "# This is a full-line comment"}},
						TrailingComment: &Trivia{Lexeme: "# This is a comment at the end of a line"},
						Key: &KeyNode{
							Segments: []string{"key"},
						},
						Value: &StringNode{
							Value: "value",
						},
					},
					&KeyValueNode{
						Key: &KeyNode{
							Segments: []string{"another"},
						},
						Value: &StringNode{
							Value: "# This is not a comment",
						},
					},
				},
			},
		},
		"Accumulated comments with tables": {
			tokens: InitializeTokens(
				Comment("# This is a comment"),
				Comment("# This is another comment"),
				LeftBracket(), BareKey("key"), RightBracket(), Comment("# This is in the same line as the table"),
			),
			expectedDocument: &Document{
				Content: []Node{
					&TableNode{
						LeadingComments: []Trivia{{Lexeme: "# This is a comment"}, {Lexeme: "# This is another comment"}},
						TrailingComment: &Trivia{"# This is in the same line as the table"},
						Key: &KeyNode{
							Segments: []string{"key"},
						},
					},
				},
			},
		},
		// Key-value pair
		"Basic key-value pair": {
			tokens: InitializeTokens(
				BareKey("key"), Equal(), BasicString("value"),
			),
			expectedDocument: &Document{
				Content: []Node{
					&KeyValueNode{
						Key: &KeyNode{
							Segments: []string{"key"},
						},
						Value: &StringNode{
							Value: "value",
						},
					},
				},
			},
		},
		"Should error on unspecified values": {
			tokens: InitializeTokens(
				BareKey("key"), Equal(),
			),
			shouldErr:  true,
			errorCount: 1,
			errorCodes: []ParseErrorCode{ErrUnspecifiedValueForKey},
		},
		// Keys
		"BareKeys": {
			tokens: InitializeTokens(
				BareKey("key"), Equal(), BasicString("value"),
				BareKey("bare_key"), Equal(), BasicString("value"),
				BareKey("bare-key"), Equal(), BasicString("value"),
				BareKey("1234"), Equal(), BasicString("value"),
			),
			expectedDocument: &Document{
				Content: []Node{
					&KeyValueNode{
						Key: &KeyNode{
							Segments: []string{"key"},
						},
						Value: &StringNode{
							Value: "value",
						},
					},
					&KeyValueNode{
						Key: &KeyNode{
							Segments: []string{"bare_key"},
						},
						Value: &StringNode{
							Value: "value",
						},
					},
					&KeyValueNode{
						Key: &KeyNode{
							Segments: []string{"bare-key"},
						},
						Value: &StringNode{
							Value: "value",
						},
					},
					&KeyValueNode{
						Key: &KeyNode{
							Segments: []string{"1234"},
						},
						Value: &StringNode{
							Value: "value",
						},
					},
				},
			},
		},
		"Quoted Keys": {
			tokens: InitializeTokens(
				BasicString("127.0.0.1"), Equal(), BasicString("value"),
				BasicString("character encoding"), Equal(), BasicString("value"),
				LiteralString("key2"), Equal(), BasicString("value"),
				LiteralString("ʎǝʞ"), Equal(), BasicString("value"),
				LiteralString("quoted \"value\""), Equal(), BasicString("value")),
			expectedDocument: &Document{
				Content: []Node{
					&KeyValueNode{
						Key: &KeyNode{
							Segments: []string{"127.0.0.1"},
						},
						Value: &StringNode{
							Value: "value",
						},
					},
					&KeyValueNode{
						Key: &KeyNode{
							Segments: []string{"character encoding"},
						},
						Value: &StringNode{
							Value: "value",
						},
					},
					&KeyValueNode{
						Key: &KeyNode{
							Segments: []string{"key2"},
						},
						Value: &StringNode{
							Value: "value",
						},
					},
					&KeyValueNode{
						Key: &KeyNode{
							Segments: []string{"ʎǝʞ"},
						},
						Value: &StringNode{
							Value: "value",
						},
					},
					&KeyValueNode{
						Key: &KeyNode{
							Segments: []string{"quoted \"value\""},
						},
						Value: &StringNode{
							Value: "value",
						},
					},
				},
			},
		},
		"Dotted-keys": {
			tokens: InitializeTokens(
				BareKey("name"), Equal(), BasicString("Orange"),
				BareKey("physical"), Dot(), BasicString("color"), Equal(), BasicString("orange"),
				BareKey("physical"), Dot(), BasicString("shape"), Equal(), BasicString("round"),
				BareKey("site"), Dot(), BasicString("google.com"), Equal(), True(),
			),
			expectedDocument: &Document{
				Content: []Node{
					&KeyValueNode{
						Key: &KeyNode{
							Segments: []string{"name"},
						},
						Value: &StringNode{
							Value: "Orange",
						},
					},
					&KeyValueNode{
						Key: &KeyNode{
							Segments: []string{"physical", "color"},
						},
						Value: &StringNode{
							Value: "orange",
						},
					},
					&KeyValueNode{
						Key: &KeyNode{
							Segments: []string{"physical", "shape"},
						},
						Value: &StringNode{
							Value: "round",
						},
					},
					&KeyValueNode{
						Key: &KeyNode{
							Segments: []string{"site", "google.com"},
						},
						Value: &BooleanNode{
							Value: true,
						},
					},
				},
			},
		},
		"Duplicate keys should error": {
			tokens: InitializeTokens(
				BareKey("name"), Equal(), BasicString("Tom"),
				BareKey("name"), Equal(), BasicString("Pradyun"),
			),
			shouldErr:  true,
			errorCount: 1,
			errorCodes: []ParseErrorCode{ErrDuplicateKey},
		},
		"Barekeys and quotted keys are equivalent on duplicate check": {
			tokens: InitializeTokens(
				BareKey("spelling"), Equal(), BasicString("favorite"),
				BasicString("spelling"), Equal(), BasicString("favourite"),
			),
			shouldErr:  true,
			errorCount: 1,
			errorCodes: []ParseErrorCode{ErrDuplicateKey},
		},
		"Valid but discouraged key": {
			tokens: InitializeTokens(BareKey("3"), Dot(), BareKey("14159"), Equal(), BasicString("pi")),
			expectedDocument: &Document{
				Content: []Node{
					&KeyValueNode{
						Key: &KeyNode{
							Segments: []string{"3", "14159"},
						},
						Value: &StringNode{
							Value: "pi",
						},
					},
				},
			},
		},
	}

	for test, params := range tests {
		t.Run(test, func(t *testing.T) {
			parser := NewParser(params.tokens)
			doc, errs := parser.parse()
			if params.shouldErr {
				if len(errs) <= 0 {
					t.Fatalf("expecting errors, found none")
				}

				if len(errs) != params.errorCount {
					t.Fatalf("expecting %d errors, found %d: %v", params.errorCount, len(errs), errs)
				}

				for _, code := range params.errorCodes {
					if !containsErrorCode(errs, code) {
						t.Fatalf("expected error code %v but was not found in %v", code, errs)
					}
				}
				return
			}

			if len(errs) != 0 {
				t.Fatalf("incorrect parse tree: %+v", parser.errors)
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

func assertNode(t *testing.T, expected, got Node) {
	t.Helper()

	if expected == nil {
		t.Fatalf("'Expected' Node is nil")
	}

	if got == nil {
		t.Fatalf("'Got' Node is nil")
	}

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

		assertComments(t, e.LeadingComments, g.LeadingComments, "leading")
		if e.TrailingComment != nil {
			if g.TrailingComment.Lexeme == "" {
				t.Fatalf("expected trailing comment %q, got \"\"", e.TrailingComment.Lexeme)
			}

			if e.TrailingComment.Lexeme != g.TrailingComment.Lexeme {
				t.Errorf("trailing comment: expected %q, got %q", e.TrailingComment.Lexeme, g.TrailingComment.Lexeme)
			}
		}
	case *KeyValueNode:
		g, ok := got.(*KeyValueNode)
		if !ok {
			t.Fatalf("expected *KeyValueNode, got %T", got)
		}
		assertKeyNode(t, e.Key, g.Key)
		assertNode(t, e.Value, g.Value)
		assertComments(t, e.LeadingComments, g.LeadingComments, "leading")
		if e.TrailingComment != nil {
			if g.TrailingComment == nil {
				t.Fatal("expected trailing comment, got nil")
			}

			if e.TrailingComment.Lexeme != g.TrailingComment.Lexeme {
				t.Errorf("trailing comment: expected %q, got %q", e.TrailingComment.Lexeme, g.TrailingComment.Lexeme)
			}
		}
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
	default:
		t.Fatalf("Unrecognized Node type %T", e)
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

func assertComments(t *testing.T, expected, got []Trivia, label string) {
	t.Helper()

	if expected == nil {
		return
	}

	if len(expected) != len(got) {
		t.Fatalf("%s: expected %d comments, got %d", label, len(expected), len(got))
	}

	for i := range len(expected) {
		if expected[i].Lexeme != got[i].Lexeme {
			t.Fatalf("%s comment %d: expected %q, got %q", label, i, expected[i], got[i])
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

type TokenOpt func() Token

func InitializeTokens(opts ...TokenOpt) []Token {
	var tokens []Token
	for _, opt := range opts {
		tokens = append(tokens, opt())
	}

	tokens = append(tokens, Eof()())
	return tokens
}

func Comment(comment string) TokenOpt {
	return func() Token {
		return Token{
			Type:    COMMENT,
			Literal: comment,
			Lexeme:  comment,
		}
	}
}

func LeftBracket() TokenOpt {
	return func() Token {
		return Token{
			Type:    LEFT_BRACKET,
			Literal: string("["),
			Lexeme:  "[",
		}
	}
}

func RightBracket() TokenOpt {
	return func() Token {
		return Token{
			Type:    RIGHT_BRACKET,
			Literal: string("]"),
			Lexeme:  "]",
		}
	}
}

func Comma() TokenOpt {
	return func() Token {
		return Token{
			Type:    COMMA,
			Literal: string(","),
			Lexeme:  ",",
		}
	}
}

func Dot() TokenOpt {
	return func() Token {
		return Token{
			Type:    DOT,
			Literal: string("."),
			Lexeme:  ".",
		}
	}
}

func Minus() TokenOpt {
	return func() Token {
		return Token{
			Type:    MINUS,
			Literal: string("-"),
			Lexeme:  "-",
		}
	}
}

func Plus() TokenOpt {
	return func() Token {
		return Token{
			Type:    PLUS,
			Literal: string("+"),
			Lexeme:  "+",
		}
	}
}

func Slash() TokenOpt {
	return func() Token {
		return Token{
			Type:    SLASH,
			Literal: string("\\"),
			Lexeme:  "\\",
		}
	}
}

func Star() TokenOpt {
	return func() Token {
		return Token{
			Type:    STAR,
			Literal: string("*"),
			Lexeme:  "*",
		}
	}
}

func Equal() TokenOpt {
	return func() Token {
		return Token{
			Type:    EQUAL,
			Literal: string("="),
			Lexeme:  "=",
		}
	}
}

func BasicString(value string) TokenOpt {
	return func() Token {
		return Token{
			Type:    BASIC_STRING,
			Literal: value,
			Lexeme:  fmt.Sprintf("\"%s\"", value),
		}
	}
}

func LiteralString(value string) TokenOpt {
	return func() Token {
		return Token{
			Type:    LITERAL_STRING,
			Literal: value,
			Lexeme:  fmt.Sprintf("'%s'", value),
		}
	}
}

func Float(value float64) TokenOpt {
	return func() Token {
		return Token{
			Type:    FLOAT,
			Literal: value,
			Lexeme:  fmt.Sprintf("%v", value),
		}
	}
}

func Integer(value int64) TokenOpt {
	return func() Token {
		return Token{
			Type:    INTEGER,
			Literal: value,
			Lexeme:  fmt.Sprintf("%v", value),
		}
	}
}

func BareKey(literal string) TokenOpt {
	return func() Token {
		return Token{
			Type:    BARE_KEY,
			Literal: string(literal),
			Lexeme:  literal,
		}
	}
}

func False() TokenOpt {
	return func() Token {
		return Token{
			Type:    FALSE,
			Literal: bool(false),
			Lexeme:  "false",
		}
	}
}

func True() TokenOpt {
	return func() Token {
		return Token{
			Type:    TRUE,
			Literal: bool(true),
			Lexeme:  "true",
		}
	}
}

func Inf() TokenOpt {
	return func() Token {
		return Token{
			Type:    INF,
			Literal: string("inf"),
			Lexeme:  "inf",
		}
	}
}

func Eof() TokenOpt {
	return func() Token {
		return Token{
			Type: EOF,
		}
	}
}
