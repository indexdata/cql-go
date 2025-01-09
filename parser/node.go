package parser

const (
	Top int = iota
	Modifier
	BoolOp
	SearchTerm
	Sort
)

type Node struct {
	kind     int
	index    string
	relation string
	term     string
	children []*Node
}
