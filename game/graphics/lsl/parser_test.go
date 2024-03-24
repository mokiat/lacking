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
		outShader, outErr = lsl.Parse2(inSource) // TODO: Change to Parse
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

	When("uniform blocks are present", func() {
		BeforeEach(func() {
			inSource = `
				#uniform { // header
					color vec3, // field1
					// has two fields
					intensity float, // field2
				}

				// comment here

				#uniform {
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
				#varying { // header
					color vec3, // field 1
					// two fields
					intensity float, // field2
				} // footer

				#varying {
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

		PIt("produces a shader with the variable declarations", func() {
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
										&lsl.FloatLiteral{
											Value: -0.5,
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

	DescribeTable("ParseNewLine", func(inSource string) {
		parser := lsl.NewParser(inSource)
		Expect(parser.ParseNewLine()).To(Succeed())
	},
		Entry("new line",
			`
			`,
		),
		Entry("new line after spacing",
			`  
			`,
		),
	)

	DescribeTable("ParseComment", func(inSource string) {
		parser := lsl.NewParser(inSource)
		Expect(parser.ParseComment()).To(Succeed())
	},
		Entry("plain comment",
			`// a comment`,
		),
		Entry("comment after spacing",
			`// a comment`,
		),
		Entry("comment with new line",
			`// some comment
			`,
		),
	)

	DescribeTable("ParseOptionalRemainder", func(inSource string) {
		parser := lsl.NewParser(inSource)
		Expect(parser.ParseOptionalRemainder()).To(Succeed())
	},
		Entry("empty",
			``,
		),
		Entry("new line",
			`
			`,
		),
		Entry("comment",
			`// a comment`,
		),
	)

	DescribeTable("ParseBlockStart", func(inSource string) {
		parser := lsl.NewParser(inSource)
		Expect(parser.ParseBlockStart()).To(Succeed())
	},
		Entry("with new line",
			`  {
			`,
		),
		Entry("with comment",
			`{ // closing bracket
			`,
		),
	)

	DescribeTable("ParseBlockEnd", func(inSource string) {
		parser := lsl.NewParser(inSource)
		Expect(parser.ParseBlockEnd()).To(Succeed())
	},
		Entry("just closing bracket",
			`}`,
		),
		Entry("with new line",
			`}
			`,
		),
		Entry("with comment",
			`} // closing bracket
			`,
		),
	)

	DescribeTable("ParseNamedParameterList", func(inSource string, expectedFields []lsl.Field) {
		parser := lsl.NewParser(inSource)
		fields, err := parser.ParseNamedParameterList()
		Expect(err).ToNot(HaveOccurred())
		Expect(fields).To(Equal(expectedFields))
	},
		Entry("empty list",
			``,
			nil,
		),
		Entry("single parameter",
			`color vec4`,
			[]lsl.Field{
				{Name: "color", Type: "vec4"},
			},
		),
		Entry("multiple parameters, single line",
			`color vec4, intensity float`,
			[]lsl.Field{
				{Name: "color", Type: "vec4"},
				{Name: "intensity", Type: "float"},
			},
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
		),
	)

	DescribeTable("ParseUnnamedParameterList", func(inSource string, expectedFields []lsl.Field) {
		parser := lsl.NewParser(inSource)
		fields, err := parser.ParseUnnamedParameterList()
		Expect(err).ToNot(HaveOccurred())
		Expect(fields).To(Equal(expectedFields))
	},
		Entry("empty list",
			``,
			nil,
		),
		Entry("single parameter",
			`vec4`,
			[]lsl.Field{
				{Type: "vec4"},
			},
		),
		Entry("multiple parameters, single line",
			`vec4, float`,
			[]lsl.Field{
				{Type: "vec4"},
				{Type: "float"},
			},
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
		),
	)

	DescribeTable("ParseUniformBlock", func(inSource string, expectedBlock *lsl.UniformBlockDeclaration) {
		parser := lsl.NewParser(inSource)
		block, err := parser.ParseUniformBlock()
		Expect(err).ToNot(HaveOccurred())
		Expect(block).To(Equal(expectedBlock))
	},
		Entry("empty",
			`#uniform {
			}`,
			&lsl.UniformBlockDeclaration{},
		),
		Entry("with fields",
			`#uniform {
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
			`#varying {
			}`,
			&lsl.VaryingBlockDeclaration{},
		),
		Entry("with fields",
			`#varying {
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
		// TODO: function call
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

		// TODO: Test body with statements
	)
})
