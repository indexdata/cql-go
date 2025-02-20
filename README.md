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
