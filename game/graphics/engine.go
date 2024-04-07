package graphics

import (
	"fmt"

	"github.com/mokiat/gog"
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/debug/log"
	"github.com/mokiat/lacking/game/graphics/internal"
	"github.com/mokiat/lacking/game/graphics/lsl"
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/util/blob"
)

func NewEngine(api render.API, shaders ShaderCollection, shaderBuilder ShaderBuilder) *Engine {
	stageData := newCommonStageData(api)
	renderer := newRenderer(api, shaders, stageData)

	return &Engine{
		api:           api,
		shaders:       shaders,
		shaderBuilder: shaderBuilder,

		stageData: stageData,
		renderer:  renderer,

		debug: &Debug{
			renderer: renderer,
		},
	}
}

// Engine represents an entrypoint to 3D graphics rendering.
type Engine struct {
	api           render.API
	shaders       ShaderCollection
	shaderBuilder ShaderBuilder

	stageData *commonStageData
	renderer  *sceneRenderer

	debug *Debug

	freeRenderPassKey uint32
}

func (e *Engine) API() render.API {
	return e.api
}

// Create initializes this 3D engine.
func (e *Engine) Create() {
	e.stageData.Allocate()
	e.renderer.Allocate()
}

// Destroy releases resources allocated by this
// 3D engine.
func (e *Engine) Destroy() {
	defer e.stageData.Release()
	defer e.renderer.Release()
}

// Debug allows the rendering of debug lines on the screen.
//
// Deprecated: Figure out how to fix/improve this. Maybe not needed anymore
// with custom shaders and forward passes?
func (e *Engine) Debug() *Debug {
	return e.debug
}

// DefaultGeometryShader returns a basic implementation of a GeometryShader.
//
// Deprecated: Use custom shaders instead.
func (e *Engine) DefaultGeometryShader(alphaTesting, albedoTexture bool) GeometryShader {
	return &defaultGeometryShader{
		shaders:          e.shaders,
		useAlphaTesting:  alphaTesting,
		useAlbedoTexture: albedoTexture,
	}
}

// DefaultShadowShader returns a basic implementation of a ShadowShader.
//
// Deprecated: Use custom shaders instead.
func (e *Engine) DefaultShadowShader() ShadowShader {
	return &defaultShadowShader{
		shaders: e.shaders,
	}
}

// CreateGeometryShader creates a new custom GeometryShader using the
// specified info object.
func (e *Engine) CreateGeometryShader(info ShaderInfo) GeometryShader {
	ast, err := lsl.Parse(info.SourceCode)
	if err != nil {
		log.Error("Failed to parse geometry shader: %v", err)
		ast = &lsl.Shader{Declarations: []lsl.Declaration{}} // TODO: Something meaningful
	}
	// TODO: Validate against Geometry globals.
	return &customGeometryShader{
		builder: e.shaderBuilder,
		ast:     ast,
	}
}

// CreateShadowShader creates a new custom ShadowShader using the
// specified info object.
func (e *Engine) CreateShadowShader(info ShaderInfo) ShadowShader {
	ast, err := lsl.Parse(info.SourceCode)
	if err != nil {
		log.Error("Failed to parse shadow shader: %v", err)
		ast = &lsl.Shader{Declarations: []lsl.Declaration{}} // TODO: Something meaningful
	}
	// TODO: Validate against Shadow globals.
	return &customShadowShader{
		builder: e.shaderBuilder,
		ast:     ast,
	}
}

// CreateForwardShader creates a new custom ForwardShader using the
// specified info object.
func (e *Engine) CreateForwardShader(info ShaderInfo) ForwardShader {
	ast, err := lsl.Parse(info.SourceCode)
	if err != nil {
		log.Error("Failed to parse forward shader: %v", err)
		ast = &lsl.Shader{Declarations: []lsl.Declaration{}} // TODO: Something meaningful
	}
	// TODO: Validate against Forward globals.
	return &customForwardShader{
		builder: e.shaderBuilder,
		ast:     ast,
	}
}

