package cql

import (
	"testing"
)

type tokenResult struct {
	token token
	value string
}

func TestLexer(t *testing.T) {
	for _, testcase := range []struct {
		name     string
		input    string
		strict   bool
		expected []tokenResult
	}{
		{
			name:   "relop",
			input:  "= == > >= < <= <>",
			strict: false,
			expected: []tokenResult{
				{token: tokenRelOp, value: "="},
				{token: tokenRelOp, value: "=="},
				{token: tokenRelOp, value: ">"},
				{token: tokenRelOp, value: ">="},
				{token: tokenRelOp, value: "<"},
				{token: tokenRelOp, value: "<="},
				{token: tokenRelOp, value: "<>"},
				{token: tokenEos, value: ""},
			},
		},
		{
			name:   "otherops",
			input:  "()/",
			strict: false,
			expected: []tokenResult{
				{token: tokenLp, value: "("},
				{token: tokenRp, value: ")"},
				{token: tokenModifier, value: "/"},
				{token: tokenEos, value: ""},
			},
		},
		{
			name:   "qstrings1",
			input:  "\"\" \"x\\\"y\"(",
			strict: false,
			expected: []tokenResult{
				{token: tokenSimpleString, value: ""},
				{token: tokenSimpleString, value: "x\\\"y"},
				{token: tokenLp, value: "("},
				{token: tokenEos, value: ""},
			},
		},
		{
			name:   "qstrings2",
			input:  "\"", // unterminated quoted string
			strict: false,
			expected: []tokenResult{
				{token: tokenSimpleString, value: ""},
				{token: tokenEos, value: ""},
			},
		},
		{
			name:   "qstrings3",
			input:  "\"\\", // unterminated backslash sequence
			strict: false,
			expected: []tokenResult{
				{token: tokenSimpleString, value: "\\"},
				{token: tokenEos, value: ""},
			},
		},
		{
			name:   "unquoted strings",
			input:  "And oR Not PROX Sortby name x.relation all any adj x\\.y",
			strict: false,
			expected: []tokenResult{
				{token: tokenBoolOp, value: "And"},
				{token: tokenBoolOp, value: "oR"},
				{token: tokenBoolOp, value: "Not"},
				{token: tokenBoolOp, value: "PROX"},
				{token: tokenSortby, value: "Sortby"},
				{token: tokenSimpleString, value: "name"},
				{token: tokenPrefixName, value: "x.relation"},
				{token: tokenPrefixName, value: "all"},
				{token: tokenPrefixName, value: "any"},
				{token: tokenPrefixName, value: "adj"},
				{token: tokenSimpleString, value: "x\\.y"},
				{token: tokenEos, value: ""},
			},
		},
		{
			name:   "unquoted strings",
			input:  "And oR Not PROX Sortby name x.relation all any adj x\\.y",
			strict: true,
			expected: []tokenResult{
				{token: tokenBoolOp, value: "And"},
				{token: tokenBoolOp, value: "oR"},
				{token: tokenBoolOp, value: "Not"},
				{token: tokenBoolOp, value: "PROX"},
				{token: tokenSortby, value: "Sortby"},
				{token: tokenPrefixName, value: "name"},
				{token: tokenPrefixName, value: "x.relation"},
				{token: tokenPrefixName, value: "all"},
				{token: tokenPrefixName, value: "any"},
				{token: tokenPrefixName, value: "adj"},
				{token: tokenPrefixName, value: "x\\.y"},
				{token: tokenEos, value: ""},
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			var l lexer
			l.init(testcase.input, testcase.strict)
			last := false
			for i := range testcase.expected {
				if last {
					t.Fatalf("EOS earlier than expected")
				}
				tok, val := l.lex()
				if tok != testcase.expected[i].token {
					t.Fatalf("token mismatch %v != %v at %v", tok, testcase.expected[i].token, i)
				}
				if val != testcase.expected[i].value {
					t.Fatalf("value mismatch %v != %v at %v", val, testcase.expected[i].value, i)
				}
				last = tok == tokenEos
			}
			if !last {
				t.Fatalf("EOS after expected results")
			}
		})
	}
}
