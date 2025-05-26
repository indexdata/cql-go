package pgcql

import (
	"reflect"
	"strings"
	"testing"

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

func TestParsing(t *testing.T) {
	def := NewPgDefinition()
	title := &FieldString{}
	title.WithExact().SetColumn("Title")

	assert.Equal(t, title.GetColumn(), "Title", "GetColumn() should return the column name")

	author := &FieldString{}
	author.WithLikeOps().SetColumn("Author")

	serverChoice := &FieldString{}
	serverChoice.WithExact().SetColumn("T")

	full := &FieldString{}
	full.WithFullText("english")

	def.AddField("title", title).AddField("author", author).AddField("cql.serverChoice", serverChoice).AddField("full", full)

	price := &FieldNumber{}
	def.AddField("price", price)

	for _, testcase := range []struct {
		query        string
		expected     string
		expectedArgs []any
	}{
		{"abc", "T = $1", []any{"abc"}},
		{"\"\"", "T IS NOT NULL", []any{}},
		{"au=2", "error: unknown field au", nil},
		{"title>2", "error: unsupported relation >", nil},
		{"title=2", "Title = $1", []any{"2"}},
		{"title==2", "Title = $1", []any{"2"}},
		{"title exact 2", "Title = $1", []any{"2"}},
		{"title<>2", "Title <> $1", []any{"2"}},
		{"a or b and c", "(T = $1 OR T = $2) AND T = $3", []any{"a", "b", "c"}},
		{"title = abc", "Title = $1", []any{"abc"}},
		{"author = \"test\"", "Author = $1", []any{"test"}},
		{"author <> \"test\"", "Author <> $1", []any{"test"}},
		{"author = \"test*\"", "Author LIKE $1", []any{"test%"}},
		{"author <> \"test*\"", "Author NOT LIKE $1", []any{"test%"}},
		{"title = a AND author = b c", "Title = $1 AND Author = $2", []any{"a", "b c"}},
		{"title = 'a' OR author = 'b'", "Title = $1 OR Author = $2", []any{"'a'", "'b'"}},
		{"title = a NOT author = b", "Title = $1 AND NOT Author = $2", []any{"a", "b"}},
		{"a prox b", "error: unsupported operator prox", []any{}},
		{"a sortby title", "error: sorting not supported", []any{}},
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
		{"full = \"abc\"", "to_tsvector('english', full) @@ phraseto_tsquery('english', $1)", []any{"abc"}},
		{"full adj \"abc\"", "to_tsvector('english', full) @@ phraseto_tsquery('english', $1)", []any{"abc"}},
		{"full all \"abc\"", "to_tsvector('english', full) @@ plainto_tsquery('english', $1)", []any{"abc"}},
		{"full=\"a*\"", "error: masking op * unsupported", nil},
		{"full any x", "error: exact search not supported", nil},
		{"price = 10", "price = $1", []any{10.0}},
		{"price == 10", "price = $1", []any{10.0}},
		{"price exact 10", "price = $1", []any{10.0}},
		{"price > 10.95", "price > $1", []any{10.95}},
		{"price < 10.95", "price < $1", []any{10.95}},
		{"price >= 10.95", "price >= $1", []any{10.95}},
		{"price < 10.95", "price < $1", []any{10.95}},
		{"price <= 10.95", "price <= $1", []any{10.95}},
		{"price <= beta", "error: invalid number beta", nil},
		{"price all 10.95", "error: unsupported operator all", nil},
		{"price = \"\"", "price IS NOT NULL", []any{}},
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
		if pgQuery.GetWhereClause() != testcase.expected {
			t.Errorf("%s: Expected %s, got %s", testcase.query, testcase.expected, pgQuery.GetWhereClause())
		}
		if !reflect.DeepEqual(pgQuery.GetQueryArguments(), testcase.expectedArgs) {
			t.Errorf("%s: Expected arguments %v, got %v", testcase.query, testcase.expectedArgs, pgQuery.GetQueryArguments())
		}
		if pgQuery.GetOrderByClause() != "" {
			t.Errorf("%s: Expected empty order by clause, got %s", testcase.query, pgQuery.GetOrderByClause())
		}
		if pgQuery.GetOrderByFields() != "" {
			t.Errorf("%s: Expected empty order by fields, got %s", testcase.query, pgQuery.GetOrderByFields())
		}
	}
}
