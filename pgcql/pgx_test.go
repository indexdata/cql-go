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
	_, err = conn.Exec(ctx, "CREATE TABLE mytable (id SERIAL PRIMARY KEY, title TEXT, author TEXT, year INT)")
	assert.NoError(t, err, "failed to create mytable")

	var rows pgx.Rows

	rows, err = conn.Query(ctx, "INSERT INTO mytable (title, author, year) VALUES ($1, $2, $3)", "the art of computer programming, volume 1", "donald e. knuth", 1968)
	assert.NoError(t, err, "failed to insert data")
	rows.Close()

	rows, err = conn.Query(ctx, "INSERT INTO mytable (title, author, year) VALUES ($1, $2, $3)", "the TeXbook", "d. e. knuth", 1984)
	assert.NoError(t, err, "failed to insert data")
	rows.Close()

	rows, err = conn.Query(ctx, "INSERT INTO mytable (title, year) VALUES ($1, $2)", "anonymous", 2025)
	assert.NoError(t, err, "failed to insert data")
	rows.Close()

	t.Run("exact", func(t *testing.T) {
		def := &PgDefinition{}

		def.AddField("title", (&FieldString{}).WithExact())
		def.AddField("author", (&FieldString{}).WithExact())
		def.AddField("year", (&FieldString{}).WithExact())

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
		} {
			q, err := parser.Parse(testcase.query)
			assert.NoErrorf(t, err, "failed to parse cql query '%s'", testcase.query)
			res, err := def.Parse(q, 1)
			assert.NoErrorf(t, err, "failed to parse pgcql query for cql query '%s'", testcase.query)
			rows, err = conn.Query(ctx, "SELECT id FROM mytable WHERE "+res.GetWhereClause(), res.GetQueryArguments()...)
			assert.NoErrorf(t, err, "failed to execute query for cql query '%s'", testcase.query)
			ids := make([]int, 0)
			for rows.Next() {
				var id int
				err := rows.Scan(&id)
				assert.NoErrorf(t, err, "failed to scan for cql query '%s'", testcase.query)
				ids = append(ids, id)
			}
			assert.Equal(t, testcase.expectedIds, ids, "expected ids %v, got %v for query '%s'", testcase.expectedIds, ids, testcase.query)
			rows.Close()
		}
	})
	err = pgContainer.Terminate(ctx)
	assert.NoError(t, err, "failed to stop db container")
}
