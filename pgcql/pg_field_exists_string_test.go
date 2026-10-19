package pgcql

import (
	"testing"

	"github.com/indexdata/cql-go/cql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFieldExistsStringGenerate(t *testing.T) {
	for _, tc := range []struct {
		name, term, relation, predicate string
		args                            []any
		negative                        bool
	}{
		{"equality", "AbC", "=", "v.value = $7", []any{"AbC"}, false},
		{"exact", "AbC", "==", "v.value = $7", []any{"AbC"}, false},
		{"exact relation", "AbC", "exact", "v.value = $7", []any{"AbC"}, false},
		{"wildcards", "Ab*?", "=", "v.value LIKE $7", []any{"Ab%_"}, false},
		{"inequality", "AbC", "<>", "v.value = $7", []any{"AbC"}, true},
		{"wildcard inequality", "Ab*", "<>", "v.value LIKE $7", []any{"Ab%"}, true},
		{"escaped wildcard", `Ab\*`, "=", "v.value = $7", []any{"Ab*"}, false},
		{"SQL wildcard literals", "A%_*", "=", "v.value LIKE $7", []any{`A\%\_%`}, false},
		{"quote parameter", "O'Brien", "=", "v.value = $7", []any{"O'Brien"}, false},
		{"empty equality", "", "=", "v.value IS NOT NULL AND v.value <> ''", nil, true},
		{"empty exact", "", "==", "v.value IS NOT NULL AND v.value <> ''", nil, true},
		{"empty exact relation", "", "exact", "v.value IS NOT NULL AND v.value <> ''", nil, true},
		{"empty inequality", "", "<>", "v.value IS NOT NULL AND v.value <> ''", nil, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := NewFieldExistsString("child", "v", "v.parent_id = p.id", "v.value")
			sql, args, err := f.Generate(cql.SearchClause{Term: tc.term, Relation: cql.Relation(tc.relation)}, 7)
			require.NoError(t, err)
			want := "EXISTS (SELECT 1 FROM child v WHERE v.parent_id = p.id AND " + tc.predicate + ")"
			if tc.negative {
				want = "NOT " + want
			}
			assert.Equal(t, want, sql)
			assert.Equal(t, tc.args, args)
		})
	}
}

func TestFieldExistsStringConfiguration(t *testing.T) {
	f := NewFieldExistsString("child", "", "child.parent_id = p.id", "child.value")
	assert.Empty(t, f.GetColumn())
	assert.Empty(t, f.Sort())
	f.SetColumn("related_value")
	assert.Equal(t, "related_value", f.GetColumn())
	assert.Same(t, f, f.WithSortExpression("p.name"))
	assert.Equal(t, "p.name", f.Sort())
	assert.Same(t, f, f.WithField(NewFieldString().WithLikeOps().WithLower()))
	for _, relation := range []cql.Relation{cql.EQ, cql.NE} {
		sql, args, err := f.Generate(cql.SearchClause{Term: "Ab*", Relation: relation}, 4)
		require.NoError(t, err)
		want := "EXISTS (SELECT 1 FROM child WHERE child.parent_id = p.id AND lower(child.value) LIKE lower($4))"
		if relation == cql.NE {
			want = "NOT " + want
		}
		assert.Equal(t, want, sql)
		assert.Equal(t, []any{"Ab%"}, args)
	}
}

func TestFieldExistsStringErrors(t *testing.T) {
	f := NewFieldExistsString("child", "v", "v.parent_id = p.id", "v.value")
	for _, sc := range []cql.SearchClause{
		{Term: "abc", Relation: cql.GT},
		{Term: "abc", Relation: cql.ALL},
		{Term: `abc\`, Relation: cql.EQ},
		{Term: `abc\`, Relation: cql.NE},
	} {
		sql, args, err := f.Generate(sc, 1)
		require.Error(t, err)
		assert.Empty(t, sql)
		assert.Nil(t, args)
	}
}

func TestFieldExistsStringDefinition(t *testing.T) {
	def := NewPgDefinition()
	f := NewFieldExistsString("child", "v", "v.parent_id = p.id", "v.value").WithField(NewFieldString().WithLikeOps().WithLower())
	def.AddField("value", f)
	def.AddField("name", NewFieldString().WithExact().WithColumn("p.name"))
	assert.Equal(t, "value", f.GetColumn())
	var parser cql.Parser
	q, err := parser.Parse(`value = "Ab*" and name = "Smith"`)
	require.NoError(t, err)
	result, err := def.Parse(q, 3)
	require.NoError(t, err)
	assert.Equal(t, "EXISTS (SELECT 1 FROM child v WHERE v.parent_id = p.id AND lower(v.value) LIKE lower($3)) AND p.name = $4", result.GetWhereClause())
	assert.Equal(t, []any{"Ab%", "Smith"}, result.GetQueryArguments())
	q, err = parser.Parse(`value = "abc" sortby value`)
	require.NoError(t, err)
	_, err = def.Parse(q, 1)
	require.ErrorContains(t, err, "does not support sorting")
	f.WithSortExpression("p.name")
	result, err = def.Parse(q, 1)
	require.NoError(t, err)
	assert.Contains(t, result.GetOrderByClause(), "p.name")
}
