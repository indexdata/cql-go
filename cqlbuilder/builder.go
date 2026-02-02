// Package cqlbuilder provides a fluent, validated builder for CQL queries.
package cqlbuilder

import (
	"fmt"
	"strings"

	"github.com/indexdata/cql-go/cql"
)

// NewQuery creates a new query builder.
func NewQuery() *QueryBuilder {
	return &QueryBuilder{}
}

// NewQueryFromString initializes a builder from a CQL string.
func NewQueryFromString(input string) (*QueryBuilder, error) {
	var parser cql.Parser
	query, err := parser.Parse(input)
	if err != nil {
		return nil, err
	}
	return NewQueryFrom(query), nil
}

// NewQueryFrom initializes a builder with an existing query.
func NewQueryFrom(query cql.Query) *QueryBuilder {
	clause := query.Clause
	return &QueryBuilder{
		sorts: query.SortSpec,
		root:  &clause,
	}
}

// BeginClause starts a grouped root clause.
func (qb *QueryBuilder) BeginClause() *ClauseBuilder {
	if qb.err == nil && qb.root != nil {
		qb.err = fmt.Errorf("query already has a root clause")
	}
	return &ClauseBuilder{
		ctx: &clauseContext{root: qb},
	}
}

// QueryBuilder builds a validated cql.Query.
type QueryBuilder struct {
	prefixes []cql.Prefix
	sorts    []cql.Sort
	root     *cql.Clause
	err      error
}

// Prefix adds a prefix declaration.
func (qb *QueryBuilder) Prefix(prefix, uri string) *QueryBuilder {
	if qb.err != nil {
		return qb
	}
	if strings.TrimSpace(uri) == "" {
		qb.err = fmt.Errorf("prefix uri must be non-empty")
		return qb
	}
	qb.prefixes = append(qb.prefixes, cql.Prefix{
		Prefix: prefix,
		Uri:    uri,
	})
	return qb
}

// Search starts a search expression as the root clause.
func (qb *QueryBuilder) Search(index string) *SearchBuilder {
	if qb.err == nil && qb.root != nil {
		qb.err = fmt.Errorf("query already has a root clause")
	}
	return &SearchBuilder{
		index:  index,
		finish: qb.finishRoot,
		build:  qb.Build,
		qb:     qb,
		err:    qb.err,
	}
}

// And appends an AND boolean expression to the existing root clause.
func (qb *QueryBuilder) And() *JoinBuilder {
	return qb.append(cql.AND)
}

// Or appends an OR boolean expression to the existing root clause.
func (qb *QueryBuilder) Or() *JoinBuilder {
	return qb.append(cql.OR)
}

// Not appends a NOT boolean expression to the existing root clause.
func (qb *QueryBuilder) Not() *JoinBuilder {
	return qb.append(cql.NOT)
}

// Prox appends a PROX boolean expression to the existing root clause.
func (qb *QueryBuilder) Prox() *JoinBuilder {
	return qb.append(cql.PROX)
}

func (qb *QueryBuilder) append(op cql.Operator) *JoinBuilder {
	if qb.err == nil && qb.root == nil {
		qb.err = fmt.Errorf("query requires a root clause before appending")
	}
	var left cql.Clause
	if qb.root != nil {
		left = *qb.root
	}
	return &JoinBuilder{
		finish: qb.finishAppend,
		build:  qb.Build,
		qb:     qb,
		left:   left,
		op:     op,
		err:    qb.err,
	}
}

// SortBy adds a sort criterion with simple (name-only) modifiers.
func (qb *QueryBuilder) SortBy(index string, mods ...cql.CqlModifier) *QueryBuilder {
	if qb.err != nil {
		return qb
	}
	if strings.TrimSpace(index) == "" {
		qb.err = fmt.Errorf("sort index must be non-empty")
		return qb
	}
	var out []cql.Modifier
	for _, mod := range mods {
		if strings.TrimSpace(string(mod)) == "" {
			qb.err = fmt.Errorf("sort modifier name must be non-empty")
			return qb
		}
		out = append(out, cql.Modifier{Name: string(mod)})
	}
	qb.sorts = append(qb.sorts, cql.Sort{Index: index, Modifiers: out})
	return qb
}

