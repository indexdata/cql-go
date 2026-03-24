package pgcql

import (
	"fmt"
	"strings"

	"github.com/indexdata/cql-go/cql"
)

type FieldTsVector struct {
	FieldCommon
	language        string
	serverChoiceRel cql.Relation
}

func NewFieldTsVector() *FieldTsVector {
	return &FieldTsVector{language: "simple"}
}

func (f *FieldTsVector) WithColumn(column string) *FieldTsVector {
	f.column = column
	return f
}

func (f *FieldTsVector) WithLanguage(language string) *FieldTsVector {
	if language == "" {
		f.language = "simple"
	} else {
		f.language = language
	}
	return f
}

func (f *FieldTsVector) WithServerChoiceRel(relation cql.Relation) *FieldTsVector {
	f.serverChoiceRel = relation
	return f
}

func (f *FieldTsVector) Generate(sc cql.SearchClause, queryArgumentIndex int) (string, []any, error) {
	sql := f.handleEmptyTerm(sc)
	if sql != "" {
		return sql, nil, nil
	}
	if f.serverChoiceRel != "" && (sc.Relation == cql.EQ || sc.Relation == cql.SCR) {
		sc.Relation = f.serverChoiceRel
	}
	switch sc.Relation {
	case cql.ADJ, cql.EQ:
		return f.generateTsQuery(sc, "<->", queryArgumentIndex)
	case cql.ALL:
		return f.generateTsQuery(sc, "&", queryArgumentIndex)
	case cql.ANY:
		return f.generateTsQuery(sc, "|", queryArgumentIndex)
	}
	return "", nil, &PgError{message: "unsupported relation " + string(sc.Relation)}
}

func (f *FieldTsVector) generateTsQuery(sc cql.SearchClause, termOp string, queryArgumentIndex int) (string, []any, error) {
	pgTerms, err := maskedSplit(sc.Term, " ")
	if err != nil {
		return "", nil, err
	}
	for i, v := range pgTerms {
		pgTerms[i] = "'" + strings.ReplaceAll(v, "'", "''") + "'"
	}
	sql := f.column + " @@ to_tsquery('" + f.language + "', " + fmt.Sprintf("$%d", queryArgumentIndex) + ")"
	return sql, []any{strings.Join(pgTerms, termOp)}, nil
}
