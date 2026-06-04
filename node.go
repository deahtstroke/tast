package tast

import (
	"fmt"
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
	LeadingTrivia []Trivia

	// Comment in the same line as the table
	LineTrivia *Trivia

	// Any comments that are leftover after the table
	// itself
	TrailingTrivia []Trivia

	// Tokens that make up the table-key
	Tokens []Token
}

func (n *TableNode) NodeLexeme() string {
	return "[" + n.Key.NodeLiteral() + "]"
}

func (n *TableNode) Accept(v Visitor) error {
	return v.VisitTableNode(n)
}

func (n *TableNode) Set(key string, value any) error {
	k, exists := n.FindKey(key)
	if !exists {
		return fmt.Errorf("Cannot find key %s", key)
	}

	var node Node
	switch v := value.(type) {
	case int64:
		node = &IntegerNode{
			Value: v,
		}
	case string:
		node = &StringNode{
			Value: v,
		}
	case float64:
		node = &FloatNode{
			Value: v,
		}
	case bool:
		node = &BooleanNode{
			Value: v,
		}
	default:
	}

	k.Value = node
	return nil
}

func (n *TableNode) Delete(key string) bool {
	i := n.findNodeIndex(key)
	if i == -1 {
		return false
	}

	n.Children = append(n.Children[:i], n.Children[i+1:]...)
	return false
}

func (n *TableNode) FindKey(key string) (*KeyValueNode, bool) {
	i := n.findNodeIndex(key)
	if i == -1 {
		return nil, false
	}

	return n.Children[i].(*KeyValueNode), true
}

func (n *TableNode) findNodeIndex(key string) int {
	for i, child := range n.Children {
		if kv, ok := child.(*KeyValueNode); ok && KeysMatch(key, kv.Key.Segments) {
			return i
		}
	}

	return -1
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
	LeadingTrivia []Trivia

	// Comment at the end of the line in a key-value node
	LineTrivia *Trivia

	// Any comments at the end of a key-value node
	// Note: This field only gets filled if the only tokens leftover
	// after the dangling trivia are New lines and/or EOF
	TrailingTrivia []Trivia
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

func KeysMatch(key string, segments []string) bool {
	keyArr := strings.Split(key, ".")
	if len(keyArr) != len(segments) {
		return false
	}

	for i := range len(keyArr) {
		if keyArr[i] != segments[i] {
			return false
		}
	}

	return true
}
