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

func (xcql *Xcql) toXmlMod(modifiers []*CqlNode, level int) {
	number := 0
	for _, n := range modifiers {
		mod := n.Search
		if mod == nil {
			return
		}
		if number == 0 {
			xcql.pr(level, "<modifiers>\n")
		}
		number++
		xcql.pr(level+1, "<modifier>\n")
		xcql.pr(level+2, "<type>")
		xcql.cdata(mod.Index)
		xcql.pr(0, "</type>\n")
		if len(mod.Relation) > 0 {
			xcql.pr(level+2, "<comparison>")
			xcql.cdata(mod.Relation)
			xcql.pr(0, "</comparison>\n")
		}
		if len(mod.Term) > 0 {
			xcql.pr(level+2, "<value>")
			xcql.cdata(mod.Term)
			xcql.pr(0, "</value>\n")
		}
		xcql.pr(level+1, "</modifier>\n")
	}
	if number > 0 {
		xcql.pr(level, "</modifiers>\n")
	}
}

func (xcql *Xcql) toXmlSb(node *CqlNode, level int) {
	if node.Search != nil {
		searchNode := node.Search
		xcql.pr(level, "<searchClause>\n")
		xcql.pr(level+1, "<index>")
		xcql.cdata(searchNode.Index)
		xcql.pr(0, "</index>\n")

		xcql.pr(level+1, "<relation>\n")
		xcql.pr(level+2, "<value>")
		xcql.cdata(searchNode.Relation)
		xcql.pr(0, "</value>\n")
		xcql.pr(level+1, "</relation>\n")
		xcql.toXmlMod(searchNode.Modifiers, level+1)
		xcql.pr(level+1, "<term>")
		xcql.cdata(searchNode.Term)
		xcql.pr(0, "</term>\n")
		xcql.pr(level, "</searchClause>\n")
	} else if node.Boolean != nil {
		booleanNode := node.Boolean
		xcql.pr(level, "<triple>\n")
		xcql.pr(level+1, "<Boolean>\n")
		xcql.pr(level+2, "<value>")
		xcql.cdata(booleanNode.Operator)
		xcql.pr(0, "</value>\n")
		xcql.toXmlMod(booleanNode.Modifiers, level+2)
		xcql.pr(level+1, "</Boolean>\n")

		xcql.pr(level+1, "<leftOperand>\n")
		xcql.toXmlSb(booleanNode.Left, level+2)
		xcql.pr(level+1, "</leftOperand>\n")

		xcql.pr(level+1, "<rightOperand>\n")
		xcql.toXmlSb(booleanNode.Right, level+2)
		xcql.pr(level+1, "</rightOperand>\n")
		xcql.pr(level, "</triple>\n")
	} else if node.Prefix != nil {
		// XCQL can ONLY represent prefixes
		xcql.toXmlSb(node.Prefix.Next, level)
	}
}

func (xcql *Xcql) toXmlPrefix(node *CqlNode, level int) {
	number := 0
	for ; node.Prefix != nil; number++ {
		prefixNode := node.Prefix
		if number == 0 {
			xcql.pr(level, "<prefixes>\n")
		}
		xcql.pr(level+1, "<prefix>\n")
		xcql.pr(level+2, "<name>")
		xcql.cdata(prefixNode.Prefix)
		xcql.pr(0, "</name>\n")
		xcql.pr(level+2, "<identifier>")
		xcql.cdata(prefixNode.Uri)
		xcql.pr(0, "</identifier>\n")
		xcql.pr(level+1, "</prefix>\n")
		node = prefixNode.Next
	}
	if number > 0 {
		xcql.pr(level, "</prefixes>\n")
	}
	if node.Search != nil {
		xcql.pr(level, "<triple>\n")
		xcql.toXmlSb(node, level+1)
		xcql.pr(level, "</triple>\n")
	} else if node.Boolean != nil {
		xcql.toXmlSb(node, level)
	}
}

func (xcql *Xcql) toXmlSort(node *CqlNode, level int) {
	// skip over sort nodes as they have to be emitted last
	var node1 = node
	for node1.Sort != nil {
		node1 = node1.Sort.Next
	}
	xcql.toXmlPrefix(node1, level)

	// sort nodes at the end
	number := 0
	for ; node.Sort != nil; number++ {
		sortNode := node.Sort
		if number == 0 {
			xcql.pr(level, "<sortKeys>\n")
		}
		xcql.pr(level+1, "<key>\n")
		xcql.pr(level+2, "<index>")
		xcql.cdata(sortNode.Index)
		xcql.pr(0, "</index>\n")
		xcql.toXmlMod(sortNode.Modifiers, level+2)
		xcql.pr(level+1, "</key>\n")
		node = sortNode.Next
	}
	if number > 0 {
		xcql.pr(level, "</sortKeys>\n")
	}
}

func (xcql *Xcql) ToString(node *CqlNode, tab int) string {
	xcql.sb.Reset()
	xcql.tab = tab
	xcql.pr(0, "<xcql xmlns=\"http://docs.oasis-open.org/ns/search-ws/xcql\">\n")
	xcql.toXmlSort(node, 1)
	xcql.pr(0, "</xcql>\n")
	return xcql.sb.String()
}