// SortByModifiers adds a sort criterion with fully-specified modifiers.
func (qb *QueryBuilder) SortByModifiers(index string, mods ...cql.Modifier) *QueryBuilder {
	if qb.err != nil {
		return qb
	}
	if strings.TrimSpace(index) == "" {
		qb.err = fmt.Errorf("sort index must be non-empty")
		return qb
	}
	for i := range mods {
		if strings.TrimSpace(mods[i].Name) == "" {
			qb.err = fmt.Errorf("sort modifier name must be non-empty")
			return qb
		}
		if mods[i].Value != "" {
			mods[i].Value = EscapeSpecialChars(mods[i].Value)
		}
		if mods[i].Relation == "" && mods[i].Value != "" {
			mods[i].Relation = cql.EQ
		}
		if mods[i].Relation != "" && !isValidRelation(mods[i].Relation) {
			qb.err = fmt.Errorf("invalid modifier relation: %q", mods[i].Relation)
			return qb
		}
	}
	qb.sorts = append(qb.sorts, cql.Sort{Index: index, Modifiers: mods})
	return qb
}

// Build validates and returns the final query.
func (qb *QueryBuilder) Build() (cql.Query, error) {
	if qb.err != nil {
		return cql.Query{}, qb.err
	}
	if qb.root == nil {
		return cql.Query{}, fmt.Errorf("query requires a root clause")
	}
	root := *qb.root
	if len(qb.prefixes) > 0 {
		root.PrefixMap = append(root.PrefixMap, qb.prefixes...)
	}
	return cql.Query{
		Clause:   root,
		SortSpec: qb.sorts,
	}, nil
}

func (qb *QueryBuilder) finishAppend(clause cql.Clause) *ExprBuilder {
	if qb.err != nil {
		return &ExprBuilder{finish: qb.finishAppend, build: qb.Build, qb: qb, end: nil, err: qb.err}
	}
	qb.root = &clause
	return &ExprBuilder{finish: qb.finishAppend, build: qb.Build, qb: qb, end: nil, clause: clause}
}

func (qb *QueryBuilder) finishRoot(clause cql.Clause) *ExprBuilder {
	if qb.err != nil {
		return &ExprBuilder{finish: qb.finishRoot, build: qb.Build, qb: qb, end: nil, err: qb.err}
	}
	qb.root = &clause
	return &ExprBuilder{finish: qb.finishRoot, build: qb.Build, qb: qb, end: nil, clause: clause}
}

// ExprBuilder represents a completed expression that can be extended with boolean operators.
type ExprBuilder struct {
	finish func(cql.Clause) *ExprBuilder
	build  func() (cql.Query, error)
	qb     *QueryBuilder
	end    func(cql.Clause) *ExprBuilder
	clause cql.Clause
	err    error
}

// And starts an AND boolean expression.
func (eb *ExprBuilder) And() *JoinBuilder {
	return eb.join(cql.AND)
}

// Or starts an OR boolean expression.
func (eb *ExprBuilder) Or() *JoinBuilder {
	return eb.join(cql.OR)
}

// Not starts a NOT boolean expression.
func (eb *ExprBuilder) Not() *JoinBuilder {
	return eb.join(cql.NOT)
}

// Prox starts a PROX boolean expression.
func (eb *ExprBuilder) Prox() *JoinBuilder {
	return eb.join(cql.PROX)
}

// EndClause closes a grouped clause and returns to the parent expression.
func (eb *ExprBuilder) EndClause() *ExprBuilder {
	if eb.end == nil {
		eb.err = fmt.Errorf("no open clause to end")
		return eb
	}
	return eb.end(eb.clause)
}

// SortBy adds a sort criterion with simple (name-only) modifiers.
func (eb *ExprBuilder) SortBy(index string, mods ...cql.CqlModifier) *ExprBuilder {
	if eb.err != nil || eb.qb == nil {
		return eb
	}
	eb.qb.SortBy(index, mods...)
	if eb.err == nil && eb.qb.err != nil {
		eb.err = eb.qb.err
	}
	return eb
}

// SortByModifiers adds a sort criterion with fully-specified modifiers.
func (eb *ExprBuilder) SortByModifiers(index string, mods ...cql.Modifier) *ExprBuilder {
	if eb.err != nil || eb.qb == nil {
		return eb
	}
	eb.qb.SortByModifiers(index, mods...)
	if eb.err == nil && eb.qb.err != nil {
		eb.err = eb.qb.err
	}
	return eb
}

// Build finalizes and returns the query.
func (eb *ExprBuilder) Build() (cql.Query, error) {
	if eb.build == nil {
		return cql.Query{}, fmt.Errorf("builder is missing query context")
	}
	return eb.build()
}

