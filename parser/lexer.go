package parser

import (
	"strings"
	"unicode/utf8"
)

type token int

const (
	token_eos token = iota
	token_rel_op
	token_bool_op
	token_simple_string
	token_prefix_name
	token_sortby
	token_modifier
	token_lp
	token_rp
	token_error
)

type lexer struct {
	input  string
	pos    int
	ch     rune
	strict bool
}

func (l *lexer) next() rune {
	if l.pos == len(l.input) {
		return 0
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	if r == utf8.RuneError {
		return r
	}
	l.pos += w
	return r
}

func isspace(ch rune) bool {
	return ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n'
}

func (l *lexer) lex() (tok token, value string) {
	for isspace(l.ch) {
		l.ch = l.next()
	}
	switch l.ch {
	case 0:
		return token_eos, ""
	case utf8.RuneError:
		return token_error, ""
	case '=':
		l.ch = l.next()
		if l.ch == '=' {
			l.ch = l.next()
			return token_rel_op, "=="
		}
		return token_rel_op, "="
	case '<':
		l.ch = l.next()
		if l.ch == '=' {
			l.ch = l.next()
			return token_rel_op, "<="
		}
		if l.ch == '>' {
			l.ch = l.next()
			return token_rel_op, "<>"
		}
		return token_rel_op, "<"
	case '>':
		l.ch = l.next()
		if l.ch == '=' {
			l.ch = l.next()
			return token_rel_op, ">="
		}
		return token_rel_op, ">"
	case '/':
		l.ch = l.next()
		return token_modifier, "/"
	case '(':
		l.ch = l.next()
		return token_lp, "("
	case ')':
		l.ch = l.next()
		return token_rp, ")"
	case '"':
		l.ch = l.next()
		var sb strings.Builder
		for l.ch != 0 && l.ch != utf8.RuneError {
			if l.ch == '"' {
				l.ch = l.next()
				break
			}
			sb.WriteRune(l.ch)
			if l.ch == '\\' {
				l.ch = l.next()
				if l.ch != 0 && l.ch != utf8.RuneError {
					sb.WriteRune(l.ch)
				}
			}
			l.ch = l.next()
		}
		return token_simple_string, sb.String()
	default:
		var sb strings.Builder
		var relation_like bool = l.strict
		for l.ch != 0 && l.ch != utf8.RuneError {
			if strings.ContainsRune(" \n()=<>/", l.ch) {
				break
			}
			if l.ch == '.' {
				relation_like = true
			}
			sb.WriteRune(l.ch)
			if l.ch == '\\' {
				l.ch = l.next()
				if l.ch != 0 && l.ch != utf8.RuneError {
					sb.WriteRune(l.ch)
				}
			}
			l.ch = l.next()
		}
		value := sb.String()
		if strings.EqualFold(value, "and") ||
			strings.EqualFold(value, "or") ||
			strings.EqualFold(value, "not") ||
			strings.EqualFold(value, "prox") {
			return token_bool_op, value
		}
		if strings.EqualFold(value, "sortby") {
			return token_sortby, value
		}
		if strings.EqualFold(value, "all") ||
			strings.EqualFold(value, "any") ||
			strings.EqualFold(value, "adj") {
			relation_like = true
		}
		if relation_like {
			return token_prefix_name, value
		}
		return token_simple_string, value
	}
}

func (l *lexer) init(input string, strict bool) {
	l.strict = strict
	l.input = input
	l.pos = 0
	l.ch = l.next()
}
