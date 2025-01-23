package cql

import (
	"testing"
)

func TestParseXml(t *testing.T) {
	var p Parser

	for _, testcase := range []struct {
		name   string
		input  string
		tab    int
		ok     bool
		expect string
	}{
		{
			name:  "single term",
			input: "myterm",
			ok:    true,
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
			name:  "single term",
			input: "\"<&>\"",
			tab:   2,
			ok:    true,
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
			name:  "empty term",
			input: "\"\"",
			tab:   2,
			ok:    true,
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
			name:  "term rel value",
			input: "dc.title all andersen",
			ok:    true,
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
			name:  "term namedrelation value",
			input: "dc.title cql.exact andersen",
			ok:    true,
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
			name:  "relation modifiers",
			input: "dc.title =/k1=v1/k2 andersen",
			ok:    true,
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
			name:  "booleans1",
			input: "year < 1990 and b",
			ok:    true,
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
			name:  "booleans2",
			input: "year > 1990 or b not c",
			ok:    true,
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
			name:  "prox",
			input: "(a prox/order=1/default=\"\" (b))",
			ok:    true,
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
			name:  "sort",
			input: "myterm1 sortby title",
			ok:    true,
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
			name:  "sort2",
			input: "myterm1 sortby title year/asc",
			ok:    true,
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
			name:  "sort3",
			input: "ti=a and b sortby title",
			ok:    true,
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
			name:  "prefix1",
			input: ">dc = uri dc.ti = a",
			ok:    true,
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
			name:  "prefix2",
			input: ">a =uri1>uri2>b=uri3 dc.ti = a",
			ok:    true,
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
			name:  "prefix3",
			input: "a and (>dc=uri dc.ti = a)",
			ok:    true,
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
			ok:     false,
			expect: "search term expected near pos 0",
		},
		{
			name:   "(",
			input:  "(",
			ok:     false,
			expect: "search term expected near pos 1",
		},
		{
			name:   "(a)",
			input:  "(a",
			ok:     false,
			expect: "missing ) near pos 2",
		},
		{
			name:   "(a))",
			input:  "(a))",
			ok:     false,
			expect: "EOF expected near pos 4",
		},
		{
			name:   "dc.ti =",
			input:  "dc.ti =",
			ok:     false,
			expect: "search term expected near pos 7",
		},
		{
			name:   "a and",
			input:  "a and",
			ok:     false,
			expect: "search term expected near pos 5",
		},
		{
			name:   "a and /",
			input:  "a and /",
			ok:     false,
			expect: "missing modifier key near pos 7",
		},
		{
			name:   "a =/",
			input:  "a =/",
			ok:     false,
			expect: "missing modifier key near pos 4",
		},
		{
			name:   "a =/",
			input:  "a =/b=",
			ok:     false,
			expect: "missing modifier value near pos 6",
		},
		{
			name:   ">",
			input:  ">",
			ok:     false,
			expect: "term expected after > near pos 1",
		},
		{
			name:   ">dc=()",
			input:  ">dc=()",
			ok:     false,
			expect: "term expected after = near pos 6",
		},
		{
			name:   ">dc=uri",
			input:  ">dc=uri",
			ok:     false,
			expect: "search term expected near pos 7",
		},
		{
			name:   "a sortby year/",
			input:  "a sortby year/",
			ok:     false,
			expect: "missing modifier key near pos 14",
		},
		{
			name:   "bad rune",
			input:  string([]byte{65, 192, 32, 65}),
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

func TestMultiTermAndSymRelStrict(t *testing.T) {
	var p Parser
	p.Strict = true
	in := "a"
	q, err := p.Parse(in)
	if err != nil {
		t.Fatalf("parse error: %s", err)
	}
	out := q.String()
	if in != out {
		t.Fatalf("expected:\n%s\nwas:\n%s", in, out)
	}
	in = "cql.serverChoice scr a"
	q, err = p.Parse(in)
	if err != nil {
		t.Fatalf("parse error: %s", err)
	}
	out = q.String()
	exp := "a"
	if exp != out {
		t.Fatalf("expected:\n%s\nwas:\n%s", exp, out)
	}
	in = "a b"
	q, err = p.Parse(in)
	if err == nil || err.Error() != "EOF expected near pos 3" {
		t.Fatalf("expected parse error but was: %v", err)
	}
	out = q.String()
	if in == out {
		t.Fatalf("expected not equals: %s, %s", in, out)
	}
	in = "a b c"
	q, err = p.Parse(in)
	if err == nil || err.Error() != "EOF expected near pos 4" {
		t.Fatalf("expected parse error but was: %v", err)
	}
	out = q.String()
	if in == out {
		t.Fatalf("expected not equals: %s, %s", in, out)
	}
	in = "a b.c"
	q, err = p.Parse(in)
	if err == nil || err.Error() != "EOF expected near pos 5" {
		t.Fatalf("expected parse error but was: %v", err)
	}
	out = q.String()
	if in == out {
		t.Fatalf("expected not equals: %s, %s", in, out)
	}
	in = "a b.c d"
	q, err = p.Parse(in)
	if err == nil || err.Error() != "EOF expected near pos 6" {
		t.Fatalf("expected parse error but was: %v", err)
	}
	out = q.String()
	if in == out {
		t.Fatalf("expected:\n%s\nwas:\n%s", in, out)
	}
	in = "a within d"
	q, err = p.Parse(in)
	if err != nil {
		t.Fatalf("parse error: %s", err)
	}
	out = q.String()
	if in != out {
		t.Fatalf("expected:\n%s\nwas:\n%s", in, out)
	}
	in = "a b adj"
	q, err = p.Parse(in)
	if err == nil || err.Error() != "EOF expected near pos 4" {
		t.Fatalf("expected parse error but was: %v", err)
	}
	out = q.String()
	if in == out {
		t.Fatalf("expected:\n%s\nwas:\n%s", in, out)
	}
	in = "a b adj c"
	q, err = p.Parse(in)
	if err == nil || err.Error() != "EOF expected near pos 4" {
		t.Fatalf("expected parse error but was: %v", err)
	}
}

func TestMultiTermAndSymRel(t *testing.T) {
	in := "a b"
	var p Parser
	q, err := p.Parse(in)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}
	out := q.String()
	exp := "\"a b\""
	if exp != out {
		t.Fatalf("Expected: %s, got %s", exp, out)
	}
	in = "a b c"
	q, err = p.Parse(in)
	if err != nil {
		t.Fatalf("got error: %v", err)
	}
	out = q.String()
	exp = "\"a b c\""
	if exp != out {
		t.Fatalf("Expected: %s, got %s", exp, out)
	}
	in = "a b.c"
	q, err = p.Parse(in)
	if err != nil {
		t.Fatalf("parse error but was: %v", err)
	}
	out = q.String()
	exp = "\"a b.c\""
	if exp != out {
		t.Fatalf("expected not equals: %s, %s", exp, out)
	}
	in = "a b.c d"
	q, err = p.Parse(in)
	if err != nil {
		t.Fatalf("parse error: %s", err)
	}
	out = q.String()
	exp = "\"a b.c d\""
	if exp != out {
		t.Fatalf("expected not equals: %s, %s", exp, out)
	}
	in = "a within d"
	q, err = p.Parse(in)
	if err != nil {
		t.Fatalf("parse error: %s", err)
	}
	out = q.String()
	if in != out {
		t.Fatalf("expected:\n%s\nwas:\n%s", in, out)
	}
	in = "a b adj"
	q, err = p.Parse(in)
	if err != nil {
		t.Fatalf("parse error: %s", err)
	}
	out = q.String()
	exp = "\"a b adj\""
	if exp != out {
		t.Fatalf("expected:\n%s\nwas:\n%s", exp, out)
	}
	in = "a b adj adj"
	q, err = p.Parse(in)
	if err != nil {
		t.Fatalf("parse error: %s", err)
	}
	out = q.String()
	exp = "\"a b adj adj\""
	if exp != out {
		t.Fatalf("expected:\n%s\nwas:\n%s", exp, out)
	}
}

func TestQueryString(t *testing.T) {
	in := "> dc = \"http://deepcustard.org/\" > \"\" dc.title any/\"\" \"\" or (dc.creator =/x=y sanderson and dc.identifier = id:1234567) sortBy dc.date/sort.descending/special=1 dc.title/sort.ascending"
	var p Parser
	q, err := p.Parse(in)
	if err != nil {
		t.Fatalf("parse error: %s", err)
	}
	out := q.String()
	if in != out {
		t.Fatalf("expected:\n%s\nwas:\n%s", in, out)
	}
}

func TestQueryBrackets(t *testing.T) {
	in := "a = x and b = y or c = z"
	var p Parser
	q, err := p.Parse(in)
	if err != nil {
		t.Fatalf("parse error: %s", err)
	}
	out := q.String()
	if in != out {
		t.Fatalf("expected:\n%s\nwas:\n%s", in, out)
	}
	in = "a = x and (b = y or c = z)"
	q, err = p.Parse(in)
	if err != nil {
		t.Fatalf("parse error: %s", err)
	}
	out = q.String()
	if in != out {
		t.Fatalf("expected: %s, was: %s", in, out)
	}
	in = "(a = x or a = y) and (b = z or b = q)"
	expected := "a = x or a = y and (b = z or b = q)"
	q, err = p.Parse(in)
	if err != nil {
		t.Fatalf("parse error: %s", err)
	}
	out = q.String()
	if out != expected {
		t.Fatalf("expected: %s, was: %s", expected, out)
	}
}

func TestSortString(t *testing.T) {
	sort := Sort{}
	exp := "\"\""
	out := sort.String()
	if exp != out {
		t.Fatalf("expected:\n%s\nwas:\n%s", exp, out)
	}
	sort = Sort{Index: "title"}
	exp = "title"
	out = sort.String()
	if exp != out {
		t.Fatalf("expected:\n%s\nwas:\n%s", exp, out)
	}
	sort = Sort{Index: "title", Modifiers: []Modifier{{Name: "case"}}}
	exp = "title/case"
	out = sort.String()
	if exp != out {
		t.Fatalf("expected:\n%s\nwas:\n%s", exp, out)
	}
}

func TestModifierString(t *testing.T) {
	mod := Modifier{}
	exp := "\"\""
	out := mod.String()
	if exp != out {
		t.Fatalf("expected:\n%s\nwas:\n%s", exp, out)
	}
	mod = Modifier{Name: "case", Value: "true"}
	exp = "case=true"
	out = mod.String()
	if exp != out {
		t.Fatalf("expected:\n%s\nwas:\n%s", exp, out)
	}
	mod = Modifier{Relation: EQ}
	exp = "\"\""
	out = mod.String()
	if exp != out {
		t.Fatalf("expected:\n%s\nwas:\n%s", exp, out)
	}
	mod = Modifier{Value: "true"}
	exp = "\"\"=true"
	out = mod.String()
	if exp != out {
		t.Fatalf("expected:\n%s\nwas:\n%s", exp, out)
	}
	mod = Modifier{Name: "case", Relation: EQ}
	exp = "case"
	out = mod.String()
	if exp != out {
		t.Fatalf("expected:\n%s\nwas:\n%s", exp, out)
	}
	mod = Modifier{Name: "case", Relation: EQ, Value: "true"}
	exp = "case=true"
	out = mod.String()
	if exp != out {
		t.Fatalf("expected:\n%s\nwas:\n%s", exp, out)
	}
}

func TestPrefixString(t *testing.T) {
	px := Prefix{}
	exp := "> \"\""
	out := px.String()
	if exp != out {
		t.Fatalf("expected:\n%s\nwas:\n%s", exp, out)
	}
	px = Prefix{Uri: "http://deepcustard.org/"}
	exp = "> \"http://deepcustard.org/\""
	out = px.String()
	if exp != out {
		t.Fatalf("expected:\n%s\nwas:\n%s", exp, out)
	}
	px = Prefix{Prefix: "dc"}
	exp = "> dc = \"\""
	out = px.String()
	if exp != out {
		t.Fatalf("expected:\n%s\nwas:\n%s", exp, out)
	}
	px = Prefix{Prefix: "dc", Uri: "http://deepcustard.org/"}
	exp = "> dc = \"http://deepcustard.org/\""
	out = px.String()
	if exp != out {
		t.Fatalf("expected:\n%s\nwas:\n%s", exp, out)
	}
}

func TestSearchClauseString(t *testing.T) {
	searchClause := SearchClause{Index: "title", Relation: EQ}
	exp := "title = \"\""
	out := searchClause.String()
	if exp != out {
		t.Fatalf("expected:\n%s\nwas:\n%s", exp, out)
	}
	searchClause = SearchClause{Relation: EQ}
	exp = "\"\""
	out = searchClause.String()
	if exp != out {
		t.Fatalf("expected:\n%s\nwas:\n%s", exp, out)
	}
	searchClause = SearchClause{Relation: NE}
	exp = "cql.serverChoice <> \"\""
	out = searchClause.String()
	if exp != out {
		t.Fatalf("expected:\n%s\nwas:\n%s", exp, out)
	}
	searchClause = SearchClause{Index: "title"}
	exp = "title = \"\""
	out = searchClause.String()
	if exp != out {
		t.Fatalf("expected:\n%s\nwas:\n%s", exp, out)
	}
	searchClause = SearchClause{Term: "lord of the rings"}
	exp = "\"lord of the rings\""
	out = searchClause.String()
	if exp != out {
		t.Fatalf("expected:\n%s\nwas:\n%s", exp, out)
	}
	searchClause = SearchClause{Index: "title", Relation: EQ, Term: "lord of the rings"}
	exp = "title = \"lord of the rings\""
	out = searchClause.String()
	if exp != out {
		t.Fatalf("expected:\n%s\nwas:\n%s", exp, out)
	}
}

func TestBoolClauseString(t *testing.T) {
	clause1 := Clause{SearchClause: &SearchClause{Term: "x"}}
	clause2 := Clause{SearchClause: &SearchClause{Term: "y"}}
	boolClause := BoolClause{Left: clause1, Operator: AND, Right: clause2}
	in := "x and y"
	out := boolClause.String()
	if in != out {
		t.Fatalf("expected:\n%s\nwas:\n%s", in, out)
	}
	boolClause = BoolClause{Left: clause1, Right: clause2}
	in = "x and y"
	out = boolClause.String()
	if in != out {
		t.Fatalf("expected:\n%s\nwas:\n%s", in, out)
	}
	boolClause = BoolClause{Right: clause2}
	in = "cql.allRecords = 1 and y"
	out = boolClause.String()
	if in != out {
		t.Fatalf("expected:\n%s\nwas:\n%s", in, out)
	}
	boolClause = BoolClause{Left: clause2}
	in = "y and cql.allRecords = 1"
	out = boolClause.String()
	if in != out {
		t.Fatalf("expected:\n%s\nwas:\n%s", in, out)
	}
	in = "x"
	out = clause1.String()
	if in != out {
		t.Fatalf("expected:\n%s\nwas:\n%s", in, out)
	}
	in = "y"
	out = clause2.String()
	if in != out {
		t.Fatalf("expected:\n%s\nwas:\n%s", in, out)
	}
}
