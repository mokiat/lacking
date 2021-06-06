package graphics

import (
	"fmt"

	"github.com/go-gl/gl/v4.6-core/gl"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/framework/opengl"
	"github.com/mokiat/lacking/game/graphics"
)

const (
	framebufferWidth  = int32(1920)
	framebufferHeight = int32(1080)
)

func newRenderer() *Renderer {
	return &Renderer{
		framebufferWidth:  framebufferWidth,
		framebufferHeight: framebufferHeight,

		geometryAlbedoTexture: opengl.NewTwoDTexture(),
		geometryNormalTexture: opengl.NewTwoDTexture(),
		geometryDepthTexture:  opengl.NewTwoDTexture(),
		geometryFramebuffer:   opengl.NewFramebuffer(),

		lightingAlbedoTexture: opengl.NewTwoDTexture(),
		lightingFramebuffer:   opengl.NewFramebuffer(),

		screenFramebuffer: opengl.DefaultFramebuffer(),

		postprocessingMaterial: newPostprocessingMaterial(),

		quadMesh: newQuadMesh(),

		skyboxMaterial: newSkyboxMaterial(),
		skyboxMesh:     newSkyboxMesh(),
	}
}

type Renderer struct {
	framebufferWidth  int32
	framebufferHeight int32

	geometryAlbedoTexture *opengl.TwoDTexture
	geometryNormalTexture *opengl.TwoDTexture
	geometryDepthTexture  *opengl.TwoDTexture
	geometryFramebuffer   *opengl.Framebuffer

	lightingAlbedoTexture *opengl.TwoDTexture
	lightingFramebuffer   *opengl.Framebuffer

	screenFramebuffer *opengl.Framebuffer

	postprocessingMaterial *PostprocessingMaterial

	quadMesh *QuadMesh

	skyboxMaterial *SkyboxMaterial
	skyboxMesh     *SkyboxMesh
}

func (r *Renderer) Allocate() {
	geometryAlbedoTextureInfo := opengl.TwoDTextureAllocateInfo{
		Width:             framebufferWidth,
		Height:            framebufferHeight,
		MinFilter:         gl.NEAREST,
		MagFilter:         gl.NEAREST,
		InternalFormat:    gl.RGBA8,
		DataFormat:        gl.RGBA,
		DataComponentType: gl.UNSIGNED_BYTE,
	}
	r.geometryAlbedoTexture.Allocate(geometryAlbedoTextureInfo)

	geometryNormalTextureInfo := opengl.TwoDTextureAllocateInfo{
		Width:             framebufferWidth,
		Height:            framebufferHeight,
		MinFilter:         gl.NEAREST,
		MagFilter:         gl.NEAREST,
		InternalFormat:    gl.RGBA32F,
		DataFormat:        gl.RGBA,
		DataComponentType: gl.FLOAT,
	}
	r.geometryNormalTexture.Allocate(geometryNormalTextureInfo)

	geometryDepthTextureInfo := opengl.TwoDTextureAllocateInfo{
		Width:             framebufferWidth,
		Height:            framebufferHeight,
		MinFilter:         gl.NEAREST,
		MagFilter:         gl.NEAREST,
		InternalFormat:    gl.DEPTH_COMPONENT32,
		DataFormat:        gl.DEPTH_COMPONENT,
		DataComponentType: gl.FLOAT,
	}
	r.geometryDepthTexture.Allocate(geometryDepthTextureInfo)

	geometryFramebufferInfo := opengl.FramebufferAllocateInfo{
		ColorAttachments: []*opengl.Texture{
			&r.geometryAlbedoTexture.Texture,
			&r.geometryNormalTexture.Texture,
		},
		DepthAttachment: &r.geometryDepthTexture.Texture,
	}
	r.geometryFramebuffer.Allocate(geometryFramebufferInfo)

	lightingAlbedoTextureInfo := opengl.TwoDTextureAllocateInfo{
		Width:             framebufferWidth,
		Height:            framebufferHeight,
		MinFilter:         gl.NEAREST,
		MagFilter:         gl.NEAREST,
		InternalFormat:    gl.RGBA32F,
		DataFormat:        gl.RGBA,
		DataComponentType: gl.FLOAT,
	}
	r.lightingAlbedoTexture.Allocate(lightingAlbedoTextureInfo)

	lightingFramebufferInfo := opengl.FramebufferAllocateInfo{
		ColorAttachments: []*opengl.Texture{
			&r.lightingAlbedoTexture.Texture,
		},
		DepthAttachment: &r.geometryDepthTexture.Texture,
	}
	r.lightingFramebuffer.Allocate(lightingFramebufferInfo)

	r.postprocessingMaterial.Allocate(ReinhardToneMapping)

	r.quadMesh.Allocate()

	r.skyboxMesh.Allocate()

	r.skyboxMaterial.Allocate()
}

