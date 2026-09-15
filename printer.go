package tast

import (
	"io"
	"strings"
)

type printer struct {
	writer io.Writer
	prefix string
	n      int64
	err    error
}

func newPrinter(w io.Writer) *printer {
	return &printer{writer: w}
}

func (p *printer) write(s string) {
	if p.err != nil {
		return
	}

	_, p.err = io.WriteString(p.writer, s)
}

func (p *printer) print(doc *Document) error {
	for _, node := range doc.content {
		if err := node.accept(p); err != nil {
			return err
		}
	}
	return nil
}

func (p *printer) visitTableNode(n *TableNode) error {
	for _, comment := range n.leadingTrivia {
		p.write(comment.Lexeme)
	}

	if n.isImplicit {
		previous := p.prefix
		p.prefix += n.key.segments[0] + "."
		for _, child := range n.children {
			if err := child.accept(p); err != nil {
				return err
			}
		}

		p.prefix = previous
		return nil
	}

	p.write("[")
	if err := n.key.accept(p); err != nil {
		return err
	}
	p.write("]")

	for _, trivia := range n.lineTrivia {
		p.write(trivia.Lexeme)
	}

	for _, c := range n.children {
		if err := c.accept(p); err != nil {
			return err
		}
	}

	for _, trivia := range n.trailingTrivia {
		p.write(trivia.Lexeme)
	}

	return nil
}

func (p *printer) visitKeyValueNode(n *KeyValueNode) error {
	for _, t := range n.leadingTrivia {
		p.write(t.Lexeme)
	}

	// dotted-key prefix (if there is one)
	p.write(p.prefix)

	if err := n.key.accept(p); err != nil {
		return err
	}

	p.write(" = ")

	if err := n.value.accept(p); err != nil {
		return err
	}

	for _, trivia := range n.lineTrivia {
		p.write(trivia.Lexeme)
	}

	for _, trivia := range n.trailingTrivia {
		p.write(trivia.Lexeme)
	}
	return nil
}

func (p *printer) visitKeyNode(n *keyNode) error {
	lexemes := []string{}
	for _, token := range n.tokens {
		lexemes = append(lexemes, token.Lexeme)
	}
	p.write(strings.Join(lexemes, "."))
	return nil
}

func (p *printer) visitStringNode(n *stringNode) error {
	p.write(n.token.Lexeme)
	return nil
}

func (p *printer) visitIntegerNode(n *integerNode) error {
	p.write(n.token.Lexeme)
	return nil
}

func (p *printer) visitFloatNode(n *floatNode) error {
	p.write(n.token.Lexeme)
	return nil
}

func (p *printer) visitBooleanNode(n *booleanNode) error {
	p.write(n.token.Lexeme)
	return nil
}
