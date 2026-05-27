package tast

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"
)

type Parser struct {
	Tokens  []Token
	current int

	errors []ParseError
	keys   map[string]struct{}
}

func NewParser(tokens []Token) *Parser {
	return &Parser{
		Tokens: tokens,
	}
}

func (p *Parser) registerKey(key *KeyNode) error {
	literalKey := strings.Join(key.Segments, ".")
	if _, exists := p.keys[literalKey]; exists {
		msg := fmt.Sprintf("Duplicate key exists with signature %s", literalKey)
		p.addParseErrorNoTokens(msg, ErrDuplicateKey)
		return errors.New(msg)
	}

	p.keys[literalKey] = struct{}{}
	return nil
}

func (p *Parser) parse() (*Document, []ParseError) {
	documentNode := Document{}
	for !p.isAtEnd() {
		node := p.parseEntry()
		if node != nil {
			documentNode.Content = append(documentNode.Content, node)
		}
	}

	return &documentNode, p.errors
}

func (p *Parser) Synchronize() {
	p.advance() // skip the current token

	for !p.isAtEnd() {
		if p.previous().Type == NEW_LINE {
			return
		}

		p.advance()
	}
}

func (p *Parser) addParseErrorNoTokens(msg string, code ParseErrorCode) {
	p.errors = append(p.errors, ParseError{Message: msg, Code: code})
}

func (p *Parser) addParseError(token Token, msg string, code ParseErrorCode) {
	p.errors = append(p.errors, ParseError{Token: token, Message: msg, Code: code})
}

func (p *Parser) parseEntry() Node {
	switch {
	case p.Match(LEFT_BRACKET):
		node := p.Table()
		if node == nil {
			p.Synchronize()
		}
		return node
	case p.Match(BARE_KEY, BASIC_STRING):
		node := p.KeyValue()
		if node == nil {
			p.Synchronize()
		}
		return node
	default:
		p.advance()
		return nil
	}
}

func (p *Parser) KeyValue() *KeyValueNode {
	key := p.Key()
	if key == nil {
		return nil
	}

	if err := p.registerKey(key); err != nil {
		return nil
	}

	if !p.Match(EQUAL) {
		p.addParseError(p.peek(), "expecting assignment operator '=' after key", ErrMissingAssignmentAfterKey)
		return nil
	}

	value := p.value()
	if value == nil {
		return nil
	}

	return &KeyValueNode{
		Key:   key,
		Value: value,
	}
}

func (p *Parser) value() Node {
	if p.Match(MINUS, PLUS) {
		operator := p.previous().Type
		switch {
		case p.Match(FLOAT):
			return createFloatNode(p, operator)
		case p.Match(INTEGER):
			return createIntNode(p, operator)
		case p.Match(INF):
			return createInfinityNode(p, operator)
		default:
			p.addParseError(p.peek(), "Unable to recognize token that follows -/+", ErrUnrecognizedToken)
			return nil
		}
	}

	switch {
	case p.Match(FLOAT):
		return createFloatNode(p, 0)
	case p.Match(INTEGER):
		return createIntNode(p, 0)
	case p.Match(FALSE):
		return createBoolNode(p, FALSE)
	case p.Match(TRUE):
		return createBoolNode(p, TRUE)
	case p.Match(INF):
		return createInfinityNode(p, 0)
	case p.Match(BASIC_STRING, MULTILINE_BASIC_STRING):
		return createStringNode(p)
	default:
	}

	return nil
}

func createStringNode(p *Parser) Node {
	val, ok := p.previous().Literal.(string)
	if !ok {
		p.addParseError(p.peek(), "Unable to parse value as string", ErrParsingString)
		return nil
	}

	return &StringNode{
		Value: val,
		Token: p.previous(),
	}
}

func createBoolNode(p *Parser, b TokenType) Node {
	return &BooleanNode{
		Value: b == TRUE,
		Token: p.previous(),
	}
}

func createInfinityNode(p *Parser, operator TokenType) Node {
	val := math.MaxInt64
	if operator == MINUS {
		val = -val
	}

	return &IntegerNode{
		Value: int64(val),
		Token: p.previous(),
	}
}

func createIntNode(p *Parser, operator TokenType) Node {
	val, ok := p.previous().Literal.(int64)
	if !ok {
		p.addParseError(p.peek(), "Unable to parse value as int64", ErrParsingInt)
		return nil
	}

	if operator == MINUS {
		val = -val
	}

	return &IntegerNode{
		Value: val,
		Token: p.previous(),
	}
}

func createFloatNode(p *Parser, operator TokenType) Node {
	val, ok := p.previous().Literal.(float64)
	if !ok {
		p.addParseError(p.peek(), "Unable to parse value to float64", ErrParsingFloat)
		return nil
	}

	if operator == MINUS {
		val = -val
	}

	return &FloatNode{
		Value: val,
		Token: p.previous(),
	}
}

// Parse a TOML table which follows the grammar rule:
// table -> LEFT_BRACKET  RIGHT_BRACKET
func (p *Parser) Table() *TableNode {
	if !p.Match(BARE_KEY, BASIC_STRING) {
		p.addParseError(p.peek(), "Expected a key after left-bracket", ErrMalformedTableKey)
		return nil
	}

	key := p.Key()

	if !p.Match(RIGHT_BRACKET) {
		p.addParseError(p.peek(), "Expecting closing bracket ']' after key definition", ErrMissingClosingBracket)
		return nil
	}

	outer := p.keys
	p.keys = make(map[string]struct{})
	defer func() {
		p.keys = outer
	}()

	var children []Node
	for !p.isAtEnd() && !p.check(LEFT_BRACKET) {
		kv := p.KeyValue()
		if kv != nil {
			children = append(children, kv)
		}
	}

	return &TableNode{
		Key:      key,
		Children: children,
	}
}

// Parse a TOML key which follows the grammar rule:
// key -> (BARE_KEY | STRING) (DOT (BARE_KEY | STRING))*
func (p *Parser) Key() *KeyNode {
	curr := p.previous()

	literal, ok := curr.Literal.(string)
	if !ok {
		p.addParseError(p.peek(), fmt.Sprintf("Unable to convert token literal %s to string", curr.Lexeme), ErrParsingString)
		return nil
	}

	node := &KeyNode{
		Segments: []string{literal},
		Tokens:   []Token{curr},
	}

	for p.Match(DOT) {
		if !p.Match(BASIC_STRING, BARE_KEY) {
			p.addParseError(p.peek(), "Expected string or bare key after dot '.'", ErrNoKeyAfterDot)
			return nil
		}

		segment := p.previous()
		node.Segments = append(node.Segments, segment.Literal.(string))
		node.Tokens = append(node.Tokens, segment)
	}

	return node
}

func (p *Parser) Match(types ...TokenType) bool {
	if slices.ContainsFunc(types, p.check) {
		p.advance()
		return true
	}

	return false
}

// Checks the current token for the specific type without consuming it
func (p *Parser) check(token TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peek().Type == token
}

func (p *Parser) advance() Token {
	if !p.isAtEnd() {
		p.current++
	}
	return p.previous()
}

func (p *Parser) peek() Token {
	return p.Tokens[p.current]
}

func (p *Parser) isAtEnd() bool {
	return p.peek().Type == EOF
}

func (p *Parser) previous() Token {
	return p.Tokens[p.current-1]
}
