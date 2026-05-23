package token

import "testing"

func Test_IntegerNode(t *testing.T) {
	tests := map[string]struct {
		source    string
		tokenType TokenType
		lexeme    string
		literal   any
		shouldErr bool
	}{
		"\"normal\" integer": {
			source:    `12345`,
			lexeme:    "12345",
			tokenType: INTEGER,
			literal:   int(12345),
		},
		"integer with underscores": {
			source:    `12_345`,
			lexeme:    "12_345",
			tokenType: INTEGER,
			literal:   int(12_345),
		},
		"edge case 1": {
			source:    `1_2_3_4_5`,
			lexeme:    `1_2_3_4_5`,
			tokenType: INTEGER,
			literal:   int(12345),
		},
		"edge case 2": {
			source:    `53_49_221`,
			lexeme:    `53_49_221`,
			tokenType: INTEGER,
			literal:   int(5349221),
		},
		"edge case 3": {
			source:    `5_349_221`,
			lexeme:    `5_349_221`,
			tokenType: INTEGER,
			literal:   int(5349221),
		},
		"regular floating point": {
			source:    "3.14",
			lexeme:    "3.14",
			tokenType: FLOAT,
			literal:   float64(3.14),
		},
		"floating point with underscores on integer": {
			source:    "1_341.890",
			lexeme:    "1_341.890",
			tokenType: FLOAT,
			literal:   float64(1341.890),
		},
	}

	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			s := Scanner{
				Source:  []byte(tt.source),
				start:   0,
				line:    0,
				current: 0,
			}

			s.Scan()

			if s.Tokens[0].Type != tt.tokenType {
				t.Fatalf("Incorrect token type: Expected %v. Got %v", INTEGER, s.Tokens[0].Type)
			}

			if s.Tokens[0].Literal != tt.literal {
				t.Fatalf("Incorrect literal value for token: Expected: %v. Got: %v", tt.literal, s.Tokens[0].Literal)
			}

			if s.Tokens[0].Lexeme != tt.lexeme {
				t.Fatalf("Incorrect lexeme value for token: Expected: %s. Got: %s", tt.lexeme, s.Tokens[0].Lexeme)
			}
		})
	}
}