// CreateSkyShader creates a new custom SkyShader using the
// specified info object.
func (e *Engine) CreateSkyShader(info ShaderInfo) *SkyShader {
	ast, err := lsl.Parse(info.SourceCode)
	if err != nil {
		log.Error("Failed to parse sky shader: %v", err)
		ast = &lsl.Shader{Declarations: []lsl.Declaration{}} // TODO: Something meaningful
	}
	// TODO: Validate against Sky globals.
	return &SkyShader{
		builder: e.shaderBuilder,
		ast:     ast,
	}
}

// CreateSkyDefinition creates a new SkyDefinition using the specified info
// object.
func (e *Engine) CreateSkyDefinition(info SkyDefinitionInfo) *SkyDefinition {
	return newSkyDefinition(e, info)
}

// CreateTwoDTexture creates a new TwoDTexture using the
// specified definition.
func (e *Engine) CreateTwoDTexture(definition TwoDTextureDefinition) *TwoDTexture {
	return newTwoDTexture(e.api.CreateColorTexture2D(render.ColorTexture2DInfo{
		Width:           definition.Width,
		Height:          definition.Height,
		GenerateMipmaps: definition.GenerateMipmaps,
		GammaCorrection: true,
		Format:          e.convertFormat(definition.DataFormat),
		Data:            definition.Data,
	}))
}

// CreateMaterial creates a new Material from the specified info object.
func (e *Engine) CreateMaterial(info MaterialInfo) *Material {
	geometryPasses := gog.Map(info.GeometryPasses, func(passInfo GeometryRenderPassInfo) internal.MaterialRenderPassDefinition {
		if len(passInfo.Textures) > 8 {
			passInfo.Textures = passInfo.Textures[:8]
		}

		var textures [8]render.Texture
		for i, textureInfo := range passInfo.Textures {
			textures[i] = textureInfo.Texture
		}

		var samplers [8]render.Sampler
		for i, textureInfo := range passInfo.Textures {
			samplers[i] = e.api.CreateSampler(render.SamplerInfo{
				Wrapping:   textureInfo.Wrapping,
				Filtering:  textureInfo.Filtering,
				Mipmapping: textureInfo.Mipmapping,
			})
		}

		return internal.MaterialRenderPassDefinition{
			Layer:           passInfo.Layer,
			Culling:         passInfo.Culling.ValueOrDefault(render.CullModeNone),
			FrontFace:       passInfo.FrontFace.ValueOrDefault(render.FaceOrientationCCW),
			DepthTest:       passInfo.DepthTest.ValueOrDefault(true),
			DepthWrite:      passInfo.DepthWrite.ValueOrDefault(true),
			DepthComparison: passInfo.DepthComparison.ValueOrDefault(render.ComparisonLessOrEqual),
			Blending:        false,

			Textures:    textures,
			Samplers:    samplers,
			UniformData: passInfo.MaterialDataStd140,

			Shader: passInfo.Shader,
		}
	})

	shadowPasses := gog.Map(info.ShadowPasses, func(passInfo ShadowRenderPassInfo) internal.MaterialRenderPassDefinition {
		if len(passInfo.Textures) > 8 {
			passInfo.Textures = passInfo.Textures[:8]
		}

		var textures [8]render.Texture
		for i, textureInfo := range passInfo.Textures {
			textures[i] = textureInfo.Texture
		}

		var samplers [8]render.Sampler
		for i, textureInfo := range passInfo.Textures {
			samplers[i] = e.api.CreateSampler(render.SamplerInfo{
				Wrapping:   textureInfo.Wrapping,
				Filtering:  textureInfo.Filtering,
				Mipmapping: textureInfo.Mipmapping,
			})
		}

		return internal.MaterialRenderPassDefinition{
			Layer:           0,
			Culling:         passInfo.Culling.ValueOrDefault(render.CullModeNone),
			FrontFace:       passInfo.FrontFace.ValueOrDefault(render.FaceOrientationCCW),
			DepthTest:       true,
			DepthWrite:      true,
			DepthComparison: render.ComparisonLessOrEqual,
			Blending:        false,

			Textures:    textures,
			Samplers:    samplers,
			UniformData: passInfo.MaterialDataStd140,

			Shader: passInfo.Shader,
		}
	})

	forwardPasses := gog.Map(info.ForwardPasses, func(passInfo ForwardRenderPassInfo) internal.MaterialRenderPassDefinition {
		if len(passInfo.Textures) > 8 {
			passInfo.Textures = passInfo.Textures[:8]
		}

		var textures [8]render.Texture
		for i, textureInfo := range passInfo.Textures {
			textures[i] = textureInfo.Texture
		}

		var samplers [8]render.Sampler
		for i, textureInfo := range passInfo.Textures {
			samplers[i] = e.api.CreateSampler(render.SamplerInfo{
				Wrapping:   textureInfo.Wrapping,
				Filtering:  textureInfo.Filtering,
				Mipmapping: textureInfo.Mipmapping,
			})
		}

		return internal.MaterialRenderPassDefinition{
			Layer:           passInfo.Layer,
			Culling:         passInfo.Culling.ValueOrDefault(render.CullModeNone),
			FrontFace:       passInfo.FrontFace.ValueOrDefault(render.FaceOrientationCCW),
			DepthTest:       passInfo.DepthTest.ValueOrDefault(true),
			DepthWrite:      passInfo.DepthWrite.ValueOrDefault(true),
			DepthComparison: passInfo.DepthComparison.ValueOrDefault(render.ComparisonLessOrEqual),
			Blending:        passInfo.AlphaBlending.ValueOrDefault(false),

			Textures:    textures,
			Samplers:    samplers,
			UniformData: passInfo.MaterialDataStd140,

			Shader: passInfo.Shader,
		}
	})

	return &Material{
		name:           info.Name,
		geometryPasses: geometryPasses,
		shadowPasses:   shadowPasses,
		forwardPasses:  forwardPasses,
	}
}

