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
		Entry("struct declarations",
			openTestFile("parser", "parse-shader", "structs.lsl"),
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
		),
	)
})
