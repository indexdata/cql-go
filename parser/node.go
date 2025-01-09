package parser

import (
	"strings"
	"unicode/utf8"
)

const (
	Modifier int = iota
	BoolOp
	SearchTerm
)

type Node struct {
	kind     int
	index    string
	relation string
	term     string
	children []*Node
}

func cdata(sb *strings.Builder, msg string) {
	pos := 0
	for pos < len(msg) {
		r, w := utf8.DecodeRuneInString(msg[pos:])
		switch r {
		case utf8.RuneError:
			return
		case '&':
			sb.WriteString("&amp;")
		case '<':
			sb.WriteString("&lt;")
		case '>':
			sb.WriteString("&gt;")
		default:
			sb.WriteRune(r)
		}
		pos += w
	}
}

func pr(sb *strings.Builder, level int, msg string) {
	for i := 0; i < level; i++ {
		sb.WriteString("  ")
	}
	sb.WriteString(msg)
}

func (n *Node) toXmlMod(sb *strings.Builder, level int) {
	number := 0
	for _, elem := range n.children {
		if elem.kind == Modifier {
			if number == 0 {
				pr(sb, level, "<modifiers>\n")
			}
			pr(sb, level+1, "<modifier>\n")
			number++
			pr(sb, level+2, "<type>")
			cdata(sb, elem.index)
			pr(sb, 0, "<type>\n")
			if len(elem.relation) > 0 {
				pr(sb, level+2, "<comparison>")
				cdata(sb, elem.relation)
				pr(sb, 0, "</comparison>\n")
			}
			if len(elem.term) > 0 {
				pr(sb, level+2, "<value>")
				cdata(sb, elem.term)
				pr(sb, 0, "</value>\n")
			}
			pr(sb, level+1, "</modifier>\n")
		}
	}
	if number > 0 {
		pr(sb, level, "</modifiers>\n")
	}

}

func (n *Node) toXmlSb(sb *strings.Builder, level int) {
	switch n.kind {
	case SearchTerm:
		pr(sb, level, "<searchClause>\n")
		pr(sb, level+1, "<index>")
		cdata(sb, n.index)
		pr(sb, 0, "<index>\n")

		pr(sb, level+1, "<relation>\n")
		pr(sb, level+2, "<value>")
		cdata(sb, n.relation)
		pr(sb, 0, "</value>\n")
		pr(sb, level+1, "</relation>\n")
		n.toXmlMod(sb, level+1)
		pr(sb, level+1, "<term>")
		cdata(sb, n.term)
		pr(sb, 0, "<term>\n")

		pr(sb, level, "</searchClause>\n")
	case BoolOp:
		pr(sb, level, "<triple>\n")
		pr(sb, level+1, "<boolean>\n")
		pr(sb, level+2, "<value>")
		cdata(sb, n.index)
		pr(sb, 0, "</value>\n")
		n.toXmlMod(sb, level+2)
		pr(sb, level+1, "</boolean>\n")

		pr(sb, level+1, "<leftOperand>\n")
		n.children[0].toXmlSb(sb, level+2)
		pr(sb, level+1, "</leftOperand>\n")

		pr(sb, level+1, "<rightOperand>\n")
		n.children[1].toXmlSb(sb, level+2)
		pr(sb, level+1, "</rightOperand>\n")
		pr(sb, level, "</triple>\n")
	}
}

func (n *Node) ToXml() string {
	var sb strings.Builder
	n.toXmlSb(&sb, 0)
	return sb.String()
}