func (eb *ExprBuilder) join(op cql.Operator) *JoinBuilder {
	return &JoinBuilder{
		finish: eb.finish,
		build:  eb.build,
		qb:     eb.qb,
		end:    eb.end,
		left:   eb.clause,
		op:     op,
		err:    eb.err,
	}
}

// JoinBuilder builds a boolean clause by providing the right-hand side.
type JoinBuilder struct {
	finish func(cql.Clause) *ExprBuilder
	build  func() (cql.Query, error)
	qb     *QueryBuilder
	end    func(cql.Clause) *ExprBuilder
	left   cql.Clause
	op     cql.Operator
	mods   []cql.Modifier
	err    error
}

// Mod adds a modifier to the boolean operator (name-only).
func (jb *JoinBuilder) Mod(name cql.CqlModifier) *JoinBuilder {
	if jb.err != nil {
		return jb
	}
	if strings.TrimSpace(string(name)) == "" {
		jb.err = fmt.Errorf("modifier name must be non-empty")
		return jb
	}
	jb.mods = append(jb.mods, cql.Modifier{Name: string(name)})
	return jb
}

// ModRel adds a modifier with relation and value to the boolean operator.
func (jb *JoinBuilder) ModRel(name cql.CqlModifier, rel cql.Relation, value string) *JoinBuilder {
	if jb.err != nil {
		return jb
	}
	if strings.TrimSpace(string(name)) == "" {
		jb.err = fmt.Errorf("modifier name must be non-empty")
		return jb
	}
	if rel == "" {
		rel = cql.EQ
	}
	if !isValidRelation(rel) {
		jb.err = fmt.Errorf("invalid modifier relation: %q", rel)
		return jb
	}
	jb.mods = append(jb.mods, cql.Modifier{
		Name:     string(name),
		Relation: rel,
		Value:    EscapeSpecialChars(value),
	})
	return jb
}

// BeginClause starts a grouped boolean clause as the right-hand side.
func (jb *JoinBuilder) BeginClause() *ClauseBuilder {
	return &ClauseBuilder{
		ctx: &clauseContext{parent: jb},
	}
}

// Search provides the right-hand search clause.
func (jb *JoinBuilder) Search(index string) *SearchBuilder {
	return &SearchBuilder{
		index:  index,
		finish: jb.finishRight,
		end:    jb.end,
		err:    jb.err,
	}
}

func (jb *JoinBuilder) finishRight(right cql.Clause) *ExprBuilder {
	if jb.err != nil {
		return &ExprBuilder{finish: jb.finish, build: jb.build, qb: jb.qb, end: jb.end, err: jb.err}
	}
	if !isValidOperator(jb.op) {
		jb.err = fmt.Errorf("invalid boolean operator: %q", jb.op)
		return &ExprBuilder{finish: jb.finish, build: jb.build, qb: jb.qb, end: jb.end, err: jb.err}
	}
	bc := cql.BoolClause{
		Left:      jb.left,
		Operator:  jb.op,
		Modifiers: jb.mods,
		Right:     right,
	}
	clause := cql.Clause{BoolClause: &bc}
	return jb.finish(clause)
}

type clauseContext struct {
	root   *QueryBuilder
	parent *JoinBuilder
}

func (cc *clauseContext) finish(clause cql.Clause) *ExprBuilder {
	return &ExprBuilder{
		finish: cc.finish,
		end:    cc.end,
		clause: clause,
		err:    cc.err(),
	}
}

func (cc *clauseContext) end(clause cql.Clause) *ExprBuilder {
	if cc.root != nil {
		if cc.root.err != nil {
			return &ExprBuilder{finish: cc.finish, build: cc.root.Build, qb: cc.root, end: nil, err: cc.root.err}
		}
		cc.root.root = &clause
		return &ExprBuilder{finish: cc.root.finishRoot, build: cc.root.Build, qb: cc.root, end: nil, clause: clause}
	}
	return cc.parent.finishRight(clause)
}

func (cc *clauseContext) err() error {
	if cc.root != nil {
		return cc.root.err
	}
	return cc.parent.err
}

// ClauseBuilder builds a grouped clause on the right-hand side of a boolean operator.
type ClauseBuilder struct {
	ctx *clauseContext
}

// Search starts the grouped clause with a search expression.
func (cb *ClauseBuilder) Search(index string) *SearchBuilder {
	return &SearchBuilder{
		index:  index,
		finish: cb.ctx.finish,
		end:    cb.ctx.end,
		err:    cb.ctx.err(),
	}
}

