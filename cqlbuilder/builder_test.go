package cqlbuilder

import (
	"testing"

	"github.com/indexdata/cql-go/cql"
)

func TestBuilderSearch(t *testing.T) {
	query, err := NewQuery().
		Search("dc.title").
		Rel(cql.EQ).
		Term("hello").
		Build()
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	if got, want := query.String(), "dc.title = hello"; got != want {
		t.Fatalf("unexpected query: got %q want %q", got, want)
	}
}

func TestBuilderSearchMultiWord(t *testing.T) {
	query, err := NewQuery().
		Search("dc.title").
		Rel(cql.EQ).
		Term("hello world").
		Build()
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	if got, want := query.String(), "dc.title = \"hello world\""; got != want {
		t.Fatalf("unexpected query: got %q want %q", got, want)
	}
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
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	if got, want := query.String(), "a = one and b >= 2"; got != want {
		t.Fatalf("unexpected query: got %q want %q", got, want)
	}
}

func TestBuilderPrefixSortAndEscaping(t *testing.T) {
	query, err := NewQuery().
		Prefix("dc", "http://purl.org/dc/elements/1.1/").
		Search("dc.title").
		Term("the \"little\" prince").
		SortBy("dc.title", cql.IgnoreCase).
		Build()
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	want := "> dc = \"http://purl.org/dc/elements/1.1/\" dc.title = \"the \\\"little\\\" prince\" sortBy dc.title/ignoreCase"
	if got := query.String(); got != want {
		t.Fatalf("unexpected query: got %q want %q", got, want)
	}
}

func TestBuilderSafe(t *testing.T) {
	query, err := NewQuery().
		Search("title").
		Term("a*b?c\\^d").
		Build()
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	if got, want := query.String(), "title = a\\*b\\?c\\\\\\^d"; got != want {
		t.Fatalf("unexpected query: got %q want %q", got, want)
	}

}

func TestBuilderTermUnsafe(t *testing.T) {
	query, err := NewQuery().
		Search("title").
		TermUnsafe("a*b?c\\^d").
		Build()
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	if got, want := query.String(), "title = a*b?c\\^d"; got != want {
		t.Fatalf("unexpected query: got %q want %q", got, want)
	}
}

func TestBuilderTermUnsafeEmpty(t *testing.T) {
	query, err := NewQuery().
		Search("title").
		TermUnsafe("").
		Build()
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	if got, want := query.String(), "title = \"\""; got != want {
		t.Fatalf("unexpected query: got %q want %q", got, want)
	}
}

func TestBuilderValidation(t *testing.T) {
	if _, err := NewQuery().Search("a").Term("").Build(); err == nil {
		t.Fatalf("expected error for empty term")
	}

	if _, err := NewQuery().Search("a").Rel("bogus").Term("x").Build(); err == nil {
		t.Fatalf("expected error for invalid relation")
	}

	qb := NewQuery()
	_, _ = qb.Search("a").Term("x").Build()
	if _, err := qb.Search("b").Term("y").Build(); err == nil {
		t.Fatalf("expected error for duplicate root")
	}
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
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	if got, want := query.String(), "title = base and author = alice"; got != want {
		t.Fatalf("unexpected query: got %q want %q", got, want)
	}
}

func TestBuilderFromString(t *testing.T) {
	qb, err := NewQueryFromString("title = base")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	query, err := qb.And().Search("author").Term("alice").Build()
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	if got, want := query.String(), "title = base and author = alice"; got != want {
		t.Fatalf("unexpected query: got %q want %q", got, want)
	}
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
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}
	//CQL is left associative so the stringifier skips left parentheses
	if got, want := query.String(), "a = a or b = b and (d = d or b = b)"; got != want {
		t.Fatalf("unexpected query: got %q want %q", got, want)
	}
}

func TestBuilderFromStringInjectionSafe(t *testing.T) {
	qb, err := NewQueryFromString("title = base")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	query, err := qb.And().Search("author").Term("\" OR injected=true").Build()
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	if got, want := query.String(), "title = base and author = \"\\\" OR injected=true\""; got != want {
		t.Fatalf("unexpected query: got %q want %q", got, want)
	}
}

func TestBuilderFromStringInjectionUnsafe(t *testing.T) {
	qb, err := NewQueryFromString("title = base")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	query, err := qb.And().Search("author").TermUnsafe("\" OR injected=true").Build()
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	if got, want := query.String(), "title = base and author = \"\" OR injected=true\""; got != want {
		t.Fatalf("unexpected query: got %q want %q", got, want)
	}
}

func TestBuilderErrorsAndModifiers(t *testing.T) {
	if _, err := NewQuery().Build(); err == nil {
		t.Fatalf("expected error for missing root clause")
	}

	if _, err := NewQuery().Prefix("p", "").Build(); err == nil {
		t.Fatalf("expected error for empty prefix uri")
	}

	if _, err := NewQuery().SortBy("").Build(); err == nil {
		t.Fatalf("expected error for empty sort index")
	}

	if _, err := NewQuery().
		Search("a").
		Term("b").
		SortBy("title", "").
		Build(); err == nil {
		t.Fatalf("expected error for empty sort modifier name")
	}

	if _, err := NewQuery().
		Search("a").
		Term("b").
		SortByModifiers("title", cql.Modifier{Name: "", Value: "x"}).
		Build(); err == nil {
		t.Fatalf("expected error for empty modifier name")
	}

	if _, err := NewQuery().
		Search("a").
		Term("b").
		SortByModifiers("title", cql.Modifier{Name: "x", Relation: "bogus"}).
		Build(); err == nil {
		t.Fatalf("expected error for invalid modifier relation")
	}
}

