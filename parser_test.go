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
		errCodes    []parserErrorCode
	}{
		// Comments
		"General comments with key-value pairs": {
			tokens: DeclareTokens(
				Comment("# This is a full-line comment"), NewLine(),
				BareKey("key"), Equal(), BasicString("value"), Comment("# This is a comment at the end of a line"), NewLine(),
				BareKey("another"), Equal(), BasicString("# This is not a comment"), NewLine(),
			),
			expectedDoc: &Document{
				content: []node{
					&KeyValueNode{
						leadingTrivia: []trivia{{Lexeme: "# This is a full-line comment", Type: commentTrivia}, {Lexeme: "\n", Type: newLineTrivia}},
						lineTrivia:    []trivia{{Lexeme: "# This is a comment at the end of a line", Type: commentTrivia}, {Lexeme: "\n", Type: newLineTrivia}},
						key: &keyNode{
							segments: []string{"key"},
						},
						value: &stringNode{
							value: "value",
						},
					},
					&KeyValueNode{
						key: &keyNode{
							segments: []string{"another"},
						},
						value: &stringNode{
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
			errCodes: []parserErrorCode{errMissingAssignmentAfterKey},
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
				content: []node{
					&TableNode{
						leadingTrivia: []trivia{{Lexeme: "# Trivia 1", Type: commentTrivia}, {Lexeme: "\n", Type: newLineTrivia}},
						key: &keyNode{
							segments: []string{"key1"},
						},
						children: []node{
							&KeyValueNode{
								leadingTrivia: []trivia{{Lexeme: "# Trivia 2", Type: commentTrivia}, {Lexeme: "\n", Type: newLineTrivia}},
								key: &keyNode{
									segments: []string{"key2"},
								},
								value: &floatNode{
									value: math.Inf(1),
								},
							},
							&KeyValueNode{
								leadingTrivia: []trivia{{Lexeme: "# Trivia 3", Type: commentTrivia}, {Lexeme: "\n", Type: newLineTrivia}},
								key: &keyNode{
									segments: []string{"key3"},
								},
								value: &floatNode{
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
				Comment("# This is a comment"), NewLine(),
				Comment("# This is another comment"), NewLine(),
				LeftBracket(), BareKey("key"), RightBracket(), Comment("# This is in the same line as the table"),
			),
			expectedDoc: &Document{
				content: []node{
					&TableNode{
						leadingTrivia: []trivia{
							{Lexeme: "# This is a comment", Type: commentTrivia},
							{Lexeme: "\n", Type: newLineTrivia},
							{Lexeme: "# This is another comment", Type: commentTrivia},
							{Lexeme: "\n", Type: newLineTrivia}},
						lineTrivia: []trivia{
							{Lexeme: "# This is in the same line as the table", Type: commentTrivia},
						},
						key: &keyNode{
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
				content: []node{
					&TableNode{
						leadingTrivia: []trivia{
							{Lexeme: "# Comment 1", Type: commentTrivia},
							{Lexeme: "\n", Type: newLineTrivia},
						},
						key: &keyNode{
							segments: []string{"key"},
						},
					},
					&TableNode{
						leadingTrivia: []trivia{
							{Lexeme: "# Comment 2", Type: commentTrivia},
							{Lexeme: "\n", Type: newLineTrivia},
						},
						key: &keyNode{
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
				Comment("# Comment 2"),
			),
			expectedDoc: &Document{
				content: []node{
					&TableNode{
						leadingTrivia: []trivia{
							{Lexeme: "# Comment 1", Type: commentTrivia},
							{Lexeme: "\n", Type: newLineTrivia},
						},
						trailingTrivia: []trivia{{Lexeme: "# Comment 2", Type: commentTrivia}},
						key: &keyNode{
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
				content: []node{
					&KeyValueNode{
						key: &keyNode{
							segments: []string{"key"},
						},
						value: &stringNode{
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
			errCodes: []parserErrorCode{errUnspecifiedValueForKey},
		},
		// Keys
		"BareKeys": {
			tokens: DeclareTokens(
				BareKey("key"), Equal(), BasicString("value"), NewLine(),
				BareKey("bare_key"), Equal(), BasicString("value"), NewLine(),
				BareKey("bare-key"), Equal(), BasicString("value"), NewLine(),
				BareKey("1234"), Equal(), BasicString("value"),
			),
			expectedDoc: &Document{
				content: []node{
					&KeyValueNode{
						key: &keyNode{
							segments: []string{"key"},
						},
						value: &stringNode{
							value: "value",
						},
					},
					&KeyValueNode{
						key: &keyNode{
							segments: []string{"bare_key"},
						},
						value: &stringNode{
							value: "value",
						},
					},
					&KeyValueNode{
						key: &keyNode{
							segments: []string{"bare-key"},
						},
						value: &stringNode{
							value: "value",
						},
					},
					&KeyValueNode{
						key: &keyNode{
							segments: []string{"1234"},
						},
						value: &stringNode{
							value: "value",
						},
					},
				},
			},
		},
		"Quoted Keys": {
			tokens: DeclareTokens(
				BasicString("127.0.0.1"), Equal(), BasicString("value"), NewLine(),
				BasicString("character encoding"), Equal(), BasicString("value"), NewLine(),
				LiteralString("key2"), Equal(), BasicString("value"), NewLine(),
				LiteralString("ʎǝʞ"), Equal(), BasicString("value"), NewLine(),
				LiteralString("quoted \"value\""), Equal(), BasicString("value")),
			expectedDoc: &Document{
				content: []node{
					&KeyValueNode{
						key: &keyNode{
							segments: []string{"127.0.0.1"},
						},
						value: &stringNode{
							value: "value",
						},
					},
					&KeyValueNode{
						key: &keyNode{
							segments: []string{"character encoding"},
						},
						value: &stringNode{
							value: "value",
						},
					},
					&KeyValueNode{
						key: &keyNode{
							segments: []string{"key2"},
						},
						value: &stringNode{
							value: "value",
						},
					},
					&KeyValueNode{
						key: &keyNode{
							segments: []string{"ʎǝʞ"},
						},
						value: &stringNode{
							value: "value",
						},
					},
					&KeyValueNode{
						key: &keyNode{
							segments: []string{"quoted \"value\""},
						},
						value: &stringNode{
							value: "value",
						},
					},
				},
			},
		},
		"Dotted-keys": {
			tokens: DeclareTokens(
				BareKey("name"), Equal(), BasicString("Orange"), NewLine(),
				BareKey("physical"), Dot(), BasicString("color"), Equal(), BasicString("orange"), NewLine(),
				BareKey("site"), Dot(), BasicString("google.com"), Equal(), True(),
			),
			expectedDoc: &Document{
				content: []node{
					&KeyValueNode{
						key: &keyNode{
							segments: []string{"name"},
						},
						value: &stringNode{
							value: "Orange",
						},
					},
					&TableNode{
						key: &keyNode{
							segments: []string{"physical"},
						},
						children: []node{
							&KeyValueNode{
								key: &keyNode{
									segments: []string{"color"},
								},
								value: &stringNode{
									value: "orange",
								},
							},
						},
					},
					&TableNode{
						key: &keyNode{
							segments: []string{"site"},
						},
						children: []node{
							&KeyValueNode{
								key: &keyNode{
									segments: []string{"google.com"},
								},
								value: &booleanNode{
									value: true,
								},
							},
						},
					},
				},
			},
		},
		"Duplicate keys should error": {
			tokens: DeclareTokens(
				BareKey("name"), Equal(), BasicString("Tom"), NewLine(),
				BareKey("name"), Equal(), BasicString("Pradyun"),
			),
			wantErr:  true,
			errCount: 1,
			errCodes: []parserErrorCode{errDuplicateKey},
		},
		"Barekeys and quotted keys are equivalent on duplicate check": {
			tokens: DeclareTokens(
				BareKey("spelling"), Equal(), BasicString("favorite"), NewLine(),
				BasicString("spelling"), Equal(), BasicString("favourite"),
			),
			wantErr:  true,
			errCount: 1,
			errCodes: []parserErrorCode{errDuplicateKey},
		},
		"Valid but discouraged key": {
			tokens: DeclareTokens(BareKey("3"), Dot(), BareKey("14159"), Equal(), BasicString("pi")),
			expectedDoc: &Document{
				content: []node{
					&KeyValueNode{
						key: &keyNode{
							segments: []string{"3", "14159"},
						},
						value: &stringNode{
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
			errCodes: []parserErrorCode{errDuplicateTable},
		},
		"Table with duplicate KVs should error": {
			tokens: DeclareTokens(
				LeftBracket(), BareKey("key"), RightBracket(), NewLine(),
				BareKey("hello"), Equal(), BasicString("world"), NewLine(),
				BasicString("hello"), Equal(), BasicString("world?"),
			),
			wantErr:  true,
			errCount: 1,
			errCodes: []parserErrorCode{errDuplicateKey},
		},
		"Table with duplicate dotted KVs should error": {
			tokens: DeclareTokens(
				LeftBracket(), BareKey("key"), RightBracket(), NewLine(),
				BareKey("foo"), Dot(), BareKey("bar"), Equal(), BasicString("world"), NewLine(),
				BasicString("foo"), Dot(), BasicString("bar"), Equal(), BasicString("world??"),
			),
			wantErr:  true,
			errCount: 1,
			errCodes: []parserErrorCode{errDuplicateKey},
		},
		"Missing NewLine after Key-Value should error": {
			tokens: DeclareTokens(
				BareKey("key"), Equal(), Integer(123123),
				BareKey("key"), Equal(), Integer(123123),
			),
			wantErr:  true,
			errCount: 1,
			errCodes: []parserErrorCode{errMissingNewLine},
		},
		"Missing NewLine after Key-Value should error #2": {
			tokens: DeclareTokens(
				BareKey("key"), Equal(), Integer(123123), Comment("# This is a comment with no following newline"),
				BareKey("key"), Equal(), Integer(123123),
			),
			wantErr:  true,
			errCount: 1,
			errCodes: []parserErrorCode{errMissingNewLine},
		},
		"Missing NewLine after table should error": {
			tokens: DeclareTokens(
				LeftBracket(), BareKey("key"), RightBracket(),
				BareKey("foo"), Dot(), BareKey("bar"), Equal(), BasicString("world"), NewLine(),
			),
			wantErr:  true,
			errCount: 1,
			errCodes: []parserErrorCode{errMissingNewLine},
		},
		"Missing NewLine after table should error #2": {
			tokens: DeclareTokens(
				LeftBracket(), BareKey("key"), RightBracket(), Comment("# This is a comment with no following newline"),
				BareKey("foo"), Dot(), BareKey("bar"), Equal(), BasicString("world"), NewLine(),
			),
			wantErr:  true,
			errCount: 1,
			errCodes: []parserErrorCode{errMissingNewLine},
		},
	}

	for test, expected := range tests {
		t.Run(test, func(t *testing.T) {
			parser := newParser(expected.tokens)
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

func assertNode(t *testing.T, expected, got node) {
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
	case *stringNode:
		g, ok := got.(*stringNode)
		if !ok {
			t.Fatalf("expected *StringNode, got %T", got)
		}

		if g.value != e.value {
			t.Fatalf("expected value %s, got %s", e.value, g.value)
		}
	case *integerNode:
		g, ok := got.(*integerNode)
		if !ok {
			t.Fatalf("expected *IntegerNode, got %T", got)
		}

		if g.value != e.value {
			t.Fatalf("expected value %d, got %d", e.value, g.value)
		}
	case *floatNode:
		g, ok := got.(*floatNode)
		if !ok {
			t.Fatalf("expected *FloatNode, got %T", got)
		}

		if g.value != e.value {
			t.Fatalf("expected value %f, got %f", e.value, g.value)
		}
	case *booleanNode:
		g, ok := got.(*booleanNode)
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

func assertKeyNode(t *testing.T, expected, got *keyNode) {
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

func assertTrivia(t *testing.T, expected, got []trivia, label string) {
	t.Helper()

	if expected == nil {
		return
	}

	if len(expected) != len(got) {
		t.Fatalf("%s: expected %d trivia, got %d", label, len(expected), len(got))
	}

	for i := range len(expected) {
		if expected[i].Lexeme != got[i].Lexeme {
			t.Fatalf("%s trivia lexeme %d: expected %s, got %s", label, i, expected[i].Lexeme, got[i].Lexeme)
		}

		if expected[i].Type != got[i].Type {
			t.Fatalf("%s trivia type %d: expected %s, got %s", label, i, triviaTypes[expected[i].Type], triviaTypes[got[i].Type])
		}
	}
}

func containsErrorCode(errs []parseError, code parserErrorCode) bool {
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

func Comment(c string) TokenOpt {
	return func() token {
		return token{
			Type:    comment,
			Literal: c,
			Lexeme:  c,
		}
	}
}

func LeftBracket() TokenOpt {
	return func() token {
		return token{
			Type:    leftBracket,
			Literal: string("["),
			Lexeme:  "[",
		}
	}
}

func RightBracket() TokenOpt {
	return func() token {
		return token{
			Type:    rightBracket,
			Literal: string("]"),
			Lexeme:  "]",
		}
	}
}

func Comma() TokenOpt {
	return func() token {
		return token{
			Type:    comma,
			Literal: string(","),
			Lexeme:  ",",
		}
	}
}

func Dot() TokenOpt {
	return func() token {
		return token{
			Type:    dot,
			Literal: string("."),
			Lexeme:  ".",
		}
	}
}

func Minus() TokenOpt {
	return func() token {
		return token{
			Type:    minus,
			Literal: string("-"),
			Lexeme:  "-",
		}
	}
}

func Plus() TokenOpt {
	return func() token {
		return token{
			Type:    plus,
			Literal: string("+"),
			Lexeme:  "+",
		}
	}
}

func Slash() TokenOpt {
	return func() token {
		return token{
			Type:    slash,
			Literal: string("\\"),
			Lexeme:  "\\",
		}
	}
}

func Star() TokenOpt {
	return func() token {
		return token{
			Type:    star,
			Literal: string("*"),
			Lexeme:  "*",
		}
	}
}

func Equal() TokenOpt {
	return func() token {
		return token{
			Type:    equal,
			Literal: string("="),
			Lexeme:  "=",
		}
	}
}

func BasicString(value string) TokenOpt {
	return func() token {
		return token{
			Type:    basicString,
			Literal: value,
			Lexeme:  fmt.Sprintf("\"%s\"", value),
		}
	}
}

func LiteralString(value string) TokenOpt {
	return func() token {
		return token{
			Type:    literalString,
			Literal: value,
			Lexeme:  fmt.Sprintf("'%s'", value),
		}
	}
}

func Float(value float64) TokenOpt {
	return func() token {
		return token{
			Type:    floatPoint,
			Literal: value,
			Lexeme:  fmt.Sprintf("%v", value),
		}
	}
}

func Integer(value int64) TokenOpt {
	return func() token {
		return token{
			Type:    integer,
			Literal: value,
			Lexeme:  fmt.Sprintf("%v", value),
		}
	}
}

func BareKey(literal string) TokenOpt {
	return func() token {
		return token{
			Type:    bareKey,
			Literal: string(literal),
			Lexeme:  literal,
		}
	}
}

func False() TokenOpt {
	return func() token {
		return token{
			Type:    boolean,
			Literal: bool(false),
			Lexeme:  "false",
		}
	}
}

func True() TokenOpt {
	return func() token {
		return token{
			Type:    boolean,
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
			Type:    infinity,
			Literal: literal,
			Lexeme:  "inf",
		}
	}
}

func Eof() TokenOpt {
	return func() token {
		return token{
			Type: eof,
		}
	}
}

func NewLine() TokenOpt {
	return func() token {
		return token{
			Type:    newLine,
			Lexeme:  "\n",
			Literal: string("\n"),
		}
	}
}
