package main

import (
	"fmt"
	"os"

	"github.com/indexdata/cql-go/parser"
)

func main() {
	var p parser.CqlParser
	for _, arg := range os.Args[1:] {
		node, err := p.Parse(arg)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Parse failed:", err.Error())
			os.Exit(1)
		}
		var xcql parser.Xcql
		fmt.Println(xcql.ToString(node, 2))
	}
}