func TestBuilderAppendErrors(t *testing.T) {
	if _, err := NewQuery().
		And().
		Search("a").
		Term("b").
		Build(); err == nil {
		t.Fatalf("expected error when appending without root")
	}
}

func TestBuilderBeginClauseErrors(t *testing.T) {
	qb := NewQuery()
	_, err := qb.
		BeginClause().
		Search("a").
		Term("b").
		EndClause().
		Build()
	if err != nil {
		t.Fatalf("unexpected error building first root clause: %v", err)
	}

	_, err = qb.
		BeginClause().
		Search("c").
		Term("d").
		EndClause().
		Build()
	if err == nil {
		t.Fatalf("expected error when starting a second root clause")
	}
}

func TestBuilderJoinModifiersValidation(t *testing.T) {
	if _, err := NewQuery().
		Search("a").
		Term("b").
		And().
		Mod("").
		Search("c").
		Term("d").
		Build(); err == nil {
		t.Fatalf("expected error for empty boolean modifier name")
	}

	if _, err := NewQuery().
		Search("a").
		Term("b").
		And().
		ModRel(cql.Distance, "bogus", "1").
		Search("c").
		Term("d").
		Build(); err == nil {
		t.Fatalf("expected error for invalid boolean modifier relation")
	}
}

func TestBuilderRelationValidation(t *testing.T) {
	if _, err := NewQuery().
		Search("a").
		Rel("bogus").
		Term("b").
		Build(); err == nil {
		t.Fatalf("expected error for invalid relation")
	}
}

func TestBuilderEndClauseWithoutStart(t *testing.T) {
	expr := &ExprBuilder{}
	if _, err := expr.EndClause().Build(); err == nil {
		t.Fatalf("expected error for EndClause without BeginClause")
	}
}

func TestBuilderMultiplePrefixes(t *testing.T) {
	query, err := NewQuery().
		Prefix("dc", "http://purl.org/dc/elements/1.1/").
		Prefix("bath", "http://z3950.org/bath/2.0/").
		Search("dc.title").
		Term("hello").
		Build()
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	want := "> dc = \"http://purl.org/dc/elements/1.1/\" > bath = \"http://z3950.org/bath/2.0/\" dc.title = hello"
	if got := query.String(); got != want {
		t.Fatalf("unexpected query: got %q want %q", got, want)
	}
}

func TestBuilderSortByModifiersEscaping(t *testing.T) {
	query, err := NewQuery().
		Search("title").
		Term("hello").
		SortByModifiers("title", cql.Modifier{Name: "locale", Relation: cql.EQ, Value: "en\"US"}).
		Build()
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	if got, want := query.String(), "title = hello sortBy title/locale=\"en\\\"US\""; got != want {
		t.Fatalf("unexpected query: got %q want %q", got, want)
	}
}

func TestBuilderSortByModifiersDefaultRelation(t *testing.T) {
	query, err := NewQuery().
		Search("title").
		Term("hello").
		SortByModifiers("title", cql.Modifier{Name: "locale", Value: "en_US"}).
		Build()
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	if got, want := query.String(), "title = hello sortBy title/locale=en_US"; got != want {
		t.Fatalf("unexpected query: got %q want %q", got, want)
	}
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
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	if got, want := query.String(), "a = a and (b = b or c = c)"; got != want {
		t.Fatalf("unexpected query: got %q want %q", got, want)
	}
}

func TestBuilderSearchModifiers(t *testing.T) {
	query, err := NewQuery().
		Search("title").
		Rel(cql.EQ).
		Mod(cql.Locale).
		ModRel(cql.Locale, cql.EQ, "en\"US").
		Term("hello").
		Build()
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	if got, want := query.String(), "title =/locale/locale=\"en\\\"US\" hello"; got != want {
		t.Fatalf("unexpected query: got %q want %q", got, want)
	}
}

func TestBuilderSearchModifiersValidation(t *testing.T) {
	if _, err := NewQuery().
		Search("title").
		Mod("").
		Term("hello").
		Build(); err == nil {
		t.Fatalf("expected error for empty search modifier name")
	}

	if _, err := NewQuery().
		Search("title").
		ModRel(cql.Locale, "bogus", "en").
		Term("hello").
		Build(); err == nil {
		t.Fatalf("expected error for invalid search modifier relation")
	}
}

func TestBuilderSearchRelationDefaults(t *testing.T) {
	query, err := NewQuery().
		Search("title").
		Rel("").
		Term("hello").
		Build()
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	if got, want := query.String(), "title = hello"; got != want {
		t.Fatalf("unexpected query: got %q want %q", got, want)
	}
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
	if _, err := jb.Search("c").Term("d").Build(); err == nil {
		t.Fatalf("expected error for invalid boolean operator")
	}
}

func TestBuilderEscapeHelpers(t *testing.T) {
	if got, want := EscapeSpecialChars("a\\b\"c"), "a\\\\b\\\"c"; got != want {
		t.Fatalf("unexpected EscapeSpecialChars: got %q want %q", got, want)
	}

	if got, want := EscapeMaskingChars("*?^"), "\\*\\?\\^"; got != want {
		t.Fatalf("unexpected EscapeMaskingChars: got %q want %q", got, want)
	}
}
