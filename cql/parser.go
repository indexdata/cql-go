package cql

import (
	"fmt"
	"slices"
	"strings"
)

const combiningTildeBelow = string('\u0330')

// Indicates query parsing error
type ParseError struct {
	query   string
	message string
	pos     int
}

// Formats error for display
func (e *ParseError) Error() string {
	return fmt.Sprintf("%s at position %d: %s", e.message, e.pos, e.Marked())
}

// Raw parser error message
func (e *ParseError) Message() string {
	return e.message
}

// Error position in the query
func (e *ParseError) Pos() int {
	return e.pos
}

// Query that caused the error
func (e *ParseError) Query() string {
	return e.query
}

// Query with the error position marked
func (e *ParseError) Marked() string {
	return e.query[:e.pos] + combiningTildeBelow + e.query[e.pos:]
}

// CQL parser, non-strict by default
type Parser struct {
	Strict bool //if true, multi term values, e.g. `a b c` are not allowed
	look   token
	value  string
	lexer  lexer
}

type context struct {
	index         string
	relation      Relation
	relation_mods []Modifier
	prefixes      []string
	custom        bool
}

func (p *Parser) next() {
	p.look, p.value = p.lexer.lex()
}

func (p *Parser) isSearchTerm() bool {
	return p.look == tokenSimpleString ||
		p.look == tokenPrefixName ||
		p.look == tokenAnd ||
		p.look == tokenOr ||
		p.look == tokenNot ||
		p.look == tokenProx ||
		p.look == tokenSortby ||
		p.look == tokenRelSym
}

func (p *Parser) isRelation(prefixes []string, custom bool) bool {
	return p.look == tokenRelOp || p.look == tokenRelSym ||
		(p.look == tokenPrefixName && slices.Contains(prefixes, strings.Split(p.value, ".")[0])) ||
		(p.look == tokenSimpleString && custom)
}

func (p *Parser) modifiers() ([]Modifier, error) {
	var mods []Modifier
	for p.look == tokenModifier {
		p.next()
		if !p.isSearchTerm() {
			return mods, &ParseError{p.lexer.input, "missing modifier key", p.lexer.pos}
		}
		modifier := p.value
		p.next()
		if p.look == tokenRelOp {
			relation := Relation(p.value)
			p.next()
			if !p.isSearchTerm() {
				return mods, &ParseError{p.lexer.input, "missing modifier value", p.lexer.pos}
			}
			mod := Modifier{Name: modifier, Relation: relation, Value: p.value}
			p.next()
			mods = append(mods, mod)
		} else {
			mod := Modifier{Name: modifier}
			mods = append(mods, mod)
		}
	}
	return mods, nil
}

func (p *Parser) searchClause(ctx *context) (Clause, error) {
	if p.look == tokenLp {
		p.next()
		node, err := p.cqlQuery(ctx)
		if err != nil {
			return node, err
		}
		if p.look != tokenRp {
			return node, &ParseError{p.lexer.input, "missing )", p.lexer.pos}
		}
		p.next()
		return node, nil
	}
	var node Clause
	if !p.isSearchTerm() {
		return node, &ParseError{p.lexer.input, "search term expected", p.lexer.pos}
	}
	indexOrTerm := p.value
	relPos := p.lexer.pos
	p.next()
	if p.isRelation(ctx.prefixes, ctx.custom) {
		relation := Relation(p.value)
		p.next()
		mods, err := p.modifiers()
		if err != nil {
			return node, err
		}
		ctx := context{index: indexOrTerm, relation: relation, relation_mods: mods, prefixes: ctx.prefixes, custom: ctx.custom}
		return p.searchClause(&ctx)
	}
	var sb strings.Builder
	sb.WriteString(indexOrTerm)
	for p.look == tokenSimpleString || p.look == tokenPrefixName || p.look == tokenRelSym {
		if p.Strict {
			return node, &ParseError{p.lexer.input, "relation expected", relPos}
		} else {
			sb.WriteString(" " + p.value)
			p.next()
		}
	}
	sc := SearchClause{Index: ctx.index, Relation: ctx.relation, Term: sb.String(), Modifiers: ctx.relation_mods}
	node.SearchClause = &sc
	return node, nil
}

func (p *Parser) scopedClause(ctx *context) (Clause, error) {
	left, err := p.searchClause(ctx)
	if err != nil {
		return left, err
	}
	for {
		var op Operator
		switch p.look {
		case tokenAnd:
			op = "and"
		case tokenOr:
			op = "or"
		case tokenNot:
			op = "not"
		case tokenProx:
			op = "prox"
		default:
			return left, nil
		}
		p.next()
		mods, err := p.modifiers()
		if err != nil {
			return left, err
		}
		right, err := p.searchClause(ctx)
		if err != nil {
			return left, err
		}
		bnode := BoolClause{Operator: op, Modifiers: mods, Left: left, Right: right}
		left = Clause{BoolClause: &bnode}
	}
}

func (p *Parser) cqlQuery(ctx *context) (Clause, error) {
	var prefixes []Prefix
	var node Clause
	subctx := *ctx
	for p.look == tokenRelOp && p.value == ">" {
		p.next()
		if p.look != tokenSimpleString {
			return node, &ParseError{p.lexer.input, "prefix or uri expected", p.lexer.pos}
		}
		var uri string
		value := p.value
		p.next()
		if p.look == tokenRelOp && p.value == "=" {
			p.next()
			if p.look != tokenSimpleString {
				return node, &ParseError{p.lexer.input, "uri expected", p.lexer.pos}
			}
			uri = p.value
			subctx.prefixes = append(ctx.prefixes, value)
			p.next()
		} else {
			subctx.custom = true
			uri = value
			value = ""
		}
		prefix := Prefix{Prefix: value, Uri: uri}
		prefixes = append(prefixes, prefix)
	}
	node, err := p.scopedClause(&subctx)
	node.PrefixMap = prefixes
	return node, err
}

func (p *Parser) sortKeys() ([]Sort, error) {
	var sortList []Sort

	for p.isSearchTerm() {
		index := p.value
		p.next()
		mods, err := p.modifiers()
		if err != nil {
			return sortList, err
		}
		sort := Sort{Index: index, Modifiers: mods}
		sortList = append(sortList, sort)
	}
	return sortList, nil
}

// Parse input query string into a syntax tree or return an error.
func (p *Parser) Parse(input string) (Query, error) {
	p.lexer.init(input)
	p.look, p.value = p.lexer.lex()

	ctx := context{index: "cql.serverChoice", relation: "=", prefixes: []string{"cql"}}
	var query Query

	node, err := p.cqlQuery(&ctx)
	if err != nil {
		return query, err
	}
	query.Clause = node
	if p.look == tokenSortby {
		p.next()
		query.SortSpec, err = p.sortKeys()
	}
	if p.look != tokenEos {
		return query, &ParseError{p.lexer.input, "EOF expected", p.lexer.pos}
	}
	return query, err
}
