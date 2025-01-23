# CQL-Go 

## Introduction

CQL-Go is a library for decoding and encoding SRU
[CQL](https://www.loc.gov/standards/sru/cql/) queries.

The projects consists of a library as well as command-line
tool for parsing CQL queries.

## Usage

The primary purpose of the CQL library is to convert from string to
query structure. Here as a test case:


Minimal program which illustrates decoding from string (CQL) and
and encoding to string:

    func TestExampleParse(t *testing.T) {
        var parser cql.Parser

        in := "dc.title = abc"
        query, err := parser.Parse(in)
        assert.Nil(t, err)
        assert.Equal(t, in, query.String())
    }

The query tree can also be constructed manually:

    searchClause := &cql.SearchClause{Index: "dc.title", Relation: "=", Term: "abc"}
    query := cql.Query{Clause: cql.Clause{SearchClause: searchClause}}
    assert.Equal(t, "dc.title = abc", query.String())

## Further reading

 * [CQL](https://www.loc.gov/standards/sru/cql/)
 * [SRU](https://www.loc.gov/standards/sru/)




