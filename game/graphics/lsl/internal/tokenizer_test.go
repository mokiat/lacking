package internal_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/lacking/game/graphics/lsl/internal"
)

var _ = Describe("Tokenizer", func() {

	DescribeTable("NextToken", func(inSource string, expectedTokens []internal.Token) {
		tokenizer := internal.NewTokenizer(inSource)
		var outTokens []internal.Token
		for {
			token := tokenizer.Next()
			if token.Type == internal.TokenTypeOEF {
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
			[]internal.Token{
				{Type: internal.TokenTypeNewLine, Value: "\n"},
				{Type: internal.TokenTypeComment, Value: "This is a comment"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},
				{Type: internal.TokenTypeComment, Value: "This is another comment"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},
			},
		),
		Entry("uniform block declaration",
			`
			uniform {
			}
			`,
			[]internal.Token{
				{Type: internal.TokenTypeNewLine, Value: "\n"},
				{Type: internal.TokenTypeIdentifier, Value: "uniform"},
				{Type: internal.TokenTypeOperator, Value: "{"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},
				{Type: internal.TokenTypeOperator, Value: "}"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},
			},
		),
		Entry("varying block declaration",
			`
			varying {
			}
			`,
			[]internal.Token{
				{Type: internal.TokenTypeNewLine, Value: "\n"},
				{Type: internal.TokenTypeIdentifier, Value: "varying"},
				{Type: internal.TokenTypeOperator, Value: "{"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},
				{Type: internal.TokenTypeOperator, Value: "}"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},
			},
		),
		Entry("uniform/varying block content",
			`
			color vec3
			intensity float
			`,
			[]internal.Token{
				{Type: internal.TokenTypeNewLine, Value: "\n"},

				{Type: internal.TokenTypeIdentifier, Value: "color"},
				{Type: internal.TokenTypeIdentifier, Value: "vec3"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},

				{Type: internal.TokenTypeIdentifier, Value: "intensity"},
				{Type: internal.TokenTypeIdentifier, Value: "float"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},
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
			[]internal.Token{
				{Type: internal.TokenTypeNewLine, Value: "\n"},

				{Type: internal.TokenTypeIdentifier, Value: "func"},
				{Type: internal.TokenTypeIdentifier, Value: "hello"},
				{Type: internal.TokenTypeOperator, Value: "("},
				{Type: internal.TokenTypeIdentifier, Value: "a"},
				{Type: internal.TokenTypeIdentifier, Value: "vec3"},
				{Type: internal.TokenTypeOperator, Value: ","},
				{Type: internal.TokenTypeIdentifier, Value: "b"},
				{Type: internal.TokenTypeIdentifier, Value: "float"},
				{Type: internal.TokenTypeOperator, Value: ")"},
				{Type: internal.TokenTypeOperator, Value: "("},
				{Type: internal.TokenTypeIdentifier, Value: "vec4"},
				{Type: internal.TokenTypeOperator, Value: ","},
				{Type: internal.TokenTypeIdentifier, Value: "float"},
				{Type: internal.TokenTypeOperator, Value: ")"},
				{Type: internal.TokenTypeOperator, Value: "{"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},
				{Type: internal.TokenTypeOperator, Value: "}"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},

				{Type: internal.TokenTypeIdentifier, Value: "func"},
				{Type: internal.TokenTypeIdentifier, Value: "world"},
				{Type: internal.TokenTypeOperator, Value: "("},
				{Type: internal.TokenTypeOperator, Value: ")"},
				{Type: internal.TokenTypeIdentifier, Value: "vec3"},
				{Type: internal.TokenTypeOperator, Value: "{"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},
				{Type: internal.TokenTypeOperator, Value: "}"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},

				{Type: internal.TokenTypeIdentifier, Value: "func"},
				{Type: internal.TokenTypeIdentifier, Value: "test"},
				{Type: internal.TokenTypeOperator, Value: "("},
				{Type: internal.TokenTypeOperator, Value: ")"},
				{Type: internal.TokenTypeOperator, Value: "{"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},
				{Type: internal.TokenTypeOperator, Value: "}"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},
			},
		),
		Entry("var declaration",
			`
			var a vec3
			var b float = 1.0
			var c float = -1.0
			`,
			[]internal.Token{
				{Type: internal.TokenTypeNewLine, Value: "\n"},

				{Type: internal.TokenTypeIdentifier, Value: "var"},
				{Type: internal.TokenTypeIdentifier, Value: "a"},
				{Type: internal.TokenTypeIdentifier, Value: "vec3"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},

				{Type: internal.TokenTypeIdentifier, Value: "var"},
				{Type: internal.TokenTypeIdentifier, Value: "b"},
				{Type: internal.TokenTypeIdentifier, Value: "float"},
				{Type: internal.TokenTypeOperator, Value: "="},
				{Type: internal.TokenTypeNumber, Value: "1.0"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},

				{Type: internal.TokenTypeIdentifier, Value: "var"},
				{Type: internal.TokenTypeIdentifier, Value: "c"},
				{Type: internal.TokenTypeIdentifier, Value: "float"},
				{Type: internal.TokenTypeOperator, Value: "="},
				{Type: internal.TokenTypeOperator, Value: "-"},
				{Type: internal.TokenTypeNumber, Value: "1.0"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},
			},
		),
		Entry("field access",
			`
			color.x=1.0
			color.y=-color.x
			`,
			[]internal.Token{
				{Type: internal.TokenTypeNewLine, Value: "\n"},

				{Type: internal.TokenTypeIdentifier, Value: "color"},
				{Type: internal.TokenTypeOperator, Value: "."},
				{Type: internal.TokenTypeIdentifier, Value: "x"},
				{Type: internal.TokenTypeOperator, Value: "="},
				{Type: internal.TokenTypeNumber, Value: "1.0"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},

				{Type: internal.TokenTypeIdentifier, Value: "color"},
				{Type: internal.TokenTypeOperator, Value: "."},
				{Type: internal.TokenTypeIdentifier, Value: "y"},
				{Type: internal.TokenTypeOperator, Value: "="},
				{Type: internal.TokenTypeOperator, Value: "-"},
				{Type: internal.TokenTypeIdentifier, Value: "color"},
				{Type: internal.TokenTypeOperator, Value: "."},
				{Type: internal.TokenTypeIdentifier, Value: "x"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},
			},
		),
		Entry("numeric expression",
			`
			x = (7.1 * ((13 / 2)%5.5)+3.0) - 5
			`,
			[]internal.Token{
				{Type: internal.TokenTypeNewLine, Value: "\n"},
				{Type: internal.TokenTypeIdentifier, Value: "x"},
				{Type: internal.TokenTypeOperator, Value: "="},
				{Type: internal.TokenTypeOperator, Value: "("},
				{Type: internal.TokenTypeNumber, Value: "7.1"},
				{Type: internal.TokenTypeOperator, Value: "*"},
				{Type: internal.TokenTypeOperator, Value: "("},
				{Type: internal.TokenTypeOperator, Value: "("},
				{Type: internal.TokenTypeNumber, Value: "13"},
				{Type: internal.TokenTypeOperator, Value: "/"},
				{Type: internal.TokenTypeNumber, Value: "2"},
				{Type: internal.TokenTypeOperator, Value: ")"},
				{Type: internal.TokenTypeOperator, Value: "%"},
				{Type: internal.TokenTypeNumber, Value: "5.5"},
				{Type: internal.TokenTypeOperator, Value: ")"},
				{Type: internal.TokenTypeOperator, Value: "+"},
				{Type: internal.TokenTypeNumber, Value: "3.0"},
				{Type: internal.TokenTypeOperator, Value: ")"},
				{Type: internal.TokenTypeOperator, Value: "-"},
				{Type: internal.TokenTypeNumber, Value: "5"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},
			},
		),
		Entry("bit-wise expression",
			`
			x = ((6 << 1) ^ (8 >> 2)) | (3 & 5)
			`,
			[]internal.Token{
				{Type: internal.TokenTypeNewLine, Value: "\n"},
				{Type: internal.TokenTypeIdentifier, Value: "x"},
				{Type: internal.TokenTypeOperator, Value: "="},
				{Type: internal.TokenTypeOperator, Value: "("},
				{Type: internal.TokenTypeOperator, Value: "("},
				{Type: internal.TokenTypeNumber, Value: "6"},
				{Type: internal.TokenTypeOperator, Value: "<<"},
				{Type: internal.TokenTypeNumber, Value: "1"},
				{Type: internal.TokenTypeOperator, Value: ")"},
				{Type: internal.TokenTypeOperator, Value: "^"},
				{Type: internal.TokenTypeOperator, Value: "("},
				{Type: internal.TokenTypeNumber, Value: "8"},
				{Type: internal.TokenTypeOperator, Value: ">>"},
				{Type: internal.TokenTypeNumber, Value: "2"},
				{Type: internal.TokenTypeOperator, Value: ")"},
				{Type: internal.TokenTypeOperator, Value: ")"},
				{Type: internal.TokenTypeOperator, Value: "|"},
				{Type: internal.TokenTypeOperator, Value: "("},
				{Type: internal.TokenTypeNumber, Value: "3"},
				{Type: internal.TokenTypeOperator, Value: "&"},
				{Type: internal.TokenTypeNumber, Value: "5"},
				{Type: internal.TokenTypeOperator, Value: ")"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},
			},
		),
		Entry("boolean expression",
			`
			x = ((x.z == z) && (5 != 4) && (1 < 2) && (1.1>3) && true) || (false&&(5 >= 9)&& (2 <= 1))
			`,
			[]internal.Token{
				{Type: internal.TokenTypeNewLine, Value: "\n"},
				{Type: internal.TokenTypeIdentifier, Value: "x"},
				{Type: internal.TokenTypeOperator, Value: "="},
				{Type: internal.TokenTypeOperator, Value: "("},
				{Type: internal.TokenTypeOperator, Value: "("},
				{Type: internal.TokenTypeIdentifier, Value: "x"},
				{Type: internal.TokenTypeOperator, Value: "."},
				{Type: internal.TokenTypeIdentifier, Value: "z"},
				{Type: internal.TokenTypeOperator, Value: "=="},
				{Type: internal.TokenTypeIdentifier, Value: "z"},
				{Type: internal.TokenTypeOperator, Value: ")"},
				{Type: internal.TokenTypeOperator, Value: "&&"},
				{Type: internal.TokenTypeOperator, Value: "("},
				{Type: internal.TokenTypeNumber, Value: "5"},
				{Type: internal.TokenTypeOperator, Value: "!="},
				{Type: internal.TokenTypeNumber, Value: "4"},
				{Type: internal.TokenTypeOperator, Value: ")"},
				{Type: internal.TokenTypeOperator, Value: "&&"},
				{Type: internal.TokenTypeOperator, Value: "("},
				{Type: internal.TokenTypeNumber, Value: "1"},
				{Type: internal.TokenTypeOperator, Value: "<"},
				{Type: internal.TokenTypeNumber, Value: "2"},
				{Type: internal.TokenTypeOperator, Value: ")"},
				{Type: internal.TokenTypeOperator, Value: "&&"},
				{Type: internal.TokenTypeOperator, Value: "("},
				{Type: internal.TokenTypeNumber, Value: "1.1"},
				{Type: internal.TokenTypeOperator, Value: ">"},
				{Type: internal.TokenTypeNumber, Value: "3"},
				{Type: internal.TokenTypeOperator, Value: ")"},
				{Type: internal.TokenTypeOperator, Value: "&&"},
				{Type: internal.TokenTypeIdentifier, Value: "true"},
				{Type: internal.TokenTypeOperator, Value: ")"},
				{Type: internal.TokenTypeOperator, Value: "||"},
				{Type: internal.TokenTypeOperator, Value: "("},
				{Type: internal.TokenTypeIdentifier, Value: "false"},
				{Type: internal.TokenTypeOperator, Value: "&&"},
				{Type: internal.TokenTypeOperator, Value: "("},
				{Type: internal.TokenTypeNumber, Value: "5"},
				{Type: internal.TokenTypeOperator, Value: ">="},
				{Type: internal.TokenTypeNumber, Value: "9"},
				{Type: internal.TokenTypeOperator, Value: ")"},
				{Type: internal.TokenTypeOperator, Value: "&&"},
				{Type: internal.TokenTypeOperator, Value: "("},
				{Type: internal.TokenTypeNumber, Value: "2"},
				{Type: internal.TokenTypeOperator, Value: "<="},
				{Type: internal.TokenTypeNumber, Value: "1"},
				{Type: internal.TokenTypeOperator, Value: ")"},
				{Type: internal.TokenTypeOperator, Value: ")"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},
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
			[]internal.Token{
				{Type: internal.TokenTypeNewLine, Value: "\n"},

				{Type: internal.TokenTypeIdentifier, Value: "a"},
				{Type: internal.TokenTypeOperator, Value: "="},
				{Type: internal.TokenTypeNumber, Value: "13"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},

				{Type: internal.TokenTypeIdentifier, Value: "b"},
				{Type: internal.TokenTypeOperator, Value: "="},
				{Type: internal.TokenTypeNumber, Value: "55.3"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},

				{Type: internal.TokenTypeIdentifier, Value: "c"},
				{Type: internal.TokenTypeOperator, Value: "+="},
				{Type: internal.TokenTypeNumber, Value: "5"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},

				{Type: internal.TokenTypeIdentifier, Value: "d"},
				{Type: internal.TokenTypeOperator, Value: "+="},
				{Type: internal.TokenTypeOperator, Value: "-"},
				{Type: internal.TokenTypeNumber, Value: "5"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},

				{Type: internal.TokenTypeIdentifier, Value: "e"},
				{Type: internal.TokenTypeOperator, Value: "-="},
				{Type: internal.TokenTypeNumber, Value: "5"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},

				{Type: internal.TokenTypeIdentifier, Value: "f"},
				{Type: internal.TokenTypeOperator, Value: "-="},
				{Type: internal.TokenTypeOperator, Value: "-"},
				{Type: internal.TokenTypeNumber, Value: "5"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},

				{Type: internal.TokenTypeIdentifier, Value: "g"},
				{Type: internal.TokenTypeOperator, Value: "*="},
				{Type: internal.TokenTypeNumber, Value: "5"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},

				{Type: internal.TokenTypeIdentifier, Value: "h"},
				{Type: internal.TokenTypeOperator, Value: "*="},
				{Type: internal.TokenTypeOperator, Value: "-"},
				{Type: internal.TokenTypeNumber, Value: "5"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},

				{Type: internal.TokenTypeIdentifier, Value: "i"},
				{Type: internal.TokenTypeOperator, Value: "/="},
				{Type: internal.TokenTypeNumber, Value: "5"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},

				{Type: internal.TokenTypeIdentifier, Value: "j"},
				{Type: internal.TokenTypeOperator, Value: "/="},
				{Type: internal.TokenTypeOperator, Value: "-"},
				{Type: internal.TokenTypeNumber, Value: "5"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},

				{Type: internal.TokenTypeIdentifier, Value: "k"},
				{Type: internal.TokenTypeOperator, Value: "%="},
				{Type: internal.TokenTypeNumber, Value: "5"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},

				{Type: internal.TokenTypeIdentifier, Value: "l"},
				{Type: internal.TokenTypeOperator, Value: "%="},
				{Type: internal.TokenTypeOperator, Value: "-"},
				{Type: internal.TokenTypeNumber, Value: "5"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},

				{Type: internal.TokenTypeIdentifier, Value: "m"},
				{Type: internal.TokenTypeOperator, Value: ">>="},
				{Type: internal.TokenTypeNumber, Value: "5"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},

				{Type: internal.TokenTypeIdentifier, Value: "n"},
				{Type: internal.TokenTypeOperator, Value: ">>="},
				{Type: internal.TokenTypeOperator, Value: "-"},
				{Type: internal.TokenTypeNumber, Value: "5"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},

				{Type: internal.TokenTypeIdentifier, Value: "o"},
				{Type: internal.TokenTypeOperator, Value: "<<="},
				{Type: internal.TokenTypeNumber, Value: "5"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},

				{Type: internal.TokenTypeIdentifier, Value: "p"},
				{Type: internal.TokenTypeOperator, Value: "<<="},
				{Type: internal.TokenTypeOperator, Value: "-"},
				{Type: internal.TokenTypeNumber, Value: "5"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},

				{Type: internal.TokenTypeIdentifier, Value: "q"},
				{Type: internal.TokenTypeOperator, Value: "&="},
				{Type: internal.TokenTypeNumber, Value: "5"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},

				{Type: internal.TokenTypeIdentifier, Value: "r"},
				{Type: internal.TokenTypeOperator, Value: "&="},
				{Type: internal.TokenTypeOperator, Value: "-"},
				{Type: internal.TokenTypeNumber, Value: "5"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},

				{Type: internal.TokenTypeIdentifier, Value: "s"},
				{Type: internal.TokenTypeOperator, Value: "^="},
				{Type: internal.TokenTypeNumber, Value: "5"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},

				{Type: internal.TokenTypeIdentifier, Value: "t"},
				{Type: internal.TokenTypeOperator, Value: "^="},
				{Type: internal.TokenTypeOperator, Value: "-"},
				{Type: internal.TokenTypeNumber, Value: "5"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},

				{Type: internal.TokenTypeIdentifier, Value: "u"},
				{Type: internal.TokenTypeOperator, Value: "|="},
				{Type: internal.TokenTypeNumber, Value: "5"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},

				{Type: internal.TokenTypeIdentifier, Value: "v"},
				{Type: internal.TokenTypeOperator, Value: "|="},
				{Type: internal.TokenTypeOperator, Value: "-"},
				{Type: internal.TokenTypeNumber, Value: "5"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},
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
			[]internal.Token{
				{Type: internal.TokenTypeNewLine, Value: "\n"},
				{Type: internal.TokenTypeIdentifier, Value: "doSomething"},
				{Type: internal.TokenTypeOperator, Value: "("},
				{Type: internal.TokenTypeNumber, Value: "15.55"},
				{Type: internal.TokenTypeOperator, Value: ","},
				{Type: internal.TokenTypeNumber, Value: "10"},
				{Type: internal.TokenTypeOperator, Value: ","},
				{Type: internal.TokenTypeIdentifier, Value: "someVar"},
				{Type: internal.TokenTypeOperator, Value: ")"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},
				{Type: internal.TokenTypeIdentifier, Value: "var"},
				{Type: internal.TokenTypeIdentifier, Value: "a"},
				{Type: internal.TokenTypeIdentifier, Value: "vec4"},
				{Type: internal.TokenTypeOperator, Value: "="},
				{Type: internal.TokenTypeIdentifier, Value: "calculate"},
				{Type: internal.TokenTypeOperator, Value: "("},
				{Type: internal.TokenTypeOperator, Value: ")"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},
				{Type: internal.TokenTypeIdentifier, Value: "var"},
				{Type: internal.TokenTypeIdentifier, Value: "b"},
				{Type: internal.TokenTypeIdentifier, Value: "float"},
				{Type: internal.TokenTypeOperator, Value: "="},
				{Type: internal.TokenTypeIdentifier, Value: "sum"},
				{Type: internal.TokenTypeOperator, Value: "("},
				{Type: internal.TokenTypeNewLine, Value: "\n"},
				{Type: internal.TokenTypeNumber, Value: "115.3"},
				{Type: internal.TokenTypeOperator, Value: ","},
				{Type: internal.TokenTypeNewLine, Value: "\n"},
				{Type: internal.TokenTypeIdentifier, Value: "abs"},
				{Type: internal.TokenTypeOperator, Value: "("},
				{Type: internal.TokenTypeOperator, Value: "-"},
				{Type: internal.TokenTypeNumber, Value: "15"},
				{Type: internal.TokenTypeOperator, Value: ")"},
				{Type: internal.TokenTypeOperator, Value: ","},
				{Type: internal.TokenTypeNewLine, Value: "\n"},
				{Type: internal.TokenTypeOperator, Value: ")"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},
			},
		),
	)
})