// CreatePBRMaterial creates a new Material that is based on PBR properties.
//
// Deprecated: Use CreateMaterial instead.
func (e *Engine) CreatePBRMaterial(info PBRMaterialInfo) *Material {
	var textures []TextureBindingInfo
	if info.AlbedoTexture != nil {
		textures = append(textures, TextureBindingInfo{
			Texture:    info.AlbedoTexture.texture,
			Wrapping:   render.WrapModeClamp,
			Filtering:  render.FilterModeLinear,
			Mipmapping: true,
		})
	}
	if info.NormalTexture != nil {
		textures = append(textures, TextureBindingInfo{
			Texture:    info.NormalTexture.texture,
			Wrapping:   render.WrapModeClamp,
			Filtering:  render.FilterModeNearest,
			Mipmapping: true,
		})
	}
	if info.MetallicRoughnessTexture != nil {
		textures = append(textures, TextureBindingInfo{
			Texture:    info.MetallicRoughnessTexture.texture,
			Wrapping:   render.WrapModeClamp,
			Filtering:  render.FilterModeLinear,
			Mipmapping: true,
		})
	}

	uniformData := make([]byte, 3*4*4)
	plotter := blob.NewPlotter(uniformData)
	plotter.PlotSPVec4(info.AlbedoColor)
	plotter.PlotFloat32(info.AlphaThreshold)
	plotter.PlotFloat32(info.NormalScale)
	plotter.PlotFloat32(info.Metallic)
	plotter.PlotFloat32(info.Roughness)
	plotter.PlotSPVec4(info.EmissiveColor)

	culling := render.CullModeNone
	if info.BackfaceCulling {
		culling = render.CullModeBack
	}

	return e.CreateMaterial(MaterialInfo{
		Name: "PBR-Unspecified",
		GeometryPasses: []GeometryRenderPassInfo{
			{
				Culling:            opt.V(culling),
				MaterialDataStd140: uniformData,
				Textures:           textures,
				Shader:             e.DefaultGeometryShader(info.AlphaTesting, info.AlbedoTexture != nil),
			},
		},
		ShadowPasses: []ShadowRenderPassInfo{
			{
				Culling:            opt.V(culling),
				MaterialDataStd140: uniformData,
				Shader:             e.DefaultShadowShader(),
			},
		},
	})

}

