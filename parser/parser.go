package parser

import (
	"fmt"
	"slices"
)

type CqlError struct {
	message string
	pos     int
}

func (e *CqlError) Error() string {
	return fmt.Sprintf("%s near pos %d", e.message, e.pos)
}

type CqlParser struct {
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

func (p *CqlParser) next() {
	p.look, p.value = p.lexer.lex()
}

func (p *CqlParser) isSearchTerm() bool {
	return p.look == tokenSimpleString ||
		p.look == tokenPrefixName ||
		p.look == tokenBoolOp ||
		p.look == tokenSortby
}

func (p *CqlParser) isRelation() bool {
	return p.look == tokenRelOp || p.look == tokenPrefixName
}

func (p *CqlParser) modifiers() ([]*Node, error) {
	var mods []*Node
	for p.look == tokenModifier {
		p.next()
		if !p.isSearchTerm() {
			return nil, &CqlError{"missing modifier key", p.lexer.pos}
		}
		modifier := p.value
		p.next()
		if p.look == tokenRelOp {
			relation := p.value
			p.next()
			if !p.isSearchTerm() {
				return nil, &CqlError{"missing modifier value", p.lexer.pos}
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

func (p *CqlParser) searchClause(ctx *context) (*Node, error) {
	if p.look == tokenLp {
		p.next()
		node, err := p.cqlQuery(ctx)
		if err != nil {
			return nil, err
		}
		if p.look != tokenRp {
			return nil, &CqlError{"missing )", p.lexer.pos}
		}
		p.next()
		return node, nil
	}
	if !p.isSearchTerm() {
		return nil, &CqlError{"search term expected", p.lexer.pos}
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
		ctx := context{index: indexOrTerm, relation: relation, relation_mods: mods}
		return p.searchClause(&ctx)
	}
	node := Node{kind: SearchTerm, index: ctx.index, relation: ctx.relation, term: indexOrTerm, children: ctx.relation_mods}
	return &node, nil
}

func (p *CqlParser) scopedClause(ctx *context) (*Node, error) {
	left, err := p.searchClause(ctx)
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
		right, err := p.searchClause(ctx)
		if err != nil {
			return nil, err
		}
		mods = slices.Insert(mods, 0, left, right)
		left = &Node{kind: BoolOp, index: op, children: mods}
	}
	return left, nil
}

func (p *CqlParser) cqlQuery(ctx *context) (*Node, error) {
	if p.look == tokenRelOp && p.value == ">" {
		p.next()
		if p.look != tokenSimpleString {
			return nil, &CqlError{"term expected after >", p.lexer.pos}
		}
		var uri string
		prefix := p.value
		p.next()
		if p.look == tokenRelOp && p.value == "=" {
			p.next()
			if p.look != tokenSimpleString {
				return nil, &CqlError{"term expected after =", p.lexer.pos}
			}
			uri = p.value
			p.next()
		} else {
			uri = prefix
			prefix = ""
		}
		node, err := p.cqlQuery(ctx)
		if err != nil {
			return nil, err
		}
		pnode := &Node{kind: Prefix, index: prefix, term: uri, children: [](*Node){node}}
		return pnode, nil
	}
	return p.scopedClause(ctx)
}

func (p *CqlParser) Parse(input string) (*Node, error) {
	p.lexer.init(input, p.strict)
	p.look, p.value = p.lexer.lex()

	ctx := context{index: "cql.serverChoice", relation: "="}
	node, err := p.cqlQuery(&ctx)
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
		return nil, &CqlError{"EOF expected", p.lexer.pos}
	}
	return node, err
}
