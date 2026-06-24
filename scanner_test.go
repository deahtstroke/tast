package tast

import (
	"math"
	"testing"
)

func Test_Scan(t *testing.T) {
	tests := map[string]struct {
		sourceBytes    []byte
		expectedTokens []token
		wantErr        bool
	}{
		"simple key value": {
			sourceBytes: []byte(`foo = "bar"`),
			expectedTokens: []token{
				{Type: bareKey, Lexeme: "foo", Literal: "foo", Line: 1, Column: 0},
				{Type: equal, Lexeme: "=", Literal: "=", Line: 1, Column: 5},
				{Type: basicString, Lexeme: `"bar"`, Literal: "bar", Line: 1, Column: 7},
				{Type: eof},
			},
		},
		"simple key value with integer": {
			sourceBytes: []byte(`foo = +23`),
			expectedTokens: []token{
				{Type: bareKey, Lexeme: "foo", Literal: "foo", Line: 1, Column: 0},
				{Type: equal, Lexeme: "=", Literal: "=", Line: 1, Column: 5},
				{Type: plus, Lexeme: "+", Literal: "+", Line: 1, Column: 6},
				{Type: integer, Lexeme: "23", Literal: int64(23), Line: 1, Column: 7},
				{Type: eof},
			},
		},
		"simple key value with floating point": {
			sourceBytes: []byte(`foo = 5_123.12`),
			expectedTokens: []token{
				{Type: bareKey, Lexeme: "foo", Literal: "foo", Line: 1, Column: 0},
				{Type: equal, Lexeme: "=", Literal: "=", Line: 1, Column: 5},
				{Type: floatPoint, Lexeme: "5_123.12", Literal: float64(5123.12), Line: 1, Column: 7},
				{Type: eof},
			},
		},
		"simple key value with infinity": {
			sourceBytes: []byte(`foo = inf`),
			expectedTokens: []token{
				{Type: bareKey, Lexeme: "foo", Literal: "foo", Line: 1, Column: 0},
				{Type: equal, Lexeme: "=", Literal: "=", Line: 1, Column: 5},
				{Type: infinity, Lexeme: "inf", Literal: float64(math.Inf(1)), Line: 1, Column: 7},
				{Type: eof},
			},
		},
		"simple key value with Nan": {
			sourceBytes: []byte(`foo = nan`),
			expectedTokens: []token{
				{Type: bareKey, Lexeme: "foo", Literal: "foo", Line: 1, Column: 0},
				{Type: equal, Lexeme: "=", Literal: "=", Line: 1, Column: 5},
				{Type: nan, Lexeme: "nan", Literal: float64(math.NaN()), Line: 1, Column: 7},
				{Type: eof},
			},
		},
		"unterminated string should error": {
			sourceBytes: []byte(`foo = "bar`),
			wantErr:     true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			scanner := newScanner(tt.sourceBytes)
			tokens, err := scanner.scan()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expecting error, got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected scanner error: %v", err)
			}

			if len(tokens) != len(tt.expectedTokens) {
				t.Fatalf("expected %d tokens, got %d: %v", len(tt.expectedTokens), len(tokens), tokens)
			}

			for i, want := range tt.expectedTokens {
				got := tokens[i]
				if got.Type != want.Type {
					t.Errorf("token %d: expected type %v, got %v", i, want.Type, got.Type)
				}

				if want.Lexeme != "" && got.Lexeme != want.Lexeme {
					t.Errorf("token %d: expected lexeme %v, got %v", i, want.Lexeme, got.Lexeme)
				}

				if want.Literal != nil && got.Literal != want.Literal {
					// Special case for NaN since NaN is unequal to everything, even itself
					if want.Type == nan {
						f, ok := got.Literal.(float64)
						if !ok {
							t.Fatalf("Did not get Float64 for NaN value")
						}

						if !math.IsNaN(f) {
							t.Fatalf("Expecting NaN literal, got: %v", f)
						}

						return
					}
					t.Errorf("token %d: expected literal %v, got %v", i, want.Literal, got.Literal)
				}
			}
		})
	}
}

