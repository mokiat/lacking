package lsl_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/lacking/game/graphics/lsl"
)

var _ = FDescribe("Validator", func() {
	var (
		inShader *lsl.Shader
		inSchema lsl.Schema
		outErr   error
	)

	itPassesValidation := func() {
		GinkgoHelper()
		It("passes validation", func() {
			Expect(outErr).ToNot(HaveOccurred())
		})
	}

	itFailsValidation := func() {
		GinkgoHelper()
		It("fails validation", func() {
			Expect(outErr).To(HaveOccurred())
		})
	}

	JustBeforeEach(func() {
		outErr = lsl.Validate(inShader, inSchema)
	})

	BeforeEach(func() {
		inSchema = lsl.DefaultSchema()
	})

	When("empty shader", func() {
		BeforeEach(func() {
			inShader = &lsl.Shader{}
		})
		itPassesValidation()
	})

	When("valid texture block", func() {
		BeforeEach(func() {
			inShader = lsl.MustParse(`
				textures {
					diffuse sampler2D,
					reflection samplerCube,
				}
			`)
		})
		itPassesValidation()
	})

	When("multiple texture blocks", func() {
		BeforeEach(func() {
			inShader = lsl.MustParse(`
				textures {
					diffuse sampler2D,
				}
				textures {
					reflection samplerCube,
				}
			`)
		})
		itFailsValidation()
	})

	When("texture block with protected field name", func() {
		BeforeEach(func() {
			inShader = lsl.MustParse(`
				textures {
					#diffuse sampler2D,
				}
			`)
		})
		itFailsValidation()
	})

	When("texture block with unsupported type", func() {
		BeforeEach(func() {
			inShader = lsl.MustParse(`
				textures {
					diffuse vec2,
				}
			`)
		})
		itFailsValidation()
	})

	When("valid uniform block", func() {
		BeforeEach(func() {
			inShader = lsl.MustParse(`
				uniforms {
					color vec4,
					normal vec3,
				}
			`)
		})
		itPassesValidation()
	})

	When("multiple uniform blocks", func() {
		BeforeEach(func() {
			inShader = lsl.MustParse(`
				uniforms {
					color vec4,
				}
				uniforms {
					normal vec3,
				}
			`)
		})
		itFailsValidation()
	})

	When("uniform block with protected field name", func() {
		BeforeEach(func() {
			inShader = lsl.MustParse(`
				uniforms {
					#color vec4,
					#normal vec3,
				}
			`)
		})
		itFailsValidation()
	})

	When("uniform block with unsupported type", func() {
		BeforeEach(func() {
			inShader = lsl.MustParse(`
				uniforms {
					color sampler2D,
				}
			`)
		})
		itFailsValidation()
	})

})
