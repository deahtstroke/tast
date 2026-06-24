package tast

import (
	"strings"
)

type printer struct {
	buf strings.Builder
}

func newPrinter() *printer {
	return &printer{}
}

func (p *printer) print(doc *Document) (string, error) {
	for _, node := range doc.content {
		if err := node.accept(p); err != nil {
			return "", err
		}
	}
	return p.buf.String(), nil
}

func (p *printer) visitTableNode(n *TableNode) error {
	for _, comment := range n.leadingTrivia {
		p.buf.WriteString(comment.Lexeme)
	}

	p.buf.WriteString("[")
	n.key.accept(p)
	p.buf.WriteString("]")

	for _, trivia := range n.lineTrivia {
		p.buf.WriteString(trivia.Lexeme)
	}

	for _, c := range n.children {
		if err := c.accept(p); err != nil {
			return err
		}
	}

	for _, trivia := range n.trailingTrivia {
		p.buf.WriteString(trivia.Lexeme)
	}

	return nil
}

func (p *printer) visitKeyValueNode(n *KeyValueNode) error {
	for _, t := range n.leadingTrivia {
		p.buf.WriteString(t.Lexeme)
	}

	if err := n.key.accept(p); err != nil {
		return err
	}

	p.buf.WriteString(" = ")

	if err := n.value.accept(p); err != nil {
		return err
	}

	for _, trivia := range n.lineTrivia {
		p.buf.WriteString(trivia.Lexeme)
	}

	for _, trivia := range n.trailingTrivia {
		p.buf.WriteString(trivia.Lexeme)
	}
	return nil
}

func (p *printer) visitKeyNode(n *keyNode) error {
	lexemes := []string{}
	for _, token := range n.tokens {
		lexemes = append(lexemes, token.Lexeme)
	}
	p.buf.WriteString(strings.Join(lexemes, "."))
	return nil
}

func (p *printer) visitStringNode(n *stringNode) error {
	p.buf.WriteString(n.token.Lexeme)
	return nil
}

func (p *printer) visitIntegerNode(n *integerNode) error {
	p.buf.WriteString(n.token.Lexeme)
	return nil
}

func (p *printer) visitFloatNode(n *floatNode) error {
	p.buf.WriteString(n.token.Lexeme)
	return nil
}

func (p *printer) visitBooleanNode(n *booleanNode) error {
	p.buf.WriteString(n.token.Lexeme)
	return nil
}
