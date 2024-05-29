package internal

import (
	"slices"
	"unicode"
)

func IsWhitespaceChar(ch rune) bool {
	return ch == ' ' || ch == '\t'
}

func IsNewLineChar(ch rune) bool {
	return ch == '\n' || ch == '\r'
}

func IsIdentifierChar(ch rune, pos int) bool {
	return unicode.IsLetter(ch) || (pos > 0 && unicode.IsDigit(ch)) || (pos == 0 && ch == '#')
}

func IsOperatorChar(ch rune) bool {
	return slices.Contains(operatorChars, ch)
}

func IsNumberChar(ch rune, pos int) bool {
	return unicode.IsDigit(ch) || (pos > 0 && ch == '.')
}
