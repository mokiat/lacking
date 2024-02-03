package graphics

import (
	"fmt"
	"math"
	"time"

	"github.com/mokiat/gblob"
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/dtos"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/gomath/stod"
	"github.com/mokiat/lacking/debug/metric"
	"github.com/mokiat/lacking/game/graphics/internal"
	"github.com/mokiat/lacking/render"
	renderutil "github.com/mokiat/lacking/render/util"
	"github.com/mokiat/lacking/util/blob"
	"github.com/mokiat/lacking/util/spatial"
	"github.com/x448/float16"
	"golang.org/x/exp/slices"
)

const (
	shadowMapWidth  = 2048
	shadowMapHeight = 2048

	commandBufferSize = 2 * 1024 * 1024  // 2MB
	uniformBufferSize = 32 * 1024 * 1024 // 32MB
)

var ShowLightView bool

func newRenderer(api render.API, shaders ShaderCollection) *sceneRenderer {
	return &sceneRenderer{
		api:     api,
		shaders: shaders,

		exposureBufferData: make([]byte, 4*render.SizeF32), // Worst case RGBA32F
		exposureTarget:     1.0,

		visibleStaticMeshes: spatial.NewVisitorBucket[uint32](2_000),
		visibleMeshes:       spatial.NewVisitorBucket[*Mesh](2_000),

		ambientLightBucket: spatial.NewVisitorBucket[*AmbientLight](16),

		pointLightBucket: spatial.NewVisitorBucket[*PointLight](16),

		spotLightBucket: spatial.NewVisitorBucket[*SpotLight](16),

		directionalLightBucket: spatial.NewVisitorBucket[*DirectionalLight](16),
	}
}

type sceneRenderer struct {
	api           render.API
	shaders       ShaderCollection
	commandBuffer render.CommandBuffer

	quadShape   *internal.Shape
	cubeShape   *internal.Shape
	sphereShape *internal.Shape
	coneShape   *internal.Shape

	framebufferWidth  int
	framebufferHeight int

	geometryAlbedoTexture render.Texture
	geometryNormalTexture render.Texture
	geometryDepthTexture  render.Texture
	geometryFramebuffer   render.Framebuffer

	lightingAlbedoTexture render.Texture
	lightingFramebuffer   render.Framebuffer

	forwardFramebuffer render.Framebuffer

	shadowDepthTexture render.Texture
	shadowFramebuffer  render.Framebuffer

	exposureAlbedoTexture   render.Texture
	exposureFramebuffer     render.Framebuffer
	exposureFormat          render.DataFormat
	exposureProgram         render.Program
	exposurePipeline        render.Pipeline
	exposureBufferData      gblob.LittleEndianBlock
	exposureBuffer          render.Buffer
	exposureSync            render.Fence
	exposureTarget          float32
	exposureUpdateTimestamp time.Time

	postprocessingProgram  render.Program
	postprocessingPipeline render.Pipeline

	ambientLightProgram  render.Program
	ambientLightPipeline render.Pipeline
	ambientLightBucket   *spatial.VisitorBucket[*AmbientLight]

	pointLightProgram  render.Program
	pointLightPipeline render.Pipeline
	pointLightBucket   *spatial.VisitorBucket[*PointLight]

	spotLightProgram  render.Program
	spotLightPipeline render.Pipeline
	spotLightBucket   *spatial.VisitorBucket[*SpotLight]

	directionalLightProgram  render.Program
	directionalLightPipeline render.Pipeline
	directionalLightBucket   *spatial.VisitorBucket[*DirectionalLight]

	skyboxProgram    render.Program
	skyboxPipeline   render.Pipeline
	skycolorProgram  render.Program
	skycolorPipeline render.Pipeline

	debugLines        []debugLine
	debugVertexData   []byte
	debugVertexBuffer render.Buffer
	debugVertexArray  render.VertexArray
	debugProgram      render.Program
	debugPipeline     render.Pipeline

	visibleStaticMeshes *spatial.VisitorBucket[uint32]
	visibleMeshes       *spatial.VisitorBucket[*Mesh]
	renderItems         []renderItem

	uniforms                  *renderutil.UniformBlockBuffer
	modelUniformBufferData    gblob.LittleEndianBlock
	cameraPlacement           renderutil.UniformPlacement
	directionalLightPlacement renderutil.UniformPlacement
}

func (r *sceneRenderer) createFramebuffers(width, height int) {
	r.framebufferWidth = width
	r.framebufferHeight = height

	r.geometryAlbedoTexture = r.api.CreateColorTexture2D(render.ColorTexture2DInfo{
		Width:           r.framebufferWidth,
		Height:          r.framebufferHeight,
		Wrapping:        render.WrapModeClamp,
		Filtering:       render.FilterModeNearest,
		Mipmapping:      false,
		GammaCorrection: false,
		Format:          render.DataFormatRGBA8,
	})
	r.geometryNormalTexture = r.api.CreateColorTexture2D(render.ColorTexture2DInfo{
		Width:           r.framebufferWidth,
		Height:          r.framebufferHeight,
		Wrapping:        render.WrapModeClamp,
		Filtering:       render.FilterModeNearest,
		Mipmapping:      false,
		GammaCorrection: false,
		Format:          render.DataFormatRGBA16F,
	})
	r.geometryDepthTexture = r.api.CreateDepthTexture2D(render.DepthTexture2DInfo{
		Width:  r.framebufferWidth,
		Height: r.framebufferHeight,
	})
	r.geometryFramebuffer = r.api.CreateFramebuffer(render.FramebufferInfo{
		ColorAttachments: [4]render.Texture{
			r.geometryAlbedoTexture,
			r.geometryNormalTexture,
		},
		DepthAttachment: r.geometryDepthTexture,
	})

	r.lightingAlbedoTexture = r.api.CreateColorTexture2D(render.ColorTexture2DInfo{
		Width:           r.framebufferWidth,
		Height:          r.framebufferHeight,
		Wrapping:        render.WrapModeClamp,
		Filtering:       render.FilterModeNearest,
		Mipmapping:      false,
		GammaCorrection: false,
		Format:          render.DataFormatRGBA16F,
	})
	r.lightingFramebuffer = r.api.CreateFramebuffer(render.FramebufferInfo{
		ColorAttachments: [4]render.Texture{
			r.lightingAlbedoTexture,
		},
	})

	r.forwardFramebuffer = r.api.CreateFramebuffer(render.FramebufferInfo{
		ColorAttachments: [4]render.Texture{
			r.lightingAlbedoTexture,
		},
		DepthAttachment: r.geometryDepthTexture,
	})
}

func (r *sceneRenderer) releaseFramebuffers() {
	defer r.geometryAlbedoTexture.Release()
	defer r.geometryNormalTexture.Release()
	defer r.geometryDepthTexture.Release()
	defer r.geometryFramebuffer.Release()

	defer r.lightingAlbedoTexture.Release()
	defer r.lightingFramebuffer.Release()

	defer r.forwardFramebuffer.Release()
}

