package graphics

import (
	"github.com/mokiat/lacking/game/graphics/internal"
	"github.com/mokiat/lacking/game/graphics/shading"
	"github.com/mokiat/lacking/render"
)

type ShadingInfo struct {
	ShadowFunc   shading.ShadowFunc
	GeometryFunc shading.GeometryFunc
	EmissiveFunc shading.EmissiveFunc
	ForwardFunc  shading.ForwardFunc
	LightingFunc shading.LightingFunc
}

// Shading represents an algorithm for shading a particular mesh and material.
type Shading interface {

	// ShadowPipeline constructs a render Pipeline for the specified mesh and
	// material definitions to be used in the shadow pass.
	ShadowPipeline(meshDef *MeshDefinition, fragmentDef *meshFragmentDefinition) render.Pipeline

	// GeometryPipeline constructs a render Pipeline for the specified mesh and
	// material definitions to be used in the geometry pass.
	GeometryPipeline(meshDef *MeshDefinition, fragmentDef *meshFragmentDefinition) render.Pipeline

	// EmissivePipeline constructs a render Pipeline for the specified mesh and
	// material definitions to be used in the emissive pass.
	EmissivePipeline(meshDef *MeshDefinition, fragmentDef *meshFragmentDefinition) render.Pipeline

	// ForwardPipeline constructs a render Pipeline for the specified mesh and
	// material definitions to be used in the forward pass.
	ForwardPipeline(meshDef *MeshDefinition, fragmentDef *meshFragmentDefinition) render.Pipeline
}

type customShading struct {
	api     render.API
	shaders ShaderCollection // FIXME: Only builder is of interest.

	shadowFunc   shading.ShadowFunc
	geometryFunc shading.GeometryFunc
	emissiveFunc shading.EmissiveFunc
	forwardFunc  shading.ForwardFunc
}

func (s *customShading) ShadowPipeline(meshDef *MeshDefinition, fragmentDef *meshFragmentDefinition) render.Pipeline {
	if s.shadowFunc == nil {
		return nil
	}
	return nil // TODO
}

func (s *customShading) GeometryPipeline(meshDef *MeshDefinition, fragmentDef *meshFragmentDefinition) render.Pipeline {
	if s.geometryFunc == nil {
		return nil
	}

	materialDef := fragmentDef.material.definition
	meshCfg := MeshConfig{
		HasArmature: meshDef.needsArmature,
	}
	shaderSet := s.shaders.BuildGeometry(meshCfg, s.geometryFunc)
	program := internal.NewGeometryProgram(s.api, shaderSet.VertexShader, shaderSet.FragmentShader)
	cullMode := render.CullModeNone
	if materialDef.backfaceCulling {
		cullMode = render.CullModeBack
	}
	return s.api.CreatePipeline(render.PipelineInfo{
		Program: program,

		// TODO: Move mesh outside pipeline for better reuse.
		VertexArray: meshDef.vertexArray,
		Topology:    fragmentDef.topology,

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

func (s *customShading) EmissivePipeline(meshDef *MeshDefinition, fragmentDef *meshFragmentDefinition) render.Pipeline {
	if s.emissiveFunc == nil {
		return nil
	}
	return nil // TODO
}

func (s *customShading) ForwardPipeline(meshDef *MeshDefinition, fragmentDef *meshFragmentDefinition) render.Pipeline {
	if s.forwardFunc == nil {
		return nil
	}

	materialDef := fragmentDef.material.definition
	meshCfg := MeshConfig{
		HasArmature: meshDef.needsArmature,
	}
	shaderSet := s.shaders.BuildForward(meshCfg, s.forwardFunc)
	program := internal.NewGeometryProgram(s.api, shaderSet.VertexShader, shaderSet.FragmentShader)
	cullMode := render.CullModeNone
	if materialDef.backfaceCulling {
		cullMode = render.CullModeBack
	}
	return s.api.CreatePipeline(render.PipelineInfo{
		Program: program,

		// TODO: Move mesh outside pipeline for better reuse.
		VertexArray: meshDef.vertexArray,
		Topology:    fragmentDef.topology,

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

type pbrShading struct {
	api     render.API
	shaders ShaderCollection
}

func (s *pbrShading) GeometryPipeline(meshDef *MeshDefinition, fragmentDef *meshFragmentDefinition) render.Pipeline {
	material := fragmentDef.material
	materialDef := material.definition
	// TODO: Cache programs
	shaderSet := s.shaders.PBRGeometrySet(PBRGeometryShaderConfig{
		HasArmature:      meshDef.needsArmature,
		HasAlphaTesting:  materialDef.alphaTesting,
		HasVertexColors:  meshDef.hasVertexColors,
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

func (s *pbrShading) EmissivePipeline(meshDef *MeshDefinition, fragmentDef *meshFragmentDefinition) render.Pipeline {
	return nil // TODO
}

func (s *pbrShading) ShadowPipeline(meshDef *MeshDefinition, fragmentDef *meshFragmentDefinition) render.Pipeline {
	material := fragmentDef.material
	materialDef := material.definition
	// TODO: Cache programs
	shaderSet := s.shaders.ShadowMappingSet(ShadowMappingShaderConfig{
		HasArmature: meshDef.needsArmature,
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

func (s *pbrShading) ForwardPipeline(meshDef *MeshDefinition, fragmentDef *meshFragmentDefinition) render.Pipeline {
	return nil // TODO
}