func (r *Renderer) Release() {
	r.skyboxMaterial.Release()
	r.skyboxMesh.Release()

	r.quadMesh.Release()

	r.postprocessingMaterial.Release()

	r.lightingFramebuffer.Release()
	r.lightingAlbedoTexture.Release()

	r.geometryFramebuffer.Release()
	r.geometryDepthTexture.Release()
	r.geometryNormalTexture.Release()
	r.geometryAlbedoTexture.Release()
}

type renderCtx struct {
	scene            *Scene
	x                int
	y                int
	width            int
	height           int
	projectionMatrix [16]float32
	cameraMatrix     [16]float32
	viewMatrix       [16]float32
	camera           *Camera
}

func (r *Renderer) Render(viewport graphics.Viewport, scene *Scene, camera *Camera) {
	projectionMatrix := r.evaluateProjectionMatrix(camera, viewport.Width, viewport.Height)
	cameraMatrix := r.evaluateCameraMatrix(camera)
	viewMatrix := sprec.InverseMat4(cameraMatrix)

	gl.Enable(gl.FRAMEBUFFER_SRGB)

	ctx := renderCtx{
		scene:            scene,
		x:                viewport.X,
		y:                viewport.Y,
		width:            viewport.Width,
		height:           viewport.Height,
		projectionMatrix: projectionMatrix.ColumnMajorArray(),
		cameraMatrix:     cameraMatrix.ColumnMajorArray(),
		viewMatrix:       viewMatrix.ColumnMajorArray(),
		camera:           camera,
	}
	r.renderGeometryPass(ctx)
	r.renderForwardPass(ctx)
	r.renderPostprocessingPass(ctx)
}

func (r *Renderer) evaluateProjectionMatrix(camera *Camera, width, height int) sprec.Mat4 {
	const (
		near = float32(0.5)
		far  = float32(900.0)
	)
	var (
		fWidth  = sprec.Max(1.0, float32(width))
		fHeight = sprec.Max(1.0, float32(height))
	)

	switch camera.fovMode {
	case graphics.FoVModeHorizontalPlus:
		halfHeight := near * sprec.Tan(camera.fov/2.0)
		halfWidth := halfHeight * (fWidth / fHeight)
		return sprec.PerspectiveMat4(
			-halfWidth, halfWidth, -halfHeight, halfHeight, near, far,
		)

	case graphics.FoVModeVertialMinus:
		halfWidth := near * sprec.Tan(camera.fov/2.0)
		halfHeight := halfWidth * (fHeight / fWidth)
		return sprec.PerspectiveMat4(
			-halfWidth, halfWidth, -halfHeight, halfHeight, near, far,
		)

	case graphics.FoVModePixelBased:
		halfWidth := fWidth / 2.0
		halfHeight := fHeight / 2.0
		return sprec.OrthoMat4(
			-halfWidth, halfWidth, halfHeight, -halfHeight, near, far,
		)

	default:
		panic(fmt.Errorf("unsupported fov mode: %s", camera.fovMode))
	}
}

