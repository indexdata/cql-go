package parser

import (
	"testing"
)

type tokenResult struct {
	token token
	value string
}

func TestTokenVals(t *testing.T) {

	if token_eos != 0 {
		t.Errorf("eos != 0")
	}

	if token_rel_op != 1 {
		t.Errorf("relop != 1")
	}
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
				{token: token_rel_op, value: "="},
				{token: token_rel_op, value: "=="},
				{token: token_rel_op, value: ">"},
				{token: token_rel_op, value: ">="},
				{token: token_rel_op, value: "<"},
				{token: token_rel_op, value: "<="},
				{token: token_rel_op, value: "<>"},
				{token: token_eos, value: ""},
			},
		},
		{
			name:  "otherops",
			input: "()/",
			expected: []tokenResult{
				{token: token_lp, value: "("},
				{token: token_rp, value: ")"},
				{token: token_modifier, value: "/"},
				{token: token_eos, value: ""},
			},
		},
		{
			name:  "qstrings1",
			input: "\"\" \"x\\\"y\"(",
			expected: []tokenResult{
				{token: token_simple_string, value: ""},
				{token: token_simple_string, value: "x\\\"y"},
				{token: token_lp, value: "("},
				{token: token_eos, value: ""},
			},
		},
		{
			name:  "qstrings2",
			input: "\"", // unterminated quoted string
			expected: []tokenResult{
				{token: token_simple_string, value: ""},
				{token: token_eos, value: ""},
			},
		},
		{
			name:  "qstrings3",
			input: "\"\\", // unterminated backslash sequence
			expected: []tokenResult{
				{token: token_simple_string, value: "\\"},
				{token: token_eos, value: ""},
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
				last = tok == token_eos
			}
			if !last {
				t.Fatalf("EOS after expected results")
			}
		})
	}
}