// CreateMeshGeometry creates a new MeshGeometry using the specified
// info object.
func (e *Engine) CreateMeshGeometry(info MeshGeometryInfo) *MeshGeometry {
	vertexBuffer := e.api.CreateVertexBuffer(render.BufferInfo{
		Dynamic: false,
		Data:    info.VertexData,
	})
	indexBuffer := e.api.CreateIndexBuffer(render.BufferInfo{
		Dynamic: false,
		Data:    info.IndexData,
	})

	var attributes []render.VertexArrayAttribute
	if info.VertexFormat.HasCoord {
		attributes = append(attributes, render.VertexArrayAttribute{
			Binding:  0,
			Location: internal.CoordAttributeIndex,
			Format:   render.VertexAttributeFormatRGB32F,
			Offset:   info.VertexFormat.CoordOffsetBytes,
		})
	}
	if info.VertexFormat.HasNormal {
		attributes = append(attributes, render.VertexArrayAttribute{
			Binding:  0,
			Location: internal.NormalAttributeIndex,
			Format:   render.VertexAttributeFormatRGB16F,
			Offset:   info.VertexFormat.NormalOffsetBytes,
		})
	}
	if info.VertexFormat.HasTangent {
		attributes = append(attributes, render.VertexArrayAttribute{
			Binding:  0,
			Location: internal.TangentAttributeIndex,
			Format:   render.VertexAttributeFormatRGB16F,
			Offset:   info.VertexFormat.TangentOffsetBytes,
		})
	}
	if info.VertexFormat.HasTexCoord {
		attributes = append(attributes, render.VertexArrayAttribute{
			Binding:  0,
			Location: internal.TexCoordAttributeIndex,
			Format:   render.VertexAttributeFormatRG16F,
			Offset:   info.VertexFormat.TexCoordOffsetBytes,
		})
	}
	if info.VertexFormat.HasColor {
		attributes = append(attributes, render.VertexArrayAttribute{
			Binding:  0,
			Location: internal.ColorAttributeIndex,
			Format:   render.VertexAttributeFormatRGBA8UN,
			Offset:   info.VertexFormat.ColorOffsetBytes,
		})
	}
	if info.VertexFormat.HasWeights {
		attributes = append(attributes, render.VertexArrayAttribute{
			Binding:  0,
			Location: internal.WeightsAttributeIndex,
			Format:   render.VertexAttributeFormatRGBA8UN,
			Offset:   info.VertexFormat.WeightsOffsetBytes,
		})
	}
	if info.VertexFormat.HasJoints {
		attributes = append(attributes, render.VertexArrayAttribute{
			Binding:  0,
			Location: internal.JointsAttributeIndex,
			Format:   render.VertexAttributeFormatRGBA8IU,
			Offset:   info.VertexFormat.JointsOffsetBytes,
		})
	}

	vertexArray := e.api.CreateVertexArray(render.VertexArrayInfo{
		Bindings: []render.VertexArrayBinding{
			{
				VertexBuffer: vertexBuffer,
				Stride:       info.VertexFormat.CoordStrideBytes, // FIXME: Not accurate
			},
		},
		Attributes:  attributes,
		IndexBuffer: indexBuffer,
		IndexFormat: e.convertIndexType(info.IndexFormat),
	})

	return &MeshGeometry{
		vertexBuffer: vertexBuffer,
		indexBuffer:  indexBuffer,
		vertexArray:  vertexArray,
		vertexFormat: info.VertexFormat,
		fragments: gog.Map(info.Fragments, func(fragmentInfo MeshGeometryFragmentInfo) MeshGeometryFragment {
			return MeshGeometryFragment{
				name:            fragmentInfo.Name,
				topology:        fragmentInfo.Topology,
				indexByteOffset: fragmentInfo.IndexByteOffset,
				indexCount:      fragmentInfo.IndexCount,
			}
		}),
		boundingSphereRadius: info.BoundingSphereRadius,
	}
}

