package parser

import (
	"fmt"
	"slices"
)

type ParseError struct {
	message string
	pos     int
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("Parse error %s near pos %d", e.message, e.pos)
}

type Parser struct {
	look   token
	value  string
	lexer  lexer
	strict bool
}

type context struct {
	index    string
	relation string
	children []*Node
}

func (p *Parser) next() {
	p.look, p.value = p.lexer.lex()
}

func (p *Parser) isSearchTerm() bool {
	return p.look == tokenSimpleString ||
		p.look == tokenPrefixName ||
		p.look == tokenBoolOp ||
		p.look == tokenSortby
}

func (p *Parser) isRelation() bool {
	return p.look == tokenRelOp || p.look == tokenPrefixName
}

func (p *Parser) modifiers() ([]*Node, error) {
	var mods []*Node
	for p.look == tokenModifier {
		p.next()
		if !p.isSearchTerm() {
			return nil, &ParseError{"Missing modifier key", p.lexer.pos}
		}
		modifier := p.value
		p.next()
		if p.look == tokenRelOp {
			relation := p.value
			p.next()
			if !p.isSearchTerm() {
				return nil, &ParseError{"Missing modifier value", p.lexer.pos}
			}
			node := &Node{kind: Modifier, index: modifier, relation: relation, term: p.value}
			p.next()
			mods = append(mods, node)
		} else {
			node := &Node{kind: Modifier, index: modifier}
			mods = append(mods, node)
		}
	}
	return mods, nil
}

func (p *Parser) searchClause(c *context) (*Node, error) {
	if p.look == tokenLp {
		p.next()
		node, err := p.cqlQuery(c)
		if err != nil {
			return nil, err
		}
		if p.look != tokenRp {
			return nil, &ParseError{"Missing )", p.lexer.pos}
		}
		p.next()
		return node, nil
	}
	if !p.isSearchTerm() {
		return nil, &ParseError{"search term expected", p.lexer.pos}
	}
	indexOrTerm := p.value
	p.next()
	if p.isRelation() {
		relation := p.value
		p.next()
		mods, err := p.modifiers()
		if err != nil {
			return nil, err
		}
		c := context{index: indexOrTerm, relation: relation, children: mods}
		return p.searchClause(&c)
	}
	node := Node{kind: SearchTerm, index: c.index, relation: c.relation, term: indexOrTerm, children: c.children}
	return &node, nil
}

func (p *Parser) scopedClause(c *context) (*Node, error) {
	left, err := p.searchClause(c)
	if err != nil {
		return nil, err
	}
	for p.look == tokenBoolOp || p.look == tokenPrefixName {
		op := p.value
		p.next()
		mods, err := p.modifiers()
		if err != nil {
			return nil, err
		}
		right, err := p.searchClause(c)
		if err != nil {
			return nil, err
		}
		mods = slices.Insert(mods, 0, left, right)
		left = &Node{kind: BoolOp, index: op, children: mods}
	}
	return left, nil
}

func (p *Parser) cqlQuery(c *context) (*Node, error) {
	return p.scopedClause(c)
}

func (p *Parser) Parse(input string) (*Node, error) {
	p.lexer.init(input, p.strict)
	p.look, p.value = p.lexer.lex()

	c := context{index: "cql.serverChoice", relation: "="}
	node, err := p.cqlQuery(&c)
	if err == nil && p.look != tokenEos {
		return nil, &ParseError{"EOF expected", p.lexer.pos}
	}
	return node, err
}
