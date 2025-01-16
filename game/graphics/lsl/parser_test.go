package lsl_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/lacking/game/graphics/lsl"
)

var _ = Describe("Parse", func() {
	var (
		inSource  string
		outShader *lsl.Shader
		outErr    error
	)

	JustBeforeEach(func() {
		outShader, outErr = lsl.Parse(inSource)
	})

	When("empty shader", func() {
		BeforeEach(func() {
			inSource = ``
		})

		It("produces an empty shader", func() {
			Expect(outErr).ToNot(HaveOccurred())
			Expect(outShader).To(Equal(&lsl.Shader{}))
		})
	})

	When("comments are present", func() {
		BeforeEach(func() {
			inSource = `
				// This is a comment
			`
		})

		It("ignores the comments", func() {
			Expect(outErr).ToNot(HaveOccurred())
			Expect(outShader).To(Equal(&lsl.Shader{}))
		})
	})

	When("texture blocks are present", func() {
		BeforeEach(func() {
			inSource = `
				textures { // header
					color sampler2D, // field1
					// has two fields
					intensity samplerCube, // field2
				}

				// comment here

				textures {
					value sampler2D,
				}
			`
		})

		It("produces a shader with the texture blocks", func() {
			Expect(outErr).ToNot(HaveOccurred())
			Expect(outShader).To(Equal(&lsl.Shader{
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
			}))
		})
	})

	When("uniform blocks are present", func() {
		BeforeEach(func() {
			inSource = `
				uniforms { // header
					color vec3, // field1
					// has two fields
					intensity float, // field2
				}

				// comment here

				uniforms {
					value vec4,
				}
			`
		})

		It("produces a shader with the uniform blocks", func() {
			Expect(outErr).ToNot(HaveOccurred())
			Expect(outShader).To(Equal(&lsl.Shader{
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
			}))
		})
	})

	When("varying blocks are present", func() {
		BeforeEach(func() {
			inSource = `
				varyings { // header
					color vec3, // field 1
					// two fields
					intensity float, // field2
				} // footer

				varyings {
					value vec4,
				}
			`
		})

		It("produces a shader with the varying blocks", func() {
			Expect(outErr).ToNot(HaveOccurred())
			Expect(outShader).To(Equal(&lsl.Shader{
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
			}))
		})
	})

	When("function declarations are present", func() {
		BeforeEach(func() {
			inSource = `
				func vertex(a vec3, b vec4) (float, vec2) {
				}
				
				func #fragment() {
				}
			`
		})

		It("produces a shader with the function declaration", func() {
			Expect(outErr).ToNot(HaveOccurred())
			Expect(outShader).To(Equal(&lsl.Shader{
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
			}))
		})
	})

	When("variable declarations are present", func() {
		BeforeEach(func() {
			inSource = `
				func test() { // a test function
					var color vec3 = vec3(1.0,-0.5, 0.1) // this has assignment
					// some comment
					var intensity float // this is just a declaration
				} // so much for it
			`
		})

		It("produces a shader with the variable declarations", func() {
			Expect(outErr).ToNot(HaveOccurred())
			Expect(outShader).To(Equal(&lsl.Shader{
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
			}))
		})
	})
})

