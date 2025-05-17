package cql

import (
	"encoding/xml"
	"io"
)

type Xcql struct {
	w   io.Writer
	err error
	tab int
}

func (xcql *Xcql) cdata(msg string) {
	err := xml.EscapeText(xcql.w, []byte(msg))
	if err != nil && xcql.err == nil {
		xcql.err = err
	}
}

func (xcql *Xcql) write(msg string) {
	_, err := xcql.w.Write([]byte(msg))
	if err != nil && xcql.err == nil {
		xcql.err = err
	}
}

func (xcql *Xcql) pr(level int, msg string) {
	for i := 0; i < level*xcql.tab; i++ {
		xcql.write(" ")
	}
	xcql.write(msg)
}

func (xcql *Xcql) toXmlMod(modifiers []Modifier, level int) {
	number := 0
	for _, mod := range modifiers {
		if number == 0 {
			xcql.pr(level, "<modifiers>\n")
		}
		number++
		xcql.pr(level+1, "<modifier>\n")
		xcql.pr(level+2, "<type>")
		xcql.cdata(mod.Name)
		xcql.pr(0, "</type>\n")
		if len(mod.Relation) > 0 {
			xcql.pr(level+2, "<comparison>")
			xcql.cdata(string(mod.Relation))
			xcql.pr(0, "</comparison>\n")
			xcql.pr(level+2, "<value>")
			xcql.cdata(mod.Value)
			xcql.pr(0, "</value>\n")
		}
		xcql.pr(level+1, "</modifier>\n")
	}
	if number > 0 {
		xcql.pr(level, "</modifiers>\n")
	}
}

func (xcql *Xcql) toXmlNode(node Clause, level int) {
	// there could be prefix handling here, but we only deal with them in toXmlPrefix
	// to conform to XSCL schema
	if node.SearchClause != nil {
		xcql.pr(level, "<searchClause>\n")
		xcql.pr(level+1, "<index>")
		xcql.cdata(node.SearchClause.Index)
		xcql.pr(0, "</index>\n")

		xcql.pr(level+1, "<relation>\n")
		xcql.pr(level+2, "<value>")
		xcql.cdata(string(node.SearchClause.Relation))
		xcql.pr(0, "</value>\n")
		xcql.pr(level+1, "</relation>\n")
		xcql.toXmlMod(node.SearchClause.Modifiers, level+1)
		xcql.pr(level+1, "<term>")
		xcql.cdata(node.SearchClause.Term)
		xcql.pr(0, "</term>\n")
		xcql.pr(level, "</searchClause>\n")
	} else if node.BoolClause != nil {
		xcql.pr(level, "<triple>\n")
		xcql.pr(level+1, "<Boolean>\n") // XCQL schema: Capital B! , unlike earlier versions
		xcql.pr(level+2, "<value>")
		xcql.cdata(string(node.BoolClause.Operator))
		xcql.pr(0, "</value>\n")
		xcql.toXmlMod(node.BoolClause.Modifiers, level+2)
		xcql.pr(level+1, "</Boolean>\n")

		xcql.pr(level+1, "<leftOperand>\n")
		xcql.toXmlNode(node.BoolClause.Left, level+2)
		xcql.pr(level+1, "</leftOperand>\n")

		xcql.pr(level+1, "<rightOperand>\n")
		xcql.toXmlNode(node.BoolClause.Right, level+2)
		xcql.pr(level+1, "</rightOperand>\n")
		xcql.pr(level, "</triple>\n")
	}
}

func (xcql *Xcql) toXmlPrefix(node Clause, level int) {
	number := 0
	for _, prefix := range node.PrefixMap {
		if number == 0 {
			xcql.pr(level, "<prefixes>\n")
		}
		number++
		xcql.pr(level+1, "<prefix>\n")
		xcql.pr(level+2, "<name>")
		xcql.cdata(prefix.Prefix)
		xcql.pr(0, "</name>\n")
		xcql.pr(level+2, "<identifier>")
		xcql.cdata(prefix.Uri)
		xcql.pr(0, "</identifier>\n")
		xcql.pr(level+1, "</prefix>\n")
	}
	if number > 0 {
		xcql.pr(level, "</prefixes>\n")
	}
	if node.SearchClause != nil {
		xcql.pr(level, "<triple>\n") // very unfortunate that XCQL schema requires this
		xcql.toXmlNode(node, level+1)
		xcql.pr(level, "</triple>\n")
	} else if node.BoolClause != nil {
		xcql.toXmlNode(node, level)
	}
}

func (xcql *Xcql) toXmlSort(query Query, level int) {
	xcql.toXmlPrefix(query.Clause, level)
	number := 0
	for _, sort := range query.SortSpec {
		if number == 0 {
			xcql.pr(level, "<sortKeys>\n")
		}
		number++
		xcql.pr(level+1, "<key>\n")
		xcql.pr(level+2, "<index>")
		xcql.cdata(sort.Index)
		xcql.pr(0, "</index>\n")
		xcql.toXmlMod(sort.Modifiers, level+2)
		xcql.pr(level+1, "</key>\n")
	}
	if number > 0 {
		xcql.pr(level, "</sortKeys>\n")
	}
}

func (xcql *Xcql) Write(query Query, tab int, w io.Writer) error {
	xcql.w = w
	xcql.tab = tab
	xcql.pr(0, "<xcql xmlns=\"http://docs.oasis-open.org/ns/search-ws/xcql\">\n")
	xcql.toXmlSort(query, 1)
	xcql.pr(0, "</xcql>\n")
	return xcql.err
}
