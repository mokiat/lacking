package graphics

import (
	"github.com/mokiat/gog"
	"github.com/mokiat/lacking/debug/log"
	"github.com/mokiat/lacking/game/graphics/internal"
	"github.com/mokiat/lacking/game/graphics/lsl"
	"github.com/mokiat/lacking/render"
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

// CreateGeometryShader creates a new custom GeometryShader using the
// specified info object.
func (e *Engine) CreateGeometryShader(info ShaderInfo) *GeometryShader {
	ast, err := lsl.Parse(info.SourceCode)
	if err != nil {
		log.Error("Failed to parse geometry shader: %v", err)
		ast = &lsl.Shader{Declarations: []lsl.Declaration{}} // TODO: Something meaningful
	}
	// TODO: Validate against Geometry globals.
	return &GeometryShader{
		builder: e.shaderBuilder,
		ast:     ast,
	}
}

// CreateShadowShader creates a new custom ShadowShader using the
// specified info object.
func (e *Engine) CreateShadowShader(info ShaderInfo) *ShadowShader {
	ast, err := lsl.Parse(info.SourceCode)
	if err != nil {
		log.Error("Failed to parse shadow shader: %v", err)
		ast = &lsl.Shader{Declarations: []lsl.Declaration{}} // TODO: Something meaningful
	}
	// TODO: Validate against Shadow globals.
	return &ShadowShader{
		builder: e.shaderBuilder,
		ast:     ast,
	}
}

// CreateForwardShader creates a new custom ForwardShader using the
// specified info object.
func (e *Engine) CreateForwardShader(info ShaderInfo) *ForwardShader {
	ast, err := lsl.Parse(info.SourceCode)
	if err != nil {
		log.Error("Failed to parse forward shader: %v", err)
		ast = &lsl.Shader{Declarations: []lsl.Declaration{}} // TODO: Something meaningful
	}
	// TODO: Validate against Forward globals.
	return &ForwardShader{
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
	// TODO: Rework ast based on sky shader constraints.
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

// CreateMaterial creates a new Material from the specified info object.
func (e *Engine) CreateMaterial(info MaterialInfo) *Material {
	geometryPasses := gog.Map(info.GeometryPasses, func(passInfo GeometryRenderPassInfo) internal.MaterialRenderPassDefinition {
		return internal.MaterialRenderPassDefinition{
			Layer:           passInfo.Layer,
			Culling:         passInfo.Culling.ValueOrDefault(render.CullModeNone),
			FrontFace:       passInfo.FrontFace.ValueOrDefault(render.FaceOrientationCCW),
			DepthTest:       passInfo.DepthTest.ValueOrDefault(true),
			DepthWrite:      passInfo.DepthWrite.ValueOrDefault(true),
			DepthComparison: passInfo.DepthComparison.ValueOrDefault(render.ComparisonLessOrEqual),
			Blending:        false,
			TextureSet:      internal.NewShaderTextureSet(passInfo.Shader.ast),
			UniformSet:      internal.NewShaderUniformSet(passInfo.Shader.ast),
			Shader:          passInfo.Shader,
		}
	})

	shadowPasses := gog.Map(info.ShadowPasses, func(passInfo ShadowRenderPassInfo) internal.MaterialRenderPassDefinition {
		return internal.MaterialRenderPassDefinition{
			Layer:           0,
			Culling:         passInfo.Culling.ValueOrDefault(render.CullModeNone),
			FrontFace:       passInfo.FrontFace.ValueOrDefault(render.FaceOrientationCCW),
			DepthTest:       true,
			DepthWrite:      true,
			DepthComparison: render.ComparisonLessOrEqual,
			Blending:        false,
			TextureSet:      internal.NewShaderTextureSet(passInfo.Shader.ast),
			UniformSet:      internal.NewShaderUniformSet(passInfo.Shader.ast),
			Shader:          passInfo.Shader,
		}
	})

	forwardPasses := gog.Map(info.ForwardPasses, func(passInfo ForwardRenderPassInfo) internal.MaterialRenderPassDefinition {
		return internal.MaterialRenderPassDefinition{
			Layer:           passInfo.Layer,
			Culling:         passInfo.Culling.ValueOrDefault(render.CullModeNone),
			FrontFace:       passInfo.FrontFace.ValueOrDefault(render.FaceOrientationCCW),
			DepthTest:       passInfo.DepthTest.ValueOrDefault(true),
			DepthWrite:      passInfo.DepthWrite.ValueOrDefault(true),
			DepthComparison: passInfo.DepthComparison.ValueOrDefault(render.ComparisonLessOrEqual),
			Blending:        passInfo.Blending.ValueOrDefault(false),
			TextureSet:      internal.NewShaderTextureSet(passInfo.Shader.ast),
			UniformSet:      internal.NewShaderUniformSet(passInfo.Shader.ast),
			Shader:          passInfo.Shader,
		}
	})

	return &Material{
		name:           info.Name,
		geometryPasses: geometryPasses,
		shadowPasses:   shadowPasses,
		forwardPasses:  forwardPasses,
	}
}

// CreateMeshGeometry creates a new MeshGeometry using the specified
// info object.
func (e *Engine) CreateMeshGeometry(info MeshGeometryInfo) *MeshGeometry {
	vertexBuffers := make([]render.Buffer, len(info.VertexBuffers))
	for i, bufferInfo := range info.VertexBuffers {
		vertexBuffer := e.api.CreateVertexBuffer(render.BufferInfo{
			Dynamic: false,
			Data:    bufferInfo.Data,
		})
		vertexBuffers[i] = vertexBuffer
	}
	indexBuffer := e.api.CreateIndexBuffer(render.BufferInfo{
		Dynamic: false,
		Data:    info.IndexBuffer.Data,
	})

	var attributes []render.VertexArrayAttribute
	if attribInfo := info.VertexFormat.Coord; attribInfo.Specified {
		attributes = append(attributes, render.VertexArrayAttribute{
			Binding:  uint(attribInfo.Value.BufferIndex),
			Location: internal.CoordAttributeIndex,
			Format:   attribInfo.Value.Format,
			Offset:   attribInfo.Value.ByteOffset,
		})
	}
	if attribInfo := info.VertexFormat.Normal; attribInfo.Specified {
		attributes = append(attributes, render.VertexArrayAttribute{
			Binding:  uint(attribInfo.Value.BufferIndex),
			Location: internal.NormalAttributeIndex,
			Format:   attribInfo.Value.Format,
			Offset:   attribInfo.Value.ByteOffset,
		})
	}
	if attribInfo := info.VertexFormat.Tangent; attribInfo.Specified {
		attributes = append(attributes, render.VertexArrayAttribute{
			Binding:  uint(attribInfo.Value.BufferIndex),
			Location: internal.TangentAttributeIndex,
			Format:   attribInfo.Value.Format,
			Offset:   attribInfo.Value.ByteOffset,
		})
	}
	if attribInfo := info.VertexFormat.TexCoord; attribInfo.Specified {
		attributes = append(attributes, render.VertexArrayAttribute{
			Binding:  uint(attribInfo.Value.BufferIndex),
			Location: internal.TexCoordAttributeIndex,
			Format:   attribInfo.Value.Format,
			Offset:   attribInfo.Value.ByteOffset,
		})
	}
	if attribInfo := info.VertexFormat.Color; attribInfo.Specified {
		attributes = append(attributes, render.VertexArrayAttribute{
			Binding:  uint(attribInfo.Value.BufferIndex),
			Location: internal.ColorAttributeIndex,
			Format:   attribInfo.Value.Format,
			Offset:   attribInfo.Value.ByteOffset,
		})
	}
	if attribInfo := info.VertexFormat.Weights; attribInfo.Specified {
		attributes = append(attributes, render.VertexArrayAttribute{
			Binding:  uint(attribInfo.Value.BufferIndex),
			Location: internal.WeightsAttributeIndex,
			Format:   attribInfo.Value.Format,
			Offset:   attribInfo.Value.ByteOffset,
		})
	}
	if attribInfo := info.VertexFormat.Joints; attribInfo.Specified {
		attributes = append(attributes, render.VertexArrayAttribute{
			Binding:  uint(attribInfo.Value.BufferIndex),
			Location: internal.JointsAttributeIndex,
			Format:   attribInfo.Value.Format,
			Offset:   attribInfo.Value.ByteOffset,
		})
	}

	vertexArray := e.api.CreateVertexArray(render.VertexArrayInfo{
		Bindings: gog.MapIndex(info.VertexBuffers, func(index int, bufferInfo MeshGeometryVertexBuffer) render.VertexArrayBinding {
			return render.VertexArrayBinding{
				VertexBuffer: vertexBuffers[index],
				Stride:       bufferInfo.ByteStride,
			}
		}),
		Attributes:  attributes,
		IndexBuffer: indexBuffer,
		IndexFormat: info.IndexBuffer.Format,
	})

	return &MeshGeometry{
		vertexBuffers: vertexBuffers,
		indexBuffer:   indexBuffer,
		vertexArray:   vertexArray,
		vertexFormat:  info.VertexFormat,
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
	return newScene(e, e.renderer)
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
		for i := range uint(min(8, len(textureBlock.Fields))) {
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
