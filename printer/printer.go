package printer

import (
	"strings"

	"github.com/deahtstroke/toml-ast/parser"
)

type Printer struct {
	buf strings.Builder
}

func NewPrinter() *Printer {
	return &Printer{}
}

func (p *Printer) Print(doc *parser.Document) (string, error) {
	var prev parser.Node
	for _, node := range doc.Content {
		if prev != nil {
			_, currIsTable := node.(*parser.TableNode)
			_, prevIsKV := prev.(*parser.KeyValueNode)
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

func (p *Printer) VisitTableNode(n *parser.TableNode) error {
	for _, comment := range n.LeadingComments {
		p.buf.WriteString(comment.Lexeme)
		p.buf.WriteString("\n")
	}

	p.buf.WriteString("[")
	n.Key.Accept(p)
	p.buf.WriteString("]")

	if n.TrailingComment != nil {
		p.buf.WriteString(" ")
		p.buf.WriteString(n.TrailingComment.Lexeme)
	}

	p.buf.WriteString("\n")

	return nil
}

func (p *Printer) VisitKeyValueNode(n *parser.KeyValueNode) error {
	for _, comment := range n.LeadingComments {
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

	p.buf.WriteString(n.TrailingComment.Lexeme)
	p.buf.WriteString("\n")
	return nil
}

func (p *Printer) VisitKeyNode(n *parser.KeyNode) error {
	lexemes := []string{}
	for _, token := range n.Tokens {
		lexemes = append(lexemes, token.Lexeme)
	}
	p.buf.WriteString(strings.Join(lexemes, "."))
	return nil
}

func (p *Printer) VisitStringNode(n *parser.StringNode) error {
	p.buf.WriteString(n.Token.Lexeme)
	return nil
}

func (p *Printer) VisitIntegerNode(n *parser.IntegerNode) error {
	p.buf.WriteString(n.Token.Lexeme)
	return nil
}

func (p *Printer) VisitFloatNode(n *parser.FloatNode) error {
	p.buf.WriteString(n.Token.Lexeme)
	return nil
}

func (p *Printer) VisitBooleanNode(n *parser.BooleanNode) error {
	p.buf.WriteString(n.Token.Lexeme)
	return nil
}