var _ = Describe("Parser", func() {

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
			"\n irrelevant",
			nil,
		),
		Entry("carriage return and new line",
			"\r\n irrelevant",
			nil,
		),
		Entry("new line after spacing",
			"  \t  \n irrelevant",
			nil,
		),
		Entry("no tokens",
			"",
			&lsl.ParseError{
				Pos:     at(1, 1),
				Message: "expected a new line token",
			},
		),
		Entry("identifier",
			"hello \n",
			&lsl.ParseError{
				Pos:     at(1, 1),
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
			"// a comment",
			nil,
		),
		Entry("comment after spacing",
			" \t // a comment",
			nil,
		),
		Entry("comment with new line",
			`// some comment
			irrelevant`,
			nil,
		),
		Entry("operator",
			`; // a comment`,
			&lsl.ParseError{
				Pos:     at(1, 1),
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
			"",
			nil,
		),
		Entry("new line",
			"\r\n",
			nil,
		),
		Entry("comment",
			"// a comment",
			nil,
		),
		Entry("vital token",
			"+= 5",
			&lsl.ParseError{
				Pos:     at(1, 1),
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
		Entry("just opening bracket",
			"{",
			nil,
		),
		Entry("with new line",
			" \t {\n",
			nil,
		),
		Entry("with comment",
			"{ // closing bracket",
			nil,
		),
		Entry("not opening bracket",
			"5.0\n",
			&lsl.ParseError{
				Pos:     at(1, 1),
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
		Entry("just closing bracket",
			"}",
			nil,
		),
		Entry("with new line",
			" \t }\n",
			nil,
		),
		Entry("with comment",
			"} // closing bracket\n",
			nil,
		),
		Entry("not closing bracket",
			"5.0\n",
			&lsl.ParseError{
				Pos:     at(1, 1),
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
			"",
			nil,
			nil,
		),
		Entry("single parameter",
			"color vec4",
			[]lsl.Field{
				{Name: "color", Type: "vec4"},
			},
			nil,
		),
		Entry("multiple parameters, single line",
			" \t color vec4, \t intensity float \t ",
			[]lsl.Field{
				{Name: "color", Type: "vec4"},
				{Name: "intensity", Type: "float"},
			},
			nil,
		),
		Entry("multiple parameters, multiple lines",
			`
			color vec4, // first param here
			// there will be a second param

			intensity float,

			`,
			[]lsl.Field{
				{Name: "color", Type: "vec4"},
				{Name: "intensity", Type: "float"},
			},
			nil,
		),
		Entry("ending on a non-comma operator",
			"color vec4)",
			[]lsl.Field{
				{Name: "color", Type: "vec4"},
			},
			nil,
		),
		Entry("ending on a comma operator",
			"color vec4,",
			[]lsl.Field{
				{Name: "color", Type: "vec4"},
			},
			nil,
		),
		Entry("ending on a twin-comma operator",
			"color vec4,,",
			nil,
			&lsl.ParseError{
				Pos:     at(1, 12),
				Message: "unexpected comma",
			},
		),
		Entry("non-identifier name",
			"5 vec4",
			nil,
			&lsl.ParseError{
				Pos:     at(1, 1),
				Message: "expected a name identifier or end of list",
			},
		),
		Entry("non-identifier type",
			"color 5",
			nil,
			&lsl.ParseError{
				Pos:     at(1, 7),
				Message: "expected a type identifier",
			},
		),
		Entry("non-comma or operator after type",
			"color vec4 hello",
			nil,
			&lsl.ParseError{
				Pos:     at(1, 12),
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
			"",
			nil,
			nil,
		),
		Entry("single parameter",
			"vec4",
			[]lsl.Field{
				{Type: "vec4"},
			},
			nil,
		),
		Entry("multiple parameters, single line",
			" \t vec4, \t float \t ",
			[]lsl.Field{
				{Type: "vec4"},
				{Type: "float"},
			},
			nil,
		),
		Entry("multiple parameters, multiple lines",
			`
			vec4, // first param here
			// there will be a second param

			float,

			`,
			[]lsl.Field{
				{Type: "vec4"},
				{Type: "float"},
			},
			nil,
		),
		Entry("ending on a non-comma operator",
			"vec4)",
			[]lsl.Field{
				{Type: "vec4"},
			},
			nil,
		),
		Entry("ending on a comma operator",
			"vec4,",
			[]lsl.Field{
				{Type: "vec4"},
			},
			nil,
		),
		Entry("ending on a twin-comma operator",
			"vec4,,",
			nil,
			&lsl.ParseError{
				Pos:     at(1, 6),
				Message: "unexpected comma",
			},
		),
		Entry("non-identifier type",
			"5",
			nil,
			&lsl.ParseError{
				Pos:     at(1, 1),
				Message: "expected a type identifier or end of list",
			},
		),
		Entry("non-comma or operator after type",
			"vec4 hello",
			nil,
			&lsl.ParseError{
				Pos:     at(1, 6),
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
			`textures {
			}`,
			&lsl.TextureBlockDeclaration{},
			nil,
		),
		Entry("with fields",
			`textures {
				first sampler2D,
				second samplerCube,
			}`,
			&lsl.TextureBlockDeclaration{
				Fields: []lsl.Field{
					{Name: "first", Type: "sampler2D"},
					{Name: "second", Type: "samplerCube"},
				},
			},
			nil,
		),
		Entry("with comments and spaces",
			`textures { // block start

			first 	 sampler2D  	, // first field

			second samplerCube,    // second field

			} // block end`,
			&lsl.TextureBlockDeclaration{
				Fields: []lsl.Field{
					{Name: "first", Type: "sampler2D"},
					{Name: "second", Type: "samplerCube"},
				},
			},
			nil,
		),
		Entry("closing on same line",
			`textures {}`,
			nil,
			&lsl.ParseError{
				Pos:     at(1, 11),
				Message: "expected a comment, new line or end of file",
			},
		),
		Entry("other block type",
			`uniforms {
			}`,
			nil,
			&lsl.ParseError{
				Pos:     at(1, 1),
				Message: "expected 'textures' keyword",
			},
		),
	)

	DescribeTable("ParseUniformBlock", func(inSource string, expectedBlock *lsl.UniformBlockDeclaration) {
		parser := lsl.NewParser(inSource)
		block, err := parser.ParseUniformBlock()
		Expect(err).ToNot(HaveOccurred())
		Expect(block).To(Equal(expectedBlock))
	},
		Entry("empty",
			`uniforms {
			}`,
			&lsl.UniformBlockDeclaration{},
		),
		Entry("with fields",
			`uniforms {
				color vec4,
				intensity float,
			}`,
			&lsl.UniformBlockDeclaration{
				Fields: []lsl.Field{
					{Name: "color", Type: "vec4"},
					{Name: "intensity", Type: "float"},
				},
			},
		),
	)

	DescribeTable("ParseVaryingBlock", func(inSource string, expectedBlock *lsl.VaryingBlockDeclaration) {
		parser := lsl.NewParser(inSource)
		block, err := parser.ParseVaryingBlock()
		Expect(err).ToNot(HaveOccurred())
		Expect(block).To(Equal(expectedBlock))
	},
		Entry("empty",
			`varyings {
			}`,
			&lsl.VaryingBlockDeclaration{},
		),
		Entry("with fields",
			`varyings {
				color vec4,
				intensity float,
			}`,
			&lsl.VaryingBlockDeclaration{
				Fields: []lsl.Field{
					{Name: "color", Type: "vec4"},
					{Name: "intensity", Type: "float"},
				},
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
			`5.3`,
			&lsl.FloatLiteral{
				Value: 5.3,
			},
		),
		Entry("int literal",
			`3999`,
			&lsl.IntLiteral{
				Value: 3999,
			},
		),
		Entry("unary (+) operator",
			`+10`,
			&lsl.UnaryExpression{
				Operator: "+",
				Operand: &lsl.IntLiteral{
					Value: 10,
				},
			},
		),
		Entry("unary (-) operator",
			`-10`,
			&lsl.UnaryExpression{
				Operator: "-",
				Operand: &lsl.IntLiteral{
					Value: 10,
				},
			},
		),
		Entry("unary (^) operator",
			`^10`,
			&lsl.UnaryExpression{
				Operator: "^",
				Operand: &lsl.IntLiteral{
					Value: 10,
				},
			},
		),
		Entry("unary (!) operator",
			`!10`,
			&lsl.UnaryExpression{
				Operator: "!",
				Operand: &lsl.IntLiteral{
					Value: 10,
				},
			},
		),
		Entry("identifier",
			`hello`,
			&lsl.Identifier{
				Name: "hello",
			},
		),
		Entry("field identifier",
			`hello.world`,
			&lsl.FieldIdentifier{
				ObjName:   "hello",
				FieldName: "world",
			},
		),
		Entry("function call",
			`rand()`,
			&lsl.FunctionCall{
				Name:      "rand",
				Arguments: nil,
			},
		),
		Entry("function call with args",
			`test(200, 1.5)`,
			&lsl.FunctionCall{
				Name: "test",
				Arguments: []lsl.Expression{
					&lsl.IntLiteral{Value: 200},
					&lsl.FloatLiteral{Value: 1.5},
				},
			},
		),
		Entry("function call with args and new lines",
			`test(
				200, 
				1.5,
			)`,
			&lsl.FunctionCall{
				Name: "test",
				Arguments: []lsl.Expression{
					&lsl.IntLiteral{Value: 200},
					&lsl.FloatLiteral{Value: 1.5},
				},
			},
		),
		Entry("function call with args and comments",
			`test( // function
				200, // first arg
				// some comment here
				1.5, // second arg
			) // end`,
			&lsl.FunctionCall{
				Name: "test",
				Arguments: []lsl.Expression{
					&lsl.IntLiteral{Value: 200},
					&lsl.FloatLiteral{Value: 1.5},
				},
			},
		),
		Entry("binary expression (numbers)",
			`1 + 2.3`,
			&lsl.BinaryExpression{
				Operator: "+",
				Left:     &lsl.IntLiteral{Value: 1},
				Right:    &lsl.FloatLiteral{Value: 2.3},
			},
		),
		Entry("binary expression (identifiers)",
			`amount + color.x`,
			&lsl.BinaryExpression{
				Operator: "+",
				Left:     &lsl.Identifier{Name: "amount"},
				Right:    &lsl.FieldIdentifier{ObjName: "color", FieldName: "x"},
			},
		),
		Entry("binary expression (functions)",
			`first() * second()`,
			&lsl.BinaryExpression{
				Operator: "*",
				Left:     &lsl.FunctionCall{Name: "first"},
				Right:    &lsl.FunctionCall{Name: "second"},
			},
		),
		Entry("complex expression",
			`5.5 + hello * (13 / 2 - 77)`,
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
			`func test() {
			}`,
			&lsl.FunctionDeclaration{
				Name:    "test",
				Inputs:  nil,
				Outputs: nil,
				Body:    nil,
			},
		),
		Entry("with inputs",
			`func test(color vec3, intensity float) {
			}`,
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
			`func test(
				// color param to follow
				color vec3,  // color param
				// intensity param to follow
				intensity float, // intensity param
				// all done
			) {
			}`,
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
			`func test() (vec3) {
			}`,
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
			`func test() (vec3, float) {
			}`,
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
			`func test(color vec3, intensity float) (vec3, float) {
			}`,
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
			`func test() {
				// some comment
			}`,
			&lsl.FunctionDeclaration{
				Name:    "test",
				Inputs:  nil,
				Outputs: nil,
				Body:    nil,
			},
		),
		Entry("with function call statements",
			`func test() {
				doFirst()
				doSecond()
			}`,
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
			`func test() {
				var x float = 5.3
				var y int = 3
				var z vec3 = vec3(1.0, 0.0, -0.5)
			}`,
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
			`func test() {
				color.x += 5.3
				color.y *= 3
				z = vec3(1.0, 0.0, -0.5)
			}`,
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
			`func test() {
				if 10 > 5 {
					doFirst()
				}
				if 10 > 20 {
					doFirst()
				} else {
					doSecond()
				}
				if 10 > 20 {
					doFirst()
				} else if 10 > 5 {
					doSecond()
				} else {
					doThird()
				}
			}`,
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
			`func test() {
				discard
				discard // with comment
			}`,
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
})
