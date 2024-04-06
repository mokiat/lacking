package graphics

import (
	"github.com/mokiat/gog"
	"github.com/mokiat/lacking/game/graphics/internal"
	"github.com/mokiat/lacking/render"
)

// SkyDefinitionInfo holds the information required to create a new sky
// definition.
type SkyDefinitionInfo struct {
	Layers []SkyLayerDefinitionInfo
}

// SkyLayerDefinitionInfo holds the information required to create a new sky
// layer definition.
type SkyLayerDefinitionInfo struct {
	Shader   *SkyShader
	Blending bool
}

func newSkyDefinition(engine *Engine, info SkyDefinitionInfo) *SkyDefinition {
	return &SkyDefinition{
		layers: gog.Map(info.Layers, func(layerInfo SkyLayerDefinitionInfo) internal.SkyLayerDefinition {
			return newSkyLayerDefinition(engine, layerInfo)
		}),
	}
}

// SkyDefinition represents a sky that can be rendered in the scene.
type SkyDefinition struct {
	layers []internal.SkyLayerDefinition
}

// Texture returns the texture with the specified name.
// If the texture is not found, nil is returned.
func (d *SkyDefinition) Texture(name string) render.Texture {
	for i := range d.layers {
		if texture := d.layers[i].TextureSet.Texture(name); texture != nil {
			return texture
		}
	}
	return nil
}

// SetTexture sets the texture with the specified name.
func (d *SkyDefinition) SetTexture(name string, texture render.Texture) {
	gog.Mutate(d.layers, func(layer *internal.SkyLayerDefinition) {
		layer.TextureSet.SetTexture(name, texture)
	})
}

// Sampler returns the sampler with the specified name.
// If the sampler is not found, nil is returned.
func (d *SkyDefinition) Sampler(name string) render.Sampler {
	for i := range d.layers {
		if sampler := d.layers[i].TextureSet.Sampler(name); sampler != nil {
			return sampler
		}
	}
	return nil
}

// SetSampler sets the sampler with the specified name.
func (d *SkyDefinition) SetSampler(name string, sampler render.Sampler) {
	gog.Mutate(d.layers, func(layer *internal.SkyLayerDefinition) {
		layer.TextureSet.SetSampler(name, sampler)
	})
}

// Property returns the property with the specified name.
// If the property is not found, nil is returned.
func (d *SkyDefinition) Property(name string) any {
	for i := range d.layers {
		if value := d.layers[i].UniformSet.Property(name); value != nil {
			return value
		}
	}
	return nil
}

// SetProperty sets the property with the specified name.
func (d *SkyDefinition) SetProperty(name string, value any) {
	gog.Mutate(d.layers, func(layer *internal.SkyLayerDefinition) {
		layer.UniformSet.SetProperty(name, value)
	})
}

// Delete deletes the sky definition and releases its resources.
func (d *SkyDefinition) Delete() {
	gog.Mutate(d.layers, func(layer *internal.SkyLayerDefinition) {
		layer.Delete()
	})
}

func newSkyLayerDefinition(engine *Engine, info SkyLayerDefinitionInfo) internal.SkyLayerDefinition {
	programCode := info.Shader.createProgramCode()
	program := engine.createSkyProgram(programCode)
	pipeline, indexByteOffset, indexCount := engine.createSkyPipeline(internal.SkyPipelineInfo{
		Program:  program,
		Blending: info.Blending,
	})

	return internal.SkyLayerDefinition{
		TextureSet: internal.NewShaderTextureSet(info.Shader.ast),
		UniformSet: internal.NewShaderUniformSet(info.Shader.ast),

		Program:         program,
		Pipeline:        pipeline,
		IndexByteOffset: indexByteOffset,
		IndexCount:      indexCount,
	}
}
