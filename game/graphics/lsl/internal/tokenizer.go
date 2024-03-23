package internal

import (
	"strings"
	"unicode/utf8"
)

func NewTokenizer(source string) *Tokenizer {
	if !utf8.ValidString(source) {
		panic("source contains invalid UTF-8 characters")
	}
	return &Tokenizer{
		source: source,
		offset: 0,
	}
}

type Tokenizer struct {
	source string
	offset int
}

func (t *Tokenizer) Next() Token {
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
		Type: TokenTypeOEF,
	}
}

func (t *Tokenizer) nextIsWhitespace() bool {
	ch, _ := t.peekRune(t.offset)
	return IsWhitespace(ch)
}

func (t *Tokenizer) scanWhitespace() {
	for t.offset < len(t.source) {
		ch, newOffset := t.peekRune(t.offset)
		if !IsWhitespace(ch) {
			break
		}
		t.offset = newOffset
	}
}

func (t *Tokenizer) nextIsNewLine() bool {
	ch, _ := t.peekRune(t.offset)
	return IsNewLine(ch)
}

func (t *Tokenizer) scanNewLine() Token {
	start := t.offset
	for t.offset < len(t.source) {
		ch, newOffset := t.peekRune(t.offset)
		if !IsNewLine(ch) {
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
		if IsNewLine(ch) {
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
	return IsIdentifier(0, ch)
}

func (t *Tokenizer) scanIdentifier() Token {
	start := t.offset
	charIndex := 0
	for t.offset < len(t.source) {
		ch, newOffset := t.peekRune(t.offset)
		if !IsIdentifier(charIndex, ch) {
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
	return IsOperator(ch)
}

func (t *Tokenizer) scanOperator() Token {
	ch1, ch1Offset := t.peekRune(t.offset)
	ch2, ch2Offset := t.peekRune(ch1Offset)
	ch3, ch3Offset := t.peekRune(ch2Offset)
	switch {
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
	case ch1 == '=':
		t.offset = ch1Offset
		return Token{
			Type:  TokenTypeOperator,
			Value: "=",
		}
	case ch1 == '-':
		t.offset = ch1Offset
		return Token{
			Type:  TokenTypeOperator,
			Value: "-",
		}
	default:
		return Token{}
	}
}

func (t *Tokenizer) nextIsNumber() bool {
	ch, _ := t.peekRune(t.offset)
	return IsNumber(0, ch)
}

func (t *Tokenizer) scanNumber() Token {
	start := t.offset
	charIndex := 0
	for t.offset < len(t.source) {
		ch, newOffset := t.peekRune(t.offset)
		if !IsNumber(charIndex, ch) {
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
		Type: TokenTypeOEF, // FIXME
	}
}

func (t *Tokenizer) peekRune(offset int) (rune, int) {
	if offset >= len(t.source) {
		return -1, offset
	}
	r, size := utf8.DecodeRuneInString(t.source[offset:])
	return r, offset + size
}
