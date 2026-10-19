package pgcql

import (
	"fmt"

	"github.com/indexdata/cql-go/cql"
)

// FieldExistsString searches string values in related rows using correlated EXISTS
// subqueries. Positive relations match any related row; <> matches only when no
// related row matches the corresponding equality predicate, including when there
// are no related rows. Empty equality terms match when no related row has a
// non-NULL, non-empty value; empty inequality terms require such a value.
// Sorting is disabled unless WithSortExpression supplies an outer-query expression.
type FieldExistsString struct {
	column          string
	table           string
	alias           string
	correlation     string
	valueExpression string
	sortExpression  string
	field           *FieldString
}

// NewFieldExistsString creates a field using case-sensitive, wildcard-aware string
// matching. table, alias, correlation, and valueExpression are trusted SQL
// fragments supplied by the application, never user input. alias may be empty;
// correlation links the related rows to the outer query. Query terms are bound
// as parameters. Use WithField to customize string matching.
func NewFieldExistsString(table, alias, correlation, valueExpression string) *FieldExistsString {
	field := NewFieldString().WithLikeOps()
	field.SetColumn(valueExpression)
	return &FieldExistsString{
		table:           table,
		alias:           alias,
		correlation:     correlation,
		valueExpression: valueExpression,
		field:           field,
	}
}

// GetColumn returns the field registration name, not the related value expression.
func (f *FieldExistsString) GetColumn() string {
	return f.column
}

// SetColumn sets the registration name without changing the related value expression.
func (f *FieldExistsString) SetColumn(column string) {
	f.column = column
}

// Sort returns the configured outer-query sort expression, or empty to reject sorting.
func (f *FieldExistsString) Sort() string {
	return f.sortExpression
}

// WithSortExpression enables sorting by a trusted SQL expression valid in the outer query.
func (f *FieldExistsString) WithSortExpression(sortExpression string) *FieldExistsString {
	f.sortExpression = sortExpression
	return f
}

// WithField replaces the string matcher and sets its column to the related value
// expression. field must be non-nil and must not be shared with other definitions.
func (f *FieldExistsString) WithField(field *FieldString) *FieldExistsString {
	f.field = field
	f.field.SetColumn(f.valueExpression)
	return f
}

// Generate translates a search clause, numbering parameters from queryArgumentIndex.
func (f *FieldExistsString) Generate(sc cql.SearchClause, queryArgumentIndex int) (string, []any, error) {
	switch {
	case sc.Term == "" && isPositiveStringRelation(sc.Relation):
		return "NOT " + f.existsSql(f.nonEmptyValuePredicate()), nil, nil
	case sc.Term == "" && sc.Relation == cql.NE:
		return f.existsSql(f.nonEmptyValuePredicate()), nil, nil
	case sc.Relation == cql.NE:
		sc.Relation = cql.EQ
		predicate, args, err := f.field.Generate(sc, queryArgumentIndex)
		if err != nil {
			return "", nil, err
		}
		return "NOT " + f.existsSql(predicate), args, nil
	default:
		predicate, args, err := f.field.Generate(sc, queryArgumentIndex)
		if err != nil {
			return "", nil, err
		}
		return f.existsSql(predicate), args, nil
	}
}

func isPositiveStringRelation(relation cql.Relation) bool {
	return relation == cql.EQ || relation == cql.EXACT || relation == "=="
}

func (f *FieldExistsString) existsSql(predicate string) string {
	source := f.table
	if f.alias != "" {
		source += " " + f.alias
	}
	return fmt.Sprintf("EXISTS (SELECT 1 FROM %s WHERE %s AND %s)", source, f.correlation, predicate)
}

func (f *FieldExistsString) nonEmptyValuePredicate() string {
	return fmt.Sprintf("%s IS NOT NULL AND %s <> ''", f.valueExpression, f.valueExpression)
}
