package cql

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

type Sort struct {
	Index     string
	Modifiers []Modifier
}

type Modifier struct {
	Name     string
	Relation Relation
	Value    string
}

type Clause struct {
	PrefixMap    []Prefix
	SearchClause *SearchClause
	BoolClause   *BoolClause
}

type Prefix struct {
	Prefix string
	Uri    string
}

type SearchClause struct {
	Index     string
	Relation  Relation
	Modifiers []Modifier
	Term      string
}

type BoolClause struct {
	Left      Clause
	Operator  Operator
	Modifiers []Modifier
	Right     Clause
}
