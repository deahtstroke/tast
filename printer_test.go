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
			withLeading([]string{"# Database details for Rivenbot", "# Dev only"})),
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
	return &Document{Content: nodes}
}

type TableOption func(*TableNode)

func makeTable(keyNode *KeyNode, opts ...TableOption) *TableNode {
	table := &TableNode{
		Key:    keyNode,
		Tokens: keyNode.Tokens,
	}

	for _, opt := range opts {
		opt(table)
	}

	return table
}

func withLeading(comments []string) TableOption {
	return func(tn *TableNode) {
		for _, c := range comments {
			tn.LeadingTrivia = append(tn.LeadingTrivia, *makeTrivia(c))
		}
	}
}

func withTrailing(comment string) TableOption {
	return func(tn *TableNode) {
		tn.LineTrivia = makeTrivia(comment)
	}
}

func makeTrivia(lex string) *Trivia {
	if lex == "" {
		return nil
	}

	return &Trivia{
		Lexeme: lex,
	}
}

func makeKV(keys []string, value Node) *KeyValueNode {
	return &KeyValueNode{
		Key:   makeKey(keys...),
		Value: value,
	}
}

func makeKey(keys ...string) *KeyNode {
	var tokens []Token
	for _, key := range keys {
		tokens = append(tokens, Token{Lexeme: key})
	}
	return &KeyNode{
		Segments: keys,
		Tokens:   tokens,
	}
}

func makeVal(value any) Node {
	switch v := value.(type) {
	case float64:
		return &FloatNode{
			Value: v,
			Token: Token{
				Lexeme: fmt.Sprintf("%v", v),
			},
		}
	case int64:
		return &IntegerNode{
			Value: int64(v),
			Token: Token{
				Lexeme: fmt.Sprintf("%v", v),
			},
		}
	case string:
		return &StringNode{
			Value: v,
			Token: Token{
				Lexeme: fmt.Sprintf("\"%v\"", v),
			},
		}
	case bool:
		return &BooleanNode{
			Value: v,
			Token: Token{
				Lexeme: fmt.Sprintf("%v", v),
			},
		}
	}
	return nil
}
