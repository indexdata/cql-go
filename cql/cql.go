package cql

import (
	"strings"
)

type Relation string

const (
	EQ       Relation = "="
	NE       Relation = "<>"
	LT       Relation = "<"
	GT       Relation = ">"
	LE       Relation = "<="
	GE       Relation = ">="
	ADJ      Relation = "adj"
	ALL      Relation = "all"
	ANY      Relation = "any"
	SCR      Relation = "scr"
	ENCLOSES Relation = "encloses"
	EXACT    Relation = "exact"
	WITHIN   Relation = "within"
)

type Operator string

const (
	AND  Operator = "and"
	NOT  Operator = "not"
	OR   Operator = "or"
	PROX Operator = "prox"
)

type CqlIndex string

const (
	AllRecords   CqlIndex = "cql.allRecords"
	AllIndexes   CqlIndex = "cql.allIndexes"
	AnyIndexes   CqlIndex = "cql.anyIndexes"
	Anywhere     CqlIndex = "cql.anywhere"
	Keywords     CqlIndex = "cql.keywords"
	ServerChoice CqlIndex = "cql.serverChoice"
	ResultSetId  CqlIndex = "cql.resultSetId"
)

type Query struct {
	Clause
	SortSpec []Sort
}

func (q *Query) write(sb *strings.Builder) {
	q.Clause.write(sb, false)
	if len(q.SortSpec) > 0 {
		sb.WriteString(" sortBy")
	}
	for _, sort := range q.SortSpec {
		sb.WriteString(" ")
		sort.write(sb)
	}
}

func (q *Query) String() string {
	var sb strings.Builder
	q.write(&sb)
	return sb.String()
}

type Sort struct {
	Index     string
	Modifiers []Modifier
}

func (s *Sort) write(sb *strings.Builder) {
	quote(sb, s.Index)
	for _, mod := range s.Modifiers {
		sb.WriteString("/")
		mod.write(sb)
	}
}

func (s *Sort) String() string {
	var sb strings.Builder
	s.write(&sb)
	return sb.String()
}

type Modifier struct {
	Name     string
	Relation Relation
	Value    string
}

func (m *Modifier) write(sb *strings.Builder) {
	quote(sb, m.Name)
	if m.Value != "" {
		sb.WriteString(defVal(string(m.Relation), string(EQ)))
		quote(sb, m.Value)
	}
}

func (m *Modifier) String() string {
	var sb strings.Builder
	m.write(&sb)
	return sb.String()
}

type Clause struct {
	PrefixMap    []Prefix
	SearchClause *SearchClause
	BoolClause   *BoolClause
}

func (c *Clause) write(sb *strings.Builder, brackets bool) {
	for _, p := range c.PrefixMap {
		p.write(sb)
		sb.WriteString(" ")
	}
	if c.SearchClause != nil {
		c.SearchClause.write(sb)
		return
	}
	if c.BoolClause != nil {
		if brackets {
			sb.WriteString("(")
		}
		c.BoolClause.write(sb)
		if brackets {
			sb.WriteString(")")
		}
		return
	}
	sb.WriteString(string(AllRecords))
	sb.WriteString(" ")
	sb.WriteString(string(EQ))
	sb.WriteString(" ")
	sb.WriteString("1")
}

func (c *Clause) String() string {
	var sb strings.Builder
	c.write(&sb, false)
	return sb.String()
}

type Prefix struct {
	Prefix string
	Uri    string
}

func (p *Prefix) write(sb *strings.Builder) {
	sb.WriteString("> ")
	if len(p.Prefix) > 0 {
		quote(sb, p.Prefix)
		sb.WriteString(" = ")
	}
	quote(sb, p.Uri)
}

func (p *Prefix) String() string {
	var sb strings.Builder
	p.write(&sb)
	return sb.String()
}

type SearchClause struct {
	Index     string
	Relation  Relation
	Modifiers []Modifier
	Term      string
}

func (sc *SearchClause) write(sb *strings.Builder) {
	idx := defVal(sc.Index, string(ServerChoice))
	rel := defVal(string(sc.Relation), string(EQ))
	if idx != string(ServerChoice) ||
		(rel != string(EQ) && rel != string(SCR)) {
		quote(sb, idx)
		sb.WriteString(" ")
		sb.WriteString(rel)
		for _, mod := range sc.Modifiers {
			sb.WriteString("/")
			mod.write(sb)
		}
		sb.WriteString(" ")
	}
	quote(sb, sc.Term)
}

func (sc *SearchClause) String() string {
	var sb strings.Builder
	sc.write(&sb)
	return sb.String()
}

type BoolClause struct {
	Left      Clause
	Operator  Operator
	Modifiers []Modifier
	Right     Clause
}

func (bc *BoolClause) write(sb *strings.Builder) {
	bc.Left.write(sb, false)
	sb.WriteString(" ")
	sb.WriteString(defVal(string(bc.Operator), string(AND)))
	sb.WriteString(" ")
	bc.Right.write(sb, true)
}

func (bc *BoolClause) String() string {
	var sb strings.Builder
	bc.write(&sb)
	return sb.String()
}

func quote(sb *strings.Builder, s string) {
	if s == "" || strings.ContainsAny(s, " ()=<>\"/") {
		sb.WriteString("\"")
		sb.WriteString(s)
		sb.WriteString("\"")
	} else {
		sb.WriteString(s)
	}
}

func defVal(val string, def string) string {
	if len(val) > 0 {
		return val
	} else {
		return def
	}
}
