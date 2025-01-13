package cql

type NodeType int

type SearchClauseNode struct {
	Index     string
	Relation  string
	Term      string
	Modifiers []*CqlNode
}

type BooleanNode struct {
	Operator  string
	Relation  string
	Modifiers []*CqlNode
	Left      *CqlNode
	Right     *CqlNode
}

type SortNode struct {
	Index     string
	Modifiers []*CqlNode
	Next      *CqlNode
}

type PrefixNode struct {
	Prefix string
	Uri    string
	Next   *CqlNode
}

type CqlNode struct {
	Search  *SearchClauseNode
	Boolean *BooleanNode
	Sort    *SortNode
	Prefix  *PrefixNode
}