// SearchBuilder builds a search clause.
type SearchBuilder struct {
	index  string
	rel    cql.Relation
	mods   []cql.Modifier
	finish func(cql.Clause) *ExprBuilder
	build  func() (cql.Query, error)
	qb     *QueryBuilder
	end    func(cql.Clause) *ExprBuilder
	err    error
}

// Rel sets the relation for the search clause.
func (sb *SearchBuilder) Rel(rel cql.Relation) *SearchBuilder {
	if sb.err != nil {
		return sb
	}
	if rel != "" && !isValidRelation(rel) {
		sb.err = fmt.Errorf("invalid relation: %q", rel)
		return sb
	}
	sb.rel = rel
	return sb
}

// Mod adds a modifier (name-only).
func (sb *SearchBuilder) Mod(name cql.CqlModifier) *SearchBuilder {
	if sb.err != nil {
		return sb
	}
	if strings.TrimSpace(string(name)) == "" {
		sb.err = fmt.Errorf("modifier name must be non-empty")
		return sb
	}
	sb.mods = append(sb.mods, cql.Modifier{Name: string(name)})
	return sb
}

// ModRel adds a modifier with relation and value.
func (sb *SearchBuilder) ModRel(name cql.CqlModifier, rel cql.Relation, value string) *SearchBuilder {
	if sb.err != nil {
		return sb
	}
	if strings.TrimSpace(string(name)) == "" {
		sb.err = fmt.Errorf("modifier name must be non-empty")
		return sb
	}
	if rel == "" {
		rel = cql.EQ
	}
	if !isValidRelation(rel) {
		sb.err = fmt.Errorf("invalid modifier relation: %q", rel)
		return sb
	}
	sb.mods = append(sb.mods, cql.Modifier{
		Name:     string(name),
		Relation: rel,
		Value:    EscapeSpecialChars(value),
	})
	return sb
}

// Term finalizes the search clause and returns an expression builder.
// It escapes backslashes, quotes, and masking characters (*, ?, ^) and disallows empty terms.
func (sb *SearchBuilder) Term(term string) *ExprBuilder {
	if strings.TrimSpace(term) == "" {
		sb.err = fmt.Errorf("search term must be non-empty")
		return &ExprBuilder{finish: sb.finish, build: sb.build, qb: sb.qb, end: sb.end, err: sb.err}
	}
	return sb.termWithEscaper(term, escapeSpecialAndMaskingChars)
}

// TermUnsafe finalizes the search clause and returns an expression builder.
// It does not escape or alter the term.
func (sb *SearchBuilder) TermUnsafe(term string) *ExprBuilder {
	return sb.termWithEscaper(term, identityValue)
}

func (sb *SearchBuilder) termWithEscaper(term string, esc func(string) string) *ExprBuilder {
	if sb.err != nil {
		return &ExprBuilder{finish: sb.finish, build: sb.build, qb: sb.qb, end: sb.end, err: sb.err}
	}
	if sb.rel != "" && !isValidRelation(sb.rel) {
		sb.err = fmt.Errorf("invalid relation: %q", sb.rel)
		return &ExprBuilder{finish: sb.finish, build: sb.build, qb: sb.qb, end: sb.end, err: sb.err}
	}
	clause := cql.Clause{
		SearchClause: &cql.SearchClause{
			Index:     sb.index,
			Relation:  sb.rel,
			Modifiers: sb.mods,
			Term:      esc(term),
		},
	}
	return sb.finish(clause)
}

// Escapes backslashes and quotes in a string.
func EscapeSpecialChars(s string) string {
	if s == "" {
		return s
	}
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}

// Escapes masking characters (*, ?, ^) in a string.
func EscapeMaskingChars(s string) string {
	if s == "" {
		return s
	}
	s = strings.ReplaceAll(s, "*", "\\*")
	s = strings.ReplaceAll(s, "?", "\\?")
	s = strings.ReplaceAll(s, "^", "\\^")
	return s
}

func escapeSpecialAndMaskingChars(s string) string {
	return EscapeMaskingChars(EscapeSpecialChars(s))
}

func identityValue(s string) string {
	return s
}

func isValidRelation(rel cql.Relation) bool {
	switch rel {
	case cql.EQ, cql.NE, cql.LT, cql.GT, cql.LE, cql.GE,
		cql.ADJ, cql.ALL, cql.ANY, cql.SCR, cql.ENCLOSES,
		cql.EXACT, cql.WITHIN:
		return true
	default:
		return false
	}
}

func isValidOperator(op cql.Operator) bool {
	switch op {
	case cql.AND, cql.NOT, cql.OR, cql.PROX:
		return true
	default:
		return false
	}
}
