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

func TestBuilderFromStringInjection(t *testing.T) {
	qb, err := NewQueryFromString("title = base")
	if err != nil {
		t.Fatalf("parse failed: %v", err)
	}

	query, err := qb.And().Search("author").Term("OR injected=true").Build()
	if err != nil {
		t.Fatalf("build failed: %v", err)
	}

	if got, want := query.String(), "title = base and author = \"OR injected=true\""; got != want {
		t.Fatalf("unexpected query: got %q want %q", got, want)
	}
}
