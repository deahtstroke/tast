package tast

import (
	"fmt"
	"strconv"
	"strings"
)

const stringSeparator = "."

type visitor interface {
	visitTableNode(*TableNode) error
	visitKeyValueNode(*KeyValueNode) error
	visitKeyNode(*keyNode) error
	visitStringNode(*stringNode) error
	visitIntegerNode(*integerNode) error
	visitFloatNode(*floatNode) error
	visitBooleanNode(*booleanNode) error
}

// The node interface represents all the shared methods between all different
// types of Nodes specified in the TOML spec
type node interface {
	accept(visitor) error
	getToken() token
}

// Document represents the top-most node of a TOML parse-tree
//
// This is where most operations will be done like finding a table,
// mutating a key-value pair, or fetching a value
type Document struct {
	content []node
}

// FindKey() retrieves a key-value node defined by the 'key' parameter
// The 'key' parameter can be defined as a bare key or a dotted-key
// in accordance to the TOML spec
func (d *Document) FindKey(key string) (*KeyValueNode, bool) {
	for _, node := range d.content {
		kv, ok := node.(*KeyValueNode)
		if ok && keysMatch(key, kv.key.segments) {
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
		if ok && keysMatch(key, t.key.segments) {
			return t, true
		}
	}
	return nil, false
}

// Returns a string representation of the document
func (d *Document) String() (string, error) {
	p := newPrinter()
	return p.print(d)
}

type triviaType int

const (
	newLineTrivia triviaType = iota
	tabTrivia
	commentTrivia
)

var triviaTypes map[triviaType]string = map[triviaType]string{
	newLineTrivia: "NewLineTrivia",
	tabTrivia:     "TabTrivia",
	commentTrivia: "CommentTrivia",
}

// Trivia represents any source characters that are not part of the
// semantic value of the document — comments, newlines, whitespace,
// and other non-meaningful tokens that are preserved for
// source-faithful round-tripping
type trivia struct {
	Type   triviaType
	Lexeme string
}

// Node reprsentation of a TOML table as specified in
// https://toml.io/en/v1.1.0#table
type TableNode struct {
	// The key of the TableNode
	key *keyNode

	// Child nodes that belong to this table
	children []node

	// Any leading comments that may come before the
	// table itself
	leadingTrivia []trivia

	// Comment in the same line as the table
	lineTrivia []trivia

	// Any comments that are leftover after the table
	// itself
	trailingTrivia []trivia

	// tokens that make up the table-key
	tokens []token

	// if table was created from a dotted key
	isImplicit bool
}

func (n *TableNode) accept(v visitor) error {
	return v.visitTableNode(n)
}

func (n *TableNode) getToken() token {
	return n.tokens[0]
}

// Key() return the string representation of a TOML key
// TOML keys according to the spec., can be bare-keys, basic strings,
// or a combination of both
func (n *TableNode) Key() string {
	return n.key.key()
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
		kv.value = &integerNode{
			value: v,
			token: token{
				Type:    integer,
				Lexeme:  strconv.FormatInt(v, 10),
				Literal: v,
				Line:    kv.value.getToken().Line,
				Column:  kv.value.getToken().Column,
			},
		}
	case string:
		kv.value = &stringNode{
			value: v,
			token: token{
				Type:    basicString,
				Lexeme:  `"` + v + `"`,
				Literal: v,
				Line:    kv.value.getToken().Line,
				Column:  kv.value.getToken().Column,
			},
		}
	case float32:
		return kv.Set(float64(v))
	case float64:
		kv.value = &floatNode{
			value: v,
			token: token{
				Type:    floatPoint,
				Lexeme:  strconv.FormatFloat(v, 'f', -1, 64),
				Literal: v,
				Line:    kv.value.getToken().Line,
				Column:  kv.value.getToken().Column,
			},
		}
	case bool:
		lexeme := "false"
		if v {
			lexeme = "true"
		}
		kv.value = &booleanNode{
			value: v,
			token: token{
				Type:    boolean,
				Lexeme:  lexeme,
				Literal: v,
				Line:    kv.value.getToken().Line,
				Column:  kv.value.getToken().Column,
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
		if kv, ok := child.(*KeyValueNode); ok && keysMatch(key, kv.key.segments) {
			return i
		}
	}

	return -1
}

// Node representation of a Key in the TOML specification
// https://toml.io/en/v1.1.0#keys
type keyNode struct {
	// segments that make up the Key
	// KeyNodes can be made up of bare-keys and quoted-keys
	segments []string

	// List of scanner tokens that make up this KeyNode
	tokens []token
}

func (n *keyNode) key() string {
	if len(n.segments) == 0 {
		return ""
	}

	return strings.Join(n.segments, stringSeparator)
}

func (n *keyNode) accept(v visitor) error {
	return v.visitKeyNode(n)
}

func (n *keyNode) getToken() token {
	return n.tokens[0]
}

// Node representation of a Key-Value pair in the TOML specification
// https://toml.io/en/v1.1.0#keyvalue-pair
type KeyValueNode struct {
	// key identifier that represents this key-value pair
	key *keyNode

	// value that this key-value pair has
	value node

	// List of scanner tokens that make up this key-value pair
	tokens []token

	// Any leading comments that may come before a key-value node
	leadingTrivia []trivia

	// Comment at the end of the line in a key-value node
	lineTrivia []trivia

	// Any comments at the end of a key-value node
	// Note: This field only gets filled if the only tokens leftover
	// after the dangling trivia are New lines and/or EOF
	trailingTrivia []trivia
}

func (kv *KeyValueNode) accept(v visitor) error {
	return v.visitKeyValueNode(kv)
}

func (kv *KeyValueNode) getToken() token {
	return kv.tokens[0]
}

func (kv *KeyValueNode) Int() (int64, bool) {
	i, ok := kv.value.(*integerNode)
	if !ok {
		return -1, false
	}

	return i.value, true
}

func (kv *KeyValueNode) String() (string, bool) {
	s, ok := kv.value.(*stringNode)
	if !ok {
		return "", false
	}

	return s.value, true
}

func (kv *KeyValueNode) Bool() (bool, bool) {
	b, ok := kv.value.(*booleanNode)
	if !ok {
		return false, false
	}

	return b.value, true
}

func (kv *KeyValueNode) Float() (float64, bool) {
	f, ok := kv.value.(*floatNode)
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
		kv.value = &integerNode{
			value: v,
			token: token{
				Type:    integer,
				Lexeme:  strconv.FormatInt(v, 10),
				Literal: v,
				Line:    kv.value.getToken().Line,
				Column:  kv.value.getToken().Column,
			},
		}
	case string:
		kv.value = &stringNode{
			value: v,
			token: token{
				Type:    basicString,
				Lexeme:  `"` + v + `"`,
				Literal: v,
				Line:    kv.value.getToken().Line,
				Column:  kv.value.getToken().Column,
			},
		}
	case float32:
		return kv.Set(float64(v))
	case float64:
		kv.value = &floatNode{
			value: v,
			token: token{
				Type:    floatPoint,
				Lexeme:  strconv.FormatFloat(v, 'f', -1, 64),
				Literal: v,
				Line:    kv.value.getToken().Line,
				Column:  kv.value.getToken().Column,
			},
		}
	case bool:
		lexeme := "false"
		if v {
			lexeme = "true"
		}
		kv.value = &booleanNode{
			value: v,
			token: token{
				Type:    boolean,
				Lexeme:  lexeme,
				Literal: v,
				Line:    kv.value.getToken().Line,
				Column:  kv.value.getToken().Column,
			},
		}
	default:
		return fmt.Errorf("tast: unsupported value type %T", value)
	}

	return nil
}

type stringNode struct {
	value string
	token token
}

func (n *stringNode) accept(v visitor) error {
	return v.visitStringNode(n)
}

func (n *stringNode) getToken() token {
	return n.token
}

type integerNode struct {
	value int64
	token token
}

func (n *integerNode) accept(v visitor) error {
	return v.visitIntegerNode(n)
}

func (n *integerNode) getToken() token {
	return n.token
}

type floatNode struct {
	value float64
	token token
}

func (n *floatNode) accept(v visitor) error {
	return v.visitFloatNode(n)
}

func (n *floatNode) getToken() token {
	return n.token
}

type booleanNode struct {
	value bool
	token token
}

func (n *booleanNode) accept(v visitor) error {
	return v.visitBooleanNode(n)
}

func (n *booleanNode) getToken() token {
	return n.token
}

func keysMatch(key string, segments []string) bool {
	keyArr := strings.Split(key, stringSeparator)
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
