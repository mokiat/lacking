package graphics

import (
	"github.com/mokiat/lacking/game/graphics/internal"
	"github.com/mokiat/lacking/render"
)

// Shading represents an algorithm for shading a particular mesh and material.
type Shading interface {

	// ShadowPipeline constructs a render Pipeline for the specified mesh and
	// material definitions to be used in the shadow pass.
	ShadowPipeline(meshDef *MeshDefinition, fragmentDef *MeshFragmentDefinition) render.Pipeline

	// GeometryPipeline constructs a render Pipeline for the specified mesh and
	// material definitions to be used in the geometry pass.
	GeometryPipeline(meshDef *MeshDefinition, fragmentDef *MeshFragmentDefinition) render.Pipeline

	// EmissivePipeline constructs a render Pipeline for the specified mesh and
	// material definitions to be used in the emissive pass.
	EmissivePipeline(meshDef *MeshDefinition, fragmentDef *MeshFragmentDefinition) render.Pipeline

	// ForwardPipeline constructs a render Pipeline for the specified mesh and
	// material definitions to be used in the forward pass.
	ForwardPipeline(meshDef *MeshDefinition, fragmentDef *MeshFragmentDefinition) render.Pipeline
}

type pbrShading struct {
	api     render.API
	shaders ShaderCollection
}

func (s *pbrShading) GeometryPipeline(meshDef *MeshDefinition, fragmentDef *MeshFragmentDefinition) render.Pipeline {
	material := fragmentDef.material
	materialDef := material.definition
	// TODO: Cache programs
	shaderSet := s.shaders.PBRGeometrySet(PBRGeometryShaderConfig{
		HasArmature:      meshDef.hasArmature,
		HasAlphaTesting:  materialDef.alphaTesting,
		HasAlbedoTexture: len(materialDef.twoDTextures) > 0 && materialDef.twoDTextures[0] != nil,
	})
	program := internal.NewGeometryProgram(s.api, shaderSet.VertexShader, shaderSet.FragmentShader)
	cullMode := render.CullModeNone
	if materialDef.backfaceCulling {
		cullMode = render.CullModeBack
	}
	return s.api.CreatePipeline(render.PipelineInfo{
		Program:         program,
		VertexArray:     meshDef.vertexArray,
		Topology:        fragmentDef.topology,
		Culling:         cullMode,
		FrontFace:       render.FaceOrientationCCW,
		DepthTest:       true,
		DepthWrite:      true,
		DepthComparison: render.ComparisonLessOrEqual,
		StencilTest:     false,
		ColorWrite:      render.ColorMaskTrue,
		BlendEnabled:    false,
	})
}

func (s *pbrShading) EmissivePipeline(meshDef *MeshDefinition, fragmentDef *MeshFragmentDefinition) render.Pipeline {
	return nil // TODO
}

func (s *pbrShading) ShadowPipeline(meshDef *MeshDefinition, fragmentDef *MeshFragmentDefinition) render.Pipeline {
	material := fragmentDef.material
	materialDef := material.definition
	// TODO: Cache programs
	shaderSet := s.shaders.ShadowMappingSet(ShadowMappingShaderConfig{
		HasArmature: meshDef.hasArmature,
	})
	program := internal.NewShadowProgram(s.api, shaderSet.VertexShader, shaderSet.FragmentShader)
	cullMode := render.CullModeNone
	if materialDef.backfaceCulling {
		cullMode = render.CullModeBack
	}
	return s.api.CreatePipeline(render.PipelineInfo{
		Program:         program,
		VertexArray:     meshDef.vertexArray,
		Topology:        fragmentDef.topology,
		Culling:         cullMode,
		FrontFace:       render.FaceOrientationCCW,
		DepthTest:       true,
		DepthWrite:      true,
		DepthComparison: render.ComparisonLess,
		StencilTest:     false,
		ColorWrite:      render.ColorMaskFalse,
		BlendEnabled:    false,
	})
}

func (s *pbrShading) ForwardPipeline(meshDef *MeshDefinition, fragmentDef *MeshFragmentDefinition) render.Pipeline {
	return nil // TODO
}