func (r *sceneRenderer) Allocate() {
	r.commandBuffer = r.api.CreateCommandBuffer(commandBufferSize)

	r.quadShape = internal.CreateQuadShape(r.api)
	r.cubeShape = internal.CreateCubeShape(r.api)
	r.sphereShape = internal.CreateSphereShape(r.api)
	r.coneShape = internal.CreateConeShape(r.api)

	r.createFramebuffers(800, 600)

	r.shadowDepthTexture = r.api.CreateDepthTexture2D(render.DepthTexture2DInfo{
		Width:        shadowMapWidth,
		Height:       shadowMapHeight,
		ClippedValue: opt.V(float32(1.0)),
		Comparable:   true,
	})
	r.shadowFramebuffer = r.api.CreateFramebuffer(render.FramebufferInfo{
		DepthAttachment: r.shadowDepthTexture,
	})

	r.exposureAlbedoTexture = r.api.CreateColorTexture2D(render.ColorTexture2DInfo{
		Width:           1,
		Height:          1,
		Wrapping:        render.WrapModeClamp,
		Filtering:       render.FilterModeNearest,
		Mipmapping:      false,
		GammaCorrection: false,
		Format:          render.DataFormatRGBA16F,
	})
	r.exposureFramebuffer = r.api.CreateFramebuffer(render.FramebufferInfo{
		ColorAttachments: [4]render.Texture{
			r.exposureAlbedoTexture,
		},
	})
	r.exposureFormat = r.api.DetermineContentFormat(r.exposureFramebuffer)
	r.exposureProgram = r.api.CreateProgram(render.ProgramInfo{
		SourceCode: r.shaders.ExposureSet(),
		TextureBindings: []render.TextureBinding{
			render.NewTextureBinding("fbColor0TextureIn", internal.TextureBindingLightingFramebufferColor0),
		},
		UniformBindings: []render.UniformBinding{},
	})
	r.exposurePipeline = r.api.CreatePipeline(render.PipelineInfo{
		Program:      r.exposureProgram,
		VertexArray:  r.quadShape.VertexArray(),
		Topology:     r.quadShape.Topology(),
		Culling:      render.CullModeBack,
		FrontFace:    render.FaceOrientationCCW,
		DepthTest:    false,
		DepthWrite:   false,
		StencilTest:  false,
		ColorWrite:   render.ColorMaskTrue,
		BlendEnabled: false,
	})
	r.exposureBuffer = r.api.CreatePixelTransferBuffer(render.BufferInfo{
		Dynamic: true,
		Size:    len(r.exposureBufferData),
	})

	r.postprocessingProgram = r.api.CreateProgram(render.ProgramInfo{
		SourceCode: r.shaders.PostprocessingSet(PostprocessingShaderConfig{
			ToneMapping: ExponentialToneMapping,
		}),
		TextureBindings: []render.TextureBinding{
			render.NewTextureBinding("fbColor0TextureIn", internal.TextureBindingPostprocessFramebufferColor0),
		},
		UniformBindings: []render.UniformBinding{
			render.NewUniformBinding("Postprocess", internal.UniformBufferBindingPostprocess),
		},
	})
	r.postprocessingPipeline = r.api.CreatePipeline(render.PipelineInfo{
		Program:         r.postprocessingProgram,
		VertexArray:     r.quadShape.VertexArray(),
		Topology:        r.quadShape.Topology(),
		Culling:         render.CullModeBack,
		FrontFace:       render.FaceOrientationCCW,
		DepthTest:       false,
		DepthWrite:      false,
		DepthComparison: render.ComparisonAlways,
		StencilTest:     false,
		ColorWrite:      [4]bool{true, true, true, true},
		BlendEnabled:    false,
	})

	r.ambientLightProgram = r.api.CreateProgram(render.ProgramInfo{
		SourceCode: r.shaders.AmbientLightSet(),
		TextureBindings: []render.TextureBinding{
			render.NewTextureBinding("fbColor0TextureIn", internal.TextureBindingLightingFramebufferColor0),
			render.NewTextureBinding("fbColor1TextureIn", internal.TextureBindingLightingFramebufferColor1),
			render.NewTextureBinding("fbDepthTextureIn", internal.TextureBindingLightingFramebufferDepth),
			render.NewTextureBinding("reflectionTextureIn", internal.TextureBindingLightingReflectionTexture),
			render.NewTextureBinding("refractionTextureIn", internal.TextureBindingLightingRefractionTexture),
		},
		UniformBindings: []render.UniformBinding{
			render.NewUniformBinding("Camera", internal.UniformBufferBindingCamera),
		},
	})
	r.ambientLightPipeline = r.api.CreatePipeline(render.PipelineInfo{
		Program:                     r.ambientLightProgram,
		VertexArray:                 r.quadShape.VertexArray(),
		Topology:                    r.quadShape.Topology(),
		Culling:                     render.CullModeBack,
		FrontFace:                   render.FaceOrientationCCW,
		DepthTest:                   false,
		DepthWrite:                  false,
		DepthComparison:             render.ComparisonAlways,
		StencilTest:                 false,
		ColorWrite:                  render.ColorMaskTrue,
		BlendEnabled:                true,
		BlendColor:                  [4]float32{0.0, 0.0, 0.0, 0.0},
		BlendSourceColorFactor:      render.BlendFactorOne,
		BlendSourceAlphaFactor:      render.BlendFactorOne,
		BlendDestinationColorFactor: render.BlendFactorOne,
		BlendDestinationAlphaFactor: render.BlendFactorZero,
		BlendOpColor:                render.BlendOperationAdd,
		BlendOpAlpha:                render.BlendOperationAdd,
	})

	r.pointLightProgram = r.api.CreateProgram(render.ProgramInfo{
		SourceCode: r.shaders.PointLightSet(),
		TextureBindings: []render.TextureBinding{
			render.NewTextureBinding("fbColor0TextureIn", internal.TextureBindingLightingFramebufferColor0),
			render.NewTextureBinding("fbColor1TextureIn", internal.TextureBindingLightingFramebufferColor1),
			render.NewTextureBinding("fbDepthTextureIn", internal.TextureBindingLightingFramebufferDepth),
		},
		UniformBindings: []render.UniformBinding{
			render.NewUniformBinding("Camera", internal.UniformBufferBindingCamera),
			render.NewUniformBinding("Light", internal.UniformBufferBindingLight),
			render.NewUniformBinding("LightProperties", internal.UniformBufferBindingLightProperties),
		},
	})
	r.pointLightPipeline = r.api.CreatePipeline(render.PipelineInfo{
		Program:                     r.pointLightProgram,
		VertexArray:                 r.sphereShape.VertexArray(),
		Topology:                    r.sphereShape.Topology(),
		Culling:                     render.CullModeFront,
		FrontFace:                   render.FaceOrientationCCW,
		DepthTest:                   false,
		DepthWrite:                  false,
		DepthComparison:             render.ComparisonAlways,
		StencilTest:                 false,
		ColorWrite:                  render.ColorMaskTrue,
		BlendEnabled:                true,
		BlendColor:                  [4]float32{0.0, 0.0, 0.0, 0.0},
		BlendSourceColorFactor:      render.BlendFactorOne,
		BlendSourceAlphaFactor:      render.BlendFactorOne,
		BlendDestinationColorFactor: render.BlendFactorOne,
		BlendDestinationAlphaFactor: render.BlendFactorZero,
		BlendOpColor:                render.BlendOperationAdd,
		BlendOpAlpha:                render.BlendOperationAdd,
	})

	r.spotLightProgram = r.api.CreateProgram(render.ProgramInfo{
		SourceCode: r.shaders.SpotLightSet(),
		TextureBindings: []render.TextureBinding{
			render.NewTextureBinding("fbColor0TextureIn", internal.TextureBindingLightingFramebufferColor0),
			render.NewTextureBinding("fbColor1TextureIn", internal.TextureBindingLightingFramebufferColor1),
			render.NewTextureBinding("fbDepthTextureIn", internal.TextureBindingLightingFramebufferDepth),
		},
		UniformBindings: []render.UniformBinding{
			render.NewUniformBinding("Camera", internal.UniformBufferBindingCamera),
			render.NewUniformBinding("Light", internal.UniformBufferBindingLight),
			render.NewUniformBinding("LightProperties", internal.UniformBufferBindingLightProperties),
		},
	})
	r.spotLightPipeline = r.api.CreatePipeline(render.PipelineInfo{
		Program:                     r.spotLightProgram,
		VertexArray:                 r.coneShape.VertexArray(),
		Topology:                    r.coneShape.Topology(),
		Culling:                     render.CullModeFront,
		FrontFace:                   render.FaceOrientationCCW,
		DepthTest:                   false,
		DepthWrite:                  false,
		DepthComparison:             render.ComparisonAlways,
		StencilTest:                 false,
		ColorWrite:                  render.ColorMaskTrue,
		BlendEnabled:                true,
		BlendColor:                  [4]float32{0.0, 0.0, 0.0, 0.0},
		BlendSourceColorFactor:      render.BlendFactorOne,
		BlendSourceAlphaFactor:      render.BlendFactorOne,
		BlendDestinationColorFactor: render.BlendFactorOne,
		BlendDestinationAlphaFactor: render.BlendFactorZero,
		BlendOpColor:                render.BlendOperationAdd,
		BlendOpAlpha:                render.BlendOperationAdd,
	})

	r.directionalLightProgram = r.api.CreateProgram(render.ProgramInfo{
		SourceCode: r.shaders.DirectionalLightSet(),
		TextureBindings: []render.TextureBinding{
			render.NewTextureBinding("fbColor0TextureIn", internal.TextureBindingLightingFramebufferColor0),
			render.NewTextureBinding("fbColor1TextureIn", internal.TextureBindingLightingFramebufferColor1),
			render.NewTextureBinding("fbDepthTextureIn", internal.TextureBindingLightingFramebufferDepth),
			render.NewTextureBinding("fbShadowTextureIn", internal.TextureBindingShadowFramebufferDepth),
		},
		UniformBindings: []render.UniformBinding{
			render.NewUniformBinding("Camera", internal.UniformBufferBindingCamera),
			render.NewUniformBinding("Light", internal.UniformBufferBindingLight),
			render.NewUniformBinding("LightProperties", internal.UniformBufferBindingLightProperties),
		},
	})
	r.directionalLightPipeline = r.api.CreatePipeline(render.PipelineInfo{
		Program:                     r.directionalLightProgram,
		VertexArray:                 r.quadShape.VertexArray(),
		Topology:                    r.quadShape.Topology(),
		Culling:                     render.CullModeBack,
		FrontFace:                   render.FaceOrientationCCW,
		DepthTest:                   false,
		DepthWrite:                  false,
		DepthComparison:             render.ComparisonAlways,
		StencilTest:                 false,
		ColorWrite:                  render.ColorMaskTrue,
		BlendEnabled:                true,
		BlendColor:                  [4]float32{0.0, 0.0, 0.0, 0.0},
		BlendSourceColorFactor:      render.BlendFactorOne,
		BlendSourceAlphaFactor:      render.BlendFactorOne,
		BlendDestinationColorFactor: render.BlendFactorOne,
		BlendDestinationAlphaFactor: render.BlendFactorZero,
		BlendOpColor:                render.BlendOperationAdd,
		BlendOpAlpha:                render.BlendOperationAdd,
	})

	r.skyboxProgram = r.api.CreateProgram(render.ProgramInfo{
		SourceCode: r.shaders.SkyboxSet(),
		TextureBindings: []render.TextureBinding{
			render.NewTextureBinding("albedoCubeTextureIn", internal.TextureBindingSkyboxAlbedoTexture),
		},
		UniformBindings: []render.UniformBinding{
			render.NewUniformBinding("Camera", internal.UniformBufferBindingCamera),
		},
	})
	r.skyboxPipeline = r.api.CreatePipeline(render.PipelineInfo{
		Program:     r.skyboxProgram,
		VertexArray: r.cubeShape.VertexArray(),
		Topology:    r.cubeShape.Topology(),
		Culling:     render.CullModeBack,
		// We are looking from within the cube shape so we need to flip the winding.
		FrontFace:       render.FaceOrientationCW,
		DepthTest:       true,
		DepthWrite:      false,
		DepthComparison: render.ComparisonLessOrEqual,
		StencilTest:     false,
		ColorWrite:      render.ColorMaskTrue,
		BlendEnabled:    false,
	})

	r.skycolorProgram = r.api.CreateProgram(render.ProgramInfo{
		SourceCode:      r.shaders.SkycolorSet(),
		TextureBindings: []render.TextureBinding{},
		UniformBindings: []render.UniformBinding{
			render.NewUniformBinding("Camera", internal.UniformBufferBindingCamera),
			render.NewUniformBinding("Skybox", internal.UniformBufferBindingSkybox),
		},
	})
	r.skycolorPipeline = r.api.CreatePipeline(render.PipelineInfo{
		Program:     r.skycolorProgram,
		VertexArray: r.cubeShape.VertexArray(),
		Topology:    r.cubeShape.Topology(),
		Culling:     render.CullModeBack,
		// We are looking from within the cube shape so we need to flip the winding.
		FrontFace:       render.FaceOrientationCW,
		DepthTest:       true,
		DepthWrite:      false,
		DepthComparison: render.ComparisonLessOrEqual,
		StencilTest:     false,
		ColorWrite:      render.ColorMaskTrue,
		BlendEnabled:    false,
	})

	r.debugLines = make([]debugLine, 131072)
	r.debugVertexData = make([]byte, len(r.debugLines)*4*4*2)
	r.debugVertexBuffer = r.api.CreateVertexBuffer(render.BufferInfo{
		Dynamic: true,
		Data:    r.debugVertexData,
	})
	r.debugVertexArray = r.api.CreateVertexArray(render.VertexArrayInfo{
		Bindings: []render.VertexArrayBinding{
			render.NewVertexArrayBinding(r.debugVertexBuffer, 4*4),
		},
		Attributes: []render.VertexArrayAttribute{
			render.NewVertexArrayAttribute(0, internal.CoordAttributeIndex, 0, render.VertexAttributeFormatRGB32F),
			render.NewVertexArrayAttribute(0, internal.ColorAttributeIndex, 3*4, render.VertexAttributeFormatRGB8UN),
		},
	})
	r.debugProgram = r.api.CreateProgram(render.ProgramInfo{
		SourceCode:      r.shaders.DebugSet(),
		TextureBindings: []render.TextureBinding{},
		UniformBindings: []render.UniformBinding{
			render.NewUniformBinding("Camera", internal.UniformBufferBindingCamera),
		},
	})
	r.debugPipeline = r.api.CreatePipeline(render.PipelineInfo{
		Program:         r.debugProgram,
		VertexArray:     r.debugVertexArray,
		Topology:        render.TopologyLineList,
		Culling:         render.CullModeNone,
		FrontFace:       render.FaceOrientationCCW,
		DepthTest:       true,
		DepthWrite:      false,
		DepthComparison: render.ComparisonLessOrEqual,
		StencilTest:     false,
		ColorWrite:      render.ColorMaskTrue,
		BlendEnabled:    false,
	})

	r.uniforms = renderutil.NewUniformBlockBuffer(r.api, uniformBufferSize)
	r.modelUniformBufferData = make([]byte, 64*256) // 256 x mat4
}

