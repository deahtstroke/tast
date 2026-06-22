package tast

import (
	"fmt"
	"testing"
)

func Test_PrinterSuccess(t *testing.T) {
	doc := makeDoc(
		makeKV([]string{"concurrency"}, makeVal(int64(100))),
		makeKV([]string{"output.errors"}, makeVal("stderr")),
		makeKV([]string{"output.\"logs\""}, makeVal("stdout")),
		makeTable(makeKey("database", "rivenbot"),
			withLeading("# Database details for Rivenbot", "# Dev only")),
		makeKV([]string{"url"}, makeVal("postgres://localhost:5432/rivenbot")),
		makeKV([]string{"username"}, makeVal("daniel")),
	)

	res, err := NewPrinter().print(doc)
	if err != nil {
		t.Fatalf("Got an error while calling 'print': %v", err)
	}

	expected := `concurrency = 100
output.errors = "stderr"
output."logs" = "stdout"

# Database details for Rivenbot
# Dev only
[database.rivenbot]
url = "postgres://localhost:5432/rivenbot"
username = "daniel"
`

	if res != expected {
		t.Fatalf("Expected string is formatted incorrectly. \n\nExpected: \n%v \nGot: \n%v", expected, res)
	}
}

func makeDoc(nodes ...Node) *Document {
	return &Document{content: nodes}
}

type TableOption func(*TableNode)

func makeTable(keyNode *KeyNode, opts ...TableOption) *TableNode {
	table := &TableNode{
		key:    keyNode,
		tokens: keyNode.tokens,
	}

	for _, opt := range opts {
		opt(table)
	}

	return table
}

func withLeading(comments ...string) TableOption {
	return func(tn *TableNode) {
		for _, c := range comments {
			tn.leadingTrivia = append(tn.leadingTrivia, makeTrivia(c))
		}
	}
}

func withTrailing(comments ...string) TableOption {
	return func(tn *TableNode) {
		var lineTrivia []Trivia
		for _, c := range comments {
			lineTrivia = append(lineTrivia, makeTrivia(c))
		}
		tn.lineTrivia = lineTrivia
	}
}

func makeTrivia(lex string) Trivia {
	return Trivia{
		lexeme: lex,
	}
}

func makeKV(keys []string, value Node) *KeyValueNode {
	return &KeyValueNode{
		key:   makeKey(keys...),
		value: value,
	}
}

func makeKey(keys ...string) *KeyNode {
	var tokens []token
	for _, key := range keys {
		tokens = append(tokens, token{Lexeme: key})
	}
	return &KeyNode{
		segments: keys,
		tokens:   tokens,
	}
}

func makeVal(value any) Node {
	switch v := value.(type) {
	case float64:
		return &FloatNode{
			value: v,
			token: token{
				Lexeme: fmt.Sprintf("%v", v),
			},
		}
	case int64:
		return &IntegerNode{
			value: int64(v),
			token: token{
				Lexeme: fmt.Sprintf("%v", v),
			},
		}
	case string:
		return &StringNode{
			value: v,
			token: token{
				Lexeme: fmt.Sprintf("\"%v\"", v),
			},
		}
	case bool:
		return &BooleanNode{
			value: v,
			token: token{
				Lexeme: fmt.Sprintf("%v", v),
			},
		}
	}
	return nil
}
