package scanner

import "testing"

func Test_BasicStringNode(t *testing.T) {
	tests := map[string]struct {
		source    string
		text      string
		tokenType TokenType
		shouldErr bool
	}{
		"normal string no escape characters": {
			source:    `"Hello world!"`,
			tokenType: BASIC_STRING,
			text:      "Hello world!",
		},
		"string with escaped quotes": {
			source:    "\"Hello world!\"",
			tokenType: BASIC_STRING,
			text:      "Hello world!",
		},
		"multi-line string (should trim the first newline)": {
			source:    "\"\"\"\nHello my name is\nDaniel!\n\"\"\"",
			tokenType: MULTILINE_BASIC_STRING,
			text:      "Hello my name is\nDaniel!\n",
		},
		"multi-line string (just for Go)": {
			source: `"""Hello World!
My name is.
"""`,
			tokenType: MULTILINE_BASIC_STRING,
			text:      "Hello World!\nMy name is.\n",
		},
	}

	for test, tt := range tests {
		t.Run(test, func(t *testing.T) {
			s := Scanner{
				Source:  []byte(tt.source),
				start:   0,
				line:    0,
				current: 0,
			}

			s.Scan()

			if s.Tokens[0].Type != tt.tokenType {
				t.Fatalf("Incorrect token type: Expected String %v. Got %v", tt.tokenType, s.Tokens[0].Type)
			}

			if s.Tokens[0].Literal != tt.text {
				t.Fatalf("Incorrect literal value for token: Expected: %s. Got: %v", tt.text, s.Tokens[0].Literal)
			}
		})
	}
}