func (r *sceneRenderer) Release() {
	defer r.quadShape.Release()
	defer r.cubeShape.Release()
	defer r.sphereShape.Release()
	defer r.coneShape.Release()

	defer r.releaseFramebuffers()

	defer r.shadowDepthTexture.Release()
	defer r.shadowFramebuffer.Release()

	defer r.exposureAlbedoTexture.Release()
	defer r.exposureFramebuffer.Release()
	defer r.exposureProgram.Release()
	defer r.exposurePipeline.Release()
	defer r.exposureBuffer.Release()

	defer r.postprocessingProgram.Release()
	defer r.postprocessingPipeline.Release()

	defer r.ambientLightProgram.Release()
	defer r.ambientLightPipeline.Release()

	defer r.pointLightProgram.Release()
	defer r.pointLightPipeline.Release()

	defer r.spotLightProgram.Release()
	defer r.spotLightPipeline.Release()

	defer r.directionalLightProgram.Release()
	defer r.directionalLightPipeline.Release()

	defer r.skyboxProgram.Release()
	defer r.skyboxPipeline.Release()

	defer r.skycolorProgram.Release()
	defer r.skycolorPipeline.Release()

	defer r.debugVertexBuffer.Release()
	defer r.debugVertexArray.Release()
	defer r.debugProgram.Release()
	defer r.debugPipeline.Release()

	defer r.uniforms.Release()
}

func (r *sceneRenderer) ResetDebugLines() {
	r.debugLines = r.debugLines[:0]
}

func (r *sceneRenderer) QueueDebugLine(line debugLine) {
	if len(r.debugLines) == cap(r.debugLines)-1 {
		logger.Warn("Debug lines limit reached!")
	}
	if len(r.debugLines) == cap(r.debugLines) {
		return
	}
	r.debugLines = append(r.debugLines, line)
}

func (r *sceneRenderer) Ray(viewport Viewport, camera *Camera, x, y int) (dprec.Vec3, dprec.Vec3) {
	projectionMatrix := stod.Mat4(r.evaluateProjectionMatrix(camera, viewport.Width, viewport.Height))
	inverseProjection := dprec.InverseMat4(projectionMatrix)

	cameraMatrix := stod.Mat4(camera.gfxMatrix())

	pX := (float64(x-viewport.X)/float64(viewport.Width))*2.0 - 1.0
	pY := (float64(viewport.Y-y)/float64(viewport.Height))*2.0 + 1.0

	a := dprec.Mat4Vec4Prod(inverseProjection, dprec.NewVec4(
		pX, pY, -1.0, 1.0,
	))
	b := dprec.Mat4Vec4Prod(inverseProjection, dprec.NewVec4(
		pX, pY, 1.0, 1.0,
	))
	a = dprec.Vec4Quot(a, a.W)
	b = dprec.Vec4Quot(b, b.W)

	a = dprec.Mat4Vec4Prod(cameraMatrix, a)
	b = dprec.Mat4Vec4Prod(cameraMatrix, b)

	return a.VecXYZ(), b.VecXYZ()
}

