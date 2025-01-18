package cql

import (
	"testing"
)

func TestParseXml(t *testing.T) {
	var p Parser

	for _, testcase := range []struct {
		name   string
		input  string
		strict bool
		tab    int
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
			name:   "single term",
			input:  "\"<&>\"",
			strict: false,
			tab:    2,
			ok:     true,
			expect: `<xcql xmlns="http://docs.oasis-open.org/ns/search-ws/xcql">
  <triple>
    <searchClause>
      <index>cql.serverChoice</index>
      <relation>
        <value>=</value>
      </relation>
      <term>&lt;&amp;&gt;</term>
    </searchClause>
  </triple>
</xcql>
`,
		},
		{
			name:   "empty term",
			input:  "\"\"",
			strict: false,
			tab:    2,
			ok:     true,
			expect: `<xcql xmlns="http://docs.oasis-open.org/ns/search-ws/xcql">
  <triple>
    <searchClause>
      <index>cql.serverChoice</index>
      <relation>
        <value>=</value>
      </relation>
      <term></term>
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
			name:   "term namedrelation value",
			input:  "dc.title cql.exact andersen",
			strict: false,
			ok:     true,
			expect: `<xcql xmlns="http://docs.oasis-open.org/ns/search-ws/xcql">
<triple>
<searchClause>
<index>dc.title</index>
<relation>
<value>cql.exact</value>
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
			name:   "booleans1",
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
			name:   "booleans2",
			input:  "year > 1990 or b not c",
			strict: false,
			ok:     true,
			expect: `<xcql xmlns="http://docs.oasis-open.org/ns/search-ws/xcql">
<triple>
<Boolean>
<value>not</value>
</Boolean>
<leftOperand>
<triple>
<Boolean>
<value>or</value>
</Boolean>
<leftOperand>
<searchClause>
<index>year</index>
<relation>
<value>&gt;</value>
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
</leftOperand>
<rightOperand>
<searchClause>
<index>cql.serverChoice</index>
<relation>
<value>=</value>
</relation>
<term>c</term>
</searchClause>
</rightOperand>
</triple>
</xcql>
`,
		},
		{
			name:   "prox",
			input:  "(a prox/order=1/default=\"\" (b))",
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
<modifier>
<type>default</type>
<comparison>=</comparison>
<value></value>
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
		{
			name:   "bad rune",
			input:  string([]byte{65, 192, 32, 65}),
			strict: false,
			ok:     false,
			expect: "EOF expected near pos 3",
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			node, err := p.Parse(testcase.input)
			if testcase.ok {
				if err != nil {
					t.Fatalf("expected OK for query %s . Got error: %s", testcase.input, err)
				}
				var xcql Xcql
				xml := xcql.Marshal(node, testcase.tab)
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

func TestQueryString(t *testing.T) {
	in := "> dc = \"http://deepcustard.org/\" dc.title any \"\" or (dc.creator =/x = y sanderson and dc.identifier = id:1234567) sortBy dc.date/sort.descending/special = 1 dc.title/sort.ascending"
	var p Parser
	q, err := p.Parse(in)
	if err != nil {
		t.Fatalf("parse error: %s", err)
	}
	out := q.String()
	if in != out {
		t.Fatalf("expected:\n%s\nwas:\n%s", in, out)
	}
	in = "a b c"
	p.strict = true
	q, err = p.Parse(in)
	if err != nil {
		t.Fatalf("parse error: %s", err)
	}
	out = q.String()
	if in != out {
		t.Fatalf("expected: %s, was: %s", in, out)
	}
}

func TestSortString(t *testing.T) {
	sort := Sort{Index: "title", Modifiers: []Modifier{{Name: "case"}}}
	in := "title/case"
	out := sort.String()
	if in != out {
		t.Fatalf("expected:\n%s\nwas:\n%s", in, out)
	}
}

func TestModifierString(t *testing.T) {
	mod := Modifier{Name: "case", Relation: "=", Value: "true"}
	in := "case = true"
	out := mod.String()
	if in != out {
		t.Fatalf("expected:\n%s\nwas:\n%s", in, out)
	}
}

func TestClauseString(t *testing.T) {
	searchClause := SearchClause{Term: "y"} // in reality there would always be Index + Relation
	clause := Clause{SearchClause: &searchClause}
	in := "y"
	out := clause.String()
	if in != out {
		t.Fatalf("expected:\n%s\nwas:\n%s", in, out)
	}
}
