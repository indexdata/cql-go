package pgcql

import (
	"github.com/indexdata/cql-go/cql"
)

type FieldCommon struct {
	column string
}

func (f *FieldCommon) GetColumn() string {
	return f.column
}
func (f *FieldCommon) SetColumn(column string) {
	f.column = column
}

func (f *FieldCommon) handleUnorderedRelation(sc cql.SearchClause) (string, error) {
	switch sc.Relation {
	case "==", cql.EXACT, cql.EQ:
		return "=", nil
	case cql.NE:
		return "<>", nil
	default:
		return "", &PgError{message: "unsupported relation " + string(sc.Relation)}
	}
}

func (f *FieldCommon) handleOrderedRelation(sc cql.SearchClause) (string, error) {
	switch sc.Relation {
	case "==", cql.EXACT:
		return "=", nil
	case "=", "<>", ">", "<", "<=", ">=":
		return string(sc.Relation), nil
	default:
		return "", &PgError{message: "unsupported relation " + string(sc.Relation)}
	}
}

func (f *FieldCommon) handleEmptyTerm(sc cql.SearchClause) string {
	if sc.Term == "" && sc.Relation == cql.EQ {
		return f.column + " IS NOT NULL"
	}
	return ""
}
