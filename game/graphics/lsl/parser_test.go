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
				uniform { // header
					color vec3 // field1
					// has two fields
					intensity float // field2
				}

				// comment here

				uniform {
					value vec4
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
				varying { // header
					color vec3 // field 1
					// two fields
					intensity float // field2
				} // footer

				varying {
					value vec4
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

var _ = FDescribe("Parser", func() {

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
})
