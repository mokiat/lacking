package graphics

import (
	"github.com/mokiat/lacking/game/graphics/internal"
)

// SkyDefinitionInfo holds the information required to create a new sky
// definition.
type SkyDefinitionInfo struct {

	// Material is the material that is used to render the sky.
	Material *Material
}

func newSkyDefinition(engine *Engine, info SkyDefinitionInfo) *SkyDefinition {
	def := &SkyDefinition{
		engine:   engine,
		material: info.Material,
	}
	def.createPipeline()
	return def
}

// SkyDefinition represents a sky that can be rendered in the scene.
type SkyDefinition struct {
	engine       *Engine
	material     *Material
	renderPasses []internal.MeshRenderPass
}

// Material returns the material that is used to render the sky.
func (d *SkyDefinition) Material() *Material {
	return d.material
}

// SetMaterial sets the material that is used to render the sky.
func (d *SkyDefinition) SetMaterial(material *Material) {
	if material != d.material {
		d.material = material
		d.deletePipeline()
		d.createPipeline()
	}
}

// Delete deletes the sky definition and releases its resources.
func (d *SkyDefinition) Delete() {
	d.deletePipeline()
}

func (d *SkyDefinition) createPipeline() {
	d.renderPasses = make([]internal.MeshRenderPass, len(d.material.skyPasses))
	for i, pass := range d.material.skyPasses {
		programCode := d.engine.createProgramCode(pass.Shader, internal.ShaderProgramCodeInfo{
			ShaderMeshInfo: internal.ShaderMeshInfo{
				MeshHasCoords: true,
			},
		})
		program := d.engine.createSkyProgram(programCode, pass.Shader)
		pipeline, indexByteOffset, indexCount := d.engine.createSkyPipeline(internal.SkyPipelineInfo{
			Program:  program,
			Blending: pass.Blending,
		})
		textureSet := pass.TextureSet
		uniformSet := pass.UniformSet

		d.renderPasses[i] = internal.MeshRenderPass{
			Layer:           pass.Layer,
			Program:         program,
			Pipeline:        pipeline,
			IndexByteOffset: indexByteOffset,
			IndexCount:      indexCount,
			Key:             d.engine.pickFreeRenderPassKey(),
			TextureSet:      textureSet,
			UniformSet:      uniformSet,
		}
	}
}

func (d *SkyDefinition) deletePipeline() {
	for _, pass := range d.renderPasses {
		pass.Program.Release()
		pass.Pipeline.Release()
	}
	d.renderPasses = nil
}
