package tast

import (
	"fmt"
	"math"
	"testing"
)

func Test_Parser(t *testing.T) {
	tests := map[string]struct {
		tokens      []token
		expectedDoc *Document
		wantErr     bool
		errCount    int
		errCodes    []ParseErrorCode
	}{
		// Comments
		"General comments with key-value pairs": {
			tokens: DeclareTokens(
				Comment("# This is a full-line comment"), NewLine(),
				BareKey("key"), Equal(), BasicString("value"), Comment("# This is a comment at the end of a line"), NewLine(),
				BareKey("another"), Equal(), BasicString("# This is not a comment"), NewLine(),
			),
			expectedDoc: &Document{
				content: []Node{
					&KeyValueNode{
						leadingTrivia: []Trivia{{lexeme: "# This is a full-line comment", Type: CommentTrivia}, {lexeme: "\n", Type: NewLineTrivia}},
						lineTrivia:    []Trivia{{lexeme: "# This is a comment at the end of a line", Type: CommentTrivia}, {lexeme: "\n", Type: NewLineTrivia}},
						key: &KeyNode{
							segments: []string{"key"},
						},
						value: &StringNode{
							value: "value",
						},
					},
					&KeyValueNode{
						key: &KeyNode{
							segments: []string{"another"},
						},
						value: &StringNode{
							value: "# This is not a comment",
						},
					},
				},
			},
		},
		"Should report error on missing assignment after key": {
			tokens: DeclareTokens(
				BareKey("key"), BasicString("value"),
			),
			wantErr:  true,
			errCount: 1,
			errCodes: []ParseErrorCode{ErrMissingAssignmentAfterKey},
		},
		"Leading comments belong to the respective nodes: Table + KVs": {
			tokens: DeclareTokens(
				Comment("# Trivia 1"), NewLine(),
				LeftBracket(), BareKey("key1"), RightBracket(), NewLine(),
				Comment("# Trivia 2"), NewLine(),
				BareKey("key2"), Equal(), Inf(1), NewLine(),
				Comment("# Trivia 3"), NewLine(),
				BareKey("key3"), Equal(), Float(3.14), NewLine(),
			),
			expectedDoc: &Document{
				content: []Node{
					&TableNode{
						leadingTrivia: []Trivia{{lexeme: "# Trivia 1"}},
						key: &KeyNode{
							segments: []string{"key1"},
						},
						children: []Node{
							&KeyValueNode{
								leadingTrivia: []Trivia{{lexeme: "# Trivia 2"}},
								key: &KeyNode{
									segments: []string{"key2"},
								},
								value: &FloatNode{
									value: math.Inf(1),
								},
							},
							&KeyValueNode{
								leadingTrivia: []Trivia{{lexeme: "# Trivia 3"}},
								key: &KeyNode{
									segments: []string{"key3"},
								},
								value: &FloatNode{
									value: 3.14,
								},
							},
						},
					},
				},
			},
		},
		"Accumulated >1 comments before a table": {
			tokens: DeclareTokens(
				Comment("# This is a comment"),
				Comment("# This is another comment"),
				LeftBracket(), BareKey("key"), RightBracket(), Comment("# This is in the same line as the table"),
			),
			expectedDoc: &Document{
				content: []Node{
					&TableNode{
						leadingTrivia: []Trivia{{lexeme: "# This is a comment"}, {lexeme: "# This is another comment"}},
						lineTrivia:    &Trivia{lexeme: "# This is in the same line as the table"},
						key: &KeyNode{
							segments: []string{"key"},
						},
					},
				},
			},
		},
		"Comments should belong to the respective nodes: Tables with no KVs": {
			tokens: DeclareTokens(
				Comment("# Comment 1"), NewLine(),
				LeftBracket(), BareKey("key"), RightBracket(), NewLine(),
				Comment("# Comment 2"), NewLine(),
				LeftBracket(), BareKey("key2"), RightBracket(), NewLine(),
			),
			expectedDoc: &Document{
				content: []Node{
					&TableNode{
						leadingTrivia: []Trivia{{lexeme: "# Comment 1"}},
						key: &KeyNode{
							segments: []string{"key"},
						},
					},
					&TableNode{
						leadingTrivia: []Trivia{{lexeme: "# Comment 2"}},
						key: &KeyNode{
							segments: []string{"key2"},
						},
					},
				},
			},
		},
		"Comments at the end of the document are assigned to last node": {
			tokens: DeclareTokens(
				Comment("# Comment 1"), NewLine(),
				LeftBracket(), BareKey("key"), RightBracket(), NewLine(),
				Comment("# Comment 2"), NewLine(),
			),
			expectedDoc: &Document{
				content: []Node{
					&TableNode{
						leadingTrivia:  []Trivia{{lexeme: "# Comment 1"}},
						trailingTrivia: []Trivia{{lexeme: "# Comment 2"}},
						key: &KeyNode{
							segments: []string{"key"},
						},
					},
				},
			},
		},
		// Key-value pair
		"Basic key-value pair": {
			tokens: DeclareTokens(
				BareKey("key"), Equal(), BasicString("value"),
			),
			expectedDoc: &Document{
				content: []Node{
					&KeyValueNode{
						key: &KeyNode{
							segments: []string{"key"},
						},
						value: &StringNode{
							value: "value",
						},
					},
				},
			},
		},
		"Should error on unspecified values": {
			tokens: DeclareTokens(
				BareKey("key"), Equal(),
			),
			wantErr:  true,
			errCount: 1,
			errCodes: []ParseErrorCode{ErrUnspecifiedValueForKey},
		},
		// Keys
		"BareKeys": {
			tokens: DeclareTokens(
				BareKey("key"), Equal(), BasicString("value"),
				BareKey("bare_key"), Equal(), BasicString("value"),
				BareKey("bare-key"), Equal(), BasicString("value"),
				BareKey("1234"), Equal(), BasicString("value"),
			),
			expectedDoc: &Document{
				content: []Node{
					&KeyValueNode{
						key: &KeyNode{
							segments: []string{"key"},
						},
						value: &StringNode{
							value: "value",
						},
					},
					&KeyValueNode{
						key: &KeyNode{
							segments: []string{"bare_key"},
						},
						value: &StringNode{
							value: "value",
						},
					},
					&KeyValueNode{
						key: &KeyNode{
							segments: []string{"bare-key"},
						},
						value: &StringNode{
							value: "value",
						},
					},
					&KeyValueNode{
						key: &KeyNode{
							segments: []string{"1234"},
						},
						value: &StringNode{
							value: "value",
						},
					},
				},
			},
		},
		"Quoted Keys": {
			tokens: DeclareTokens(
				BasicString("127.0.0.1"), Equal(), BasicString("value"),
				BasicString("character encoding"), Equal(), BasicString("value"),
				LiteralString("key2"), Equal(), BasicString("value"),
				LiteralString("ʎǝʞ"), Equal(), BasicString("value"),
				LiteralString("quoted \"value\""), Equal(), BasicString("value")),
			expectedDoc: &Document{
				content: []Node{
					&KeyValueNode{
						key: &KeyNode{
							segments: []string{"127.0.0.1"},
						},
						value: &StringNode{
							value: "value",
						},
					},
					&KeyValueNode{
						key: &KeyNode{
							segments: []string{"character encoding"},
						},
						value: &StringNode{
							value: "value",
						},
					},
					&KeyValueNode{
						key: &KeyNode{
							segments: []string{"key2"},
						},
						value: &StringNode{
							value: "value",
						},
					},
					&KeyValueNode{
						key: &KeyNode{
							segments: []string{"ʎǝʞ"},
						},
						value: &StringNode{
							value: "value",
						},
					},
					&KeyValueNode{
						key: &KeyNode{
							segments: []string{"quoted \"value\""},
						},
						value: &StringNode{
							value: "value",
						},
					},
				},
			},
		},
		"Dotted-keys": {
			tokens: DeclareTokens(
				BareKey("name"), Equal(), BasicString("Orange"),
				BareKey("physical"), Dot(), BasicString("color"), Equal(), BasicString("orange"),
				BareKey("physical"), Dot(), BasicString("shape"), Equal(), BasicString("round"),
				BareKey("site"), Dot(), BasicString("google.com"), Equal(), True(),
			),
			expectedDoc: &Document{
				content: []Node{
					&KeyValueNode{
						key: &KeyNode{
							segments: []string{"name"},
						},
						value: &StringNode{
							value: "Orange",
						},
					},
					&KeyValueNode{
						key: &KeyNode{
							segments: []string{"physical", "color"},
						},
						value: &StringNode{
							value: "orange",
						},
					},
					&KeyValueNode{
						key: &KeyNode{
							segments: []string{"physical", "shape"},
						},
						value: &StringNode{
							value: "round",
						},
					},
					&KeyValueNode{
						key: &KeyNode{
							segments: []string{"site", "google.com"},
						},
						value: &BooleanNode{
							value: true,
						},
					},
				},
			},
		},
		"Duplicate keys should error": {
			tokens: DeclareTokens(
				BareKey("name"), Equal(), BasicString("Tom"),
				BareKey("name"), Equal(), BasicString("Pradyun"),
			),
			wantErr:  true,
			errCount: 1,
			errCodes: []ParseErrorCode{ErrDuplicateKey},
		},
		"Barekeys and quotted keys are equivalent on duplicate check": {
			tokens: DeclareTokens(
				BareKey("spelling"), Equal(), BasicString("favorite"),
				BasicString("spelling"), Equal(), BasicString("favourite"),
			),
			wantErr:  true,
			errCount: 1,
			errCodes: []ParseErrorCode{ErrDuplicateKey},
		},
		"Valid but discouraged key": {
			tokens: DeclareTokens(BareKey("3"), Dot(), BareKey("14159"), Equal(), BasicString("pi")),
			expectedDoc: &Document{
				content: []Node{
					&KeyValueNode{
						key: &KeyNode{
							segments: []string{"3", "14159"},
						},
						value: &StringNode{
							value: "pi",
						},
					},
				},
			},
		},
		"Duplicate tables should error": {
			tokens: DeclareTokens(
				LeftBracket(), BareKey("key"), RightBracket(), NewLine(),
				LeftBracket(), BareKey("key"), RightBracket(),
			),
			wantErr:  true,
			errCount: 1,
			errCodes: []ParseErrorCode{ErrDuplicateTable},
		},
		"Table with duplicate KVs should error": {
			tokens: DeclareTokens(
				LeftBracket(), BareKey("key"), RightBracket(), NewLine(),
				BareKey("hello"), Equal(), BasicString("world"), NewLine(),
				BasicString("hello"), Equal(), BasicString("world?"),
			),
			wantErr:  true,
			errCount: 1,
			errCodes: []ParseErrorCode{ErrDuplicateKey},
		},
		"Table with duplicate dotted KVs should error": {
			tokens: DeclareTokens(
				LeftBracket(), BareKey("key"), RightBracket(), NewLine(),
				BareKey("foo"), Dot(), BareKey("bar"), Equal(), BasicString("world"), NewLine(),
				BasicString("foo"), Dot(), BasicString("bar"), Equal(), BasicString("world??"),
			),
			wantErr:  true,
			errCount: 1,
			errCodes: []ParseErrorCode{ErrDuplicateKey},
		},
	}

	for test, expected := range tests {
		t.Run(test, func(t *testing.T) {
			parser := NewParser(expected.tokens)
			got, errs := parser.parse()
			if expected.wantErr {
				if len(errs) <= 0 {
					t.Fatalf("expecting errors, found none")
				}

				if len(errs) != expected.errCount {
					t.Fatalf("expecting %d errors, found %d: %v", expected.errCount, len(errs), errs)
				}

				for _, code := range expected.errCodes {
					if !containsErrorCode(errs, code) {
						t.Fatalf("expected error code %v but was not found in %v", code, errs)
					}
				}
				return
			}

			if len(errs) != 0 {
				t.Fatalf("incorrect parse tree: %+v", parser.errors)
			}

			assertDocument(t, expected.expectedDoc, got)
		})
	}
}