func (r *sceneRenderer) Point(viewport Viewport, camera *Camera, position dprec.Vec3) dprec.Vec2 {
	pos := dprec.NewVec4(position.X, position.Y, position.Z, 1.0)
	projectionMatrix := stod.Mat4(r.evaluateProjectionMatrix(camera, viewport.Width, viewport.Height))
	viewMatrix := stod.Mat4(sprec.InverseMat4(camera.gfxMatrix()))
	ndc := dprec.Mat4Vec4Prod(projectionMatrix, dprec.Mat4Vec4Prod(viewMatrix, pos))
	if dprec.Abs(ndc.W) < 0.0001 {
		return dprec.ZeroVec2()
	}
	clip := dprec.Vec4Quot(ndc, ndc.W)
	return dprec.NewVec2((clip.X+1.0)*float64(viewport.Width)/2.0, (1.0-clip.Y)*float64(viewport.Height)/2.0)
}

func (r *sceneRenderer) Render(framebuffer render.Framebuffer, viewport Viewport, scene *Scene, camera *Camera) {
	if viewport.Width != r.framebufferWidth || viewport.Height != r.framebufferHeight {
		r.releaseFramebuffers()
		r.createFramebuffers(viewport.Width, viewport.Height)
	}

	r.uniforms.Reset()

	projectionMatrix := r.evaluateProjectionMatrix(camera, viewport.Width, viewport.Height)
	cameraMatrix := camera.gfxMatrix()
	viewMatrix := sprec.InverseMat4(cameraMatrix)
	projectionViewMatrix := sprec.Mat4Prod(projectionMatrix, viewMatrix)
	frustum := spatial.ProjectionRegion(stod.Mat4(projectionViewMatrix))

	if ShowLightView {
		r.directionalLightBucket.Reset()
		scene.directionalLightSet.VisitHexahedronRegion(&frustum, r.directionalLightBucket)

		var directionalLight *DirectionalLight
		for _, light := range r.directionalLightBucket.Items() {
			if light.active {
				directionalLight = light
				break
			}
		}
		if directionalLight != nil {
			projectionMatrix = lightOrtho()
			cameraMatrix = directionalLight.gfxMatrix()
			viewMatrix = sprec.InverseMat4(cameraMatrix)
			projectionViewMatrix = sprec.Mat4Prod(projectionMatrix, viewMatrix)
			frustum = spatial.ProjectionRegion(stod.Mat4(projectionViewMatrix))
		}
	}

	ctx := renderCtx{
		framebuffer: framebuffer,
		scene:       scene,
		x:           viewport.X,
		y:           viewport.Y,
		width:       viewport.Width,
		height:      viewport.Height,
		camera:      camera,
		frustum:     frustum,
	}

	r.cameraPlacement = renderutil.WriteUniform(r.uniforms, internal.CameraUniform{
		ProjectionMatrix: projectionMatrix,
		ViewMatrix:       viewMatrix,
		CameraMatrix:     cameraMatrix,
		Viewport: sprec.NewVec4(
			float32(viewport.X),
			float32(viewport.Y),
			float32(viewport.Width),
			float32(viewport.Height),
		),
	})

	r.renderShadowPass(ctx)
	r.renderGeometryPass(ctx)
	r.renderLightingPass(ctx)
	r.renderForwardPass(ctx)
	if camera.autoExposureEnabled {
		r.renderExposureProbePass(ctx)
	}
	r.renderPostprocessingPass(ctx)

	uniformSpan := metric.BeginRegion("upload")
	r.uniforms.Upload()
	uniformSpan.End()

	submitSpan := metric.BeginRegion("submit")
	r.api.Queue().Invalidate()
	r.api.Queue().Submit(r.commandBuffer)
	submitSpan.End()
}

func (r *sceneRenderer) evaluateProjectionMatrix(camera *Camera, width, height int) sprec.Mat4 {
	var (
		near    = camera.Near()
		far     = camera.Far()
		fWidth  = sprec.Max(1.0, float32(width))
		fHeight = sprec.Max(1.0, float32(height))
	)

	switch camera.fovMode {
	case FoVModeHorizontalPlus:
		halfHeight := near * sprec.Tan(camera.fov/2.0)
		halfWidth := halfHeight * (fWidth / fHeight)
		return sprec.PerspectiveMat4(
			-halfWidth, halfWidth, -halfHeight, halfHeight, near, far,
		)

	case FoVModeVertialMinus:
		halfWidth := near * sprec.Tan(camera.fov/2.0)
		halfHeight := halfWidth * (fHeight / fWidth)
		return sprec.PerspectiveMat4(
			-halfWidth, halfWidth, -halfHeight, halfHeight, near, far,
		)

	case FoVModePixelBased:
		halfWidth := fWidth / 2.0
		halfHeight := fHeight / 2.0
		return sprec.OrthoMat4(
			-halfWidth, halfWidth, halfHeight, -halfHeight, near, far,
		)

	default:
		panic(fmt.Errorf("unsupported fov mode: %s", camera.fovMode))
	}
}

func lightOrtho() sprec.Mat4 {
	return sprec.OrthoMat4(-32, 32, 32, -32, 0, 256)
}

func (r *sceneRenderer) renderShadowPass(ctx renderCtx) {
	defer metric.BeginRegion("shadow").End()

	r.directionalLightBucket.Reset()
	ctx.scene.directionalLightSet.VisitHexahedronRegion(&ctx.frustum, r.directionalLightBucket)

	var directionalLight *DirectionalLight
	for _, light := range r.directionalLightBucket.Items() {
		if light.active {
			directionalLight = light
			break
		}
	}
	if directionalLight == nil {
		return
	}

	projectionMatrix := lightOrtho()
	lightMatrix := directionalLight.gfxMatrix()
	lightMatrix.M14 = floor(lightMatrix.M14*shadowMapWidth) / float32(shadowMapWidth)
	lightMatrix.M24 = floor(lightMatrix.M24*shadowMapWidth) / float32(shadowMapWidth)
	lightMatrix.M34 = floor(lightMatrix.M34*shadowMapWidth) / float32(shadowMapWidth)
	viewMatrix := sprec.InverseMat4(lightMatrix)
	projectionViewMatrix := sprec.Mat4Prod(projectionMatrix, viewMatrix)
	frustum := spatial.ProjectionRegion(stod.Mat4(projectionViewMatrix))

	r.renderItems = r.renderItems[:0]

	r.visibleMeshes.Reset()
	ctx.scene.dynamicMeshSet.VisitHexahedronRegion(&frustum, r.visibleMeshes)
	for _, mesh := range r.visibleMeshes.Items() {
		r.queueShadowMesh(ctx, mesh)
	}

	r.visibleStaticMeshes.Reset()
	ctx.scene.staticMeshOctree.VisitHexahedronRegion(&frustum, r.visibleStaticMeshes)
	for _, meshIndex := range r.visibleStaticMeshes.Items() {
		r.queueStaticShadowMesh(ctx, &ctx.scene.staticMeshes[meshIndex])
	}

	slices.SortFunc(r.renderItems, func(a, b renderItem) int {
		// TODO: If fragment IDs are stored inside the items themselves, there
		// would be fewer pointer jumps and cache misses.
		return a.fragment.id - b.fragment.id
	})

	r.commandBuffer.BeginRenderPass(render.RenderPassInfo{
		Framebuffer: r.shadowFramebuffer,
		Viewport: render.Area{
			X:      0,
			Y:      0,
			Width:  shadowMapWidth,
			Height: shadowMapHeight,
		},
		DepthLoadOp:     render.LoadOperationClear,
		DepthStoreOp:    render.StoreOperationStore,
		DepthClearValue: 1.0,
		StencilLoadOp:   render.LoadOperationDontCare,
		StencilStoreOp:  render.StoreOperationDontCare,
	})

	r.directionalLightPlacement = renderutil.WriteUniform(r.uniforms, internal.LightUniform{
		ProjectionMatrix: projectionMatrix,
		ViewMatrix:       viewMatrix,
		LightMatrix:      lightMatrix,
	})
	r.renderShadowMeshesList(ctx)

	r.commandBuffer.EndRenderPass()
}

func (r *sceneRenderer) queueShadowMesh(ctx renderCtx, mesh *Mesh) {
	modelMatrix := mesh.matrixData
	definition := mesh.definition
	for i := range definition.fragments {
		// TODO: Fragment is a somewhat shallow structure. Consider copying the
		// relevant information instead of dealing with pointers.
		fragment := &definition.fragments[i]
		r.renderItems = append(r.renderItems, renderItem{
			fragment:    fragment,
			modelMatrix: modelMatrix,
			armature:    mesh.armature,
		})
	}
}

