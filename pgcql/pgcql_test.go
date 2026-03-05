package pgcql

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/indexdata/cql-go/cql"
	"github.com/stretchr/testify/assert"
)

func TestBadSearchClause(t *testing.T) {
	def := NewPgDefinition()

	assert.Nil(t, def.GetFieldType("foo"))

	q := cql.Query{}
	_, err := def.Parse(q, 1)
	assert.Error(t, err, "Expected error for empty query")
	assert.Equal(t, "unsupported clause type", err.Error())
}

func TestOrderByFields(t *testing.T) {
	def := NewPgDefinition()
	title := &FieldString{}
	title.WithExact().SetColumn("Title")
	def.AddField("title", title)

	tag := &FieldString{}
	tag.WithSplit().WithExact().SetColumn("T")
	def.AddField("tag", tag)

	for _, testcase := range []struct {
		query    string
		expected []string
	}{
		{"title = a", []string{}},
		{"title = a sortby title", []string{"Title"}},
		{"title = a sortby title tag", []string{"Title", "T"}},
	} {
		var parser cql.Parser
		q, err := parser.Parse(testcase.query)
		assert.NoErrorf(t, err, "failed to parse cql query '%s'", testcase.query)
		pgQuery, err := def.Parse(q, 1)
		assert.NoErrorf(t, err, "failed to pg parse cql query '%s'", testcase.query)
		if !reflect.DeepEqual(pgQuery.GetOrderByFields(), testcase.expected) {
			t.Errorf("%s: Expected order by fields %v, got %v", testcase.query, testcase.expected, pgQuery.GetOrderByFields())
		}
	}
}

