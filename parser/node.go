package parser

type NodeType int

const (
	Modifier NodeType = iota
	BoolOp
	SearchTerm
	SortOp
)

type Node struct {
	kind     NodeType
	index    string
	relation string
	term     string
	children []*Node
}
