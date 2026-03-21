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
	relOrdered, err := f.handleOrderedRelation(sc)
	if err != nil {
		return "", nil, err
	}
	term := strings.Join(sc.Terms, " ")
	number, err := f.parseTerm(term)
	if err != nil {
		if f.isDate {
			return "", nil, &PgError{message: fmt.Sprintf("invalid date %s, it should be in format YYYY-MM-DD", term)}
		} else {
			return "", nil, &PgError{message: fmt.Sprintf("invalid date time %s, it should be in format YYYY-MM-DD, YYYY-MM-DD HH:MM:SS, YYYY-MM-DDTHH:MM:SSZ, YYYY-MM-DDTHH:MM:SS±HH:MM", term)}
		}
	}
	return f.column + " " + relOrdered + fmt.Sprintf(" $%d", queryArgumentIndex), []any{number}, nil
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
