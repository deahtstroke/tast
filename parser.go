package tast

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"
)

type Parser struct {
	tokens  []token
	current int

	errors        []ParseError
	keys          map[string]struct{}
	tables        map[string]struct{}
	pendingTrivia []Trivia
}

func NewParser(tokens []token) *Parser {
	return &Parser{
		tokens: tokens,
		keys:   make(map[string]struct{}),
		tables: make(map[string]struct{}),
	}
}

// Parse() starts the parsing process of iterating the tokens assigned
// to the parser struct and create a TOML-Document
// In the case where there are errors while parsing, the accumulated errors
// will be returned in the slice of ParseErrors
func (p *Parser) Parse() (*Document, []ParseError) {
	document := &Document{}
	for !p.isAtEnd() {
		node := p.nextNode()
		if node != nil {
			document.content = append(document.content, node)
		}
	}

	p.handleOrphanedTrivia(document)
	return document, p.errors
}

func (p *Parser) registerTable(table *KeyNode) error {
	tableKey := strings.Join(table.segments, ".")
	if _, exists := p.tables[tableKey]; exists {
		msg := fmt.Sprintf("Duplicate table exists with signature %s", tableKey)
		p.addErrorNoToken(msg, ErrDuplicateTable)
	}

	p.tables[tableKey] = struct{}{}
	return nil
}

func (p *Parser) registerKey(key *KeyNode) error {
	literalKey := strings.Join(key.segments, ".")
	if _, exists := p.keys[literalKey]; exists {
		msg := fmt.Sprintf("Duplicate key exists with signature %s", literalKey)
		p.addErrorNoToken(msg, ErrDuplicateKey)
		return errors.New(msg)
	}

	p.keys[literalKey] = struct{}{}
	return nil
}

func (p *Parser) handleOrphanedTrivia(document *Document) {
	if p.pendingTrivia != nil {
		lastNode := document.content[len(document.content)-1]
		switch n := lastNode.(type) {
		case *KeyValueNode:
			n.trailingTrivia = p.pendingTrivia
		case *TableNode:
			n.trailingTrivia = p.pendingTrivia
		default:
		}
	}
}

