package cql

import (
	"strings"
)

type Relation string

const (
	EQ Relation = "="
	NE Relation = "<>"
	LT Relation = "<"
	GT Relation = ">"
	LE Relation = "<="
	GE Relation = ">="
)

type Operator string

const (
	AND  Operator = "and"
	NOT  Operator = "not"
	OR   Operator = "or"
	PROX Operator = "prox"
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

func (q *Sort) String() string {
	var sb strings.Builder
	q.write(&sb)
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
		sb.WriteString(" ")
		sb.WriteString(string(m.Relation))
		sb.WriteString(" ")
	}
	quote(sb, m.Value)
}

func (q *Modifier) String() string {
	var sb strings.Builder
	q.write(&sb)
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
	}
	if c.BoolClause != nil {
		if brackets {
			sb.WriteString("(")
		}
		c.BoolClause.write(sb)
		if brackets {
			sb.WriteString(")")
		}
	}
}

func (q *Clause) String() string {
	var sb strings.Builder
	q.write(&sb, false)
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

type SearchClause struct {
	Index     string
	Relation  Relation
	Modifiers []Modifier
	Term      string
}

func (sc *SearchClause) write(sb *strings.Builder) {
	if sc.Index != "" {
		quote(sb, sc.Index)
		sb.WriteString(" ")
		quoteRel(sb, string(sc.Relation))
		for _, mod := range sc.Modifiers {
			sb.WriteString("/")
			mod.write(sb)
		}
		sb.WriteString(" ")
	}
	quote(sb, sc.Term)
}

type BoolClause struct {
	Left      Clause
	Operator  Operator
	Modifiers []Modifier
	Right     Clause
}

func (bc *BoolClause) write(sb *strings.Builder) {
	bc.Left.write(sb, true)
	sb.WriteString(" ")
	sb.WriteString(string(bc.Operator))
	sb.WriteString(" ")
	bc.Right.write(sb, true)
}

func quote(sb *strings.Builder, s string) {
	if strings.ContainsAny(s, " ()=<>\"/") {
		sb.WriteString("\"")
		sb.WriteString(s)
		sb.WriteString("\"")
	} else {
		sb.WriteString(s)
	}
}

func quoteRel(sb *strings.Builder, s string) {
	if strings.ContainsAny(s, " ()\"/") {
		sb.WriteString("\"")
		sb.WriteString(s)
		sb.WriteString("\"")
	} else {
		sb.WriteString(s)
	}
}
