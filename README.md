# Introduction

This module implements a parser for the `Contextual Query Language` (CQL) which is a part of the
`Search/Retrieval via URL` (SRU) and `Search/Retrieval via Web Service` (SRW) family of protocols.
See the [Library of Congress spec](https://www.loc.gov/standards/sru/cql/) for details.

## Usage

Running `make` or `go build -o . ./...` will compile the library and create the command-line tool `cql-cli`.

The primary purpose of the library is to parse a CQL string and produce `cql.Query` which is a parse-tree really.


```go
import (
   "fmt"
   "github.com/indexdata/cql-go/cql"
)

func main() {
   var parser cql.Parser
   query, err := parser.Parse("title=hello")
   if err != nil {
      fmt.Fprintln(os.Stderr, "ERROR", err)
   } else {
      fmt.Println(&query)
   }
}
```

See the [cql-cli source](cmd/cql-cli/main.go) for a more complete example.

## Building CQL programmatically

If you want to construct valid CQL queries without hand-assembling the AST, use the
fluent builder in `cqlbuilder` which validates inputs and escapes terms.

```go
import (
   "fmt"

   "github.com/indexdata/cql-go/cql"
   "github.com/indexdata/cql-go/cqlbuilder"
)

func main() {
   query, err := cqlbuilder.NewQuery().
      Prefix("dc", "http://purl.org/dc/elements/1.1/").
      Search("dc.title").
      Rel(cql.EQ).
      Term("the \"little\" prince").
      SortBy("dc.title", cql.IgnoreCase).
      Build()
   if err != nil {
      fmt.Fprintln(os.Stderr, "ERROR", err)
      return
   }
   fmt.Println(query.String())
}
```

## Conformance

The CQL specification requires that a query consist of a single search term with an optional index and relation:

`(index relation) term`

If multiple terms are provided in a query, they must be quoted:

`(index relation) "term1 term2"`

To remain compatible with existing implementations, like
[yaz](https://github.com/indexdata/yaz) or [CQL-java](https://github.com/indexdata/cql-java),
this parser relaxes this requirement and allows for providing multiple unquoted terms in a query:

`(index relation) term1 term2`

This behavior is enabled by default and can be disabled by setting the `Parser{Strict: true}` flag.

To allow for multiple terms in a query, the parser disambiguates between terms and relations
by assuming that a relation is either:

1. a symbolic relation: `=`, `==`, `<>`, `<` , `<=`, `>`, `>=`
2. a built-in named relation from the default context set:\
   `adj`, `any`, `all`, `exact`, `encloses`, `within`
4. a custom prefixed relation for which a prefix is declared, e.g:\
   `> dc = "http://deepcustard.org/" index dc.relation term`\
   Note that, unlike yaz or CQL-Java, terms that simply contain a `.` are not be considered relations unless the query includes a prefix declaration.
5. a custom relation, if a custom default context set is declared, e.g:\
   `> "http://deepcustard.org/" index relation term`\
   Note that, effectively, any term on the second position is considered a relation when a custom default context is declared.

These rules are applied in both the non-strict (default) and strict parsing modes.

# PGCQL

The pgcql package converts CQL to PostgreSQL.
CQL, while very limited compared to SQL, gives the ability to offer a subset
suitable for at least limiting query results.

The procedure is simple. First, define which fields are offered. At run time
for each incoming query, parse it (for syntax errors, etc) and secondly convert
the resulting tree to be used with a SQL query.

Example where we define and insert entries to a table by using the [pgx](https://github.com/jackc/pgx) package:

    conn, err := pgx.Connect(ctx, connStr)
    assert.NoError(t, err, "failed to connect to db")
    _, err = conn.Exec(ctx, "CREATE TABLE mytable (title TEXT, year INT, address JSONB)")
    assert.NoError(t, err, "failed to create mytable")
    _, err = conn.Query(ctx, "INSERT INTO mytable (title, year, address) "VALUES ($1, $2, $3)",
        "the art of computer programming", 1968, `{"city": "Reading", "country": "USA", "zip": 19601}`)
    assert.NoError(t, err, "failed to insert")

Now create the definition and add allowed CQL fields:

    def := pgcql.NewPgDefinition()

    def.AddField("title", pgcql.NewFieldString().WithExact())
    def.AddField("year", pgcql.NewFieldNumber())
    def.AddField("city", pgcql.NewFieldString().WithOps().WithColumn("address->>'city'"))

Handle query and inspect rows

    var query string

    var parser cql.Parser
    q, err := parser.Parse(query)
    assert.NoError(t, err, "parse of query failed")
    // create the SQL filter, where makes the query use $1 as first argument
    res, err := def.Parse(q, 1)
    assert.NoError(t, err, "pgcql conversion failed")
    var rows pgx.Rows
    rows, err = conn.Query(ctx, "SELECT id FROM mytable WHERE "+res.GetWhereClause(), res.GetQueryArguments()...)
    assert.NoErrorf(t, err, "failed to execute query '%s' whereClause='%s'", query, res.GetWhereClause())
    // inspect rows
    rows.Close()
