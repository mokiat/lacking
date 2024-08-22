package graphics

import (
	"cmp"
	"fmt"
	"math"
	"slices"

	"github.com/mokiat/gblob"
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/dtos"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/gomath/stod"
	"github.com/mokiat/lacking/debug/metric"
	"github.com/mokiat/lacking/game/graphics/internal"
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/render/ubo"
	"github.com/mokiat/lacking/util/spatial"
)

const (
	shadowMapWidth  = 2048
	shadowMapHeight = 2048

	commandBufferSize = 2 * 1024 * 1024  // 2MB
	uniformBufferSize = 32 * 1024 * 1024 // 32MB
)

var ShowLightView bool

func newRenderer(api render.API, shaders ShaderCollection, stageData *commonStageData) *sceneRenderer {
	return &sceneRenderer{
		api:     api,
		shaders: shaders,

		stageData: stageData,

		visibleStaticMeshes: spatial.NewVisitorBucket[uint32](2_000),
		visibleMeshes:       spatial.NewVisitorBucket[*Mesh](2_000),

		litStaticMeshes: spatial.NewVisitorBucket[uint32](2_000),
		litMeshes:       spatial.NewVisitorBucket[*Mesh](2_000),

		ambientLightBucket: spatial.NewVisitorBucket[*AmbientLight](16),

		pointLightBucket: spatial.NewVisitorBucket[*PointLight](16),

		spotLightBucket: spatial.NewVisitorBucket[*SpotLight](16),

		directionalLightBucket: spatial.NewVisitorBucket[*DirectionalLight](16),

		debugLines: make([]DebugLine, debugMaxLineCount),
	}
}

type sceneRenderer struct {
	api     render.API
	shaders ShaderCollection

	stageData *commonStageData

	framebufferWidth  uint32
	framebufferHeight uint32

	// TODO: Use the ones from stageData
	nearestSampler render.Sampler
	linearSampler  render.Sampler
	depthSampler   render.Sampler

	// TODO: Create dedicated Source stages for these.
	geometryAlbedoTexture render.Texture
	geometryNormalTexture render.Texture
	geometryDepthTexture  render.Texture
	geometryFramebuffer   render.Framebuffer

	lightingAlbedoTexture render.Texture
	lightingFramebuffer   render.Framebuffer

	shadowDepthTexture render.Texture
	shadowFramebuffer  render.Framebuffer

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

	debugLines []DebugLine

	visibleStaticMeshes *spatial.VisitorBucket[uint32]
	visibleMeshes       *spatial.VisitorBucket[*Mesh]

	litStaticMeshes *spatial.VisitorBucket[uint32]
	litMeshes       *spatial.VisitorBucket[*Mesh]

	renderItems []renderItem

	meshRenderer *meshRenderer

	modelUniformBufferData gblob.LittleEndianBlock
	cameraPlacement        ubo.UniformPlacement

	stages []Stage
}

func (r *sceneRenderer) createFramebuffers(width, height uint32) {
	r.framebufferWidth = width
	r.framebufferHeight = height

	r.geometryAlbedoTexture = r.api.CreateColorTexture2D(render.ColorTexture2DInfo{
		Width:           r.framebufferWidth,
		Height:          r.framebufferHeight,
		GenerateMipmaps: false,
		GammaCorrection: false,
		Format:          render.DataFormatRGBA8,
	})
	r.geometryNormalTexture = r.api.CreateColorTexture2D(render.ColorTexture2DInfo{
		Width:           r.framebufferWidth,
		Height:          r.framebufferHeight,
		GenerateMipmaps: false,
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
		GenerateMipmaps: false,
		GammaCorrection: false,
		Format:          render.DataFormatRGBA16F,
	})
	r.lightingFramebuffer = r.api.CreateFramebuffer(render.FramebufferInfo{
		ColorAttachments: [4]render.Texture{
			r.lightingAlbedoTexture,
		},
	})
}

func (r *sceneRenderer) releaseFramebuffers() {
	defer r.geometryAlbedoTexture.Release()
	defer r.geometryNormalTexture.Release()
	defer r.geometryDepthTexture.Release()
	defer r.geometryFramebuffer.Release()

	defer r.lightingAlbedoTexture.Release()
	defer r.lightingFramebuffer.Release()
}

