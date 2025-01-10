package parser

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
<boolean>
<value>and</value>
</boolean>
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
<boolean>
<value>prox</value>
<modifiers>
<modifier>
<type>order</type>
<comparison>=</comparison>
<value>1</value>
</modifier>
</modifiers>
</boolean>
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
	} {
		t.Run(testcase.name, func(t *testing.T) {
			node, err := p.Parse(testcase.input)
			if testcase.ok {
				if err != nil {
					t.Fatalf("expected OK for query %s", testcase.input)
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
			}
		})
	}

}
