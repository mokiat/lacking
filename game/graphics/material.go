package graphics

import (
	"github.com/mokiat/lacking/game/graphics/internal"
	"github.com/mokiat/lacking/render"
)

// MaterialInfo contains the information needed to create a Material.
type MaterialInfo struct {

	// Name specifies the name of the material.
	Name string

	// GeometryPasses specifies a list of geometry passes to be applied.
	// This is used for opaque materials that go through deferred shading.
	GeometryPasses []MaterialPassInfo

	// ShadowPasses specifies a list of shadow passes to be applied.
	// This should be omitted if the material does not cast shadows.
	ShadowPasses []MaterialPassInfo

	// ForwardPasses specifies a list of forward passes to be applied.
	// This is useful when dealing with translucent materials or when special
	// effects are needed.
	ForwardPasses []MaterialPassInfo

	// SkyPasses specifies a list of sky passes to be applied.
	// This is used for materials that are used to render the sky.
	SkyPasses []MaterialPassInfo

	// PostprocessingPasses specifies a list of postprocess passes to be applied.
	PostprocessingPasses []MaterialPassInfo
}

// Material determines the appearance of a mesh on the screen.
type Material struct {
	name string

	geometryPasses    []internal.MaterialRenderPass
	shadowPasses      []internal.MaterialRenderPass
	forwardPasses     []internal.MaterialRenderPass
	skyPasses         []internal.MaterialRenderPass
	postprocessPasses []internal.MaterialRenderPass
}

func (m *Material) Name() string {
	return m.name
}

// Texture returns the texture with the specified name.
// If the texture is not found, nil is returned.
func (m *Material) Texture(name string) render.Texture {
	var result render.Texture
	m.eachPass(func(pass *internal.MaterialRenderPass) {
		if texture := pass.TextureSet.Texture(name); texture != nil {
			result = texture
		}
	})
	return result
}

// SetTexture sets the texture with the specified name.
func (m *Material) SetTexture(name string, texture render.Texture) {
	m.eachPass(func(pass *internal.MaterialRenderPass) {
		pass.TextureSet.SetTexture(name, texture)
	})
}

// Sampler returns the sampler with the specified name.
// If the sampler is not found, nil is returned.
func (m *Material) Sampler(name string) render.Sampler {
	var result render.Sampler
	m.eachPass(func(pass *internal.MaterialRenderPass) {
		if sampler := pass.TextureSet.Sampler(name); sampler != nil {
			result = sampler
		}
	})
	return result
}

// SetSampler sets the sampler with the specified name.
func (m *Material) SetSampler(name string, sampler render.Sampler) {
	m.eachPass(func(pass *internal.MaterialRenderPass) {
		pass.TextureSet.SetSampler(name, sampler)
	})
}

// Property returns the property with the specified name.
// If the property is not found, nil is returned.
func (m *Material) Property(name string) any {
	var result any
	m.eachPass(func(pass *internal.MaterialRenderPass) {
		if value := pass.UniformSet.Property(name); value != nil {
			result = value
		}
	})
	return result
}

// SetProperty sets the property with the specified name.
func (m *Material) SetProperty(name string, value any) {
	m.eachPass(func(pass *internal.MaterialRenderPass) {
		pass.UniformSet.SetProperty(name, value)
	})
}

func (m *Material) eachPass(cb func(pass *internal.MaterialRenderPass)) {
	for i := range m.geometryPasses {
		cb(&m.geometryPasses[i])
	}
	for i := range m.shadowPasses {
		cb(&m.shadowPasses[i])
	}
	for i := range m.forwardPasses {
		cb(&m.forwardPasses[i])
	}
	for i := range m.skyPasses {
		cb(&m.skyPasses[i])
	}
	for i := range m.postprocessPasses {
		cb(&m.postprocessPasses[i])
	}
}
