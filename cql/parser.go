package cql

import (
	"fmt"
	"strings"
)

type ParseError struct {
	message string
	pos     int
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("%s near pos %d", e.message, e.pos)
}

type Parser struct {
	Strict bool //if true: multi term values are not allowed
	look   token
	value  string
	lexer  lexer
}

type context struct {
	index         string
	relation      Relation
	relation_mods []Modifier
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
		p.look == tokenSortby
}

func (p *Parser) isRelation() bool {
	return p.look == tokenRelOp || p.look == tokenPrefixName
}

func (p *Parser) modifiers() ([]Modifier, error) {
	var mods []Modifier
	for p.look == tokenModifier {
		p.next()
		if !p.isSearchTerm() {
			return mods, &ParseError{"missing modifier key", p.lexer.pos}
		}
		modifier := p.value
		p.next()
		if p.look == tokenRelOp {
			relation := Relation(p.value)
			p.next()
			if !p.isSearchTerm() {
				return mods, &ParseError{"missing modifier value", p.lexer.pos}
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
			return node, &ParseError{"missing )", p.lexer.pos}
		}
		p.next()
		return node, nil
	}
	var node Clause
	if !p.isSearchTerm() {
		return node, &ParseError{"search term expected", p.lexer.pos}
	}
	indexOrTerm := p.value
	p.next()
	if p.isRelation() {
		relation := Relation(p.value)
		p.next()
		mods, err := p.modifiers()
		if err != nil {
			return node, err
		}
		ctx := context{index: indexOrTerm, relation: relation, relation_mods: mods}
		return p.searchClause(&ctx)
	}
	var sb strings.Builder

	sb.WriteString(indexOrTerm)
	for p.look == tokenSimpleString {
		sb.WriteString(" " + p.value)
		p.next()
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
	for p.look == tokenRelOp && p.value == ">" {
		p.next()
		if p.look != tokenSimpleString {
			return node, &ParseError{"term expected after >", p.lexer.pos}
		}
		var uri string
		value := p.value
		p.next()
		if p.look == tokenRelOp && p.value == "=" {
			p.next()
			if p.look != tokenSimpleString {
				return node, &ParseError{"term expected after =", p.lexer.pos}
			}
			uri = p.value
			p.next()
		} else {
			uri = value
			value = ""
		}
		prefix := Prefix{Prefix: value, Uri: uri}
		prefixes = append(prefixes, prefix)
	}
	node, err := p.scopedClause(ctx)
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

func (p *Parser) Parse(input string) (Query, error) {
	p.lexer.init(input, p.Strict)
	p.look, p.value = p.lexer.lex()

	ctx := context{index: "cql.serverChoice", relation: "="}
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
		return query, &ParseError{"EOF expected", p.lexer.pos}
	}
	return query, err
}
