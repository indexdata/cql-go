package cql

import (
	"testing"
)

func TestSimple(t *testing.T) {
	var p Parser

	node, err := p.Parse("beta")
	if err != nil || node == nil {
		t.Errorf("expected ok")
	}

	_, err = p.Parse("")
	if err == nil {
		t.Errorf("expected error")
	}
}

func TestParseXml(t *testing.T) {
	var p Parser

	for _, testcase := range []struct {
		name   string
		input  string
		strict bool
		ok     bool
		expect string
	}{
		{
			name:   "single term",
			input:  "myterm",
			strict: false,
			ok:     true,
			expect: `<xcql xmlns="http://docs.oasis-open.org/ns/search-ws/xcql">
<triple>
<searchClause>
<index>cql.serverChoice</index>
<relation>
<value>=</value>
</relation>
<term>myterm</term>
</searchClause>
</triple>
</xcql>
`,
		},
		{
			name:   "term rel value",
			input:  "dc.title all andersen",
			strict: false,
			ok:     true,
			expect: `<xcql xmlns="http://docs.oasis-open.org/ns/search-ws/xcql">
<triple>
<searchClause>
<index>dc.title</index>
<relation>
<value>all</value>
</relation>
<term>andersen</term>
</searchClause>
</triple>
</xcql>
`,
		},
		{
			name:   "relation modifiers",
			input:  "dc.title =/k1=v1/k2 andersen",
			strict: false,
			ok:     true,
			expect: `<xcql xmlns="http://docs.oasis-open.org/ns/search-ws/xcql">
<triple>
<searchClause>
<index>dc.title</index>
<relation>
<value>=</value>
</relation>
<modifiers>
<modifier>
<type>k1</type>
<comparison>=</comparison>
<value>v1</value>
</modifier>
<modifier>
<type>k2</type>
</modifier>
</modifiers>
<term>andersen</term>
</searchClause>
</triple>
</xcql>
`,
		},
		{
			name:   "booleans",
			input:  "year < 1990 and b",
			strict: false,
			ok:     true,
			expect: `<xcql xmlns="http://docs.oasis-open.org/ns/search-ws/xcql">
<triple>
<Boolean>
<value>and</value>
</Boolean>
<leftOperand>
<searchClause>
<index>year</index>
<relation>
<value>&lt;</value>
</relation>
<term>1990</term>
</searchClause>
</leftOperand>
<rightOperand>
<searchClause>
<index>cql.serverChoice</index>
<relation>
<value>=</value>
</relation>
<term>b</term>
</searchClause>
</rightOperand>
</triple>
</xcql>
`,
		},
		{
			name:   "prox",
			input:  "(a prox/order=1 (b))",
			strict: false,
			ok:     true,
			expect: `<xcql xmlns="http://docs.oasis-open.org/ns/search-ws/xcql">
<triple>
<Boolean>
<value>prox</value>
<modifiers>
<modifier>
<type>order</type>
<comparison>=</comparison>
<value>1</value>
</modifier>
</modifiers>
</Boolean>
<leftOperand>
<searchClause>
<index>cql.serverChoice</index>
<relation>
<value>=</value>
</relation>
<term>a</term>
</searchClause>
</leftOperand>
<rightOperand>
<searchClause>
<index>cql.serverChoice</index>
<relation>
<value>=</value>
</relation>
<term>b</term>
</searchClause>
</rightOperand>
</triple>
</xcql>
`,
		},
		{
			name:   "sort",
			input:  "myterm1 sortby title",
			strict: false,
			ok:     true,
			expect: `<xcql xmlns="http://docs.oasis-open.org/ns/search-ws/xcql">
<triple>
<searchClause>
<index>cql.serverChoice</index>
<relation>
<value>=</value>
</relation>
<term>myterm1</term>
</searchClause>
</triple>
<sortKeys>
<key>
<index>title</index>
</key>
</sortKeys>
</xcql>
`,
		},
		{
			name:   "sort2",
			input:  "myterm1 sortby title year/asc",
			strict: false,
			ok:     true,
			expect: `<xcql xmlns="http://docs.oasis-open.org/ns/search-ws/xcql">
<triple>
<searchClause>
<index>cql.serverChoice</index>
<relation>
<value>=</value>
</relation>
<term>myterm1</term>
</searchClause>
</triple>
<sortKeys>
<key>
<index>title</index>
</key>
<key>
<index>year</index>
<modifiers>
<modifier>
<type>asc</type>
</modifier>
</modifiers>
</key>
</sortKeys>
</xcql>
`,
		},
		{
			name:   "sort3",
			input:  "ti=a and b sortby title",
			strict: false,
			ok:     true,
			expect: `<xcql xmlns="http://docs.oasis-open.org/ns/search-ws/xcql">
<triple>
<Boolean>
<value>and</value>
</Boolean>
<leftOperand>
<searchClause>
<index>ti</index>
<relation>
<value>=</value>
</relation>
<term>a</term>
</searchClause>
</leftOperand>
<rightOperand>
<searchClause>
<index>cql.serverChoice</index>
<relation>
<value>=</value>
</relation>
<term>b</term>
</searchClause>
</rightOperand>
</triple>
<sortKeys>
<key>
<index>title</index>
</key>
</sortKeys>
</xcql>
`,
		},
		{
			name:   "prefix1",
			input:  ">dc = uri dc.ti = a",
			strict: false,
			ok:     true,
			expect: `<xcql xmlns="http://docs.oasis-open.org/ns/search-ws/xcql">
<prefixes>
<prefix>
<name>dc</name>
<identifier>uri</identifier>
</prefix>
</prefixes>
<triple>
<searchClause>
<index>dc.ti</index>
<relation>
<value>=</value>
</relation>
<term>a</term>
</searchClause>
</triple>
</xcql>
`,
		},
		{
			name:   "prefix2",
			input:  ">a =uri1>uri2>b=uri3 dc.ti = a",
			strict: false,
			ok:     true,
			expect: `<xcql xmlns="http://docs.oasis-open.org/ns/search-ws/xcql">
<prefixes>
<prefix>
<name>a</name>
<identifier>uri1</identifier>
</prefix>
<prefix>
<name></name>
<identifier>uri2</identifier>
</prefix>
<prefix>
<name>b</name>
<identifier>uri3</identifier>
</prefix>
</prefixes>
<triple>
<searchClause>
<index>dc.ti</index>
<relation>
<value>=</value>
</relation>
<term>a</term>
</searchClause>
</triple>
</xcql>
`,
		},
		{
			name:   "prefix3",
			input:  "a and (>dc=uri dc.ti = a)",
			strict: false,
			ok:     true,
			expect: `<xcql xmlns="http://docs.oasis-open.org/ns/search-ws/xcql">
<triple>
<Boolean>
<value>and</value>
</Boolean>
<leftOperand>
<searchClause>
<index>cql.serverChoice</index>
<relation>
<value>=</value>
</relation>
<term>a</term>
</searchClause>
</leftOperand>
<rightOperand>
<searchClause>
<index>dc.ti</index>
<relation>
<value>=</value>
</relation>
<term>a</term>
</searchClause>
</rightOperand>
</triple>
</xcql>
`,
		},
		{
			name:   "",
			input:  "",
			strict: false,
			ok:     false,
			expect: "search term expected near pos 0",
		},
		{
			name:   "(",
			input:  "(",
			strict: false,
			ok:     false,
			expect: "search term expected near pos 1",
		},
		{
			name:   "(a)",
			input:  "(a",
			strict: false,
			ok:     false,
			expect: "missing ) near pos 2",
		},
		{
			name:   "(a))",
			input:  "(a))",
			strict: false,
			ok:     false,
			expect: "EOF expected near pos 4",
		},
		{
			name:   "dc.ti =",
			input:  "dc.ti =",
			strict: false,
			ok:     false,
			expect: "search term expected near pos 7",
		},
		{
			name:   "a and",
			input:  "a and",
			strict: false,
			ok:     false,
			expect: "search term expected near pos 5",
		},
		{
			name:   "a and /",
			input:  "a and /",
			strict: false,
			ok:     false,
			expect: "missing modifier key near pos 7",
		},
		{
			name:   "a =/",
			input:  "a =/",
			strict: false,
			ok:     false,
			expect: "missing modifier key near pos 4",
		},
		{
			name:   "a =/",
			input:  "a =/b=",
			strict: false,
			ok:     false,
			expect: "missing modifier value near pos 6",
		},
		{
			name:   ">",
			input:  ">",
			strict: false,
			ok:     false,
			expect: "term expected after > near pos 1",
		},
		{
			name:   ">dc=()",
			input:  ">dc=()",
			strict: false,
			ok:     false,
			expect: "term expected after = near pos 6",
		},
		{
			name:   ">dc=uri",
			input:  ">dc=uri",
			strict: false,
			ok:     false,
			expect: "search term expected near pos 7",
		},
		{
			name:   "a sortby year/",
			input:  "a sortby year/",
			strict: false,
			ok:     false,
			expect: "missing modifier key near pos 14",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			node, err := p.Parse(testcase.input)
			if testcase.ok {
				if err != nil {
					t.Fatalf("expected OK for query %s . Got error: %s", testcase.input, err)
				}
				var xcql Xcql
				xml := xcql.ToString(node, 0)
				if xml != testcase.expect {
					t.Fatalf("Different XML for query %s\nExpect:\n%s\nGot:\n%s", testcase.input, testcase.expect, xml)
				}
			} else {
				if err == nil {
					t.Fatalf("expected Failure for query %s", testcase.input)
				}
				if err.Error() != testcase.expect {
					t.Fatalf("Different error for query %s\nExpect:\n%s\nGot:\n%s", testcase.input, testcase.expect, err.Error())
				}
			}
		})
	}

}
