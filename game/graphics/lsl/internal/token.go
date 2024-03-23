package internal

import (
	"fmt"
	"slices"
	"unicode"
)

var (
	operatorChars = []rune{
		'{', '}', '=', '(', ')', ',', '-', ';', '+', '*', '/', '%', '!', '<', '>', '&', '|', '^', '.',
	}
)

type TokenType uint8

func (t TokenType) String() string {
	switch t {
	case TokenTypeOEF:
		return "EOF"
	case TokenTypeNewLine:
		return "NewLine"
	case TokenTypeComment:
		return "Comment"
	case TokenTypeIdentifier:
		return "Identifier"
	case TokenTypeOperator:
		return "Operator"
	case TokenTypeNumber:
		return "Number"
	}
	return "Unknown"
}

const (
	TokenTypeOEF TokenType = iota
	TokenTypeNewLine
	TokenTypeComment
	TokenTypeIdentifier
	TokenTypeOperator
	TokenTypeNumber
)

type Token struct {
	Type  TokenType
	Value string
}

func (t Token) String() string {
	return fmt.Sprintf("%s(%s)", t.Type, t.Value)
}

func (t Token) IsEOF() bool {
	return t.Type == TokenTypeOEF
}

func (t Token) IsNewLine() bool {
	return t.Type == TokenTypeNewLine
}

func (t Token) IsComment() bool {
	return t.Type == TokenTypeComment
}

func (t Token) IsIdentifier() bool {
	return t.Type == TokenTypeIdentifier
}

func (t Token) IsSpecificIdentifier(value string) bool {
	return t.Type == TokenTypeIdentifier && t.Value == value
}

func (t Token) IsOperator() bool {
	return t.Type == TokenTypeOperator
}

func (t Token) IsSpecificOperator(value string) bool {
	return t.Type == TokenTypeOperator && t.Value == value
}

func (t Token) IsNumber() bool {
	return t.Type == TokenTypeNumber
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

func IsNumber(p int, ch rune) bool {
	return unicode.IsDigit(ch) || (p > 0 && ch == '.')
}
