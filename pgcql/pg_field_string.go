package pgcql

import (
	"fmt"
	"strings"

	"github.com/indexdata/cql-go/cql"
)

type FieldString struct {
	FieldCommon
	language    string
	enableLike  bool
	enableExact bool
	enableSplit bool
}

func NewFieldString() *FieldString {
	return &FieldString{}
}

func (f *FieldString) WithColumn(column string) *FieldString {
	f.column = column
	return f
}

func (f *FieldString) WithFullText(language string) *FieldString {
	if language == "" {
		f.language = "simple"
	} else {
		f.language = language
	}
	return f
}

func (f *FieldString) WithLikeOps() *FieldString {
	f.enableExact = true
	f.enableLike = true
	return f
}

func (f *FieldString) WithExact() *FieldString {
	f.enableExact = true
	return f
}

func (f *FieldString) WithSplit() *FieldString {
	f.enableSplit = true
	return f
}

func maskedExact(cqlTerm string) (string, error) {
	terms, err := maskedSplit(cqlTerm, "")
	if err != nil {
		return "", err
	}
	return terms[0], nil
}

func maskedSplit(cqlTerm string, splitChars string) ([]string, error) {
	terms := make([]string, 0)
	var pgTerm []rune
	backslash := false

	for _, c := range cqlTerm {
		if backslash {
			switch c {
			case '*', '"', '?', '^', '\\':
				pgTerm = append(pgTerm, c)
			default:
				return terms, fmt.Errorf("a masking backslash in a CQL string must be followed by *, ?, ^, \" or \\")
			}
			backslash = false
		} else {
			switch c {
			case '*':
				return terms, fmt.Errorf("masking op * unsupported")
			case '?':
				return terms, fmt.Errorf("masking op ? unsupported")
			case '^':
				return terms, fmt.Errorf("anchor op ^ unsupported")
			case '\\':
				backslash = true
			default:
				if strings.ContainsRune(splitChars, c) {
					if len(pgTerm) > 0 {
						terms = append(terms, string(pgTerm))
					}
					pgTerm = []rune{}
					continue
				}
				pgTerm = append(pgTerm, c)
			}
		}
	}
	if backslash {
		return terms, fmt.Errorf("a CQL string must not end with a masking backslash")
	}
	if len(pgTerm) > 0 || len(terms) == 0 {
		terms = append(terms, string(pgTerm))
	}
	return terms, nil
}

func maskedLike(cqlTerm string) (string, bool, error) {
	var pgTerm []rune
	ops := false
	backslash := false

	for _, c := range cqlTerm {
		if backslash {
			switch c {
			case '*', '?', '^', '"':
				pgTerm = append(pgTerm, c)
			case '\\':
				pgTerm = append(pgTerm, '\\', '\\')
			default:
				return "", false, fmt.Errorf("a masking backslash in a CQL string must be followed by *, ?, ^, \" or \\")
			}
			backslash = false
		} else {
			switch c {
			case '*':
				pgTerm = append(pgTerm, '%')
				ops = true
			case '?':
				pgTerm = append(pgTerm, '_')
				ops = true
			case '^':
				return "", false, fmt.Errorf("anchor op ^ unsupported")
			case '\\':
				backslash = true
			case '%', '_':
				pgTerm = append(pgTerm, '\\', c)
			default:
				pgTerm = append(pgTerm, c)
			}
		}
	}
	if backslash {
		return "", false, fmt.Errorf("a CQL string must not end with a masking backslash")
	}
	return string(pgTerm), ops, nil
}

func (f *FieldString) handleEmptyTerm(sc cql.SearchClause) string {
	if sc.Term == "" && sc.Relation == cql.EQ {
		return f.column + " IS NOT NULL"
	}
	return ""
}

func (f *FieldString) generateTsQuery(sc cql.SearchClause, termOp string, queryArgumentIndex int) (string, []any, error) {
	pgTerms, err := maskedSplit(sc.Term, " ")
	if err != nil {
		return "", nil, err
	}
	for i, v := range pgTerms {
		pgTerms[i] = "'" + strings.ReplaceAll(v, "'", "''") + "'"
	}
	sql := "to_tsvector('" + f.language + "', " + f.column + ") @@ to_tsquery('" + f.language + "', " + fmt.Sprintf("$%d", queryArgumentIndex) + ")"
	return sql, []any{strings.Join(pgTerms, termOp)}, nil
}

func (f *FieldString) generateIn(sc cql.SearchClause, queryArgumentIndex int, not bool) (string, []any, error) {
	pgTerms, err := maskedSplit(sc.Term, " ")
	if err != nil {
		return "", nil, err
	}
	sql := f.column
	if not {
		sql += " NOT"
	}
	sql += " IN("
	anyTerms := make([]any, len(pgTerms))
	for i, v := range pgTerms {
		if i > 0 {
			sql += ", "
		}
		sql += fmt.Sprintf("$%d", queryArgumentIndex+i)
		anyTerms[i] = v
	}
	sql += ")"
	return sql, anyTerms, nil
}

func (f *FieldString) Generate(sc cql.SearchClause, queryArgumentIndex int) (string, []any, error) {
	sql := f.handleEmptyTerm(sc)
	if sql != "" {
		return sql, nil, nil
	}
	fulltext := f.language != ""
	if fulltext {
		switch sc.Relation {
		case cql.ADJ, cql.EQ:
			return f.generateTsQuery(sc, "<->", queryArgumentIndex)
		case cql.ALL:
			return f.generateTsQuery(sc, "&", queryArgumentIndex)
		case cql.ANY:
			return f.generateTsQuery(sc, "|", queryArgumentIndex)
		}
	}
	if f.enableSplit {
		if sc.Relation == cql.ANY {
			return f.generateIn(sc, queryArgumentIndex, false)
		}
		if sc.Relation == cql.NE {
			return f.generateIn(sc, queryArgumentIndex, true)
		}
	}
	if !f.enableExact {
		return "", nil, &PgError{message: "unsupported relation " + string(sc.Relation)}
	}
	if f.enableLike && (sc.Relation == cql.EQ || sc.Relation == cql.EXACT || sc.Relation == cql.NE) {
		pgTerm, ops, err := maskedLike(sc.Term)
		if err != nil {
			return "", nil, err
		}
		if ops {
			pgOp := "LIKE"
			if sc.Relation == cql.NE {
				pgOp = "NOT LIKE"
			}
			return f.column + " " + pgOp + fmt.Sprintf(" $%d", queryArgumentIndex), []any{pgTerm}, nil
		}
	}
	pgTerm, err := maskedExact(sc.Term)
	if err != nil {
		return "", nil, err
	}
	pgOp, err := f.handleUnorderedRelation(sc)
	if err != nil {
		return "", nil, err
	}
	return f.column + " " + pgOp + fmt.Sprintf(" $%d", queryArgumentIndex), []any{pgTerm}, nil
}
