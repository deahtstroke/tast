package tast

import (
	"strings"
)

type Visitor interface {
	VisitTableNode(*TableNode) error
	VisitKeyNode(*KeyNode) error
	VisitKeyValueNode(*KeyValueNode) error
	VisitStringNode(*StringNode) error
	VisitIntegerNode(*IntegerNode) error
	VisitFloatNode(*FloatNode) error
	VisitBooleanNode(*BooleanNode) error
}

type Node interface {
	NodeLexeme() string
	Accept(Visitor) error
}

// Top-most node representation of a TOML file
// in the AST
type Document struct {
	Content []Node
}

// Trivia really is just comments that start with '#'
// the only worthwhile state saving for trivia is the
// raw literal string in the comment
type Trivia struct {
	Lexeme string
}

// Node reprsentation of a TOML table as specified in
// https://toml.io/en/v1.1.0#table
type TableNode struct {
	// The key of the TableNode
	Key *KeyNode

	// Child nodes that belong to this table
	Children []Node

	// Any leading comments that may come before the
	// table itself
	LeadingComments []*Trivia

	// Comment in the same line as the table
	TrailingComment *Trivia

	// Tokens that make up the table-key
	Tokens []Token
}

func (n *TableNode) NodeLexeme() string {
	return "[" + n.Key.NodeLiteral() + "]"
}

func (n *TableNode) Accept(v Visitor) error {
	return v.VisitTableNode(n)
}

// Node representation of a Key in the TOML specification
// https://toml.io/en/v1.1.0#keys
type KeyNode struct {
	// Segments that make up the Key
	// KeyNodes can be made up of bare-keys and quoted-keys
	Segments []string

	// List of scanner tokens that make up this KeyNode
	Tokens []Token
}

func (n *KeyNode) NodeLiteral() string {
	return strings.Join(n.Segments, ".")
}

func (n *KeyNode) Accept(v Visitor) error {
	return v.VisitKeyNode(n)
}

// Node representation of a Key-Value pair in the TOML specification
// https://toml.io/en/v1.1.0#keyvalue-pair
type KeyValueNode struct {
	// Key identifier that represents this key-value pair
	Key *KeyNode

	// Value that this key-value pair has
	Value Node

	// List of scanner tokens that make up this key-value pair
	Tokens []Token

	// Any leading comments that may come before a key-value node
	LeadingComments []Trivia

	// Comment at the end of the line in a key-value node
	TrailingComment Trivia
}

func (n *KeyValueNode) NodeLexeme() string {
	segs := []string{n.Key.NodeLiteral(), n.Value.NodeLexeme()}
	return strings.Join(segs, " = ")
}

func (n *KeyValueNode) Accept(v Visitor) error {
	return v.VisitKeyValueNode(n)
}

type StringNode struct {
	Value string
	Token Token
}

func (n *StringNode) NodeLexeme() string {
	return n.Token.Lexeme
}

func (n *StringNode) Accept(v Visitor) error {
	return v.VisitStringNode(n)
}

type IntegerNode struct {
	Value int64
	Token Token
}

func (n *IntegerNode) NodeLexeme() string {
	return n.Token.Lexeme
}

func (n *IntegerNode) Accept(v Visitor) error {
	return v.VisitIntegerNode(n)
}

type FloatNode struct {
	Value float64
	Token Token
}

func (n *FloatNode) NodeLexeme() string {
	return n.Token.Lexeme
}

func (n *FloatNode) Accept(v Visitor) error {
	return v.VisitFloatNode(n)
}

type BooleanNode struct {
	Value bool
	Token Token
}

func (n *BooleanNode) NodeLexeme() string {
	if n.Value {
		return "true"
	} else {
		return "false"
	}
}

func (n *BooleanNode) Accept(v Visitor) error {
	return v.VisitBooleanNode(n)
}
