package graphics

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/graphics/internal"
	"github.com/mokiat/lacking/render"
)

// MaterialInfo contains the information needed to create a Material.
type MaterialInfo struct {

	// Name specifies the name of the material.
	Name string

	// GeometryPasses specifies a list of geometry passes to be applied.
	// This is used for opaque materials that go through deferred shading.
	GeometryPasses []GeometryRenderPassInfo

	// ShadowPasses specifies a list of shadow passes to be applied.
	// This should be omitted if the material does not cast shadows.
	ShadowPasses []ShadowRenderPassInfo

	// ForwardPasses specifies a list of forward passes to be applied.
	// This is useful when dealing with translucent materials or when special
	// effects are needed.
	ForwardPasses []ForwardRenderPassInfo
}

// Material determines the appearance of a mesh on the screen.
type Material struct {
	name string

	// TODO: Restructure in fixed array
	geometryPasses []internal.MaterialRenderPassDefinition
	shadowPasses   []internal.MaterialRenderPassDefinition
	forwardPasses  []internal.MaterialRenderPassDefinition
}

func (m *Material) Name() string {
	return m.name
}

// Texture returns the texture with the specified name.
// If the texture is not found, nil is returned.
func (m *Material) Texture(name string) render.Texture {
	var result render.Texture
	m.eachPass(func(pass *internal.MaterialRenderPassDefinition) {
		if texture := pass.TextureSet.Texture(name); texture != nil {
			result = texture
		}
	})
	return result
}

// SetTexture sets the texture with the specified name.
func (m *Material) SetTexture(name string, texture render.Texture) {
	m.eachPass(func(pass *internal.MaterialRenderPassDefinition) {
		pass.TextureSet.SetTexture(name, texture)
	})
}

// Sampler returns the sampler with the specified name.
// If the sampler is not found, nil is returned.
func (m *Material) Sampler(name string) render.Sampler {
	var result render.Sampler
	m.eachPass(func(pass *internal.MaterialRenderPassDefinition) {
		if sampler := pass.TextureSet.Sampler(name); sampler != nil {
			result = sampler
		}
	})
	return result
}

// SetSampler sets the sampler with the specified name.
func (m *Material) SetSampler(name string, sampler render.Sampler) {
	m.eachPass(func(pass *internal.MaterialRenderPassDefinition) {
		pass.TextureSet.SetSampler(name, sampler)
	})
}

// Property returns the property with the specified name.
// If the property is not found, nil is returned.
func (m *Material) Property(name string) any {
	var result any
	m.eachPass(func(pass *internal.MaterialRenderPassDefinition) {
		if value := pass.UniformSet.Property(name); value != nil {
			result = value
		}
	})
	return result
}

// SetProperty sets the property with the specified name.
func (m *Material) SetProperty(name string, value any) {
	m.eachPass(func(pass *internal.MaterialRenderPassDefinition) {
		pass.UniformSet.SetProperty(name, value)
	})
}

func (m *Material) eachPass(cb func(pass *internal.MaterialRenderPassDefinition)) {
	for i := range m.geometryPasses {
		cb(&m.geometryPasses[i])
	}
	for i := range m.shadowPasses {
		cb(&m.shadowPasses[i])
	}
	for i := range m.forwardPasses {
		cb(&m.forwardPasses[i])
	}
}

/// OLD STUFF BELOW

// PBRMaterialInfo contains the information needed to create a PBR Material.
type PBRMaterialInfo struct {
	BackfaceCulling          bool
	AlphaBlending            bool
	AlphaTesting             bool
	AlphaThreshold           float32
	Metallic                 float32
	Roughness                float32
	MetallicRoughnessTexture render.Texture
	AlbedoColor              sprec.Vec4
	AlbedoTexture            render.Texture
	NormalScale              float32
	NormalTexture            render.Texture
	EmissiveColor            sprec.Vec4
}
