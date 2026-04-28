package pgcql

import (
	"fmt"
	"strings"

	"github.com/indexdata/cql-go/cql"
)

type FieldString struct {
	FieldCommon
	language        string
	assumeTsVector  bool
	enableLower     bool
	enableLike      bool
	enableILike     bool
	enableExact     bool
	enableSplit     bool
	serverChoiceRel cql.Relation
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
	f.enableILike = false
	return f
}

// WithILikeOps enables wildcard-aware case-insensitive matching and disables exact match fallback.
// For good performance, this is typically paired with a pg_trgm GIN/GiST index.
func (f *FieldString) WithILikeOps() *FieldString {
	f.enableExact = false
	f.enableILike = true
	f.enableLike = false
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

func (f *FieldString) WithServerChoiceRel(relation cql.Relation) *FieldString {
	f.serverChoiceRel = relation
	return f
}

func (f *FieldString) WithAssumeTsVector() *FieldString {
	f.assumeTsVector = true
	return f
}

// WithLower applies lower() to column and term.
// Ignored when using WithFullText/WithAssumeTsVector/WithILikeOps.
func (f *FieldString) WithLower() *FieldString {
	f.enableLower = true
	return f
}

func (f *FieldString) getQueryColumn() string {
	if f.enableLower {
		return "lower(" + f.column + ")"
	}
	return f.column
}

func (f *FieldString) getQueryArg(index int) string {
	if f.enableLower {
		return "lower(" + fmt.Sprintf("$%d", index) + ")"
	}
	return fmt.Sprintf("$%d", index)
}

func appendMaskedChar(pgTerm []rune, c rune) ([]rune, error) {
	switch c {
	case '*', '"', '?', '^', '\\':
		return append(pgTerm, c), nil
	default:
		return pgTerm, fmt.Errorf("a masking backslash in a CQL string must be followed by *, ?, ^, \" or \\")
	}
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
			var err error
			pgTerm, err = appendMaskedChar(pgTerm, c)
			if err != nil {
				return terms, err
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

func maskedSplitTsTerms(cqlTerm string, splitChars string) ([]string, error) {
	terms := make([]string, 0)
	var pgTerm []rune
	backslash := false
	wildcard := false

	appendTerm := func() {
		if len(pgTerm) == 0 {
			return
		}
		term := "'" + strings.ReplaceAll(string(pgTerm), "'", "''") + "'"
		if wildcard {
			term += ":*"
		}
		terms = append(terms, term)
		pgTerm = []rune{}
		wildcard = false
	}

	for _, c := range cqlTerm {
		if backslash {
			if wildcard {
				return terms, fmt.Errorf("masking op * supported only at end of term")
			}
			var err error
			pgTerm, err = appendMaskedChar(pgTerm, c)
			if err != nil {
				return terms, err
			}
			backslash = false
			continue
		}

		switch c {
		case '*':
			if wildcard {
				return terms, fmt.Errorf("masking op * supported only at end of term")
			}
			if len(pgTerm) == 0 {
				return terms, fmt.Errorf("masking op * unsupported")
			}
			wildcard = true
		case '?':
			return terms, fmt.Errorf("masking op ? unsupported")
		case '^':
			return terms, fmt.Errorf("anchor op ^ unsupported")
		case '\\':
			backslash = true
		default:
			if strings.ContainsRune(splitChars, c) {
				appendTerm()
				continue
			}
			if wildcard {
				return terms, fmt.Errorf("masking op * supported only at end of term")
			}
			pgTerm = append(pgTerm, c)
		}
	}
	if backslash {
		return terms, fmt.Errorf("a CQL string must not end with a masking backslash")
	}
	appendTerm()
	if len(terms) == 0 {
		terms = append(terms, "''")
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

func (f *FieldString) generateTsQuery(sc cql.SearchClause, termOp string, queryArgumentIndex int) (string, []any, error) {
	pgTerms, err := maskedSplitTsTerms(sc.Term, " ")
	if err != nil {
		return "", nil, err
	}
	sql := ""
	if f.assumeTsVector {
		sql += f.column + " "
	} else {
		sql += "to_tsvector('" + f.language + "', " + f.column + ") "
	}
	sql += "@@ to_tsquery('" + f.language + "', " + fmt.Sprintf("$%d", queryArgumentIndex) + ")"
	return sql, []any{strings.Join(pgTerms, termOp)}, nil
}

func (f *FieldString) generateIn(sc cql.SearchClause, queryArgumentIndex int, not bool) (string, []any, error) {
	pgTerms, err := maskedSplit(sc.Term, " ")
	if err != nil {
		return "", nil, err
	}
	sql := f.getQueryColumn()
	if not {
		sql += " NOT"
	}
	sql += " IN("
	anyTerms := make([]any, len(pgTerms))
	for i, v := range pgTerms {
		if i > 0 {
			sql += ", "
		}
		sql += f.getQueryArg(queryArgumentIndex + i)
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
	if f.serverChoiceRel != "" && (sc.Relation == cql.EQ || sc.Relation == cql.SCR) {
		sc.Relation = f.serverChoiceRel
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
	if (f.enableLike || f.enableILike) && (sc.Relation == cql.EQ || sc.Relation == cql.EXACT || sc.Relation == cql.NE) {
		pgTerm, ops, err := maskedLike(sc.Term)
		if err != nil {
			return "", nil, err
		}
		if !f.enableExact || ops {
			pgOp := "LIKE"
			if f.enableILike {
				pgOp = "ILIKE"
			}
			if sc.Relation == cql.NE {
				pgOp = "NOT LIKE"
				if f.enableILike {
					pgOp = "NOT ILIKE"
				}
			}
			if f.enableILike {
				return f.column + " " + pgOp + fmt.Sprintf(" $%d", queryArgumentIndex), []any{pgTerm}, nil
			}
			return f.getQueryColumn() + " " + pgOp + " " + f.getQueryArg(queryArgumentIndex), []any{pgTerm}, nil
		}
	}
	if !f.enableExact {
		return "", nil, &PgError{message: "unsupported relation " + string(sc.Relation)}
	}
	pgTerm, err := maskedExact(sc.Term)
	if err != nil {
		return "", nil, err
	}
	pgOp, err := f.handleUnorderedRelation(sc)
	if err != nil {
		return "", nil, err
	}
	return f.getQueryColumn() + " " + pgOp + " " + f.getQueryArg(queryArgumentIndex), []any{pgTerm}, nil
}