func (r *Renderer) evaluateCameraMatrix(camera *Camera) sprec.Mat4 {
	camPosition := camera.Position()
	camRotation := camera.Rotation()
	camScale := camera.Scale()

	return sprec.Mat4MultiProd(
		sprec.TranslationMat4(
			camPosition.X,
			camPosition.Y,
			camPosition.Z,
		),
		sprec.OrientationMat4(
			camRotation.OrientationX(),
			camRotation.OrientationY(),
			camRotation.OrientationZ(),
		),
		sprec.ScaleMat4(
			camScale.X,
			camScale.Y,
			camScale.Z,
		),
	)
}

func (r *Renderer) renderGeometryPass(ctx renderCtx) {
	r.geometryFramebuffer.Use()
	r.geometryFramebuffer.ClearColor(0, sprec.NewVec4(
		ctx.scene.sky.backgroundColor.X,
		ctx.scene.sky.backgroundColor.Y,
		ctx.scene.sky.backgroundColor.Z,
		1.0,
	))
	r.geometryFramebuffer.ClearDepth(1.0)
	gl.Viewport(0, 0, r.framebufferWidth, r.framebufferHeight)
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthMask(true)
	gl.DepthFunc(gl.LEQUAL)

	// TODO: Traverse octree
}

func (r *Renderer) renderForwardPass(ctx renderCtx) {
	r.lightingFramebuffer.Use()

	// TODO: Remove once lighting pass is implemented
	r.lightingFramebuffer.ClearColor(0, sprec.NewVec4(
		ctx.scene.sky.backgroundColor.X,
		ctx.scene.sky.backgroundColor.Y,
		ctx.scene.sky.backgroundColor.Z,
		1.0,
	))

	gl.Viewport(0, 0, r.framebufferWidth, r.framebufferHeight)
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthMask(false)
	gl.DepthFunc(gl.LEQUAL)

	if texture := ctx.scene.sky.skyboxTexture; texture != nil {
		gl.Enable(gl.CULL_FACE)
		r.skyboxMaterial.Program.Use()

		location := r.skyboxMaterial.Program.UniformLocation("projectionMatrixIn")
		gl.UniformMatrix4fv(location, 1, false, &ctx.projectionMatrix[0])

		location = r.skyboxMaterial.Program.UniformLocation("viewMatrixIn")
		gl.UniformMatrix4fv(location, 1, false, &ctx.viewMatrix[0])

		gl.BindTextureUnit(0, texture.ID())
		location = r.skyboxMaterial.Program.UniformLocation("albedoCubeTextureIn")
		gl.Uniform1i(location, 0)

		gl.BindVertexArray(r.skyboxMesh.VertexArray.ID())
		gl.DrawElements(r.skyboxMesh.Primitive, r.skyboxMesh.IndexCount, gl.UNSIGNED_SHORT, gl.PtrOffset(r.skyboxMesh.IndexOffsetBytes))
	}
}

func (r *Renderer) renderPostprocessingPass(ctx renderCtx) {
	r.screenFramebuffer.Use()
	gl.Viewport(int32(ctx.x), int32(ctx.y), int32(ctx.width), int32(ctx.height))
	gl.Scissor(int32(ctx.x), int32(ctx.y), int32(ctx.width), int32(ctx.height))

	gl.Disable(gl.DEPTH_TEST)
	gl.DepthMask(false)
	gl.DepthFunc(gl.LEQUAL)

	gl.Enable(gl.CULL_FACE)
	r.postprocessingMaterial.Program.Use()

	gl.BindTextureUnit(0, r.lightingAlbedoTexture.Texture.ID())
	location := r.postprocessingMaterial.Program.UniformLocation("fbColor0TextureIn")
	gl.Uniform1i(location, 0)
	location = r.postprocessingMaterial.Program.UniformLocation("exposureIn")
	gl.Uniform1f(location, ctx.camera.exposure)

	gl.BindVertexArray(r.quadMesh.VertexArray.ID())
	gl.DrawElements(r.quadMesh.Primitive, r.quadMesh.IndexCount, gl.UNSIGNED_SHORT, gl.PtrOffset(r.quadMesh.IndexOffsetBytes))
}