func Test_IntegerNode(t *testing.T) {
	tests := map[string]struct {
		source    string
		tokenType tokenType
		lexeme    string
		literal   any
		shouldErr bool
	}{
		"\"normal\" integer": {
			source:    `12345`,
			lexeme:    "12345",
			tokenType: integer,
			literal:   int64(12345),
		},
		"integer with underscores": {
			source:    `12_345`,
			lexeme:    "12_345",
			tokenType: integer,
			literal:   int64(12_345),
		},
		"edge case 1": {
			source:    `1_2_3_4_5`,
			lexeme:    `1_2_3_4_5`,
			tokenType: integer,
			literal:   int64(12345),
		},
		"edge case 2": {
			source:    `53_49_221`,
			lexeme:    `53_49_221`,
			tokenType: integer,
			literal:   int64(5349221),
		},
		"edge case 3": {
			source:    `5_349_221`,
			lexeme:    `5_349_221`,
			tokenType: integer,
			literal:   int64(5349221),
		},
		"regular floating point": {
			source:    "3.14",
			lexeme:    "3.14",
			tokenType: floatPoint,
			literal:   float64(3.14),
		},
		"floating point with underscores on integer": {
			source:    "1_341.890",
			lexeme:    "1_341.890",
			tokenType: floatPoint,
			literal:   float64(1341.890),
		},
	}

	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			s := scanner{
				source:  []byte(tt.source),
				start:   0,
				line:    0,
				current: 0,
			}

			tokens, err := s.scan()
			if err != nil {
				t.Fatalf("Not expecting error, got: %v", err)
			}

			if tokens[0].Type != tt.tokenType {
				t.Fatalf("Incorrect token type: Expected %v. Got %v", integer, tokens[0].Type)
			}

			if tokens[0].Literal != tt.literal {
				t.Fatalf("Incorrect literal value for token: Expected: %v. Got: %v", tt.literal, tokens[0].Literal)
			}

			if tokens[0].Lexeme != tt.lexeme {
				t.Fatalf("Incorrect lexeme value for token: Expected: %s. Got: %s", tt.lexeme, tokens[0].Lexeme)
			}
		})
	}
}

func Test_KeyNode(t *testing.T) {
	tests := map[string]struct {
		source    []byte
		literal   string
		tokenType tokenType
	}{
		"bare key": {
			source:    []byte("this_is_a_key = \"World!\""),
			literal:   "this_is_a_key",
			tokenType: bareKey,
		},
		"bare key [no space between keys]": {
			source:    []byte("this_is_a_key=\"World!\""),
			literal:   "this_is_a_key",
			tokenType: bareKey,
		},
		"bare key [space between key/value]": {
			source:    []byte("this_is_a_key = \"World!\""),
			literal:   "this_is_a_key",
			tokenType: bareKey,
		},
	}

	for test, tt := range tests {
		t.Run(test, func(t *testing.T) {
			scanner := scanner{
				source:  tt.source,
				current: 0,
				start:   0,
				line:    0,
			}

			tokens, err := scanner.scan()
			if err != nil {
				t.Fatalf("Not expecting error, got: %v", err)
			}
			if tokens[0].Type != tt.tokenType {
				t.Fatalf("Incorrect token type: Expected %v. Got %v", integer, tokens[0].Type)
			}

			if tokens[0].Literal != tt.literal {
				t.Fatalf("Incorrect literal value for token: Expected: %v. Got: %v", tt.literal, tokens[0].Literal)
			}
		})
	}
}

func Test_ReservedKeys(t *testing.T) {
	tests := map[string]struct {
		source    []byte
		tokenType tokenType
	}{
		"false keyword": {
			source:    []byte("false"),
			tokenType: boolean,
		},
		"true keyword": {
			source:    []byte("true"),
			tokenType: boolean,
		},
		"nan keyword": {
			source:    []byte("nan"),
			tokenType: nan,
		},
		"inf": {
			source:    []byte("inf"),
			tokenType: infinity,
		},
	}

	for test, tt := range tests {
		t.Run(test, func(t *testing.T) {
			scanner := scanner{
				source:  tt.source,
				current: 0,
				start:   0,
				line:    0,
			}

			tokens, err := scanner.scan()
			if err != nil {
				t.Fatalf("Not expecting error, got: %v", err)
			}

			if tokens[0].Type != tt.tokenType {
				t.Fatalf("Incorrect token type: Expected %v. Got %v", tt.tokenType, tokens[0].Type)
			}
		})
	}
}

func Test_BasicStringNode(t *testing.T) {
	tests := map[string]struct {
		source    string
		text      string
		tokenType tokenType
		shouldErr bool
	}{
		"normal string no escape characters": {
			source:    `"Hello world!"`,
			tokenType: basicString,
			text:      "Hello world!",
		},
		"string with escaped quotes": {
			source:    "\"Hello world!\"",
			tokenType: basicString,
			text:      "Hello world!",
		},
		"multi-line string (should trim the first newline)": {
			source:    "\"\"\"\nHello my name is\nDaniel!\n\"\"\"",
			tokenType: multilineBasicString,
			text:      "Hello my name is\nDaniel!\n",
		},
		"multi-line string (just for Go)": {
			source: `"""Hello World!
My name is.
"""`,
			tokenType: multilineBasicString,
			text:      "Hello World!\nMy name is.\n",
		},
	}

	for test, tt := range tests {
		t.Run(test, func(t *testing.T) {
			s := scanner{
				source:  []byte(tt.source),
				start:   0,
				line:    0,
				current: 0,
			}

			tokens, err := s.scan()
			if err != nil {
				t.Fatalf("Not expecting error, got: %v", err)
			}

			if tokens[0].Type != tt.tokenType {
				t.Fatalf("Incorrect token type: Expected String %v. Got %v", tt.tokenType, tokens[0].Type)
			}

			if tokens[0].Literal != tt.text {
				t.Fatalf("Incorrect literal value for token: Expected: %s. Got: %v", tt.text, tokens[0].Literal)
			}
		})
	}
}
