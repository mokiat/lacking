package graphics

import (
	"github.com/mokiat/lacking/game/graphics/internal"
	"github.com/mokiat/lacking/render"
)

type SkyDefinitionInfo struct {
	Layers []SkyLayerDefinitionInfo
}

type SkyLayerDefinitionInfo struct {
	Shader      SkyShader
	Textures    [8]render.Texture
	Samplers    [8]render.Sampler
	UniformData []byte
	Blending    bool
}

type SkyDefinition struct {
	layers []SkyLayerDefinition
}

func (d *SkyDefinition) Delete() {
	for i := range d.layers {
		layer := &d.layers[i]
		layer.delete()
	}
}

func (d *SkyDefinition) LayerCount() int {
	return len(d.layers)
}

func (d *SkyDefinition) Layer(index int) *SkyLayerDefinition {
	return &d.layers[index]
}

type SkyLayerDefinition struct {
	engine *Engine

	shader   SkyShader
	blending bool

	program  render.Program
	pipeline render.Pipeline

	textures    [8]render.Texture
	samplers    [8]render.Sampler
	uniformData []byte

	indexByteOffset uint32
	indexCount      uint32
}

func (d *SkyLayerDefinition) Shader() SkyShader {
	return d.shader
}

func (d *SkyLayerDefinition) SetShader(shader SkyShader) {
	d.deletePipeline()
	d.shader = shader
	d.createPipeline()
}

func (d *SkyLayerDefinition) Blending() bool {
	return d.blending
}

func (d *SkyLayerDefinition) SetBlending(blending bool) {
	d.deletePipeline()
	d.blending = blending
	d.createPipeline()
}

func (d *SkyLayerDefinition) delete() {
	d.deletePipeline()
	d.shader = nil
	d.program = nil
	d.pipeline = nil
	d.engine = nil
}

func (d *SkyLayerDefinition) deletePipeline() {
	defer d.program.Release()
	defer d.pipeline.Release()
}

func (d *SkyLayerDefinition) createPipeline() {
	programCode := d.shader.CreateProgramCode(internal.ShaderProgramCodeInfo{})
	d.program = d.engine.createSkyProgram(programCode)
	d.pipeline, d.indexByteOffset, d.indexCount = d.engine.createSkyPipeline(internal.SkyPipelineInfo{
		Program:  d.program,
		Blending: d.blending,
	})
}
