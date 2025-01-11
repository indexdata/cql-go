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
	return fmt.Sprintf("%s near pos %d", e.message, e.pos)
}

type Parser struct {
	look   token
	value  string
	lexer  lexer
	strict bool
}

type context struct {
	index         string
	relation      string
	relation_mods []*Node
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
			return nil, &ParseError{"missing modifier key", p.lexer.pos}
		}
		modifier := p.value
		p.next()
		if p.look == tokenRelOp {
			relation := p.value
			p.next()
			if !p.isSearchTerm() {
				return nil, &ParseError{"missing modifier value", p.lexer.pos}
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
			return nil, &ParseError{"missing )", p.lexer.pos}
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
		c := context{index: indexOrTerm, relation: relation, relation_mods: mods}
		return p.searchClause(&c)
	}
	node := Node{kind: SearchTerm, index: c.index, relation: c.relation, term: indexOrTerm, children: c.relation_mods}
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
	if p.look == tokenRelOp && p.value == ">" {
		p.next()
		if p.look != tokenSimpleString {
			return nil, &ParseError{"term expected after >", p.lexer.pos}
		}
		var uri string
		prefix := p.value
		p.next()
		if p.look == tokenRelOp && p.value == "=" {
			p.next()
			if p.look != tokenSimpleString {
				return nil, &ParseError{"term expected after =", p.lexer.pos}
			}
			uri = p.value
			p.next()
		} else {
			uri = prefix
			prefix = ""
		}
		node, err := p.cqlQuery(c)
		if err != nil {
			return nil, err
		}
		pnode := &Node{kind: Prefix, index: prefix, term: uri, children: [](*Node){node}}
		return pnode, nil
	}
	return p.scopedClause(c)
}

func (p *Parser) Parse(input string) (*Node, error) {
	p.lexer.init(input, p.strict)
	p.look, p.value = p.lexer.lex()

	c := context{index: "cql.serverChoice", relation: "="}
	node, err := p.cqlQuery(&c)
	if err != nil {
		return nil, err
	}
	if p.look == tokenSortby {
		p.next()
		var children []*Node
		children = append(children, node)
		for p.isSearchTerm() {
			index := p.value
			p.next()
			mods, err := p.modifiers()
			if err != nil {
				return nil, err
			}
			snode := &Node{kind: SortOp, index: index, children: mods}
			children = append(children, snode)
		}
		node = &Node{kind: SortOp, children: children}
	}
	if p.look != tokenEos {
		return nil, &ParseError{"EOF expected", p.lexer.pos}
	}
	return node, err
}
