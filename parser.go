package tast

import (
	"math"
	"slices"

	"github.com/deahtstroke/toml-ast/token"
)

type Parser struct {
	Tokens  []token.Token
	current int

	errors []ParseError
	keys   map[string]bool
}

func NewParser(tokens []token.Token) *Parser {
	return &Parser{
		Tokens: tokens,
	}
}

// Starts parsing tokens, reports any errors if
// the length of errors is non-zero
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
		if p.previous().Type == token.NEW_LINE {
			return
		}

		p.advance()
	}
}

func (p *Parser) addParseError(token token.Token, msg string, code ParseErrorCode) {
	p.errors = append(p.errors, ParseError{Token: token, Message: msg, Code: code})
}

func (p *Parser) parseEntry() Node {
	switch {
	case p.Match(token.LEFT_BRACKET):
		node := p.Table()
		if node == nil {
			p.Synchronize()
		}
		return node
	case p.Match(token.BARE_KEY, token.BASIC_STRING):
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

	if !p.Match(token.EQUAL) {
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
	if p.Match(token.MINUS, token.PLUS) {
		operator := p.previous().Type
		switch {
		case p.Match(token.FLOAT):
			return createFloatNode(p, operator)
		case p.Match(token.INTEGER):
			return createIntNode(p, operator)
		case p.Match(token.INF):
			return createInfinityNode(p, operator)
		default:
			p.addParseError(p.peek(), "Unable to recognize token that follows -/+", ErrUnrecognizedToken)
			return nil
		}
	}

	switch {
	case p.Match(token.FLOAT):
		return createFloatNode(p, 0)
	case p.Match(token.INTEGER):
		return createIntNode(p, 0)
	case p.Match(token.FALSE):
		return createBoolNode(p, token.FALSE)
	case p.Match(token.TRUE):
		return createBoolNode(p, token.TRUE)
	case p.Match(token.INF):
		return createInfinityNode(p, 0)
	case p.Match(token.BASIC_STRING, token.MULTILINE_BASIC_STRING):
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

func createBoolNode(p *Parser, b token.TokenType) Node {
	return &BooleanNode{
		Value: b == token.TRUE,
		Token: p.previous(),
	}
}

func createInfinityNode(p *Parser, operator token.TokenType) Node {
	val := math.MaxInt64
	if operator == token.MINUS {
		val = -val
	}

	return &IntegerNode{
		Value: int64(val),
		Token: p.previous(),
	}
}

func createIntNode(p *Parser, operator token.TokenType) Node {
	val, ok := p.previous().Literal.(int64)
	if !ok {
		p.addParseError(p.peek(), "Unable to parse value as int64", ErrParsingInt)
		return nil
	}

	if operator == token.MINUS {
		val = -val
	}

	return &IntegerNode{
		Value: val,
		Token: p.previous(),
	}
}

func createFloatNode(p *Parser, operator token.TokenType) Node {
	val, ok := p.previous().Literal.(float64)
	if !ok {
		p.addParseError(p.peek(), "Unable to parse value to float64", ErrParsingFloat)
		return nil
	}

	if operator == token.MINUS {
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
	if !p.Match(token.BARE_KEY, token.BASIC_STRING) {
		p.addParseError(p.peek(), "expected a key after left-bracket", ErrMalformedTableKey)
		return nil
	}

	key := p.Key()

	if !p.Match(token.RIGHT_BRACKET) {
		p.addParseError(p.peek(), "Expecting closing bracket ']' after key definition", ErrMissingClosingBracket)
		return nil
	}

	return &TableNode{
		Key: key,
	}
}

// Parse a TOML key which follows the grammar rule:
// key -> (BARE_KEY | STRING) (DOT (BARE_KEY | STRING))*
func (p *Parser) Key() *KeyNode {
	curr := p.previous()
	node := &KeyNode{
		Segments: []string{curr.Lexeme},
		Tokens:   []token.Token{curr},
	}

	for p.Match(token.DOT) {
		if !p.Match(token.BASIC_STRING, token.BARE_KEY) {
			p.addParseError(p.peek(), "expected string or barekey after dot '.'", ErrNoKeyAfterDot)
			return nil
		}

		segment := p.previous()
		node.Segments = append(node.Segments, segment.Literal.(string))
		node.Tokens = append(node.Tokens, segment)
	}

	return node
}

func (p *Parser) Match(types ...token.TokenType) bool {
	if slices.ContainsFunc(types, p.check) {
		p.advance()
		return true
	}

	return false
}

func (p *Parser) check(token token.TokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peek().Type == token
}

func (p *Parser) advance() token.Token {
	if !p.isAtEnd() {
		p.current++
	}
	return p.previous()
}

func (p *Parser) peek() token.Token {
	return p.Tokens[p.current]
}

func (p *Parser) isAtEnd() bool {
	return p.peek().Type == token.EOF
}

func (p *Parser) previous() token.Token {
	return p.Tokens[p.current-1]
}
