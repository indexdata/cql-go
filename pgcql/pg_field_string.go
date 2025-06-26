package pgcql

import (
	"fmt"

	"github.com/indexdata/cql-go/cql"
)

type FieldString struct {
	FieldCommon
	language    string
	enableLike  bool
	enableExact bool
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

func maskedExact(cqlTerm string) (string, error) {
	var pgTerm []rune
	backslash := false

	for _, c := range cqlTerm {
		if backslash {
			switch c {
			case '*', '"', '?', '^', '\\':
				pgTerm = append(pgTerm, c)
			default:
				return "", fmt.Errorf("a masking backslash in a CQL string must be followed by *, ?, ^, \" or \\")
			}
			backslash = false
		} else {
			switch c {
			case '*':
				return "", fmt.Errorf("masking op * unsupported")
			case '?':
				return "", fmt.Errorf("masking op ? unsupported")
			case '^':
				return "", fmt.Errorf("anchor op ^ unsupported")
			case '\\':
				// Do nothing, just set backslash to true
			default:
				pgTerm = append(pgTerm, c)
			}
			backslash = c == '\\'
		}
	}
	if backslash {
		return "", fmt.Errorf("a CQL string must not end with a masking backslash")
	}
	return string(pgTerm), nil
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
				// Do nothing, just set backslash to true
			case '%', '_':
				pgTerm = append(pgTerm, '\\', c)
			default:
				pgTerm = append(pgTerm, c)
			}
			backslash = c == '\\'
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

func (f *FieldString) Generate(sc cql.SearchClause, queryArgumentIndex int) (string, []any, error) {
	sql := f.handleEmptyTerm(sc)
	if sql != "" {
		return sql, nil, nil
	}
	fulltext := f.language != ""
	var pgfunc string
	if fulltext {
		if sc.Relation == cql.ADJ || sc.Relation == cql.EQ {
			pgfunc = "phraseto_tsquery"
		} else if sc.Relation == cql.ALL {
			pgfunc = "plainto_tsquery"
		}
	}
	if pgfunc != "" {
		pgTerm, err := maskedExact(sc.Term)
		if err != nil {
			return "", nil, err
		}
		sql := "to_tsvector('" + f.language + "', " + f.column + ") @@ " + pgfunc + "('" + f.language + "', " + fmt.Sprintf("$%d", queryArgumentIndex) + ")"
		return sql, []any{pgTerm}, nil
	}
	if !f.enableExact {
		return "", nil, &PgError{message: "exact search not supported"}
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
