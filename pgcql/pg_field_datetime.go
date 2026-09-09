package pgcql

import (
	"fmt"
	"strings"
	"time"

	"github.com/indexdata/cql-go/cql"
)

const dateFormat = "2006-01-02"
const dateTimeFormat = "2006-01-02 15:04:05"

type FieldDateTime struct {
	FieldCommon
	isDate bool
}

func NewFieldDate() *FieldDateTime {
	return &FieldDateTime{}
}

func (f *FieldDateTime) WithColumn(column string) *FieldDateTime {
	f.column = column
	return f
}

func (f *FieldDateTime) WithOnlyDate() *FieldDateTime {
	f.isDate = true
	return f
}

func (f *FieldDateTime) Generate(sc cql.SearchClause, queryArgumentIndex int) (string, []any, error) {
	s := f.handleEmptyTerm(sc)
	if s != "" {
		return s, []any{}, nil
	}
	if sc.Relation == cql.WITHIN {
		return f.generateWithin(sc.Term, queryArgumentIndex)
	}
	relOrdered, err := f.handleOrderedRelation(sc)
	if err != nil {
		return "", nil, err
	}
	number, err := f.parseTerm(sc.Term)
	if err != nil {
		if f.isDate {
			return "", nil, &PgError{message: fmt.Sprintf("invalid date %s, it should be in format YYYY-MM-DD", sc.Term)}
		} else {
			return "", nil, &PgError{message: fmt.Sprintf("invalid date time %s, it should be in format YYYY-MM-DD, YYYY-MM-DD HH:MM:SS, YYYY-MM-DDTHH:MM:SSZ, YYYY-MM-DDTHH:MM:SS±HH:MM", sc.Term)}
		}
	}
	return f.column + " " + relOrdered + fmt.Sprintf(" $%d", queryArgumentIndex), []any{number}, nil
}

func (f *FieldDateTime) generateWithin(term string, queryArgumentIndex int) (string, []any, error) {
	parts := strings.Fields(term)
	// Each endpoint can contain a space between its date and time.
	if len(parts) >= 2 && len(parts) <= 4 {
		for i := 1; i < len(parts); i++ {
			lower, err := f.parseTerm(strings.Join(parts[:i], " "))
			if err != nil {
				continue
			}
			upper, err := f.parseTerm(strings.Join(parts[i:], " "))
			if err != nil {
				continue
			}
			return fmt.Sprintf("(%s >= $%d AND %s <= $%d)", f.column, queryArgumentIndex, f.column, queryArgumentIndex+1), []any{lower, upper}, nil
		}
	}
	return "", nil, &PgError{message: fmt.Sprintf("invalid within range %s, expected two valid date or date time endpoints", term)}
}

func (f *FieldDateTime) parseTerm(term string) (time.Time, error) {
	if f.isDate {
		date, err := time.Parse(dateFormat, term)
		if err != nil {
			return time.Time{}, err
		}
		return date, nil
	} else {
		layouts := []string{
			dateFormat,
			dateTimeFormat,
			time.RFC3339,
		}
		var err error
		for _, layout := range layouts {
			t, e := time.Parse(layout, term)
			if e == nil {
				return t, nil
			}
			err = e
		}
		return time.Time{}, err
	}
}