// CreateMeshDefinition creates a new MeshDefinition using the specified
// info object.
func (e *Engine) CreateMeshDefinition(info MeshDefinitionInfo) *MeshDefinition {
	geometry := info.Geometry

	if len(info.Materials) != len(geometry.fragments) {
		log.Warn("Number of materials (%d) does not match number of fragments (%d)", len(info.Materials), len(geometry.fragments))
	}

	result := &MeshDefinition{
		engine:         e,
		geometry:       geometry,
		materials:      make([]*Material, len(geometry.fragments)),
		materialPasses: make([][internal.MeshRenderPassTypeCount][]internal.MeshRenderPass, len(geometry.fragments)),
	}
	for i := range min(len(info.Materials), len(geometry.fragments)) {
		result.SetMaterial(i, info.Materials[i])
	}
	return result
}

// CreateScene creates a new 3D Scene. Entities managed
// within a given scene are isolated within that scene.
func (e *Engine) CreateScene() *Scene {
	return newScene(e.renderer)
}

func (e *Engine) createGeometryPassProgram(programCode render.ProgramCode) render.Program {
	return e.api.CreateProgram(render.ProgramInfo{
		SourceCode: programCode,
		TextureBindings: []render.TextureBinding{
			render.NewTextureBinding("lackingTexture0", 0),
			render.NewTextureBinding("lackingTexture1", 1),
			render.NewTextureBinding("lackingTexture2", 2),
			render.NewTextureBinding("lackingTexture3", 3),
			render.NewTextureBinding("lackingTexture4", 4),
			render.NewTextureBinding("lackingTexture5", 5),
			render.NewTextureBinding("lackingTexture6", 6),
			render.NewTextureBinding("lackingTexture7", 7),
		},
		UniformBindings: []render.UniformBinding{
			render.NewUniformBinding("Camera", internal.UniformBufferBindingCamera),
			render.NewUniformBinding("Model", internal.UniformBufferBindingModel),
			render.NewUniformBinding("Material", internal.UniformBufferBindingMaterial),
			render.NewUniformBinding("Armature", internal.UniformBufferBindingArmature),
		},
	})
}

func (e *Engine) createGeometryPassPipeline(info internal.RenderPassPipelineInfo) render.Pipeline {
	return e.api.CreatePipeline(render.PipelineInfo{
		Program: info.Program,

		VertexArray: info.MeshVertexArray,
		Topology:    info.FragmentTopology,

		Culling:   info.PassDefinition.Culling,
		FrontFace: info.PassDefinition.FrontFace,

		DepthTest:       info.PassDefinition.DepthTest,
		DepthWrite:      info.PassDefinition.DepthWrite,
		DepthComparison: info.PassDefinition.DepthComparison,

		StencilTest:  false,                          // the GBuffer does not have a stencil component
		StencilFront: render.StencilOperationState{}, // irrelevant
		StencilBack:  render.StencilOperationState{}, // irrelevant

		ColorWrite: render.ColorMaskTrue,

		BlendEnabled:                false,                    // the GBuffer does not have an alpha component
		BlendColor:                  [4]float32{},             // irrelevant
		BlendSourceColorFactor:      render.BlendFactorZero,   // irrelevant
		BlendDestinationColorFactor: render.BlendFactorZero,   // irrelevant
		BlendSourceAlphaFactor:      render.BlendFactorZero,   // irrelevant
		BlendDestinationAlphaFactor: render.BlendFactorZero,   // irrelevant
		BlendOpColor:                render.BlendOperationAdd, // irrelevant
		BlendOpAlpha:                render.BlendOperationAdd, // irrelevant
	})
}