func (r *sceneRenderer) queueStaticShadowMesh(ctx renderCtx, mesh *StaticMesh) {
	modelMatrix := mesh.matrixData
	definition := mesh.definition
	for i := range definition.fragments {
		// TODO: Fragment is a somewhat shallow structure. Consider copying the
		// relevant information instead of dealing with pointers.
		fragment := &definition.fragments[i]
		r.renderItems = append(r.renderItems, renderItem{
			fragment:    fragment,
			modelMatrix: modelMatrix,
		})
	}
}

func (r *sceneRenderer) renderShadowMeshesList(ctx renderCtx) {
	var lastFragment *meshFragmentDefinition
	count := 0
	for _, item := range r.renderItems {
		fragment := item.fragment
		material := fragment.material

		// just append the modelmatrix
		if fragment == lastFragment && item.armature == nil {
			copy(r.modelUniformBufferData[count*64:], item.modelMatrix)
			count++

			// flush batch
			if count >= 256 {
				modelPlacement := renderutil.WriteUniform(r.uniforms, internal.ModelUniform{
					ModelMatrices: r.modelUniformBufferData,
				})
				r.commandBuffer.UniformBufferUnit(
					internal.UniformBufferBindingModel,
					modelPlacement.Buffer,
					modelPlacement.Offset,
					modelPlacement.Size,
				)
				r.commandBuffer.DrawIndexed(lastFragment.indexOffsetBytes, lastFragment.indexCount, count)
				count = 0
			}
		} else {
			// flush batch
			if lastFragment != nil && count > 0 {
				modelPlacement := renderutil.WriteUniform(r.uniforms, internal.ModelUniform{
					ModelMatrices: r.modelUniformBufferData,
				})
				r.commandBuffer.UniformBufferUnit(
					internal.UniformBufferBindingModel,
					modelPlacement.Buffer,
					modelPlacement.Offset,
					modelPlacement.Size,
				)
				r.commandBuffer.DrawIndexed(lastFragment.indexOffsetBytes, lastFragment.indexCount, count)
			}

			count = 0
			lastFragment = fragment
			r.commandBuffer.BindPipeline(material.shadowPipeline)
			r.commandBuffer.UniformBufferUnit(
				internal.UniformBufferBindingLight,
				r.directionalLightPlacement.Buffer,
				r.directionalLightPlacement.Offset,
				r.directionalLightPlacement.Size,
			)

			if item.armature == nil {
				copy(r.modelUniformBufferData[count*64:], item.modelMatrix)
				count++
			} else {
				// No batching for submeshes with armature. Zero index is the model
				// matrix
				copy(r.modelUniformBufferData, item.armature.uniformBufferData)
				modelPlacement := renderutil.WriteUniform(r.uniforms, internal.ModelUniform{
					ModelMatrices: r.modelUniformBufferData,
				})
				r.commandBuffer.UniformBufferUnit(
					internal.UniformBufferBindingModel,
					modelPlacement.Buffer,
					modelPlacement.Offset,
					modelPlacement.Size,
				)
				r.commandBuffer.DrawIndexed(lastFragment.indexOffsetBytes, lastFragment.indexCount, 1)
			}
		}
	}

	// flush remainder
	if lastFragment != nil && count > 0 {
		modelPlacement := renderutil.WriteUniform(r.uniforms, internal.ModelUniform{
			ModelMatrices: r.modelUniformBufferData,
		})
		r.commandBuffer.UniformBufferUnit(
			internal.UniformBufferBindingModel,
			modelPlacement.Buffer,
			modelPlacement.Offset,
			modelPlacement.Size,
		)
		r.commandBuffer.DrawIndexed(lastFragment.indexOffsetBytes, lastFragment.indexCount, count)
	}
}

func (r *sceneRenderer) renderGeometryPass(ctx renderCtx) {
	defer metric.BeginRegion("geometry").End()

	r.commandBuffer.BeginRenderPass(render.RenderPassInfo{
		Framebuffer: r.geometryFramebuffer,
		Viewport: render.Area{
			X:      0,
			Y:      0,
			Width:  r.framebufferWidth,
			Height: r.framebufferHeight,
		},
		DepthLoadOp:     render.LoadOperationClear,
		DepthStoreOp:    render.StoreOperationStore,
		DepthClearValue: 1.0,
		StencilLoadOp:   render.LoadOperationDontCare,
		StencilStoreOp:  render.StoreOperationDontCare,
		Colors: [4]render.ColorAttachmentInfo{
			{
				LoadOp:  render.LoadOperationClear,
				StoreOp: render.StoreOperationStore,
				ClearValue: [4]float32{
					ctx.scene.sky.backgroundColor.X,
					ctx.scene.sky.backgroundColor.Y,
					ctx.scene.sky.backgroundColor.Z,
					1.0,
				},
			},
			{
				LoadOp:     render.LoadOperationClear,
				StoreOp:    render.StoreOperationStore,
				ClearValue: [4]float32{0.0, 0.0, 1.0, 0.0},
			},
		},
	})

	r.renderItems = r.renderItems[:0]

	r.visibleMeshes.Reset()
	ctx.scene.dynamicMeshSet.VisitHexahedronRegion(&ctx.frustum, r.visibleMeshes)
	for _, mesh := range r.visibleMeshes.Items() {
		r.queueMesh(ctx, mesh)
	}

	r.visibleStaticMeshes.Reset()
	ctx.scene.staticMeshOctree.VisitHexahedronRegion(&ctx.frustum, r.visibleStaticMeshes)
	for _, meshIndex := range r.visibleStaticMeshes.Items() {
		r.queueStaticMesh(ctx, &ctx.scene.staticMeshes[meshIndex])
	}

	slices.SortFunc(r.renderItems, func(a, b renderItem) int {
		return a.fragment.id - b.fragment.id
	})
	r.renderMeshesList(ctx)

	r.commandBuffer.EndRenderPass()
}

func (r *sceneRenderer) queueMesh(ctx renderCtx, mesh *Mesh) {
	modelMatrix := mesh.matrixData
	definition := mesh.definition
	for i := range definition.fragments {
		fragment := &definition.fragments[i]
		r.renderItems = append(r.renderItems, renderItem{
			fragment:    fragment,
			modelMatrix: modelMatrix,
			armature:    mesh.armature,
		})
	}
}

func (r *sceneRenderer) queueStaticMesh(ctx renderCtx, mesh *StaticMesh) {
	modelMatrix := mesh.matrixData
	definition := mesh.definition
	for i := range definition.fragments {
		fragment := &definition.fragments[i]
		r.renderItems = append(r.renderItems, renderItem{
			fragment:    fragment,
			modelMatrix: modelMatrix[:],
			armature:    nil,
		})
	}
}