func (p *Parser) nextNode() Node {
	// Accumulate all trivia before appending to next node
	leading := p.getLeadingTrivia()
	switch {
	case p.MatchAny(LEFT_BRACKET):
		node := p.Table()
		if node == nil {
			p.synchronize()
			return nil
		}

		node.leadingTrivia = leading
		return node
	case p.MatchAny(BARE_KEY, BASIC_STRING, LITERAL_STRING):
		node := p.KeyValue()
		if node == nil {
			p.synchronize()
			return nil
		}

		node.leadingTrivia = leading
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
			newLine := p.advance()
			trivia = append(trivia, Trivia{lexeme: newLine.Lexeme, Type: NewLineTrivia})
			continue
		}

		if p.check(COMMENT) {
			comment := p.advance()
			trivia = append(trivia, Trivia{lexeme: comment.Lexeme, Type: CommentTrivia})
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

func (p *Parser) addError(token token, msg string, code ParseErrorCode) {
	p.errors = append(p.errors, ParseError{Token: token, Message: msg, Code: code})
}

func (p *Parser) addErrorNoToken(msg string, code ParseErrorCode) {
	p.addError(token{}, msg, code)
}

func (p *Parser) KeyValue() *KeyValueNode {
	keyValueNode := &KeyValueNode{}
	key := p.Key()
	if key == nil {
		return nil
	}
	keyValueNode.key = key

	if err := p.registerKey(key); err != nil {
		return nil
	}

	if !p.MatchAny(EQUAL) {
		p.addError(p.peek(), "expecting assignment operator '=' after key", ErrMissingAssignmentAfterKey)
		return nil
	}

	value := p.value()
	if value == nil {
		p.addError(p.peek(), "unspecified value for after key", ErrUnspecifiedValueForKey)
		return nil
	}

	// Check to see if there was a new line after the value
	lineTrivia := p.getLineTrivia()
	hasNewLine := len(lineTrivia) > 0 && lineTrivia[len(lineTrivia)-1].Type == NewLineTrivia
	if !hasNewLine && !p.isAtEnd() {
		p.addError(p.peek(), "expected new line after value", ErrMissingNewLine)
		return nil
	}

	keyValueNode.value = value
	keyValueNode.lineTrivia = lineTrivia

	return keyValueNode
}

func (p *Parser) value() Node {
	if p.MatchAny(MINUS, PLUS) {
		operator := p.previous().Type
		switch {
		case p.MatchAny(FLOAT):
			return createFloatNode(p, operator)
		case p.MatchAny(INTEGER):
			return createIntNode(p, operator)
		case p.MatchAny(INF):
			return createInfinityNode(p, operator)
		default:
			p.addError(p.peek(), "Unable to recognize token that follows -/+", ErrUnrecognizedToken)
			return nil
		}
	}

	switch {
	case p.MatchAny(FLOAT):
		return createFloatNode(p, 0)
	case p.MatchAny(INTEGER):
		return createIntNode(p, 0)
	case p.MatchAny(FALSE):
		return createBoolNode(p, FALSE)
	case p.MatchAny(TRUE):
		return createBoolNode(p, TRUE)
	case p.MatchAny(INF):
		return createInfinityNode(p, 0)
	case p.MatchAny(BASIC_STRING, MULTILINE_BASIC_STRING):
		return createStringNode(p)
	default:
	}

	return nil
}

func createStringNode(p *Parser) Node {
	val, ok := p.previous().Literal.(string)
	if !ok {
		p.addError(p.peek(), "Unable to parse value as string", ErrParsingString)
		return nil
	}

	return &StringNode{
		value: val,
		token: p.previous(),
	}
}

func createBoolNode(p *Parser, b TokenType) Node {
	return &BooleanNode{
		value: b == TRUE,
		token: p.previous(),
	}
}

func createInfinityNode(p *Parser, operator TokenType) Node {
	val := math.Inf(1)
	if operator == MINUS {
		val = math.Inf(-1)
	}

	// According to IEEE 754, inf should be treated as float64
	// aka, a FloatNode in tast-speak
	return &FloatNode{
		value: val,
		token: p.previous(),
	}
}

func createIntNode(p *Parser, operator TokenType) Node {
	val, ok := p.previous().Literal.(int64)
	if !ok {
		p.addError(p.peek(), "Unable to parse value as int64", ErrParsingInt)
		return nil
	}

	if operator == MINUS {
		val = -val
	}

	return &IntegerNode{
		value: val,
		token: p.previous(),
	}
}

func createFloatNode(p *Parser, operator TokenType) Node {
	val, ok := p.previous().Literal.(float64)
	if !ok {
		p.addError(p.peek(), "Unable to parse value to float64", ErrParsingFloat)
		return nil
	}

	if operator == MINUS {
		val = -val
	}

	return &FloatNode{
		value: val,
		token: p.previous(),
	}
}

// Parse a TOML table which follows the grammar rule:
// table -> LEFT_BRACKET  RIGHT_BRACKET
func (p *Parser) Table() *TableNode {
	tableNode := &TableNode{}
	if !p.MatchAny(BARE_KEY, BASIC_STRING) {
		p.addError(p.peek(), "Expected a key after left-bracket", ErrMalformedTableKey)
		return nil
	}

	key := p.Key()
	if key == nil {
		return nil
	}

	if err := p.registerTable(key); err != nil {
		return nil
	}
	tableNode.key = key

	if !p.MatchAny(RIGHT_BRACKET) {
		p.addError(p.peek(), "Expecting closing bracket ']' after key definition", ErrMissingClosingBracket)
		return nil
	}

	// Check for in-line trivia
	// Break out of loop once we find newline or there are none, in that case it'll error
	lineTrivia := p.getLineTrivia()
	hasNewLine := len(lineTrivia) > 0 && lineTrivia[len(lineTrivia)-1].Type == NewLineTrivia
	if !hasNewLine && !p.isAtEnd() {
		p.addError(p.peek(), "Expected newline character after table header", ErrMissingNewLine)
		return nil
	}

	tableNode.lineTrivia = lineTrivia

	outer := p.keys
	p.keys = make(map[string]struct{})
	defer func() {
		p.keys = outer
	}()

	var children []Node
	var pendingTrivia []Trivia

	// Process children KV nodes until next table
	for !p.isAtEnd() && !p.check(LEFT_BRACKET) {
		pendingTrivia = p.getLeadingTrivia()
		if p.isAtEnd() || p.check(LEFT_BRACKET) {
			break
		}

		// Consume the first part of the key-value
		if p.MatchAny(BARE_KEY, BASIC_STRING, LITERAL_STRING) {
			kv := p.KeyValue()

			if kv != nil {
				kv.leadingTrivia = pendingTrivia
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

	tableNode.children = children
	return tableNode
}

func (p *Parser) getLineTrivia() []Trivia {
	var lineTrivia []Trivia
	for !p.isAtEnd() {
		if p.check(COMMENT) {
			comment := p.advance()
			lineTrivia = append(lineTrivia, Trivia{lexeme: comment.Lexeme, Type: CommentTrivia})
			continue
		}

		if p.check(NEW_LINE) {
			newLine := p.advance()
			lineTrivia = append(lineTrivia, Trivia{lexeme: newLine.Lexeme, Type: NewLineTrivia})
			break
		}

		break
	}
	return lineTrivia
}

// Parse a TOML key which follows the grammar rule:
// key -> (BARE_KEY | STRING) (DOT (BARE_KEY | STRING))*
func (p *Parser) Key() *KeyNode {
	curr := p.previous()

	literal, ok := curr.Literal.(string)
	if !ok {
		p.addError(p.peek(), fmt.Sprintf("Unable to convert token literal %s to string", curr.Lexeme), ErrParsingString)
		return nil
	}

	node := &KeyNode{
		segments: []string{literal},
		tokens:   []token{curr},
	}

	for p.MatchAny(DOT) {
		if !p.MatchAny(BASIC_STRING, BARE_KEY, LITERAL_STRING) {
			p.addError(p.peek(), "Expected string or bare key after dot '.'", ErrNoKeyAfterDot)
			return nil
		}

		segment := p.previous()
		node.segments = append(node.segments, segment.Literal.(string))
		node.tokens = append(node.tokens, segment)
	}

	return node
}

func (p *Parser) MatchAny(types ...TokenType) bool {
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

func (p *Parser) advance() token {
	if !p.isAtEnd() {
		p.current++
	}
	return p.previous()
}

func (p *Parser) peek() token {
	return p.tokens[p.current]
}

func (p *Parser) isAtEnd() bool {
	return p.peek().Type == EOF
}

func (p *Parser) previous() token {
	return p.tokens[p.current-1]
}
