package tast

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"
)

type parser struct {
	tokens  []token
	current int

	errors        []parseError
	keys          map[string]struct{}
	tables        map[string]struct{}
	pendingTrivia []trivia
}

func newParser(tokens []token) *parser {
	return &parser{
		tokens: tokens,
		keys:   make(map[string]struct{}),
		tables: make(map[string]struct{}),
	}
}

// parse() starts the parsing process of iterating the tokens assigned
// to the parse struct and create a TOML-Document
// In the case where there are errors while parsing, the accumulated errors
// will be returned in the slice of ParseErrors
func (p *parser) parse() (*Document, []parseError) {
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

func (p *parser) registerTable(table *keyNode) error {
	tableKey := strings.Join(table.segments, ".")
	if _, exists := p.tables[tableKey]; exists {
		msg := fmt.Sprintf("Duplicate table exists with signature %s", tableKey)
		p.addErrorNoToken(msg, errDuplicateTable)
	}

	p.tables[tableKey] = struct{}{}
	return nil
}

func (p *parser) registerKey(key *keyNode) error {
	literalKey := strings.Join(key.segments, ".")
	if _, exists := p.keys[literalKey]; exists {
		msg := fmt.Sprintf("Duplicate key exists with signature %s", literalKey)
		p.addErrorNoToken(msg, errDuplicateKey)
		return errors.New(msg)
	}

	p.keys[literalKey] = struct{}{}
	return nil
}

func (p *parser) handleOrphanedTrivia(document *Document) {
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

func (p *parser) nextNode() node {
	// Accumulate all trivia before appending to next node
	leading := p.getLeadingTrivia()
	switch {
	case p.matchAny(leftBracket):
		node := p.table()
		if node == nil {
			p.synchronize()
			return nil
		}

		node.leadingTrivia = leading
		return node
	case p.matchAny(bareKey, basicString, literalString):
		node := p.keyValue()
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

func (p *parser) getLeadingTrivia() []trivia {
	var tr []trivia

	// Check first if there's orphaned trivia to be
	// processed from earlier tables
	if p.pendingTrivia != nil {
		tr = p.pendingTrivia
		p.pendingTrivia = nil
		return tr
	}

	for !p.isAtEnd() {
		if p.check(newLine) {
			newLine := p.advance()
			tr = append(tr, trivia{Lexeme: newLine.Lexeme, Type: newLineTrivia})
			continue
		}

		if p.check(comment) {
			comment := p.advance()
			tr = append(tr, trivia{Lexeme: comment.Lexeme, Type: commentTrivia})
			continue
		}

		break
	}
	return tr
}

func (p *parser) synchronize() {
	p.advance() // skip the current token

	for !p.isAtEnd() {
		if p.previous().Type == newLine {
			return
		}

		p.advance()
	}
}

func (p *parser) addError(token token, msg string, code parserErrorCode) {
	p.errors = append(p.errors, parseError{Token: token, Message: msg, Code: code})
}

func (p *parser) addErrorNoToken(msg string, code parserErrorCode) {
	p.addError(token{}, msg, code)
}

func (p *parser) keyValue() *KeyValueNode {
	keyValueNode := &KeyValueNode{}
	key := p.key()
	if key == nil {
		return nil
	}
	keyValueNode.key = key

	if err := p.registerKey(key); err != nil {
		return nil
	}

	if !p.matchAny(equal) {
		p.addError(p.peek(), "expecting assignment operator '=' after key", errMissingAssignmentAfterKey)
		return nil
	}

	value := p.value()
	if value == nil {
		p.addError(p.peek(), "unspecified value for after key", errUnspecifiedValueForKey)
		return nil
	}

	// Check to see if there was a new line after the value
	lineTrivia := p.getLineTrivia()
	hasNewLine := len(lineTrivia) > 0 && lineTrivia[len(lineTrivia)-1].Type == newLineTrivia
	if !hasNewLine && !p.isAtEnd() {
		p.addError(p.peek(), "expected new line after value", errMissingNewLine)
		return nil
	}

	keyValueNode.value = value
	keyValueNode.lineTrivia = lineTrivia

	return keyValueNode
}

func (p *parser) value() node {
	if p.matchAny(minus, plus) {
		operator := p.previous().Type
		switch {
		case p.matchAny(floatPoint):
			return createFloatNode(p, operator)
		case p.matchAny(integer):
			return createIntNode(p, operator)
		case p.matchAny(infinity):
			return createInfinityNode(p, operator)
		default:
			p.addError(p.peek(), "Unable to recognize token that follows -/+", errUnrecognizedToken)
			return nil
		}
	}

	switch {
	case p.matchAny(floatPoint):
		return createFloatNode(p, 0)
	case p.matchAny(integer):
		return createIntNode(p, 0)
	case p.matchAny(boolean):
		return createBooleanNode(p)
	case p.matchAny(infinity):
		return createInfinityNode(p, 0)
	case p.matchAny(basicString, multilineBasicString):
		return createStringNode(p)
	default:
	}

	return nil
}

func createStringNode(p *parser) node {
	val, ok := p.previous().Literal.(string)
	if !ok {
		p.addError(p.peek(), "Unable to parse value as string", errParsingString)
		return nil
	}

	return &stringNode{
		value: val,
		token: p.previous(),
	}
}

func createInfinityNode(p *parser, operator tokenType) node {
	val := math.Inf(1)
	if operator == minus {
		val = math.Inf(-1)
	}

	// According to IEEE 754, inf should be treated as float64
	// aka, a FloatNode in tast-speak
	return &floatNode{
		value: val,
		token: p.previous(),
	}
}

func createIntNode(p *parser, operator tokenType) node {
	val, ok := p.previous().Literal.(int64)
	if !ok {
		p.addError(p.peek(), "Unable to parse value as int64", errParsingInt)
		return nil
	}

	if operator == minus {
		val = -val
	}

	return &integerNode{
		value: val,
		token: p.previous(),
	}
}

func createFloatNode(p *parser, operator tokenType) node {
	val, ok := p.previous().Literal.(float64)
	if !ok {
		p.addError(p.peek(), "Unable to parse value to float64", errParsingFloat)
		return nil
	}

	if operator == minus {
		val = -val
	}

	return &floatNode{
		value: val,
		token: p.previous(),
	}
}

func createBooleanNode(p *parser) node {
	val, ok := p.previous().Literal.(bool)
	if !ok {
		p.addError(p.peek(), "Unable to parse value to bool", errParsingBool)
	}

	return &booleanNode{
		value: val,
		token: p.previous(),
	}
}

// Parse a TOML table which follows the grammar rule:
// table -> LEFT_BRACKET  RIGHT_BRACKET
func (p *parser) table() *TableNode {
	tableNode := &TableNode{}
	if !p.matchAny(bareKey, basicString) {
		p.addError(p.peek(), "Expected a key after left-bracket", errMalformedTableKey)
		return nil
	}

	key := p.key()
	if key == nil {
		return nil
	}

	if err := p.registerTable(key); err != nil {
		return nil
	}
	tableNode.key = key

	if !p.matchAny(rightBracket) {
		p.addError(p.peek(), "Expecting closing bracket ']' after key definition", errMissingClosingBracket)
		return nil
	}

	// Check for in-line trivia
	// Break out of loop once we find newline or there are none, in that case it'll error
	lineTrivia := p.getLineTrivia()
	hasNewLine := len(lineTrivia) > 0 && lineTrivia[len(lineTrivia)-1].Type == newLineTrivia
	if !hasNewLine && !p.isAtEnd() {
		p.addError(p.peek(), "Expected newline character after table header", errMissingNewLine)
		return nil
	}

	tableNode.lineTrivia = lineTrivia

	outer := p.keys
	p.keys = make(map[string]struct{})
	defer func() {
		p.keys = outer
	}()

	var children []node
	var pendingTrivia []trivia

	// Process children KV nodes until next table
	for !p.isAtEnd() && !p.check(leftBracket) {
		pendingTrivia = p.getLeadingTrivia()
		if p.isAtEnd() || p.check(leftBracket) {
			break
		}

		// Consume the first part of the key-value
		if p.matchAny(bareKey, basicString, literalString) {
			kv := p.keyValue()

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

func (p *parser) getLineTrivia() []trivia {
	var lineTrivia []trivia
	for !p.isAtEnd() {
		if p.check(comment) {
			comment := p.advance()
			lineTrivia = append(lineTrivia, trivia{Lexeme: comment.Lexeme, Type: commentTrivia})
			continue
		}

		if p.check(newLine) {
			newLine := p.advance()
			lineTrivia = append(lineTrivia, trivia{Lexeme: newLine.Lexeme, Type: newLineTrivia})
			break
		}

		break
	}
	return lineTrivia
}

// Parse a TOML key which follows the grammar rule:
// key -> (BARE_KEY | STRING) (DOT (BARE_KEY | STRING))*
func (p *parser) key() *keyNode {
	curr := p.previous()

	literal, ok := curr.Literal.(string)
	if !ok {
		p.addError(p.peek(), fmt.Sprintf("Unable to convert token literal %s to string", curr.Lexeme), errParsingString)
		return nil
	}

	node := &keyNode{
		segments: []string{literal},
		tokens:   []token{curr},
	}

	for p.matchAny(dot) {
		if !p.matchAny(basicString, bareKey, literalString) {
			p.addError(p.peek(), "Expected string or bare key after dot '.'", errNoKeyAfterDot)
			return nil
		}

		segment := p.previous()
		node.segments = append(node.segments, segment.Literal.(string))
		node.tokens = append(node.tokens, segment)
	}

	return node
}

func (p *parser) matchAny(types ...tokenType) bool {
	if slices.ContainsFunc(types, p.check) {
		p.advance()
		return true
	}

	return false
}

// Checks the current token for the specific type without consuming it
func (p *parser) check(token tokenType) bool {
	if p.isAtEnd() {
		return false
	}
	return p.peek().Type == token
}

func (p *parser) advance() token {
	if !p.isAtEnd() {
		p.current++
	}
	return p.previous()
}

func (p *parser) peek() token {
	return p.tokens[p.current]
}

func (p *parser) isAtEnd() bool {
	return p.peek().Type == eof
}

func (p *parser) previous() token {
	return p.tokens[p.current-1]
}
