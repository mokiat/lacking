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
		Entry("comments",
			`// This is a comment
			//This is another comment`,
			[]internal.Token{
				{Type: internal.TokenTypeComment, Value: "This is a comment"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},
				{Type: internal.TokenTypeComment, Value: "This is another comment"},
			},
		),
		Entry("uniform block",
			`uniform {
				color vec3
				intensity float
			}`,
			[]internal.Token{
				{Type: internal.TokenTypeIdentifier, Value: "uniform"},
				{Type: internal.TokenTypeOperator, Value: "{"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},
				{Type: internal.TokenTypeIdentifier, Value: "color"},
				{Type: internal.TokenTypeIdentifier, Value: "vec3"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},
				{Type: internal.TokenTypeIdentifier, Value: "intensity"},
				{Type: internal.TokenTypeIdentifier, Value: "float"},
				{Type: internal.TokenTypeNewLine, Value: "\n"},
				{Type: internal.TokenTypeOperator, Value: "}"},
			},
		),
	)
})
