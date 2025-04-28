package glsl_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/graphics/glsl"
	"github.com/mokiat/lacking/game/graphics/lsl"
)

var _ = Describe("Translator", func() {
	DescribeTable("Translate 300 es", func(shader *lsl.Shader, constraints graphics.ShaderConstraints, expectedVertexCode, expectedFragmentCode string) {
		translator := glsl.NewTranslator("300 es", true)
		program := translator.Translate(shader, constraints)
		Expect(program.VertexCode).To(Equal(expectedVertexCode))
		Expect(program.FragmentCode).To(Equal(expectedFragmentCode))
	},

		Entry("empty",
			&lsl.Shader{},
			graphics.ShaderConstraints{},
			openTestFile("translate-300-es", "empty.vert.glsl"),
			openTestFile("translate-300-es", "empty.frag.glsl"),
		),
	)

	DescribeTable("Translate 410", func(shader *lsl.Shader, constraints graphics.ShaderConstraints, expectedVertexCode, expectedFragmentCode string) {
		translator := glsl.NewTranslator("410", false)
		program := translator.Translate(shader, constraints)
		Expect(program.VertexCode).To(Equal(expectedVertexCode))
		Expect(program.FragmentCode).To(Equal(expectedFragmentCode))
	},

		Entry("empty",
			&lsl.Shader{},
			graphics.ShaderConstraints{},
			openTestFile("translate-410", "empty.vert.glsl"),
			openTestFile("translate-410", "empty.frag.glsl"),
		),

		Entry("coord attribute",
			&lsl.Shader{},
			graphics.ShaderConstraints{
				HasCoords: true,
			},
			openTestFile("translate-410", "attrib-coord.vert.glsl"),
			openTestFile("translate-410", "empty.frag.glsl"),
		),

		Entry("normal attribute",
			&lsl.Shader{},
			graphics.ShaderConstraints{
				HasNormals: true,
			},
			openTestFile("translate-410", "attrib-normal.vert.glsl"),
			openTestFile("translate-410", "empty.frag.glsl"),
		),

		Entry("tangent attribute",
			&lsl.Shader{},
			graphics.ShaderConstraints{
				HasTangents: true,
			},
			openTestFile("translate-410", "attrib-tangent.vert.glsl"),
			openTestFile("translate-410", "empty.frag.glsl"),
		),

		Entry("tex coord attribute",
			&lsl.Shader{},
			graphics.ShaderConstraints{
				HasTexCoords: true,
			},
			openTestFile("translate-410", "attrib-tex-coord.vert.glsl"),
			openTestFile("translate-410", "empty.frag.glsl"),
		),

		Entry("color attribute",
			&lsl.Shader{},
			graphics.ShaderConstraints{
				HasVertexColors: true,
			},
			openTestFile("translate-410", "attrib-color.vert.glsl"),
			openTestFile("translate-410", "empty.frag.glsl"),
		),

		Entry("armature attributes",
			&lsl.Shader{},
			graphics.ShaderConstraints{
				HasArmature: true,
			},
			openTestFile("translate-410", "attrib-armature.vert.glsl"),
			openTestFile("translate-410", "empty.frag.glsl"),
		),

		Entry("framebuffer output 0",
			&lsl.Shader{},
			graphics.ShaderConstraints{
				HasOutput0: true,
			},
			openTestFile("translate-410", "empty.vert.glsl"),
			openTestFile("translate-410", "framebuffer-output-0.frag.glsl"),
		),

		Entry("framebuffer output 1",
			&lsl.Shader{},
			graphics.ShaderConstraints{
				HasOutput1: true,
			},
			openTestFile("translate-410", "empty.vert.glsl"),
			openTestFile("translate-410", "framebuffer-output-1.frag.glsl"),
		),

		Entry("framebuffer output 2",
			&lsl.Shader{},
			graphics.ShaderConstraints{
				HasOutput2: true,
			},
			openTestFile("translate-410", "empty.vert.glsl"),
			openTestFile("translate-410", "framebuffer-output-2.frag.glsl"),
		),

		Entry("framebuffer output 3",
			&lsl.Shader{},
			graphics.ShaderConstraints{
				HasOutput3: true,
			},
			openTestFile("translate-410", "empty.vert.glsl"),
			openTestFile("translate-410", "framebuffer-output-3.frag.glsl"),
		),

		Entry("camera",
			&lsl.Shader{},
			graphics.ShaderConstraints{
				HasCamera: true,
			},
			openTestFile("translate-410", "camera.vert.glsl"),
			openTestFile("translate-410", "camera.frag.glsl"),
		),

		Entry("textures",
			&lsl.Shader{
				Declarations: []lsl.Declaration{
					&lsl.TextureBlockDeclaration{
						Fields: []lsl.Field{
							{
								Name: "first",
								Type: lsl.TypeNameSampler2D,
							},
							{
								Name: "second",
								Type: lsl.TypeNameSamplerCube,
							},
						},
					},
					&lsl.TextureBlockDeclaration{
						Fields: []lsl.Field{
							{
								Name: "third",
								Type: lsl.TypeNameSampler2D,
							},
						},
					},
				},
			},
			graphics.ShaderConstraints{},
			openTestFile("translate-410", "textures.vert.glsl"),
			openTestFile("translate-410", "textures.frag.glsl"),
		),

		Entry("uniforms",
			&lsl.Shader{
				Declarations: []lsl.Declaration{
					&lsl.UniformBlockDeclaration{
						Fields: []lsl.Field{
							{
								Name: "first",
								Type: lsl.TypeNameVec4,
							},
							{
								Name: "second",
								Type: lsl.TypeNameUVec4,
							},
						},
					},
					&lsl.UniformBlockDeclaration{
						Fields: []lsl.Field{
							{
								Name: "third",
								Type: lsl.TypeNameVec2,
							},
						},
					},
				},
			},
			graphics.ShaderConstraints{},
			openTestFile("translate-410", "uniforms.vert.glsl"),
			openTestFile("translate-410", "uniforms.frag.glsl"),
		),

		Entry("varyings",
			&lsl.Shader{
				Declarations: []lsl.Declaration{
					&lsl.VaryingBlockDeclaration{
						Fields: []lsl.Field{
							{
								Name: "first",
								Type: lsl.TypeNameFloat,
							},
							{
								Name: "second",
								Type: lsl.TypeNameVec2,
							},
						},
					},
					&lsl.VaryingBlockDeclaration{
						Fields: []lsl.Field{
							{
								Name: "third",
								Type: lsl.TypeNameVec4,
							},
						},
					},
				},
			},
			graphics.ShaderConstraints{},
			openTestFile("translate-410", "varyings.vert.glsl"),
			openTestFile("translate-410", "varyings.frag.glsl"),
		),
	)
})
