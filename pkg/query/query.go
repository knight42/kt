package query

import (
	"fmt"
	"sort"
	"strings"
	"unicode"
)

type Expr interface {
	Match(line []byte) bool
	Terms() [][]byte
}

type keyword struct {
	term []byte
}

func (k *keyword) Match(line []byte) bool {
	return bytesContainsFold(line, k.term)
}

func (k *keyword) Terms() [][]byte {
	return [][]byte{k.term}
}

type andExpr struct {
	left, right Expr
}

func (a *andExpr) Match(line []byte) bool {
	return a.left.Match(line) && a.right.Match(line)
}

func (a *andExpr) Terms() [][]byte {
	return append(a.left.Terms(), a.right.Terms()...)
}

type orExpr struct {
	left, right Expr
}

func (o *orExpr) Match(line []byte) bool {
	return o.left.Match(line) || o.right.Match(line)
}

func (o *orExpr) Terms() [][]byte {
	return append(o.left.Terms(), o.right.Terms()...)
}

// Parse parses a query DSL string into an Expr.
//
// Grammar:
//
//	expr   = term ("or" term)*
//	term   = factor ("and" factor)*
//	factor = KEYWORD | "(" expr ")"
func Parse(input string) (Expr, error) {
	tokens, err := lex(input)
	if err != nil {
		return nil, err
	}
	if len(tokens) == 0 {
		return nil, fmt.Errorf("empty query")
	}
	p := &parser{tokens: tokens}
	expr, err := p.parseExpr()
	if err != nil {
		return nil, err
	}
	if p.pos < len(p.tokens) {
		return nil, fmt.Errorf("unexpected token: %q", p.tokens[p.pos].value)
	}
	return expr, nil
}

type tokenKind int

const (
	tokenKeyword tokenKind = iota
	tokenAnd
	tokenOr
	tokenLParen
	tokenRParen
)

type token struct {
	kind  tokenKind
	value string
}

func lex(input string) ([]token, error) {
	var tokens []token
	i := 0
	for i < len(input) {
		ch := rune(input[i])
		if unicode.IsSpace(ch) {
			i++
			continue
		}
		if ch == '(' {
			tokens = append(tokens, token{kind: tokenLParen, value: "("})
			i++
			continue
		}
		if ch == ')' {
			tokens = append(tokens, token{kind: tokenRParen, value: ")"})
			i++
			continue
		}
		if ch == '"' {
			end := strings.IndexByte(input[i+1:], '"')
			if end < 0 {
				return nil, fmt.Errorf("unterminated quoted string starting at position %d", i)
			}
			tokens = append(tokens, token{kind: tokenKeyword, value: input[i+1 : i+1+end]})
			i += end + 2
			continue
		}
		start := i
		for i < len(input) && !unicode.IsSpace(rune(input[i])) && input[i] != '(' && input[i] != ')' {
			i++
		}
		word := input[start:i]
		switch strings.ToLower(word) {
		case "and":
			tokens = append(tokens, token{kind: tokenAnd, value: word})
		case "or":
			tokens = append(tokens, token{kind: tokenOr, value: word})
		default:
			tokens = append(tokens, token{kind: tokenKeyword, value: word})
		}
	}
	return tokens, nil
}

type parser struct {
	tokens []token
	pos    int
}

func (p *parser) peek() *token {
	if p.pos >= len(p.tokens) {
		return nil
	}
	return &p.tokens[p.pos]
}

func (p *parser) next() *token {
	t := p.peek()
	if t != nil {
		p.pos++
	}
	return t
}

func (p *parser) parseExpr() (Expr, error) {
	left, err := p.parseTerm()
	if err != nil {
		return nil, err
	}
	for t := p.peek(); t != nil && t.kind == tokenOr; t = p.peek() {
		p.next()
		right, err := p.parseTerm()
		if err != nil {
			return nil, err
		}
		left = &orExpr{left: left, right: right}
	}
	return left, nil
}

func (p *parser) parseTerm() (Expr, error) {
	left, err := p.parseFactor()
	if err != nil {
		return nil, err
	}
	for t := p.peek(); t != nil && t.kind == tokenAnd; t = p.peek() {
		p.next()
		right, err := p.parseFactor()
		if err != nil {
			return nil, err
		}
		left = &andExpr{left: left, right: right}
	}
	return left, nil
}

func (p *parser) parseFactor() (Expr, error) {
	t := p.next()
	if t == nil {
		return nil, fmt.Errorf("unexpected end of query")
	}
	switch t.kind {
	case tokenKeyword:
		return &keyword{term: []byte(t.value)}, nil
	case tokenLParen:
		expr, err := p.parseExpr()
		if err != nil {
			return nil, err
		}
		closing := p.next()
		if closing == nil || closing.kind != tokenRParen {
			return nil, fmt.Errorf("expected closing parenthesis")
		}
		return expr, nil
	default:
		return nil, fmt.Errorf("unexpected token: %q", t.value)
	}
}

func bytesContainsFold(haystack, needle []byte) bool {
	nl := len(needle)
	hl := len(haystack)
	if nl > hl {
		return false
	}
	for i := 0; i <= hl-nl; i++ {
		if equalFold(haystack[i:i+nl], needle) {
			return true
		}
	}
	return false
}

func equalFold(a, b []byte) bool {
	for i := range a {
		ca, cb := a[i], b[i]
		if ca == cb {
			continue
		}
		if 'A' <= ca && ca <= 'Z' {
			ca += 'a' - 'A'
		}
		if 'A' <= cb && cb <= 'Z' {
			cb += 'a' - 'A'
		}
		if ca != cb {
			return false
		}
	}
	return true
}

var (
	highlightStart = []byte("\033[1;31m")
	highlightReset = []byte("\033[0m")
)

func Highlight(line []byte, terms [][]byte) []byte {
	if len(terms) == 0 {
		return line
	}
	sorted := make([][]byte, len(terms))
	copy(sorted, terms)
	sort.Slice(sorted, func(i, j int) bool {
		return len(sorted[i]) > len(sorted[j])
	})

	var buf []byte
	i := 0
	for i < len(line) {
		matched := false
		for _, term := range sorted {
			tl := len(term)
			if i+tl <= len(line) && equalFold(line[i:i+tl], term) {
				buf = append(buf, highlightStart...)
				buf = append(buf, line[i:i+tl]...)
				buf = append(buf, highlightReset...)
				i += tl
				matched = true
				break
			}
		}
		if !matched {
			buf = append(buf, line[i])
			i++
		}
	}
	return buf
}
