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
				{Name: "color", Type: "vec4"},
			},
			nil,
		),
		Entry("multiple parameters, single line",
			openTestFile("parser", "parse-named-parameter-list", "multiple-parameters-single-line.lsl"),
			[]lsl.Field{
				{Name: "color", Type: "vec4"},
				{Name: "intensity", Type: "float"},
			},
			nil,
		),
		Entry("multiple parameters, multiple lines",
			openTestFile("parser", "parse-named-parameter-list", "multiple-parameters-multiple-lines.lsl"),
			[]lsl.Field{
				{Name: "color", Type: "vec4"},
				{Name: "intensity", Type: "float"},
			},
			nil,
		),
		Entry("ending on a non-comma operator",
			openTestFile("parser", "parse-named-parameter-list", "ending-on-non-comma-operator.lsl"),
			[]lsl.Field{
				{Name: "color", Type: "vec4"},
			},
			nil,
		),
		Entry("ending on a comma operator",
			openTestFile("parser", "parse-named-parameter-list", "ending-on-comma-operator.lsl"),
			[]lsl.Field{
				{Name: "color", Type: "vec4"},
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

	DescribeTable("ParseUnnamedParameterList", func(inSource string, expectedFields []lsl.Field, expectedErr error) {
		parser := lsl.NewParser(inSource)
		fields, err := parser.ParseUnnamedParameterList()
		if expectedErr == nil {
			Expect(err).ToNot(HaveOccurred())
			Expect(fields).To(Equal(expectedFields))
		} else {
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(expectedErr))
		}
	},
		Entry("empty list",
			openTestFile("parser", "parse-unnamed-parameter-list", "empty.lsl"),
			nil,
			nil,
		),
		Entry("single parameter",
			openTestFile("parser", "parse-unnamed-parameter-list", "single-parameter.lsl"),
			[]lsl.Field{
				{Type: "vec4"},
			},
			nil,
		),
		Entry("multiple parameters, single line",
			openTestFile("parser", "parse-unnamed-parameter-list", "multiple-parameters-single-line.lsl"),
			[]lsl.Field{
				{Type: "vec4"},
				{Type: "float"},
			},
			nil,
		),
		Entry("multiple parameters, multiple lines",
			openTestFile("parser", "parse-unnamed-parameter-list", "multiple-parameters-multiple-lines.lsl"),
			[]lsl.Field{
				{Type: "vec4"},
				{Type: "float"},
			},
			nil,
		),
		Entry("ending on a non-comma operator",
			openTestFile("parser", "parse-unnamed-parameter-list", "ending-on-non-comma-operator.lsl"),
			[]lsl.Field{
				{Type: "vec4"},
			},
			nil,
		),
		Entry("ending on a comma operator",
			openTestFile("parser", "parse-unnamed-parameter-list", "ending-on-comma-operator.lsl"),
			[]lsl.Field{
				{Type: "vec4"},
			},
			nil,
		),
		Entry("ending on a twin-comma operator",
			openTestFile("parser", "parse-unnamed-parameter-list", "ending-on-twin-comma-operator.lsl"),
			nil,
			&lsl.ParseError{
				Pos:     lsl.At(1, 6),
				Message: "unexpected comma",
			},
		),
		Entry("non-identifier type",
			openTestFile("parser", "parse-unnamed-parameter-list", "non-identifier-type.lsl"),
			nil,
			&lsl.ParseError{
				Pos:     lsl.At(1, 1),
				Message: "expected a type identifier or end of list",
			},
		),
		Entry("non-comma or operator after type",
			openTestFile("parser", "parse-unnamed-parameter-list", "non-comma-operator-after-type.lsl"),
			nil,
			&lsl.ParseError{
				Pos:     lsl.At(1, 6),
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
			&lsl.TextureBlockDeclaration{},
			nil,
		),
		Entry("with fields",
			openTestFile("parser", "parse-texture-block", "with-fields.lsl"),
			&lsl.TextureBlockDeclaration{
				Fields: []lsl.Field{
					{Name: "first", Type: "sampler2D"},
					{Name: "second", Type: "samplerCube"},
				},
			},
			nil,
		),
		Entry("with comments and spaces",
			openTestFile("parser", "parse-texture-block", "with-comments-and-spaces.lsl"),
			&lsl.TextureBlockDeclaration{
				Fields: []lsl.Field{
					{Name: "first", Type: "sampler2D"},
					{Name: "second", Type: "samplerCube"},
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
			&lsl.UniformBlockDeclaration{},
			nil,
		),
		Entry("with fields",
			openTestFile("parser", "parse-uniform-block", "with-fields.lsl"),
			&lsl.UniformBlockDeclaration{
				Fields: []lsl.Field{
					{Name: "color", Type: "vec4"},
					{Name: "intensity", Type: "float"},
				},
			},
			nil,
		),
		Entry("with comments and spaces",
			openTestFile("parser", "parse-uniform-block", "with-comments-and-spaces.lsl"),
			&lsl.UniformBlockDeclaration{
				Fields: []lsl.Field{
					{Name: "color", Type: "vec4"},
					{Name: "intensity", Type: "float"},
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
			&lsl.VaryingBlockDeclaration{},
			nil,
		),
		Entry("with fields",
			openTestFile("parser", "parse-varying-block", "with-fields.lsl"),
			&lsl.VaryingBlockDeclaration{
				Fields: []lsl.Field{
					{Name: "color", Type: "vec4"},
					{Name: "intensity", Type: "float"},
				},
			},
			nil,
		),
		Entry("with comments and spaces",
			openTestFile("parser", "parse-varying-block", "with-comments-and-spaces.lsl"),
			&lsl.VaryingBlockDeclaration{
				Fields: []lsl.Field{
					{Name: "color", Type: "vec4"},
					{Name: "intensity", Type: "float"},
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

	DescribeTable("ParseExpression", func(inSource string, expectedExp lsl.Expression) {
		parser := lsl.NewParser(inSource)
		exp, err := parser.ParseExpression()
		Expect(err).ToNot(HaveOccurred())
		Expect(exp).To(Equal(expectedExp))
	},
		Entry("float literal",
			openTestFile("parser", "parse-expression", "float-literal.lsl"),
			&lsl.FloatLiteral{
				Value: 5.3,
			},
		),
		Entry("int literal",
			openTestFile("parser", "parse-expression", "int-literal.lsl"),
			&lsl.IntLiteral{
				Value: 3999,
			},
		),
		Entry("unary (+) operator",
			openTestFile("parser", "parse-expression", "unary-plus-operator.lsl"),
			&lsl.UnaryExpression{
				Operator: "+",
				Operand: &lsl.IntLiteral{
					Value: 10,
				},
			},
		),
		Entry("unary (-) operator",
			openTestFile("parser", "parse-expression", "unary-minus-operator.lsl"),
			&lsl.UnaryExpression{
				Operator: "-",
				Operand: &lsl.IntLiteral{
					Value: 10,
				},
			},
		),
		Entry("unary (^) operator",
			openTestFile("parser", "parse-expression", "unary-bit-not-operator.lsl"),
			&lsl.UnaryExpression{
				Operator: "^",
				Operand: &lsl.IntLiteral{
					Value: 10,
				},
			},
		),
		Entry("unary (!) operator",
			openTestFile("parser", "parse-expression", "unary-not-operator.lsl"),
			&lsl.UnaryExpression{
				Operator: "!",
				Operand: &lsl.IntLiteral{
					Value: 10,
				},
			},
		),
		Entry("identifier",
			openTestFile("parser", "parse-expression", "identifier.lsl"),
			&lsl.Identifier{
				Name: "hello",
			},
		),
		Entry("field identifier",
			openTestFile("parser", "parse-expression", "field-identifier.lsl"),
			&lsl.FieldIdentifier{
				ObjName:   "hello",
				FieldName: "world",
			},
		),
		Entry("function call",
			openTestFile("parser", "parse-expression", "function-call.lsl"),
			&lsl.FunctionCall{
				Name:      "rand",
				Arguments: nil,
			},
		),
		Entry("function call with args",
			openTestFile("parser", "parse-expression", "function-call-with-args.lsl"),
			&lsl.FunctionCall{
				Name: "test",
				Arguments: []lsl.Expression{
					&lsl.IntLiteral{Value: 200},
					&lsl.FloatLiteral{Value: 1.5},
				},
			},
		),
		Entry("function call with args and new lines",
			openTestFile("parser", "parse-expression", "function-call-with-args-and-new-lines.lsl"),
			&lsl.FunctionCall{
				Name: "test",
				Arguments: []lsl.Expression{
					&lsl.IntLiteral{Value: 200},
					&lsl.FloatLiteral{Value: 1.5},
				},
			},
		),
		Entry("function call with args and comments",
			openTestFile("parser", "parse-expression", "function-call-with-args-and-comments.lsl"),
			&lsl.FunctionCall{
				Name: "test",
				Arguments: []lsl.Expression{
					&lsl.IntLiteral{Value: 200},
					&lsl.FloatLiteral{Value: 1.5},
				},
			},
		),
		Entry("binary expression (numbers)",
			openTestFile("parser", "parse-expression", "binary-expression-numbers.lsl"),
			&lsl.BinaryExpression{
				Operator: "+",
				Left:     &lsl.IntLiteral{Value: 1},
				Right:    &lsl.FloatLiteral{Value: 2.3},
			},
		),
		Entry("binary expression (identifiers)",
			openTestFile("parser", "parse-expression", "binary-expression-identifiers.lsl"),
			&lsl.BinaryExpression{
				Operator: "+",
				Left:     &lsl.Identifier{Name: "amount"},
				Right:    &lsl.FieldIdentifier{ObjName: "color", FieldName: "x"},
			},
		),
		Entry("binary expression (functions)",
			openTestFile("parser", "parse-expression", "binary-expression-functions.lsl"),
			&lsl.BinaryExpression{
				Operator: "*",
				Left:     &lsl.FunctionCall{Name: "first"},
				Right:    &lsl.FunctionCall{Name: "second"},
			},
		),
		Entry("complex expression",
			openTestFile("parser", "parse-expression", "complex-expression.lsl"),
			&lsl.BinaryExpression{
				Operator: "+",
				Left:     &lsl.FloatLiteral{Value: 5.5},
				Right: &lsl.BinaryExpression{
					Operator: "*",
					Left:     &lsl.Identifier{Name: "hello"},
					Right: &lsl.BinaryExpression{
						Operator: "-",
						Left: &lsl.BinaryExpression{
							Operator: "/",
							Left:     &lsl.IntLiteral{Value: 13},
							Right:    &lsl.IntLiteral{Value: 2},
						},
						Right: &lsl.IntLiteral{Value: 77},
					},
				},
			},
		),

		// TODO: Test logical expressions
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
				Name:    "test",
				Inputs:  nil,
				Outputs: nil,
				Body:    nil,
			},
		),
		Entry("with inputs",
			openTestFile("parser", "parse-function", "with-inputs.lsl"),
			&lsl.FunctionDeclaration{
				Name: "test",
				Inputs: []lsl.Field{
					{Name: "color", Type: "vec3"},
					{Name: "intensity", Type: "float"},
				},
				Outputs: nil,
				Body:    nil,
			},
		),
		Entry("with inputs on new lines",
			openTestFile("parser", "parse-function", "with-inputs-on-new-lines.lsl"),
			&lsl.FunctionDeclaration{
				Name: "test",
				Inputs: []lsl.Field{
					{Name: "color", Type: "vec3"},
					{Name: "intensity", Type: "float"},
				},
				Outputs: nil,
				Body:    nil,
			},
		),
		Entry("with single output",
			openTestFile("parser", "parse-function", "with-simple-output.lsl"),
			&lsl.FunctionDeclaration{
				Name:   "test",
				Inputs: nil,
				Outputs: []lsl.Field{
					{Type: "vec3"},
				},
				Body: nil,
			},
		),
		Entry("with multiple outputs",
			openTestFile("parser", "parse-function", "with-multiple-outputs.lsl"),
			&lsl.FunctionDeclaration{
				Name:   "test",
				Inputs: nil,
				Outputs: []lsl.Field{
					{Type: "vec3"},
					{Type: "float"},
				},
				Body: nil,
			},
		),
		Entry("with inputs and outputs",
			openTestFile("parser", "parse-function", "with-inputs-and-outputs.lsl"),
			&lsl.FunctionDeclaration{
				Name: "test",
				Inputs: []lsl.Field{
					{Name: "color", Type: "vec3"},
					{Name: "intensity", Type: "float"},
				},
				Outputs: []lsl.Field{
					{Type: "vec3"},
					{Type: "float"},
				},
				Body: nil,
			},
		),
		Entry("with comment in body",
			openTestFile("parser", "parse-function", "with-comment-in-body.lsl"),
			&lsl.FunctionDeclaration{
				Name:    "test",
				Inputs:  nil,
				Outputs: nil,
				Body:    nil,
			},
		),
		Entry("with function call statements",
			openTestFile("parser", "parse-function", "with-function-call-statements.lsl"),
			&lsl.FunctionDeclaration{
				Name:    "test",
				Inputs:  nil,
				Outputs: nil,
				Body: []lsl.Statement{
					&lsl.FunctionCall{Name: "doFirst"},
					&lsl.FunctionCall{Name: "doSecond"},
				},
			},
		),
		Entry("with variable declarations",
			openTestFile("parser", "parse-function", "with-variable-declarations.lsl"),
			&lsl.FunctionDeclaration{
				Name:    "test",
				Inputs:  nil,
				Outputs: nil,
				Body: []lsl.Statement{
					&lsl.VariableDeclaration{
						Name:       "x",
						Type:       "float",
						Assignment: &lsl.FloatLiteral{Value: 5.3},
					},
					&lsl.VariableDeclaration{
						Name:       "y",
						Type:       "int",
						Assignment: &lsl.IntLiteral{Value: 3},
					},
					&lsl.VariableDeclaration{
						Name: "z",
						Type: "vec3",
						Assignment: &lsl.FunctionCall{
							Name: "vec3",
							Arguments: []lsl.Expression{
								&lsl.FloatLiteral{Value: 1.0},
								&lsl.FloatLiteral{Value: 0.0},
								&lsl.UnaryExpression{
									Operator: "-",
									Operand:  &lsl.FloatLiteral{Value: 0.5},
								},
							},
						},
					},
				},
			},
		),
		Entry("with variable assignments",
			openTestFile("parser", "parse-function", "with-variable-assignments.lsl"),
			&lsl.FunctionDeclaration{
				Name:    "test",
				Inputs:  nil,
				Outputs: nil,
				Body: []lsl.Statement{
					&lsl.Assignment{
						Target:     &lsl.FieldIdentifier{ObjName: "color", FieldName: "x"},
						Operator:   "+=",
						Expression: &lsl.FloatLiteral{Value: 5.3},
					},
					&lsl.Assignment{
						Target:     &lsl.FieldIdentifier{ObjName: "color", FieldName: "y"},
						Operator:   "*=",
						Expression: &lsl.IntLiteral{Value: 3},
					},
					&lsl.Assignment{
						Target:   &lsl.Identifier{Name: "z"},
						Operator: "=",
						Expression: &lsl.FunctionCall{
							Name: "vec3",
							Arguments: []lsl.Expression{
								&lsl.FloatLiteral{Value: 1.0},
								&lsl.FloatLiteral{Value: 0.0},
								&lsl.UnaryExpression{
									Operator: "-",
									Operand:  &lsl.FloatLiteral{Value: 0.5},
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
				Name:    "test",
				Inputs:  nil,
				Outputs: nil,
				Body: []lsl.Statement{
					&lsl.Conditional{
						Condition: &lsl.BinaryExpression{
							Operator: ">",
							Left:     &lsl.IntLiteral{Value: 10},
							Right:    &lsl.IntLiteral{Value: 5},
						},
						Then: []lsl.Statement{
							&lsl.FunctionCall{Name: "doFirst"},
						},
					},
					&lsl.Conditional{
						Condition: &lsl.BinaryExpression{
							Operator: ">",
							Left:     &lsl.IntLiteral{Value: 10},
							Right:    &lsl.IntLiteral{Value: 20},
						},
						Then: []lsl.Statement{
							&lsl.FunctionCall{Name: "doFirst"},
						},
						Else: []lsl.Statement{
							&lsl.FunctionCall{Name: "doSecond"},
						},
					},
					&lsl.Conditional{
						Condition: &lsl.BinaryExpression{
							Operator: ">",
							Left:     &lsl.IntLiteral{Value: 10},
							Right:    &lsl.IntLiteral{Value: 20},
						},
						Then: []lsl.Statement{
							&lsl.FunctionCall{Name: "doFirst"},
						},
						ElseIf: &lsl.Conditional{
							Condition: &lsl.BinaryExpression{
								Operator: ">",
								Left:     &lsl.IntLiteral{Value: 10},
								Right:    &lsl.IntLiteral{Value: 5},
							},
							Then: []lsl.Statement{
								&lsl.FunctionCall{Name: "doSecond"},
							},
							Else: []lsl.Statement{
								&lsl.FunctionCall{Name: "doThird"},
							},
						},
					},
				},
			},
		),
		Entry("with discard",
			openTestFile("parser", "parse-function", "with-discard.lsl"),
			&lsl.FunctionDeclaration{
				Name:    "test",
				Inputs:  nil,
				Outputs: nil,
				Body: []lsl.Statement{
					&lsl.Discard{},
					&lsl.Discard{},
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
						Fields: []lsl.Field{
							{
								Name: "color",
								Type: lsl.TypeNameSampler2D,
							},
							{
								Name: "intensity",
								Type: lsl.TypeNameSamplerCube,
							},
						},
					},
					&lsl.TextureBlockDeclaration{
						Fields: []lsl.Field{
							{
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
						Fields: []lsl.Field{
							{
								Name: "color",
								Type: lsl.TypeNameVec3,
							},
							{
								Name: "intensity",
								Type: lsl.TypeNameFloat,
							},
						},
					},
					&lsl.UniformBlockDeclaration{
						Fields: []lsl.Field{
							{
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
						Fields: []lsl.Field{
							{
								Name: "color",
								Type: lsl.TypeNameVec3,
							},
							{
								Name: "intensity",
								Type: lsl.TypeNameFloat,
							},
						},
					},
					&lsl.VaryingBlockDeclaration{
						Fields: []lsl.Field{
							{
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
						Name: "vertex",
						Inputs: []lsl.Field{
							{
								Name: "a",
								Type: lsl.TypeNameVec3,
							},
							{
								Name: "b",
								Type: lsl.TypeNameVec4,
							},
						},
						Outputs: []lsl.Field{
							{
								Name: "",
								Type: lsl.TypeNameFloat,
							},
							{
								Name: "",
								Type: lsl.TypeNameVec2,
							},
						},
					},
					&lsl.FunctionDeclaration{
						Name:    "#fragment",
						Inputs:  nil,
						Outputs: nil,
					},
				},
			},
		),
		Entry("variable declarations",
			openTestFile("parser", "parse-shader", "variable-declarations.lsl"),
			&lsl.Shader{
				Declarations: []lsl.Declaration{
					&lsl.FunctionDeclaration{
						Name:    "test",
						Inputs:  nil,
						Outputs: nil,
						Body: []lsl.Statement{
							&lsl.VariableDeclaration{
								Name: "color",
								Type: lsl.TypeNameVec3,
								Assignment: &lsl.FunctionCall{
									Name: "vec3",
									Arguments: []lsl.Expression{
										&lsl.FloatLiteral{
											Value: 1.0,
										},
										&lsl.UnaryExpression{
											Operator: "-",
											Operand: &lsl.FloatLiteral{
												Value: 0.5,
											},
										},
										&lsl.FloatLiteral{
											Value: 0.1,
										},
									},
								},
							},
							&lsl.VariableDeclaration{
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