func TestParsing(t *testing.T) {
	def := NewPgDefinition()
	title := &FieldString{}
	title.WithExact().SetColumn("Title")

	assert.Equal(t, title.GetColumn(), "Title", "GetColumn() should return the column name")

	author := NewFieldString().WithLikeOps().WithColumn("Author")

	tag := &FieldString{}
	tag.WithSplit().WithExact().SetColumn("T")

	full := &FieldString{}
	full.WithFullText("english")

	serverChoice := NewFieldCombo(true, []Field{full, title, author})
	anyf := NewFieldCombo(false, []Field{full, title, author})
	alwaysTrue := NewFieldCombo(true, []Field{})

	def.AddField("title", title).
		AddField("author", author).
		AddField("cql.serverChoice", serverChoice).
		AddField("full", full).AddField("tag", tag).
		AddField("any", anyf).
		AddField("alwaysTrue", alwaysTrue)

	price := NewFieldNumber()
	def.AddField("price", price)

	dateField := NewFieldDate().WithOnlyDate()
	def.AddField("date", dateField)

	dateTimeField := NewFieldDate()
	def.AddField("datetime", dateTimeField)

	dateTimeWithZone, err := time.Parse(time.RFC3339, "2026-03-05T09:34:27+01:00")
	assert.NoError(t, err)

	for _, testcase := range []struct {
		query        string
		expected     string
		expectedArgs []any
	}{
		{"tag = abc", "T = $1", []any{"abc"}},
		{"tag = \"\"", "T IS NOT NULL", []any{}},
		{"au=2", "error: unknown field au", nil},
		{"title>2", "error: unsupported relation >", nil},
		{"title=2", "Title = $1", []any{"2"}},
		{"title==2", "Title = $1", []any{"2"}},
		{"title exact 2", "Title = $1", []any{"2"}},
		{"title<>2", "Title <> $1", []any{"2"}},
		{"tag any \"1 23 45\"", "T IN($1, $2, $3)", []any{"1", "23", "45"}},
		{"tag <> \" 1 23 45 \"", "T NOT IN($1, $2, $3)", []any{"1", "23", "45"}},
		{"tag any \"*\"", "error: masking op * unsupported", nil},
		{"tag=a or tag=b and tag=c", "(T = $1 OR T = $2) AND T = $3", []any{"a", "b", "c"}},
		{"title = abc", "Title = $1", []any{"abc"}},
		{"author = \"test\"", "Author = $1", []any{"test"}},
		{"author <> \"test\"", "Author <> $1", []any{"test"}},
		{"author = \"test*\"", "Author LIKE $1", []any{"test%"}},
		{"author <> \"test*\"", "Author NOT LIKE $1", []any{"test%"}},
		{"title = a AND author = b c", "Title = $1 AND Author = $2", []any{"a", "b c"}},
		{"title = 'a' OR author = 'b'", "Title = $1 OR Author = $2", []any{"'a'", "'b'"}},
		{"title = a NOT author = b", "Title = $1 AND NOT Author = $2", []any{"a", "b"}},
		{"a prox b", "error: unsupported operator prox", []any{}},
		{"author = a sortby title", "Author = $1 ORDER BY Title DESC", []any{"a"}},
		{"author = a sortby title/sort.descending author/sort.ascending", "Author = $1 ORDER BY Title DESC, Author ASC", []any{"a"}},
		{"author = a sortby gyf", "error: unknown field gyf", nil},
		{"author = a sortby any", "error: field any does not support sorting", nil},
		{"author = a sortby title/sort.foo", "error: unsupported sort modifier sort.foo", nil},
		{"au=2 or a", "error: unknown field au", nil},
		{"a or au=2", "error: unknown field au", nil},
		{"author=\"ab?%\"", "Author LIKE $1", []any{"ab_\\%"}},
		{"author=\"ab*_\"", "Author LIKE $1", []any{"ab%\\_"}},
		{"author=\"a^\"", "error: anchor op ^ unsupported", nil},
		{"author=\"a*\\", "error: a CQL string must not end with a masking backslash", nil},
		{"author=\"a*\\x\"", "error: a masking backslash in a CQL string must be followed by *, ?, ^, \" or \\", nil},
		{"author=\"a*\\*\\\"\\?\\^\\\\", "Author LIKE $1", []any{"a%*\"?^\\\\"}},
		{"author=\"a\\*\\\"\\?\\^\\\\", "Author = $1", []any{"a*\"?^\\"}},
		{"title=\"a*\"", "error: masking op * unsupported", nil},
		{"title=\"a?\"", "error: masking op ? unsupported", nil},
		{"title=\"a^\"", "error: anchor op ^ unsupported", nil},
		{"title=\"a\\*\"", "Title = $1", []any{"a*"}},
		{"title=\"a\\?\"", "Title = $1", []any{"a?"}},
		{"title=\"a\\^\"", "Title = $1", []any{"a^"}},
		{"title=\"a\\", "error: a CQL string must not end with a masking backslash", nil},
		{"title=\"a\\x\"", "error: a masking backslash in a CQL string must be followed by *, ?, ^, \" or \\", nil},
		{"full = abc", "to_tsvector('english', full) @@ to_tsquery('english', $1)", []any{"'abc'"}},
		{"full = \"abc\"", "to_tsvector('english', full) @@ to_tsquery('english', $1)", []any{"'abc'"}},
		{"full = \"abc \"", "to_tsvector('english', full) @@ to_tsquery('english', $1)", []any{"'abc'"}},
		{"full adj \"a b\"", "to_tsvector('english', full) @@ to_tsquery('english', $1)", []any{"'a'<->'b'"}},
		{"full all \"a b\"", "to_tsvector('english', full) @@ to_tsquery('english', $1)", []any{"'a'&'b'"}},
		{"full any \"a b\"", "to_tsvector('english', full) @@ to_tsquery('english', $1)", []any{"'a'|'b'"}},
		{"full=\"a*\"", "error: masking op * unsupported", nil},
		{"full > x", "error: unsupported relation >", nil},
		{"a", "(to_tsvector('english', full) @@ to_tsquery('english', $1) OR Title = $2 OR Author = $3)", []any{"'a'", "a", "a"}},
		{"cql.serverChoice=a", "(to_tsvector('english', full) @@ to_tsquery('english', $1) OR Title = $2 OR Author = $3)", []any{"'a'", "a", "a"}},
		{"cql.serverChoice>a", "error: unsupported relation >", nil},
		{"cql.serverChoice==a", "(Title = $1 OR Author = $2)", []any{"a", "a"}},
		{"a and b", "(to_tsvector('english', full) @@ to_tsquery('english', $1) OR Title = $2 OR Author = $3) AND " +
			"(to_tsvector('english', full) @@ to_tsquery('english', $4) OR Title = $5 OR Author = $6)", []any{"'a'", "a", "a", "'b'", "b", "b"}},
		{"any==a", "error: unsupported relation ==", nil},
		{"any=\"\"", "(full IS NOT NULL OR Title IS NOT NULL OR Author IS NOT NULL)", []any{}},
		{"alwaysTrue=a", "TRUE", []any{}},
		{"price = 10", "price = $1", []any{10.0}},
		{"price == 10", "price = $1", []any{10.0}},
		{"price exact 10", "price = $1", []any{10.0}},
		{"price > 10.95", "price > $1", []any{10.95}},
		{"price < 10.95", "price < $1", []any{10.95}},
		{"price >= 10.95", "price >= $1", []any{10.95}},
		{"price < 10.95", "price < $1", []any{10.95}},
		{"price <= 10.95", "price <= $1", []any{10.95}},
		{"price <= beta", "error: invalid number beta", nil},
		{"price all 10.95", "error: unsupported relation all", nil},
		{"price = \"\"", "price IS NOT NULL", []any{}},
		{"date = 2026-03-05", "date = $1", []any{time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC)}},
		{"date == 2026-03-05", "date = $1", []any{time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC)}},
		{"date exact 2026-03-05", "date = $1", []any{time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC)}},
		{"date > 2026-03-05", "date > $1", []any{time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC)}},
		{"date < 2026-03-05", "date < $1", []any{time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC)}},
		{"date >= 2026-03-05", "date >= $1", []any{time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC)}},
		{"date <= 2026-03-05", "date <= $1", []any{time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC)}},
		{"date = April", "error: invalid date April, it should be in format YYYY-MM-DD", nil},
		{"date all 2026-03-05", "error: unsupported relation all", nil},
		{"date = \"\"", "date IS NOT NULL", []any{}},
		{"datetime = 2026-03-05 09:34:27", "datetime = $1", []any{time.Date(2026, 3, 5, 9, 34, 27, 0, time.UTC)}},
		{"datetime = 2026-03-05T09:34:27Z", "datetime = $1", []any{time.Date(2026, 3, 5, 9, 34, 27, 0, time.UTC)}},
		{"datetime = 2026-03-05T09:34:27+01:00", "datetime = $1", []any{dateTimeWithZone}},
		{"datetime = 2026-03-05", "datetime = $1", []any{time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC)}},
		{"datetime == 2026-03-05", "datetime = $1", []any{time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC)}},
		{"datetime exact 2026-03-05", "datetime = $1", []any{time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC)}},
		{"datetime > 2026-03-05", "datetime > $1", []any{time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC)}},
		{"datetime < 2026-03-05", "datetime < $1", []any{time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC)}},
		{"datetime >= 2026-03-05", "datetime >= $1", []any{time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC)}},
		{"datetime <= 2026-03-05", "datetime <= $1", []any{time.Date(2026, 3, 5, 0, 0, 0, 0, time.UTC)}},
		{"datetime = April", "error: invalid date time April, it should be in format YYYY-MM-DD, YYYY-MM-DD HH:MM:SS, YYYY-MM-DDTHH:MM:SSZ, YYYY-MM-DDTHH:MM:SS±HH:MM", nil},
		{"datetime all 2026-03-05", "error: unsupported relation all", nil},
		{"datetime = \"\"", "datetime IS NOT NULL", []any{}},
	} {
		var parser cql.Parser
		q, err := parser.Parse(testcase.query)
		if err != nil {
			t.Errorf("%s: CQL parse error: %v", testcase.query, err)
			continue
		}
		pgQuery, err := def.Parse(q, 1)

		expectedError := strings.HasPrefix(testcase.expected, "error: ")

		if err != nil {
			if expectedError {
				if strings.TrimPrefix(testcase.expected, "error: ") != err.Error() {
					t.Errorf("%s: Expected error %s, got %s", testcase.query, strings.TrimPrefix(testcase.expected, "error: "), err)
				}
			} else {
				t.Errorf("%s: Failed to parse: %v", testcase.query, err)
			}
			continue
		}
		if expectedError {
			t.Errorf("%s: Expected error, but got OK", testcase.query)
			continue
		}
		fullResult := pgQuery.GetWhereClause() + pgQuery.GetOrderByClause()
		if fullResult != testcase.expected {
			t.Errorf("%s: Expected %s, got %s", testcase.query, testcase.expected, fullResult)
		}
		if !reflect.DeepEqual(pgQuery.GetQueryArguments(), testcase.expectedArgs) {
			t.Errorf("%s: Expected arguments %v, got %v", testcase.query, testcase.expectedArgs, pgQuery.GetQueryArguments())
		}
	}
}
