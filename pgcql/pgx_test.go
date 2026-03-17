package pgcql

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
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
	if err != nil {
		return
	}
	var rows pgx.Rows
	fullQuery := "SELECT id FROM mytable LEFT JOIN publisher ON mytable.publisher_id = publisher.idp WHERE " +
		res.GetWhereClause() + res.GetOrderByClause()
	rows, err = conn.Query(ctx, fullQuery, res.GetQueryArguments()...)
	assert.NoErrorf(t, err, "failed to execute pgx query for cql query '%s' whereClause='%s'", query, fullQuery)

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

	_, err = conn.Exec(ctx, "CREATE TABLE publisher (idp uuid NOT NULL PRIMARY KEY, name TEXT)")
	assert.NoError(t, err, "failed to create publisher")

	_, err = conn.Exec(ctx, "CREATE TABLE mytable (id SERIAL PRIMARY KEY, title TEXT, author TEXT, tag TEXT, year INT, address JSONB, "+
		"publisher_id uuid REFERENCES publisher(idp), "+
		"start_date date, created_at timestamp)")
	assert.NoError(t, err, "failed to create mytable")

	var rows pgx.Rows

	uuid1 := uuid.New()
	rows, err = conn.Query(ctx, "INSERT INTO publisher (idp, name) VALUES ($1, $2)", uuid1, "Addision-Wesley")
	assert.NoError(t, err, "failed to insert data")
	rows.Close()

	uuid2 := uuid.New()
	rows, err = conn.Query(ctx, "INSERT INTO publisher (idp, name) VALUES ($1, $2)", uuid2, "Unknown publisher")
	assert.NoError(t, err, "failed to insert data")
	rows.Close()

	rows, err = conn.Query(ctx, "INSERT INTO mytable (title, author, tag, year, address, start_date, created_at, publisher_id) "+
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", "the art of computer programming, volume 1", "donald e. knuth", "tag1", 1968,
		`{"city": "Reading", "country": "USA", "zip": 19601}`, "2026-03-05", "2026-03-05 09:34:27", uuid1)
	assert.NoError(t, err, "failed to insert data")
	rows.Close()

	rows, err = conn.Query(ctx, "INSERT INTO mytable (title, author, tag, year, address, start_date, created_at, publisher_id) "+
		"VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", "the TeXbook", "d. e. knuth", "tag2", 1984,
		`{"city": "Stanford", "country": "USA", "zip": 67890}`, "2026-03-06", "2026-03-06 09:34:27", uuid1)
	assert.NoError(t, err, "failed to insert data")
	rows.Close()

	rows, err = conn.Query(ctx, "INSERT INTO mytable (title, year, address, publisher_id) "+
		"VALUES ($1, $2, $3, $4)", "anonymous' list", 2025,
		`{"city": "Unknown", "country": "Unknown country"}`, uuid2)
	assert.NoError(t, err, "failed to insert data")
	rows.Close()

	t.Run("exact ops", func(t *testing.T) {
		def := NewPgDefinition()

		def.AddField("title", NewFieldString().WithExact())
		def.AddField("author", NewFieldString().WithExact())
		def.AddField("year", NewFieldNumber())
		def.AddField("tag", NewFieldString().WithSplit())
		def.AddField("city", NewFieldString().WithExact().WithColumn("address->>'city'"))
		def.AddField("country", NewFieldString().WithExact().WithColumn("address->>'country'"))
		def.AddField("zip", NewFieldNumber().WithColumn("address->'zip'"))
		def.AddField("zip2", NewFieldNumber().WithColumn("(address->'zip')::numeric"))
		def.AddField("start_date", NewFieldDate().WithOnlyDate())
		def.AddField("created_at", NewFieldDate())

		var parser cql.Parser
		for _, testcase := range []struct {
			query       string
			expectedIds []int
		}{
			{"title = \"the TeXbook\"", []int{2}},
			{"title = \"the texbook\"", []int{}},
			{"title = \"the \"", []int{}},
			{"title = \"\"", []int{1, 2, 3}},
			{"title = \"\" sortby title", []int{2, 1, 3}},
			{"title = \"\" sortby title/sort.ascending", []int{3, 1, 2}},
			{"title = \"\" sortby country title", []int{2, 1, 3}},
			{"author = \"\"", []int{1, 2}},
			{"title = \"the art of computer programming, volume 1\"", []int{1}},
			{"title = \"the art of computer programming, volume\"", []int{}},
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
			{"city = \"Reading\"", []int{1}},
			{"city = \"Stanford\"", []int{2}},
			{"city = \"Unknown\"", []int{3}},
			{"country = \"USA\"", []int{1, 2}},
			{"country = \"Unknown country\"", []int{3}},
			{"zip = 19601", []int{1}},
			{"zip = 67890", []int{2}},
			{"zip >= 0", []int{1, 2}},
			{"zip = \"\"", []int{1, 2}},
			{"zip2 = 19601", []int{1}},
			{"start_date >= 2026-03-05", []int{1, 2}},
			{"start_date > 2026-03-05", []int{2}},
			{"start_date = 2026-03-05", []int{1}},
			{"created_at > 2026-03-05", []int{1, 2}},
			{"created_at > 2026-03-05 10:00:00", []int{2}},
			{"created_at = \"\"", []int{1, 2}},
		} {
			runQuery(t, parser, conn, ctx, def, testcase.query, testcase.expectedIds)
		}
	})

	t.Run("like ops", func(t *testing.T) {
		def := NewPgDefinition()

		def.AddField("title", NewFieldString().WithLikeOps())
		def.AddField("author", NewFieldString().WithLikeOps())
		def.AddField("year", NewFieldNumber())
		def.AddField("city", NewFieldString().WithLikeOps().WithColumn("address->>'city'"))

		var parser cql.Parser
		for _, testcase := range []struct {
			query       string
			expectedIds []int
		}{
			{"title = \"the TeX*\"", []int{2}},
			{"title = \"the Te?book\"", []int{2}},
			{"title = \"anonymous' l*\"", []int{3}},
			{"city = \"Read*\"", []int{1}},
			{"city = \"reading\"", []int{}},
		} {
			runQuery(t, parser, conn, ctx, def, testcase.query, testcase.expectedIds)
		}
	})

	t.Run("fulltext ops", func(t *testing.T) {
		def := NewPgDefinition()

		def.AddField("title", NewFieldString().WithFullText("simple"))
		def.AddField("author", NewFieldString().WithFullText(""))
		def.AddField("year", NewFieldNumber())
		def.AddField("city", NewFieldString().WithFullText("").WithColumn("address->>'city'"))
		def.AddField("address", NewFieldString().WithFullText("").WithColumn("address"))

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
			{"author any \"e f\"", []int{1, 2}},
			{"author adj \"e | f\"", []int{}},
			{"city = \"Reading\"", []int{1}},
			{"city = \"reading\"", []int{1}},
			{"address = USA", []int{1, 2}},
			{"address all \"Reading USA\"", []int{1}},
			{"address all \"usa reading\"", []int{1}},
			{"address all \"usa 1984\"", []int{}},
			{"address = \"unknown country\"", []int{3}},
			{"address = \"country unknown\"", []int{}},
			{"address adj \"unknown country\"", []int{3}},
			{"address any \"unknown reading\"", []int{1, 3}},
		} {
			runQuery(t, parser, conn, ctx, def, testcase.query, testcase.expectedIds)
		}
	})

	t.Run("combo ops", func(t *testing.T) {
		def := NewPgDefinition()

		titleFull := NewFieldString().WithFullText("simple").WithColumn("title")
		authorFull := NewFieldString().WithFullText("simple").WithColumn("author")
		def.AddField("cql.serverChoice", NewFieldCombo(false, []Field{titleFull, authorFull}))

		var parser cql.Parser
		for _, testcase := range []struct {
			query       string
			expectedIds []int
		}{
			{"\"the TeXbook\"", []int{2}},
			{"\"knuth\"", []int{1, 2}},
			{"\"knuth texbook\"", []int{}},
			{"knuth texbook", []int{}},
		} {
			runQuery(t, parser, conn, ctx, def, testcase.query, testcase.expectedIds)
		}
	})

	t.Run("joined ops", func(t *testing.T) {
		def := NewPgDefinition()

		titleFull := NewFieldString().WithFullText("simple").WithColumn("title")
		publisherField := NewFieldString().WithFullText("simple").WithColumn("publisher.name")
		def.AddField("publisher", publisherField)
		def.AddField("cql.serverChoice", NewFieldCombo(false, []Field{titleFull, publisherField}))
		allRefcordField := NewFieldCombo(true, []Field{})
		def.AddField("cql.allRecords", allRefcordField)

		var parser cql.Parser
		for _, testcase := range []struct {
			query       string
			expectedIds []int
		}{
			{"publisher = \"Addision-Wesley\"", []int{1, 2}},
			{"\"the TeXbook\"", []int{2}},
			{"\"Addision Wesley\"", []int{1, 2}},
			{"\"Unknown publisher\"", []int{3}},
			{"cql.allRecords=1 sortby publisher", []int{3, 1, 2}},
		} {
			runQuery(t, parser, conn, ctx, def, testcase.query, testcase.expectedIds)
		}
	})

	err = pgContainer.Terminate(ctx)
	assert.NoError(t, err, "failed to stop db container")
}
