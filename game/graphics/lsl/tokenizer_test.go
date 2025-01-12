package lsl_test

import (
	"slices"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"

	"github.com/mokiat/lacking/game/graphics/lsl"
)

var _ = Describe("Tokenizer", func() {

	at := func(line, column uint32) lsl.Position {
		if line == 1 {
			// The first line does not include any tabs in front of it, because
			// gofmt does not allow to move the backtick character to the beginning.
			return lsl.At(line, column)
		} else {
			// NOTE: Adding 3 characters due to three tabs in front of each line
			// because of the way the test cases are formatted.
			return lsl.At(line, column+3)
		}
	}

	token := func(kind lsl.TokenType, value string, pos lsl.Position) lsl.Token {
		return lsl.Token{
			Type:  kind,
			Value: value,
			Pos:   pos,
		}
	}

	DescribeTable("Next", func(inSource string, expectedTokens []lsl.Token) {
		format.MaxLength = 100000
		outTokens := slices.Collect(lsl.Tokenize(inSource))
		Expect(outTokens).To(Equal(expectedTokens))
	},
		Entry("blank source",
			"",
			nil,
		),
		Entry("empty source",
			" \t ",
			nil,
		),
		Entry("comment",
			`
			// This is a comment
			//This is another comment
			`,
			[]lsl.Token{
				token(lsl.TokenTypeNewLine, "\n", at(1, 1)),
				token(lsl.TokenTypeComment, "This is a comment", at(2, 1)),
				token(lsl.TokenTypeNewLine, "\n", at(2, 21)),
				token(lsl.TokenTypeComment, "This is another comment", at(3, 1)),
				token(lsl.TokenTypeNewLine, "\n", at(3, 26)),
			},
		),
		Entry("block declaration",
			`
			uniforms {
			}
			uniforms{
			}
			`,
			[]lsl.Token{
				token(lsl.TokenTypeNewLine, "\n", at(1, 1)),
				token(lsl.TokenTypeIdentifier, "uniforms", at(2, 1)),
				token(lsl.TokenTypeOperator, "{", at(2, 10)),
				token(lsl.TokenTypeNewLine, "\n", at(2, 11)),
				token(lsl.TokenTypeOperator, "}", at(3, 1)),
				token(lsl.TokenTypeNewLine, "\n", at(3, 2)),
				token(lsl.TokenTypeIdentifier, "uniforms", at(4, 1)),
				token(lsl.TokenTypeOperator, "{", at(4, 9)),
				token(lsl.TokenTypeNewLine, "\n", at(4, 10)),
				token(lsl.TokenTypeOperator, "}", at(5, 1)),
				token(lsl.TokenTypeNewLine, "\n", at(5, 2)),
			},
		),
		Entry("block content declaration",
			`
			color vec3
			intensity 	 float
			`,
			[]lsl.Token{
				token(lsl.TokenTypeNewLine, "\n", at(1, 1)),
				token(lsl.TokenTypeIdentifier, "color", at(2, 1)),
				token(lsl.TokenTypeIdentifier, "vec3", at(2, 7)),
				token(lsl.TokenTypeNewLine, "\n", at(2, 11)),
				token(lsl.TokenTypeIdentifier, "intensity", at(3, 1)),
				token(lsl.TokenTypeIdentifier, "float", at(3, 13)),
				token(lsl.TokenTypeNewLine, "\n", at(3, 18)),
			},
		),
		Entry("function declaration",
			`
			func hello(a vec3, b float) (vec4, float) {
			}
			func world() vec3 {
			}
			func test() {
			}
			`,
			[]lsl.Token{
				token(lsl.TokenTypeNewLine, "\n", at(1, 1)),

				token(lsl.TokenTypeIdentifier, "func", at(2, 1)),
				token(lsl.TokenTypeIdentifier, "hello", at(2, 6)),
				token(lsl.TokenTypeOperator, "(", at(2, 11)),
				token(lsl.TokenTypeIdentifier, "a", at(2, 12)),
				token(lsl.TokenTypeIdentifier, "vec3", at(2, 14)),
				token(lsl.TokenTypeOperator, ",", at(2, 18)),
				token(lsl.TokenTypeIdentifier, "b", at(2, 20)),
				token(lsl.TokenTypeIdentifier, "float", at(2, 22)),
				token(lsl.TokenTypeOperator, ")", at(2, 27)),
				token(lsl.TokenTypeOperator, "(", at(2, 29)),
				token(lsl.TokenTypeIdentifier, "vec4", at(2, 30)),
				token(lsl.TokenTypeOperator, ",", at(2, 34)),
				token(lsl.TokenTypeIdentifier, "float", at(2, 36)),
				token(lsl.TokenTypeOperator, ")", at(2, 41)),
				token(lsl.TokenTypeOperator, "{", at(2, 43)),
				token(lsl.TokenTypeNewLine, "\n", at(2, 44)),
				token(lsl.TokenTypeOperator, "}", at(3, 1)),
				token(lsl.TokenTypeNewLine, "\n", at(3, 2)),

				token(lsl.TokenTypeIdentifier, "func", at(4, 1)),
				token(lsl.TokenTypeIdentifier, "world", at(4, 6)),
				token(lsl.TokenTypeOperator, "(", at(4, 11)),
				token(lsl.TokenTypeOperator, ")", at(4, 12)),
				token(lsl.TokenTypeIdentifier, "vec3", at(4, 14)),
				token(lsl.TokenTypeOperator, "{", at(4, 19)),
				token(lsl.TokenTypeNewLine, "\n", at(4, 20)),
				token(lsl.TokenTypeOperator, "}", at(5, 1)),
				token(lsl.TokenTypeNewLine, "\n", at(5, 2)),

				token(lsl.TokenTypeIdentifier, "func", at(6, 1)),
				token(lsl.TokenTypeIdentifier, "test", at(6, 6)),
				token(lsl.TokenTypeOperator, "(", at(6, 10)),
				token(lsl.TokenTypeOperator, ")", at(6, 11)),
				token(lsl.TokenTypeOperator, "{", at(6, 13)),
				token(lsl.TokenTypeNewLine, "\n", at(6, 14)),
				token(lsl.TokenTypeOperator, "}", at(7, 1)),
				token(lsl.TokenTypeNewLine, "\n", at(7, 2)),
			},
		),
		Entry("var declaration",
			`
			var a vec3
			var b float = 1.0
			var c float = -1.0
			`,
			[]lsl.Token{
				token(lsl.TokenTypeNewLine, "\n", at(1, 1)),

				token(lsl.TokenTypeIdentifier, "var", at(2, 1)),
				token(lsl.TokenTypeIdentifier, "a", at(2, 5)),
				token(lsl.TokenTypeIdentifier, "vec3", at(2, 7)),
				token(lsl.TokenTypeNewLine, "\n", at(2, 11)),

				token(lsl.TokenTypeIdentifier, "var", at(3, 1)),
				token(lsl.TokenTypeIdentifier, "b", at(3, 5)),
				token(lsl.TokenTypeIdentifier, "float", at(3, 7)),
				token(lsl.TokenTypeOperator, "=", at(3, 13)),
				token(lsl.TokenTypeNumber, "1.0", at(3, 15)),
				token(lsl.TokenTypeNewLine, "\n", at(3, 18)),

				token(lsl.TokenTypeIdentifier, "var", at(4, 1)),
				token(lsl.TokenTypeIdentifier, "c", at(4, 5)),
				token(lsl.TokenTypeIdentifier, "float", at(4, 7)),
				token(lsl.TokenTypeOperator, "=", at(4, 13)),
				token(lsl.TokenTypeOperator, "-", at(4, 15)),
				token(lsl.TokenTypeNumber, "1.0", at(4, 16)),
				token(lsl.TokenTypeNewLine, "\n", at(4, 19)),
			},
		),
		Entry("field access",
			`
			color.x=1.0
			color.y=-color.x
			`,
			[]lsl.Token{
				token(lsl.TokenTypeNewLine, "\n", at(1, 1)),

				token(lsl.TokenTypeIdentifier, "color", at(2, 1)),
				token(lsl.TokenTypeOperator, ".", at(2, 6)),
				token(lsl.TokenTypeIdentifier, "x", at(2, 7)),
				token(lsl.TokenTypeOperator, "=", at(2, 8)),
				token(lsl.TokenTypeNumber, "1.0", at(2, 9)),
				token(lsl.TokenTypeNewLine, "\n", at(2, 12)),

				token(lsl.TokenTypeIdentifier, "color", at(3, 1)),
				token(lsl.TokenTypeOperator, ".", at(3, 6)),
				token(lsl.TokenTypeIdentifier, "y", at(3, 7)),
				token(lsl.TokenTypeOperator, "=", at(3, 8)),
				token(lsl.TokenTypeOperator, "-", at(3, 9)),
				token(lsl.TokenTypeIdentifier, "color", at(3, 10)),
				token(lsl.TokenTypeOperator, ".", at(3, 15)),
				token(lsl.TokenTypeIdentifier, "x", at(3, 16)),
				token(lsl.TokenTypeNewLine, "\n", at(3, 17)),
			},
		),
		Entry("numeric expression",
			`
			x = (7.1 * ((13 / 2)%5.5)+3.0) - 5
			`,
			[]lsl.Token{
				token(lsl.TokenTypeNewLine, "\n", at(1, 1)),

				token(lsl.TokenTypeIdentifier, "x", at(2, 1)),
				token(lsl.TokenTypeOperator, "=", at(2, 3)),
				token(lsl.TokenTypeOperator, "(", at(2, 5)),
				token(lsl.TokenTypeNumber, "7.1", at(2, 6)),
				token(lsl.TokenTypeOperator, "*", at(2, 10)),
				token(lsl.TokenTypeOperator, "(", at(2, 12)),
				token(lsl.TokenTypeOperator, "(", at(2, 13)),
				token(lsl.TokenTypeNumber, "13", at(2, 14)),
				token(lsl.TokenTypeOperator, "/", at(2, 17)),
				token(lsl.TokenTypeNumber, "2", at(2, 19)),
				token(lsl.TokenTypeOperator, ")", at(2, 20)),
				token(lsl.TokenTypeOperator, "%", at(2, 21)),
				token(lsl.TokenTypeNumber, "5.5", at(2, 22)),
				token(lsl.TokenTypeOperator, ")", at(2, 25)),
				token(lsl.TokenTypeOperator, "+", at(2, 26)),
				token(lsl.TokenTypeNumber, "3.0", at(2, 27)),
				token(lsl.TokenTypeOperator, ")", at(2, 30)),
				token(lsl.TokenTypeOperator, "-", at(2, 32)),
				token(lsl.TokenTypeNumber, "5", at(2, 34)),
				token(lsl.TokenTypeNewLine, "\n", at(2, 35)),
			},
		),
		Entry("bit-wise expression",
			`
			x = ((6 << 1) ^ (8 >> 2)) | (3 & 5)
			`,
			[]lsl.Token{
				token(lsl.TokenTypeNewLine, "\n", at(1, 1)),

				token(lsl.TokenTypeIdentifier, "x", at(2, 1)),
				token(lsl.TokenTypeOperator, "=", at(2, 3)),
				token(lsl.TokenTypeOperator, "(", at(2, 5)),
				token(lsl.TokenTypeOperator, "(", at(2, 6)),
				token(lsl.TokenTypeNumber, "6", at(2, 7)),
				token(lsl.TokenTypeOperator, "<<", at(2, 9)),
				token(lsl.TokenTypeNumber, "1", at(2, 12)),
				token(lsl.TokenTypeOperator, ")", at(2, 13)),
				token(lsl.TokenTypeOperator, "^", at(2, 15)),
				token(lsl.TokenTypeOperator, "(", at(2, 17)),
				token(lsl.TokenTypeNumber, "8", at(2, 18)),
				token(lsl.TokenTypeOperator, ">>", at(2, 20)),
				token(lsl.TokenTypeNumber, "2", at(2, 23)),
				token(lsl.TokenTypeOperator, ")", at(2, 24)),
				token(lsl.TokenTypeOperator, ")", at(2, 25)),
				token(lsl.TokenTypeOperator, "|", at(2, 27)),
				token(lsl.TokenTypeOperator, "(", at(2, 29)),
				token(lsl.TokenTypeNumber, "3", at(2, 30)),
				token(lsl.TokenTypeOperator, "&", at(2, 32)),
				token(lsl.TokenTypeNumber, "5", at(2, 34)),
				token(lsl.TokenTypeOperator, ")", at(2, 35)),
				token(lsl.TokenTypeNewLine, "\n", at(2, 36)),
			},
		),
		Entry("boolean expression",
			`
			x = ((x.z == z) && (5 != 4) && (1 < 2) && (1.1>3) && true) || (false&&(5 >= 9)&& (2 <= 1))
			y = !true
			`,
			[]lsl.Token{
				token(lsl.TokenTypeNewLine, "\n", at(1, 1)),

				token(lsl.TokenTypeIdentifier, "x", at(2, 1)),
				token(lsl.TokenTypeOperator, "=", at(2, 3)),
				token(lsl.TokenTypeOperator, "(", at(2, 5)),
				token(lsl.TokenTypeOperator, "(", at(2, 6)),
				token(lsl.TokenTypeIdentifier, "x", at(2, 7)),
				token(lsl.TokenTypeOperator, ".", at(2, 8)),
				token(lsl.TokenTypeIdentifier, "z", at(2, 9)),
				token(lsl.TokenTypeOperator, "==", at(2, 11)),
				token(lsl.TokenTypeIdentifier, "z", at(2, 14)),
				token(lsl.TokenTypeOperator, ")", at(2, 15)),
				token(lsl.TokenTypeOperator, "&&", at(2, 17)),
				token(lsl.TokenTypeOperator, "(", at(2, 20)),
				token(lsl.TokenTypeNumber, "5", at(2, 21)),
				token(lsl.TokenTypeOperator, "!=", at(2, 23)),
				token(lsl.TokenTypeNumber, "4", at(2, 26)),
				token(lsl.TokenTypeOperator, ")", at(2, 27)),
				token(lsl.TokenTypeOperator, "&&", at(2, 29)),
				token(lsl.TokenTypeOperator, "(", at(2, 32)),
				token(lsl.TokenTypeNumber, "1", at(2, 33)),
				token(lsl.TokenTypeOperator, "<", at(2, 35)),
				token(lsl.TokenTypeNumber, "2", at(2, 37)),
				token(lsl.TokenTypeOperator, ")", at(2, 38)),
				token(lsl.TokenTypeOperator, "&&", at(2, 40)),
				token(lsl.TokenTypeOperator, "(", at(2, 43)),
				token(lsl.TokenTypeNumber, "1.1", at(2, 44)),
				token(lsl.TokenTypeOperator, ">", at(2, 47)),
				token(lsl.TokenTypeNumber, "3", at(2, 48)),
				token(lsl.TokenTypeOperator, ")", at(2, 49)),
				token(lsl.TokenTypeOperator, "&&", at(2, 51)),
				token(lsl.TokenTypeIdentifier, "true", at(2, 54)),
				token(lsl.TokenTypeOperator, ")", at(2, 58)),
				token(lsl.TokenTypeOperator, "||", at(2, 60)),
				token(lsl.TokenTypeOperator, "(", at(2, 63)),
				token(lsl.TokenTypeIdentifier, "false", at(2, 64)),
				token(lsl.TokenTypeOperator, "&&", at(2, 69)),
				token(lsl.TokenTypeOperator, "(", at(2, 71)),
				token(lsl.TokenTypeNumber, "5", at(2, 72)),
				token(lsl.TokenTypeOperator, ">=", at(2, 74)),
				token(lsl.TokenTypeNumber, "9", at(2, 77)),
				token(lsl.TokenTypeOperator, ")", at(2, 78)),
				token(lsl.TokenTypeOperator, "&&", at(2, 79)),
				token(lsl.TokenTypeOperator, "(", at(2, 82)),
				token(lsl.TokenTypeNumber, "2", at(2, 83)),
				token(lsl.TokenTypeOperator, "<=", at(2, 85)),
				token(lsl.TokenTypeNumber, "1", at(2, 88)),
				token(lsl.TokenTypeOperator, ")", at(2, 89)),
				token(lsl.TokenTypeOperator, ")", at(2, 90)),
				token(lsl.TokenTypeNewLine, "\n", at(2, 91)),

				token(lsl.TokenTypeIdentifier, "y", at(3, 1)),
				token(lsl.TokenTypeOperator, "=", at(3, 3)),
				token(lsl.TokenTypeOperator, "!", at(3, 5)),
				token(lsl.TokenTypeIdentifier, "true", at(3, 6)),
				token(lsl.TokenTypeNewLine, "\n", at(3, 10)),
			},
		),
		Entry("assignments",
			`
			a=13
			b = 55.3
			c += 5
			d+=-5
			e -= 5
			f-=-5
			g *= 5
			h*=-5
			i /= 5
			j/=-5
			k %= 5
			l%=-5
			m >>= 5
			n>>=-5
			o <<= 5
			p<<=-5
			q &= 5
			r&=-5
			s ^= 5
			t^=-5
			u |= 5
			v|=-5
			w := 8
			x:=8
			`,
			[]lsl.Token{
				token(lsl.TokenTypeNewLine, "\n", at(1, 1)),

				token(lsl.TokenTypeIdentifier, "a", at(2, 1)),
				token(lsl.TokenTypeOperator, "=", at(2, 2)),
				token(lsl.TokenTypeNumber, "13", at(2, 3)),
				token(lsl.TokenTypeNewLine, "\n", at(2, 5)),

				token(lsl.TokenTypeIdentifier, "b", at(3, 1)),
				token(lsl.TokenTypeOperator, "=", at(3, 3)),
				token(lsl.TokenTypeNumber, "55.3", at(3, 5)),
				token(lsl.TokenTypeNewLine, "\n", at(3, 9)),

				token(lsl.TokenTypeIdentifier, "c", at(4, 1)),
				token(lsl.TokenTypeOperator, "+=", at(4, 3)),
				token(lsl.TokenTypeNumber, "5", at(4, 6)),
				token(lsl.TokenTypeNewLine, "\n", at(4, 7)),

				token(lsl.TokenTypeIdentifier, "d", at(5, 1)),
				token(lsl.TokenTypeOperator, "+=", at(5, 2)),
				token(lsl.TokenTypeOperator, "-", at(5, 4)),
				token(lsl.TokenTypeNumber, "5", at(5, 5)),
				token(lsl.TokenTypeNewLine, "\n", at(5, 6)),

				token(lsl.TokenTypeIdentifier, "e", at(6, 1)),
				token(lsl.TokenTypeOperator, "-=", at(6, 3)),
				token(lsl.TokenTypeNumber, "5", at(6, 6)),
				token(lsl.TokenTypeNewLine, "\n", at(6, 7)),

				token(lsl.TokenTypeIdentifier, "f", at(7, 1)),
				token(lsl.TokenTypeOperator, "-=", at(7, 2)),
				token(lsl.TokenTypeOperator, "-", at(7, 4)),
				token(lsl.TokenTypeNumber, "5", at(7, 5)),
				token(lsl.TokenTypeNewLine, "\n", at(7, 6)),

				token(lsl.TokenTypeIdentifier, "g", at(8, 1)),
				token(lsl.TokenTypeOperator, "*=", at(8, 3)),
				token(lsl.TokenTypeNumber, "5", at(8, 6)),
				token(lsl.TokenTypeNewLine, "\n", at(8, 7)),

				token(lsl.TokenTypeIdentifier, "h", at(9, 1)),
				token(lsl.TokenTypeOperator, "*=", at(9, 2)),
				token(lsl.TokenTypeOperator, "-", at(9, 4)),
				token(lsl.TokenTypeNumber, "5", at(9, 5)),
				token(lsl.TokenTypeNewLine, "\n", at(9, 6)),

				token(lsl.TokenTypeIdentifier, "i", at(10, 1)),
				token(lsl.TokenTypeOperator, "/=", at(10, 3)),
				token(lsl.TokenTypeNumber, "5", at(10, 6)),
				token(lsl.TokenTypeNewLine, "\n", at(10, 7)),

				token(lsl.TokenTypeIdentifier, "j", at(11, 1)),
				token(lsl.TokenTypeOperator, "/=", at(11, 2)),
				token(lsl.TokenTypeOperator, "-", at(11, 4)),
				token(lsl.TokenTypeNumber, "5", at(11, 5)),
				token(lsl.TokenTypeNewLine, "\n", at(11, 6)),

				token(lsl.TokenTypeIdentifier, "k", at(12, 1)),
				token(lsl.TokenTypeOperator, "%=", at(12, 3)),
				token(lsl.TokenTypeNumber, "5", at(12, 6)),
				token(lsl.TokenTypeNewLine, "\n", at(12, 7)),

				token(lsl.TokenTypeIdentifier, "l", at(13, 1)),
				token(lsl.TokenTypeOperator, "%=", at(13, 2)),
				token(lsl.TokenTypeOperator, "-", at(13, 4)),
				token(lsl.TokenTypeNumber, "5", at(13, 5)),
				token(lsl.TokenTypeNewLine, "\n", at(13, 6)),

				token(lsl.TokenTypeIdentifier, "m", at(14, 1)),
				token(lsl.TokenTypeOperator, ">>=", at(14, 3)),
				token(lsl.TokenTypeNumber, "5", at(14, 7)),
				token(lsl.TokenTypeNewLine, "\n", at(14, 8)),

				token(lsl.TokenTypeIdentifier, "n", at(15, 1)),
				token(lsl.TokenTypeOperator, ">>=", at(15, 2)),
				token(lsl.TokenTypeOperator, "-", at(15, 5)),
				token(lsl.TokenTypeNumber, "5", at(15, 6)),
				token(lsl.TokenTypeNewLine, "\n", at(15, 7)),

				token(lsl.TokenTypeIdentifier, "o", at(16, 1)),
				token(lsl.TokenTypeOperator, "<<=", at(16, 3)),
				token(lsl.TokenTypeNumber, "5", at(16, 7)),
				token(lsl.TokenTypeNewLine, "\n", at(16, 8)),

				token(lsl.TokenTypeIdentifier, "p", at(17, 1)),
				token(lsl.TokenTypeOperator, "<<=", at(17, 2)),
				token(lsl.TokenTypeOperator, "-", at(17, 5)),
				token(lsl.TokenTypeNumber, "5", at(17, 6)),
				token(lsl.TokenTypeNewLine, "\n", at(17, 7)),

				token(lsl.TokenTypeIdentifier, "q", at(18, 1)),
				token(lsl.TokenTypeOperator, "&=", at(18, 3)),
				token(lsl.TokenTypeNumber, "5", at(18, 6)),
				token(lsl.TokenTypeNewLine, "\n", at(18, 7)),

				token(lsl.TokenTypeIdentifier, "r", at(19, 1)),
				token(lsl.TokenTypeOperator, "&=", at(19, 2)),
				token(lsl.TokenTypeOperator, "-", at(19, 4)),
				token(lsl.TokenTypeNumber, "5", at(19, 5)),
				token(lsl.TokenTypeNewLine, "\n", at(19, 6)),

				token(lsl.TokenTypeIdentifier, "s", at(20, 1)),
				token(lsl.TokenTypeOperator, "^=", at(20, 3)),
				token(lsl.TokenTypeNumber, "5", at(20, 6)),
				token(lsl.TokenTypeNewLine, "\n", at(20, 7)),

				token(lsl.TokenTypeIdentifier, "t", at(21, 1)),
				token(lsl.TokenTypeOperator, "^=", at(21, 2)),
				token(lsl.TokenTypeOperator, "-", at(21, 4)),
				token(lsl.TokenTypeNumber, "5", at(21, 5)),
				token(lsl.TokenTypeNewLine, "\n", at(21, 6)),

				token(lsl.TokenTypeIdentifier, "u", at(22, 1)),
				token(lsl.TokenTypeOperator, "|=", at(22, 3)),
				token(lsl.TokenTypeNumber, "5", at(22, 6)),
				token(lsl.TokenTypeNewLine, "\n", at(22, 7)),

				token(lsl.TokenTypeIdentifier, "v", at(23, 1)),
				token(lsl.TokenTypeOperator, "|=", at(23, 2)),
				token(lsl.TokenTypeOperator, "-", at(23, 4)),
				token(lsl.TokenTypeNumber, "5", at(23, 5)),
				token(lsl.TokenTypeNewLine, "\n", at(23, 6)),

				token(lsl.TokenTypeIdentifier, "w", at(24, 1)),
				token(lsl.TokenTypeOperator, ":=", at(24, 3)),
				token(lsl.TokenTypeNumber, "8", at(24, 6)),
				token(lsl.TokenTypeNewLine, "\n", at(24, 7)),

				token(lsl.TokenTypeIdentifier, "x", at(25, 1)),
				token(lsl.TokenTypeOperator, ":=", at(25, 2)),
				token(lsl.TokenTypeNumber, "8", at(25, 4)),
				token(lsl.TokenTypeNewLine, "\n", at(25, 5)),
			},
		),
		Entry("func invocation",
			`
			doSomething(15.55, 10, someVar)
			var a vec4 = calculate()
			var b float = sum(
				115.3,
				abs(-15),
			)
			`,
			[]lsl.Token{
				token(lsl.TokenTypeNewLine, "\n", at(1, 1)),

				token(lsl.TokenTypeIdentifier, "doSomething", at(2, 1)),
				token(lsl.TokenTypeOperator, "(", at(2, 12)),
				token(lsl.TokenTypeNumber, "15.55", at(2, 13)),
				token(lsl.TokenTypeOperator, ",", at(2, 18)),
				token(lsl.TokenTypeNumber, "10", at(2, 20)),
				token(lsl.TokenTypeOperator, ",", at(2, 22)),
				token(lsl.TokenTypeIdentifier, "someVar", at(2, 24)),
				token(lsl.TokenTypeOperator, ")", at(2, 31)),
				token(lsl.TokenTypeNewLine, "\n", at(2, 32)),

				token(lsl.TokenTypeIdentifier, "var", at(3, 1)),
				token(lsl.TokenTypeIdentifier, "a", at(3, 5)),
				token(lsl.TokenTypeIdentifier, "vec4", at(3, 7)),
				token(lsl.TokenTypeOperator, "=", at(3, 12)),
				token(lsl.TokenTypeIdentifier, "calculate", at(3, 14)),
				token(lsl.TokenTypeOperator, "(", at(3, 23)),
				token(lsl.TokenTypeOperator, ")", at(3, 24)),
				token(lsl.TokenTypeNewLine, "\n", at(3, 25)),

				token(lsl.TokenTypeIdentifier, "var", at(4, 1)),
				token(lsl.TokenTypeIdentifier, "b", at(4, 5)),
				token(lsl.TokenTypeIdentifier, "float", at(4, 7)),
				token(lsl.TokenTypeOperator, "=", at(4, 13)),
				token(lsl.TokenTypeIdentifier, "sum", at(4, 15)),
				token(lsl.TokenTypeOperator, "(", at(4, 18)),
				token(lsl.TokenTypeNewLine, "\n", at(4, 19)),
				token(lsl.TokenTypeNumber, "115.3", at(5, 2)),
				token(lsl.TokenTypeOperator, ",", at(5, 7)),
				token(lsl.TokenTypeNewLine, "\n", at(5, 8)),
				token(lsl.TokenTypeIdentifier, "abs", at(6, 2)),
				token(lsl.TokenTypeOperator, "(", at(6, 5)),
				token(lsl.TokenTypeOperator, "-", at(6, 6)),
				token(lsl.TokenTypeNumber, "15", at(6, 7)),
				token(lsl.TokenTypeOperator, ")", at(6, 9)),
				token(lsl.TokenTypeOperator, ",", at(6, 10)),
				token(lsl.TokenTypeNewLine, "\n", at(6, 11)),
				token(lsl.TokenTypeOperator, ")", at(7, 1)),
				token(lsl.TokenTypeNewLine, "\n", at(7, 2)),
			},
		),
	)
})
