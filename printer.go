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
	for _, node := range doc.Content {
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
	for _, comment := range n.LeadingTrivia {
		p.buf.WriteString(comment.Lexeme)
		p.buf.WriteString("\n")
	}

	p.buf.WriteString("[")
	n.Key.Accept(p)
	p.buf.WriteString("]")

	if n.LineComment != nil {
		p.buf.WriteString(" ")
		p.buf.WriteString(n.LineComment.Lexeme)
	}

	p.buf.WriteString("\n")

	return nil
}

func (p *Printer) VisitKeyValueNode(n *KeyValueNode) error {
	for _, comment := range n.LeadingTrivia {
		p.buf.WriteString(comment.Lexeme)
		p.buf.WriteString("\n")
	}

	if err := n.Key.Accept(p); err != nil {
		return err
	}

	p.buf.WriteString(" = ")

	if err := n.Value.Accept(p); err != nil {
		return err
	}

	p.buf.WriteString(n.LineTrivia.Lexeme)
	p.buf.WriteString("\n")
	return nil
}

func (p *Printer) VisitKeyNode(n *KeyNode) error {
	lexemes := []string{}
	for _, token := range n.Tokens {
		lexemes = append(lexemes, token.Lexeme)
	}
	p.buf.WriteString(strings.Join(lexemes, "."))
	return nil
}

func (p *Printer) VisitStringNode(n *StringNode) error {
	p.buf.WriteString(n.Token.Lexeme)
	return nil
}

func (p *Printer) VisitIntegerNode(n *IntegerNode) error {
	p.buf.WriteString(n.Token.Lexeme)
	return nil
}

func (p *Printer) VisitFloatNode(n *FloatNode) error {
	p.buf.WriteString(n.Token.Lexeme)
	return nil
}

func (p *Printer) VisitBooleanNode(n *BooleanNode) error {
	p.buf.WriteString(n.Token.Lexeme)
	return nil
}