func (e *Engine) createShadowPassProgram(programCode render.ProgramCode) render.Program {
	return e.api.CreateProgram(render.ProgramInfo{
		SourceCode: programCode,
		TextureBindings: []render.TextureBinding{
			render.NewTextureBinding("lackingTexture0", 0),
			render.NewTextureBinding("lackingTexture1", 1),
			render.NewTextureBinding("lackingTexture2", 2),
			render.NewTextureBinding("lackingTexture3", 3),
			render.NewTextureBinding("lackingTexture4", 4),
			render.NewTextureBinding("lackingTexture5", 5),
			render.NewTextureBinding("lackingTexture6", 6),
			render.NewTextureBinding("lackingTexture7", 7),
		},
		UniformBindings: []render.UniformBinding{
			render.NewUniformBinding("Camera", internal.UniformBufferBindingCamera),
			render.NewUniformBinding("Model", internal.UniformBufferBindingModel),
			render.NewUniformBinding("Material", internal.UniformBufferBindingMaterial),
			render.NewUniformBinding("Armature", internal.UniformBufferBindingArmature),
		},
	})
}

func (e *Engine) createShadowPassPipeline(info internal.RenderPassPipelineInfo) render.Pipeline {
	return e.api.CreatePipeline(render.PipelineInfo{
		Program: info.Program,

		VertexArray: info.MeshVertexArray,
		Topology:    info.FragmentTopology,

		Culling:   info.PassDefinition.Culling,
		FrontFace: info.PassDefinition.FrontFace,

		DepthTest:       info.PassDefinition.DepthTest,
		DepthWrite:      info.PassDefinition.DepthWrite,
		DepthComparison: info.PassDefinition.DepthComparison,

		StencilTest:  false,                          // the only target is a depth buffer
		StencilFront: render.StencilOperationState{}, // irrelevant
		StencilBack:  render.StencilOperationState{}, // irrelevant

		ColorWrite: render.ColorMaskFalse, // the only target is a depth buffer

		BlendEnabled:                false,                    // the only target is a depth buffer
		BlendColor:                  [4]float32{},             // irrelevant
		BlendSourceColorFactor:      render.BlendFactorZero,   // irrelevant
		BlendDestinationColorFactor: render.BlendFactorZero,   // irrelevant
		BlendSourceAlphaFactor:      render.BlendFactorZero,   // irrelevant
		BlendDestinationAlphaFactor: render.BlendFactorZero,   // irrelevant
		BlendOpColor:                render.BlendOperationAdd, // irrelevant
		BlendOpAlpha:                render.BlendOperationAdd, // irrelevant
	})
}

func (e *Engine) createForwardPassProgram(programCode render.ProgramCode) render.Program {
	return e.api.CreateProgram(render.ProgramInfo{
		SourceCode: programCode,
		TextureBindings: []render.TextureBinding{
			render.NewTextureBinding("lackingTexture0", 0),
			render.NewTextureBinding("lackingTexture1", 1),
			render.NewTextureBinding("lackingTexture2", 2),
			render.NewTextureBinding("lackingTexture3", 3),
			render.NewTextureBinding("lackingTexture4", 4),
			render.NewTextureBinding("lackingTexture5", 5),
			render.NewTextureBinding("lackingTexture6", 6),
			render.NewTextureBinding("lackingTexture7", 7),
		},
		UniformBindings: []render.UniformBinding{
			render.NewUniformBinding("Camera", internal.UniformBufferBindingCamera),
			render.NewUniformBinding("Model", internal.UniformBufferBindingModel),
			render.NewUniformBinding("Material", internal.UniformBufferBindingMaterial),
			render.NewUniformBinding("Armature", internal.UniformBufferBindingArmature),
		},
	})
}

