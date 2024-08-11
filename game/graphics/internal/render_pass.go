package internal

import (
	"github.com/mokiat/lacking/game/graphics/lsl"
	"github.com/mokiat/lacking/render"
)

type MaterialRenderPass struct {
	Layer           int32
	Culling         render.CullMode
	FrontFace       render.FaceOrientation
	DepthTest       bool
	DepthWrite      bool
	DepthComparison render.Comparison
	Blending        bool
	TextureSet      TextureSet
	UniformSet      UniformSet
	// TODO: Add blending
	Shader *lsl.Shader
}

type MeshRenderPassType uint8

const (
	MeshRenderPassTypeShadow MeshRenderPassType = iota
	MeshRenderPassTypeGeometry
	MeshRenderPassTypeForward
	MeshRenderPassTypeSky
	MeshRenderPassTypePostprocess
	MeshRenderPassTypeCount
)

type MeshRenderPass struct {
	Layer           int32
	Program         render.Program
	Pipeline        render.Pipeline
	IndexByteOffset uint32
	IndexCount      uint32

	Key        uint32
	TextureSet TextureSet
	UniformSet UniformSet
}

type RenderPassPipelineInfo struct {
	Program          render.Program
	MeshVertexArray  render.VertexArray
	FragmentTopology render.Topology
	PassDefinition   MaterialRenderPass
}
