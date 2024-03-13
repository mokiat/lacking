package internal

import "github.com/mokiat/lacking/render"

type MeshRenderPassType uint8

const (
	MeshRenderPassTypeShadow MeshRenderPassType = iota
	MeshRenderPassTypeGeometry
	MeshRenderPassTypeForward
	MeshRenderPassTypeCount
)

type MeshRenderPassDefinition struct {
	Layer           int32
	Program         render.Program
	Pipeline        render.Pipeline
	IndexByteOffset uint32
	IndexCount      uint32
}

type MeshRenderPass struct {
	MeshRenderPassDefinition
	Key          uint32
	Textures     [8]render.Texture
	Samplers     [8]render.Sampler
	MaterialData []byte
}

// TODO: Rename to non-definition
type MaterialRenderPassDefinition struct {
	Layer           int32
	Culling         render.CullMode
	FrontFace       render.FaceOrientation
	DepthTest       bool
	DepthWrite      bool
	DepthComparison render.Comparison
	Blending        bool

	Textures    [8]render.Texture
	Samplers    [8]render.Sampler
	UniformData []byte

	Shader Shader
}

type RenderPassPipelineInfo struct {
	Program          render.Program
	MeshVertexArray  render.VertexArray
	FragmentTopology render.Topology
	PassDefinition   MaterialRenderPassDefinition
}
