package graphics

import (
	"fmt"
	"time"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/gomath/stod"
	"github.com/mokiat/lacking/game/graphics/internal"
	"github.com/mokiat/lacking/log"
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/util/blob"
	"github.com/mokiat/lacking/util/metrics"
	"github.com/mokiat/lacking/util/shape"
	"github.com/mokiat/lacking/util/spatial"
	"github.com/x448/float16"
	"golang.org/x/exp/slices"
)

const (
	shadowMapWidth  = 4096
	shadowMapHeight = 4096
)

var ShowLightView bool

func newRenderer(api render.API, shaders ShaderCollection) *sceneRenderer {
	return &sceneRenderer{
		api:     api,
		shaders: shaders,

		exposureBufferData: make([]byte, 4*render.SizeF32), // Worst case RGBA32F
		exposureTarget:     1.0,

		visibleMeshes: spatial.NewVisitorBucket[*Mesh](2_000_000),
	}
}

type sceneRenderer struct {
	api      render.API
	shaders  ShaderCollection
	commands render.CommandQueue

	framebufferWidth  int
	framebufferHeight int

	quadShape *internal.Shape
	cubeShape *internal.Shape

	geometryAlbedoTexture render.Texture
	geometryNormalTexture render.Texture
	geometryDepthTexture  render.Texture
	geometryFramebuffer   render.Framebuffer

	lightingAlbedoTexture render.Texture
	lightingFramebuffer   render.Framebuffer

	forwardFramebuffer render.Framebuffer

	shadowFramebuffer  render.Framebuffer
	shadowDepthTexture render.Texture

	exposureAlbedoTexture   render.Texture
	exposureFramebuffer     render.Framebuffer
	exposurePresentation    *internal.LightingPresentation
	exposurePipeline        render.Pipeline
	exposureBufferData      blob.Buffer
	exposureBuffer          render.Buffer
	exposureFormat          render.DataFormat
	exposureSync            render.Fence
	exposureTarget          float32
	exposureUpdateTimestamp time.Time

	postprocessingPresentation *internal.PostprocessingPresentation
	postprocessingPipeline     render.Pipeline

	directionalLightPresentation *internal.LightingPresentation
	directionalLightPipeline     render.Pipeline
	ambientLightPresentation     *internal.LightingPresentation
	ambientLightPipeline         render.Pipeline
	pointLightPresentation       *internal.LightingPresentation
	pointLightPipeline           render.Pipeline

	skyboxPresentation   *internal.SkyboxPresentation
	skyboxPipeline       render.Pipeline
	skycolorPresentation *internal.SkyboxPresentation
	skycolorPipeline     render.Pipeline

	debugLines        []debugLine
	debugVertexData   []byte
	debugVertexBuffer render.Buffer
	debugVertexArray  render.VertexArray
	debugPipeline     render.Pipeline

	cameraUniformBufferData blob.Buffer
	cameraUniformBuffer     render.Buffer

	modelUniformBufferData blob.Buffer
	modelUniformBuffer     render.Buffer

	materialUniformBufferData blob.Buffer
	materialUniformBuffer     render.Buffer

	lightUniformBufferData blob.Buffer
	lightUniformBuffer     render.Buffer

	visibleMeshes *spatial.VisitorBucket[*Mesh]
	renderItems   []renderItem
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
	r.commands = r.api.CreateCommandQueue()

	r.createFramebuffers(800, 600)

	r.quadShape = internal.CreateQuadShape(r.api)
	r.cubeShape = internal.CreateCubeShape(r.api)

	defaultShadowDepth := float32(1.0)
	r.shadowDepthTexture = r.api.CreateDepthTexture2D(render.DepthTexture2DInfo{
		Width:        shadowMapWidth,
		Height:       shadowMapHeight,
		ClippedValue: &defaultShadowDepth,
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
	exposureShaders := r.shaders.ExposureSet()
	r.exposurePresentation = internal.NewLightingPresentation(r.api,
		exposureShaders.VertexShader,
		exposureShaders.FragmentShader,
	)
	r.exposurePipeline = r.api.CreatePipeline(render.PipelineInfo{
		Program:      r.exposurePresentation.Program,
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
	r.exposureFormat = r.api.DetermineContentFormat(r.exposureFramebuffer)

	postprocessingShaders := r.shaders.PostprocessingSet(PostprocessingShaderConfig{
		ToneMapping: ExponentialToneMapping,
	})
	r.postprocessingPresentation = internal.NewPostprocessingPresentation(r.api,
		postprocessingShaders.VertexShader,
		postprocessingShaders.FragmentShader,
	)
	r.postprocessingPipeline = r.api.CreatePipeline(render.PipelineInfo{
		Program:         r.postprocessingPresentation.Program,
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

	directionalLightShaders := r.shaders.DirectionalLightSet()
	r.directionalLightPresentation = internal.NewLightingPresentation(r.api,
		directionalLightShaders.VertexShader,
		directionalLightShaders.FragmentShader,
	)
	r.directionalLightPipeline = r.api.CreatePipeline(render.PipelineInfo{
		Program:                     r.directionalLightPresentation.Program,
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
	ambientLightShaders := r.shaders.AmbientLightSet()
	r.ambientLightPresentation = internal.NewLightingPresentation(r.api,
		ambientLightShaders.VertexShader,
		ambientLightShaders.FragmentShader,
	)
	r.ambientLightPipeline = r.api.CreatePipeline(render.PipelineInfo{
		Program:                     r.ambientLightPresentation.Program,
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
	pointLightShaders := r.shaders.PointLightSet()
	r.pointLightPresentation = internal.NewLightingPresentation(r.api,
		pointLightShaders.VertexShader,
		pointLightShaders.FragmentShader,
	)
	r.pointLightPipeline = r.api.CreatePipeline(render.PipelineInfo{
		Program:                     r.pointLightPresentation.Program,
		VertexArray:                 r.quadShape.VertexArray(), // TODO: Sphere (r=1) mesh!
		Topology:                    r.quadShape.Topology(),    // TODO: Sphere (r=1) mesh!
		Culling:                     render.CullModeBack,
		FrontFace:                   render.FaceOrientationCCW,
		DepthTest:                   false, // TODO: True, once sphere shape is used!
		DepthWrite:                  false,
		DepthComparison:             render.ComparisonAlways, // TODO: LEQUAL, once sphere shape is used!
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

	skyboxShaders := r.shaders.SkyboxSet()
	r.skyboxPresentation = internal.NewSkyboxPresentation(r.api,
		skyboxShaders.VertexShader,
		skyboxShaders.FragmentShader,
	)
	r.skyboxPipeline = r.api.CreatePipeline(render.PipelineInfo{
		Program:     r.skyboxPresentation.Program,
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

	skycolorShaders := r.shaders.SkycolorSet()
	r.skycolorPresentation = internal.NewSkyboxPresentation(r.api,
		skycolorShaders.VertexShader,
		skycolorShaders.FragmentShader,
	)
	r.skycolorPipeline = r.api.CreatePipeline(render.PipelineInfo{
		Program:     r.skycolorPresentation.Program,
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
		Bindings: []render.VertexArrayBindingInfo{
			{
				VertexBuffer: r.debugVertexBuffer,
				Stride:       4 * 4,
			},
		},
		Attributes: []render.VertexArrayAttributeInfo{
			{
				Binding:  0,
				Location: internal.CoordAttributeIndex,
				Format:   render.VertexAttributeFormatRGB32F,
				Offset:   0,
			},
			{
				Binding:  0,
				Location: internal.ColorAttributeIndex,
				Format:   render.VertexAttributeFormatRGB8UN,
				Offset:   3 * 4,
			},
		},
	})
	debugShaders := r.shaders.DebugSet()
	debugProgram := internal.BuildProgram(r.api, debugShaders.VertexShader, debugShaders.FragmentShader, nil, []render.UniformBinding{
		{
			Name:  "Camera",
			Index: internal.UniformBufferBindingCamera,
		},
	})
	r.debugPipeline = r.api.CreatePipeline(render.PipelineInfo{
		Program:         debugProgram,
		VertexArray:     r.debugVertexArray,
		Topology:        render.TopologyLines,
		Culling:         render.CullModeNone,
		FrontFace:       render.FaceOrientationCCW,
		DepthTest:       true,
		DepthWrite:      false,
		DepthComparison: render.ComparisonLessOrEqual,
		StencilTest:     false,
		ColorWrite:      render.ColorMaskTrue,
		BlendEnabled:    false,
	})

	r.cameraUniformBufferData = make([]byte, 3*64) // 3 x mat4
	r.cameraUniformBuffer = r.api.CreateUniformBuffer(render.BufferInfo{
		Dynamic: true,
		Size:    len(r.cameraUniformBufferData),
	})
	r.modelUniformBufferData = make([]byte, 64*256) // 256 x mat4
	r.modelUniformBuffer = r.api.CreateUniformBuffer(render.BufferInfo{
		Dynamic: true,
		Size:    len(r.modelUniformBufferData),
	})
	r.materialUniformBufferData = make([]byte, 2*4*4) // 2 x vec4
	r.materialUniformBuffer = r.api.CreateUniformBuffer(render.BufferInfo{
		Dynamic: true,
		Size:    len(r.materialUniformBufferData),
	})
	r.lightUniformBufferData = make([]byte, 3*64) // 2 x mat4
	r.lightUniformBuffer = r.api.CreateUniformBuffer(render.BufferInfo{
		Dynamic: true,
		Size:    len(r.lightUniformBufferData),
	})
}

func (r *sceneRenderer) Release() {
	defer r.commands.Release()

	defer r.releaseFramebuffers()

	defer r.quadShape.Release()
	defer r.cubeShape.Release()

	defer r.shadowDepthTexture.Release()
	defer r.shadowFramebuffer.Release()

	defer r.exposureAlbedoTexture.Release()
	defer r.exposureFramebuffer.Release()
	defer r.exposurePresentation.Delete()
	defer r.exposurePipeline.Release()
	defer r.exposureBuffer.Release()

	defer r.postprocessingPresentation.Delete()
	defer r.postprocessingPipeline.Release()

	defer r.directionalLightPresentation.Delete()
	defer r.directionalLightPipeline.Release()
	defer r.ambientLightPresentation.Delete()
	defer r.ambientLightPipeline.Release()

	defer r.skyboxPresentation.Delete()
	defer r.skyboxPipeline.Release()
	defer r.skycolorPresentation.Delete()
	defer r.skycolorPipeline.Release()

	defer r.cameraUniformBuffer.Release()
	defer r.modelUniformBuffer.Release()
	defer r.materialUniformBuffer.Release()
	defer r.lightUniformBuffer.Release()
}

func (r *sceneRenderer) ResetDebugLines() {
	r.debugLines = r.debugLines[:0]
}

func (r *sceneRenderer) QueueDebugLine(line debugLine) {
	if len(r.debugLines) == cap(r.debugLines) {
		log.Warn("No more debug lines allowed")
		return
	}
	r.debugLines = append(r.debugLines, line)
}

func (r *sceneRenderer) Ray(viewport Viewport, camera *Camera, x, y int) shape.StaticLine {
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

	return shape.NewStaticLine(shape.Point(a.VecXYZ()), shape.Point(b.VecXYZ()))
}

func (r *sceneRenderer) Render(framebuffer render.Framebuffer, viewport Viewport, scene *Scene, camera *Camera) {
	r.api.Invalidate()

	if viewport.Width != r.framebufferWidth || viewport.Height != r.framebufferHeight {
		r.releaseFramebuffers()
		r.createFramebuffers(viewport.Width, viewport.Height)
	}

	projectionMatrix := r.evaluateProjectionMatrix(camera, viewport.Width, viewport.Height)
	cameraMatrix := camera.gfxMatrix()
	viewMatrix := sprec.InverseMat4(cameraMatrix)
	projectionViewMatrix := sprec.Mat4Prod(projectionMatrix, viewMatrix)
	frustum := spatial.ProjectionRegion(stod.Mat4(projectionViewMatrix))

	if scene.firstLight != nil && ShowLightView {
		dirLight := scene.firstLight.next
		projectionMatrix = lightOrtho()
		cameraMatrix = dirLight.gfxMatrix()
		viewMatrix = sprec.InverseMat4(cameraMatrix)
		projectionViewMatrix = sprec.Mat4Prod(projectionMatrix, viewMatrix)
		frustum = spatial.ProjectionRegion(stod.Mat4(projectionViewMatrix))
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

	cameraPlotter := blob.NewPlotter(r.cameraUniformBufferData)
	cameraPlotter.PlotSPMat4(projectionMatrix)
	cameraPlotter.PlotSPMat4(viewMatrix)
	cameraPlotter.PlotSPMat4(cameraMatrix)
	r.cameraUniformBuffer.Update(render.BufferUpdateInfo{
		Data: r.cameraUniformBufferData,
	})

	r.api.UniformBufferUnit(internal.UniformBufferBindingCamera, r.cameraUniformBuffer)
	r.api.UniformBufferUnit(internal.UniformBufferBindingModel, r.modelUniformBuffer)
	r.api.UniformBufferUnit(internal.UniformBufferBindingMaterial, r.materialUniformBuffer)
	r.api.UniformBufferUnit(internal.UniformBufferBindingLight, r.lightUniformBuffer)

	r.renderShadowPass(ctx)
	r.renderGeometryPass(ctx)
	r.renderLightingPass(ctx)
	r.renderForwardPass(ctx)
	if camera.autoExposureEnabled {
		r.renderExposureProbePass(ctx)
	}
	r.renderPostprocessingPass(ctx)
}

func (r *sceneRenderer) evaluateProjectionMatrix(camera *Camera, width, height int) sprec.Mat4 {
	const (
		near = float32(0.5)
		far  = float32(1600.0) // At 400 on a flat plane with forests, you don't really notice the far clipping plane.
	)
	var (
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
	return sprec.OrthoMat4(-64, 64, 64, -64, 0, 200)
}

func (r *sceneRenderer) renderShadowPass(ctx renderCtx) {
	defer metrics.BeginSpan("shadow pass").End()

	// TODO: Support array of shadow-casting lights, not just one
	var directionalLight *Light
	for light := ctx.scene.firstLight; light != nil; light = light.next {
		if light.mode == LightModeDirectional {
			directionalLight = light
			break
		}
	}
	if directionalLight == nil {
		return
	}

	projectionMatrix := lightOrtho()
	lightMatrix := directionalLight.gfxMatrix()
	viewMatrix := sprec.InverseMat4(lightMatrix)
	projectionViewMatrix := sprec.Mat4Prod(projectionMatrix, viewMatrix)
	frustum := spatial.ProjectionRegion(stod.Mat4(projectionViewMatrix))

	r.renderItems = r.renderItems[:0]
	ctx.scene.meshOctree.VisitHexahedronRegion(&frustum, r.visibleMeshes)
	for _, mesh := range r.visibleMeshes.Items() {
		r.queueShadowMesh(ctx, mesh)
	}
	slices.SortFunc(r.renderItems, func(a, b renderItem) bool {
		// TODO: If fragment IDs are stored inside the items thelsevles, there
		// would be fewer pointer jumps and cache misses.
		return a.fragment.id < b.fragment.id
	})

	r.api.BeginRenderPass(render.RenderPassInfo{
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

	lightPlotter := blob.NewPlotter(r.lightUniformBufferData)
	lightPlotter.PlotSPMat4(projectionMatrix)
	lightPlotter.PlotSPMat4(viewMatrix)
	lightPlotter.PlotSPMat4(lightMatrix)
	r.lightUniformBuffer.Update(render.BufferUpdateInfo{
		Data: r.lightUniformBufferData,
	})
	r.renderShadowMeshesList(ctx)

	r.api.SubmitQueue(r.commands)
	r.api.EndRenderPass()
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
				r.commands.UpdateBufferData(r.modelUniformBuffer, render.BufferUpdateInfo{
					Data:   r.modelUniformBufferData,
					Offset: 0,
				})
				r.commands.DrawIndexed(lastFragment.indexOffsetBytes, lastFragment.indexCount, count)
				count = 0
			}
		} else {
			// flush batch
			if lastFragment != nil && count > 0 {
				r.commands.UpdateBufferData(r.modelUniformBuffer, render.BufferUpdateInfo{
					Data:   r.modelUniformBufferData,
					Offset: 0,
				})
				r.commands.DrawIndexed(lastFragment.indexOffsetBytes, lastFragment.indexCount, count)
			}

			count = 0
			lastFragment = fragment
			r.commands.BindPipeline(material.shadowPipeline)
			if item.armature == nil {
				copy(r.modelUniformBufferData[count*64:], item.modelMatrix)
				count++
			} else {
				// No batching for submeshes with armature. Zero index is the model
				// matrix
				copy(r.modelUniformBufferData, item.armature.uniformBufferData)
				r.commands.UpdateBufferData(r.modelUniformBuffer, render.BufferUpdateInfo{
					Data:   r.modelUniformBufferData,
					Offset: 0,
				})
				r.commands.DrawIndexed(lastFragment.indexOffsetBytes, lastFragment.indexCount, 1)
			}
		}
	}

	// flush remainder
	if lastFragment != nil && count > 0 {
		r.commands.UpdateBufferData(r.modelUniformBuffer, render.BufferUpdateInfo{
			Data:   r.modelUniformBufferData,
			Offset: 0,
		})
		r.commands.DrawIndexed(lastFragment.indexOffsetBytes, lastFragment.indexCount, count)
	}
}

func (r *sceneRenderer) renderGeometryPass(ctx renderCtx) {
	defer metrics.BeginSpan("geometry pass").End()

	r.api.BeginRenderPass(render.RenderPassInfo{
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
	ctx.scene.meshOctree.VisitHexahedronRegion(&ctx.frustum, r.visibleMeshes)
	for _, mesh := range r.visibleMeshes.Items() {
		r.queueMesh(ctx, mesh)
	}
	slices.SortFunc(r.renderItems, func(a, b renderItem) bool {
		return a.fragment.id < b.fragment.id
	})
	r.renderMeshesList(ctx)

	r.api.SubmitQueue(r.commands)
	r.api.EndRenderPass()
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

func (r *sceneRenderer) renderMeshesList(ctx renderCtx) {
	var lastFragment *meshFragmentDefinition
	count := 0
	for _, item := range r.renderItems {
		fragment := item.fragment

		// just append the modelmatrix
		if fragment == lastFragment && item.armature == nil {
			copy(r.modelUniformBufferData[count*64:], item.modelMatrix)
			count++

			// flush batch
			if count >= 256 {
				r.commands.UpdateBufferData(r.modelUniformBuffer, render.BufferUpdateInfo{
					Data:   r.modelUniformBufferData,
					Offset: 0,
				})
				r.commands.DrawIndexed(lastFragment.indexOffsetBytes, lastFragment.indexCount, count)
				count = 0
			}
		} else {
			// flush batch
			if lastFragment != nil && count > 0 {
				r.commands.UpdateBufferData(r.modelUniformBuffer, render.BufferUpdateInfo{
					Data:   r.modelUniformBufferData,
					Offset: 0,
				})
				r.commands.DrawIndexed(lastFragment.indexOffsetBytes, lastFragment.indexCount, count)
			}

			count = 0
			lastFragment = fragment
			material := fragment.material
			materialValues := material.definition
			copy(r.materialUniformBufferData, materialValues.uniformData)
			r.commands.BindPipeline(material.geometryPipeline)
			if materialValues.twoDTextures[0] != nil {
				r.commands.TextureUnit(internal.TextureBindingGeometryAlbedoTexture, materialValues.twoDTextures[0])
			}
			r.commands.UpdateBufferData(r.materialUniformBuffer, render.BufferUpdateInfo{
				Data:   r.materialUniformBufferData,
				Offset: 0,
			})
			if item.armature == nil {
				copy(r.modelUniformBufferData[count*64:], item.modelMatrix)
				count++
			} else {
				// No batching for submeshes with armature. Zero index is the model
				// matrix
				copy(r.modelUniformBufferData, item.armature.uniformBufferData)
				r.commands.UpdateBufferData(r.modelUniformBuffer, render.BufferUpdateInfo{
					Data:   r.modelUniformBufferData,
					Offset: 0,
				})
				r.commands.DrawIndexed(lastFragment.indexOffsetBytes, lastFragment.indexCount, 1)
			}
		}
	}

	// flush remainder
	if lastFragment != nil && count > 0 {
		r.commands.UpdateBufferData(r.modelUniformBuffer, render.BufferUpdateInfo{
			Data:   r.modelUniformBufferData,
			Offset: 0,
		})
		r.commands.DrawIndexed(lastFragment.indexOffsetBytes, lastFragment.indexCount, count)
	}
}

func (r *sceneRenderer) renderLightingPass(ctx renderCtx) {
	defer metrics.BeginSpan("lighting pass").End()

	r.api.BeginRenderPass(render.RenderPassInfo{
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
	// TODO: Traverse octree
	for light := ctx.scene.firstLight; light != nil; light = light.next {
		switch light.mode {
		case LightModeDirectional:
			r.renderDirectionalLight(ctx, light)
		case LightModeAmbient:
			r.renderAmbientLight(ctx, light)
		case LightModePoint:
			r.renderPointLight(ctx, light)
		}
	}
	r.api.SubmitQueue(r.commands)
	r.api.EndRenderPass()
}

func (r *sceneRenderer) renderAmbientLight(ctx renderCtx, light *Light) {
	r.commands.BindPipeline(r.ambientLightPipeline)
	r.commands.TextureUnit(internal.TextureBindingLightingFramebufferColor0, r.geometryAlbedoTexture)
	r.commands.TextureUnit(internal.TextureBindingLightingFramebufferColor1, r.geometryNormalTexture)
	r.commands.TextureUnit(internal.TextureBindingLightingFramebufferDepth, r.geometryDepthTexture)
	r.commands.TextureUnit(internal.TextureBindingLightingReflectionTexture, light.reflectionTexture.texture)
	r.commands.TextureUnit(internal.TextureBindingLightingRefractionTexture, light.refractionTexture.texture)
	r.commands.DrawIndexed(0, r.quadShape.IndexCount(), 1)
}

func (r *sceneRenderer) renderDirectionalLight(ctx renderCtx, light *Light) {
	// TODO: Update Light uniform

	r.commands.BindPipeline(r.directionalLightPipeline)
	direction := light.gfxMatrix().OrientationZ()
	r.commands.Uniform3f(r.directionalLightPresentation.LightDirection, direction.Array())
	intensity := light.intensity
	r.commands.Uniform3f(r.directionalLightPresentation.LightIntensity, intensity.Array())
	r.commands.TextureUnit(internal.TextureBindingLightingFramebufferColor0, r.geometryAlbedoTexture)
	r.commands.TextureUnit(internal.TextureBindingLightingFramebufferColor1, r.geometryNormalTexture)
	r.commands.TextureUnit(internal.TextureBindingLightingFramebufferDepth, r.geometryDepthTexture)
	r.commands.TextureUnit(internal.TextureBindingShadowFramebufferDepth, r.shadowDepthTexture)
	r.commands.DrawIndexed(0, r.quadShape.IndexCount(), 1)
}

func (r *sceneRenderer) renderPointLight(ctx renderCtx, light *Light) {
	projectionMatrix := sprec.IdentityMat4()
	lightMatrix := light.gfxMatrix()
	viewMatrix := sprec.InverseMat4(lightMatrix)

	lightPlotter := blob.NewPlotter(r.lightUniformBufferData)
	lightPlotter.PlotSPMat4(projectionMatrix)
	lightPlotter.PlotSPMat4(viewMatrix)
	lightPlotter.PlotSPMat4(lightMatrix)

	r.commands.BindPipeline(r.pointLightPipeline)
	r.commands.UpdateBufferData(r.lightUniformBuffer, render.BufferUpdateInfo{
		Data: r.lightUniformBufferData,
	})
	r.commands.Uniform3f(r.pointLightPresentation.LightIntensity, light.intensity.Array())
	r.commands.TextureUnit(internal.TextureBindingLightingFramebufferColor0, r.geometryAlbedoTexture)
	r.commands.TextureUnit(internal.TextureBindingLightingFramebufferColor1, r.geometryNormalTexture)
	r.commands.TextureUnit(internal.TextureBindingLightingFramebufferDepth, r.geometryDepthTexture)
	// TODO: Use a sphere mesh positioned where the light is!
	r.commands.DrawIndexed(0, r.quadShape.IndexCount(), 1)
}

func (r *sceneRenderer) renderForwardPass(ctx renderCtx) {
	defer metrics.BeginSpan("forward pass").End()

	r.api.BeginRenderPass(render.RenderPassInfo{
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
		r.commands.BindPipeline(r.skyboxPipeline)
		r.commands.TextureUnit(internal.TextureBindingSkyboxAlbedoTexture, texture.texture)
		r.commands.DrawIndexed(0, r.cubeShape.IndexCount(), 1)
	} else {
		r.commands.BindPipeline(r.skycolorPipeline)
		r.commands.Uniform4f(r.skycolorPresentation.AlbedoColorLocation, [4]float32{
			sky.backgroundColor.X,
			sky.backgroundColor.Y,
			sky.backgroundColor.Z,
			1.0,
		})
		r.commands.DrawIndexed(0, r.cubeShape.IndexCount(), 1)
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
		r.commands.UpdateBufferData(r.debugVertexBuffer, render.BufferUpdateInfo{
			Data:   r.debugVertexData[:plotter.Offset()],
			Offset: 0,
		})
		r.commands.BindPipeline(r.debugPipeline)
		r.commands.Draw(0, len(r.debugLines)*2, 1)
	}

	r.api.SubmitQueue(r.commands)
	r.api.EndRenderPass()
}

func (r *sceneRenderer) renderExposureProbePass(ctx renderCtx) {
	defer metrics.BeginSpan("exposure pass").End()

	if r.exposureFormat != render.DataFormatRGBA16F && r.exposureFormat != render.DataFormatRGBA32F {
		log.Error("Skipping exposure due to unsupported framebuffer format %q", r.exposureFormat)
		return
	}

	if r.exposureSync != nil {
		switch r.exposureSync.Status() {
		case render.FenceStatusSuccess:
			r.exposureBuffer.Fetch(render.BufferFetchInfo{
				Offset: 0,
				Target: r.exposureBufferData,
			})
			var brightness float32
			switch r.exposureFormat {
			case render.DataFormatRGBA16F:
				brightness = float16.Frombits(r.exposureBufferData.Uint16(0 * 2)).Float32()
			case render.DataFormatRGBA32F:
				brightness = blob.Buffer(r.exposureBufferData).Float32(0 * 4)
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

		case render.FenceStatusDeviceLost:
			r.exposureSync.Delete()
			r.exposureSync = nil

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
		r.api.BeginRenderPass(render.RenderPassInfo{
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
		r.commands.BindPipeline(r.exposurePipeline)
		r.commands.TextureUnit(internal.TextureBindingLightingFramebufferColor0, r.lightingAlbedoTexture)
		r.commands.DrawIndexed(0, r.quadShape.IndexCount(), 1)
		r.commands.CopyContentToBuffer(render.CopyContentToBufferInfo{
			Buffer: r.exposureBuffer,
			X:      0,
			Y:      0,
			Width:  1,
			Height: 1,
			Format: r.exposureFormat,
		})
		r.api.SubmitQueue(r.commands)
		r.exposureSync = r.api.CreateFence()
		r.api.EndRenderPass()
	}
}

func (r *sceneRenderer) renderPostprocessingPass(ctx renderCtx) {
	defer metrics.BeginSpan("postprocessing pass").End()

	r.api.BeginRenderPass(render.RenderPassInfo{
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

	r.commands.BindPipeline(r.postprocessingPipeline)
	r.commands.TextureUnit(internal.TextureBindingPostprocessFramebufferColor0, r.lightingAlbedoTexture)
	r.commands.Uniform1f(r.postprocessingPresentation.ExposureLocation, ctx.camera.exposure)
	r.commands.DrawIndexed(0, r.quadShape.IndexCount(), 1)
	r.api.SubmitQueue(r.commands)
	r.api.EndRenderPass()
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
