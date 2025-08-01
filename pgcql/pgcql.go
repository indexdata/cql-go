package pgcql

import (
	"github.com/indexdata/cql-go/cql"
)

type PgError struct {
	message string
}

func (e *PgError) Error() string {
	return e.message
}

type Field interface {
	GetColumn() string
	SetColumn(column string)
	Generate(sc cql.SearchClause, queryArgumentIndex int) (string, []any, error)
}

type Definition interface {
	AddField(name string, field Field) Definition
	GetFieldType(name string) Field
	Parse(q cql.Query, queryArgumentIndex int) (Query, error)
}

type Query interface {
	GetWhereClause() string
	GetQueryArguments() []any
	GetOrderByClause() string
	GetOrderByFields() string
}
