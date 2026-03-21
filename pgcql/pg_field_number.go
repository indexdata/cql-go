package pgcql

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/indexdata/cql-go/cql"
)

type FieldNumber struct {
	FieldCommon
}

func NewFieldNumber() *FieldNumber {
	return &FieldNumber{}
}

func (f *FieldNumber) WithColumn(column string) *FieldNumber {
	f.column = column
	return f
}

func (f *FieldNumber) Generate(sc cql.SearchClause, queryArgumentIndex int) (string, []any, error) {
	s := f.handleEmptyTerm(sc)
	if s != "" {
		return s, []any{}, nil
	}
	relOrdered, err := f.handleOrderedRelation(sc)
	if err != nil {
		return "", nil, err
	}
	term := strings.Join(sc.Terms, " ")
	number, err := strconv.ParseFloat(term, 64)
	if err != nil {
		return "", nil, &PgError{message: fmt.Sprintf("invalid number %s", term)}
	}
	return f.column + " " + relOrdered + fmt.Sprintf(" $%d", queryArgumentIndex), []any{number}, nil
}
