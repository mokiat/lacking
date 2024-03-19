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

		It("ignores the line", func() {
			Expect(outErr).ToNot(HaveOccurred())
			Expect(outShader).To(Equal(&lsl.Shader{}))
		})
	})

	When("uniform blocks are present", func() {
		BeforeEach(func() {
			inSource = `
				uniform {
					color vec3
					intensity float
				}

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
				varying {
					color vec3
					intensity float
				}

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

	When("function definitions are present", func() {
		BeforeEach(func() {
			inSource = `
				func Vertex(a vec3, b vec4) (float, vec2) {
				}

				func Fragment() {
				}
			`
		})

		It("produces a shader with the function definitions", func() {
			Expect(outErr).ToNot(HaveOccurred())
			Expect(outShader).To(Equal(&lsl.Shader{
				Declarations: []lsl.Declaration{
					&lsl.FunctionDeclaration{
						Name: "Vertex",
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
						Name:    "Fragment",
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
			func test() {
				var color vec3 = vec3(1.0, 0.0, 1.0)
				var intensity float
			}
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
