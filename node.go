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
	Accept(Visitor) error
}

// Top-most node representation of a TOML file
// in the AST
type Document struct {
	content []Node
}

// Retrieves a root-level key-value node defined by their key
// Key argument can be defined as a barekey or a dotted-key
func (d *Document) KeyValue(key string) (*KeyValueNode, bool) {
	for _, node := range d.content {
		kv, ok := node.(*KeyValueNode)
		if ok && KeysMatch(key, kv.key.segments) {
			return kv, true
		}
	}
	return nil, false
}

// Retrives a Table defined by their key
// Key argument can be defined as a barekey or a dotted-key
func (d *Document) Table(key string) (*TableNode, bool) {
	for _, node := range d.content {
		t, ok := node.(*TableNode)
		if ok && KeysMatch(key, t.key.segments) {
			return t, true
		}
	}
	return nil, false
}

// Trivia really is just comments that start with '#'
// the only worthwhile state saving for trivia is the
// raw literal string in the comment
type Trivia struct {
	lexeme string
}

// Node reprsentation of a TOML table as specified in
// https://toml.io/en/v1.1.0#table
type TableNode struct {
	// The key of the TableNode
	key *KeyNode

	// Child nodes that belong to this table
	children []Node

	// Any leading comments that may come before the
	// table itself
	leadingTrivia []Trivia

	// Comment in the same line as the table
	lineTrivia *Trivia

	// Any comments that are leftover after the table
	// itself
	trailingTrivia []Trivia

	// tokens that make up the table-key
	tokens []token
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
			value: v,
		}
	case string:
		node = &StringNode{
			value: v,
		}
	case float64:
		node = &FloatNode{
			value: v,
		}
	case bool:
		node = &BooleanNode{
			value: v,
		}
	default:
	}

	k.value = node
	return nil
}

func (n *TableNode) Delete(key string) bool {
	i := n.findNodeIndex(key)
	if i == -1 {
		return false
	}

	n.children = append(n.children[:i], n.children[i+1:]...)
	return true
}

func (n *TableNode) FindKey(key string) (*KeyValueNode, bool) {
	i := n.findNodeIndex(key)
	if i == -1 {
		return nil, false
	}

	return n.children[i].(*KeyValueNode), true
}

func (n *TableNode) findNodeIndex(key string) int {
	for i, child := range n.children {
		if kv, ok := child.(*KeyValueNode); ok && KeysMatch(key, kv.key.segments) {
			return i
		}
	}

	return -1
}

// Node representation of a Key in the TOML specification
// https://toml.io/en/v1.1.0#keys
type KeyNode struct {
	// segments that make up the Key
	// KeyNodes can be made up of bare-keys and quoted-keys
	segments []string

	// List of scanner tokens that make up this KeyNode
	tokens []token
}

func (n *KeyNode) Accept(v Visitor) error {
	return v.VisitKeyNode(n)
}

// Node representation of a Key-Value pair in the TOML specification
// https://toml.io/en/v1.1.0#keyvalue-pair
type KeyValueNode struct {
	// key identifier that represents this key-value pair
	key *KeyNode

	// value that this key-value pair has
	value Node

	// List of scanner tokens that make up this key-value pair
	tokens []token

	// Any leading comments that may come before a key-value node
	leadingTrivia []Trivia

	// Comment at the end of the line in a key-value node
	lineTrivia *Trivia

	// Any comments at the end of a key-value node
	// Note: This field only gets filled if the only tokens leftover
	// after the dangling trivia are New lines and/or EOF
	trailingTrivia []Trivia
}

func (n *KeyValueNode) Accept(v Visitor) error {
	return v.VisitKeyValueNode(n)
}

type StringNode struct {
	value string
	token token
}

func (n *StringNode) Accept(v Visitor) error {
	return v.VisitStringNode(n)
}

type IntegerNode struct {
	value int64
	token token
}

func (n *IntegerNode) Accept(v Visitor) error {
	return v.VisitIntegerNode(n)
}

type FloatNode struct {
	value float64
	token token
}

func (n *FloatNode) Accept(v Visitor) error {
	return v.VisitFloatNode(n)
}

type BooleanNode struct {
	value bool
	token token
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