func (e *Engine) createForwardPassPipeline(info internal.RenderPassPipelineInfo) render.Pipeline {
	return e.api.CreatePipeline(render.PipelineInfo{
		Program: info.Program,

		VertexArray: info.MeshVertexArray,
		Topology:    info.FragmentTopology,

		Culling:   info.PassDefinition.Culling,   // default: render.CullModeNone
		FrontFace: info.PassDefinition.FrontFace, // default: render.FaceOrientationCCW

		DepthTest:       info.PassDefinition.DepthTest,       // default: true
		DepthWrite:      info.PassDefinition.DepthWrite,      // default: true
		DepthComparison: info.PassDefinition.DepthComparison, // default: render.ComparisonLessOrEqual

		StencilTest:  false,                          // the lighting buffer does not have a stencil component
		StencilFront: render.StencilOperationState{}, // irrelevant
		StencilBack:  render.StencilOperationState{}, // irrelevant

		ColorWrite: render.ColorMaskTrue,

		BlendEnabled:                info.PassDefinition.Blending, // default: false
		BlendColor:                  [4]float32{0.0, 0.0, 0.0, 0.0},
		BlendSourceColorFactor:      render.BlendFactorOne,
		BlendSourceAlphaFactor:      render.BlendFactorOne,
		BlendDestinationColorFactor: render.BlendFactorOne,
		BlendDestinationAlphaFactor: render.BlendFactorZero,
		BlendOpColor:                render.BlendOperationAdd,
		BlendOpAlpha:                render.BlendOperationAdd,
	})
}

func (e *Engine) createSkyProgram(programCode render.ProgramCode, shader *lsl.Shader) render.Program {
	var textureBindings []render.TextureBinding

	if textureBlock, ok := shader.FindTextureBlock(); ok {
		for i := range min(8, len(textureBlock.Fields)) {
			textureBindings = append(textureBindings, render.NewTextureBinding(textureBlock.Fields[i].Name, i))
		}
	}

	return e.api.CreateProgram(render.ProgramInfo{
		SourceCode:      programCode,
		TextureBindings: textureBindings,
		UniformBindings: []render.UniformBinding{
			render.NewUniformBinding("Camera", internal.UniformBufferBindingCamera),
			render.NewUniformBinding("Material", internal.UniformBufferBindingMaterial),
		},
	})
}

func (e *Engine) createSkyPipeline(info internal.SkyPipelineInfo) (render.Pipeline, uint32, uint32) {
	cubeShape := e.stageData.CubeShape()
	pipeline := e.api.CreatePipeline(render.PipelineInfo{
		Program:                     info.Program,
		VertexArray:                 cubeShape.VertexArray(),
		Topology:                    cubeShape.Topology(),
		Culling:                     render.CullModeBack,
		FrontFace:                   render.FaceOrientationCW, // we are inside the cube
		DepthTest:                   true,
		DepthWrite:                  false,
		DepthComparison:             render.ComparisonLessOrEqual,
		StencilTest:                 false,
		ColorWrite:                  render.ColorMaskTrue,
		BlendEnabled:                info.Blending,
		BlendSourceColorFactor:      render.BlendFactorSourceAlpha,
		BlendSourceAlphaFactor:      render.BlendFactorOne,
		BlendDestinationColorFactor: render.BlendFactorOneMinusSourceAlpha,
		BlendDestinationAlphaFactor: render.BlendFactorZero,
		BlendOpColor:                render.BlendOperationAdd,
		BlendOpAlpha:                render.BlendOperationAdd,
		BlendColor:                  [4]float32{0.0, 0.0, 0.0, 0.0},
	})
	return pipeline, 0, uint32(cubeShape.IndexCount())
}

func (e *Engine) pickFreeRenderPassKey() uint32 {
	e.freeRenderPassKey++
	return e.freeRenderPassKey
}

func (e *Engine) convertFormat(dataFormat DataFormat) render.DataFormat {
	switch dataFormat {
	case DataFormatRGBA8:
		return render.DataFormatRGBA8
	case DataFormatRGBA16F:
		return render.DataFormatRGBA16F
	case DataFormatRGBA32F:
		return render.DataFormatRGBA32F
	default:
		panic(fmt.Errorf("unknown data format: %d", dataFormat))
	}
}

func (e *Engine) convertIndexType(indexFormat IndexFormat) render.IndexFormat {
	switch indexFormat {
	case IndexFormatU16:
		return render.IndexFormatUnsignedShort
	case IndexFormatU32:
		return render.IndexFormatUnsignedInt
	default:
		panic(fmt.Errorf("unknown index format: %d", indexFormat))
	}
}