func assertDocument(t *testing.T, expected, got *Document) {
	t.Helper()
	if expected == nil && got == nil {
		return
	}

	if len(expected.content) != len(got.content) {
		t.Fatalf("Expected %d nodes, got %d", len(expected.content), len(got.content))
	}

	for i := range len(got.content) {
		assertNode(t, expected.content[i], got.content[i])
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

		assertKeyNode(t, e.key, g.key)

		if len(e.children) != len(g.children) {
			t.Fatalf("expected %d children, got: %d", len(e.children), len(g.children))
		}

		for i := range len(g.children) {
			assertNode(t, e.children[i], g.children[i])
		}

		assertTrivia(t, e.leadingTrivia, g.leadingTrivia, "leading")
		assertTrivia(t, e.lineTrivia, g.lineTrivia, "inline")
		assertTrivia(t, e.trailingTrivia, g.trailingTrivia, "trailing")
	case *KeyValueNode:
		g, ok := got.(*KeyValueNode)
		if !ok {
			t.Fatalf("expected *KeyValueNode, got %T", got)
		}
		assertKeyNode(t, e.key, g.key)
		assertNode(t, e.value, g.value)

		assertTrivia(t, e.leadingTrivia, g.leadingTrivia, "leading")
		assertTrivia(t, e.lineTrivia, g.lineTrivia, "inline")
		assertTrivia(t, e.trailingTrivia, g.trailingTrivia, "trailing")
	case *StringNode:
		g, ok := got.(*StringNode)
		if !ok {
			t.Fatalf("expected *StringNode, got %T", got)
		}

		if g.value != e.value {
			t.Fatalf("expected value %s, got %s", e.value, g.value)
		}
	case *IntegerNode:
		g, ok := got.(*IntegerNode)
		if !ok {
			t.Fatalf("expected *IntegerNode, got %T", got)
		}

		if g.value != e.value {
			t.Fatalf("expected value %d, got %d", e.value, g.value)
		}
	case *FloatNode:
		g, ok := got.(*FloatNode)
		if !ok {
			t.Fatalf("expected *FloatNode, got %T", got)
		}

		if g.value != e.value {
			t.Fatalf("expected value %f, got %f", e.value, g.value)
		}
	case *BooleanNode:
		g, ok := got.(*BooleanNode)
		if !ok {
			t.Fatalf("expected *BooleanNode, got %T", got)
		}

		if g.value != e.value {
			t.Fatalf("expected value %t, got %t", e.value, g.value)
		}
	default:
		t.Fatalf("Unrecognized Node type %T", e)
	}
}

