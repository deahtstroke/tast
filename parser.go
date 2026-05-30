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

	errors        []ParseError
	keys          map[string]struct{}
	pendingTrivia []Trivia
}

func NewParser(tokens []Token) *Parser {
	return &Parser{
		Tokens: tokens,
		keys:   make(map[string]struct{}),
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
	document := &Document{}
	for !p.isAtEnd() {
		node := p.parseEntry()
		if node != nil {
			document.Content = append(document.Content, node)
		}
	}

	p.handleOrphanedTrivia(document)

	return document, p.errors
}

func (p *Parser) handleOrphanedTrivia(document *Document) {
	if p.pendingTrivia != nil {
		lastNode := document.Content[len(document.Content)-1]
		switch n := lastNode.(type) {
		case *KeyValueNode:
			n.TrailingTrivia = p.pendingTrivia
		case *TableNode:
			n.TrailingComments = p.pendingTrivia
		default:
		}
	}
}

func (p *Parser) parseEntry() Node {
	// Accumulate all trivia before appending to next node
	leading := p.getLeadingTrivia()
	switch {
	case p.Match(LEFT_BRACKET):
		node := p.Table()
		if node == nil {
			p.synchronize()
			return nil
		}

		node.LeadingTrivia = leading
		return node
	case p.Match(BARE_KEY, BASIC_STRING, LITERAL_STRING):
		node := p.KeyValue()
		if node == nil {
			p.synchronize()
			return nil
		}

		node.LeadingTrivia = leading
		return node
	default:
		p.advance()
		return nil
	}
}

func (p *Parser) getLeadingTrivia() []Trivia {
	var trivia []Trivia

	// Check first if there's orphaned trivia to be
	// processed from earlier tables
	if p.pendingTrivia != nil {
		trivia = p.pendingTrivia
		p.pendingTrivia = nil
		return trivia
	}
	for !p.isAtEnd() {
		if p.check(NEW_LINE) {
			p.advance()
			continue
		}

		if p.check(COMMENT) {
			comment := p.advance()
			trivia = append(trivia, Trivia{Lexeme: comment.Lexeme})
			continue
		}

		break
	}
	return trivia
}

func (p *Parser) synchronize() {
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

func (p *Parser) KeyValue() *KeyValueNode {
	keyValueNode := &KeyValueNode{}
	key := p.Key()
	if key == nil {
		return nil
	}
	keyValueNode.Key = key

	if err := p.registerKey(key); err != nil {
		return nil
	}

	if !p.Match(EQUAL) {
		p.addParseError(p.peek(), "expecting assignment operator '=' after key", ErrMissingAssignmentAfterKey)
		return nil
	}

	value := p.value()
	if value == nil {
		p.addParseError(p.peek(), "unspecified value for after key", ErrUnspecifiedValueForKey)
		return nil
	}
	keyValueNode.Value = value

	if p.check(COMMENT) {
		comment := p.advance()
		keyValueNode.LineTrivia = &Trivia{Lexeme: comment.Lexeme}
	}

	return keyValueNode
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
	tableNode := &TableNode{}
	if !p.Match(BARE_KEY, BASIC_STRING) {
		p.addParseError(p.peek(), "Expected a key after left-bracket", ErrMalformedTableKey)
		return nil
	}

	key := p.Key()
	if key == nil {
		return nil
	}
	tableNode.Key = key

	if !p.Match(RIGHT_BRACKET) {
		p.addParseError(p.peek(), "Expecting closing bracket ']' after key definition", ErrMissingClosingBracket)
		return nil
	}

	// Trailing comment
	if p.check(COMMENT) {
		comment := p.advance()
		tableNode.LineComment = &Trivia{Lexeme: comment.Lexeme}
	}

	// TODO: Come back to duplicate keys on tables
	// outer := p.keys
	// p.keys = make(map[string]struct{})
	// defer func() {
	// 	p.keys = outer
	// }()

	var children []Node
	var pendingTrivia []Trivia
	for !p.isAtEnd() && !p.check(LEFT_BRACKET) {

		if p.check(NEW_LINE) {
			p.advance()
			continue
		}

		pendingTrivia = p.getLeadingTrivia()
		if p.isAtEnd() || p.check(LEFT_BRACKET) {
			break
		}
		// TODO: Removed here a check for the next table and
		// end of the document itself

		// Consume the first part of the key-value
		if p.Match(BARE_KEY, BASIC_STRING, LITERAL_STRING) {
			kv := p.KeyValue()

			if kv != nil {
				kv.LeadingTrivia = pendingTrivia
				pendingTrivia = nil

				children = append(children, kv)
			} else {
				p.synchronize()
				break
			}
		}
	}

	// Buffer the pending trivia if there's any
	p.pendingTrivia = pendingTrivia

	tableNode.Children = children
	return tableNode
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
		if !p.Match(BASIC_STRING, BARE_KEY, LITERAL_STRING) {
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
