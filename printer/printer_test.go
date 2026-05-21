package printer

import (
	"fmt"
	"testing"

	"github.com/deahtstroke/toml-ast/parser"
	"github.com/deahtstroke/toml-ast/scanner"
)

func Test_Printer(t *testing.T) {
	doc := makeDoc(
		makeKV([]string{"hello", "world"}, makeVal("This is a string")),
		makeKV([]string{"foo"}, makeVal("bar")),
		makeTable(makeKey("database", "rivenbot"),
			withLeading([]string{"# Database details for Rivenbot", "# Dev only"})),
		makeKV([]string{"url"}, makeVal("postgres://localhost:5432/rivenbot")),
		makeKV([]string{"username"}, makeVal("daniel")),
	)

	res, err := NewPrinter().Print(doc)
	if err != nil {
		t.Fatalf("Got an error while calling 'print': %v", err)
	}

	expected := `hello.world = "This is a string"
foo = "bar"

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

func makeDoc(nodes ...parser.Node) *parser.Document {
	return &parser.Document{Content: nodes}
}

type TableOption func(*parser.TableNode)

func makeTable(keyNode *parser.KeyNode, opts ...TableOption) *parser.TableNode {
	table := &parser.TableNode{
		Key:    keyNode,
		Tokens: keyNode.Tokens,
	}

	for _, opt := range opts {
		opt(table)
	}

	return table
}

func withLeading(comments []string) TableOption {
	return func(tn *parser.TableNode) {
		for _, c := range comments {
			tn.LeadingComments = append(tn.LeadingComments, makeTrivia(c))
		}
	}
}

func withTrailing(comment string) TableOption {
	return func(tn *parser.TableNode) {
		tn.TrailingComment = makeTrivia(comment)
	}
}

func makeTrivia(lex string) *parser.Trivia {
	if lex == "" {
		return nil
	}

	return &parser.Trivia{
		Lexeme: lex,
	}
}

func makeKV(keys []string, value parser.Node) *parser.KeyValueNode {
	return &parser.KeyValueNode{
		Key:   makeKey(keys...),
		Value: value,
	}
}

func makeKey(keys ...string) *parser.KeyNode {
	var tokens []scanner.Token
	for _, key := range keys {
		tokens = append(tokens, scanner.Token{Lexeme: key})
	}
	return &parser.KeyNode{
		Segments: keys,
		Tokens:   tokens,
	}
}

func makeVal(value any) parser.Node {
	switch v := value.(type) {
	case float64:
		return &parser.FloatNode{
			Value: v,
			Token: scanner.Token{
				Lexeme: fmt.Sprintf("%v", v),
			},
		}
	case int64:
		return &parser.IntegerNode{
			Value: v,
			Token: scanner.Token{
				Lexeme: fmt.Sprintf("%v", v),
			},
		}
	case string:
		return &parser.StringNode{
			Value: v,
			Token: scanner.Token{
				Lexeme: fmt.Sprintf("\"%v\"", v),
			},
		}
	case bool:
		return &parser.BooleanNode{
			Value: v,
			Token: scanner.Token{
				Lexeme: fmt.Sprintf("%v", v),
			},
		}
	}
	return nil
}
