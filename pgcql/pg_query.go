package pgcql

import (
	"fmt"
	"strings"

	"github.com/indexdata/cql-go/cql"
)

type PgQuery struct {
	def                *PgDefinition
	queryArgumentIndex int
	arguments          []any
	whereClause        string
	orderByClause      string
	orderByFields      []string
}

func (p *PgQuery) parse(q cql.Query, queryArgumentIndex int, def *PgDefinition) error {
	p.def = def
	p.arguments = make([]any, 0)
	p.queryArgumentIndex = queryArgumentIndex
	p.orderByFields = make([]string, 0)
	err := p.parseClause(q.Clause, 0)
	if err != nil {
		return err
	}
	return p.parseSortSpec(q.SortSpec)

}

func (p *PgQuery) parseSortSpec(sortSpec []cql.Sort) error {
	if len(sortSpec) == 0 {
		return nil
	}
	p.orderByClause = " ORDER BY "
	for i, sortField := range sortSpec {
		if i > 0 {
			p.orderByClause += ", "
		}
		fieldType := p.def.GetFieldType(sortField.Index)
		if fieldType == nil {
			return &PgError{message: fmt.Sprintf("unknown field %s", sortField.Index)}
		}
		sort := fieldType.Sort()
		if sort == "" {
			return &PgError{message: fmt.Sprintf("field %s does not support sorting", sortField.Index)}
		}
		p.orderByClause += sort
		p.orderByFields = append(p.orderByFields, sort)
		dir := ""
		for _, modifier := range sortField.Modifiers {
			if strings.EqualFold(modifier.Name, "sort.ascending") {
				dir = ""
			} else if strings.EqualFold(modifier.Name, "sort.descending") {
				dir = " DESC"
			} else {
				return &PgError{message: fmt.Sprintf("unsupported sort modifier %s", modifier.Name)}
			}
		}
		p.orderByClause += dir
	}
	return nil
}

func (p *PgQuery) parseClause(sc cql.Clause, level int) error {
	if sc.SearchClause != nil {
		index := sc.SearchClause.Index
		fieldType := p.def.GetFieldType(index)
		if fieldType == nil {
			return &PgError{message: fmt.Sprintf("unknown field %s", index)}
		}
		sql, args, err := fieldType.Generate(*sc.SearchClause, p.queryArgumentIndex)
		if err != nil {
			return err
		}
		p.whereClause += sql
		if args != nil {
			p.queryArgumentIndex += len(args)
			p.arguments = append(p.arguments, args...)
		}
		return nil
	} else if sc.BoolClause != nil {
		if level > 0 {
			p.whereClause += "("
		}
		err := p.parseClause(sc.BoolClause.Left, level+1)
		if err != nil {
			return err
		}
		switch sc.BoolClause.Operator {
		case cql.AND:
			p.whereClause += " AND "
		case cql.OR:
			p.whereClause += " OR "
		case cql.NOT:
			p.whereClause += " AND NOT "
		default:
			return &PgError{message: fmt.Sprintf("unsupported operator %s", sc.BoolClause.Operator)}
		}
		err = p.parseClause(sc.BoolClause.Right, level+1)
		if err != nil {
			return err
		}
		if level > 0 {
			p.whereClause += ")"
		}
		return nil
	}
	return &PgError{message: "unsupported clause type"}
}

func (p *PgQuery) GetWhereClause() string {
	return p.whereClause
}

func (p *PgQuery) GetQueryArguments() []any {
	return p.arguments
}

func (p *PgQuery) GetOrderByClause() string {
	return p.orderByClause
}

func (p *PgQuery) GetOrderByFields() []string {
	return p.orderByFields
}
