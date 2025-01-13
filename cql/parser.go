package cql

import (
	"fmt"
)

type CqlError struct {
	message string
	pos     int
}

func (e *CqlError) Error() string {
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
	relation_mods []*CqlNode
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

func (p *Parser) modifiers() ([]*CqlNode, error) {
	var mods []*CqlNode
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
			node := &CqlNode{Search: &SearchClauseNode{Index: modifier, Relation: relation, Term: p.value}}
			p.next()
			mods = append(mods, node)
		} else {
			node := &CqlNode{Search: &SearchClauseNode{Index: modifier}}
			mods = append(mods, node)
		}
	}
	return mods, nil
}

func (p *Parser) searchClause(ctx *context) (*CqlNode, error) {
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
	node := &CqlNode{Search: &SearchClauseNode{Index: ctx.index, Relation: ctx.relation, Term: indexOrTerm, Modifiers: ctx.relation_mods}}
	return node, nil
}

func (p *Parser) scopedClause(ctx *context) (*CqlNode, error) {
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
		left = &CqlNode{Boolean: &BooleanNode{Operator: op, Modifiers: mods, Left: left, Right: right}}
	}
	return left, nil
}

func (p *Parser) cqlQuery(ctx *context) (*CqlNode, error) {
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
		pnode := &CqlNode{Prefix: &PrefixNode{Prefix: prefix, Uri: uri, Next: node}}
		return pnode, nil
	}
	return p.scopedClause(ctx)
}

func (p *Parser) Parse(input string) (*CqlNode, error) {
	p.lexer.init(input, p.strict)
	p.look, p.value = p.lexer.lex()

	ctx := context{index: "cql.serverChoice", relation: "="}
	node, err := p.cqlQuery(&ctx)
	if err != nil {
		return nil, err
	}
	if p.look == tokenSortby {
		searchNode := node
		var prevNode *CqlNode = nil
		p.next()
		for p.isSearchTerm() {
			index := p.value
			p.next()
			mods, err := p.modifiers()
			if err != nil {
				return nil, err
			}
			sortNode := &CqlNode{Sort: &SortNode{Index: index, Modifiers: mods}}
			if prevNode != nil {
				prevNode.Sort.Next = sortNode
			} else {
				node = sortNode
			}
			prevNode = sortNode
		}
		prevNode.Sort.Next = searchNode

	}
	if p.look != tokenEos {
		return nil, &CqlError{"EOF expected", p.lexer.pos}
	}
	return node, err
}
