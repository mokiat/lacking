package lsl_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/lacking/game/graphics/lsl"
)

var _ = Describe("Parser", func() {

	DescribeTable("ParseFieldGroup", func(inSource string, expectedFields []lsl.Field, expectedErr error) {
		parser := lsl.NewParser(inSource)
		fields, err := parser.ParseFieldGroup(lsl.GroupStart, lsl.GroupEnd)
		if expectedErr == nil {
			Expect(err).ToNot(HaveOccurred())
			Expect(fields).To(Equal(expectedFields))
		} else {
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(expectedErr))
		}
	},
		Entry("empty",
			openTestFile("parser", "parse-field-group", "valid-empty.lsl"),
			nil,
			nil,
		),
		Entry("single field",
			openTestFile("parser", "parse-field-group", "valid-single-field.lsl"),
			[]lsl.Field{
				{
					Pos:  lsl.At(2, 3),
					Name: "color",
					Type: "vec4",
				},
			},
			nil,
		),
		Entry("multiple fields",
			openTestFile("parser", "parse-field-group", "valid-multiple-fields.lsl"),
			[]lsl.Field{
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
			nil,
		),
		Entry("bloated",
			openTestFile("parser", "parse-field-group", "valid-bloated.lsl"),
			[]lsl.Field{
				{
					Pos:  lsl.At(3, 3),
					Name: "color",
					Type: "vec4",
				},
				{
					Pos:  lsl.At(7, 5),
					Name: "intensity",
					Type: "float",
				},
			},
			nil,
		),
		Entry("ending on a comma",
			openTestFile("parser", "parse-field-group", "invalid-ending-on-comma.lsl"),
			nil,
			&lsl.ParseError{
				Pos:     lsl.At(2, 13),
				Message: "expected a comment, new line or end of file",
			},
		),
		Entry("non-identifier name",
			openTestFile("parser", "parse-field-group", "invalid-non-identifier-name.lsl"),
			nil,
			&lsl.ParseError{
				Pos:     lsl.At(2, 3),
				Message: "expected a name identifier or end of list",
			},
		),
		Entry("non-identifier type",
			openTestFile("parser", "parse-field-group", "invalid-non-identifier-type.lsl"),
			nil,
			&lsl.ParseError{
				Pos:     lsl.At(2, 9),
				Message: "expected a type identifier",
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
		Entry("empty",
			openTestFile("parser", "parse-texture-block", "valid-empty.lsl"),
			&lsl.TextureBlockDeclaration{
				Pos:    lsl.At(1, 1),
				Fields: nil,
			},
			nil,
		),
		Entry("single",
			openTestFile("parser", "parse-texture-block", "valid-single.lsl"),
			&lsl.TextureBlockDeclaration{
				Pos: lsl.At(1, 1),
				Fields: []lsl.Field{
					{
						Pos:  lsl.At(1, 9),
						Name: "color",
						Type: "sampler2D",
					},
				},
			},
			nil,
		),
		Entry("compact",
			openTestFile("parser", "parse-texture-block", "valid-compact.lsl"),
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
		Entry("bloated",
			openTestFile("parser", "parse-texture-block", "valid-bloated.lsl"),
			&lsl.TextureBlockDeclaration{
				Pos: lsl.At(1, 1),
				Fields: []lsl.Field{
					{
						Pos:  lsl.At(3, 1),
						Name: "first",
						Type: "sampler2D",
					},
					{
						Pos:  lsl.At(5, 3),
						Name: "second",
						Type: "samplerCube",
					},
				},
			},
			nil,
		),
		Entry("other block type",
			openTestFile("parser", "parse-texture-block", "invalid-other-block-type.lsl"),
			nil,
			&lsl.ParseError{
				Pos:     lsl.At(1, 1),
				Message: "expected \"texture\" keyword",
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
		Entry("empty",
			openTestFile("parser", "parse-uniform-block", "valid-empty.lsl"),
			&lsl.UniformBlockDeclaration{
				Pos:    lsl.At(1, 1),
				Fields: nil,
			},
			nil,
		),
		Entry("single",
			openTestFile("parser", "parse-uniform-block", "valid-single.lsl"),
			&lsl.UniformBlockDeclaration{
				Pos: lsl.At(1, 1),
				Fields: []lsl.Field{
					{
						Pos:  lsl.At(1, 9),
						Name: "color",
						Type: "vec4",
					},
				},
			},
			nil,
		),
		Entry("compact",
			openTestFile("parser", "parse-uniform-block", "valid-compact.lsl"),
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
		Entry("bloated",
			openTestFile("parser", "parse-uniform-block", "valid-bloated.lsl"),
			&lsl.UniformBlockDeclaration{
				Pos: lsl.At(1, 1),
				Fields: []lsl.Field{
					{
						Pos:  lsl.At(3, 1),
						Name: "color",
						Type: "vec4",
					},
					{
						Pos:  lsl.At(5, 3),
						Name: "intensity",
						Type: "float",
					},
				},
			},
			nil,
		),
		Entry("other block type",
			openTestFile("parser", "parse-uniform-block", "invalid-other-block-type.lsl"),
			nil,
			&lsl.ParseError{
				Pos:     lsl.At(1, 1),
				Message: "expected \"uniform\" keyword",
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
			openTestFile("parser", "parse-varying-block", "valid-empty.lsl"),
			&lsl.VaryingBlockDeclaration{
				Pos:    lsl.At(1, 1),
				Fields: nil,
			},
			nil,
		),
		Entry("single",
			openTestFile("parser", "parse-varying-block", "valid-single.lsl"),
			&lsl.VaryingBlockDeclaration{
				Pos: lsl.At(1, 1),
				Fields: []lsl.Field{
					{
						Pos:  lsl.At(1, 9),
						Name: "color",
						Type: "vec4",
					},
				},
			},
			nil,
		),
		Entry("compact",
			openTestFile("parser", "parse-varying-block", "valid-compact.lsl"),
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
		Entry("bloated",
			openTestFile("parser", "parse-varying-block", "valid-bloated.lsl"),
			&lsl.VaryingBlockDeclaration{
				Pos: lsl.At(1, 1),
				Fields: []lsl.Field{
					{
						Pos:  lsl.At(3, 1),
						Name: "color",
						Type: "vec4",
					},
					{
						Pos:  lsl.At(5, 3),
						Name: "intensity",
						Type: "float",
					},
				},
			},
			nil,
		),
		Entry("other block type",
			openTestFile("parser", "parse-varying-block", "invalid-other-block-type.lsl"),
			nil,
			&lsl.ParseError{
				Pos:     lsl.At(1, 1),
				Message: "expected \"varying\" keyword",
			},
		),
	)

	DescribeTable("ParseTypeDeclaration", func(inSource string, expectedBlock lsl.Declaration, expectedErr error) {
		parser := lsl.NewParser(inSource)
		block, err := parser.ParseTypeDeclaration()
		if expectedErr == nil {
			Expect(err).ToNot(HaveOccurred())
			Expect(block).To(Equal(expectedBlock))
		} else {
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(expectedErr))
		}
	},
		Entry("empty struct",
			openTestFile("parser", "parse-type-declaration", "valid-struct-empty.lsl"),
			&lsl.StructTypeDeclaration{
				Pos:    lsl.At(1, 1),
				Name:   "Example",
				Fields: nil,
			},
			nil,
		),
		Entry("simple struct",
			openTestFile("parser", "parse-type-declaration", "valid-struct-simple.lsl"),
			&lsl.StructTypeDeclaration{
				Pos:  lsl.At(1, 1),
				Name: "Example",
				Fields: []lsl.Field{
					{
						Pos:  lsl.At(2, 3),
						Name: "color",
						Type: "vec4",
					},
				},
			},
			nil,
		),
		Entry("bloated struct",
			openTestFile("parser", "parse-type-declaration", "valid-struct-bloated.lsl"),
			&lsl.StructTypeDeclaration{
				Pos:  lsl.At(1, 1),
				Name: "Example",
				Fields: []lsl.Field{
					{
						Pos:  lsl.At(3, 1),
						Name: "first",
						Type: "vec4",
					},
					{
						Pos:  lsl.At(5, 3),
						Name: "second",
						Type: "float",
					},
				},
			},
			nil,
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
		Entry("bool literal (true)",
			openTestFile("parser", "parse-expression", "valid-bool-literal-true.lsl"),
			&lsl.BoolLiteral{
				Pos:   lsl.At(1, 1),
				Value: true,
			},
			nil,
		),
		Entry("bool literal (false)",
			openTestFile("parser", "parse-expression", "valid-bool-literal-false.lsl"),
			&lsl.BoolLiteral{
				Pos:   lsl.At(1, 1),
				Value: false,
			},
			nil,
		),
		Entry("float literal",
			openTestFile("parser", "parse-expression", "valid-float-literal.lsl"),
			&lsl.FloatLiteral{
				Pos:   lsl.At(1, 1),
				Value: 5.3,
			},
			nil,
		),
		Entry("int literal",
			openTestFile("parser", "parse-expression", "valid-int-literal.lsl"),
			&lsl.IntLiteral{
				Pos:   lsl.At(1, 1),
				Value: 3999,
			},
			nil,
		),
		Entry("unary (+) operator",
			openTestFile("parser", "parse-expression", "valid-unary-plus-operator.lsl"),
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
			openTestFile("parser", "parse-expression", "valid-unary-minus-operator.lsl"),
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
			openTestFile("parser", "parse-expression", "valid-unary-bit-not-operator.lsl"),
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
			openTestFile("parser", "parse-expression", "valid-unary-not-operator.lsl"),
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
			openTestFile("parser", "parse-expression", "valid-identifier.lsl"),
			&lsl.Identifier{
				Pos:  lsl.At(1, 1),
				Name: "hello",
			},
			nil,
		),
		Entry("field identifier",
			openTestFile("parser", "parse-expression", "valid-field-identifier.lsl"),
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
			openTestFile("parser", "parse-expression", "valid-field-identifier-nested.lsl"),
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
			openTestFile("parser", "parse-expression", "valid-function-call.lsl"),
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
			openTestFile("parser", "parse-expression", "valid-function-call-nested.lsl"),
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
		Entry("function call (with args)",
			openTestFile("parser", "parse-expression", "valid-function-call-with-args.lsl"),
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
		Entry("function call (bloated)",
			openTestFile("parser", "parse-expression", "valid-function-call-bloated.lsl"),
			&lsl.FunctionCall{
				Owner: &lsl.Identifier{
					Pos:  lsl.At(1, 1),
					Name: "test",
				},
				Arguments: []lsl.Expression{
					&lsl.IntLiteral{
						Pos:   lsl.At(3, 3),
						Value: 200,
					},
					&lsl.FloatLiteral{
						Pos:   lsl.At(6, 5),
						Value: 1.5,
					},
				},
			},
			nil,
		),
		Entry("binary expression (numbers)",
			openTestFile("parser", "parse-expression", "valid-binary-expression-numbers.lsl"),
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
			openTestFile("parser", "parse-expression", "valid-binary-expression-identifiers.lsl"),
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
			openTestFile("parser", "parse-expression", "valid-binary-expression-functions.lsl"),
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
			openTestFile("parser", "parse-expression", "valid-complex-expression.lsl"),
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
			openTestFile("parser", "parse-expression", "valid-logical-expression.lsl"),
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
			openTestFile("parser", "parse-expression", "valid-operator-precedence.lsl"),
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
		Entry("bloated",
			openTestFile("parser", "parse-expression", "valid-bloated.lsl"),
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
			openTestFile("parser", "parse-expression", "invalid-starts-with-binary-operator.lsl"),
			nil,
			&lsl.ParseError{
				Pos:     lsl.At(1, 1),
				Message: "expected an expression value",
			},
		),
		Entry("incomplete",
			openTestFile("parser", "parse-expression", "invalid-incomplete.lsl"),
			nil,
			&lsl.ParseError{
				Pos:     lsl.At(1, 4),
				Message: "expected an expression value",
			},
		),
		Entry("invalid value",
			openTestFile("parser", "parse-expression", "invalid-operator-value.lsl"),
			nil,
			&lsl.ParseError{
				Pos:     lsl.At(1, 5),
				Message: "expected an expression value",
			},
		),
	)

	DescribeTable("ParseArgumentBlock", func(inSource string, expectedArgs []lsl.Expression, expectedErr error) {
		parser := lsl.NewParser(inSource)
		fields, err := parser.ParseArgumentBlock()
		if expectedErr == nil {
			Expect(err).ToNot(HaveOccurred())
			Expect(fields).To(Equal(expectedArgs))
		} else {
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(expectedErr))
		}
	},
		Entry("empty",
			openTestFile("parser", "parse-argument-block", "valid-empty.lsl"),
			nil,
			nil,
		),
		Entry("single argument",
			openTestFile("parser", "parse-argument-block", "valid-single-argument.lsl"),
			[]lsl.Expression{
				&lsl.IntLiteral{
					Pos:   lsl.At(1, 2),
					Value: 10,
				},
			},
			nil,
		),
		Entry("multiple arguments",
			openTestFile("parser", "parse-argument-block", "valid-multiple-arguments.lsl"),
			[]lsl.Expression{
				&lsl.IntLiteral{
					Pos:   lsl.At(1, 2),
					Value: 10,
				},
				&lsl.FloatLiteral{
					Pos:   lsl.At(1, 5),
					Value: 5.5,
				},
			},
			nil,
		),
		Entry("bloated",
			openTestFile("parser", "parse-argument-block", "valid-bloated.lsl"),
			[]lsl.Expression{
				&lsl.FunctionCall{
					Owner: &lsl.Identifier{
						Pos:  lsl.At(2, 3),
						Name: "vec3",
					},
					Arguments: []lsl.Expression{
						&lsl.FloatLiteral{
							Pos:   lsl.At(2, 8),
							Value: 0.0,
						},
					},
				},
				&lsl.BinaryExpression{
					Operator: lsl.BinaryOperatorAdd,
					Left: &lsl.FloatLiteral{
						Pos:   lsl.At(6, 3),
						Value: 5.5,
					},
					Right: &lsl.FloatLiteral{
						Pos:   lsl.At(7, 5),
						Value: 3.3,
					},
				},
			},
			nil,
		),
		Entry("missing opening bracket",
			openTestFile("parser", "parse-argument-block", "invalid-missing-opening.lsl"),
			nil,
			&lsl.ParseError{
				Pos:     lsl.At(1, 1),
				Message: "expected an opening bracket",
			},
		),
		Entry("missing closing bracket",
			openTestFile("parser", "parse-argument-block", "invalid-missing-closing.lsl"),
			nil,
			&lsl.ParseError{
				Pos:     lsl.At(1, 5),
				Message: "expected a comma or a closing bracket",
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
			openTestFile("parser", "parse-statement", "valid-discard.lsl"),
			&lsl.Discard{
				Pos: lsl.At(1, 1),
			},
			nil,
		),
		Entry("var declaration (simple)",
			openTestFile("parser", "parse-statement", "valid-var-simple.lsl"),
			&lsl.VariableDeclaration{
				Pos:        lsl.At(1, 1),
				Name:       "color",
				Type:       "vec4",
				Assignment: nil,
			},
			nil,
		),
		Entry("var declaration (expression)",
			openTestFile("parser", "parse-statement", "valid-var-expression.lsl"),
			&lsl.VariableDeclaration{
				Pos:  lsl.At(1, 1),
				Name: "intensity",
				Type: "float",
				Assignment: &lsl.FloatLiteral{
					Pos:   lsl.At(1, 23),
					Value: 1.0,
				},
			},
			nil,
		),
		Entry("var declaration (expression no type)",
			openTestFile("parser", "parse-statement", "valid-var-expression-no-type.lsl"),
			&lsl.VariableDeclaration{
				Pos:  lsl.At(1, 1),
				Name: "intensity",
				Type: "", // auto-assignment
				Assignment: &lsl.FloatLiteral{
					Pos:   lsl.At(1, 17),
					Value: 1.0,
				},
			},
			nil,
		),
		Entry("var declaration (auto)",
			openTestFile("parser", "parse-statement", "valid-var-auto.lsl"),
			&lsl.VariableDeclaration{
				Pos:  lsl.At(1, 1),
				Name: "intensity",
				Type: "", // auto-assignment
				Assignment: &lsl.FloatLiteral{
					Pos:   lsl.At(1, 14),
					Value: 1.0,
				},
			},
			nil,
		),
		Entry("var declaration (no type or expression)",
			openTestFile("parser", "parse-statement", "invalid-var-no-type-or-expression.lsl"),
			nil,
			&lsl.ParseError{
				Pos:     lsl.At(1, 11),
				Message: "expected a type identifier or an assignment operator",
			},
		),
		Entry("if condition (simple)",
			openTestFile("parser", "parse-statement", "valid-condition-simple.lsl"),
			&lsl.Conditional{
				Pos: lsl.At(1, 1),
				Condition: &lsl.BoolLiteral{
					Pos:   lsl.At(1, 5),
					Value: true,
				},
				Then: lsl.StatementList{
					&lsl.Discard{
						Pos: lsl.At(2, 3),
					},
				},
			},
			nil,
		),
		Entry("if condition (bloated)",
			openTestFile("parser", "parse-statement", "valid-condition-bloated.lsl"),
			&lsl.Conditional{
				Pos: lsl.At(1, 1),
				Condition: &lsl.BinaryExpression{
					Operator: lsl.BinaryOperatorGreater,
					Left: &lsl.Identifier{
						Pos:  lsl.At(1, 5),
						Name: "a",
					},
					Right: &lsl.IntLiteral{
						Pos:   lsl.At(1, 9),
						Value: 5,
					},
				},
				Then: lsl.StatementList{
					&lsl.FunctionCall{
						Owner: &lsl.Identifier{
							Pos:  lsl.At(4, 3),
							Name: "doFirst",
						},
					},
				},
				Else: &lsl.Conditional{
					Pos: lsl.At(6, 11),
					Condition: &lsl.BinaryExpression{
						Operator: lsl.BinaryOperatorGreater,
						Left: &lsl.Identifier{
							Pos:  lsl.At(6, 16),
							Name: "b",
						},
						Right: &lsl.IntLiteral{
							Pos:   lsl.At(6, 20),
							Value: 6,
						},
					},
					Then: lsl.StatementList{
						&lsl.FunctionCall{
							Owner: &lsl.Identifier{
								Pos:  lsl.At(7, 3),
								Name: "doSecond",
							},
						},
					},
					Else: &lsl.Conditional{
						Pos: lsl.At(8, 10),
						Condition: &lsl.BinaryExpression{
							Operator: lsl.BinaryOperatorGreater,
							Left: &lsl.Identifier{
								Pos:  lsl.At(8, 14),
								Name: "c",
							},
							Right: &lsl.IntLiteral{
								Pos:   lsl.At(8, 18),
								Value: 7,
							},
						},
						Then: lsl.StatementList{
							&lsl.FunctionCall{
								Owner: &lsl.Identifier{
									Pos:  lsl.At(9, 5),
									Name: "doThird",
								},
							},
						},
						Else: lsl.StatementList{
							&lsl.FunctionCall{
								Owner: &lsl.Identifier{
									Pos:  lsl.At(11, 3),
									Name: "doFourth",
								},
							},
						},
					},
				},
			},
			nil,
		),
		Entry("assignment",
			openTestFile("parser", "parse-statement", "valid-assignment.lsl"),
			&lsl.Assignment{
				Operator: lsl.AssignmentOperatorAdd,
				Target: &lsl.Identifier{
					Pos:  lsl.At(1, 1),
					Name: "intensity",
				},
				Expression: &lsl.FloatLiteral{
					Pos:   lsl.At(1, 14),
					Value: 1.0,
				},
			},
			nil,
		),
		Entry("assignment (nested)",
			openTestFile("parser", "parse-statement", "valid-assignment-nested.lsl"),
			&lsl.Assignment{
				Operator: lsl.AssignmentOperatorAdd,
				Target: &lsl.FieldIdentifier{
					Owner: &lsl.FieldIdentifier{
						Owner: &lsl.Identifier{
							Pos:  lsl.At(1, 1),
							Name: "vertex",
						},
						Field: lsl.Identifier{
							Pos:  lsl.At(1, 8),
							Name: "color",
						},
					},
					Field: lsl.Identifier{
						Pos:  lsl.At(1, 14),
						Name: "r",
					},
				},
				Expression: &lsl.FloatLiteral{
					Pos:   lsl.At(1, 19),
					Value: 1.5,
				},
			},
			nil,
		),
		Entry("function call",
			openTestFile("parser", "parse-statement", "valid-function-call-nested.lsl"),
			&lsl.FunctionCall{
				Owner: &lsl.FieldIdentifier{
					Owner: &lsl.Identifier{
						Pos:  lsl.At(1, 1),
						Name: "utils",
					},
					Field: lsl.Identifier{
						Pos:  lsl.At(1, 7),
						Name: "example",
					},
				},
				Arguments: []lsl.Expression{
					&lsl.FloatLiteral{
						Pos:   lsl.At(1, 15),
						Value: 1.0,
					},
					&lsl.FloatLiteral{
						Pos:   lsl.At(1, 20),
						Value: 5.0,
					},
				},
			},
			nil,
		),
		Entry("return (empty)",
			openTestFile("parser", "parse-statement", "valid-return-empty.lsl"),
			&lsl.Return{
				Pos:        lsl.At(1, 1),
				Expression: nil,
			},
			nil,
		),
		Entry("return (expression)",
			openTestFile("parser", "parse-statement", "valid-return-expression.lsl"),
			&lsl.Return{
				Pos: lsl.At(1, 1),
				Expression: &lsl.BinaryExpression{
					Operator: lsl.BinaryOperatorAdd,
					Left: &lsl.IntLiteral{
						Pos:   lsl.At(1, 8),
						Value: 5,
					},
					Right: &lsl.BinaryExpression{
						Operator: lsl.BinaryOperatorMul,
						Left: &lsl.IntLiteral{
							Pos:   lsl.At(1, 13),
							Value: 3,
						},
						Right: &lsl.IntLiteral{
							Pos:   lsl.At(1, 17),
							Value: 2,
						},
					},
				},
			},
			nil,
		),
	)

	DescribeTable("ParseFunction", func(inSource string, expectedDecl *lsl.FunctionDeclaration, expectedErr error) {
		parser := lsl.NewParser(inSource)
		decl, err := parser.ParseFunction()
		if expectedErr == nil {
			Expect(err).ToNot(HaveOccurred())
			Expect(decl).To(Equal(expectedDecl))
		} else {
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(expectedErr))
		}
	},
		Entry("empty",
			openTestFile("parser", "parse-function", "valid-empty.lsl"),
			&lsl.FunctionDeclaration{
				Pos:        lsl.At(1, 1),
				Name:       "test",
				Inputs:     nil,
				OutputType: "",
				Body:       nil,
			},
			nil,
		),
		Entry("with inputs",
			openTestFile("parser", "parse-function", "valid-with-inputs.lsl"),
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
			nil,
		),
		Entry("with inputs on new lines",
			openTestFile("parser", "parse-function", "valid-with-inputs-on-new-lines.lsl"),
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
			nil,
		),
		Entry("with single output",
			openTestFile("parser", "parse-function", "valid-with-output.lsl"),
			&lsl.FunctionDeclaration{
				Pos:        lsl.At(1, 1),
				Name:       "test",
				Inputs:     nil,
				OutputType: "vec3",
				Body:       nil,
			},
			nil,
		),
		Entry("with inputs and output",
			openTestFile("parser", "parse-function", "valid-with-inputs-and-output.lsl"),
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
			nil,
		),
		Entry("with comment in body",
			openTestFile("parser", "parse-function", "valid-with-comment-in-body.lsl"),
			&lsl.FunctionDeclaration{
				Pos:        lsl.At(1, 1),
				Name:       "test",
				Inputs:     nil,
				OutputType: "",
				Body:       nil,
			},
			nil,
		),
		Entry("with statements",
			openTestFile("parser", "parse-function", "valid-with-statements.lsl"),
			&lsl.FunctionDeclaration{
				Pos:        lsl.At(1, 1),
				Name:       "test",
				Inputs:     nil,
				OutputType: "",
				Body: lsl.StatementList{
					&lsl.VariableDeclaration{
						Pos:  lsl.At(3, 3),
						Name: "alpha",
						Type: "", // auto-assignment
						Assignment: &lsl.FieldIdentifier{
							Owner: &lsl.FunctionCall{
								Owner: &lsl.Identifier{
									Pos:  lsl.At(3, 12),
									Name: "texture",
								},
								Arguments: []lsl.Expression{
									&lsl.Identifier{
										Pos:  lsl.At(3, 20),
										Name: "uv",
									},
								},
							},
							Field: lsl.Identifier{
								Pos:  lsl.At(3, 24),
								Name: "a",
							},
						},
					},
					&lsl.Conditional{
						Pos: lsl.At(6, 3),
						Condition: &lsl.BinaryExpression{
							Operator: lsl.BinaryOperatorLess,
							Left: &lsl.Identifier{
								Pos:  lsl.At(6, 7),
								Name: "alpha",
							},
							Right: &lsl.FloatLiteral{
								Pos:   lsl.At(6, 15),
								Value: 0.5,
							},
						},
						Then: lsl.StatementList{
							&lsl.Discard{
								Pos: lsl.At(8, 7),
							},
						},
					},
					&lsl.Return{
						Pos:        lsl.At(11, 3),
						Expression: nil,
					},
				},
			},
			nil,
		),
	)

	DescribeTable("ParseShader", func(inSource string, expectedShader *lsl.Shader, expectedErr error) {
		parser := lsl.NewParser(inSource)
		shader, err := parser.ParseShader()
		if expectedErr == nil {
			Expect(err).ToNot(HaveOccurred())
			Expect(shader).To(Equal(expectedShader))
		} else {
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(expectedErr))
		}
	},
		Entry("empty",
			openTestFile("parser", "parse-shader", "valid-empty.lsl"),
			&lsl.Shader{},
			nil,
		),
		Entry("root comments",
			openTestFile("parser", "parse-shader", "valid-root-comments.lsl"),
			&lsl.Shader{},
			nil,
		),
		Entry("texture blocks",
			openTestFile("parser", "parse-shader", "valid-texture-blocks.lsl"),
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
								Pos:  lsl.At(9, 9),
								Name: "value",
								Type: lsl.TypeNameSampler2D,
							},
						},
					},
				},
			},
			nil,
		),
		Entry("uniform blocks",
			openTestFile("parser", "parse-shader", "valid-uniform-blocks.lsl"),
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
								Pos:  lsl.At(9, 9),
								Name: "value",
								Type: lsl.TypeNameVec4,
							},
						},
					},
				},
			},
			nil,
		),
		Entry("varying blocks",
			openTestFile("parser", "parse-shader", "valid-varying-blocks.lsl"),
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
								Pos:  lsl.At(7, 9),
								Name: "value",
								Type: lsl.TypeNameVec4,
							},
						},
					},
				},
			},
			nil,
		),
		Entry("function declarations",
			openTestFile("parser", "parse-shader", "valid-function-declarations.lsl"),
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
			nil,
		),
		Entry("struct declarations",
			openTestFile("parser", "parse-shader", "valid-struct-declarations.lsl"),
			&lsl.Shader{
				Declarations: []lsl.Declaration{
					&lsl.StructTypeDeclaration{
						Pos:  lsl.At(2, 1),
						Name: "Vertex",
						Fields: []lsl.Field{
							{
								Pos:  lsl.At(3, 3),
								Name: "position",
								Type: lsl.TypeNameVec3,
							},
							{
								Pos:  lsl.At(4, 3),
								Name: "color",
								Type: lsl.TypeNameVec3,
							},
						},
					},
					&lsl.StructTypeDeclaration{
						Pos:  lsl.At(7, 1),
						Name: "Lighting",
						Fields: []lsl.Field{
							{
								Pos:  lsl.At(8, 3),
								Name: "intensity",
								Type: lsl.TypeNameFloat,
							},
						},
					},
				},
			},
			nil,
		),
	)
})
