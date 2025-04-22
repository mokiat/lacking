package lsl_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/lacking/game/graphics/lsl"
)

var _ = Describe("Parser", func() {
	DescribeTable("ParseNewLine", func(inSource string, expectedErr error) {
		parser := lsl.NewParser(inSource)
		err := parser.ParseNewLine()
		if expectedErr == nil {
			Expect(err).ToNot(HaveOccurred())
		} else {
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(expectedErr))
		}
	},
		Entry("new line",
			openTestFile("parser", "parse-new-line", "new-line.lsl"),
			nil,
		),
		Entry("new line after spacing",
			openTestFile("parser", "parse-new-line", "new-line-after-spacing.lsl"),
			nil,
		),
		Entry("no tokens",
			openTestFile("parser", "parse-new-line", "empty.lsl"),
			&lsl.ParseError{
				Pos:     lsl.At(1, 1),
				Message: "expected a new line token",
			},
		),
		Entry("identifier",
			openTestFile("parser", "parse-new-line", "identifier.lsl"),
			&lsl.ParseError{
				Pos:     lsl.At(1, 1),
				Message: "expected a new line token",
			},
		),
	)

	DescribeTable("ParseComment", func(inSource string, expectedErr error) {
		parser := lsl.NewParser(inSource)
		err := parser.ParseComment()
		if expectedErr == nil {
			Expect(err).ToNot(HaveOccurred())
		} else {
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(expectedErr))
		}
	},
		Entry("plain comment",
			openTestFile("parser", "parse-comment", "plain-comment.lsl"),
			nil,
		),
		Entry("comment after spacing",
			openTestFile("parser", "parse-comment", "comment-after-spacing.lsl"),
			nil,
		),
		Entry("comment with new line",
			openTestFile("parser", "parse-comment", "comment-with-new-line.lsl"),
			nil,
		),
		Entry("operator",
			openTestFile("parser", "parse-comment", "operator.lsl"),
			&lsl.ParseError{
				Pos:     lsl.At(1, 1),
				Message: "expected a comment token",
			},
		),
	)

	DescribeTable("ParseOptionalRemainder", func(inSource string, expectedErr error) {
		parser := lsl.NewParser(inSource)
		err := parser.ParseOptionalRemainder()
		if expectedErr == nil {
			Expect(err).ToNot(HaveOccurred())
		} else {
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(expectedErr))
		}
	},
		Entry("empty",
			openTestFile("parser", "parse-optional-remainder", "empty.lsl"),
			nil,
		),
		Entry("new line",
			openTestFile("parser", "parse-optional-remainder", "new-line.lsl"),
			nil,
		),
		Entry("comment",
			openTestFile("parser", "parse-optional-remainder", "comment.lsl"),
			nil,
		),
		Entry("vital token",
			openTestFile("parser", "parse-optional-remainder", "vital-token.lsl"),
			&lsl.ParseError{
				Pos:     lsl.At(1, 1),
				Message: "expected a comment, new line or end of file",
			},
		),
	)

	DescribeTable("ParseBlockStart", func(inSource string, expectedErr error) {
		parser := lsl.NewParser(inSource)
		err := parser.ParseBlockStart()
		if expectedErr == nil {
			Expect(err).ToNot(HaveOccurred())
		} else {
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(expectedErr))
		}
	},
		Entry("opening bracket",
			openTestFile("parser", "parse-block-start", "opening-bracket.lsl"),
			nil,
		),
		Entry("with new line",
			openTestFile("parser", "parse-block-start", "with-new-line.lsl"),
			nil,
		),
		Entry("with comment",
			openTestFile("parser", "parse-block-start", "with-comment.lsl"),
			nil,
		),
		Entry("not opening bracket",
			openTestFile("parser", "parse-block-start", "not-opening-bracket.lsl"),
			&lsl.ParseError{
				Pos:     lsl.At(1, 1),
				Message: "expected an opening bracket",
			},
		),
	)

	DescribeTable("ParseBlockEnd", func(inSource string, expectedErr error) {
		parser := lsl.NewParser(inSource)
		err := parser.ParseBlockEnd()
		if expectedErr == nil {
			Expect(err).ToNot(HaveOccurred())
		} else {
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(expectedErr))
		}
	},
		Entry("closing bracket",
			openTestFile("parser", "parse-block-end", "closing-bracket.lsl"),
			nil,
		),
		Entry("with new line",
			openTestFile("parser", "parse-block-end", "with-new-line.lsl"),
			nil,
		),
		Entry("with comment",
			openTestFile("parser", "parse-block-end", "with-comment.lsl"),
			nil,
		),
		Entry("not closing bracket",
			openTestFile("parser", "parse-block-end", "not-closing-bracket.lsl"),
			&lsl.ParseError{
				Pos:     lsl.At(1, 1),
				Message: "expected a closing bracket",
			},
		),
	)

	DescribeTable("ParseNamedParameterList", func(inSource string, expectedFields []lsl.Field, expectedErr error) {
		parser := lsl.NewParser(inSource)
		fields, err := parser.ParseNamedParameterList()
		if expectedErr == nil {
			Expect(err).ToNot(HaveOccurred())
			Expect(fields).To(Equal(expectedFields))
		} else {
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(expectedErr))
		}
	},
		Entry("empty list",
			openTestFile("parser", "parse-named-parameter-list", "empty.lsl"),
			nil,
			nil,
		),
		Entry("single parameter",
			openTestFile("parser", "parse-named-parameter-list", "single-parameter.lsl"),
			[]lsl.Field{
				{
					Pos:  lsl.At(1, 1),
					Name: "color",
					Type: "vec4",
				},
			},
			nil,
		),
		Entry("multiple parameters, single line",
			openTestFile("parser", "parse-named-parameter-list", "multiple-parameters-single-line.lsl"),
			[]lsl.Field{
				{
					Pos:  lsl.At(1, 4),
					Name: "color",
					Type: "vec4",
				},
				{
					Pos:  lsl.At(1, 18),
					Name: "intensity",
					Type: "float",
				},
			},
			nil,
		),
		Entry("multiple parameters, multiple lines",
			openTestFile("parser", "parse-named-parameter-list", "multiple-parameters-multiple-lines.lsl"),
			[]lsl.Field{
				{
					Pos:  lsl.At(2, 1),
					Name: "color",
					Type: "vec4",
				},
				{
					Pos:  lsl.At(5, 1),
					Name: "intensity",
					Type: "float",
				},
			},
			nil,
		),
		Entry("ending on a non-comma operator",
			openTestFile("parser", "parse-named-parameter-list", "ending-on-non-comma-operator.lsl"),
			[]lsl.Field{
				{
					Pos:  lsl.At(1, 1),
					Name: "color",
					Type: "vec4",
				},
			},
			nil,
		),
		Entry("ending on a comma operator",
			openTestFile("parser", "parse-named-parameter-list", "ending-on-comma-operator.lsl"),
			[]lsl.Field{
				{
					Pos:  lsl.At(1, 1),
					Name: "color",
					Type: "vec4",
				},
			},
			nil,
		),
		Entry("ending on a twin-comma operator",
			openTestFile("parser", "parse-named-parameter-list", "ending-on-twin-comma-operator.lsl"),
			nil,
			&lsl.ParseError{
				Pos:     lsl.At(1, 12),
				Message: "unexpected comma",
			},
		),
		Entry("non-identifier name",
			openTestFile("parser", "parse-named-parameter-list", "non-identifier-name.lsl"),
			nil,
			&lsl.ParseError{
				Pos:     lsl.At(1, 1),
				Message: "expected a name identifier or end of list",
			},
		),
		Entry("non-identifier type",
			openTestFile("parser", "parse-named-parameter-list", "non-identifier-type.lsl"),
			nil,
			&lsl.ParseError{
				Pos:     lsl.At(1, 7),
				Message: "expected a type identifier",
			},
		),
		Entry("non-comma or operator after type",
			openTestFile("parser", "parse-named-parameter-list", "non-comma-or-operator-after-type.lsl"),
			nil,
			&lsl.ParseError{
				Pos:     lsl.At(1, 12),
				Message: "expected a comma or end of list",
			},
		),
	)

	DescribeTable("ParseTextureBlock", func(inSource string, expectedBlock *lsl.TextureBlockDeclaration, expectedErr error) {
		parser := lsl.NewParser(inSource)
		block, err := parser.ParseTextureBlock()
		if expectedErr == nil {
			Expect(err).ToNot(HaveOccurred())
			Expect(block).To(Equal(expectedBlock))
		} else {
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(expectedErr))
		}
	},
		Entry("empty block",
			openTestFile("parser", "parse-texture-block", "empty.lsl"),
			&lsl.TextureBlockDeclaration{
				Pos:    lsl.At(1, 1),
				Fields: nil,
			},
			nil,
		),
		Entry("with fields",
			openTestFile("parser", "parse-texture-block", "with-fields.lsl"),
			&lsl.TextureBlockDeclaration{
				Pos: lsl.At(1, 1),
				Fields: []lsl.Field{
					{
						Pos:  lsl.At(2, 3),
						Name: "first",
						Type: "sampler2D",
					},
					{
						Pos:  lsl.At(3, 3),
						Name: "second",
						Type: "samplerCube",
					},
				},
			},
			nil,
		),
		Entry("with comments and spaces",
			openTestFile("parser", "parse-texture-block", "with-comments-and-spaces.lsl"),
			&lsl.TextureBlockDeclaration{
				Pos: lsl.At(1, 1),
				Fields: []lsl.Field{
					{
						Pos:  lsl.At(3, 1),
						Name: "first",
						Type: "sampler2D",
					},
					{
						Pos:  lsl.At(5, 1),
						Name: "second",
						Type: "samplerCube",
					},
				},
			},
			nil,
		),
		Entry("closing on same line",
			openTestFile("parser", "parse-texture-block", "closing-on-same-line.lsl"),
			nil,
			&lsl.ParseError{
				Pos:     lsl.At(1, 11),
				Message: "expected a comment, new line or end of file",
			},
		),
		Entry("other block type",
			openTestFile("parser", "parse-texture-block", "other-block-type.lsl"),
			nil,
			&lsl.ParseError{
				Pos:     lsl.At(1, 1),
				Message: "expected 'textures' keyword",
			},
		),
	)

	DescribeTable("ParseUniformBlock", func(inSource string, expectedBlock *lsl.UniformBlockDeclaration, expectedErr error) {
		parser := lsl.NewParser(inSource)
		block, err := parser.ParseUniformBlock()
		if expectedErr == nil {
			Expect(err).ToNot(HaveOccurred())
			Expect(block).To(Equal(expectedBlock))
		} else {
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(expectedErr))
		}
	},
		Entry("empty block",
			openTestFile("parser", "parse-uniform-block", "empty.lsl"),
			&lsl.UniformBlockDeclaration{
				Pos:    lsl.At(1, 1),
				Fields: nil,
			},
			nil,
		),
		Entry("with fields",
			openTestFile("parser", "parse-uniform-block", "with-fields.lsl"),
			&lsl.UniformBlockDeclaration{
				Pos: lsl.At(1, 1),
				Fields: []lsl.Field{
					{
						Pos:  lsl.At(2, 3),
						Name: "color",
						Type: "vec4",
					},
					{
						Pos:  lsl.At(3, 3),
						Name: "intensity",
						Type: "float",
					},
				},
			},
			nil,
		),
		Entry("with comments and spaces",
			openTestFile("parser", "parse-uniform-block", "with-comments-and-spaces.lsl"),
			&lsl.UniformBlockDeclaration{
				Pos: lsl.At(1, 1),
				Fields: []lsl.Field{
					{
						Pos:  lsl.At(3, 1),
						Name: "color",
						Type: "vec4",
					},
					{
						Pos:  lsl.At(5, 1),
						Name: "intensity",
						Type: "float",
					},
				},
			},
			nil,
		),
		Entry("closing on same line",
			openTestFile("parser", "parse-uniform-block", "closing-on-same-line.lsl"),
			nil,
			&lsl.ParseError{
				Pos:     lsl.At(1, 11),
				Message: "expected a comment, new line or end of file",
			},
		),
		Entry("other block type",
			openTestFile("parser", "parse-uniform-block", "other-block-type.lsl"),
			nil,
			&lsl.ParseError{
				Pos:     lsl.At(1, 1),
				Message: "expected 'uniforms' keyword",
			},
		),
	)

	DescribeTable("ParseVaryingBlock", func(inSource string, expectedBlock *lsl.VaryingBlockDeclaration, expectedErr error) {
		parser := lsl.NewParser(inSource)
		block, err := parser.ParseVaryingBlock()
		if expectedErr == nil {
			Expect(err).ToNot(HaveOccurred())
			Expect(block).To(Equal(expectedBlock))
		} else {
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(expectedErr))
		}
	},
		Entry("empty",
			openTestFile("parser", "parse-varying-block", "empty.lsl"),
			&lsl.VaryingBlockDeclaration{
				Pos:    lsl.At(1, 1),
				Fields: nil,
			},
			nil,
		),
		Entry("with fields",
			openTestFile("parser", "parse-varying-block", "with-fields.lsl"),
			&lsl.VaryingBlockDeclaration{
				Pos: lsl.At(1, 1),
				Fields: []lsl.Field{
					{
						Pos:  lsl.At(2, 3),
						Name: "color",
						Type: "vec4",
					},
					{
						Pos:  lsl.At(3, 3),
						Name: "intensity",
						Type: "float",
					},
				},
			},
			nil,
		),
		Entry("with comments and spaces",
			openTestFile("parser", "parse-varying-block", "with-comments-and-spaces.lsl"),
			&lsl.VaryingBlockDeclaration{
				Pos: lsl.At(1, 1),
				Fields: []lsl.Field{
					{
						Pos:  lsl.At(3, 1),
						Name: "color",
						Type: "vec4",
					},
					{
						Pos:  lsl.At(5, 1),
						Name: "intensity",
						Type: "float",
					},
				},
			},
			nil,
		),
		Entry("closing on same line",
			openTestFile("parser", "parse-varying-block", "closing-on-same-line.lsl"),
			nil,
			&lsl.ParseError{
				Pos:     lsl.At(1, 11),
				Message: "expected a comment, new line or end of file",
			},
		),
		Entry("other block type",
			openTestFile("parser", "parse-varying-block", "other-block-type.lsl"),
			nil,
			&lsl.ParseError{
				Pos:     lsl.At(1, 1),
				Message: "expected 'varyings' keyword",
			},
		),
	)

	DescribeTable("ParseExpression", func(inSource string, expectedExp lsl.Expression, expectedErr error) {
		parser := lsl.NewParser(inSource)
		exp, err := parser.ParseExpression()
		if expectedErr == nil {
			Expect(err).ToNot(HaveOccurred())
			Expect(exp).To(Equal(expectedExp))
		} else {
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(expectedErr))
		}
	},
		Entry("bool literal",
			openTestFile("parser", "parse-expression", "bool-literal.lsl"),
			&lsl.BoolLiteral{
				Pos:   lsl.At(1, 1),
				Value: true,
			},
			nil,
		),
		Entry("float literal",
			openTestFile("parser", "parse-expression", "float-literal.lsl"),
			&lsl.FloatLiteral{
				Pos:   lsl.At(1, 1),
				Value: 5.3,
			},
			nil,
		),
		Entry("int literal",
			openTestFile("parser", "parse-expression", "int-literal.lsl"),
			&lsl.IntLiteral{
				Pos:   lsl.At(1, 1),
				Value: 3999,
			},
			nil,
		),
		Entry("unary (+) operator",
			openTestFile("parser", "parse-expression", "unary-plus-operator.lsl"),
			&lsl.UnaryExpression{
				Pos:      lsl.At(1, 1),
				Operator: "+",
				Operand: &lsl.IntLiteral{
					Pos:   lsl.At(1, 2),
					Value: 10,
				},
			},
			nil,
		),
		Entry("unary (-) operator",
			openTestFile("parser", "parse-expression", "unary-minus-operator.lsl"),
			&lsl.UnaryExpression{
				Pos:      lsl.At(1, 1),
				Operator: "-",
				Operand: &lsl.IntLiteral{
					Pos:   lsl.At(1, 2),
					Value: 10,
				},
			},
			nil,
		),
		Entry("unary (^) operator",
			openTestFile("parser", "parse-expression", "unary-bit-not-operator.lsl"),
			&lsl.UnaryExpression{
				Pos:      lsl.At(1, 1),
				Operator: "^",
				Operand: &lsl.IntLiteral{
					Pos:   lsl.At(1, 2),
					Value: 10,
				},
			},
			nil,
		),
		Entry("unary (!) operator",
			openTestFile("parser", "parse-expression", "unary-not-operator.lsl"),
			&lsl.UnaryExpression{
				Pos:      lsl.At(1, 1),
				Operator: "!",
				Operand: &lsl.IntLiteral{
					Pos:   lsl.At(1, 2),
					Value: 10,
				},
			},
			nil,
		),
		Entry("identifier",
			openTestFile("parser", "parse-expression", "identifier.lsl"),
			&lsl.Identifier{
				Pos:  lsl.At(1, 1),
				Name: "hello",
			},
			nil,
		),
		Entry("field identifier",
			openTestFile("parser", "parse-expression", "field-identifier.lsl"),
			&lsl.FieldIdentifier{
				Owner: &lsl.Identifier{
					Pos:  lsl.At(1, 1),
					Name: "hello",
				},
				Field: lsl.Identifier{
					Pos:  lsl.At(1, 7),
					Name: "world",
				},
			},
			nil,
		),
		Entry("nested field identifier",
			openTestFile("parser", "parse-expression", "nested-field-identifier.lsl"),
			&lsl.FieldIdentifier{
				Owner: &lsl.FieldIdentifier{
					Owner: &lsl.Identifier{
						Pos:  lsl.At(1, 1),
						Name: "first",
					},
					Field: lsl.Identifier{
						Pos:  lsl.At(1, 7),
						Name: "second",
					},
				},
				Field: lsl.Identifier{
					Pos:  lsl.At(1, 14),
					Name: "third",
				},
			},
			nil,
		),
		Entry("function call",
			openTestFile("parser", "parse-expression", "function-call.lsl"),
			&lsl.FunctionCall{
				Owner: &lsl.Identifier{
					Pos:  lsl.At(1, 1),
					Name: "rand",
				},
				Arguments: nil,
			},
			nil,
		),
		Entry("nested function call",
			openTestFile("parser", "parse-expression", "nested-function-call.lsl"),
			&lsl.FunctionCall{
				Owner: &lsl.FieldIdentifier{
					Owner: &lsl.Identifier{
						Pos:  lsl.At(1, 1),
						Name: "first",
					},
					Field: lsl.Identifier{
						Pos:  lsl.At(1, 7),
						Name: "second",
					},
				},
				Arguments: nil,
			},
			nil,
		),
		Entry("function call with args",
			openTestFile("parser", "parse-expression", "function-call-with-args.lsl"),
			&lsl.FunctionCall{
				Owner: &lsl.Identifier{
					Pos:  lsl.At(1, 1),
					Name: "test",
				},
				Arguments: []lsl.Expression{
					&lsl.IntLiteral{
						Pos:   lsl.At(1, 6),
						Value: 200,
					},
					&lsl.FloatLiteral{
						Pos:   lsl.At(1, 11),
						Value: 1.5,
					},
				},
			},
			nil,
		),
		Entry("function call with args and new lines",
			openTestFile("parser", "parse-expression", "function-call-with-args-and-new-lines.lsl"),
			&lsl.FunctionCall{
				Owner: &lsl.Identifier{
					Pos:  lsl.At(1, 1),
					Name: "test",
				},
				Arguments: []lsl.Expression{
					&lsl.IntLiteral{
						Pos:   lsl.At(2, 3),
						Value: 200,
					},
					&lsl.FloatLiteral{
						Pos:   lsl.At(3, 3),
						Value: 1.5,
					},
				},
			},
			nil,
		),
		Entry("function call with args and comments",
			openTestFile("parser", "parse-expression", "function-call-with-args-and-comments.lsl"),
			&lsl.FunctionCall{
				Owner: &lsl.Identifier{
					Pos:  lsl.At(1, 1),
					Name: "test",
				},
				Arguments: []lsl.Expression{
					&lsl.IntLiteral{
						Pos:   lsl.At(2, 3),
						Value: 200,
					},
					&lsl.FloatLiteral{
						Pos:   lsl.At(4, 3),
						Value: 1.5,
					},
				},
			},
			nil,
		),
		Entry("binary expression (numbers)",
			openTestFile("parser", "parse-expression", "binary-expression-numbers.lsl"),
			&lsl.BinaryExpression{
				Operator: "+",
				Left: &lsl.IntLiteral{
					Pos:   lsl.At(1, 1),
					Value: 1,
				},
				Right: &lsl.FloatLiteral{
					Pos:   lsl.At(1, 5),
					Value: 2.3,
				},
			},
			nil,
		),
		Entry("binary expression (identifiers)",
			openTestFile("parser", "parse-expression", "binary-expression-identifiers.lsl"),
			&lsl.BinaryExpression{
				Operator: "+",
				Left: &lsl.Identifier{
					Pos:  lsl.At(1, 1),
					Name: "amount",
				},
				Right: &lsl.FieldIdentifier{
					Owner: &lsl.Identifier{
						Pos:  lsl.At(1, 10),
						Name: "color",
					},
					Field: lsl.Identifier{
						Pos:  lsl.At(1, 16),
						Name: "x",
					},
				},
			},
			nil,
		),
		Entry("binary expression (functions)",
			openTestFile("parser", "parse-expression", "binary-expression-functions.lsl"),
			&lsl.BinaryExpression{
				Operator: "*",
				Left: &lsl.FunctionCall{
					Owner: &lsl.Identifier{
						Pos:  lsl.At(1, 1),
						Name: "first",
					},
				},
				Right: &lsl.FunctionCall{
					Owner: &lsl.Identifier{
						Pos:  lsl.At(1, 11),
						Name: "second",
					},
				},
			},
			nil,
		),
		Entry("complex expression",
			openTestFile("parser", "parse-expression", "complex-expression.lsl"),
			&lsl.BinaryExpression{
				Operator: "+",
				Left: &lsl.FloatLiteral{
					Pos:   lsl.At(1, 1),
					Value: 5.5,
				},
				Right: &lsl.BinaryExpression{
					Operator: "*",
					Left: &lsl.Identifier{
						Pos:  lsl.At(1, 7),
						Name: "hello",
					},
					Right: &lsl.BinaryExpression{
						Operator: "-",
						Left: &lsl.BinaryExpression{
							Operator: "/",
							Left: &lsl.IntLiteral{
								Pos:   lsl.At(1, 16),
								Value: 13,
							},
							Right: &lsl.IntLiteral{
								Pos:   lsl.At(1, 21),
								Value: 2,
							},
						},
						Right: &lsl.IntLiteral{
							Pos:   lsl.At(1, 25),
							Value: 77,
						},
					},
				},
			},
			nil,
		),
		Entry("logical expression",
			openTestFile("parser", "parse-expression", "logical-expression.lsl"),
			&lsl.BinaryExpression{
				Left: &lsl.BinaryExpression{
					Left: &lsl.BinaryExpression{
						Left: &lsl.BinaryExpression{
							Left: &lsl.FloatLiteral{
								Pos:   lsl.At(1, 4),
								Value: 5.0,
							},
							Operator: ">",
							Right: &lsl.IntLiteral{
								Pos:   lsl.At(1, 10),
								Value: 10,
							},
						},
						Operator: "||",
						Right: &lsl.BinaryExpression{
							Left: &lsl.Identifier{
								Pos:  lsl.At(1, 18),
								Name: "b",
							},
							Operator: ">=",
							Right: &lsl.Identifier{
								Pos:  lsl.At(1, 23),
								Name: "a",
							},
						},
					},
					Operator: "&&",
					Right: &lsl.UnaryExpression{
						Pos:      lsl.At(1, 30),
						Operator: "!",
						Operand: &lsl.Identifier{
							Pos:  lsl.At(1, 31),
							Name: "c",
						},
					},
				},
				Operator: "==",
				Right: &lsl.Identifier{
					Pos:  lsl.At(1, 37),
					Name: "d",
				},
			},
			nil,
		),
		Entry("operator precedence",
			openTestFile("parser", "parse-expression", "operator-precedence.lsl"),
			&lsl.BinaryExpression{
				Left: &lsl.BinaryExpression{
					Left: &lsl.BinaryExpression{
						Left: &lsl.BinaryExpression{
							Left: &lsl.BinaryExpression{
								Left: &lsl.IntLiteral{
									Pos:   lsl.At(1, 1),
									Value: 10,
								},
								Operator: "*",
								Right: &lsl.BinaryExpression{
									Left: &lsl.BinaryExpression{
										Left: &lsl.IntLiteral{
											Pos:   lsl.At(1, 7),
											Value: 5,
										},
										Operator: "+",
										Right: &lsl.IntLiteral{
											Pos:   lsl.At(1, 11),
											Value: 3,
										},
									},
									Operator: "-",
									Right: &lsl.IntLiteral{
										Pos:   lsl.At(1, 15),
										Value: 2,
									},
								},
							},
							Operator: "|",
							Right: &lsl.BinaryExpression{
								Left: &lsl.IntLiteral{
									Pos:   lsl.At(1, 20),
									Value: 7,
								},
								Operator: "/",
								Right: &lsl.IntLiteral{
									Pos:   lsl.At(1, 24),
									Value: 3,
								},
							},
						},
						Operator: ">=",
						Right: &lsl.IntLiteral{
							Pos:   lsl.At(1, 29),
							Value: 0,
						},
					},
					Operator: "&&",
					Right: &lsl.BoolLiteral{
						Pos:   lsl.At(1, 34),
						Value: true,
					},
				},
				Operator: "||",
				Right: &lsl.BoolLiteral{
					Pos:   lsl.At(1, 42),
					Value: false,
				},
			},
			nil,
		),
		Entry("with comments and spaces",
			openTestFile("parser", "parse-expression", "with-comments-and-spaces.lsl"),
			&lsl.BinaryExpression{
				Left: &lsl.BinaryExpression{
					Left: &lsl.UnaryExpression{
						Pos:      lsl.At(1, 5),
						Operator: "^",
						Operand: &lsl.Identifier{
							Pos:  lsl.At(1, 6),
							Name: "a",
						},
					},
					Operator: ">=",
					Right: &lsl.IntLiteral{
						Pos:   lsl.At(1, 13),
						Value: 10,
					},
				},
				Operator: "&&",
				Right: &lsl.BoolLiteral{
					Pos:   lsl.At(2, 4),
					Value: true,
				},
			},
			nil,
		),
		Entry("starts with binary operator",
			openTestFile("parser", "parse-expression", "starts-with-binary-operator.lsl"),
			nil,
			&lsl.ParseError{
				Pos:     lsl.At(1, 1),
				Message: "expected an expression value",
			},
		),
		Entry("incomplete",
			openTestFile("parser", "parse-expression", "incomplete.lsl"),
			nil,
			&lsl.ParseError{
				Pos:     lsl.At(1, 4),
				Message: "expected an expression value",
			},
		),
		Entry("invalid value",
			openTestFile("parser", "parse-expression", "invalid-value.lsl"),
			nil,
			&lsl.ParseError{
				Pos:     lsl.At(1, 5),
				Message: "expected an expression value",
			},
		),
	)

	DescribeTable("ParseStatement", func(inSource string, expectedStmt lsl.Statement, expectedErr error) {
		parser := lsl.NewParser(inSource)
		stmt, err := parser.ParseStatement()
		if expectedErr == nil {
			Expect(err).ToNot(HaveOccurred())
			Expect(stmt).To(Equal(expectedStmt))
		} else {
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(expectedErr))
		}
	},
		Entry("discard",
			openTestFile("parser", "parse-statement", "discard.lsl"),
			&lsl.Discard{
				Pos: lsl.At(1, 1),
			},
			nil,
		),
		// TODO: Add more tests for statements.
	)

	DescribeTable("ParseFunction", func(inSource string, expectedDecl *lsl.FunctionDeclaration) {
		parser := lsl.NewParser(inSource)
		decl, err := parser.ParseFunction()
		Expect(err).ToNot(HaveOccurred())
		Expect(decl).To(Equal(expectedDecl))
	},
		Entry("simple",
			openTestFile("parser", "parse-function", "simple.lsl"),
			&lsl.FunctionDeclaration{
				Pos:        lsl.At(1, 1),
				Name:       "test",
				Inputs:     nil,
				OutputType: "",
				Body:       nil,
			},
		),
		Entry("with inputs",
			openTestFile("parser", "parse-function", "with-inputs.lsl"),
			&lsl.FunctionDeclaration{
				Pos:  lsl.At(1, 1),
				Name: "test",
				Inputs: []lsl.Field{
					{
						Pos:  lsl.At(1, 11),
						Name: "color",
						Type: "vec3",
					},
					{
						Pos:  lsl.At(1, 23),
						Name: "intensity",
						Type: "float",
					},
				},
				OutputType: "",
				Body:       nil,
			},
		),
		Entry("with inputs on new lines",
			openTestFile("parser", "parse-function", "with-inputs-on-new-lines.lsl"),
			&lsl.FunctionDeclaration{
				Pos:  lsl.At(1, 1),
				Name: "test",
				Inputs: []lsl.Field{
					{
						Pos:  lsl.At(3, 3),
						Name: "color",
						Type: "vec3",
					},
					{
						Pos:  lsl.At(5, 3),
						Name: "intensity",
						Type: "float",
					},
				},
				OutputType: "",
				Body:       nil,
			},
		),
		Entry("with single output",
			openTestFile("parser", "parse-function", "with-simple-output.lsl"),
			&lsl.FunctionDeclaration{
				Pos:        lsl.At(1, 1),
				Name:       "test",
				Inputs:     nil,
				OutputType: "vec3",
				Body:       nil,
			},
		),
		Entry("with inputs and output",
			openTestFile("parser", "parse-function", "with-inputs-and-output.lsl"),
			&lsl.FunctionDeclaration{
				Pos:  lsl.At(1, 1),
				Name: "test",
				Inputs: []lsl.Field{
					{
						Pos:  lsl.At(1, 11),
						Name: "color",
						Type: "vec3",
					},
					{
						Pos:  lsl.At(1, 23),
						Name: "intensity",
						Type: "float",
					},
				},
				OutputType: "vec3",
				Body:       nil,
			},
		),
		Entry("with comment in body",
			openTestFile("parser", "parse-function", "with-comment-in-body.lsl"),
			&lsl.FunctionDeclaration{
				Pos:        lsl.At(1, 1),
				Name:       "test",
				Inputs:     nil,
				OutputType: "",
				Body:       nil,
			},
		),
		Entry("with function call statements",
			openTestFile("parser", "parse-function", "with-function-call-statements.lsl"),
			&lsl.FunctionDeclaration{
				Pos:        lsl.At(1, 1),
				Name:       "test",
				Inputs:     nil,
				OutputType: "",
				Body: lsl.StatementList{
					&lsl.FunctionCall{
						Owner: &lsl.Identifier{
							Pos:  lsl.At(2, 3),
							Name: "doFirst",
						},
					},
					&lsl.FunctionCall{
						Owner: &lsl.Identifier{
							Pos:  lsl.At(3, 3),
							Name: "doSecond",
						},
					},
				},
			},
		),
		Entry("with variable declarations",
			openTestFile("parser", "parse-function", "with-variable-declarations.lsl"),
			&lsl.FunctionDeclaration{
				Pos:        lsl.At(1, 1),
				Name:       "test",
				Inputs:     nil,
				OutputType: "",
				Body: lsl.StatementList{
					&lsl.VariableDeclaration{
						Pos:  lsl.At(2, 3),
						Name: "x",
						Type: "float",
						Assignment: &lsl.FloatLiteral{
							Pos:   lsl.At(2, 17),
							Value: 5.3,
						},
					},
					&lsl.VariableDeclaration{
						Pos:  lsl.At(3, 3),
						Name: "y",
						Type: "int",
						Assignment: &lsl.IntLiteral{
							Pos:   lsl.At(3, 15),
							Value: 3,
						},
					},
					&lsl.VariableDeclaration{
						Pos:  lsl.At(4, 3),
						Name: "z",
						Type: "vec3",
						Assignment: &lsl.FunctionCall{
							Owner: &lsl.Identifier{
								Pos:  lsl.At(4, 16),
								Name: "vec3",
							},
							Arguments: []lsl.Expression{
								&lsl.FloatLiteral{
									Pos:   lsl.At(4, 21),
									Value: 1.0,
								},
								&lsl.FloatLiteral{
									Pos:   lsl.At(4, 26),
									Value: 0.0,
								},
								&lsl.UnaryExpression{
									Pos:      lsl.At(4, 31),
									Operator: "-",
									Operand: &lsl.FloatLiteral{
										Pos:   lsl.At(4, 32),
										Value: 0.5,
									},
								},
							},
						},
					},
					&lsl.VariableDeclaration{
						Pos:  lsl.At(5, 3),
						Name: "w",
						Type: "", // auto-assignment
						Assignment: &lsl.IntLiteral{
							Pos:   lsl.At(5, 8),
							Value: 15,
						},
					},
				},
			},
		),
		Entry("with variable assignments",
			openTestFile("parser", "parse-function", "with-variable-assignments.lsl"),
			&lsl.FunctionDeclaration{
				Pos:        lsl.At(1, 1),
				Name:       "test",
				Inputs:     nil,
				OutputType: "",
				Body: lsl.StatementList{
					&lsl.Assignment{
						Target: &lsl.FieldIdentifier{
							Owner: &lsl.Identifier{
								Pos:  lsl.At(2, 3),
								Name: "color",
							},
							Field: lsl.Identifier{
								Pos:  lsl.At(2, 9),
								Name: "x",
							},
						},
						Operator: "+=",
						Expression: &lsl.FloatLiteral{
							Pos:   lsl.At(2, 14),
							Value: 5.3,
						},
					},
					&lsl.Assignment{
						Target: &lsl.FieldIdentifier{
							Owner: &lsl.Identifier{
								Pos:  lsl.At(3, 3),
								Name: "color",
							},
							Field: lsl.Identifier{
								Pos:  lsl.At(3, 9),
								Name: "y",
							},
						},
						Operator: "*=",
						Expression: &lsl.IntLiteral{
							Pos:   lsl.At(3, 14),
							Value: 3,
						},
					},
					&lsl.Assignment{
						Target: &lsl.Identifier{
							Pos:  lsl.At(4, 3),
							Name: "z",
						},
						Operator: "=",
						Expression: &lsl.FunctionCall{
							Owner: &lsl.Identifier{
								Pos:  lsl.At(4, 7),
								Name: "vec3",
							},
							Arguments: []lsl.Expression{
								&lsl.FloatLiteral{
									Pos:   lsl.At(4, 12),
									Value: 1.0,
								},
								&lsl.FloatLiteral{
									Pos:   lsl.At(4, 17),
									Value: 0.0,
								},
								&lsl.UnaryExpression{
									Pos:      lsl.At(4, 22),
									Operator: "-",
									Operand: &lsl.FloatLiteral{
										Pos:   lsl.At(4, 23),
										Value: 0.5,
									},
								},
							},
						},
					},
				},
			},
		),
		Entry("with conditionals",
			openTestFile("parser", "parse-function", "with-conditionals.lsl"),
			&lsl.FunctionDeclaration{
				Pos:        lsl.At(1, 1),
				Name:       "test",
				Inputs:     nil,
				OutputType: "",
				Body: lsl.StatementList{
					&lsl.Conditional{
						Pos: lsl.At(2, 3),
						Condition: &lsl.BinaryExpression{
							Operator: ">",
							Left: &lsl.IntLiteral{
								Pos:   lsl.At(2, 6),
								Value: 10,
							},
							Right: &lsl.IntLiteral{
								Pos:   lsl.At(2, 11),
								Value: 5,
							},
						},
						Then: lsl.StatementList{
							&lsl.FunctionCall{
								Owner: &lsl.Identifier{
									Pos:  lsl.At(3, 5),
									Name: "doFirst",
								},
							},
						},
					},
					&lsl.Conditional{
						Pos: lsl.At(5, 3),
						Condition: &lsl.BinaryExpression{
							Operator: ">",
							Left: &lsl.IntLiteral{
								Pos:   lsl.At(5, 6),
								Value: 10,
							},
							Right: &lsl.IntLiteral{
								Pos:   lsl.At(5, 11),
								Value: 20,
							},
						},
						Then: lsl.StatementList{
							&lsl.FunctionCall{
								Owner: &lsl.Identifier{
									Pos:  lsl.At(6, 5),
									Name: "doFirst",
								},
							},
						},
						Else: lsl.StatementList{
							&lsl.FunctionCall{
								Owner: &lsl.Identifier{
									Pos:  lsl.At(8, 5),
									Name: "doSecond",
								},
							},
						},
					},
					&lsl.Conditional{
						Pos: lsl.At(10, 3),
						Condition: &lsl.BinaryExpression{
							Operator: ">",
							Left: &lsl.IntLiteral{
								Pos:   lsl.At(10, 6),
								Value: 10,
							},
							Right: &lsl.IntLiteral{
								Pos:   lsl.At(10, 11),
								Value: 20,
							},
						},
						Then: lsl.StatementList{
							&lsl.FunctionCall{
								Owner: &lsl.Identifier{
									Pos:  lsl.At(11, 5),
									Name: "doFirst",
								},
							},
						},
						Else: &lsl.Conditional{
							Pos: lsl.At(12, 10),
							Condition: &lsl.BinaryExpression{
								Operator: ">",
								Left: &lsl.IntLiteral{
									Pos:   lsl.At(12, 13),
									Value: 10,
								},
								Right: &lsl.IntLiteral{
									Pos:   lsl.At(12, 18),
									Value: 5,
								},
							},
							Then: lsl.StatementList{
								&lsl.FunctionCall{
									Owner: &lsl.Identifier{
										Pos:  lsl.At(13, 5),
										Name: "doSecond",
									},
								},
							},
							Else: lsl.StatementList{
								&lsl.FunctionCall{
									Owner: &lsl.Identifier{
										Pos:  lsl.At(15, 5),
										Name: "doThird",
									},
								},
							},
						},
					},
				},
			},
		),
		Entry("with discard",
			openTestFile("parser", "parse-function", "with-discard.lsl"),
			&lsl.FunctionDeclaration{
				Pos:        lsl.At(1, 1),
				Name:       "test",
				Inputs:     nil,
				OutputType: "",
				Body: lsl.StatementList{
					&lsl.Discard{
						Pos: lsl.At(2, 3),
					},
					&lsl.Discard{
						Pos: lsl.At(3, 3),
					},
				},
			},
		),
	)

	DescribeTable("ParseShader", func(inSource string, expectedShader *lsl.Shader) {
		parser := lsl.NewParser(inSource)
		shader, err := parser.ParseShader()
		Expect(err).ToNot(HaveOccurred())
		Expect(shader).To(Equal(expectedShader))
	},
		Entry("empty",
			openTestFile("parser", "parse-shader", "empty.lsl"),
			&lsl.Shader{},
		),
		Entry("root comments",
			openTestFile("parser", "parse-shader", "root-comments.lsl"),
			&lsl.Shader{},
		),
		Entry("texture blocks",
			openTestFile("parser", "parse-shader", "texture-blocks.lsl"),
			&lsl.Shader{
				Declarations: []lsl.Declaration{
					&lsl.TextureBlockDeclaration{
						Pos: lsl.At(1, 1),
						Fields: []lsl.Field{
							{
								Pos:  lsl.At(2, 3),
								Name: "color",
								Type: lsl.TypeNameSampler2D,
							},
							{
								Pos:  lsl.At(4, 3),
								Name: "intensity",
								Type: lsl.TypeNameSamplerCube,
							},
						},
					},
					&lsl.TextureBlockDeclaration{
						Pos: lsl.At(9, 1),
						Fields: []lsl.Field{
							{
								Pos:  lsl.At(10, 3),
								Name: "value",
								Type: lsl.TypeNameSampler2D,
							},
						},
					},
				},
			},
		),
		Entry("uniform blocks",
			openTestFile("parser", "parse-shader", "uniform-blocks.lsl"),
			&lsl.Shader{
				Declarations: []lsl.Declaration{
					&lsl.UniformBlockDeclaration{
						Pos: lsl.At(1, 1),
						Fields: []lsl.Field{
							{
								Pos:  lsl.At(2, 3),
								Name: "color",
								Type: lsl.TypeNameVec3,
							},
							{
								Pos:  lsl.At(4, 3),
								Name: "intensity",
								Type: lsl.TypeNameFloat,
							},
						},
					},
					&lsl.UniformBlockDeclaration{
						Pos: lsl.At(9, 1),
						Fields: []lsl.Field{
							{
								Pos:  lsl.At(10, 3),
								Name: "value",
								Type: lsl.TypeNameVec4,
							},
						},
					},
				},
			},
		),
		Entry("varying blocks",
			openTestFile("parser", "parse-shader", "varying-blocks.lsl"),
			&lsl.Shader{
				Declarations: []lsl.Declaration{
					&lsl.VaryingBlockDeclaration{
						Pos: lsl.At(1, 1),
						Fields: []lsl.Field{
							{
								Pos:  lsl.At(2, 3),
								Name: "color",
								Type: lsl.TypeNameVec3,
							},
							{
								Pos:  lsl.At(4, 3),
								Name: "intensity",
								Type: lsl.TypeNameFloat,
							},
						},
					},
					&lsl.VaryingBlockDeclaration{
						Pos: lsl.At(7, 1),
						Fields: []lsl.Field{
							{
								Pos:  lsl.At(8, 3),
								Name: "value",
								Type: lsl.TypeNameVec4,
							},
						},
					},
				},
			},
		),
		Entry("function declarations",
			openTestFile("parser", "parse-shader", "function-declarations.lsl"),
			&lsl.Shader{
				Declarations: []lsl.Declaration{
					&lsl.FunctionDeclaration{
						Pos:  lsl.At(1, 1),
						Name: "vertex",
						Inputs: []lsl.Field{
							{
								Pos:  lsl.At(1, 13),
								Name: "a",
								Type: lsl.TypeNameVec3,
							},
							{
								Pos:  lsl.At(1, 21),
								Name: "b",
								Type: lsl.TypeNameVec4,
							},
						},
						OutputType: lsl.TypeNameFloat,
					},
					&lsl.FunctionDeclaration{
						Pos:        lsl.At(4, 1),
						Name:       "#fragment",
						Inputs:     nil,
						OutputType: "",
					},
				},
			},
		),
		Entry("variable declarations",
			openTestFile("parser", "parse-shader", "variable-declarations.lsl"),
			&lsl.Shader{
				Declarations: []lsl.Declaration{
					&lsl.FunctionDeclaration{
						Pos:        lsl.At(1, 1),
						Name:       "test",
						Inputs:     nil,
						OutputType: "",
						Body: lsl.StatementList{
							&lsl.VariableDeclaration{
								Pos:  lsl.At(2, 3),
								Name: "color",
								Type: lsl.TypeNameVec3,
								Assignment: &lsl.FunctionCall{
									Owner: &lsl.Identifier{
										Pos:  lsl.At(2, 20),
										Name: "vec3",
									},
									Arguments: []lsl.Expression{
										&lsl.FloatLiteral{
											Pos:   lsl.At(2, 25),
											Value: 1.0,
										},
										&lsl.UnaryExpression{
											Pos:      lsl.At(2, 29),
											Operator: "-",
											Operand: &lsl.FloatLiteral{
												Pos:   lsl.At(2, 30),
												Value: 0.5,
											},
										},
										&lsl.FloatLiteral{
											Pos:   lsl.At(2, 35),
											Value: 0.1,
										},
									},
								},
							},
							&lsl.VariableDeclaration{
								Pos:  lsl.At(4, 3),
								Name: "intensity",
								Type: lsl.TypeNameFloat,
							},
						},
					},
				},
			},
		),
	)
})
