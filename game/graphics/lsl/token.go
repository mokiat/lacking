package lsl

import (
	"fmt"
	"slices"
)

const (
	TokenTypeEOF TokenType = iota
	TokenTypeNewLine
	TokenTypeComment
	TokenTypeIdentifier
	TokenTypeOperator
	TokenTypeNumber
)

type TokenType uint8

func (t TokenType) String() string {
	switch t {
	case TokenTypeEOF:
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

type Token struct {
	Type  TokenType
	Value string
}

func (t Token) String() string {
	return fmt.Sprintf("%s(%s)", t.Type, t.Value)
}

func (t Token) IsEOF() bool {
	return t.Type == TokenTypeEOF
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

func (t Token) IsAssignmentOperator() bool {
	return t.IsOperator() && slices.Contains(assignmentOperators, t.Value)
}

func (t Token) IsUnaryOperator() bool {
	return t.IsOperator() && slices.Contains(unaryOperators, t.Value)
}

func (t Token) IsBinaryOperator() bool {
	return t.IsOperator() && slices.Contains(binaryOperators, t.Value)
}

func (t Token) IsSpecificOperator(value string) bool {
	return t.Type == TokenTypeOperator && t.Value == value
}

func (t Token) IsNumber() bool {
	return t.Type == TokenTypeNumber
}
