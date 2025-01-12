package lsl

import (
	"iter"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
)

// nonUTFOffset is a special offset value that indicates that the source code
// is not valid UTF-8. This is used to avoid checking for UTF-8 validity in
// every iteration of the tokenizer.
const nonUTFOffset = -1

// Tokenize is a helper function that creates a new iterator that yields tokens
// from the given source code. Internally it creates a new Tokenizer and uses
// it to generate tokens.
func Tokenize(source string) iter.Seq[Token] {
	return func(yield func(Token) bool) {
		tokenizer := NewTokenizer(source)
		for {
			token := tokenizer.Next()
			if token.IsTerminal() {
				return
			}
			if !yield(token) {
				return
			}
		}
	}
}

// NewTokenizer creates a new Tokenizer for the given source code. The source
// code is expected to be a valid UTF-8 string. If the source code is not a
// valid UTF-8 string, the tokenizer will return an error token.
func NewTokenizer(source string) *Tokenizer {
	offset := 0
	if !utf8.ValidString(source) {
		offset = nonUTFOffset
	}
	return &Tokenizer{
		source: source,
		offset: offset,
	}
}

// Tokenizer is a mechanism to split an LSL source code into key pieces of
// information, called tokens. Each token represents a single element of the
// source code, such as an identifier, a number, an operator, etc.
//
// One would normally use the Parser to process LSL source code, since the
// Tokenizer provides low-level information about the source code.
type Tokenizer struct {
	source string
	offset int
}

// Next returns the next token in the source code. If there are no more tokens
// to return, it returns a token with the type TokenTypeEOF.
func (t *Tokenizer) Next() Token {
	if t.offset == nonUTFOffset {
		return Token{
			Type:  TokenTypeError,
			Value: "source code is not a valid UTF-8 string",
		}
	}
	for t.offset < len(t.source) {
		switch {
		case t.nextIsWhitespace():
			t.scanWhitespace()
		case t.nextIsNewLine():
			return t.scanNewLine()
		case t.nextIsComment():
			return t.scanComment()
		case t.nextIsIdentifier():
			return t.scanIdentifier()
		case t.nextIsOperator():
			return t.scanOperator()
		case t.nextIsNumber():
			return t.scanNumber()
		default:
			return t.scanUnknown()
		}
	}
	return Token{
		Type: TokenTypeEOF,
	}
}

func (t *Tokenizer) nextIsWhitespace() bool {
	ch, _ := t.peekRune(t.offset)
	return isWhitespaceChar(ch)
}

func (t *Tokenizer) scanWhitespace() {
	for t.offset < len(t.source) {
		ch, newOffset := t.peekRune(t.offset)
		if !isWhitespaceChar(ch) {
			break
		}
		t.offset = newOffset
	}
}

func (t *Tokenizer) nextIsNewLine() bool {
	ch, _ := t.peekRune(t.offset)
	return isNewLineChar(ch)
}

func (t *Tokenizer) scanNewLine() Token {
	start := t.offset
	for t.offset < len(t.source) {
		ch, newOffset := t.peekRune(t.offset)
		if !isNewLineChar(ch) {
			break
		}
		t.offset = newOffset
	}
	return Token{
		Type:  TokenTypeNewLine,
		Value: t.source[start:t.offset],
	}
}

func (t *Tokenizer) nextIsComment() bool {
	ch1, _ := t.peekRune(t.offset)
	ch2, _ := t.peekRune(t.offset + 1)
	return ch1 == '/' && ch2 == '/'
}

func (t *Tokenizer) scanComment() Token {
	_, t.offset = t.peekRune(t.offset) // Skip first '/'
	_, t.offset = t.peekRune(t.offset) // Skip second '/'
	start := t.offset
	for t.offset < len(t.source) {
		ch, newOffset := t.peekRune(t.offset)
		if isNewLineChar(ch) {
			break
		}
		t.offset = newOffset
	}
	return Token{
		Type:  TokenTypeComment,
		Value: strings.TrimSpace(t.source[start:t.offset]),
	}
}

