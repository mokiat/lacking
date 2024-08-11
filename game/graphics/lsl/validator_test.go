package lsl_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/mokiat/lacking/game/graphics/lsl"
)

var _ = Describe("Validator", func() {
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

	When("texture block has duplicate field name", func() {
		BeforeEach(func() {
			inShader = lsl.MustParse(`
				textures {
					diffuse sampler2D,
					reflection samplerCube,
					diffuse samplerCube,
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

	When("uniform block has duplicate field name", func() {
		BeforeEach(func() {
			inShader = lsl.MustParse(`
				uniforms {
					color vec4,
					normal vec3,
					color vec3,
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

	When("valid varying block", func() {
		BeforeEach(func() {
			inShader = lsl.MustParse(`
				varyings {
					color vec4,
					normal vec3,
				}
			`)
		})
		itPassesValidation()
	})

	When("multiple varying blocks", func() {
		BeforeEach(func() {
			inShader = lsl.MustParse(`
				varyings {
					color vec4,
				}
				varyings {
					normal vec3,
				}
			`)
		})
		itFailsValidation()
	})

	When("varying block has duplicate field name", func() {
		BeforeEach(func() {
			inShader = lsl.MustParse(`
				varyings {
					color vec4,
					normal vec3,
					color vec3,
				}
			`)
		})
		itFailsValidation()
	})

	When("varying block with protected field name", func() {
		BeforeEach(func() {
			inShader = lsl.MustParse(`
				varyings {
					#color vec4,
					#normal vec3,
				}
			`)
		})
		itFailsValidation()
	})

	When("varying block with unsupported type", func() {
		BeforeEach(func() {
			inShader = lsl.MustParse(`
				varyings {
					color sampler2D,
				}
			`)
		})
		itFailsValidation()
	})

	When("texture and uniform blocks share field name", func() {
		BeforeEach(func() {
			inShader = lsl.MustParse(`
				textures {
					color sampler2D,
				}
				uniforms {
					color vec4,
				}
			`)
		})
		itFailsValidation()
	})

	When("uniform and varying blocks share field name", func() {
		BeforeEach(func() {
			inShader = lsl.MustParse(`
				uniforms {
					color vec4,
				}
				varyings {
					color vec4,
				}
			`)
		})
		itFailsValidation()
	})

	When("declaring #vertex function", func() {
		BeforeEach(func() {
			inShader = lsl.MustParse(`
				func #vertex() {
				}
			`)
		})
		itPassesValidation()
	})

	When("declaring multiple #vertex functions", func() {
		BeforeEach(func() {
			inShader = lsl.MustParse(`
				func #vertex() {
				}
				func #vertex() {
				}
			`)
		})
		itFailsValidation()
	})

	When("declaring #fragment function", func() {
		BeforeEach(func() {
			inShader = lsl.MustParse(`
				func #fragment() {
				}
			`)
		})
		itPassesValidation()
	})

	When("declaring multiple #fragment functions", func() {
		BeforeEach(func() {
			inShader = lsl.MustParse(`
				func #fragment() {
				}
				func #fragment() {
				}
			`)
		})
		itFailsValidation()
	})

	When("declaring a custom function", func() {
		// NOTE: Custom functions are not supported yet.
		BeforeEach(func() {
			inShader = lsl.MustParse(`
				func custom() {
				}
			`)
		})
		itFailsValidation()
	})

	When("declaring a valid variable", func() {
		BeforeEach(func() {
			inShader = lsl.MustParse(`
				func #fragment() {
					var color vec4
				}
			`)
		})
		itPassesValidation()
	})

	When("declaring a variable with protected name", func() {
		BeforeEach(func() {
			inShader = lsl.MustParse(`
				func #fragment() {
					var #color vec4
				}
			`)
		})
		itFailsValidation()
	})

	When("declaring duplicate variables", func() {
		BeforeEach(func() {
			inShader = lsl.MustParse(`
				func #fragment() {
					var color vec4
					var color vec3
				}
			`)
		})
		itFailsValidation()
	})

	When("declaring a variable of unknown type", func() {
		BeforeEach(func() {
			inShader = lsl.MustParse(`
				func #fragment() {
					var color unknown
				}
			`)
		})
		itFailsValidation()
	})

})
