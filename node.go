package tast

import (
	"fmt"
	"strconv"
	"strings"
)

const STRING_SEPARATOR = "."

type Visitor interface {
	VisitTableNode(*TableNode) error
	VisitKeyNode(*KeyNode) error
	VisitKeyValueNode(*KeyValueNode) error
	VisitStringNode(*StringNode) error
	VisitIntegerNode(*IntegerNode) error
	VisitFloatNode(*FloatNode) error
	VisitBooleanNode(*BooleanNode) error
}

// The Node interface represents all the shared methods between all different
// types of Nodes specified in the TOML spec
type Node interface {
	Accept(Visitor) error
	GetToken() token
}

// Top-most node representation of a TOML file
type Document struct {
	content []Node
}

// FindKey() retrieves a key-value node defined by the 'key' parameter
// The 'key' parameter can be defined as a bare key or a dotted-key
// in accordance to the TOML spec
func (d *Document) FindKey(key string) (*KeyValueNode, bool) {
	for _, node := range d.content {
		kv, ok := node.(*KeyValueNode)
		if ok && KeysMatch(key, kv.key.segments) {
			return kv, true
		}
	}
	return nil, false
}

// Table() attempts to retrive a Table defined by the key parameter
// The 'key' parameter can be defined as a bare key or a dotted-key
// in accordance to the TOML spec
func (d *Document) Table(key string) (*TableNode, bool) {
	for _, node := range d.content {
		t, ok := node.(*TableNode)
		if ok && KeysMatch(key, t.key.segments) {
			return t, true
		}
	}
	return nil, false
}

// Returns a string representation of the document
func (d *Document) String() (string, error) {
	p := NewPrinter()
	return p.print(d)
}

type TriviaType int

const (
	NewLineTrivia TriviaType = iota
	TabTrivia
	CommentTrivia
)

var TriviaTypes map[TriviaType]string = map[TriviaType]string{
	NewLineTrivia: "NewLineTrivia",
	TabTrivia:     "TabTrivia",
	CommentTrivia: "CommentTrivia",
}

// Trivia really is just comments that start with '#'
// the only worthwhile state saving for trivia is the
// raw literal string in the comment
type Trivia struct {
	Type   TriviaType
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
	lineTrivia []Trivia

	// Any comments that are leftover after the table
	// itself
	trailingTrivia []Trivia

	// tokens that make up the table-key
	tokens []token
}

func (n *TableNode) Accept(v Visitor) error {
	return v.VisitTableNode(n)
}

func (n *TableNode) GetToken() token {
	return n.tokens[0]
}

// Key() return the string representation of a TOML key
// TOML keys according to the spec., can be bare-keys, basic strings,
// or a combination of both
func (n *TableNode) Key() string {
	return n.key.Key()
}

// Set() mutates a given key given by the parameter 'key' with the
// passed in value 'value'. An error is returned if the key was not
// found or if the passed in type is not recognized
func (n *TableNode) Set(key string, value any) error {
	i := n.findNodeIndex(key)
	if i == -1 {
		return fmt.Errorf("tast: key %q not found", key)
	}

	kv := n.children[i].(*KeyValueNode)

	switch v := value.(type) {
	case int:
		return n.Set(key, int64(v))
	case int8:
		return n.Set(key, int64(v))
	case int32:
		return n.Set(key, int64(v))
	case int64:
		kv.value = &IntegerNode{
			value: v,
			token: token{
				Type:    INTEGER,
				Lexeme:  strconv.FormatInt(v, 10),
				Literal: v,
				Line:    kv.value.GetToken().Line,
				Column:  kv.value.GetToken().Column,
			},
		}
	case string:
		kv.value = &StringNode{
			value: v,
			token: token{
				Type:    BASIC_STRING,
				Lexeme:  `"` + v + `"`,
				Literal: v,
				Line:    kv.value.GetToken().Line,
				Column:  kv.value.GetToken().Column,
			},
		}
	case float32:
		return kv.Set(float64(v))
	case float64:
		kv.value = &FloatNode{
			value: v,
			token: token{
				Type:    FLOAT,
				Lexeme:  strconv.FormatFloat(v, 'f', -1, 64),
				Literal: v,
				Line:    kv.value.GetToken().Line,
				Column:  kv.value.GetToken().Column,
			},
		}
	case bool:
		lexeme := "false"
		nodeType := FALSE
		if v {
			nodeType = TRUE
			lexeme = "true"
		}
		kv.value = &BooleanNode{
			value: v,
			token: token{
				Type:    nodeType,
				Lexeme:  lexeme,
				Literal: v,
				Line:    kv.value.GetToken().Line,
				Column:  kv.value.GetToken().Column,
			},
		}
	default:
		return fmt.Errorf("tast: unsupported value type %T", value)
	}

	return nil
}

