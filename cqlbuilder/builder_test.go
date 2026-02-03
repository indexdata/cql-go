package cqlbuilder

import (
	"testing"

	"github.com/indexdata/cql-go/cql"
	"github.com/stretchr/testify/assert"
)

func TestBuilderSearch(t *testing.T) {
	query, err := NewQuery().
		Search("dc.title").
		Rel(cql.EQ).
		Term("hello").
		Build()
	assert.NoError(t, err, "build failed")
	assert.Equal(t, "dc.title = hello", query.String(), "unexpected query string")
}

func TestBuilderSearchMultiWord(t *testing.T) {
	query, err := NewQuery().
		Search("dc.title").
		Rel(cql.EQ).
		Term("hello world").
		Build()
	assert.NoError(t, err, "build failed")
	assert.Equal(t, "dc.title = \"hello world\"", query.String(), "unexpected query string")
}

func TestBuilderBooleanAnd(t *testing.T) {
	query, err := NewQuery().
		Search("a").
		Term("one").
		And().
		Search("b").
		Rel(cql.GE).
		Term("2").
		Build()
	assert.NoError(t, err, "build failed")
	assert.Equal(t, "a = one and b >= 2", query.String(), "unexpected query string")
}

func TestBuilderPrefixSortAndEscaping(t *testing.T) {
	query, err := NewQuery().
		Prefix("dc", "http://purl.org/dc/elements/1.1/").
		Search("dc.title").
		Term("the \"little\" prince").
		SortBy("dc.title", cql.IgnoreCase).
		Build()
	assert.NoError(t, err, "build failed")

	want := "> dc = \"http://purl.org/dc/elements/1.1/\" dc.title = \"the \\\"little\\\" prince\" sortBy dc.title/ignoreCase"
	assert.Equal(t, want, query.String(), "unexpected query string")
}

func TestBuilderSafe(t *testing.T) {
	query, err := NewQuery().
		Search("title").
		Term("a*b?c\\^d").
		Build()
	assert.NoError(t, err, "build failed")
	assert.Equal(t, "title = a\\*b\\?c\\\\\\^d", query.String(), "unexpected query string")
}

func TestBuilderTermUnsafe(t *testing.T) {
	query, err := NewQuery().
		Search("title").
		TermUnsafe("a*b?c\\^d").
		Build()
	assert.NoError(t, err, "build failed")
	assert.Equal(t, "title = a*b?c\\^d", query.String(), "unexpected query string")
}

func TestBuilderTermUnsafeEmpty(t *testing.T) {
	query, err := NewQuery().
		Search("title").
		TermUnsafe("").
		Build()
	assert.NoError(t, err, "build failed")
	assert.Equal(t, "title = \"\"", query.String(), "unexpected query string")
}

func TestBuilderValidation(t *testing.T) {
	_, err := NewQuery().Search("a").Term("").Build()
	assert.Error(t, err, "expected error for empty term")
	assert.Equal(t, "search term must be non-empty", err.Error())

	_, err = NewQuery().Search("a").Rel("bogus").Term("x").Build()
	assert.Error(t, err, "expected error for invalid relation")
	assert.Equal(t, "invalid relation: \"bogus\"", err.Error())

	_, err = NewQuery().SortBy("").Build()
	assert.Error(t, err, "expected error for empty sort index")
	assert.Equal(t, "sort index must be non-empty", err.Error())

	_, err = NewQuery().Prefix("x", "").Build()
	assert.Error(t, err, "expected error for empty prefix")
	assert.Equal(t, "prefix uri must be non-empty", err.Error())

	_, err = NewQuery().Prefix("x", "").Prefix("a", "b").SortBy("title").SortByModifiers("asc").Build()
	assert.Error(t, err, "expected error for empty prefix")
	assert.Equal(t, "prefix uri must be non-empty", err.Error())
}

func TestBuilderDuplicateRoot(t *testing.T) {
	qb := NewQuery()
	_, err := qb.Search("a").Term("x").Build()
	assert.NoError(t, err, "unexpected error building first root clause")

	_, err = qb.Search("b").Term("y").Build()
	assert.Error(t, err, "expected error for duplicate root")
	assert.Equal(t, "query already has a root clause", err.Error())
}

func TestBuilderAppendToExistingQuery(t *testing.T) {
	base := cql.Query{
		Clause: cql.Clause{
			SearchClause: &cql.SearchClause{
				Index:    "title",
				Relation: cql.EQ,
				Term:     "base",
			},
		},
	}

	query, err := NewQueryFrom(base).
		And().
		Search("author").
		Term("alice").
		Build()
	assert.NoError(t, err, "build failed")
	assert.Equal(t, "title = base and author = alice", query.String(), "unexpected query string")
}