func (t *Tokenizer) nextIsIdentifier() bool {
	ch, _ := t.peekRune(t.offset)
	return isIdentifierChar(ch, 0)
}

func (t *Tokenizer) scanIdentifier() Token {
	start := t.offset
	charIndex := 0
	for t.offset < len(t.source) {
		ch, newOffset := t.peekRune(t.offset)
		if !isIdentifierChar(ch, charIndex) {
			break
		}
		t.offset = newOffset
		charIndex++
	}
	return Token{
		Type:  TokenTypeIdentifier,
		Value: t.source[start:t.offset],
	}
}

func (t *Tokenizer) nextIsOperator() bool {
	ch, _ := t.peekRune(t.offset)
	return isOperatorChar(ch)
}

func (t *Tokenizer) scanOperator() Token {
	ch1, ch1Offset := t.peekRune(t.offset)
	ch2, ch2Offset := t.peekRune(ch1Offset)
	ch3, ch3Offset := t.peekRune(ch2Offset)
	switch {
	case ch1 == '>' && ch2 == '>' && ch3 == '=':
		t.offset = ch3Offset
		return Token{
			Type:  TokenTypeOperator,
			Value: ">>=",
		}
	case ch1 == '<' && ch2 == '<' && ch3 == '=':
		t.offset = ch3Offset
		return Token{
			Type:  TokenTypeOperator,
			Value: "<<=",
		}
	case ch1 == '=' && ch2 == '=':
		t.offset = ch2Offset
		return Token{
			Type:  TokenTypeOperator,
			Value: "==",
		}
	case ch1 == '+' && ch2 == '=':
		t.offset = ch2Offset
		return Token{
			Type:  TokenTypeOperator,
			Value: "+=",
		}
	case ch1 == '-' && ch2 == '=':
		t.offset = ch2Offset
		return Token{
			Type:  TokenTypeOperator,
			Value: "-=",
		}
	case ch1 == '*' && ch2 == '=':
		t.offset = ch2Offset
		return Token{
			Type:  TokenTypeOperator,
			Value: "*=",
		}
	case ch1 == '/' && ch2 == '=':
		t.offset = ch2Offset
		return Token{
			Type:  TokenTypeOperator,
			Value: "/=",
		}
	case ch1 == '%' && ch2 == '=':
		t.offset = ch2Offset
		return Token{
			Type:  TokenTypeOperator,
			Value: "%=",
		}
	case ch1 == '&' && ch2 == '=':
		t.offset = ch2Offset
		return Token{
			Type:  TokenTypeOperator,
			Value: "&=",
		}
	case ch1 == '^' && ch2 == '=':
		t.offset = ch2Offset
		return Token{
			Type:  TokenTypeOperator,
			Value: "^=",
		}
	case ch1 == '|' && ch2 == '=':
		t.offset = ch2Offset
		return Token{
			Type:  TokenTypeOperator,
			Value: "|=",
		}
	case ch1 == '<' && ch2 == '<':
		t.offset = ch2Offset
		return Token{
			Type:  TokenTypeOperator,
			Value: "<<",
		}
	case ch1 == '>' && ch2 == '>':
		t.offset = ch2Offset
		return Token{
			Type:  TokenTypeOperator,
			Value: ">>",
		}
	case ch1 == '!' && ch2 == '=':
		t.offset = ch2Offset
		return Token{
			Type:  TokenTypeOperator,
			Value: "!=",
		}
	case ch1 == '>' && ch2 == '=':
		t.offset = ch2Offset
		return Token{
			Type:  TokenTypeOperator,
			Value: ">=",
		}
	case ch1 == '<' && ch2 == '=':
		t.offset = ch2Offset
		return Token{
			Type:  TokenTypeOperator,
			Value: "<=",
		}
	case ch1 == '&' && ch2 == '&':
		t.offset = ch2Offset
		return Token{
			Type:  TokenTypeOperator,
			Value: "&&",
		}
	case ch1 == '|' && ch2 == '|':
		t.offset = ch2Offset
		return Token{
			Type:  TokenTypeOperator,
			Value: "||",
		}
	case ch1 == '(':
		t.offset = ch1Offset
		return Token{
			Type:  TokenTypeOperator,
			Value: "(",
		}
	case ch1 == ')':
		t.offset = ch1Offset
		return Token{
			Type:  TokenTypeOperator,
			Value: ")",
		}
	case ch1 == ',':
		t.offset = ch1Offset
		return Token{
			Type:  TokenTypeOperator,
			Value: ",",
		}
	case ch1 == '{':
		t.offset = ch1Offset
		return Token{
			Type:  TokenTypeOperator,
			Value: "{",
		}
	case ch1 == '}':
		t.offset = ch1Offset
		return Token{
			Type:  TokenTypeOperator,
			Value: "}",
		}
	case ch1 == '=':
		t.offset = ch1Offset
		return Token{
			Type:  TokenTypeOperator,
			Value: "=",
		}
	case ch1 == '+':
		t.offset = ch1Offset
		return Token{
			Type:  TokenTypeOperator,
			Value: "+",
		}
	case ch1 == '-':
		t.offset = ch1Offset
		return Token{
			Type:  TokenTypeOperator,
			Value: "-",
		}
	case ch1 == '.':
		t.offset = ch1Offset
		return Token{
			Type:  TokenTypeOperator,
			Value: ".",
		}
	case ch1 == '*':
		t.offset = ch1Offset
		return Token{
			Type:  TokenTypeOperator,
			Value: "*",
		}
	case ch1 == '/':
		t.offset = ch1Offset
		return Token{
			Type:  TokenTypeOperator,
			Value: "/",
		}
	case ch1 == '%':
		t.offset = ch1Offset
		return Token{
			Type:  TokenTypeOperator,
			Value: "%",
		}
	case ch1 == '^':
		t.offset = ch1Offset
		return Token{
			Type:  TokenTypeOperator,
			Value: "^",
		}
	case ch1 == '|':
		t.offset = ch1Offset
		return Token{
			Type:  TokenTypeOperator,
			Value: "|",
		}
	case ch1 == '&':
		t.offset = ch1Offset
		return Token{
			Type:  TokenTypeOperator,
			Value: "&",
		}
	case ch1 == '<':
		t.offset = ch1Offset
		return Token{
			Type:  TokenTypeOperator,
			Value: "<",
		}
	case ch1 == '>':
		t.offset = ch1Offset
		return Token{
			Type:  TokenTypeOperator,
			Value: ">",
		}
	case ch1 == '!':
		t.offset = ch1Offset
		return Token{
			Type:  TokenTypeOperator,
			Value: "!",
		}
	default:
		return Token{}
	}
}

