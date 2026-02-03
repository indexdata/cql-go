package cql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuote(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty", in: "", want: "\"\""},
		{name: "plain", in: "alpha", want: "alpha"},
		{name: "whitespace", in: "two words", want: "\"two words\""},
		{name: "operator", in: "a=b", want: "\"a=b\""},
		{name: "quote_unquoted", in: "a\"b", want: "a\\\"b"},
		{name: "quote_quoted", in: "a\" b", want: "\"a\\\" b\""},
		{name: "trailing_backslash", in: "abc\\", want: "abc\\\\"},
		{name: "escaped_quote", in: "a\\\"b", want: "a\\\"b"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := &Query{
				Clause: Clause{
					BoolClause: &BoolClause{
						Left: Clause{
							SearchClause: &SearchClause{
								Index: "idx",
								Term:  tt.in,
							},
						},
						Operator: AND,
						Right: Clause{
							SearchClause: &SearchClause{
								Index: "idx2",
								Term:  "val2",
							},
						},
					},
				},
			}
			expected := "idx = " + tt.want + " and idx2 = val2"
			assert.Equal(t, expected, q.String(), "unexpected quoted term")
			parsed, err := (&Parser{}).Parse(expected)
			assert.NoError(t, err, "unexpected parse error for query %q", expected)
			assert.Equal(t, expected, parsed.String(), "unexpected re-serialized query")
		})
	}
}
