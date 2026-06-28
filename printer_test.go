package tast

import (
	"fmt"
	"strings"
	"testing"
)

func Test_PrinterSuccess(t *testing.T) {
	doc := makeDoc(
		makeKV([]string{"concurrency"}, makeVal(int64(100)), withLine("\n")),
		makeTable(makeKey("output"), true, []node{makeKV([]string{"errors"}, makeVal("stderr"), withLine("\n"))}, []NodeOption{}),
		makeTable(makeKey("output"), true, []node{makeKV([]string{"\"logs\""}, makeVal("stdout"), withLine("\n"))}, []NodeOption{}),
		makeTable(makeKey("input"), true, []node{makeTable(makeKey("src"), true, []node{makeKV([]string{"compress"}, makeVal("zstd"))}, []NodeOption{})}, []NodeOption{}),
		makeTable(makeKey("database", "rivenbot"),
			false,
			[]node{
				makeKV([]string{"url"}, makeVal("postgres://localhost:5432/rivenbot"), withLine("\n")),
				makeKV([]string{"username"}, makeVal("daniel")),
			},
			[]NodeOption{withLeading("\n", "# Database details for Rivenbot", "\n", "# Dev only", "\n"), withLine("\n")}),
	)

	got, err := newPrinter().print(doc)
	if err != nil {
		t.Fatalf("Got an error while calling 'print': %v", err)
	}

	expected := `concurrency = 100
output.errors = "stderr"
output."logs" = "stdout"
input.src.compress = "zstd"

# Database details for Rivenbot
# Dev only
[database.rivenbot]
url = "postgres://localhost:5432/rivenbot"
username = "daniel"`

	if got != expected {
		t.Fatalf("String mismatch:\n%s", diffStrings(expected, got))
	}
}

func makeDoc(nodes ...node) *Document {
	return &Document{content: nodes}
}

type NodeOption func(node)

func makeTable(keyNode *keyNode, implicit bool, children []node, opts []NodeOption) *TableNode {
	table := &TableNode{
		key:        keyNode,
		isImplicit: implicit,
		tokens:     keyNode.tokens,
		children:   children,
	}

	for _, opt := range opts {
		opt(table)
	}

	return table
}

func withLeading(tr ...string) NodeOption {
	return func(node node) {
		var leadingTrivia []trivia
		for _, t := range tr {
			leadingTrivia = append(leadingTrivia, makeTrivia(t))
		}

		switch n := node.(type) {
		case *TableNode:
			n.leadingTrivia = leadingTrivia
		case *KeyValueNode:
			n.leadingTrivia = leadingTrivia
		default:
		}
	}
}

func withTrailing(tr ...string) NodeOption {
	return func(node node) {
		var trailingTrivia []trivia
		for _, c := range tr {
			trailingTrivia = append(trailingTrivia, makeTrivia(c))
		}

		switch n := node.(type) {
		case *TableNode:
			n.trailingTrivia = trailingTrivia
		case *KeyValueNode:
			n.trailingTrivia = trailingTrivia
		default:
		}
	}
}

func withLine(tr ...string) NodeOption {
	return func(node node) {
		var lineTrivia []trivia
		for _, t := range tr {
			lineTrivia = append(lineTrivia, makeTrivia(t))
		}

		switch n := node.(type) {
		case *TableNode:
			n.lineTrivia = lineTrivia
		case *KeyValueNode:
			n.lineTrivia = lineTrivia
		default:
		}
	}
}

func makeTrivia(lex string) trivia {
	return trivia{
		Lexeme: lex,
	}
}

func makeKV(keys []string, value node, opts ...NodeOption) *KeyValueNode {
	kv := &KeyValueNode{
		key:   makeKey(keys...),
		value: value,
	}

	for _, opt := range opts {
		opt(kv)
	}

	return kv
}

func makeKey(keys ...string) *keyNode {
	var tokens []token
	for _, key := range keys {
		tokens = append(tokens, token{Lexeme: key})
	}
	return &keyNode{
		segments: keys,
		tokens:   tokens,
	}
}

func makeVal(value any) node {
	switch v := value.(type) {
	case float64:
		return &floatNode{
			value: v,
			token: token{
				Lexeme: fmt.Sprintf("%v", v),
			},
		}
	case int64:
		return &integerNode{
			value: int64(v),
			token: token{
				Lexeme: fmt.Sprintf("%v", v),
			},
		}
	case string:
		return &stringNode{
			value: v,
			token: token{
				Lexeme: fmt.Sprintf("\"%v\"", v),
			},
		}
	case bool:
		return &booleanNode{
			value: v,
			token: token{
				Lexeme: fmt.Sprintf("%v", v),
			},
		}
	}
	return nil
}

// diffStrings walks two strings simultaneously and returns a diff string
// with a '^' marker pointing to the first character that differs.
func diffStrings(expected, got string) string {
	// find diff index
	minLen := len(expected)
	if len(got) < minLen {
		minLen = len(got)
	}

	diffAt := -1
	for i := 0; i < minLen; i++ {
		if expected[i] != got[i] {
			diffAt = i
			break
		}
	}

	if diffAt == -1 && len(expected) != len(got) {
		diffAt = minLen
	}

	if diffAt == -1 {
		return "strings are equal"
	}

	// find which line and column the diff is on
	line := 0
	col := 0
	for i := 0; i < diffAt; i++ {
		if expected[i] == '\n' {
			line++
			col = 0
		} else {
			col++
		}
	}

	// split expected and got into lines
	expectedLines := strings.Split(expected, "\n")
	gotLines := strings.Split(got, "\n")

	var b strings.Builder
	b.WriteString(fmt.Sprintf("first difference at line %d col %d:\n", line+1, col+1))

	if line < len(expectedLines) {
		b.WriteString(fmt.Sprintf("expected: %q\n", expectedLines[line]))
	}
	if line < len(gotLines) {
		b.WriteString(fmt.Sprintf("got:      %q\n", gotLines[line]))
	}

	b.WriteString("          ")
	for i := 0; i < col; i++ {
		b.WriteByte(' ')
	}
	b.WriteByte('^')

	return b.String()
}
