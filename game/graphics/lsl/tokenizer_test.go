package lsl_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/lacking/game/graphics/lsl"
)

var _ = Describe("Tokenizer", func() {

	DescribeTable("NextToken", func(inSource string, expectedTokens []lsl.Token) {
		tokenizer := lsl.NewTokenizer(inSource)
		var outTokens []lsl.Token
		for {
			token := tokenizer.Next()
			if token.Type == lsl.TokenTypeEOF {
				break
			}
			outTokens = append(outTokens, token)
		}
		Expect(outTokens).To(Equal(expectedTokens))
	},
		Entry("empty source",
			``,
			nil,
		),
		Entry("comment",
			`
			// This is a comment
			//This is another comment
			`,
			[]lsl.Token{
				{Type: lsl.TokenTypeNewLine, Value: "\n"},
				{Type: lsl.TokenTypeComment, Value: "This is a comment"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},
				{Type: lsl.TokenTypeComment, Value: "This is another comment"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},
			},
		),
		Entry("uniform block declaration",
			`
			uniform {
			}
			`,
			[]lsl.Token{
				{Type: lsl.TokenTypeNewLine, Value: "\n"},
				{Type: lsl.TokenTypeIdentifier, Value: "uniform"},
				{Type: lsl.TokenTypeOperator, Value: "{"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},
				{Type: lsl.TokenTypeOperator, Value: "}"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},
			},
		),
		Entry("varying block declaration",
			`
			varying {
			}
			`,
			[]lsl.Token{
				{Type: lsl.TokenTypeNewLine, Value: "\n"},
				{Type: lsl.TokenTypeIdentifier, Value: "varying"},
				{Type: lsl.TokenTypeOperator, Value: "{"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},
				{Type: lsl.TokenTypeOperator, Value: "}"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},
			},
		),
		Entry("uniform/varying block content",
			`
			color vec3
			intensity float
			`,
			[]lsl.Token{
				{Type: lsl.TokenTypeNewLine, Value: "\n"},

				{Type: lsl.TokenTypeIdentifier, Value: "color"},
				{Type: lsl.TokenTypeIdentifier, Value: "vec3"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},

				{Type: lsl.TokenTypeIdentifier, Value: "intensity"},
				{Type: lsl.TokenTypeIdentifier, Value: "float"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},
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
				{Type: lsl.TokenTypeNewLine, Value: "\n"},

				{Type: lsl.TokenTypeIdentifier, Value: "func"},
				{Type: lsl.TokenTypeIdentifier, Value: "hello"},
				{Type: lsl.TokenTypeOperator, Value: "("},
				{Type: lsl.TokenTypeIdentifier, Value: "a"},
				{Type: lsl.TokenTypeIdentifier, Value: "vec3"},
				{Type: lsl.TokenTypeOperator, Value: ","},
				{Type: lsl.TokenTypeIdentifier, Value: "b"},
				{Type: lsl.TokenTypeIdentifier, Value: "float"},
				{Type: lsl.TokenTypeOperator, Value: ")"},
				{Type: lsl.TokenTypeOperator, Value: "("},
				{Type: lsl.TokenTypeIdentifier, Value: "vec4"},
				{Type: lsl.TokenTypeOperator, Value: ","},
				{Type: lsl.TokenTypeIdentifier, Value: "float"},
				{Type: lsl.TokenTypeOperator, Value: ")"},
				{Type: lsl.TokenTypeOperator, Value: "{"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},
				{Type: lsl.TokenTypeOperator, Value: "}"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},

				{Type: lsl.TokenTypeIdentifier, Value: "func"},
				{Type: lsl.TokenTypeIdentifier, Value: "world"},
				{Type: lsl.TokenTypeOperator, Value: "("},
				{Type: lsl.TokenTypeOperator, Value: ")"},
				{Type: lsl.TokenTypeIdentifier, Value: "vec3"},
				{Type: lsl.TokenTypeOperator, Value: "{"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},
				{Type: lsl.TokenTypeOperator, Value: "}"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},

				{Type: lsl.TokenTypeIdentifier, Value: "func"},
				{Type: lsl.TokenTypeIdentifier, Value: "test"},
				{Type: lsl.TokenTypeOperator, Value: "("},
				{Type: lsl.TokenTypeOperator, Value: ")"},
				{Type: lsl.TokenTypeOperator, Value: "{"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},
				{Type: lsl.TokenTypeOperator, Value: "}"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},
			},
		),
		Entry("var declaration",
			`
			var a vec3
			var b float = 1.0
			var c float = -1.0
			`,
			[]lsl.Token{
				{Type: lsl.TokenTypeNewLine, Value: "\n"},

				{Type: lsl.TokenTypeIdentifier, Value: "var"},
				{Type: lsl.TokenTypeIdentifier, Value: "a"},
				{Type: lsl.TokenTypeIdentifier, Value: "vec3"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},

				{Type: lsl.TokenTypeIdentifier, Value: "var"},
				{Type: lsl.TokenTypeIdentifier, Value: "b"},
				{Type: lsl.TokenTypeIdentifier, Value: "float"},
				{Type: lsl.TokenTypeOperator, Value: "="},
				{Type: lsl.TokenTypeNumber, Value: "1.0"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},

				{Type: lsl.TokenTypeIdentifier, Value: "var"},
				{Type: lsl.TokenTypeIdentifier, Value: "c"},
				{Type: lsl.TokenTypeIdentifier, Value: "float"},
				{Type: lsl.TokenTypeOperator, Value: "="},
				{Type: lsl.TokenTypeOperator, Value: "-"},
				{Type: lsl.TokenTypeNumber, Value: "1.0"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},
			},
		),
		Entry("field access",
			`
			color.x=1.0
			color.y=-color.x
			`,
			[]lsl.Token{
				{Type: lsl.TokenTypeNewLine, Value: "\n"},

				{Type: lsl.TokenTypeIdentifier, Value: "color"},
				{Type: lsl.TokenTypeOperator, Value: "."},
				{Type: lsl.TokenTypeIdentifier, Value: "x"},
				{Type: lsl.TokenTypeOperator, Value: "="},
				{Type: lsl.TokenTypeNumber, Value: "1.0"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},

				{Type: lsl.TokenTypeIdentifier, Value: "color"},
				{Type: lsl.TokenTypeOperator, Value: "."},
				{Type: lsl.TokenTypeIdentifier, Value: "y"},
				{Type: lsl.TokenTypeOperator, Value: "="},
				{Type: lsl.TokenTypeOperator, Value: "-"},
				{Type: lsl.TokenTypeIdentifier, Value: "color"},
				{Type: lsl.TokenTypeOperator, Value: "."},
				{Type: lsl.TokenTypeIdentifier, Value: "x"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},
			},
		),
		Entry("numeric expression",
			`
			x = (7.1 * ((13 / 2)%5.5)+3.0) - 5
			`,
			[]lsl.Token{
				{Type: lsl.TokenTypeNewLine, Value: "\n"},
				{Type: lsl.TokenTypeIdentifier, Value: "x"},
				{Type: lsl.TokenTypeOperator, Value: "="},
				{Type: lsl.TokenTypeOperator, Value: "("},
				{Type: lsl.TokenTypeNumber, Value: "7.1"},
				{Type: lsl.TokenTypeOperator, Value: "*"},
				{Type: lsl.TokenTypeOperator, Value: "("},
				{Type: lsl.TokenTypeOperator, Value: "("},
				{Type: lsl.TokenTypeNumber, Value: "13"},
				{Type: lsl.TokenTypeOperator, Value: "/"},
				{Type: lsl.TokenTypeNumber, Value: "2"},
				{Type: lsl.TokenTypeOperator, Value: ")"},
				{Type: lsl.TokenTypeOperator, Value: "%"},
				{Type: lsl.TokenTypeNumber, Value: "5.5"},
				{Type: lsl.TokenTypeOperator, Value: ")"},
				{Type: lsl.TokenTypeOperator, Value: "+"},
				{Type: lsl.TokenTypeNumber, Value: "3.0"},
				{Type: lsl.TokenTypeOperator, Value: ")"},
				{Type: lsl.TokenTypeOperator, Value: "-"},
				{Type: lsl.TokenTypeNumber, Value: "5"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},
			},
		),
		Entry("bit-wise expression",
			`
			x = ((6 << 1) ^ (8 >> 2)) | (3 & 5)
			`,
			[]lsl.Token{
				{Type: lsl.TokenTypeNewLine, Value: "\n"},
				{Type: lsl.TokenTypeIdentifier, Value: "x"},
				{Type: lsl.TokenTypeOperator, Value: "="},
				{Type: lsl.TokenTypeOperator, Value: "("},
				{Type: lsl.TokenTypeOperator, Value: "("},
				{Type: lsl.TokenTypeNumber, Value: "6"},
				{Type: lsl.TokenTypeOperator, Value: "<<"},
				{Type: lsl.TokenTypeNumber, Value: "1"},
				{Type: lsl.TokenTypeOperator, Value: ")"},
				{Type: lsl.TokenTypeOperator, Value: "^"},
				{Type: lsl.TokenTypeOperator, Value: "("},
				{Type: lsl.TokenTypeNumber, Value: "8"},
				{Type: lsl.TokenTypeOperator, Value: ">>"},
				{Type: lsl.TokenTypeNumber, Value: "2"},
				{Type: lsl.TokenTypeOperator, Value: ")"},
				{Type: lsl.TokenTypeOperator, Value: ")"},
				{Type: lsl.TokenTypeOperator, Value: "|"},
				{Type: lsl.TokenTypeOperator, Value: "("},
				{Type: lsl.TokenTypeNumber, Value: "3"},
				{Type: lsl.TokenTypeOperator, Value: "&"},
				{Type: lsl.TokenTypeNumber, Value: "5"},
				{Type: lsl.TokenTypeOperator, Value: ")"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},
			},
		),
		Entry("boolean expression",
			`
			x = ((x.z == z) && (5 != 4) && (1 < 2) && (1.1>3) && true) || (false&&(5 >= 9)&& (2 <= 1))
			y = !true
			`,
			[]lsl.Token{
				{Type: lsl.TokenTypeNewLine, Value: "\n"},
				{Type: lsl.TokenTypeIdentifier, Value: "x"},
				{Type: lsl.TokenTypeOperator, Value: "="},
				{Type: lsl.TokenTypeOperator, Value: "("},
				{Type: lsl.TokenTypeOperator, Value: "("},
				{Type: lsl.TokenTypeIdentifier, Value: "x"},
				{Type: lsl.TokenTypeOperator, Value: "."},
				{Type: lsl.TokenTypeIdentifier, Value: "z"},
				{Type: lsl.TokenTypeOperator, Value: "=="},
				{Type: lsl.TokenTypeIdentifier, Value: "z"},
				{Type: lsl.TokenTypeOperator, Value: ")"},
				{Type: lsl.TokenTypeOperator, Value: "&&"},
				{Type: lsl.TokenTypeOperator, Value: "("},
				{Type: lsl.TokenTypeNumber, Value: "5"},
				{Type: lsl.TokenTypeOperator, Value: "!="},
				{Type: lsl.TokenTypeNumber, Value: "4"},
				{Type: lsl.TokenTypeOperator, Value: ")"},
				{Type: lsl.TokenTypeOperator, Value: "&&"},
				{Type: lsl.TokenTypeOperator, Value: "("},
				{Type: lsl.TokenTypeNumber, Value: "1"},
				{Type: lsl.TokenTypeOperator, Value: "<"},
				{Type: lsl.TokenTypeNumber, Value: "2"},
				{Type: lsl.TokenTypeOperator, Value: ")"},
				{Type: lsl.TokenTypeOperator, Value: "&&"},
				{Type: lsl.TokenTypeOperator, Value: "("},
				{Type: lsl.TokenTypeNumber, Value: "1.1"},
				{Type: lsl.TokenTypeOperator, Value: ">"},
				{Type: lsl.TokenTypeNumber, Value: "3"},
				{Type: lsl.TokenTypeOperator, Value: ")"},
				{Type: lsl.TokenTypeOperator, Value: "&&"},
				{Type: lsl.TokenTypeIdentifier, Value: "true"},
				{Type: lsl.TokenTypeOperator, Value: ")"},
				{Type: lsl.TokenTypeOperator, Value: "||"},
				{Type: lsl.TokenTypeOperator, Value: "("},
				{Type: lsl.TokenTypeIdentifier, Value: "false"},
				{Type: lsl.TokenTypeOperator, Value: "&&"},
				{Type: lsl.TokenTypeOperator, Value: "("},
				{Type: lsl.TokenTypeNumber, Value: "5"},
				{Type: lsl.TokenTypeOperator, Value: ">="},
				{Type: lsl.TokenTypeNumber, Value: "9"},
				{Type: lsl.TokenTypeOperator, Value: ")"},
				{Type: lsl.TokenTypeOperator, Value: "&&"},
				{Type: lsl.TokenTypeOperator, Value: "("},
				{Type: lsl.TokenTypeNumber, Value: "2"},
				{Type: lsl.TokenTypeOperator, Value: "<="},
				{Type: lsl.TokenTypeNumber, Value: "1"},
				{Type: lsl.TokenTypeOperator, Value: ")"},
				{Type: lsl.TokenTypeOperator, Value: ")"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},

				{Type: lsl.TokenTypeIdentifier, Value: "y"},
				{Type: lsl.TokenTypeOperator, Value: "="},
				{Type: lsl.TokenTypeOperator, Value: "!"},
				{Type: lsl.TokenTypeIdentifier, Value: "true"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},
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
			`,
			[]lsl.Token{
				{Type: lsl.TokenTypeNewLine, Value: "\n"},

				{Type: lsl.TokenTypeIdentifier, Value: "a"},
				{Type: lsl.TokenTypeOperator, Value: "="},
				{Type: lsl.TokenTypeNumber, Value: "13"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},

				{Type: lsl.TokenTypeIdentifier, Value: "b"},
				{Type: lsl.TokenTypeOperator, Value: "="},
				{Type: lsl.TokenTypeNumber, Value: "55.3"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},

				{Type: lsl.TokenTypeIdentifier, Value: "c"},
				{Type: lsl.TokenTypeOperator, Value: "+="},
				{Type: lsl.TokenTypeNumber, Value: "5"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},

				{Type: lsl.TokenTypeIdentifier, Value: "d"},
				{Type: lsl.TokenTypeOperator, Value: "+="},
				{Type: lsl.TokenTypeOperator, Value: "-"},
				{Type: lsl.TokenTypeNumber, Value: "5"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},

				{Type: lsl.TokenTypeIdentifier, Value: "e"},
				{Type: lsl.TokenTypeOperator, Value: "-="},
				{Type: lsl.TokenTypeNumber, Value: "5"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},

				{Type: lsl.TokenTypeIdentifier, Value: "f"},
				{Type: lsl.TokenTypeOperator, Value: "-="},
				{Type: lsl.TokenTypeOperator, Value: "-"},
				{Type: lsl.TokenTypeNumber, Value: "5"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},

				{Type: lsl.TokenTypeIdentifier, Value: "g"},
				{Type: lsl.TokenTypeOperator, Value: "*="},
				{Type: lsl.TokenTypeNumber, Value: "5"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},

				{Type: lsl.TokenTypeIdentifier, Value: "h"},
				{Type: lsl.TokenTypeOperator, Value: "*="},
				{Type: lsl.TokenTypeOperator, Value: "-"},
				{Type: lsl.TokenTypeNumber, Value: "5"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},

				{Type: lsl.TokenTypeIdentifier, Value: "i"},
				{Type: lsl.TokenTypeOperator, Value: "/="},
				{Type: lsl.TokenTypeNumber, Value: "5"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},

				{Type: lsl.TokenTypeIdentifier, Value: "j"},
				{Type: lsl.TokenTypeOperator, Value: "/="},
				{Type: lsl.TokenTypeOperator, Value: "-"},
				{Type: lsl.TokenTypeNumber, Value: "5"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},

				{Type: lsl.TokenTypeIdentifier, Value: "k"},
				{Type: lsl.TokenTypeOperator, Value: "%="},
				{Type: lsl.TokenTypeNumber, Value: "5"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},

				{Type: lsl.TokenTypeIdentifier, Value: "l"},
				{Type: lsl.TokenTypeOperator, Value: "%="},
				{Type: lsl.TokenTypeOperator, Value: "-"},
				{Type: lsl.TokenTypeNumber, Value: "5"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},

				{Type: lsl.TokenTypeIdentifier, Value: "m"},
				{Type: lsl.TokenTypeOperator, Value: ">>="},
				{Type: lsl.TokenTypeNumber, Value: "5"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},

				{Type: lsl.TokenTypeIdentifier, Value: "n"},
				{Type: lsl.TokenTypeOperator, Value: ">>="},
				{Type: lsl.TokenTypeOperator, Value: "-"},
				{Type: lsl.TokenTypeNumber, Value: "5"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},

				{Type: lsl.TokenTypeIdentifier, Value: "o"},
				{Type: lsl.TokenTypeOperator, Value: "<<="},
				{Type: lsl.TokenTypeNumber, Value: "5"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},

				{Type: lsl.TokenTypeIdentifier, Value: "p"},
				{Type: lsl.TokenTypeOperator, Value: "<<="},
				{Type: lsl.TokenTypeOperator, Value: "-"},
				{Type: lsl.TokenTypeNumber, Value: "5"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},

				{Type: lsl.TokenTypeIdentifier, Value: "q"},
				{Type: lsl.TokenTypeOperator, Value: "&="},
				{Type: lsl.TokenTypeNumber, Value: "5"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},

				{Type: lsl.TokenTypeIdentifier, Value: "r"},
				{Type: lsl.TokenTypeOperator, Value: "&="},
				{Type: lsl.TokenTypeOperator, Value: "-"},
				{Type: lsl.TokenTypeNumber, Value: "5"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},

				{Type: lsl.TokenTypeIdentifier, Value: "s"},
				{Type: lsl.TokenTypeOperator, Value: "^="},
				{Type: lsl.TokenTypeNumber, Value: "5"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},

				{Type: lsl.TokenTypeIdentifier, Value: "t"},
				{Type: lsl.TokenTypeOperator, Value: "^="},
				{Type: lsl.TokenTypeOperator, Value: "-"},
				{Type: lsl.TokenTypeNumber, Value: "5"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},

				{Type: lsl.TokenTypeIdentifier, Value: "u"},
				{Type: lsl.TokenTypeOperator, Value: "|="},
				{Type: lsl.TokenTypeNumber, Value: "5"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},

				{Type: lsl.TokenTypeIdentifier, Value: "v"},
				{Type: lsl.TokenTypeOperator, Value: "|="},
				{Type: lsl.TokenTypeOperator, Value: "-"},
				{Type: lsl.TokenTypeNumber, Value: "5"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},
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
				{Type: lsl.TokenTypeNewLine, Value: "\n"},
				{Type: lsl.TokenTypeIdentifier, Value: "doSomething"},
				{Type: lsl.TokenTypeOperator, Value: "("},
				{Type: lsl.TokenTypeNumber, Value: "15.55"},
				{Type: lsl.TokenTypeOperator, Value: ","},
				{Type: lsl.TokenTypeNumber, Value: "10"},
				{Type: lsl.TokenTypeOperator, Value: ","},
				{Type: lsl.TokenTypeIdentifier, Value: "someVar"},
				{Type: lsl.TokenTypeOperator, Value: ")"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},
				{Type: lsl.TokenTypeIdentifier, Value: "var"},
				{Type: lsl.TokenTypeIdentifier, Value: "a"},
				{Type: lsl.TokenTypeIdentifier, Value: "vec4"},
				{Type: lsl.TokenTypeOperator, Value: "="},
				{Type: lsl.TokenTypeIdentifier, Value: "calculate"},
				{Type: lsl.TokenTypeOperator, Value: "("},
				{Type: lsl.TokenTypeOperator, Value: ")"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},
				{Type: lsl.TokenTypeIdentifier, Value: "var"},
				{Type: lsl.TokenTypeIdentifier, Value: "b"},
				{Type: lsl.TokenTypeIdentifier, Value: "float"},
				{Type: lsl.TokenTypeOperator, Value: "="},
				{Type: lsl.TokenTypeIdentifier, Value: "sum"},
				{Type: lsl.TokenTypeOperator, Value: "("},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},
				{Type: lsl.TokenTypeNumber, Value: "115.3"},
				{Type: lsl.TokenTypeOperator, Value: ","},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},
				{Type: lsl.TokenTypeIdentifier, Value: "abs"},
				{Type: lsl.TokenTypeOperator, Value: "("},
				{Type: lsl.TokenTypeOperator, Value: "-"},
				{Type: lsl.TokenTypeNumber, Value: "15"},
				{Type: lsl.TokenTypeOperator, Value: ")"},
				{Type: lsl.TokenTypeOperator, Value: ","},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},
				{Type: lsl.TokenTypeOperator, Value: ")"},
				{Type: lsl.TokenTypeNewLine, Value: "\n"},
			},
		),
	)
})
