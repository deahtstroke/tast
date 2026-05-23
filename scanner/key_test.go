package scanner

import (
	"testing"
)

func Test_KeyNode(t *testing.T) {
	tests := map[string]struct {
		source    []byte
		literal   string
		tokenType TokenType
	}{
		"bare key": {
			source:    []byte("this_is_a_key = \"World!\""),
			literal:   "this_is_a_key",
			tokenType: BARE_KEY,
		},
		"bare key [no space between keys]": {
			source:    []byte("this_is_a_key=\"World!\""),
			literal:   "this_is_a_key",
			tokenType: BARE_KEY,
		},
		"bare key [space between key/value]": {
			source:    []byte("this_is_a_key = \"World!\""),
			literal:   "this_is_a_key",
			tokenType: BARE_KEY,
		},
	}

	for test, tt := range tests {
		t.Run(test, func(t *testing.T) {
			scanner := Scanner{
				Source:  tt.source,
				current: 0,
				start:   0,
				line:    0,
			}

			scanner.Scan()
			if scanner.Tokens[0].Type != tt.tokenType {
				t.Fatalf("Incorrect token type: Expected %v. Got %v", INTEGER, scanner.Tokens[0].Type)
			}

			if scanner.Tokens[0].Literal != tt.literal {
				t.Fatalf("Incorrect literal value for token: Expected: %v. Got: %v", tt.literal, scanner.Tokens[0].Literal)
			}
		})
	}
}

func Test_reservedKeys(t *testing.T) {
	tests := map[string]struct {
		source    []byte
		tokenType TokenType
	}{
		"false keyword": {
			source:    []byte("false"),
			tokenType: FALSE,
		},
		"true keyword": {
			source:    []byte("true"),
			tokenType: TRUE,
		},
		"nan keyword": {
			source:    []byte("nan"),
			tokenType: NAN,
		},
		"inf": {
			source:    []byte("inf"),
			tokenType: INF,
		},
	}

	for test, tt := range tests {
		t.Run(test, func(t *testing.T) {
			scanner := Scanner{
				Source:  tt.source,
				current: 0,
				start:   0,
				line:    0,
			}

			scanner.Scan()
			if scanner.Tokens[0].Type != tt.tokenType {
				t.Fatalf("Incorrect token type: Expected %v. Got %v", tt.tokenType, scanner.Tokens[0].Type)
			}
		})
	}
}
