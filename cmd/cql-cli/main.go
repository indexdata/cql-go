package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/indexdata/cql-go/cql"
)

func main() {
	var outFmt string
	flag.StringVar(&outFmt, "t", "cql", "output format: cql, xcql")
	flag.Parse()
	var p cql.Parser
	for _, arg := range flag.Args() {
		query, err := p.Parse(arg)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Parse failed:", err.Error())
			os.Exit(1)
		}
		switch outFmt {
		case "cql":
			fmt.Println(&query)
		case "xcql":
			fmt.Print((&cql.Xcql{}).Marshal(query, 2))
		default:
			fmt.Fprintln(os.Stderr, "Unknown output format:", outFmt)
			os.Exit(1)
		}
	}
}