func assertKeyNode(t *testing.T, expected, got *KeyNode) {
	t.Helper()

	if len(expected.segments) != len(got.segments) {
		t.Fatalf("expected %d segments, got %d", len(expected.segments), len(got.segments))
	}

	for i := range len(got.segments) {
		if got.segments[i] != expected.segments[i] {
			t.Fatalf("expected segment %s at index %d to be equal, got: %s", expected.segments[i], i, got.segments[i])
		}
	}
}

func assertTrivia(t *testing.T, expected, got []Trivia, label string) {
	t.Helper()

	if expected == nil {
		return
	}

	if len(expected) != len(got) {
		t.Fatalf("%s: expected %d trivia, got %d", label, len(expected), len(got))
	}

	for i := range len(expected) {
		if expected[i].lexeme != got[i].lexeme {
			t.Fatalf("%s trivia lexeme %d: expected %s, got %s", label, i, expected[i].lexeme, got[i].lexeme)
		}

		if expected[i].Type != got[i].Type {
			t.Fatalf("%s trivia type %d: expected %s, got %s", label, i, TriviaTypes[expected[i].Type], TriviaTypes[got[i].Type])
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

type TokenOpt func() token

func DeclareTokens(opts ...TokenOpt) []token {
	var tokens []token
	for _, opt := range opts {
		tokens = append(tokens, opt())
	}

	tokens = append(tokens, Eof()())
	return tokens
}

func Comment(comment string) TokenOpt {
	return func() token {
		return token{
			Type:    COMMENT,
			Literal: comment,
			Lexeme:  comment,
		}
	}
}

func LeftBracket() TokenOpt {
	return func() token {
		return token{
			Type:    LEFT_BRACKET,
			Literal: string("["),
			Lexeme:  "[",
		}
	}
}

func RightBracket() TokenOpt {
	return func() token {
		return token{
			Type:    RIGHT_BRACKET,
			Literal: string("]"),
			Lexeme:  "]",
		}
	}
}

func Comma() TokenOpt {
	return func() token {
		return token{
			Type:    COMMA,
			Literal: string(","),
			Lexeme:  ",",
		}
	}
}

func Dot() TokenOpt {
	return func() token {
		return token{
			Type:    DOT,
			Literal: string("."),
			Lexeme:  ".",
		}
	}
}

func Minus() TokenOpt {
	return func() token {
		return token{
			Type:    MINUS,
			Literal: string("-"),
			Lexeme:  "-",
		}
	}
}

func Plus() TokenOpt {
	return func() token {
		return token{
			Type:    PLUS,
			Literal: string("+"),
			Lexeme:  "+",
		}
	}
}

func Slash() TokenOpt {
	return func() token {
		return token{
			Type:    SLASH,
			Literal: string("\\"),
			Lexeme:  "\\",
		}
	}
}

func Star() TokenOpt {
	return func() token {
		return token{
			Type:    STAR,
			Literal: string("*"),
			Lexeme:  "*",
		}
	}
}

func Equal() TokenOpt {
	return func() token {
		return token{
			Type:    EQUAL,
			Literal: string("="),
			Lexeme:  "=",
		}
	}
}

func BasicString(value string) TokenOpt {
	return func() token {
		return token{
			Type:    BASIC_STRING,
			Literal: value,
			Lexeme:  fmt.Sprintf("\"%s\"", value),
		}
	}
}

func LiteralString(value string) TokenOpt {
	return func() token {
		return token{
			Type:    LITERAL_STRING,
			Literal: value,
			Lexeme:  fmt.Sprintf("'%s'", value),
		}
	}
}

func Float(value float64) TokenOpt {
	return func() token {
		return token{
			Type:    FLOAT,
			Literal: value,
			Lexeme:  fmt.Sprintf("%v", value),
		}
	}
}

func Integer(value int64) TokenOpt {
	return func() token {
		return token{
			Type:    INTEGER,
			Literal: value,
			Lexeme:  fmt.Sprintf("%v", value),
		}
	}
}

func BareKey(literal string) TokenOpt {
	return func() token {
		return token{
			Type:    BARE_KEY,
			Literal: string(literal),
			Lexeme:  literal,
		}
	}
}

func False() TokenOpt {
	return func() token {
		return token{
			Type:    FALSE,
			Literal: bool(false),
			Lexeme:  "false",
		}
	}
}

func True() TokenOpt {
	return func() token {
		return token{
			Type:    TRUE,
			Literal: bool(true),
			Lexeme:  "true",
		}
	}
}

func Inf(op int) TokenOpt {
	return func() token {
		var literal string = "inf"
		if op < 0 {
			literal = "-inf"
		}
		return token{
			Type:    INF,
			Literal: literal,
			Lexeme:  "inf",
		}
	}
}

func Eof() TokenOpt {
	return func() token {
		return token{
			Type: EOF,
		}
	}
}

func NewLine() TokenOpt {
	return func() token {
		return token{
			Type:    NEW_LINE,
			Lexeme:  "\n",
			Literal: string("\n"),
		}
	}
}