func (r *sceneRenderer) renderMeshesList(ctx renderCtx) {
	var lastFragment *meshFragmentDefinition
	count := 0
	for _, item := range r.renderItems {
		fragment := item.fragment
		if fragment.material.geometryPipeline == nil {
			continue // skip non-geometry meshes
		}

		// just append the modelmatrix
		if fragment == lastFragment && item.armature == nil {
			copy(r.modelUniformBufferData[count*64:], item.modelMatrix)
			count++

			// flush batch
			if count >= 256 {
				modelPlacement := renderutil.WriteUniform(r.uniforms, internal.ModelUniform{
					ModelMatrices: r.modelUniformBufferData,
				})
				r.commandBuffer.UniformBufferUnit(
					internal.UniformBufferBindingModel,
					modelPlacement.Buffer,
					modelPlacement.Offset,
					modelPlacement.Size,
				)
				r.commandBuffer.DrawIndexed(lastFragment.indexOffsetBytes, lastFragment.indexCount, count)
				count = 0
			}
		} else {
			// flush batch
			if lastFragment != nil && count > 0 {
				modelPlacement := renderutil.WriteUniform(r.uniforms, internal.ModelUniform{
					ModelMatrices: r.modelUniformBufferData,
				})
				r.commandBuffer.UniformBufferUnit(
					internal.UniformBufferBindingModel,
					modelPlacement.Buffer,
					modelPlacement.Offset,
					modelPlacement.Size,
				)
				r.commandBuffer.DrawIndexed(lastFragment.indexOffsetBytes, lastFragment.indexCount, count)
			}

			count = 0
			lastFragment = fragment
			material := fragment.material
			materialValues := material.definition
			r.commandBuffer.BindPipeline(material.geometryPipeline)
			if materialValues.twoDTextures[0] != nil {
				r.commandBuffer.TextureUnit(internal.TextureBindingGeometryAlbedoTexture, materialValues.twoDTextures[0])
			}
			r.commandBuffer.UniformBufferUnit(
				internal.UniformBufferBindingCamera,
				r.cameraPlacement.Buffer,
				r.cameraPlacement.Offset,
				r.cameraPlacement.Size,
			)
			materialPlacement := renderutil.WriteUniform(r.uniforms, internal.MaterialUniform{
				Data: materialValues.uniformData,
			})
			r.commandBuffer.UniformBufferUnit(
				internal.UniformBufferBindingMaterial,
				materialPlacement.Buffer,
				materialPlacement.Offset,
				materialPlacement.Size,
			)
			if item.armature == nil {
				copy(r.modelUniformBufferData[count*64:], item.modelMatrix)
				count++
			} else {
				// No batching for submeshes with armature. Zero index is the model
				// matrix
				copy(r.modelUniformBufferData, item.armature.uniformBufferData)
				modelPlacement := renderutil.WriteUniform(r.uniforms, internal.ModelUniform{
					ModelMatrices: r.modelUniformBufferData,
				})
				r.commandBuffer.UniformBufferUnit(
					internal.UniformBufferBindingModel,
					modelPlacement.Buffer,
					modelPlacement.Offset,
					modelPlacement.Size,
				)
				r.commandBuffer.DrawIndexed(lastFragment.indexOffsetBytes, lastFragment.indexCount, 1)
			}
		}
	}

	// flush remainder
	if lastFragment != nil && count > 0 {
		modelPlacement := renderutil.WriteUniform(r.uniforms, internal.ModelUniform{
			ModelMatrices: r.modelUniformBufferData,
		})
		r.commandBuffer.UniformBufferUnit(
			internal.UniformBufferBindingModel,
			modelPlacement.Buffer,
			modelPlacement.Offset,
			modelPlacement.Size,
		)
		r.commandBuffer.DrawIndexed(lastFragment.indexOffsetBytes, lastFragment.indexCount, count)
	}
}

func (r *sceneRenderer) renderLightingPass(ctx renderCtx) {
	defer metric.BeginRegion("lighting").End()

	r.commandBuffer.BeginRenderPass(render.RenderPassInfo{
		Framebuffer: r.lightingFramebuffer,
		Viewport: render.Area{
			X:      0,
			Y:      0,
			Width:  r.framebufferWidth,
			Height: r.framebufferHeight,
		},
		DepthLoadOp:    render.LoadOperationDontCare, // TODO: LoadOperationLoad: we do care
		DepthStoreOp:   render.StoreOperationStore,
		StencilLoadOp:  render.LoadOperationDontCare,
		StencilStoreOp: render.StoreOperationDontCare,
		Colors: [4]render.ColorAttachmentInfo{
			{
				LoadOp:     render.LoadOperationClear,
				StoreOp:    render.StoreOperationStore,
				ClearValue: [4]float32{0.0, 0.0, 0.0, 1.0},
			},
		},
	})

	r.ambientLightBucket.Reset()
	ctx.scene.ambientLightSet.VisitHexahedronRegion(&ctx.frustum, r.ambientLightBucket)
	for _, ambientLight := range r.ambientLightBucket.Items() {
		if ambientLight.active {
			r.renderAmbientLight(ctx, ambientLight)
		}
	}

	// TODO: Use batching (instancing) when rendering point lights
	r.pointLightBucket.Reset()
	ctx.scene.pointLightSet.VisitHexahedronRegion(&ctx.frustum, r.pointLightBucket)
	for _, pointLight := range r.pointLightBucket.Items() {
		if pointLight.active {
			r.renderPointLight(ctx, pointLight)
		}
	}

	// TODO: Use batching (instancing) when rendering spot lights
	r.spotLightBucket.Reset()
	ctx.scene.spotLightSet.VisitHexahedronRegion(&ctx.frustum, r.spotLightBucket)
	for _, spotLight := range r.spotLightBucket.Items() {
		if spotLight.active {
			r.renderSpotLight(ctx, spotLight)
		}
	}

	r.directionalLightBucket.Reset()
	ctx.scene.directionalLightSet.VisitHexahedronRegion(&ctx.frustum, r.directionalLightBucket)
	for _, directionalLight := range r.directionalLightBucket.Items() {
		if directionalLight.active {
			r.renderDirectionalLight(ctx, directionalLight)
		}
	}

	r.commandBuffer.EndRenderPass()
}

func (r *sceneRenderer) renderAmbientLight(ctx renderCtx, light *AmbientLight) {
	r.commandBuffer.BindPipeline(r.ambientLightPipeline)
	r.commandBuffer.TextureUnit(internal.TextureBindingLightingFramebufferColor0, r.geometryAlbedoTexture)
	r.commandBuffer.TextureUnit(internal.TextureBindingLightingFramebufferColor1, r.geometryNormalTexture)
	r.commandBuffer.TextureUnit(internal.TextureBindingLightingFramebufferDepth, r.geometryDepthTexture)
	r.commandBuffer.TextureUnit(internal.TextureBindingLightingReflectionTexture, light.reflectionTexture.texture)
	r.commandBuffer.TextureUnit(internal.TextureBindingLightingRefractionTexture, light.refractionTexture.texture)
	r.commandBuffer.UniformBufferUnit(
		internal.UniformBufferBindingCamera,
		r.cameraPlacement.Buffer,
		r.cameraPlacement.Offset,
		r.cameraPlacement.Size,
	)
	// TODO: Intensity based on distance and inner and outer radius
	r.commandBuffer.DrawIndexed(0, r.quadShape.IndexCount(), 1)
}

func floor(a float32) float32 {
	return float32(math.Floor(float64(a)))
}

func (r *sceneRenderer) renderDirectionalLight(ctx renderCtx, light *DirectionalLight) {
	projectionMatrix := lightOrtho()
	lightMatrix := light.gfxMatrix()
	lightMatrix.M14 = floor(lightMatrix.M14*shadowMapWidth) / float32(shadowMapWidth)
	lightMatrix.M24 = floor(lightMatrix.M24*shadowMapWidth) / float32(shadowMapWidth)
	lightMatrix.M34 = floor(lightMatrix.M34*shadowMapWidth) / float32(shadowMapWidth)
	viewMatrix := sprec.InverseMat4(lightMatrix)

	lightPlacement := renderutil.WriteUniform(r.uniforms, internal.LightUniform{
		ProjectionMatrix: projectionMatrix,
		ViewMatrix:       viewMatrix,
		LightMatrix:      lightMatrix,
	})

	lightPropertiesPlacement := renderutil.WriteUniform(r.uniforms, internal.LightPropertiesUniform{
		Color:     dtos.Vec3(light.emitColor),
		Intensity: 1.0,
	})

	r.commandBuffer.BindPipeline(r.directionalLightPipeline)
	r.commandBuffer.TextureUnit(internal.TextureBindingLightingFramebufferColor0, r.geometryAlbedoTexture)
	r.commandBuffer.TextureUnit(internal.TextureBindingLightingFramebufferColor1, r.geometryNormalTexture)
	r.commandBuffer.TextureUnit(internal.TextureBindingLightingFramebufferDepth, r.geometryDepthTexture)
	r.commandBuffer.TextureUnit(internal.TextureBindingShadowFramebufferDepth, r.shadowDepthTexture)
	r.commandBuffer.UniformBufferUnit(
		internal.UniformBufferBindingCamera,
		r.cameraPlacement.Buffer,
		r.cameraPlacement.Offset,
		r.cameraPlacement.Size,
	)
	r.commandBuffer.UniformBufferUnit(
		internal.UniformBufferBindingLight,
		lightPlacement.Buffer,
		lightPlacement.Offset,
		lightPlacement.Size,
	)
	r.commandBuffer.UniformBufferUnit(
		internal.UniformBufferBindingLightProperties,
		lightPropertiesPlacement.Buffer,
		lightPropertiesPlacement.Offset,
		lightPropertiesPlacement.Size,
	)
	r.commandBuffer.DrawIndexed(0, r.quadShape.IndexCount(), 1)
}

