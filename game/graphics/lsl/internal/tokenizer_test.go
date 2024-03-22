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
