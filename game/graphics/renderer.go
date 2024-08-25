package graphics

import (
	"fmt"

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
	shadowMapWidth  = 2048 // FIXME: Should not use this. Instead from the texture.
	shadowMapHeight = 2048 // FIXME: Should not use this. Instead from the texture.

	commandBufferSize = 2 * 1024 * 1024  // 2MB
	uniformBufferSize = 32 * 1024 * 1024 // 32MB
)

func newRenderer(api render.API, stageData *commonStageData, stages []Stage) *sceneRenderer {
	return &sceneRenderer{
		api:       api,
		stageData: stageData,
		stages:    stages,

		debugLines: make([]DebugLine, debugMaxLineCount),

		visibleAmbientLights:     spatial.NewVisitorBucket[*AmbientLight](1),
		visiblePointLights:       spatial.NewVisitorBucket[*PointLight](32),
		visibleSpotLights:        spatial.NewVisitorBucket[*SpotLight](8),
		visibleDirectionalLights: spatial.NewVisitorBucket[*DirectionalLight](2),

		visibleStaticMeshes: spatial.NewVisitorBucket[uint32](65536),
		visibleMeshes:       spatial.NewVisitorBucket[*Mesh](1024),
	}
}

type sceneRenderer struct {
	api       render.API
	stageData *commonStageData
	stages    []Stage

	debugLines []DebugLine

	visibleAmbientLights     *spatial.VisitorBucket[*AmbientLight]
	visiblePointLights       *spatial.VisitorBucket[*PointLight]
	visibleSpotLights        *spatial.VisitorBucket[*SpotLight]
	visibleDirectionalLights *spatial.VisitorBucket[*DirectionalLight]

	visibleStaticMeshes *spatial.VisitorBucket[uint32]
	visibleMeshes       *spatial.VisitorBucket[*Mesh]
}

func (r *sceneRenderer) Allocate() {
	for _, stage := range r.stages {
		stage.Allocate()
	}
}

func (r *sceneRenderer) Release() {
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

	for _, stage := range r.stages {
		stage.PreRender(viewport.Width, viewport.Height)
	}

	projectionMatrix := r.evaluateProjectionMatrix(camera, viewport.Width, viewport.Height)
	cameraMatrix := camera.gfxMatrix()
	viewMatrix := sprec.InverseMat4(cameraMatrix)
	projectionViewMatrix := sprec.Mat4Prod(projectionMatrix, viewMatrix)
	frustum := spatial.ProjectionRegion(stod.Mat4(projectionViewMatrix))

	cameraPlacement := ubo.WriteUniform(uniformBuffer, internal.CameraUniform{
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

	r.visibleAmbientLights.Reset()
	scene.ambientLightSet.VisitHexahedronRegion(&frustum, r.visibleAmbientLights)

	r.visiblePointLights.Reset()
	scene.pointLightSet.VisitHexahedronRegion(&frustum, r.visiblePointLights)

	r.visibleSpotLights.Reset()
	scene.spotLightSet.VisitHexahedronRegion(&frustum, r.visibleSpotLights)

	r.visibleDirectionalLights.Reset()
	scene.directionalLightSet.VisitHexahedronRegion(&frustum, r.visibleDirectionalLights)

	r.visibleMeshes.Reset()
	scene.dynamicMeshSet.VisitHexahedronRegion(&frustum, r.visibleMeshes)

	r.visibleStaticMeshes.Reset()
	scene.staticMeshOctree.VisitHexahedronRegion(&frustum, r.visibleStaticMeshes)

	stageCtx := StageContext{
		Scene:                    scene,
		Camera:                   camera,
		CameraPosition:           stod.Vec3(cameraMatrix.Translation()),
		CameraPlacement:          cameraPlacement,
		CameraFrustum:            frustum,
		VisibleAmbientLights:     r.visibleAmbientLights.Items(),
		VisiblePointLights:       r.visiblePointLights.Items(),
		VisibleSpotLights:        r.visibleSpotLights.Items(),
		VisibleDirectionalLights: r.visibleDirectionalLights.Items(),
		VisibleMeshes:            r.visibleMeshes.Items(),
		VisibleStaticMeshIndices: r.visibleStaticMeshes.Items(),
		DebugLines:               r.debugLines,
		Viewport:                 render.Area(viewport),
		Framebuffer:              framebuffer,
		CommandBuffer:            commandBuffer,
		UniformBuffer:            uniformBuffer,
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
	// FIXME: This should depend on the light and cascade.
	return sprec.OrthoMat4(-32, 32, 32, -32, 0, 256)
}

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
