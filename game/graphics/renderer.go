package graphics

import (
	"fmt"

	"github.com/mokiat/gblob"
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
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

func newRenderer(api render.API, shaders ShaderCollection, stageData *commonStageData, meshRenderer *meshRenderer) *sceneRenderer {
	return &sceneRenderer{
		api:          api,
		shaders:      shaders,
		stageData:    stageData,
		meshRenderer: meshRenderer,

		visibleStaticMeshes: spatial.NewVisitorBucket[uint32](2_000),
		visibleMeshes:       spatial.NewVisitorBucket[*Mesh](2_000),

		litStaticMeshes: spatial.NewVisitorBucket[uint32](2_000),
		litMeshes:       spatial.NewVisitorBucket[*Mesh](2_000),

		debugLines: make([]DebugLine, debugMaxLineCount),
	}
}

type sceneRenderer struct {
	api          render.API
	shaders      ShaderCollection
	stageData    *commonStageData
	meshRenderer *meshRenderer

	framebufferWidth  uint32
	framebufferHeight uint32

	// TODO: Create dedicated Source stages for these.
	geometryAlbedoTexture render.Texture
	geometryNormalTexture render.Texture
	geometryDepthTexture  render.Texture

	lightingAlbedoTexture render.Texture

	shadowDepthTexture render.Texture
	shadowFramebuffer  render.Framebuffer

	debugLines []DebugLine

	visibleStaticMeshes *spatial.VisitorBucket[uint32]
	visibleMeshes       *spatial.VisitorBucket[*Mesh]

	litStaticMeshes *spatial.VisitorBucket[uint32]
	litMeshes       *spatial.VisitorBucket[*Mesh]

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

	r.lightingAlbedoTexture = r.api.CreateColorTexture2D(render.ColorTexture2DInfo{
		Width:           r.framebufferWidth,
		Height:          r.framebufferHeight,
		GenerateMipmaps: false,
		GammaCorrection: false,
		Format:          render.DataFormatRGBA16F,
	})
}

func (r *sceneRenderer) releaseFramebuffers() {
	defer r.geometryAlbedoTexture.Release()
	defer r.geometryNormalTexture.Release()
	defer r.geometryDepthTexture.Release()

	defer r.lightingAlbedoTexture.Release()
}

func (r *sceneRenderer) Allocate() {
	r.createFramebuffers(800, 600)

	r.shadowDepthTexture = r.api.CreateDepthTexture2D(render.DepthTexture2DInfo{
		Width:      shadowMapWidth,
		Height:     shadowMapHeight,
		Comparable: true,
	})
	r.shadowFramebuffer = r.api.CreateFramebuffer(render.FramebufferInfo{
		DepthAttachment: r.shadowDepthTexture,
	})

	r.modelUniformBufferData = make([]byte, modelUniformBufferSize)

	geometryStage := newGeometryStage(r.api, r.meshRenderer, GeometryStageInput{
		AlbedoMetallicTexture: func() render.Texture {
			return r.geometryAlbedoTexture
		},
		NormalRoughnessTexture: func() render.Texture {
			return r.geometryNormalTexture
		},
		DepthTexture: func() render.Texture {
			return r.geometryDepthTexture
		},
	})
	lightingStage := newLightingStage(r.api, r.shaders, r.stageData, LightingStageInput{
		AlbedoMetallicTexture: func() render.Texture {
			return r.geometryAlbedoTexture
		},
		NormalRoughnessTexture: func() render.Texture {
			return r.geometryNormalTexture
		},
		DepthTexture: func() render.Texture {
			return r.geometryDepthTexture
		},
		HDRTexture: func() render.Texture {
			return r.lightingAlbedoTexture
		},
	})
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
		geometryStage,
		lightingStage,
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

	defer r.shadowDepthTexture.Release()
	defer r.shadowFramebuffer.Release()

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

	// ctx := renderCtx{
	// 	framebuffer:    framebuffer,
	// 	scene:          scene,
	// 	x:              viewport.X,
	// 	y:              viewport.Y,
	// 	width:          viewport.Width,
	// 	height:         viewport.Height,
	// 	camera:         camera,
	// 	cameraPosition: stod.Vec3(cameraMatrix.Translation()),
	// 	frustum:        frustum,
	// }

	stageCtx := StageContext{
		Scene:                    scene,
		Camera:                   camera,
		CameraPosition:           stod.Vec3(cameraMatrix.Translation()),
		CameraPlacement:          r.cameraPlacement,
		CameraFrustum:            frustum,
		VisibleMeshes:            r.visibleMeshes.Items(),
		VisibleStaticMeshIndices: r.visibleStaticMeshes.Items(),
		DebugLines:               r.debugLines,
		Viewport:                 render.Area(viewport),
		Framebuffer:              framebuffer,
		CommandBuffer:            commandBuffer,
		UniformBuffer:            uniformBuffer,
	}

	// r.renderShadowPass(ctx, stageCtx)

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

// func (r *sceneRenderer) renderShadowPass(ctx renderCtx, stageCtx StageContext) {
// 	defer metric.BeginRegion("shadow").End()

// 	r.directionalLightBucket.Reset()
// 	ctx.scene.directionalLightSet.VisitHexahedronRegion(&ctx.frustum, r.directionalLightBucket)

// 	var directionalLight *DirectionalLight
// 	for _, light := range r.directionalLightBucket.Items() {
// 		if light.active {
// 			directionalLight = light
// 			break
// 		}
// 	}
// 	if directionalLight == nil {
// 		return
// 	}

// 	projectionMatrix := lightOrtho()
// 	lightMatrix := directionalLight.gfxMatrix()
// 	lightMatrix.M14 = sprec.Floor(lightMatrix.M14*shadowMapWidth) / float32(shadowMapWidth)
// 	lightMatrix.M24 = sprec.Floor(lightMatrix.M24*shadowMapWidth) / float32(shadowMapWidth)
// 	lightMatrix.M34 = sprec.Floor(lightMatrix.M34*shadowMapWidth) / float32(shadowMapWidth)
// 	viewMatrix := sprec.InverseMat4(lightMatrix)
// 	projectionViewMatrix := sprec.Mat4Prod(projectionMatrix, viewMatrix)
// 	frustum := spatial.ProjectionRegion(stod.Mat4(projectionViewMatrix))

// 	r.litMeshes.Reset()
// 	ctx.scene.dynamicMeshSet.VisitHexahedronRegion(&frustum, r.litMeshes)

// 	r.litStaticMeshes.Reset()
// 	ctx.scene.staticMeshOctree.VisitHexahedronRegion(&frustum, r.litStaticMeshes)

// 	r.meshRenderer.DiscardRenderItems()
// 	for _, mesh := range r.litMeshes.Items() {
// 		r.meshRenderer.QueueMeshRenderItems(stageCtx, mesh, internal.MeshRenderPassTypeShadow)
// 	}
// 	for _, meshIndex := range r.litStaticMeshes.Items() {
// 		staticMesh := &ctx.scene.staticMeshes[meshIndex]
// 		r.meshRenderer.QueueStaticMeshRenderItems(stageCtx, staticMesh, internal.MeshRenderPassTypeShadow)
// 	}

// 	commandBuffer := r.stageData.CommandBuffer()
// 	commandBuffer.BeginRenderPass(render.RenderPassInfo{
// 		Framebuffer: r.shadowFramebuffer,
// 		Viewport: render.Area{
// 			X:      0,
// 			Y:      0,
// 			Width:  shadowMapWidth,
// 			Height: shadowMapHeight,
// 		},
// 		DepthLoadOp:     render.LoadOperationClear,
// 		DepthStoreOp:    render.StoreOperationStore,
// 		DepthClearValue: 1.0,
// 		StencilLoadOp:   render.LoadOperationLoad,
// 		StencilStoreOp:  render.StoreOperationDiscard,
// 	})

// 	uniformBuffer := r.stageData.UniformBuffer()
// 	lightCameraPlacement := ubo.WriteUniform(uniformBuffer, internal.CameraUniform{
// 		ProjectionMatrix: projectionMatrix,
// 		ViewMatrix:       viewMatrix,
// 		CameraMatrix:     lightMatrix,
// 		Viewport:         sprec.ZeroVec4(), // TODO?
// 		Time:             ctx.scene.Time(), // FIXME?
// 	})
// 	stageCtx.CameraPlacement = lightCameraPlacement
// 	r.meshRenderer.Render(stageCtx)
// 	commandBuffer.EndRenderPass()
// }

// type renderCtx struct {
// 	framebuffer    render.Framebuffer
// 	scene          *Scene
// 	x              uint32
// 	y              uint32
// 	width          uint32
// 	height         uint32
// 	camera         *Camera
// 	cameraPosition dprec.Vec3
// 	frustum        spatial.HexahedronRegion
// }

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
