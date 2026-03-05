package pgcql

import (
	"strings"

	"github.com/indexdata/cql-go/cql"
)

type FieldCombo struct {
	ignoreError bool
	fields      []Field
}

func NewFieldCombo(ignoreError bool, fields []Field) *FieldCombo {
	return &FieldCombo{ignoreError: ignoreError, fields: fields}
}

func (f *FieldCombo) GetColumn() string {
	return ""
}

func (f *FieldCombo) SetColumn(column string) {
}

func (f *FieldCombo) Generate(sc cql.SearchClause, queryArgumentIndex int) (string, []any, error) {
	var sqlParts []string
	var args []any
	var err error
	for _, field := range f.fields {
		var sql string
		var fieldArgs []any
		sql, fieldArgs, err = field.Generate(sc, queryArgumentIndex+len(args))
		if err != nil {
			if f.ignoreError {
				continue
			}
			return "", nil, err
		}
		sqlParts = append(sqlParts, sql)
		args = append(args, fieldArgs...)
	}
	if len(sqlParts) == 0 {
		if err != nil {
			return "", nil, err
		}
		return "TRUE", args, nil
	}
	return "(" + strings.Join(sqlParts, " OR ") + ")", args, nil
}
