package cql

import (
	"strings"
	"unicode/utf8"
)

type token int

const (
	tokenEos token = iota
	tokenRelOp
	tokenAnd
	tokenOr
	tokenNot
	tokenProx
	tokenSimpleString
	tokenPrefixName
	tokenSortby
	tokenModifier
	tokenLp
	tokenRp
	tokenError
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
		l.pos += 1
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
		return tokenEos, ""
	case utf8.RuneError:
		l.ch = l.next()
		return tokenError, ""
	case '=':
		l.ch = l.next()
		if l.ch == '=' {
			l.ch = l.next()
			return tokenRelOp, "=="
		}
		return tokenRelOp, "="
	case '<':
		l.ch = l.next()
		if l.ch == '=' {
			l.ch = l.next()
			return tokenRelOp, "<="
		}
		if l.ch == '>' {
			l.ch = l.next()
			return tokenRelOp, "<>"
		}
		return tokenRelOp, "<"
	case '>':
		l.ch = l.next()
		if l.ch == '=' {
			l.ch = l.next()
			return tokenRelOp, ">="
		}
		return tokenRelOp, ">"
	case '/':
		l.ch = l.next()
		return tokenModifier, "/"
	case '(':
		l.ch = l.next()
		return tokenLp, "("
	case ')':
		l.ch = l.next()
		return tokenRp, ")"
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
		return tokenSimpleString, sb.String()
	default:
		var sb strings.Builder
		var isRelation bool = false
		for l.ch != 0 && l.ch != utf8.RuneError {
			if strings.ContainsRune(" \n()=<>/", l.ch) {
				break
			}
			if l.ch == '.' {
				isRelation = true
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
		if strings.EqualFold(value, "and") {
			return tokenAnd, value
		}
		if strings.EqualFold(value, "or") {
			return tokenOr, value
		}
		if strings.EqualFold(value, "not") {
			return tokenNot, value
		}
		if strings.EqualFold(value, "prox") {
			return tokenProx, value
		}
		if strings.EqualFold(value, "sortby") {
			return tokenSortby, value
		}
		if strings.EqualFold(value, "all") ||
			strings.EqualFold(value, "any") ||
			strings.EqualFold(value, "adj") ||
			strings.EqualFold(value, "scr") ||
			strings.EqualFold(value, "exact") ||
			strings.EqualFold(value, "within") ||
			strings.EqualFold(value, "encloses") {
			isRelation = true
		}
		if isRelation {
			return tokenPrefixName, value
		}
		return tokenSimpleString, value
	}
}

func (l *lexer) init(input string) {
	l.input = input
	l.pos = 0
	l.ch = l.next()
}
