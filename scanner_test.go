package tast

import (
	"math"
	"reflect"
	"testing"
	"time"

	"gotest.tools/v3/assert"
)

func Test_Scan_Values(t *testing.T) {
	tests := map[string]struct {
		sourceBytes    []byte
		expectedTokens []token
		wantErr        bool
	}{
		"simple key value": {
			sourceBytes: []byte(`foo = "bar"`),
			expectedTokens: []token{
				{Type: bareKey, Lexeme: "foo", Literal: "foo", Line: 0, Column: 3},
				{Type: equal, Lexeme: "=", Literal: "=", Line: 0, Column: 5},
				{Type: basicString, Lexeme: `"bar"`, Literal: "bar", Line: 0, Column: 11},
				{Type: eof},
			},
		},
		"simple key value with integer": {
			sourceBytes: []byte(`foo = +23`),
			expectedTokens: []token{
				{Type: bareKey, Lexeme: "foo", Literal: "foo", Line: 0, Column: 3},
				{Type: equal, Lexeme: "=", Literal: "=", Line: 0, Column: 5},
				{Type: plus, Lexeme: "+", Literal: "+", Line: 0, Column: 7},
				{Type: integer, Lexeme: "23", Literal: int64(23), Line: 0, Column: 9},
				{Type: eof},
			},
		},
		"simple key value with floating point": {
			sourceBytes: []byte(`foo = 5_123.12`),
			expectedTokens: []token{
				{Type: bareKey, Lexeme: "foo", Literal: "foo", Line: 0, Column: 3},
				{Type: equal, Lexeme: "=", Literal: "=", Line: 0, Column: 5},
				{Type: floatPoint, Lexeme: "5_123.12", Literal: float64(5123.12), Line: 0, Column: 14},
				{Type: eof},
			},
		},
		"simple key value with infinity": {
			sourceBytes: []byte(`foo = inf`),
			expectedTokens: []token{
				{Type: bareKey, Lexeme: "foo", Literal: "foo", Line: 0, Column: 3},
				{Type: equal, Lexeme: "=", Literal: "=", Line: 0, Column: 5},
				{Type: infinity, Lexeme: "inf", Literal: float64(math.Inf(1)), Line: 0, Column: 7},
				{Type: eof},
			},
		},
		"simple key value with Nan": {
			sourceBytes: []byte(`foo = nan`),
			expectedTokens: []token{
				{Type: bareKey, Lexeme: "foo", Literal: "foo", Line: 0, Column: 3},
				{Type: equal, Lexeme: "=", Literal: "=", Line: 0, Column: 5},
				{Type: nan, Lexeme: "nan", Literal: float64(math.NaN()), Line: 0, Column: 7},
				{Type: eof},
			},
		},
		"simple key value with local time": {
			sourceBytes: []byte(`foo = 12:00:21`),
			expectedTokens: []token{
				{Type: bareKey, Lexeme: "foo", Literal: "foo", Line: 0, Column: 3},
				{Type: equal, Lexeme: "=", Literal: "=", Line: 0, Column: 5},
				{Type: localTime, Lexeme: "12:00:21", Literal: time.Date(0, 0, 0, 12, 0, 21, 0, time.UTC), Line: 0, Column: 11},
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
			got, err := scanner.scan()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expecting error, got none")
				}
				return
			}

			want := tt.expectedTokens
			assertTokens(t, got, want)
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
			assert.NilError(t, err, "not expecting error, got %v", err)

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

func Test_LocalTimeParsing(t *testing.T) {
	tests := map[string]struct {
		source    []byte
		tokenType tokenType
		want      time.Time
		shouldErr bool
	}{
		"inferred seconds": {
			source:    []byte(`12:00`),
			tokenType: localTime,
			want:      time.Date(0, 0, 0, 12, 0, 0, 0, time.UTC),
		},
		"malformed seconds should error": {
			source:    []byte(`12:300`),
			tokenType: localTime,
			shouldErr: true,
		},
		"seconds defined": {
			source:    []byte(`12:00:10`),
			tokenType: localTime,
			want:      time.Date(0, 0, 0, 12, 0, 10, 0, time.UTC),
		},
		"terminator character (new line)": {
			source: []byte(`12:00:00
			`),
			tokenType: localTime,
			want:      time.Date(0, 0, 0, 12, 0, 0, 0, time.UTC),
		},
		"terminator character (tab)": {
			source:    []byte(`12:00:00	`),
			tokenType: localTime,
			want:      time.Date(0, 0, 0, 12, 0, 0, 0, time.UTC),
		},
		"terminator character (space)": {
			source:    []byte(`12:00:00 `),
			tokenType: localTime,
			want:      time.Date(0, 0, 0, 12, 0, 0, 0, time.UTC),
		},
		"terminator character (hashtag)": {
			source:    []byte(`12:00:00#`),
			tokenType: localTime,
			want:      time.Date(0, 0, 0, 12, 0, 0, 0, time.UTC),
		},
		"terminator character (closing bracket)": {
			source:    []byte(`12:00:00]`),
			tokenType: localTime,
			want:      time.Date(0, 0, 0, 12, 0, 0, 0, time.UTC),
		},
		"terminator character (closing brace)": {
			source:    []byte("12:00:00}"),
			tokenType: localTime,
			want:      time.Date(0, 0, 0, 12, 0, 0, 0, time.UTC),
		},
		"malformed time with wrong seconds should error": {
			source:    []byte(`12:00:0`),
			tokenType: localTime,
			shouldErr: true,
		},
		"malformed time with wrong minutes should error": {
			source:    []byte(`12:300:00`),
			tokenType: localTime,
			shouldErr: true,
		},
		"seconds + illegal terminal": {
			source:    []byte(`12:30:20;`),
			tokenType: localTime,
			shouldErr: true,
		},
		"milliseconds precision (1 digit)": {
			source:    []byte(`12:20:00.1`),
			tokenType: localTime,
			want:      time.Date(0, 0, 0, 12, 20, 0o0, 1e8, time.UTC),
		},
		"millisecond precision (2 digits)": {
			source:    []byte(`12:20:00.12`),
			tokenType: localTime,
			want:      time.Date(0, 0, 0, 12, 20, 0o0, 12e7, time.UTC),
		},
		"millisecond precision (3 digits)": {
			source:    []byte(`12:20:00.123`),
			tokenType: localTime,
			want:      time.Date(0, 0, 0, 12, 20, 0o0, 123e6, time.UTC),
		},
		"microsecond precision (4 digits)": {
			source:    []byte(`12:20:00.1234`),
			tokenType: localTime,
			want:      time.Date(0, 0, 0, 12, 20, 0o0, 1234e5, time.UTC),
		},
		"microsecond precision (5 digits)": {
			source:    []byte(`12:20:00.12345`),
			tokenType: localTime,
			want:      time.Date(0, 0, 0, 12, 20, 0o0, 12345e4, time.UTC),
		},
		"microsecond precision (6 digits)": {
			source:    []byte(`12:20:00.123456`),
			tokenType: localTime,
			want:      time.Date(0, 0, 0, 12, 20, 0o0, 123456e3, time.UTC),
		},
		"nanosecond precision (7 digits)": {
			source:    []byte(`12:20:00.1234567`),
			tokenType: localTime,
			want:      time.Date(0, 0, 0, 12, 20, 0o0, 1234567e2, time.UTC),
		},
		"nanosecond precision (8 digits)": {
			source:    []byte(`12:20:00.12345678`),
			tokenType: localTime,
			want:      time.Date(0, 0, 0, 12, 20, 0o0, 123456780, time.UTC),
		},
		"nanosecond precision (9 digits)": {
			source:    []byte(`12:20:00.123456789`),
			tokenType: localTime,
			want:      time.Date(0, 0, 0, 12, 20, 0o0, 123456789, time.UTC),
		},
		"nanosecond precision should truncate past 9 digits": {
			source:    []byte(`12:20:00.1234567899999`),
			tokenType: localTime,
			want:      time.Date(0, 0, 0, 12, 20, 0o0, 123456789, time.UTC),
		},
		"hour out of range should error": {
			source:    []byte(`29:00:00`),
			tokenType: localTime,
			shouldErr: true,
		},
		"minute out of range should error": {
			source:    []byte(`24:94:00`),
			tokenType: localTime,
			shouldErr: true,
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
			if tt.shouldErr {
				if err == nil {
					t.Fatalf("Expecting error, got none")
				}
			} else {
				got, ok := tokens[0].Literal.(time.Time)
				if !ok {
					t.Fatalf("Not of type time.Time: %T", got)
				}

				assert.NilError(t, err, "not expecting error, got %v", err)
				assertClock(t, got, tt.want)
			}
		})
	}
}

func Test_LocalDateParsing(t *testing.T) {
	tests := map[string]struct {
		source    []byte
		token     tokenType
		want      time.Time
		shouldErr bool
	}{
		"Simple valid date": {
			source: []byte(`2026-12-02`),
			token:  localDate,
			want:   time.Date(2026, 12, 0o2, 0, 0, 0, 0, time.UTC),
		},
		"Date with invalid month": {
			source:    []byte(`2026-13-02`),
			token:     localDate,
			shouldErr: true,
		},
		"Date with invalid day of the month": {
			source:    []byte(`2026-12-32`),
			token:     localDate,
			shouldErr: true,
		},
		"Invalid format for month": {
			source:    []byte(`2026-112-01`),
			token:     localDate,
			shouldErr: true,
		},
		"Invalid format for day": {
			source:    []byte(`2026-11-101`),
			token:     localDate,
			shouldErr: true,
		},
	}

	for test, tt := range tests {
		t.Run(test, func(t *testing.T) {
			scanner := newScanner(tt.source)
			tokens, err := scanner.scan()

			if tt.shouldErr {
				if err == nil {
					t.Fatalf("Expecting error, got none")
				}
			} else {
				got, ok := tokens[0].Literal.(time.Time)
				if !ok {
					t.Fatalf("Not of type time.Time: %T", got)
				}

				want := tt.want
				assert.NilError(t, err, "Not expecting error, got %v", err)
				assertDate(t, got, want)
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

func assertTokens(t *testing.T, got []token, want []token) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("expected %d tokens, got %d", len(want), len(got))
	}

	for i := range got {
		g, w := got[i], want[i]

		if g.Type != w.Type {
			t.Fatalf("TokenType: want %s, got %s", w.Type, g.Type)
		}

		if g.Lexeme != "" && g.Lexeme != w.Lexeme {
			t.Fatalf("Lexeme: expected %s, got %s", w.Lexeme, g.Lexeme)
		}

		assertTokenLiteral(t, g, w)
	}
}

func assertTokenLiteral(t *testing.T, got, want token) {
	t.Helper()

	if got.Type == nan {
		if f, ok := got.Literal.(float64); !ok {
			t.Fatalf("Did not get float64 for NaN")
		} else if !math.IsNaN(f) {
			t.Fatalf("Expecting NaN literal, got %v", f)
		}
		return
	}

	if reflect.TypeOf(got) != reflect.TypeOf(want) {
		t.Fatalf("Unable to compare: %T != %T", got, want)
	}

	switch got.Literal.(type) {
	case time.Time:
		assertTime(t, got, want)
	case int, int64, string, bool:
		assert.Equal(t, got, want)
	default:
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("got %#v, want %#v", got, want)
		}
	}
}

func assertTime(t *testing.T, got, want token) {
	t.Helper()

	g := got.Literal.(time.Time)
	w := want.Literal.(time.Time)

	switch got.Type {
	case localTime:
		assertClock(t, g, w)
	default:
	}
}

func assertDate(t *testing.T, got, want time.Time) {
	t.Helper()

	gy, gm, gd := got.Date()
	wy, wm, wd := want.Date()

	if gy != wy {
		t.Fatalf("Date: expected years %d, got %d", wy, gy)
	}

	if gm != wm {
		t.Fatalf("Date: expected months %d, got %d", wm, gm)
	}

	if gd != wd {
		t.Fatalf("Date: expected days %d, got %d", wd, gd)
	}
}

func assertClock(t *testing.T, got, want time.Time) {
	t.Helper()
	gh, gm, gs := got.Clock()
	gn := got.Nanosecond()
	wh, wm, ws := want.Clock()
	wn := want.Nanosecond()

	if gh != wh {
		t.Fatalf("Clock: expected hours %d, got %d", wh, gh)
	}

	if gm != wm {
		t.Fatalf("Clock: expected minutes %d, got %d", wm, gm)
	}

	if gs != ws {
		t.Fatalf("Clock: expected seconds %d, got %d", ws, gs)
	}

	if gn != wn {
		t.Fatalf("Clock: expected nanoseconds %d, got %d", wn, gn)
	}
}
