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
	input string
	pos   int
	ch    rune
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
		var value strings.Builder
		for l.ch != 0 && l.ch != utf8.RuneError {
			if l.ch == '"' {
				l.ch = l.next()
				break
			}
			value.WriteRune(l.ch)
			if l.ch == '\\' {
				l.ch = l.next()
				if l.ch != 0 && l.ch != utf8.RuneError {
					value.WriteRune(l.ch)
				}
			}
			l.ch = l.next()
		}
		return token_simple_string, value.String()
	default:
		var value strings.Builder
		var relation_like bool = false
		for l.ch != 0 && l.ch != utf8.RuneError {
			if strings.ContainsRune(" \n()=<>/", l.ch) {
				break
			}
			if l.ch == '.' {
				relation_like = true
			}
			value.WriteRune(l.ch)
			if l.ch == '\\' {
				l.ch = l.next()
				if l.ch != 0 && l.ch != utf8.RuneError {
					value.WriteRune(l.ch)
				}
			}
			l.ch = l.next()
		}
		if relation_like {
			return token_prefix_name, value.String()
		}
		return token_simple_string, value.String()
	}
}

func (l *lexer) init(input string) {
	l.input = input
	l.pos = 0
	l.ch = l.next()
}
