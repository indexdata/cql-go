package pgcql

import (
	"context"
	"testing"
	"time"

	"github.com/indexdata/cql-go/cql"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

func runQuery(t *testing.T, parser cql.Parser, conn *pgx.Conn, ctx context.Context, def Definition, query string, expectedIds []int) {
	q, err := parser.Parse(query)
	assert.NoErrorf(t, err, "failed to parse cql query '%s'", query)
	res, err := def.Parse(q, 1)
	assert.NoErrorf(t, err, "failed to parse pgcql query for cql query '%s'", query)
	var rows pgx.Rows
	rows, err = conn.Query(ctx, "SELECT id FROM mytable WHERE "+res.GetWhereClause(), res.GetQueryArguments()...)
	assert.NoErrorf(t, err, "failed to execute pgx query for cql query '%s' whereClause='%s'", query, res.GetWhereClause())
	ids := make([]int, 0)
	for rows.Next() {
		var id int
		err := rows.Scan(&id)
		assert.NoErrorf(t, err, "failed to scan for cql query '%s'", query)
		ids = append(ids, id)
	}
	assert.Equal(t, expectedIds, ids, "expected ids %v, got %v for query '%s'", expectedIds, ids, query)
	rows.Close()
}

func TestPgx(t *testing.T) {
	ctx := context.Background()
	pgContainer, err := postgres.Run(ctx, "postgres",
		postgres.WithDatabase("crosslink"),
		postgres.WithUsername("crosslink"),
		postgres.WithPassword("crosslink"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).WithStartupTimeout(5*time.Second)),
	)
	assert.NoError(t, err, "failed to start db container")
	var connStr string
	connStr, err = pgContainer.ConnectionString(ctx, "sslmode=disable")
	assert.NoError(t, err, "failed to get db connection string")
	assert.NotEmpty(t, connStr, "connection string should not be empty")

	conn, err := pgx.Connect(ctx, connStr)
	assert.NoError(t, err, "failed to connect to db")
	defer func() {
		err := conn.Close(ctx)
		assert.NoError(t, err, "failed to close db connection")
	}()
	_, err = conn.Exec(ctx, "CREATE TABLE mytable (id SERIAL PRIMARY KEY, title TEXT, author TEXT, tag TEXT, year INT)")
	assert.NoError(t, err, "failed to create mytable")

	var rows pgx.Rows

	rows, err = conn.Query(ctx, "INSERT INTO mytable (title, author, tag, year) VALUES ($1, $2, $3, $4)", "the art of computer programming, volume 1", "donald e. knuth", "tag1", 1968)
	assert.NoError(t, err, "failed to insert data")
	rows.Close()

	rows, err = conn.Query(ctx, "INSERT INTO mytable (title, author, tag, year) VALUES ($1, $2, $3, $4)", "the TeXbook", "d. e. knuth", "tag2", 1984)
	assert.NoError(t, err, "failed to insert data")
	rows.Close()

	rows, err = conn.Query(ctx, "INSERT INTO mytable (title, year) VALUES ($1, $2)", "anonymous' list", 2025)
	assert.NoError(t, err, "failed to insert data")
	rows.Close()

	t.Run("exact ops", func(t *testing.T) {
		def := NewPgDefinition()

		def.AddField("title", (&FieldString{}).WithExact())
		def.AddField("author", (&FieldString{}).WithExact())
		def.AddField("year", (&FieldNumber{}))
		def.AddField("tag", (&FieldString{}).WithSplit())

		var parser cql.Parser
		for _, testcase := range []struct {
			query       string
			expectedIds []int
		}{
			{"title = \"the TeXbook\"", []int{2}},
			{"title = \"the texbook\"", []int{}},
			{"title = \"the \"", []int{}},
			{"title = \"\"", []int{1, 2, 3}},
			{"author = \"\"", []int{1, 2}},
			{"title = \"the art of computer programming, volume 1\"", []int{1}},
			{"author = \"d. e. knuth\"", []int{2}},
			{"author = \"donald e. knuth\"", []int{1}},
			{"title = \"the TeXbook\" AND author = \"d. e. knuth\"", []int{2}},
			{"title = \"the TeXbook\" AND author = \"donald e. knuth\"", []int{}},
			{"title = \"the TeXbook\" OR author = \"d. e. knuth\"", []int{2}},
			{"title = \"the TeXbook\" OR author = \"donald e. knuth\"", []int{1, 2}},
			{"title = \"the TeXbook\" AND author = \"d. e. knuth\" AND year = 1984", []int{2}},
			{"title = \"the TeXbook\" AND author = \"d. e. knuth\" AND year = 1968", []int{}},
			{"title = \"the TeXbook\" AND author = \"d. e. knuth\" AND year = 1984 AND title = \"the art of computer programming, volume 1\"", []int{}},
			{"title = \"the TeXbook\" AND author = \"d. e. knuth\" AND year = 1984 AND title = \"the TeXbook\"", []int{2}},
			{"title = \"anonymous' list\"", []int{3}},
			{"title = \"anonymous'' list\"", []int{}},
			{"title = \"anonymous list\"", []int{}},
			{"year <> 1984", []int{1, 3}},
			{"year < 1984", []int{1}},
			{"year <= 1984", []int{1, 2}},
			{"year >= 1984", []int{2, 3}},
			{"year > 1984", []int{3}},
			{"tag any \"tag1\"", []int{1}},
			{"tag <> \"tag1\"", []int{2}},
			{"tag any \"tag1 tag2 tag3\"", []int{1, 2}},
		} {
			runQuery(t, parser, conn, ctx, def, testcase.query, testcase.expectedIds)
		}
	})

	t.Run("like ops", func(t *testing.T) {
		def := NewPgDefinition()

		def.AddField("title", (&FieldString{}).WithLikeOps())
		def.AddField("author", (&FieldString{}).WithLikeOps())
		def.AddField("year", (&FieldNumber{}))

		var parser cql.Parser
		for _, testcase := range []struct {
			query       string
			expectedIds []int
		}{
			{"title = \"the TeX*\"", []int{2}},
			{"title = \"the Te?book\"", []int{2}},
			{"title = \"anonymous' l*\"", []int{3}},
		} {
			runQuery(t, parser, conn, ctx, def, testcase.query, testcase.expectedIds)
		}
	})

	t.Run("fulltext ops", func(t *testing.T) {
		def := NewPgDefinition()

		def.AddField("title", (&FieldString{}).WithFullText("simple"))
		def.AddField("author", (&FieldString{}).WithFullText(""))
		def.AddField("year", (&FieldNumber{}))

		var parser cql.Parser
		for _, testcase := range []struct {
			query       string
			expectedIds []int
		}{
			{"title = \"the TeXbook\"", []int{2}},
			{"title = \"the Texbook\"", []int{2}},
			{"title = \"Texbook\"", []int{2}},
			{"title = \"Texboo\"", []int{}},
			{"author all \"knuth d e\"", []int{2}},
			{"author = \"d e knuth\"", []int{2}},
			{"author adj \"d e knuth\"", []int{2}},
			{"author adj \"e knuth\"", []int{1, 2}},
			{"author adj \"e d knuth\"", []int{}},
		} {
			runQuery(t, parser, conn, ctx, def, testcase.query, testcase.expectedIds)
		}
	})

	err = pgContainer.Terminate(ctx)
	assert.NoError(t, err, "failed to stop db container")
}