func (r *sceneRenderer) renderPointLight(ctx renderCtx, light *PointLight) {
	projectionMatrix := sprec.IdentityMat4()
	lightMatrix := light.gfxMatrix()
	viewMatrix := sprec.InverseMat4(lightMatrix)

	lightPlacement := renderutil.WriteUniform(r.uniforms, internal.LightUniform{
		ProjectionMatrix: projectionMatrix,
		ViewMatrix:       viewMatrix,
		LightMatrix:      lightMatrix,
	})

	lightPropertiesPlacement := renderutil.WriteUniform(r.uniforms, internal.LightPropertiesUniform{
		Color:     dtos.Vec3(light.emitColor),
		Intensity: 1.0,
		Range:     float32(light.emitRange),
	})

	r.commandBuffer.BindPipeline(r.pointLightPipeline)
	r.commandBuffer.TextureUnit(internal.TextureBindingLightingFramebufferColor0, r.geometryAlbedoTexture)
	r.commandBuffer.TextureUnit(internal.TextureBindingLightingFramebufferColor1, r.geometryNormalTexture)
	r.commandBuffer.TextureUnit(internal.TextureBindingLightingFramebufferDepth, r.geometryDepthTexture)
	r.commandBuffer.UniformBufferUnit(
		internal.UniformBufferBindingCamera,
		r.cameraPlacement.Buffer,
		r.cameraPlacement.Offset,
		r.cameraPlacement.Size,
	)
	r.commandBuffer.UniformBufferUnit(
		internal.UniformBufferBindingLight,
		lightPlacement.Buffer,
		lightPlacement.Offset,
		lightPlacement.Size,
	)
	r.commandBuffer.UniformBufferUnit(
		internal.UniformBufferBindingLightProperties,
		lightPropertiesPlacement.Buffer,
		lightPropertiesPlacement.Offset,
		lightPropertiesPlacement.Size,
	)
	r.commandBuffer.DrawIndexed(0, r.sphereShape.IndexCount(), 1)
}

func (r *sceneRenderer) renderSpotLight(ctx renderCtx, light *SpotLight) {
	projectionMatrix := sprec.IdentityMat4()
	lightMatrix := light.gfxMatrix()
	viewMatrix := sprec.InverseMat4(lightMatrix)

	lightPlacement := renderutil.WriteUniform(r.uniforms, internal.LightUniform{
		ProjectionMatrix: projectionMatrix,
		ViewMatrix:       viewMatrix,
		LightMatrix:      lightMatrix,
	})

	lightPropertiesPlacement := renderutil.WriteUniform(r.uniforms, internal.LightPropertiesUniform{
		Color:      dtos.Vec3(light.emitColor),
		Intensity:  1.0,
		Range:      float32(light.emitRange),
		OuterAngle: float32(light.emitOuterConeAngle.Radians()),
		InnerAngle: float32(light.emitInnerConeAngle.Radians()),
	})

	r.commandBuffer.BindPipeline(r.spotLightPipeline)
	r.commandBuffer.TextureUnit(internal.TextureBindingLightingFramebufferColor0, r.geometryAlbedoTexture)
	r.commandBuffer.TextureUnit(internal.TextureBindingLightingFramebufferColor1, r.geometryNormalTexture)
	r.commandBuffer.TextureUnit(internal.TextureBindingLightingFramebufferDepth, r.geometryDepthTexture)
	r.commandBuffer.UniformBufferUnit(
		internal.UniformBufferBindingCamera,
		r.cameraPlacement.Buffer,
		r.cameraPlacement.Offset,
		r.cameraPlacement.Size,
	)
	r.commandBuffer.UniformBufferUnit(
		internal.UniformBufferBindingLight,
		lightPlacement.Buffer,
		lightPlacement.Offset,
		lightPlacement.Size,
	)
	r.commandBuffer.UniformBufferUnit(
		internal.UniformBufferBindingLightProperties,
		lightPropertiesPlacement.Buffer,
		lightPropertiesPlacement.Offset,
		lightPropertiesPlacement.Size,
	)
	r.commandBuffer.DrawIndexed(0, r.coneShape.IndexCount(), 1)
}

func (r *sceneRenderer) renderForwardPass(ctx renderCtx) {
	defer metric.BeginRegion("forward").End()

	r.commandBuffer.BeginRenderPass(render.RenderPassInfo{
		Framebuffer: r.forwardFramebuffer,
		Viewport: render.Area{
			X:      0,
			Y:      0,
			Width:  r.framebufferWidth,
			Height: r.framebufferHeight,
		},
		DepthLoadOp:    render.LoadOperationDontCare, // TODO: LoadOperationLoad: we do care
		DepthStoreOp:   render.StoreOperationStore,
		StencilLoadOp:  render.LoadOperationDontCare,
		StencilStoreOp: render.StoreOperationDontCare,
		Colors: [4]render.ColorAttachmentInfo{
			{
				LoadOp:  render.LoadOperationDontCare, // TODO: LoadOperationLoad: we do care
				StoreOp: render.StoreOperationStore,
			},
		},
	})

	sky := ctx.scene.sky
	if texture := sky.skyboxTexture; texture != nil {
		r.commandBuffer.BindPipeline(r.skyboxPipeline)
		r.commandBuffer.UniformBufferUnit(
			internal.UniformBufferBindingCamera,
			r.cameraPlacement.Buffer,
			r.cameraPlacement.Offset,
			r.cameraPlacement.Size,
		)
		r.commandBuffer.TextureUnit(internal.TextureBindingSkyboxAlbedoTexture, texture.texture)
		r.commandBuffer.DrawIndexed(0, r.cubeShape.IndexCount(), 1)
	} else {
		skyboxPlacement := renderutil.WriteUniform(r.uniforms, internal.SkyboxUniform{
			Color: sprec.Vec4{
				X: sky.backgroundColor.X,
				Y: sky.backgroundColor.Y,
				Z: sky.backgroundColor.Z,
				W: 1.0,
			},
		})

		r.commandBuffer.BindPipeline(r.skycolorPipeline)
		r.commandBuffer.UniformBufferUnit(
			internal.UniformBufferBindingCamera,
			r.cameraPlacement.Buffer,
			r.cameraPlacement.Offset,
			r.cameraPlacement.Size,
		)
		r.commandBuffer.UniformBufferUnit(
			internal.UniformBufferBindingSkybox,
			skyboxPlacement.Buffer,
			skyboxPlacement.Offset,
			skyboxPlacement.Size,
		)
		r.commandBuffer.DrawIndexed(0, r.cubeShape.IndexCount(), 1)
	}

	if len(r.debugLines) > 0 {
		plotter := blob.NewPlotter(r.debugVertexData)
		for _, line := range r.debugLines {
			plotter.PlotSPVec3(line.Start)
			plotter.PlotUint8(uint8(line.Color.X * 255))
			plotter.PlotUint8(uint8(line.Color.Y * 255))
			plotter.PlotUint8(uint8(line.Color.Z * 255))
			plotter.PlotUint8(uint8(255))

			plotter.PlotSPVec3(line.End)
			plotter.PlotUint8(uint8(line.Color.X * 255))
			plotter.PlotUint8(uint8(line.Color.Y * 255))
			plotter.PlotUint8(uint8(line.Color.Z * 255))
			plotter.PlotUint8(uint8(255))
		}
		r.api.Queue().WriteBuffer(r.debugVertexBuffer, 0, r.debugVertexData[:plotter.Offset()])
		r.commandBuffer.BindPipeline(r.debugPipeline)
		r.commandBuffer.UniformBufferUnit(
			internal.UniformBufferBindingCamera,
			r.cameraPlacement.Buffer,
			r.cameraPlacement.Offset,
			r.cameraPlacement.Size,
		)
		r.commandBuffer.Draw(0, len(r.debugLines)*2, 1)
	}

	// NOTE: Reusing renderItems and assuming same as geometry pass.
	r.renderForwardMeshesList(ctx)

	r.commandBuffer.EndRenderPass()
}

