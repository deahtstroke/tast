package tast

import (
	"strings"
)

type Printer struct {
	buf strings.Builder
}

func NewPrinter() *Printer {
	return &Printer{}
}

func (p *Printer) print(doc *Document) (string, error) {
	var prev Node
	for _, node := range doc.content {
		if prev != nil {
			_, currIsTable := node.(*TableNode)
			_, prevIsKV := prev.(*KeyValueNode)
			if currIsTable && prevIsKV {
				p.buf.WriteString("\n")
			}
		}

		if err := node.Accept(p); err != nil {
			return "", err
		}

		prev = node
	}
	return p.buf.String(), nil
}

func (p *Printer) VisitTableNode(n *TableNode) error {
	for _, comment := range n.leadingTrivia {
		p.buf.WriteString(comment.lexeme)
	}

	p.buf.WriteString("[")
	n.key.Accept(p)
	p.buf.WriteString("]")

	for _, trivia := range n.lineTrivia {
		p.buf.WriteString(trivia.lexeme)
	}

	for _, c := range n.children {
		if err := c.Accept(p); err != nil {
			return err
		}
	}

	for _, trivia := range n.trailingTrivia {
		p.buf.WriteString(trivia.lexeme)
	}

	return nil
}

func (p *Printer) VisitKeyValueNode(n *KeyValueNode) error {
	for _, trivia := range n.leadingTrivia {
		p.buf.WriteString(trivia.lexeme)
	}

	if err := n.key.Accept(p); err != nil {
		return err
	}

	p.buf.WriteString(" = ")

	if err := n.value.Accept(p); err != nil {
		return err
	}

	for _, trivia := range n.lineTrivia {
		p.buf.WriteString(trivia.lexeme)
	}

	for _, trivia := range n.trailingTrivia {
		p.buf.WriteString(trivia.lexeme)
	}
	return nil
}

func (p *Printer) VisitKeyNode(n *KeyNode) error {
	lexemes := []string{}
	for _, token := range n.tokens {
		lexemes = append(lexemes, token.Lexeme)
	}
	p.buf.WriteString(strings.Join(lexemes, "."))
	return nil
}

func (p *Printer) VisitStringNode(n *StringNode) error {
	p.buf.WriteString(n.token.Lexeme)
	return nil
}

func (p *Printer) VisitIntegerNode(n *IntegerNode) error {
	p.buf.WriteString(n.token.Lexeme)
	return nil
}

func (p *Printer) VisitFloatNode(n *FloatNode) error {
	p.buf.WriteString(n.token.Lexeme)
	return nil
}

func (p *Printer) VisitBooleanNode(n *BooleanNode) error {
	p.buf.WriteString(n.token.Lexeme)
	return nil
}
