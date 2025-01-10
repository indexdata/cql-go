package parser

import (
	"strings"
	"unicode/utf8"
)

type Xcql struct {
	sb  strings.Builder
	tab int
}

func (x *Xcql) init() {
	x.tab = 2
}

func (x *Xcql) cdata(msg string) {
	pos := 0
	for pos < len(msg) {
		r, w := utf8.DecodeRuneInString(msg[pos:])
		switch r {
		case utf8.RuneError:
			return
		case '&':
			x.sb.WriteString("&amp;")
		case '<':
			x.sb.WriteString("&lt;")
		case '>':
			x.sb.WriteString("&gt;")
		default:
			x.sb.WriteRune(r)
		}
		pos += w
	}
}

func (x *Xcql) pr(level int, msg string) {
	for i := 0; i < level*x.tab; i++ {
		x.sb.WriteString(" ")
	}
	x.sb.WriteString(msg)
}

func (x *Xcql) toXmlMod(n *Node, level int) {
	number := 0
	for _, elem := range n.children {
		if elem.kind == Modifier {
			if number == 0 {
				x.pr(level, "<modifiers>\n")
			}
			x.pr(level+1, "<modifier>\n")
			number++
			x.pr(level+2, "<type>")
			x.cdata(elem.index)
			x.pr(0, "</type>\n")
			if len(elem.relation) > 0 {
				x.pr(level+2, "<comparison>")
				x.cdata(elem.relation)
				x.pr(0, "</comparison>\n")
			}
			if len(elem.term) > 0 {
				x.pr(level+2, "<value>")
				x.cdata(elem.term)
				x.pr(0, "</value>\n")
			}
			x.pr(level+1, "</modifier>\n")
		}
	}
	if number > 0 {
		x.pr(level, "</modifiers>\n")
	}
}

func (x *Xcql) toXmlSb(n *Node, level int) {
	switch n.kind {
	case SearchTerm:
		x.pr(level, "<searchClause>\n")
		x.pr(level+1, "<index>")
		x.cdata(n.index)
		x.pr(0, "</index>\n")

		x.pr(level+1, "<relation>\n")
		x.pr(level+2, "<value>")
		x.cdata(n.relation)
		x.pr(0, "</value>\n")
		x.pr(level+1, "</relation>\n")
		x.toXmlMod(n, level+1)
		x.pr(level+1, "<term>")
		x.cdata(n.term)
		x.pr(0, "</term>\n")
		x.pr(level, "</searchClause>\n")
	case BoolOp:
		x.pr(level, "<triple>\n")
		x.pr(level+1, "<boolean>\n")
		x.pr(level+2, "<value>")
		x.cdata(n.index)
		x.pr(0, "</value>\n")
		x.toXmlMod(n, level+2)
		x.pr(level+1, "</boolean>\n")

		x.pr(level+1, "<leftOperand>\n")
		x.toXmlSb(n.children[0], level+2)
		x.pr(level+1, "</leftOperand>\n")

		x.pr(level+1, "<rightOperand>\n")
		x.toXmlSb(n.children[1], level+2)
		x.pr(level+1, "</rightOperand>\n")
		x.pr(level, "</triple>\n")
	}
}

func (x *Xcql) toXmlTop(n *Node, level int) {
	if n.kind == SortOp {
		x.toXmlTop(n.children[0], level)
		x.pr(level, "<sortKeys>\n")
		for _, ni := range n.children[1:] {
			x.pr(level+1, "<key>\n")
			x.pr(level+2, "<index>")
			x.cdata(ni.index)
			x.pr(0, "</index>\n")
			x.toXmlMod(ni, level+2)
			x.pr(level+1, "</key>\n")
		}
		x.pr(level, "</sortKeys>\n")
		return
	}
	if n.kind == SearchTerm {
		x.pr(level, "<triple>\n")
		x.toXmlSb(n, level+1)
		x.pr(level, "</triple>\n")
	} else if n.kind == BoolOp {
		x.toXmlSb(n, level)
	}
}

func (x *Xcql) ToString(n *Node, tab int) string {
	x.sb.Reset()
	x.tab = tab
	x.pr(0, "<xcql xmlns=\"http://docs.oasis-open.org/ns/search-ws/xcql\">\n")
	x.toXmlTop(n, 1)
	x.pr(0, "</xcql>\n")
	return x.sb.String()
}
