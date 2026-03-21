package pgcql

import (
	"fmt"
	"strings"

	"github.com/indexdata/cql-go/cql"
)

type FieldBool struct {
	FieldCommon
}

func NewFieldBool() *FieldBool {
	return &FieldBool{}
}

func (f *FieldBool) WithColumn(column string) *FieldBool {
	f.column = column
	return f
}

func (f *FieldBool) Generate(sc cql.SearchClause, queryArgumentIndex int) (string, []any, error) {
	s := f.handleEmptyTerm(sc)
	if s != "" {
		return s, []any{}, nil
	}
	relOrdered, err := f.handleUnorderedRelation(sc)
	if err != nil {
		return "", nil, err
	}

	// Map string values to boolean
	var boolValue bool
	term := strings.Join(sc.Terms, " ")
	switch strings.ToLower(term) {
	case "true", "1", "yes", "on":
		boolValue = true
	case "false", "0", "no", "off":
		boolValue = false
	default:
		return "", nil, &PgError{message: fmt.Sprintf("invalid bool %s", term)}
	}

	return f.column + " " + relOrdered + fmt.Sprintf(" $%d", queryArgumentIndex), []any{boolValue}, nil
}
