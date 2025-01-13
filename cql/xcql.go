package cql

import (
	"strings"
	"unicode/utf8"
)

type Xcql struct {
	sb  strings.Builder
	tab int
}

func (xcql *Xcql) init() {
	xcql.tab = 2
}

func (xcql *Xcql) cdata(msg string) {
	pos := 0
	for pos < len(msg) {
		r, w := utf8.DecodeRuneInString(msg[pos:])
		switch r {
		case utf8.RuneError:
			return
		case '&':
			xcql.sb.WriteString("&amp;")
		case '<':
			xcql.sb.WriteString("&lt;")
		case '>':
			xcql.sb.WriteString("&gt;")
		default:
			xcql.sb.WriteRune(r)
		}
		pos += w
	}
}

func (xcql *Xcql) pr(level int, msg string) {
	for i := 0; i < level*xcql.tab; i++ {
		xcql.sb.WriteString(" ")
	}
	xcql.sb.WriteString(msg)
}

func (xcql *Xcql) toXmlMod(node *Node, level int) {
	number := 0
	for _, n := range node.children {
		if n.kind == Modifier {
			if number == 0 {
				xcql.pr(level, "<modifiers>\n")
			}
			xcql.pr(level+1, "<modifier>\n")
			number++
			xcql.pr(level+2, "<type>")
			xcql.cdata(n.index)
			xcql.pr(0, "</type>\n")
			if len(n.relation) > 0 {
				xcql.pr(level+2, "<comparison>")
				xcql.cdata(n.relation)
				xcql.pr(0, "</comparison>\n")
			}
			if len(n.term) > 0 {
				xcql.pr(level+2, "<value>")
				xcql.cdata(n.term)
				xcql.pr(0, "</value>\n")
			}
			xcql.pr(level+1, "</modifier>\n")
		}
	}
	if number > 0 {
		xcql.pr(level, "</modifiers>\n")
	}
}

func (xcql *Xcql) toXmlSb(node *Node, level int) {
	switch node.kind {
	case SearchTerm:
		xcql.pr(level, "<searchClause>\n")
		xcql.pr(level+1, "<index>")
		xcql.cdata(node.index)
		xcql.pr(0, "</index>\n")

		xcql.pr(level+1, "<relation>\n")
		xcql.pr(level+2, "<value>")
		xcql.cdata(node.relation)
		xcql.pr(0, "</value>\n")
		xcql.pr(level+1, "</relation>\n")
		xcql.toXmlMod(node, level+1)
		xcql.pr(level+1, "<term>")
		xcql.cdata(node.term)
		xcql.pr(0, "</term>\n")
		xcql.pr(level, "</searchClause>\n")
	case BoolOp:
		xcql.pr(level, "<triple>\n")
		xcql.pr(level+1, "<Boolean>\n")
		xcql.pr(level+2, "<value>")
		xcql.cdata(node.index)
		xcql.pr(0, "</value>\n")
		xcql.toXmlMod(node, level+2)
		xcql.pr(level+1, "</Boolean>\n")

		xcql.pr(level+1, "<leftOperand>\n")
		xcql.toXmlSb(node.children[0], level+2)
		xcql.pr(level+1, "</leftOperand>\n")

		xcql.pr(level+1, "<rightOperand>\n")
		xcql.toXmlSb(node.children[1], level+2)
		xcql.pr(level+1, "</rightOperand>\n")
		xcql.pr(level, "</triple>\n")
	case Prefix:
		xcql.toXmlSb(node.children[0], level)
	}
}

func (xcql *Xcql) toXmlTop(node *Node, level int) {
	number := 0
	for ; node.kind == Prefix; number++ {
		if number == 0 {
			xcql.pr(level, "<prefixes>\n")
		}
		xcql.pr(level+1, "<prefix>\n")
		xcql.pr(level+2, "<name>")
		xcql.cdata(node.index)
		xcql.pr(0, "</name>\n")
		xcql.pr(level+2, "<identifier>")
		xcql.cdata(node.term)
		xcql.pr(0, "</identifier>\n")
		xcql.pr(level+1, "</prefix>\n")
		node = node.children[0]
	}
	if number > 0 {
		xcql.pr(level, "</prefixes>\n")
	}
	if node.kind == SortOp {
		xcql.toXmlTop(node.children[0], level)
		xcql.pr(level, "<sortKeys>\n")
		for _, n := range node.children[1:] {
			xcql.pr(level+1, "<key>\n")
			xcql.pr(level+2, "<index>")
			xcql.cdata(n.index)
			xcql.pr(0, "</index>\n")
			xcql.toXmlMod(n, level+2)
			xcql.pr(level+1, "</key>\n")
		}
		xcql.pr(level, "</sortKeys>\n")
		return
	}
	if node.kind == SearchTerm {
		xcql.pr(level, "<triple>\n")
		xcql.toXmlSb(node, level+1)
		xcql.pr(level, "</triple>\n")
	} else if node.kind == BoolOp {
		xcql.toXmlSb(node, level)
	}
}

func (xcql *Xcql) ToString(node *Node, tab int) string {
	xcql.sb.Reset()
	xcql.tab = tab
	xcql.pr(0, "<xcql xmlns=\"http://docs.oasis-open.org/ns/search-ws/xcql\">\n")
	xcql.toXmlTop(node, 1)
	xcql.pr(0, "</xcql>\n")
	return xcql.sb.String()
}
