package cql

// Chainable helpers for constructing AST nodes directly.

// Search constructs a search clause with index, relation EQ, and term.
func Search(index, term string) Clause {
	sc := SearchClause{
		Index:    index,
		Relation: EQ,
		Term:     term,
	}
	return Clause{SearchClause: &sc}
}

// SearchRel constructs a search clause with index, relation, and term.
func SearchRel(index string, rel Relation, term string) Clause {
	sc := SearchClause{
		Index:    index,
		Relation: rel,
		Term:     term,
	}
	return Clause{SearchClause: &sc}
}

// Bool constructs a boolean clause with operator and child clauses.
func Bool(op Operator, left, right Clause) Clause {
	bc := BoolClause{
		Left:     left,
		Operator: op,
		Right:    right,
	}
	return Clause{BoolClause: &bc}
}

// And constructs an AND boolean clause.
func And(left, right Clause) Clause {
	return Bool(AND, left, right)
}

// Or constructs an OR boolean clause.
func Or(left, right Clause) Clause {
	return Bool(OR, left, right)
}

// Not constructs a NOT boolean clause.
func Not(left, right Clause) Clause {
	return Bool(NOT, left, right)
}

// Prox constructs a PROX boolean clause.
func Prox(left, right Clause) Clause {
	return Bool(PROX, left, right)
}

// WithClause sets the query clause.
func (q *Query) WithClause(clause Clause) *Query {
	if q == nil {
		return q
	}
	q.Clause = clause
	return q
}

// WithSorts replaces the sort specification.
func (q *Query) WithSorts(sorts ...Sort) *Query {
	if q == nil {
		return q
	}
	q.SortSpec = sorts
	return q
}

// AddSort appends a sort specification.
func (q *Query) AddSort(sort Sort) *Query {
	if q == nil {
		return q
	}
	q.SortSpec = append(q.SortSpec, sort)
	return q
}

// WithPrefixes replaces the prefix map.
func (c *Clause) WithPrefixes(prefixes ...Prefix) *Clause {
	if c == nil {
		return c
	}
	c.PrefixMap = prefixes
	return c
}

// AddPrefix appends a prefix mapping.
func (c *Clause) AddPrefix(prefix Prefix) *Clause {
	if c == nil {
		return c
	}
	c.PrefixMap = append(c.PrefixMap, prefix)
	return c
}

// WithSearchClause sets the search clause and clears any boolean clause.
func (c *Clause) WithSearchClause(sc SearchClause) *Clause {
	if c == nil {
		return c
	}
	c.SearchClause = &sc
	c.BoolClause = nil
	return c
}

// WithBoolClause sets the boolean clause and clears any search clause.
func (c *Clause) WithBoolClause(bc BoolClause) *Clause {
	if c == nil {
		return c
	}
	c.BoolClause = &bc
	c.SearchClause = nil
	return c
}

// WithIndex sets the search index.
func (sc *SearchClause) WithIndex(index string) *SearchClause {
	if sc == nil {
		return sc
	}
	sc.Index = index
	return sc
}

// WithRelation sets the search relation.
func (sc *SearchClause) WithRelation(relation Relation) *SearchClause {
	if sc == nil {
		return sc
	}
	sc.Relation = relation
	return sc
}

// WithModifiers replaces the search modifiers.
func (sc *SearchClause) WithModifiers(mods ...Modifier) *SearchClause {
	if sc == nil {
		return sc
	}
	sc.Modifiers = mods
	return sc
}

// AddModifier appends a modifier to the search clause.
func (sc *SearchClause) AddModifier(mod Modifier) *SearchClause {
	if sc == nil {
		return sc
	}
	sc.Modifiers = append(sc.Modifiers, mod)
	return sc
}

// WithTerm sets the search term.
func (sc *SearchClause) WithTerm(term string) *SearchClause {
	if sc == nil {
		return sc
	}
	sc.Term = term
	return sc
}

// WithLeft sets the left clause.
func (bc *BoolClause) WithLeft(left Clause) *BoolClause {
	if bc == nil {
		return bc
	}
	bc.Left = left
	return bc
}

// WithOperator sets the boolean operator.
func (bc *BoolClause) WithOperator(op Operator) *BoolClause {
	if bc == nil {
		return bc
	}
	bc.Operator = op
	return bc
}

// WithModifiers replaces the boolean modifiers.
func (bc *BoolClause) WithModifiers(mods ...Modifier) *BoolClause {
	if bc == nil {
		return bc
	}
	bc.Modifiers = mods
	return bc
}

// AddModifier appends a modifier to the boolean clause.
func (bc *BoolClause) AddModifier(mod Modifier) *BoolClause {
	if bc == nil {
		return bc
	}
	bc.Modifiers = append(bc.Modifiers, mod)
	return bc
}

// WithRight sets the right clause.
func (bc *BoolClause) WithRight(right Clause) *BoolClause {
	if bc == nil {
		return bc
	}
	bc.Right = right
	return bc
}

// WithIndex sets the sort index.
func (s *Sort) WithIndex(index string) *Sort {
	if s == nil {
		return s
	}
	s.Index = index
	return s
}

// WithModifiers replaces the sort modifiers.
func (s *Sort) WithModifiers(mods ...Modifier) *Sort {
	if s == nil {
		return s
	}
	s.Modifiers = mods
	return s
}

// AddModifier appends a modifier to the sort.
func (s *Sort) AddModifier(mod Modifier) *Sort {
	if s == nil {
		return s
	}
	s.Modifiers = append(s.Modifiers, mod)
	return s
}

// WithName sets the modifier name.
func (m *Modifier) WithName(name string) *Modifier {
	if m == nil {
		return m
	}
	m.Name = name
	return m
}

// WithRelation sets the modifier relation.
func (m *Modifier) WithRelation(rel Relation) *Modifier {
	if m == nil {
		return m
	}
	m.Relation = rel
	return m
}

// WithValue sets the modifier value.
func (m *Modifier) WithValue(value string) *Modifier {
	if m == nil {
		return m
	}
	m.Value = value
	return m
}

// WithPrefix sets the prefix name.
func (p *Prefix) WithPrefix(prefix string) *Prefix {
	if p == nil {
		return p
	}
	p.Prefix = prefix
	return p
}

// WithUri sets the prefix URI.
func (p *Prefix) WithUri(uri string) *Prefix {
	if p == nil {
		return p
	}
	p.Uri = uri
	return p
}
