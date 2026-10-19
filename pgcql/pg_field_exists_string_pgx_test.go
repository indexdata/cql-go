package pgcql

import (
	"context"
	"testing"
	"time"

	"github.com/indexdata/cql-go/cql"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestFieldExistsStringPgx(t *testing.T) {
	ctx := context.Background()
	container, err := postgres.Run(ctx, "postgres:16",
		postgres.WithDatabase("cql"), postgres.WithUsername("cql"), postgres.WithPassword("cql"),
		testcontainers.WithWaitStrategy(wait.ForLog("database system is ready to accept connections").WithOccurrence(2).WithStartupTimeout(30*time.Second)))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, container.Terminate(ctx)) })
	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)
	conn, err := pgx.Connect(ctx, dsn)
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, conn.Close(ctx)) })
	_, err = conn.Exec(ctx, `CREATE TABLE parent (id integer PRIMARY KEY);
 CREATE TABLE child (parent_id integer REFERENCES parent(id), value text);
 INSERT INTO parent VALUES (1),(2),(3),(4),(5),(6);
 INSERT INTO child VALUES (1,'AbC'),(1,'Other'),(1,'AbC'),(2,NULL),(3,''),(4,'other'),(6,NULL),(6,'');`)
	require.NoError(t, err)
	// Parent 5 has no children; parent 1 has both matching and nonmatching children.
	def := NewPgDefinition()
	def.AddField("value", NewFieldExistsString("child", "v", "v.parent_id = p.id", "v.value").WithField(NewFieldString().WithLikeOps().WithLower()))
	def.AddField("sensitive", NewFieldExistsString("child", "v", "v.parent_id = p.id", "v.value"))
	for _, tc := range []struct {
		query string
		ids   []int
	}{
		{`value = "ABC"`, []int{1}},
		{`value = "aB*"`, []int{1}},
		{`value <> "ABC"`, []int{2, 3, 4, 5, 6}},
		{`value <> "Ab*"`, []int{2, 3, 4, 5, 6}},
		{`value = ""`, []int{2, 3, 5, 6}},
		{`value == ""`, []int{2, 3, 5, 6}},
		{`value <> ""`, []int{1, 4}},
		{`value = "ABC" and value = "OTHER"`, []int{1}},
		{`value = "ABC" or value = "OTHER"`, []int{1, 4}},
		{`value = "missing"`, []int{}},
		{`sensitive = "abc"`, []int{}},
		{`sensitive = "AbC"`, []int{1}},
	} {
		t.Run(tc.query, func(t *testing.T) {
			var parser cql.Parser
			q, err := parser.Parse(tc.query)
			require.NoError(t, err)
			result, err := def.Parse(q, 1)
			require.NoError(t, err)
			rows, err := conn.Query(ctx, "SELECT p.id FROM parent p WHERE "+result.GetWhereClause()+" ORDER BY p.id", result.GetQueryArguments()...)
			require.NoError(t, err)
			ids, err := pgx.CollectRows(rows, pgx.RowTo[int])
			require.NoError(t, err)
			assert.Equal(t, tc.ids, ids)
		})
	}
}