// Delete() eliminates a key alongside its value in a TOML table.
// If the key was found and removed successfully it returns true.
// Otherwise it returns false
func (n *TableNode) Delete(key string) bool {
	i := n.findNodeIndex(key)
	if i == -1 {
		return false
	}

	n.children = append(n.children[:i], n.children[i+1:]...)
	return true
}

// FindKey() attempts to find a given key-value pair inside a TOML table.
// If the key is found successfully, then it returns the node alongside true.
// Otherwise it returns false
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

func (n *KeyNode) Key() string {
	if len(n.segments) == 0 {
		return ""
	}

	return strings.Join(n.segments, STRING_SEPARATOR)
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
	lineTrivia []Trivia

	// Any comments at the end of a key-value node
	// Note: This field only gets filled if the only tokens leftover
	// after the dangling trivia are New lines and/or EOF
	trailingTrivia []Trivia
}

func (kv *KeyValueNode) Accept(v Visitor) error {
	return v.VisitKeyValueNode(kv)
}

func (kv *KeyValueNode) GetToken() token {
	return kv.tokens[0]
}

func (kv *KeyValueNode) Int() (int64, bool) {
	i, ok := kv.value.(*IntegerNode)
	if !ok {
		return -1, false
	}

	return i.value, true
}

func (kv *KeyValueNode) String() (string, bool) {
	s, ok := kv.value.(*StringNode)
	if !ok {
		return "", false
	}

	return s.value, true
}

func (kv *KeyValueNode) Bool() (bool, bool) {
	b, ok := kv.value.(*BooleanNode)
	if !ok {
		return false, false
	}

	return b.value, true
}

func (kv *KeyValueNode) Float() (float64, bool) {
	f, ok := kv.value.(*FloatNode)
	if !ok {
		return 0.0, false
	}

	return f.value, false
}

func (kv *KeyValueNode) Set(value any) error {
	switch v := value.(type) {
	case int:
		return kv.Set(int64(v))
	case int8:
		return kv.Set(int64(v))
	case int32:
		return kv.Set(int64(v))
	case int64:
		kv.value = &IntegerNode{
			value: v,
			token: token{
				Type:    INTEGER,
				Lexeme:  strconv.FormatInt(v, 10),
				Literal: v,
				Line:    kv.value.GetToken().Line,
				Column:  kv.value.GetToken().Column,
			},
		}
	case string:
		kv.value = &StringNode{
			value: v,
			token: token{
				Type:    BASIC_STRING,
				Lexeme:  `"` + v + `"`,
				Literal: v,
				Line:    kv.value.GetToken().Line,
				Column:  kv.value.GetToken().Column,
			},
		}
	case float32:
		return kv.Set(float64(v))
	case float64:
		kv.value = &FloatNode{
			value: v,
			token: token{
				Type:    FLOAT,
				Lexeme:  strconv.FormatFloat(v, 'f', -1, 64),
				Literal: v,
				Line:    kv.value.GetToken().Line,
				Column:  kv.value.GetToken().Column,
			},
		}
	case bool:
		lexeme := "false"
		nodeType := FALSE
		if v {
			nodeType = TRUE
			lexeme = "true"
		}
		kv.value = &BooleanNode{
			value: v,
			token: token{
				Type:    nodeType,
				Lexeme:  lexeme,
				Literal: v,
				Line:    kv.value.GetToken().Line,
				Column:  kv.value.GetToken().Column,
			},
		}
	default:
		return fmt.Errorf("tast: unsupported value type %T", value)
	}

	return nil
}

type StringNode struct {
	value string
	token token
}

func (n *StringNode) Accept(v Visitor) error {
	return v.VisitStringNode(n)
}

func (n *StringNode) GetToken() token {
	return n.token
}

type IntegerNode struct {
	value int64
	token token
}

func (n *IntegerNode) Accept(v Visitor) error {
	return v.VisitIntegerNode(n)
}

func (n *IntegerNode) GetToken() token {
	return n.token
}

type FloatNode struct {
	value float64
	token token
}

func (n *FloatNode) Accept(v Visitor) error {
	return v.VisitFloatNode(n)
}

func (n *FloatNode) GetToken() token {
	return n.token
}

type BooleanNode struct {
	value bool
	token token
}

func (n *BooleanNode) Accept(v Visitor) error {
	return v.VisitBooleanNode(n)
}

func (n *BooleanNode) GetToken() token {
	return n.token
}

func KeysMatch(key string, segments []string) bool {
	keyArr := strings.Split(key, STRING_SEPARATOR)
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
