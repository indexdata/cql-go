package example

import (
	"testing"

	"github.com/indexdata/cql-go/cql"
	"github.com/stretchr/testify/assert"
)

func TestExampleParse(t *testing.T) {
	var parser cql.Parser

	in := "dc.title = abc"
	query, err := parser.Parse(in)
	assert.Nil(t, err)
	assert.Equal(t, in, query.String())
}

func TestExampleSearchClause(t *testing.T) {
	searchClause := &cql.SearchClause{Index: "dc.title", Relation: "=", Term: "abc"}
	query := cql.Query{Clause: cql.Clause{SearchClause: searchClause}}
	assert.Equal(t, "dc.title = abc", query.String())
}

func TestExampleBoolean(t *testing.T) {
	sc1 := &cql.SearchClause{Index: "dc.title", Relation: "=", Term: "abc"}
	sc2 := &cql.SearchClause{Term: "other"}
	bc := &cql.BoolClause{Operator: cql.AND, Left: cql.Clause{SearchClause: sc1}, Right: cql.Clause{SearchClause: sc2}}
	query := cql.Query{Clause: cql.Clause{BoolClause: bc}}
	assert.Equal(t, "dc.title = abc and other", query.String())
}

func TestExampleXcql(t *testing.T) {
	searchClause := &cql.SearchClause{Index: "dc.title", Relation: "=", Term: "abc"}
	query := cql.Query{Clause: cql.Clause{SearchClause: searchClause}}
	var xcql cql.Xcql
	assert.Equal(t, `<xcql xmlns="http://docs.oasis-open.org/ns/search-ws/xcql">
<triple>
<searchClause>
<index>dc.title</index>
<relation>
<value>=</value>
</relation>
<term>abc</term>
</searchClause>
</triple>
</xcql>
`, xcql.Marshal(query, 0))
}
