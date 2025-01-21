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
		expected []tokenResult
	}{
		{
			name:  "relop",
			input: "= == > >= < <= <>",
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
			name:  "otherops",
			input: "()/",
			expected: []tokenResult{
				{token: tokenLp, value: "("},
				{token: tokenRp, value: ")"},
				{token: tokenModifier, value: "/"},
				{token: tokenEos, value: ""},
			},
		},
		{
			name:  "qstrings1",
			input: "\"\" \"x\\\"y\"(",
			expected: []tokenResult{
				{token: tokenSimpleString, value: ""},
				{token: tokenSimpleString, value: "x\\\"y"},
				{token: tokenLp, value: "("},
				{token: tokenEos, value: ""},
			},
		},
		{
			name:  "qstrings2",
			input: "\"", // unterminated quoted string
			expected: []tokenResult{
				{token: tokenSimpleString, value: ""},
				{token: tokenEos, value: ""},
			},
		},
		{
			name:  "qstrings3",
			input: "\"\\", // unterminated backslash sequence
			expected: []tokenResult{
				{token: tokenSimpleString, value: "\\"},
				{token: tokenEos, value: ""},
			},
		},
		{
			name:  "unquoted strings",
			input: "And oR Not PROX Sortby name x.relation all any adj x\\.y",
			expected: []tokenResult{
				{token: tokenAnd, value: "And"},
				{token: tokenOr, value: "oR"},
				{token: tokenNot, value: "Not"},
				{token: tokenProx, value: "PROX"},
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
			name:  "bad rune",
			input: string([]byte{60, 192, 32, 65}),
			expected: []tokenResult{
				{token: tokenRelOp, value: "<"},
				{token: tokenError, value: ""},
				{token: tokenSimpleString, value: "A"},
				{token: tokenEos, value: ""},
			},
		},
	} {
		t.Run(testcase.name, func(t *testing.T) {
			var l lexer
			l.init(testcase.input)
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
