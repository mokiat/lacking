package internal

import "github.com/mokiat/lacking/render"

type SkyPipelineInfo struct {
	Program  render.Program
	Blending bool
}

type SkyLayerDefinition struct {
	TextureSet TextureSet
	UniformSet UniformSet

	Program         render.Program
	Pipeline        render.Pipeline
	IndexByteOffset uint32
	IndexCount      uint32
}

func (d *SkyLayerDefinition) Delete() {
	defer d.Program.Release()
	defer d.Pipeline.Release()
}
