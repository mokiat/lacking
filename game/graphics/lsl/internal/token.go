package internal

import (
	"slices"
	"unicode"
)

var (
	operatorChars = []rune{
		'{', '}', '=',
	}
)

type TokenType uint8

const (
	TokenTypeOEF TokenType = iota
	TokenTypeNewLine
	TokenTypeWhitespace
	TokenTypeComment
	TokenTypeIdentifier
	TokenTypeOperator
)

type Token struct {
	Type  TokenType
	Value string
}

func IsWhitespace(ch rune) bool {
	return ch == ' ' || ch == '\t'
}

func IsNewLine(ch rune) bool {
	return ch == '\n' || ch == '\r'
}

func IsIdentifier(p int, ch rune) bool {
	return unicode.IsLetter(ch) || (p > 0 && unicode.IsDigit(ch)) || (p == 0 && ch == '#')
}

func IsOperator(ch rune) bool {
	return slices.Contains(operatorChars, ch)
}
