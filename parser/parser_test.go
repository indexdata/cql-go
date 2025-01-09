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
			expect: `<searchClause>
  <index>cql.serverChoice<index>
  <relation>
    <value>=</value>
  </relation>
  <term>myterm<term>
</searchClause>
`,
		},
		{
			name:   "term rel value",
			input:  "dc.title all andersen",
			strict: false,
			ok:     true,
			expect: `<searchClause>
  <index>dc.title<index>
  <relation>
    <value>all</value>
  </relation>
  <term>andersen<term>
</searchClause>
`,
		},
		{
			name:   "relation modifiers",
			input:  "dc.title =/k1=v1/k2 andersen",
			strict: false,
			ok:     true,
			expect: `<searchClause>
  <index>dc.title<index>
  <relation>
    <value>=</value>
  </relation>
  <modifiers>
    <modifier>
      <type>k1<type>
      <comparison>=</comparison>
      <value>v1</value>
    </modifier>
    <modifier>
      <type>k2<type>
    </modifier>
  </modifiers>
  <term>andersen<term>
</searchClause>
`,
		},
		{
			name:   "booleans",
			input:  "year < 1990 and b",
			strict: false,
			ok:     true,
			expect: `<triple>
  <boolean>
    <value>and</value>
  </boolean>
  <leftOperand>
    <searchClause>
      <index>year<index>
      <relation>
        <value>&lt;</value>
      </relation>
      <term>1990<term>
    </searchClause>
  </leftOperand>
  <rightOperand>
    <searchClause>
      <index>cql.serverChoice<index>
      <relation>
        <value>=</value>
      </relation>
      <term>b<term>
    </searchClause>
  </rightOperand>
</triple>
`,
		},
		{
			name:   "prox",
			input:  "(a prox/order=1 (b))",
			strict: false,
			ok:     true,
			expect: `<triple>
  <boolean>
    <value>prox</value>
    <modifiers>
      <modifier>
        <type>order<type>
        <comparison>=</comparison>
        <value>1</value>
      </modifier>
    </modifiers>
  </boolean>
  <leftOperand>
    <searchClause>
      <index>cql.serverChoice<index>
      <relation>
        <value>=</value>
      </relation>
      <term>a<term>
    </searchClause>
  </leftOperand>
  <rightOperand>
    <searchClause>
      <index>cql.serverChoice<index>
      <relation>
        <value>=</value>
      </relation>
      <term>b<term>
    </searchClause>
  </rightOperand>
</triple>
`,
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			node, err := p.Parse(testcase.input)
			if testcase.ok {
				if err != nil {
					t.Fatalf("expected OK for query %s", testcase.input)
				}
				xml := node.ToXml()
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
