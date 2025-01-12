package lsl

import (
	"fmt"
	"slices"
)

const (
	// TokenTypeEOF represents the end of the input.
	TokenTypeEOF TokenType = iota

	// TokenTypeError represents an error during tokenization.
	TokenTypeError

	// TokenTypeNewLine represents a new line.
	TokenTypeNewLine

	// TokenTypeComment represents a comment.
	TokenTypeComment

	// TokenTypeIdentifier represents an identifier (e.g. variable name,
	// type name, field, function name, etc).
	TokenTypeIdentifier

	// TokenTypeOperator represents an operator (e.g. assignment, braces, etc).
	TokenTypeOperator

	// TokenTypeNumber represents a numeric value.
	TokenTypeNumber
)

// TokenType represents the type of a token.
type TokenType uint8

// String returns a string representation of the token type.
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
	default:
		return "Unknown"
	}
}

// Token represents a single item of interest in the LSL source code. A
// Tokenizer will convert a string of LSL source code into a sequence of tokens.
type Token struct {

	// Type is the type of token.
	Type TokenType

	// Value is the value of the token.
	Value string

	// Pos is the position of the token in the source code.
	Pos Position
}

// String returns a string representation of the token.
func (t Token) String() string {
	return fmt.Sprintf("%s(%s)", t.Type, t.Value)
}

// IsEOF returns true if the token is an EOF token.
func (t Token) IsEOF() bool {
	return t.Type == TokenTypeEOF
}

// IsError returns true if the token is an error token.
func (t Token) IsError() bool {
	return t.Type == TokenTypeError
}

// IsTerminal returns true if the token is a final token and no subsequent
// tokens will be returned.
func (t Token) IsTerminal() bool {
	return t.IsEOF() || t.IsError()
}

// IsNewLine returns true if the token is a new line token.
func (t Token) IsNewLine() bool {
	return t.Type == TokenTypeNewLine
}

// IsComment returns true if the token is a comment token.
func (t Token) IsComment() bool {
	return t.Type == TokenTypeComment
}

// IsIdentifier returns true if the token is an identifier token.
func (t Token) IsIdentifier() bool {
	return t.Type == TokenTypeIdentifier
}

// IsSpecificIdentifier returns true if the token is an identifier token with
// the specified value.
func (t Token) IsSpecificIdentifier(value string) bool {
	return t.Type == TokenTypeIdentifier && t.Value == value
}

// IsOperator returns true if the token is an operator token.
func (t Token) IsOperator() bool {
	return t.Type == TokenTypeOperator
}

// IsSpecificOperator returns true if the token is an operator token with the
// specified value.
func (t Token) IsSpecificOperator(value string) bool {
	return t.Type == TokenTypeOperator && t.Value == value
}

// IsAssignmentOperator returns true if the token is an assignment operator
// token.
func (t Token) IsAssignmentOperator() bool {
	return t.IsOperator() && slices.Contains(assignmentOperators, t.Value)
}

// IsUnaryOperator returns true if the token is a unary operator token.
func (t Token) IsUnaryOperator() bool {
	return t.IsOperator() && slices.Contains(unaryOperators, t.Value)
}

// IsBinaryOperator returns true if the token is a binary operator token.
func (t Token) IsBinaryOperator() bool {
	return t.IsOperator() && slices.Contains(binaryOperators, t.Value)
}

// IsNumber returns true if the token is a numeric token.
func (t Token) IsNumber() bool {
	return t.Type == TokenTypeNumber
}