func (t *Tokenizer) nextIsNumber() bool {
	ch, _ := t.peekRune(t.offset)
	return isNumberChar(ch, 0)
}

func (t *Tokenizer) scanNumber() Token {
	start := t.offset
	charIndex := 0
	for t.offset < len(t.source) {
		ch, newOffset := t.peekRune(t.offset)
		if !isNumberChar(ch, charIndex) {
			break
		}
		t.offset = newOffset
		charIndex++
	}
	return Token{
		Type:  TokenTypeNumber,
		Value: t.source[start:t.offset],
	}
}

func (t *Tokenizer) scanUnknown() Token {
	return Token{
		Type: TokenTypeEOF,
	}
}

func (t *Tokenizer) peekRune(offset int) (rune, int) {
	if offset >= len(t.source) {
		return -1, offset
	}
	r, size := utf8.DecodeRuneInString(t.source[offset:])
	return r, offset + size
}

func isWhitespaceChar(ch rune) bool {
	return ch == ' ' || ch == '\t'
}

func isNewLineChar(ch rune) bool {
	return ch == '\n' || ch == '\r'
}

func isIdentifierChar(ch rune, pos int) bool {
	return unicode.IsLetter(ch) || (pos > 0 && unicode.IsDigit(ch)) || (pos == 0 && ch == '#')
}

func isOperatorChar(ch rune) bool {
	return slices.Contains(operatorChars, ch)
}

func isNumberChar(ch rune, pos int) bool {
	return unicode.IsDigit(ch) || (pos > 0 && ch == '.')
}
