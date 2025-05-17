package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/indexdata/cql-go/cql"
	"github.com/indexdata/cql-go/pgcql"
)

func main() {
	var serverChoiceColumn string
	flag.StringVar(&serverChoiceColumn, "s", "text", "column for cql.serverChoice")
	def := &pgcql.PgDefinition{}
	if serverChoiceColumn != "" {
		serverChoice := &pgcql.FieldString{}
		serverChoice.WithFullText("english").SetColumn(serverChoiceColumn)
		def.AddField("cql.serverChoice", serverChoice)
	}
	flag.Parse()
	if len(flag.Args()) == 0 {
		fmt.Println("Usage: pgcql-cli [-s serverchoicefield] field .. query")
		fmt.Println("Example: pgcql-cli -s notes ti \"free and ti=powerful\"")
		os.Exit(1)
	}
	for i := 0; i < len(flag.Args()); i++ {
		if i < len(flag.Args())-1 {
			field := &pgcql.FieldString{}
			field.WithLikeOps()
			def.AddField(flag.Args()[i], field)
			continue
		}
		var parser cql.Parser
		query, err := parser.Parse(flag.Args()[i])
		if err != nil {
			fmt.Printf("cql error: %s", err)
			return
		}
		res, err := def.Parse(query, 1)
		if err != nil {
			fmt.Printf("pgcql error: %s\n", err)
			return
		}
		fmt.Printf("whereByClause: %s\n", res.GetWhereClause())
		for j, qa := range res.GetQueryArguments() {
			fmt.Printf("$%d: %v\n", j+1, qa)
		}
	}
}