func TestBuilderFromString(t *testing.T) {
	qb, err := NewQueryFromString("title = base")
	assert.NoError(t, err, "parse failed")

	query, err := qb.And().Search("author").Term("alice").Build()
	assert.NoError(t, err, "build failed")
	assert.Equal(t, "title = base and author = alice", query.String(), "unexpected query string")

	_, err = NewQueryFromString("a and")
	assert.Error(t, err, "expected parse failed")
}

func TestBuilderGroupedClause(t *testing.T) {
	query, err := NewQuery().
		BeginClause().
		Search("a").
		Term("a").
		Or().
		Search("b").
		Term("b").
		EndClause().
		And().
		BeginClause().
		Search("d").
		Term("d").
		Or().
		Search("b").
		Term("b").
		EndClause().
		Build()
	assert.NoError(t, err, "build failed")
	//CQL is left associative so the stringifier skips left parentheses
	assert.Equal(t, "a = a or b = b and (d = d or b = b)", query.String(), "unexpected query string")
}

func TestBuilderFromStringInjectionSafe(t *testing.T) {
	qb, err := NewQueryFromString("title = base")
	assert.NoError(t, err, "build failed")
	query, err := qb.And().Search("author").Term("\" OR injected=true").Build()
	assert.NoError(t, err, "build failed")
	assert.Equal(t, "title = base and author = \"\\\" OR injected=true\"", query.String(), "unexpected query string")
}

func TestBuilderFromStringInjectionUnsafe(t *testing.T) {
	qb, err := NewQueryFromString("title = base")
	assert.NoError(t, err, "parse failed")

	query, err := qb.And().Search("author").TermUnsafe("\" OR injected=true").Build()
	assert.NoError(t, err, "build failed")
	assert.Equal(t, "title = base and author = \"\" OR injected=true\"", query.String(), "unexpected query string")
}

func TestBuilderErrorsAndModifiers(t *testing.T) {
	_, err := NewQuery().Build()
	assert.Error(t, err, "expected error for missing root clause")
	assert.Equal(t, "query requires a root clause", err.Error())

	_, err = NewQuery().Prefix("p", "").Build()
	assert.Error(t, err, "expected error for empty prefix uri")
	assert.Equal(t, "prefix uri must be non-empty", err.Error())

	_, err = NewQuery().SortBy("").Build()
	assert.Error(t, err, "expected error for empty sort index")
	assert.Equal(t, "sort index must be non-empty", err.Error())

	_, err = NewQuery().
		Search("a").
		Term("b").
		SortBy("title", "").
		Build()
	assert.Error(t, err, "expected error for empty sort modifier name")
	assert.Equal(t, "sort modifier name must be non-empty", err.Error())

	_, err = NewQuery().
		Search("a").
		Term("b").
		SortByModifiers("title", cql.Modifier{Name: "", Value: "x"}).
		Build()
	assert.Error(t, err, "expected error for empty modifier name")
	assert.Equal(t, "sort modifier name must be non-empty", err.Error())

	_, err = NewQuery().
		Search("a").
		Term("b").
		SortByModifiers("title", cql.Modifier{Name: "x", Relation: "bogus"}).
		Build()
	assert.Error(t, err, "expected error for invalid modifier relation")
	assert.Equal(t, "invalid modifier relation: \"bogus\"", err.Error())

	_, err = NewQuery().
		Search("a").
		Term("b").
		SortByModifiers("", cql.Modifier{Name: "x", Relation: "="}).
		Build()
	assert.Error(t, err, "expected error for empty sort index")
	assert.Equal(t, "sort index must be non-empty", err.Error())

}

func TestBuilderAppendErrors(t *testing.T) {
	_, err := NewQuery().
		And().
		Search("a").
		Term("b").
		Build()
	assert.Error(t, err, "cannot append boolean operator without existing root clause")
	assert.Equal(t, "query requires a root clause before appending", err.Error())
}

func TestBuilderBeginClauseErrors(t *testing.T) {
	qb := NewQuery()
	_, err := qb.
		BeginClause().
		Search("a").
		Term("b").
		EndClause().
		Build()
	assert.NoError(t, err, "unexpected error building first root clause")
	_, err = qb.
		BeginClause().
		Search("c").
		Term("d").
		EndClause().
		Build()
	assert.Error(t, err, "expected error when starting a second root clause")
	assert.Equal(t, "query already has a root clause", err.Error())
}

func TestBuilderJoinModifiersValidation(t *testing.T) {
	_, err := NewQuery().
		Search("a").
		Term("b").
		And().
		Mod("").
		Search("c").
		Term("d").
		Build()
	assert.Error(t, err, "expected error for empty boolean modifier name")
	assert.Equal(t, "modifier name must be non-empty", err.Error())

	_, err = NewQuery().
		Search("a").
		Term("b").
		And().
		ModRel("", "bogus", "1").
		Search("c").
		Term("d").
		Build()

	assert.Error(t, err, "expected error for invalid boolean modifier relation")
	assert.Equal(t, "modifier name must be non-empty", err.Error())

	_, err = NewQuery().
		Search("a").
		Term("b").
		And().
		ModRel(cql.Distance, "bogus", "1").
		Search("c").
		Term("d").
		Build()

	assert.Error(t, err, "expected error for invalid boolean modifier relation")
	assert.Equal(t, "invalid modifier relation: \"bogus\"", err.Error())
}