func (r *sceneRenderer) Allocate() {
	r.createFramebuffers(800, 600)

	r.nearestSampler = r.api.CreateSampler(render.SamplerInfo{
		Wrapping:   render.WrapModeClamp,
		Filtering:  render.FilterModeNearest,
		Mipmapping: false,
	})
	r.linearSampler = r.api.CreateSampler(render.SamplerInfo{
		Wrapping:   render.WrapModeClamp,
		Filtering:  render.FilterModeLinear,
		Mipmapping: false,
	})
	r.depthSampler = r.api.CreateSampler(render.SamplerInfo{
		Wrapping:   render.WrapModeClamp,
		Filtering:  render.FilterModeLinear,
		Comparison: opt.V(render.ComparisonLess),
		Mipmapping: false,
	})

	quadShape := r.stageData.QuadShape()
	sphereShape := r.stageData.SphereShape()
	coneShape := r.stageData.ConeShape()

	r.shadowDepthTexture = r.api.CreateDepthTexture2D(render.DepthTexture2DInfo{
		Width:      shadowMapWidth,
		Height:     shadowMapHeight,
		Comparable: true,
	})
	r.shadowFramebuffer = r.api.CreateFramebuffer(render.FramebufferInfo{
		DepthAttachment: r.shadowDepthTexture,
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
		VertexArray:                 quadShape.VertexArray(),
		Topology:                    quadShape.Topology(),
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
		VertexArray:                 sphereShape.VertexArray(),
		Topology:                    sphereShape.Topology(),
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
		VertexArray:                 coneShape.VertexArray(),
		Topology:                    coneShape.Topology(),
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
		VertexArray:                 quadShape.VertexArray(),
		Topology:                    quadShape.Topology(),
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

	r.modelUniformBufferData = make([]byte, modelUniformBufferSize)

	r.meshRenderer = newMeshRenderer()

	forwardStage := newForwardStage(r.api, r.shaders, r.stageData, r.meshRenderer, ForwardStageInput{
		HDRTexture: func() render.Texture {
			return r.lightingAlbedoTexture
		},
		DepthTexture: func() render.Texture {
			return r.geometryDepthTexture
		},
	})
	exposureProbeStage := newExposureProbeStage(r.api, r.shaders, r.stageData, ExposureProbeStageInput{
		HDRTexture: func() render.Texture {
			return r.lightingAlbedoTexture
		},
	})
	bloomStage := newBloomStage(r.api, r.shaders, r.stageData, BloomStageInput{
		HDRTexture: func() render.Texture {
			return r.lightingAlbedoTexture
		},
	})
	toneMappingStage := newToneMappingStage(r.api, r.shaders, r.stageData, ToneMappingStageInput{
		HDRTexture: func() render.Texture {
			return r.lightingAlbedoTexture
		},
		BloomTexture: opt.V(StageTextureParameter(bloomStage.BloomTexture)),
	})

	// TODO: Make this configurable and user-defined.
	r.stages = []Stage{
		forwardStage,
		exposureProbeStage,
		bloomStage,
		toneMappingStage,
	}

	for _, stage := range r.stages {
		stage.Allocate()
	}
}

func (r *sceneRenderer) Release() {
	defer r.releaseFramebuffers()

	defer r.nearestSampler.Release()
	defer r.linearSampler.Release()
	defer r.depthSampler.Release()

	defer r.shadowDepthTexture.Release()
	defer r.shadowFramebuffer.Release()

	defer r.ambientLightProgram.Release()
	defer r.ambientLightPipeline.Release()

	defer r.pointLightProgram.Release()
	defer r.pointLightPipeline.Release()

	defer r.spotLightProgram.Release()
	defer r.spotLightPipeline.Release()

	defer r.directionalLightProgram.Release()
	defer r.directionalLightPipeline.Release()

	for _, stage := range r.stages {
		defer stage.Release()
	}
}

func (r *sceneRenderer) ResetDebugLines() {
	r.debugLines = r.debugLines[:0]
}

func (r *sceneRenderer) QueueDebugLine(line DebugLine) {
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

	pX := (float64(x-int(viewport.X))/float64(viewport.Width))*2.0 - 1.0
	pY := (float64(int(viewport.Y)-y)/float64(viewport.Height))*2.0 + 1.0

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
	commandBuffer := r.stageData.CommandBuffer()
	uniformBuffer := r.stageData.UniformBuffer()
	uniformBuffer.Reset()

	if viewport.Width != r.framebufferWidth || viewport.Height != r.framebufferHeight {
		r.releaseFramebuffers()
		r.createFramebuffers(viewport.Width, viewport.Height)
	}
	for _, stage := range r.stages {
		stage.PreRender(viewport.Width, viewport.Height)
	}

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

	r.visibleMeshes.Reset()
	scene.dynamicMeshSet.VisitHexahedronRegion(&frustum, r.visibleMeshes)

	r.visibleStaticMeshes.Reset()
	scene.staticMeshOctree.VisitHexahedronRegion(&frustum, r.visibleStaticMeshes)

	r.cameraPlacement = ubo.WriteUniform(uniformBuffer, internal.CameraUniform{
		ProjectionMatrix: projectionMatrix,
		ViewMatrix:       viewMatrix,
		CameraMatrix:     cameraMatrix,
		Viewport: sprec.NewVec4(
			float32(viewport.X),
			float32(viewport.Y),
			float32(viewport.Width),
			float32(viewport.Height),
		),
		Time: scene.Time(),
	})

	ctx := renderCtx{
		framebuffer:    framebuffer,
		scene:          scene,
		x:              viewport.X,
		y:              viewport.Y,
		width:          viewport.Width,
		height:         viewport.Height,
		camera:         camera,
		cameraPosition: stod.Vec3(cameraMatrix.Translation()),
		frustum:        frustum,
	}

	r.renderShadowPass(ctx)
	r.renderGeometryPass(ctx)
	r.renderLightingPass(ctx)

	stageCtx := StageContext{
		Scene:                    ctx.scene,
		Camera:                   ctx.camera,
		CameraPosition:           ctx.cameraPosition,
		CameraPlacement:          r.cameraPlacement,
		VisibleMeshes:            r.visibleMeshes.Items(),
		VisibleStaticMeshIndices: r.visibleStaticMeshes.Items(),
		DebugLines:               r.debugLines,
		Viewport: render.Area{
			X:      ctx.x,
			Y:      ctx.y,
			Width:  ctx.width,
			Height: ctx.height,
		},
		Framebuffer:   ctx.framebuffer,
		CommandBuffer: commandBuffer,
		UniformBuffer: uniformBuffer,
	}
	for _, stage := range r.stages {
		stage.Render(stageCtx)
	}

	uniformSpan := metric.BeginRegion("upload")
	uniformBuffer.Upload()
	uniformSpan.End()

	submitSpan := metric.BeginRegion("submit")
	r.api.Queue().Invalidate()
	r.api.Queue().Submit(commandBuffer)
	submitSpan.End()

	for _, stage := range r.stages {
		stage.PostRender()
	}
}

func (r *sceneRenderer) evaluateProjectionMatrix(camera *Camera, width, height uint32) sprec.Mat4 {
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
	lightMatrix.M14 = sprec.Floor(lightMatrix.M14*shadowMapWidth) / float32(shadowMapWidth)
	lightMatrix.M24 = sprec.Floor(lightMatrix.M24*shadowMapWidth) / float32(shadowMapWidth)
	lightMatrix.M34 = sprec.Floor(lightMatrix.M34*shadowMapWidth) / float32(shadowMapWidth)
	viewMatrix := sprec.InverseMat4(lightMatrix)
	projectionViewMatrix := sprec.Mat4Prod(projectionMatrix, viewMatrix)
	frustum := spatial.ProjectionRegion(stod.Mat4(projectionViewMatrix))

	r.litMeshes.Reset()
	ctx.scene.dynamicMeshSet.VisitHexahedronRegion(&frustum, r.litMeshes)

	r.litStaticMeshes.Reset()
	ctx.scene.staticMeshOctree.VisitHexahedronRegion(&frustum, r.litStaticMeshes)

	r.renderItems = r.renderItems[:0]
	for _, mesh := range r.litMeshes.Items() {
		r.queueMeshRenderItems(mesh, internal.MeshRenderPassTypeShadow)
	}
	for _, meshIndex := range r.litStaticMeshes.Items() {
		staticMesh := &ctx.scene.staticMeshes[meshIndex]
		r.queueStaticMeshRenderItems(ctx, staticMesh, internal.MeshRenderPassTypeShadow)
	}

	commandBuffer := r.stageData.CommandBuffer()
	commandBuffer.BeginRenderPass(render.RenderPassInfo{
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
		StencilLoadOp:   render.LoadOperationLoad,
		StencilStoreOp:  render.StoreOperationDiscard,
	})

	uniformBuffer := r.stageData.UniformBuffer()
	lightCameraPlacement := ubo.WriteUniform(uniformBuffer, internal.CameraUniform{
		ProjectionMatrix: projectionMatrix,
		ViewMatrix:       viewMatrix,
		CameraMatrix:     lightMatrix,
		Viewport:         sprec.ZeroVec4(), // TODO?
		Time:             ctx.scene.Time(), // FIXME?
	})

	meshCtx := renderMeshContext{
		CameraPlacement: lightCameraPlacement,
	}
	r.renderMeshRenderItems(meshCtx, r.renderItems)
	commandBuffer.EndRenderPass()
}

func (r *sceneRenderer) renderGeometryPass(ctx renderCtx) {
	defer metric.BeginRegion("geometry").End()

	r.renderItems = r.renderItems[:0]
	for _, mesh := range r.visibleMeshes.Items() {
		r.queueMeshRenderItems(mesh, internal.MeshRenderPassTypeGeometry)
	}
	for _, meshIndex := range r.visibleStaticMeshes.Items() {
		staticMesh := &ctx.scene.staticMeshes[meshIndex]
		r.queueStaticMeshRenderItems(ctx, staticMesh, internal.MeshRenderPassTypeGeometry)
	}

	commandBuffer := r.stageData.CommandBuffer()
	commandBuffer.BeginRenderPass(render.RenderPassInfo{
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
		StencilLoadOp:   render.LoadOperationLoad,
		StencilStoreOp:  render.StoreOperationDiscard,
		Colors: [4]render.ColorAttachmentInfo{
			{
				LoadOp:     render.LoadOperationClear,
				StoreOp:    render.StoreOperationStore,
				ClearValue: [4]float32{0.0, 0.0, 0.0, 1.0},
			},
			{
				LoadOp:     render.LoadOperationClear,
				StoreOp:    render.StoreOperationStore,
				ClearValue: [4]float32{0.0, 0.0, 1.0, 0.0},
			},
			{
				LoadOp:     render.LoadOperationClear,
				StoreOp:    render.StoreOperationStore,
				ClearValue: [4]float32{0.0, 0.0, 0.0, 1.0},
			},
		},
	})
	meshCtx := renderMeshContext{
		CameraPlacement: r.cameraPlacement,
	}
	r.renderMeshRenderItems(meshCtx, r.renderItems)
	commandBuffer.EndRenderPass()
}

func (r *sceneRenderer) renderLightingPass(ctx renderCtx) {
	defer metric.BeginRegion("lighting").End()

	r.ambientLightBucket.Reset()
	ctx.scene.ambientLightSet.VisitHexahedronRegion(&ctx.frustum, r.ambientLightBucket)

	r.pointLightBucket.Reset()
	ctx.scene.pointLightSet.VisitHexahedronRegion(&ctx.frustum, r.pointLightBucket)

	r.spotLightBucket.Reset()
	ctx.scene.spotLightSet.VisitHexahedronRegion(&ctx.frustum, r.spotLightBucket)

	r.directionalLightBucket.Reset()
	ctx.scene.directionalLightSet.VisitHexahedronRegion(&ctx.frustum, r.directionalLightBucket)

	commandBuffer := r.stageData.CommandBuffer()
	commandBuffer.BeginRenderPass(render.RenderPassInfo{
		Framebuffer: r.lightingFramebuffer,
		Viewport: render.Area{
			X:      0,
			Y:      0,
			Width:  r.framebufferWidth,
			Height: r.framebufferHeight,
		},
		DepthLoadOp:    render.LoadOperationLoad,
		DepthStoreOp:   render.StoreOperationStore,
		StencilLoadOp:  render.LoadOperationLoad,
		StencilStoreOp: render.StoreOperationDiscard,
		Colors: [4]render.ColorAttachmentInfo{
			{
				LoadOp:     render.LoadOperationClear,
				StoreOp:    render.StoreOperationStore,
				ClearValue: [4]float32{0.0, 0.0, 0.0, 1.0},
			},
		},
	})

	// TODO: Use batching (instancing) when rendering lights, if possible.

	for _, ambientLight := range r.ambientLightBucket.Items() {
		if ambientLight.active {
			r.renderAmbientLight(ambientLight)
		}
	}
	for _, pointLight := range r.pointLightBucket.Items() {
		if pointLight.active {
			r.renderPointLight(pointLight)
		}
	}
	for _, spotLight := range r.spotLightBucket.Items() {
		if spotLight.active {
			r.renderSpotLight(spotLight)
		}
	}
	for _, directionalLight := range r.directionalLightBucket.Items() {
		if directionalLight.active {
			r.renderDirectionalLight(directionalLight)
		}
	}

	commandBuffer.EndRenderPass()
}

func (r *sceneRenderer) queueMeshRenderItems(mesh *Mesh, passType internal.MeshRenderPassType) {
	if !mesh.active {
		return
	}
	definition := mesh.definition
	passes := definition.passesByType[passType]
	for _, pass := range passes {
		r.renderItems = append(r.renderItems, renderItem{
			Layer:       pass.Layer,
			MaterialKey: pass.Key,
			ArmatureKey: mesh.armature.key(),

			Pipeline:     pass.Pipeline,
			TextureSet:   pass.TextureSet,
			UniformSet:   pass.UniformSet,
			ModelData:    mesh.matrixData,
			ArmatureData: mesh.armature.uniformData(),

			IndexByteOffset: pass.IndexByteOffset,
			IndexCount:      pass.IndexCount,
		})
	}
}

func (r *sceneRenderer) queueStaticMeshRenderItems(ctx renderCtx, mesh *StaticMesh, passType internal.MeshRenderPassType) {
	if !mesh.active {
		return
	}
	distance := dprec.Vec3Diff(mesh.position, ctx.cameraPosition).Length()
	if distance < mesh.minDistance || mesh.maxDistance < distance {
		return
	}

	// TODO: Extract common stuff between mesh and static mesh into a type
	// that is passed ot this function instead so that it can be reused.
	definition := mesh.definition
	passes := definition.passesByType[passType]
	for _, pass := range passes {
		r.renderItems = append(r.renderItems, renderItem{
			Layer:       pass.Layer,
			MaterialKey: pass.Key,
			ArmatureKey: math.MaxUint32,

			Pipeline:     pass.Pipeline,
			TextureSet:   pass.TextureSet,
			UniformSet:   pass.UniformSet,
			ModelData:    mesh.matrixData,
			ArmatureData: nil,

			IndexByteOffset: pass.IndexByteOffset,
			IndexCount:      pass.IndexCount,
		})
	}
}

func (r *sceneRenderer) renderMeshRenderItems(ctx renderMeshContext, items []renderItem) {
	const maxBatchSize = modelUniformBufferItemCount
	var (
		lastMaterialKey = uint32(math.MaxUint32)
		lastArmatureKey = uint32(math.MaxUint32)

		batchStart = 0
		batchEnd   = 0
	)

	slices.SortFunc(items, compareMeshRenderItems)

	itemCount := len(items)
	for i, item := range items {
		materialKey := item.MaterialKey
		armatureKey := item.ArmatureKey

		isSame := (materialKey == lastMaterialKey) && (armatureKey == lastArmatureKey)
		if !isSame {
			if batchStart < batchEnd {
				r.renderMeshRenderItemBatch(ctx, items[batchStart:batchEnd])
			}
			batchStart = batchEnd
		}
		batchEnd++

		batchSize := batchEnd - batchStart
		if (batchSize >= maxBatchSize) || (i == itemCount-1) {
			r.renderMeshRenderItemBatch(ctx, items[batchStart:batchEnd])
			batchStart = batchEnd
		}

		lastMaterialKey = materialKey
		lastArmatureKey = armatureKey
	}
}

func (r *sceneRenderer) renderMeshRenderItemBatch(ctx renderMeshContext, items []renderItem) {
	template := items[0]

	commandBuffer := r.stageData.CommandBuffer()
	commandBuffer.BindPipeline(template.Pipeline)

	// Camera data is shared between all items.
	cameraPlacement := ctx.CameraPlacement
	commandBuffer.UniformBufferUnit(
		internal.UniformBufferBindingCamera,
		cameraPlacement.Buffer,
		cameraPlacement.Offset,
		cameraPlacement.Size,
	)

	// Material data is shared between all items.
	uniformBuffer := r.stageData.UniformBuffer()
	if !template.UniformSet.IsEmpty() {
		materialPlacement := ubo.WriteUniform(uniformBuffer, internal.MaterialUniform{
			Data: template.UniformSet.Data(),
		})
		commandBuffer.UniformBufferUnit(
			internal.UniformBufferBindingMaterial,
			materialPlacement.Buffer,
			materialPlacement.Offset,
			materialPlacement.Size,
		)
	}

	for i := range template.TextureSet.TextureCount() {
		if texture := template.TextureSet.TextureAt(i); texture != nil {
			commandBuffer.TextureUnit(uint(i), texture)
		}
		if sampler := template.TextureSet.SamplerAt(i); sampler != nil {
			commandBuffer.SamplerUnit(uint(i), sampler)
		}
	}

	// Model data needs to be combined.
	for i, item := range items {
		start := i * modelUniformBufferItemSize
		end := start + modelUniformBufferItemSize
		copy(r.modelUniformBufferData[start:end], item.ModelData)
	}
	modelPlacement := ubo.WriteUniform(uniformBuffer, internal.ModelUniform{
		ModelMatrices: r.modelUniformBufferData,
	})
	commandBuffer.UniformBufferUnit(
		internal.UniformBufferBindingModel,
		modelPlacement.Buffer,
		modelPlacement.Offset,
		modelPlacement.Size,
	)

	// Armature data is shared between all items.
	if template.ArmatureData != nil {
		armaturePlacement := ubo.WriteUniform(uniformBuffer, internal.ArmatureUniform{
			BoneMatrices: template.ArmatureData,
		})
		commandBuffer.UniformBufferUnit(
			internal.UniformBufferBindingArmature,
			armaturePlacement.Buffer,
			armaturePlacement.Offset,
			armaturePlacement.Size,
		)
	}

	commandBuffer.DrawIndexed(template.IndexByteOffset, template.IndexCount, uint32(len(items)))
}

func (r *sceneRenderer) renderAmbientLight(light *AmbientLight) {
	quadShape := r.stageData.QuadShape()
	commandBuffer := r.stageData.CommandBuffer()
	// TODO: Ambient light intensity based on distance and inner and outer radius
	commandBuffer.BindPipeline(r.ambientLightPipeline)
	commandBuffer.TextureUnit(internal.TextureBindingLightingFramebufferColor0, r.geometryAlbedoTexture)
	commandBuffer.SamplerUnit(internal.TextureBindingLightingFramebufferColor0, r.nearestSampler)
	commandBuffer.TextureUnit(internal.TextureBindingLightingFramebufferColor1, r.geometryNormalTexture)
	commandBuffer.SamplerUnit(internal.TextureBindingLightingFramebufferColor1, r.nearestSampler)
	commandBuffer.TextureUnit(internal.TextureBindingLightingFramebufferDepth, r.geometryDepthTexture)
	commandBuffer.SamplerUnit(internal.TextureBindingLightingFramebufferDepth, r.nearestSampler)
	commandBuffer.TextureUnit(internal.TextureBindingLightingReflectionTexture, light.reflectionTexture)
	commandBuffer.SamplerUnit(internal.TextureBindingLightingReflectionTexture, r.linearSampler)
	commandBuffer.TextureUnit(internal.TextureBindingLightingRefractionTexture, light.refractionTexture)
	commandBuffer.SamplerUnit(internal.TextureBindingLightingRefractionTexture, r.linearSampler)
	commandBuffer.UniformBufferUnit(
		internal.UniformBufferBindingCamera,
		r.cameraPlacement.Buffer,
		r.cameraPlacement.Offset,
		r.cameraPlacement.Size,
	)
	commandBuffer.DrawIndexed(0, quadShape.IndexCount(), 1)
}

func (r *sceneRenderer) renderPointLight(light *PointLight) {
	sphereShape := r.stageData.SphereShape()
	projectionMatrix := sprec.IdentityMat4()
	lightMatrix := light.gfxMatrix()
	viewMatrix := sprec.InverseMat4(lightMatrix)

	uniformBuffer := r.stageData.UniformBuffer()
	lightPlacement := ubo.WriteUniform(uniformBuffer, internal.LightUniform{
		ProjectionMatrix: projectionMatrix,
		ViewMatrix:       viewMatrix,
		LightMatrix:      lightMatrix,
	})

	lightPropertiesPlacement := ubo.WriteUniform(uniformBuffer, internal.LightPropertiesUniform{
		Color:     dtos.Vec3(light.emitColor),
		Intensity: 1.0,
		Range:     float32(light.emitRange),
	})

	commandBuffer := r.stageData.CommandBuffer()
	commandBuffer.BindPipeline(r.pointLightPipeline)
	commandBuffer.TextureUnit(internal.TextureBindingLightingFramebufferColor0, r.geometryAlbedoTexture)
	commandBuffer.SamplerUnit(internal.TextureBindingLightingFramebufferColor0, r.nearestSampler)
	commandBuffer.TextureUnit(internal.TextureBindingLightingFramebufferColor1, r.geometryNormalTexture)
	commandBuffer.SamplerUnit(internal.TextureBindingLightingFramebufferColor1, r.nearestSampler)
	commandBuffer.TextureUnit(internal.TextureBindingLightingFramebufferDepth, r.geometryDepthTexture)
	commandBuffer.SamplerUnit(internal.TextureBindingLightingFramebufferDepth, r.nearestSampler)
	commandBuffer.UniformBufferUnit(
		internal.UniformBufferBindingCamera,
		r.cameraPlacement.Buffer,
		r.cameraPlacement.Offset,
		r.cameraPlacement.Size,
	)
	commandBuffer.UniformBufferUnit(
		internal.UniformBufferBindingLight,
		lightPlacement.Buffer,
		lightPlacement.Offset,
		lightPlacement.Size,
	)
	commandBuffer.UniformBufferUnit(
		internal.UniformBufferBindingLightProperties,
		lightPropertiesPlacement.Buffer,
		lightPropertiesPlacement.Offset,
		lightPropertiesPlacement.Size,
	)
	commandBuffer.DrawIndexed(0, sphereShape.IndexCount(), 1)
}

func (r *sceneRenderer) renderSpotLight(light *SpotLight) {
	coneShape := r.stageData.ConeShape()
	projectionMatrix := sprec.IdentityMat4()
	lightMatrix := light.gfxMatrix()
	viewMatrix := sprec.InverseMat4(lightMatrix)

	uniformBuffer := r.stageData.UniformBuffer()
	lightPlacement := ubo.WriteUniform(uniformBuffer, internal.LightUniform{
		ProjectionMatrix: projectionMatrix,
		ViewMatrix:       viewMatrix,
		LightMatrix:      lightMatrix,
	})

	lightPropertiesPlacement := ubo.WriteUniform(uniformBuffer, internal.LightPropertiesUniform{
		Color:      dtos.Vec3(light.emitColor),
		Intensity:  1.0,
		Range:      float32(light.emitRange),
		OuterAngle: float32(light.emitOuterConeAngle.Radians()),
		InnerAngle: float32(light.emitInnerConeAngle.Radians()),
	})

	commandBuffer := r.stageData.CommandBuffer()
	commandBuffer.BindPipeline(r.spotLightPipeline)
	commandBuffer.TextureUnit(internal.TextureBindingLightingFramebufferColor0, r.geometryAlbedoTexture)
	commandBuffer.SamplerUnit(internal.TextureBindingLightingFramebufferColor0, r.nearestSampler)
	commandBuffer.TextureUnit(internal.TextureBindingLightingFramebufferColor1, r.geometryNormalTexture)
	commandBuffer.SamplerUnit(internal.TextureBindingLightingFramebufferColor1, r.nearestSampler)
	commandBuffer.TextureUnit(internal.TextureBindingLightingFramebufferDepth, r.geometryDepthTexture)
	commandBuffer.SamplerUnit(internal.TextureBindingLightingFramebufferDepth, r.nearestSampler)
	commandBuffer.UniformBufferUnit(
		internal.UniformBufferBindingCamera,
		r.cameraPlacement.Buffer,
		r.cameraPlacement.Offset,
		r.cameraPlacement.Size,
	)
	commandBuffer.UniformBufferUnit(
		internal.UniformBufferBindingLight,
		lightPlacement.Buffer,
		lightPlacement.Offset,
		lightPlacement.Size,
	)
	commandBuffer.UniformBufferUnit(
		internal.UniformBufferBindingLightProperties,
		lightPropertiesPlacement.Buffer,
		lightPropertiesPlacement.Offset,
		lightPropertiesPlacement.Size,
	)
	commandBuffer.DrawIndexed(0, coneShape.IndexCount(), 1)
}

func (r *sceneRenderer) renderDirectionalLight(light *DirectionalLight) {
	quadShape := r.stageData.QuadShape()
	projectionMatrix := lightOrtho()
	lightMatrix := light.gfxMatrix()
	lightMatrix.M14 = sprec.Floor(lightMatrix.M14*shadowMapWidth) / float32(shadowMapWidth)
	lightMatrix.M24 = sprec.Floor(lightMatrix.M24*shadowMapWidth) / float32(shadowMapWidth)
	lightMatrix.M34 = sprec.Floor(lightMatrix.M34*shadowMapWidth) / float32(shadowMapWidth)
	viewMatrix := sprec.InverseMat4(lightMatrix)

	uniformBuffer := r.stageData.UniformBuffer()
	lightPlacement := ubo.WriteUniform(uniformBuffer, internal.LightUniform{
		ProjectionMatrix: projectionMatrix,
		ViewMatrix:       viewMatrix,
		LightMatrix:      lightMatrix,
	})

	lightPropertiesPlacement := ubo.WriteUniform(uniformBuffer, internal.LightPropertiesUniform{
		Color:     dtos.Vec3(light.emitColor),
		Intensity: 1.0,
	})

	commandBuffer := r.stageData.CommandBuffer()
	commandBuffer.BindPipeline(r.directionalLightPipeline)
	commandBuffer.TextureUnit(internal.TextureBindingLightingFramebufferColor0, r.geometryAlbedoTexture)
	commandBuffer.SamplerUnit(internal.TextureBindingLightingFramebufferColor0, r.nearestSampler)
	commandBuffer.TextureUnit(internal.TextureBindingLightingFramebufferColor1, r.geometryNormalTexture)
	commandBuffer.SamplerUnit(internal.TextureBindingLightingFramebufferColor1, r.nearestSampler)
	commandBuffer.TextureUnit(internal.TextureBindingLightingFramebufferDepth, r.geometryDepthTexture)
	commandBuffer.SamplerUnit(internal.TextureBindingLightingFramebufferDepth, r.nearestSampler)
	commandBuffer.TextureUnit(internal.TextureBindingShadowFramebufferDepth, r.shadowDepthTexture)
	commandBuffer.SamplerUnit(internal.TextureBindingShadowFramebufferDepth, r.depthSampler)
	commandBuffer.UniformBufferUnit(
		internal.UniformBufferBindingCamera,
		r.cameraPlacement.Buffer,
		r.cameraPlacement.Offset,
		r.cameraPlacement.Size,
	)
	commandBuffer.UniformBufferUnit(
		internal.UniformBufferBindingLight,
		lightPlacement.Buffer,
		lightPlacement.Offset,
		lightPlacement.Size,
	)
	commandBuffer.UniformBufferUnit(
		internal.UniformBufferBindingLightProperties,
		lightPropertiesPlacement.Buffer,
		lightPropertiesPlacement.Offset,
		lightPropertiesPlacement.Size,
	)
	commandBuffer.DrawIndexed(0, quadShape.IndexCount(), 1)
}

type renderCtx struct {
	framebuffer    render.Framebuffer
	scene          *Scene
	x              uint32
	y              uint32
	width          uint32
	height         uint32
	camera         *Camera
	cameraPosition dprec.Vec3
	frustum        spatial.HexahedronRegion
}

type renderMeshContext struct {
	CameraPlacement ubo.UniformPlacement
}

// TODO: Rename to meshRenderItem
type renderItem struct {
	Layer       int32
	MaterialKey uint32
	ArmatureKey uint32

	Pipeline render.Pipeline

	TextureSet internal.TextureSet
	UniformSet internal.UniformSet

	ModelData    []byte
	ArmatureData []byte

	IndexByteOffset uint32
	IndexCount      uint32
}

func compareMeshRenderItems(a, b renderItem) int {
	return cmp.Or(
		cmp.Compare(a.Layer, b.Layer),
		cmp.Compare(a.MaterialKey, b.MaterialKey),
		cmp.Compare(a.ArmatureKey, b.ArmatureKey),
	)
}
