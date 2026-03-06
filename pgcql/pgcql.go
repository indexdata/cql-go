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
	Sort() string
}

type Definition interface {
	AddField(name string, field Field) Definition
	GetFieldType(name string) Field
	Parse(q cql.Query, queryArgumentIndex int) (Query, error)
}

type Query interface {
	// GetWhereClause returns the SQL WHERE clause generated from the CQL query,
	// without the "WHERE" keyword.
	// The returned string will contain parameter placeholders (e.g. $1, $2, etc.)
	// corresponding to the query arguments returned by GetQueryArguments.
	GetWhereClause() string
	// GetQueryArguments returns the list of query arguments to be used in the SQL
	// query, in the order they should be applied.
	GetQueryArguments() []any
	// GetOrderByClause returns the SQL ORDER BY clause, or an empty string if no
	// sorting is specified.
	// If the query includes sorting, the returned string will start with " ORDER BY "
	// followed by the sorting fields and directions.
	GetOrderByClause() string
	// GetOrderByFields returns a list of fields used in the ORDER BY clause, or an
	// empty list if no sorting is specified.
	GetOrderByFields() []string
}