func TestBuilderRelationValidation(t *testing.T) {
	_, err := NewQuery().
		Search("a").
		Rel("bogus").
		Term("b").
		Build()
	assert.Error(t, err, "expected error for invalid relation")
	assert.Equal(t, "invalid relation: \"bogus\"", err.Error())
}

func TestBuilderEndClauseWithoutStart(t *testing.T) {
	expr := &ExprBuilder{}
	_, err := expr.EndClause().Build()
	assert.Error(t, err, "expected error for EndClause without BeginClause")
	assert.Equal(t, "no open clause to end", err.Error())
}

func TestBuilderMultiplePrefixes(t *testing.T) {
	query, err := NewQuery().
		Prefix("dc", "http://purl.org/dc/elements/1.1/").
		Prefix("bath", "http://z3950.org/bath/2.0/").
		Search("dc.title").
		Term("hello").
		Build()
	assert.NoError(t, err, "build failed")
	want := "> dc = \"http://purl.org/dc/elements/1.1/\" > bath = \"http://z3950.org/bath/2.0/\" dc.title = hello"
	assert.Equal(t, want, query.String(), "unexpected query string")
}

func TestBuilderSortByModifiersEscaping(t *testing.T) {
	query, err := NewQuery().
		Search("title").
		Term("hello").
		SortByModifiers("title", cql.Modifier{Name: "locale", Relation: cql.EQ, Value: "en\"US"}).
		Build()
	assert.NoError(t, err, "build failed")
	assert.Equal(t, "title = hello sortBy title/locale=\"en\\\"US\"", query.String(), "unexpected query string")
}

func TestBuilderSortByModifiersDefaultRelation(t *testing.T) {
	query, err := NewQuery().
		Search("title").
		Term("hello").
		SortByModifiers("title", cql.Modifier{Name: "locale", Value: "en_US"}).
		Build()
	assert.NoError(t, err, "build failed")
	assert.Equal(t, "title = hello sortBy title/locale=en_US", query.String(), "unexpected query string")
}

func TestBuilderBeginClauseRightHand(t *testing.T) {
	query, err := NewQuery().
		Search("a").
		Term("a").
		And().
		BeginClause().
		Search("b").
		Term("b").
		Or().
		Search("c").
		Term("c").
		EndClause().
		Build()

	assert.NoError(t, err, "build failed")
	assert.Equal(t, "a = a and (b = b or c = c)", query.String(), "unexpected query string")
}

func TestBuilderSearchModifiers(t *testing.T) {
	query, err := NewQuery().
		Search("title").
		Rel(cql.EQ).
		Mod(cql.Locale).
		ModRel(cql.Locale, cql.EQ, "en\"US").
		Term("hello").
		Build()

	assert.NoError(t, err, "build failed")
	assert.Equal(t, "title =/locale/locale=\"en\\\"US\" hello", query.String(), "unexpected query string")
}

func TestBuilderSearchModifiersValidation(t *testing.T) {
	_, err := NewQuery().
		Search("title").
		Mod("").
		Term("hello").
		Build()
	assert.Error(t, err, "expected error for empty search modifier name")
	assert.Equal(t, "modifier name must be non-empty", err.Error())

	_, err = NewQuery().
		Search("title").
		ModRel(cql.Locale, "bogus", "en").
		Term("hello").
		Build()
	assert.Error(t, err, "expected error for invalid search modifier relation")
	assert.Equal(t, "invalid modifier relation: \"bogus\"", err.Error())
}

func TestBuilderSearchRelationDefaults(t *testing.T) {
	query, err := NewQuery().
		Search("title").
		Rel("").
		Term("hello").
		Build()
	assert.NoError(t, err, "build failed")
	assert.Equal(t, "title = hello", query.String(), "unexpected query string")
}

func TestBuilderAppendInvalidOperator(t *testing.T) {
	qb := NewQuery()
	expr := qb.Search("a").Term("b")
	jb := &JoinBuilder{
		finish: expr.finish,
		build:  expr.build,
		qb:     expr.qb,
		left:   expr.clause,
		op:     "bogus",
	}
	_, err := jb.Search("c").Term("d").Build()
	assert.Error(t, err, "expected error for invalid boolean operator")
}

func TestBuilderEscapeHelpers(t *testing.T) {
	assert.Equal(t, "a\\\\b\\\"c", EscapeSpecialChars("a\\b\"c"), "unexpected EscapeSpecialChars")
	assert.Equal(t, "\\*\\?\\^", EscapeMaskingChars("*?^"), "unexpected EscapeMaskingChars")
}
