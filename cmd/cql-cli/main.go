package main

import (
	"encoding/json"
	"encoding/xml"
	"flag"
	"fmt"
	"os"

	"github.com/indexdata/cql-go/cql"
)

func main() {
	var outFmt string
	var strict bool
	flag.StringVar(&outFmt, "t", "cql", "output format: cql, xcql, struct")
	flag.BoolVar(&strict, "s", false, "strict CQL (e.g no multi-terms)")
	flag.Parse()
	var p cql.Parser
	for _, arg := range flag.Args() {
		p.Strict = strict
		query, err := p.Parse(arg)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Parse failed:", err.Error())
			os.Exit(1)
		}
		switch outFmt {
		case "cql":
			fmt.Println(&query)
		case "json":
			out, _ := json.MarshalIndent(&query, "", "  ")
			fmt.Printf("%s\n", out)
		case "xml":
			out, _ := xml.MarshalIndent(&query, "", "  ")
			fmt.Printf("%s\n", out)
		case "struct":
			fmt.Printf("%+v\n", query)
		case "xcql":
			fmt.Print((&cql.Xcql{}).Marshal(query, 2))
		default:
			fmt.Fprintln(os.Stderr, "Unknown output format:", outFmt)
			os.Exit(1)
		}
	}
}