func (r *sceneRenderer) renderForwardMeshesList(ctx renderCtx) {
	var lastFragment *meshFragmentDefinition
	count := 0
	for _, item := range r.renderItems {
		fragment := item.fragment
		if fragment.material.forwardPipeline == nil {
			continue // skip non-forward meshes
		}

		// just append the modelmatrix
		if fragment == lastFragment && item.armature == nil {
			copy(r.modelUniformBufferData[count*64:], item.modelMatrix)
			count++

			// flush batch
			if count >= 256 {
				modelPlacement := renderutil.WriteUniform(r.uniforms, internal.ModelUniform{
					ModelMatrices: r.modelUniformBufferData,
				})
				r.commandBuffer.UniformBufferUnit(
					internal.UniformBufferBindingModel,
					modelPlacement.Buffer,
					modelPlacement.Offset,
					modelPlacement.Size,
				)
				r.commandBuffer.DrawIndexed(lastFragment.indexOffsetBytes, lastFragment.indexCount, count)
				count = 0
			}
		} else {
			// flush batch
			if lastFragment != nil && count > 0 {
				modelPlacement := renderutil.WriteUniform(r.uniforms, internal.ModelUniform{
					ModelMatrices: r.modelUniformBufferData,
				})
				r.commandBuffer.UniformBufferUnit(
					internal.UniformBufferBindingModel,
					modelPlacement.Buffer,
					modelPlacement.Offset,
					modelPlacement.Size,
				)
				r.commandBuffer.DrawIndexed(lastFragment.indexOffsetBytes, lastFragment.indexCount, count)
			}

			count = 0
			lastFragment = fragment
			material := fragment.material
			materialValues := material.definition
			r.commandBuffer.BindPipeline(material.forwardPipeline)
			r.commandBuffer.UniformBufferUnit(
				internal.UniformBufferBindingCamera,
				r.cameraPlacement.Buffer,
				r.cameraPlacement.Offset,
				r.cameraPlacement.Size,
			)

			materialPlacement := renderutil.WriteUniform(r.uniforms, internal.MaterialUniform{
				Data: materialValues.uniformData,
			})
			r.commandBuffer.UniformBufferUnit(
				internal.UniformBufferBindingMaterial,
				materialPlacement.Buffer,
				materialPlacement.Offset,
				materialPlacement.Size,
			)
			if item.armature == nil {
				copy(r.modelUniformBufferData[count*64:], item.modelMatrix)
				count++
			} else {
				// No batching for submeshes with armature. Zero index is the model
				// matrix
				copy(r.modelUniformBufferData, item.armature.uniformBufferData)
				modelPlacement := renderutil.WriteUniform(r.uniforms, internal.ModelUniform{
					ModelMatrices: r.modelUniformBufferData,
				})
				r.commandBuffer.UniformBufferUnit(
					internal.UniformBufferBindingModel,
					modelPlacement.Buffer,
					modelPlacement.Offset,
					modelPlacement.Size,
				)
				r.commandBuffer.DrawIndexed(lastFragment.indexOffsetBytes, lastFragment.indexCount, 1)
			}
		}
	}

	// flush remainder
	if lastFragment != nil && count > 0 {
		modelPlacement := renderutil.WriteUniform(r.uniforms, internal.ModelUniform{
			ModelMatrices: r.modelUniformBufferData,
		})
		r.commandBuffer.UniformBufferUnit(
			internal.UniformBufferBindingModel,
			modelPlacement.Buffer,
			modelPlacement.Offset,
			modelPlacement.Size,
		)
		r.commandBuffer.DrawIndexed(lastFragment.indexOffsetBytes, lastFragment.indexCount, count)
	}
}

func (r *sceneRenderer) renderExposureProbePass(ctx renderCtx) {
	defer metric.BeginRegion("exposure").End()

	if r.exposureFormat != render.DataFormatRGBA16F && r.exposureFormat != render.DataFormatRGBA32F {
		logger.Error("Skipping exposure due to unsupported framebuffer format (%q)!", r.exposureFormat)
		return
	}

	if r.exposureSync != nil {
		switch r.exposureSync.Status() {
		case render.FenceStatusSuccess:
			r.api.Queue().ReadBuffer(r.exposureBuffer, 0, r.exposureBufferData)
			var brightness float32
			switch r.exposureFormat {
			case render.DataFormatRGBA16F:
				brightness = float16.Frombits(r.exposureBufferData.Uint16(0 * 2)).Float32()
			case render.DataFormatRGBA32F:
				brightness = gblob.LittleEndianBlock(r.exposureBufferData).Float32(0 * 4)
			}
			brightness = sprec.Clamp(brightness, 0.001, 1000.0)

			r.exposureTarget = 1.0 / (2 * 3.14 * brightness)
			if r.exposureTarget > ctx.camera.maxExposure {
				r.exposureTarget = ctx.camera.maxExposure
			}
			if r.exposureTarget < ctx.camera.minExposure {
				r.exposureTarget = ctx.camera.minExposure
			}
			fallthrough

		case render.FenceStatusNotReady:
			// wait until next frame
		}
	}

	if !r.exposureUpdateTimestamp.IsZero() {
		elapsedSeconds := float32(time.Since(r.exposureUpdateTimestamp).Seconds())
		ctx.camera.exposure = sprec.Mix(
			ctx.camera.exposure,
			r.exposureTarget,
			sprec.Clamp(ctx.camera.autoExposureSpeed*elapsedSeconds, 0.0, 1.0),
		)
	}
	r.exposureUpdateTimestamp = time.Now()

	if r.exposureSync == nil {
		r.commandBuffer.BeginRenderPass(render.RenderPassInfo{
			Framebuffer: r.exposureFramebuffer,
			Viewport: render.Area{
				X:      0,
				Y:      0,
				Width:  1,
				Height: 1,
			},
			DepthLoadOp:    render.LoadOperationDontCare,
			DepthStoreOp:   render.StoreOperationDontCare,
			StencilLoadOp:  render.LoadOperationDontCare,
			StencilStoreOp: render.StoreOperationDontCare,
			Colors: [4]render.ColorAttachmentInfo{
				{
					LoadOp:     render.LoadOperationClear,
					StoreOp:    render.StoreOperationDontCare,
					ClearValue: [4]float32{0.0, 0.0, 0.0, 0.0},
				},
			},
		})
		r.commandBuffer.BindPipeline(r.exposurePipeline)
		r.commandBuffer.TextureUnit(internal.TextureBindingLightingFramebufferColor0, r.lightingAlbedoTexture)
		r.commandBuffer.UniformBufferUnit(
			internal.UniformBufferBindingCamera,
			r.cameraPlacement.Buffer,
			r.cameraPlacement.Offset,
			r.cameraPlacement.Size,
		)
		r.commandBuffer.DrawIndexed(0, r.quadShape.IndexCount(), 1)
		r.commandBuffer.CopyFramebufferToBuffer(render.CopyFramebufferToBufferInfo{
			Buffer: r.exposureBuffer,
			X:      0,
			Y:      0,
			Width:  1,
			Height: 1,
			Format: r.exposureFormat,
		})
		r.exposureSync = r.api.Queue().TrackSubmittedWorkDone() // FIXME: Incorrect place
		r.commandBuffer.EndRenderPass()
	}
}

func (r *sceneRenderer) renderPostprocessingPass(ctx renderCtx) {
	defer metric.BeginRegion("post").End()

	postprocessPlacement := renderutil.WriteUniform(r.uniforms, internal.PostprocessUniform{
		Exposure: ctx.camera.exposure,
	})

	r.commandBuffer.BeginRenderPass(render.RenderPassInfo{
		Framebuffer: ctx.framebuffer,
		Viewport: render.Area{
			X:      ctx.x,
			Y:      ctx.y,
			Width:  ctx.width,
			Height: ctx.height,
		},
		DepthLoadOp:    render.LoadOperationDontCare,
		DepthStoreOp:   render.StoreOperationDontCare,
		StencilLoadOp:  render.LoadOperationDontCare,
		StencilStoreOp: render.StoreOperationStore, // TODO: We need this due to UI. Figure out how to control this.
		Colors: [4]render.ColorAttachmentInfo{
			{
				LoadOp:  render.LoadOperationDontCare,
				StoreOp: render.StoreOperationStore,
			},
		},
	})

	r.commandBuffer.BindPipeline(r.postprocessingPipeline)
	r.commandBuffer.TextureUnit(internal.TextureBindingPostprocessFramebufferColor0, r.lightingAlbedoTexture)
	r.commandBuffer.UniformBufferUnit(
		internal.UniformBufferBindingCamera,
		r.cameraPlacement.Buffer,
		r.cameraPlacement.Offset,
		r.cameraPlacement.Size,
	)
	r.commandBuffer.UniformBufferUnit(
		internal.UniformBufferBindingPostprocess,
		postprocessPlacement.Buffer,
		postprocessPlacement.Offset,
		postprocessPlacement.Size,
	)
	r.commandBuffer.DrawIndexed(0, r.quadShape.IndexCount(), 1)

	r.commandBuffer.EndRenderPass()
}

type renderCtx struct {
	framebuffer render.Framebuffer
	scene       *Scene
	x           int
	y           int
	width       int
	height      int
	camera      *Camera
	frustum     spatial.HexahedronRegion
}

type renderItem struct {
	fragment    *meshFragmentDefinition
	armature    *Armature
	modelMatrix []byte
}
