package graphics

import (
	"fmt"

	"github.com/go-gl/gl/v4.6-core/gl"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/framework/opengl"
	"github.com/mokiat/lacking/framework/opengl/game/graphics/internal"
	"github.com/mokiat/lacking/game/graphics"
)

const (
	framebufferWidth  = int32(1920)
	framebufferHeight = int32(1080)

	coordAttributeIndex    = 0
	normalAttributeIndex   = 1
	tangentAttributeIndex  = 2
	texCoordAttributeIndex = 3
	colorAttributeIndex    = 4
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
		lightingDepthTexture:  opengl.NewTwoDTexture(),
		lightingFramebuffer:   opengl.NewFramebuffer(),

		exposureAlbedoTexture: opengl.NewTwoDTexture(),
		exposureFramebuffer:   opengl.NewFramebuffer(),
		exposureBuffer:        opengl.NewBuffer(),
		exposureTarget:        1.0,

		screenFramebuffer: opengl.DefaultFramebuffer(),

		quadMesh: newQuadMesh(),

		skyboxMesh: newSkyboxMesh(),
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
	lightingDepthTexture  *opengl.TwoDTexture
	lightingFramebuffer   *opengl.Framebuffer

	exposureAlbedoTexture *opengl.TwoDTexture
	exposureFramebuffer   *opengl.Framebuffer
	exposurePresentation  *internal.LightingPresentation
	exposureBuffer        *opengl.Buffer
	exposureSync          uintptr
	exposureTarget        float32

	screenFramebuffer *opengl.Framebuffer

	postprocessingPresentation *internal.PostprocessingPresentation

	directionalLightPresentation *internal.LightingPresentation
	ambientLightPresentation     *internal.LightingPresentation

	quadMesh *QuadMesh

	skyboxPresentation   *internal.SkyboxPresentation
	skycolorPresentation *internal.SkyboxPresentation
	skyboxMesh           *SkyboxMesh
}

func (r *Renderer) Allocate() {
	r.geometryAlbedoTexture.Allocate(opengl.TwoDTextureAllocateInfo{
		Width:             framebufferWidth,
		Height:            framebufferHeight,
		MinFilter:         gl.NEAREST,
		MagFilter:         gl.NEAREST,
		InternalFormat:    gl.RGBA8,
		DataFormat:        gl.RGBA,
		DataComponentType: gl.UNSIGNED_BYTE,
	})

	r.geometryNormalTexture.Allocate(opengl.TwoDTextureAllocateInfo{
		Width:             framebufferWidth,
		Height:            framebufferHeight,
		MinFilter:         gl.NEAREST,
		MagFilter:         gl.NEAREST,
		InternalFormat:    gl.RGBA32F,
		DataFormat:        gl.RGBA,
		DataComponentType: gl.FLOAT,
	})

	r.geometryDepthTexture.Allocate(opengl.TwoDTextureAllocateInfo{
		Width:             framebufferWidth,
		Height:            framebufferHeight,
		MinFilter:         gl.NEAREST,
		MagFilter:         gl.NEAREST,
		InternalFormat:    gl.DEPTH_COMPONENT32,
		DataFormat:        gl.DEPTH_COMPONENT,
		DataComponentType: gl.FLOAT,
	})

	r.geometryFramebuffer.Allocate(opengl.FramebufferAllocateInfo{
		ColorAttachments: []*opengl.Texture{
			&r.geometryAlbedoTexture.Texture,
			&r.geometryNormalTexture.Texture,
		},
		DepthAttachment: &r.geometryDepthTexture.Texture,
	})

	r.lightingAlbedoTexture.Allocate(opengl.TwoDTextureAllocateInfo{
		Width:             framebufferWidth,
		Height:            framebufferHeight,
		MinFilter:         gl.NEAREST,
		MagFilter:         gl.NEAREST,
		InternalFormat:    gl.RGBA32F,
		DataFormat:        gl.RGBA,
		DataComponentType: gl.FLOAT,
	})

	r.lightingDepthTexture.Allocate(opengl.TwoDTextureAllocateInfo{
		Width:             framebufferWidth,
		Height:            framebufferHeight,
		MinFilter:         gl.NEAREST,
		MagFilter:         gl.NEAREST,
		InternalFormat:    gl.DEPTH_COMPONENT32,
		DataFormat:        gl.DEPTH_COMPONENT,
		DataComponentType: gl.FLOAT,
	})

	r.lightingFramebuffer.Allocate(opengl.FramebufferAllocateInfo{
		ColorAttachments: []*opengl.Texture{
			&r.lightingAlbedoTexture.Texture,
		},
		DepthAttachment: &r.lightingDepthTexture.Texture,
	})

	r.exposureAlbedoTexture.Allocate(opengl.TwoDTextureAllocateInfo{
		Width:             1,
		Height:            1,
		MinFilter:         gl.NEAREST,
		MagFilter:         gl.NEAREST,
		InternalFormat:    gl.RGBA32F,
		DataFormat:        gl.RGBA,
		DataComponentType: gl.FLOAT,
	})

	r.exposureFramebuffer.Allocate(opengl.FramebufferAllocateInfo{
		ColorAttachments: []*opengl.Texture{
			&r.exposureAlbedoTexture.Texture,
		},
	})
	r.exposurePresentation = internal.NewExposurePresentation()

	r.exposureBuffer.Allocate(opengl.BufferAllocateInfo{
		Dynamic: true,
		Data:    make([]byte, 4*4),
	})

	r.postprocessingPresentation = internal.NewTonePostprocessingPresentation(internal.ExponentialToneMapping)

	r.directionalLightPresentation = internal.NewDirectionalLightPresentation()
	r.ambientLightPresentation = internal.NewAmbientLightPresentation()

	r.quadMesh.Allocate()

	r.skyboxPresentation = internal.NewCubeSkyboxPresentation()
	r.skycolorPresentation = internal.NewColorSkyboxPresentation()
	r.skyboxMesh.Allocate()
}

func (r *Renderer) Release() {
	r.skyboxPresentation.Delete()
	r.skyboxMesh.Release()

	r.ambientLightPresentation.Delete()
	r.directionalLightPresentation.Delete()

	r.quadMesh.Release()

	r.postprocessingPresentation.Delete()

	r.exposureBuffer.Release()
	r.exposurePresentation.Delete()
	r.exposureFramebuffer.Release()
	r.exposureAlbedoTexture.Release()

	r.lightingFramebuffer.Release()
	r.lightingAlbedoTexture.Release()
	r.lightingDepthTexture.Release()

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
	cameraMatrix := camera.ModelMatrix()
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
	gl.TextureBarrier()
	r.renderLightingPass(ctx)
	r.renderForwardPass(ctx)
	if camera.autoExposureEnabled {
		gl.TextureBarrier()
		r.renderExposureProbePass(ctx)
	}
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

func (r *Renderer) renderGeometryPass(ctx renderCtx) {
	r.geometryFramebuffer.Use()

	gl.Viewport(0, 0, r.framebufferWidth, r.framebufferHeight)
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthMask(true)
	gl.DepthFunc(gl.LEQUAL)

	r.geometryFramebuffer.ClearColor(0, sprec.NewVec4(
		ctx.scene.sky.backgroundColor.X,
		ctx.scene.sky.backgroundColor.Y,
		ctx.scene.sky.backgroundColor.Z,
		1.0,
	))
	r.geometryFramebuffer.ClearDepth(1.0)

	// TODO: Traverse octree
	for mesh := ctx.scene.firstMesh; mesh != nil; mesh = mesh.next {
		r.renderMesh(ctx, mesh.ModelMatrix().ColumnMajorArray(), mesh.template)
	}
}

func (r *Renderer) renderMesh(ctx renderCtx, modelMatrix [16]float32, template *MeshTemplate) {
	for _, subMesh := range template.subMeshes {
		if subMesh.material.backfaceCulling {
			gl.Enable(gl.CULL_FACE)
		} else {
			gl.Disable(gl.CULL_FACE)
		}

		material := subMesh.material
		presentation := material.geometryPresentation
		presentation.Program.Use()

		gl.UniformMatrix4fv(presentation.ProjectionMatrixLocation, 1, false, &ctx.projectionMatrix[0])
		gl.UniformMatrix4fv(presentation.ViewMatrixLocation, 1, false, &ctx.viewMatrix[0])
		gl.UniformMatrix4fv(presentation.ModelMatrixLocation, 1, false, &modelMatrix[0])

		gl.Uniform1f(presentation.MetalnessLocation, material.vectors[1].Y)
		gl.Uniform1f(presentation.RoughnessLocation, material.vectors[1].Z)
		gl.Uniform4f(presentation.AlbedoColorLocation, material.vectors[0].X, material.vectors[0].Y, material.vectors[0].Z, material.vectors[0].Z)

		textureUnit := uint32(0)
		if material.twoDTextures[0] != nil {
			gl.BindTextureUnit(textureUnit, material.twoDTextures[0].ID())
			gl.Uniform1i(presentation.AlbedoTextureLocation, int32(textureUnit))
			textureUnit++
		}

		gl.BindVertexArray(template.vertexArray.ID())
		gl.DrawElements(subMesh.primitive, subMesh.indexCount, subMesh.indexType, gl.PtrOffset(subMesh.indexOffsetBytes))
	}
}

func (r *Renderer) renderLightingPass(ctx renderCtx) {
	gl.BlitNamedFramebuffer(r.geometryFramebuffer.ID(), r.lightingFramebuffer.ID(),
		0, 0, r.framebufferWidth, r.framebufferHeight,
		0, 0, r.framebufferWidth, r.framebufferHeight,
		gl.DEPTH_BUFFER_BIT,
		gl.NEAREST,
	)

	r.lightingFramebuffer.Use()

	gl.Viewport(0, 0, r.framebufferWidth, r.framebufferHeight)
	gl.Disable(gl.DEPTH_TEST)
	gl.DepthMask(false)
	gl.Enable(gl.CULL_FACE)

	r.lightingFramebuffer.ClearColor(0, sprec.NewVec4(0.0, 0.0, 0.0, 1.0))

	gl.Enablei(gl.BLEND, 0)
	gl.BlendEquationSeparate(gl.FUNC_ADD, gl.FUNC_ADD)
	gl.BlendFuncSeparate(gl.ONE, gl.ONE, gl.ONE, gl.ZERO)

	// TODO: Traverse octree
	for light := ctx.scene.firstLight; light != nil; light = light.next {
		switch light.mode {
		case LightModeDirectional:
			r.renderDirectionalLight(ctx, light)
		case LightModeAmbient:
			r.renderAmbientLight(ctx, light)
		}
	}

	gl.Disablei(gl.BLEND, 0)
}

func (r *Renderer) renderAmbientLight(ctx renderCtx, light *Light) {
	presentation := r.ambientLightPresentation
	presentation.Program.Use()

	gl.UniformMatrix4fv(presentation.ProjectionMatrixLocation, 1, false, &ctx.projectionMatrix[0])
	gl.UniformMatrix4fv(presentation.CameraMatrixLocation, 1, false, &ctx.cameraMatrix[0])
	gl.UniformMatrix4fv(presentation.ViewMatrixLocation, 1, false, &ctx.viewMatrix[0])

	textureUnit := uint32(0)

	gl.BindTextureUnit(textureUnit, r.geometryAlbedoTexture.ID())
	gl.Uniform1i(presentation.FramebufferDraw0Location, int32(textureUnit))
	textureUnit++

	gl.BindTextureUnit(textureUnit, r.geometryNormalTexture.ID())
	gl.Uniform1i(presentation.FramebufferDraw1Location, int32(textureUnit))
	textureUnit++

	gl.BindTextureUnit(textureUnit, r.geometryDepthTexture.ID())
	gl.Uniform1i(presentation.FramebufferDepthLocation, int32(textureUnit))
	textureUnit++

	gl.BindTextureUnit(textureUnit, light.reflectionTexture.ID())
	gl.Uniform1i(presentation.ReflectionTextureLocation, int32(textureUnit))
	textureUnit++

	gl.BindTextureUnit(textureUnit, light.refractionTexture.ID())
	gl.Uniform1i(presentation.RefractionTextureLocation, int32(textureUnit))
	textureUnit++

	gl.BindVertexArray(r.quadMesh.VertexArray.ID())
	gl.DrawElements(r.quadMesh.Primitive, r.quadMesh.IndexCount, gl.UNSIGNED_SHORT, gl.PtrOffset(r.quadMesh.IndexOffsetBytes))
}

func (r *Renderer) renderDirectionalLight(ctx renderCtx, light *Light) {
	presentation := r.directionalLightPresentation
	presentation.Program.Use()

	gl.UniformMatrix4fv(presentation.ProjectionMatrixLocation, 1, false, &ctx.projectionMatrix[0])
	gl.UniformMatrix4fv(presentation.CameraMatrixLocation, 1, false, &ctx.cameraMatrix[0])
	gl.UniformMatrix4fv(presentation.ViewMatrixLocation, 1, false, &ctx.viewMatrix[0])

	direction := light.Rotation().OrientationZ()
	gl.Uniform3f(presentation.LightDirection, direction.X, direction.Y, direction.Z)
	intensity := light.intensity
	gl.Uniform3f(presentation.LightIntensity, intensity.X, intensity.Y, intensity.Z)

	textureUnit := uint32(0)

	gl.BindTextureUnit(textureUnit, r.geometryAlbedoTexture.ID())
	gl.Uniform1i(presentation.FramebufferDraw0Location, int32(textureUnit))
	textureUnit++

	gl.BindTextureUnit(textureUnit, r.geometryNormalTexture.ID())
	gl.Uniform1i(presentation.FramebufferDraw1Location, int32(textureUnit))
	textureUnit++

	gl.BindTextureUnit(textureUnit, r.geometryDepthTexture.ID())
	gl.Uniform1i(presentation.FramebufferDepthLocation, int32(textureUnit))
	textureUnit++

	gl.BindVertexArray(r.quadMesh.VertexArray.ID())
	gl.DrawElements(r.quadMesh.Primitive, r.quadMesh.IndexCount, gl.UNSIGNED_SHORT, gl.PtrOffset(r.quadMesh.IndexOffsetBytes))
}

func (r *Renderer) renderForwardPass(ctx renderCtx) {
	r.lightingFramebuffer.Use()

	gl.Viewport(0, 0, r.framebufferWidth, r.framebufferHeight)
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthMask(false)
	gl.DepthFunc(gl.LEQUAL)

	sky := ctx.scene.sky
	if texture := sky.skyboxTexture; texture != nil {
		gl.Enable(gl.CULL_FACE)

		presentation := r.skyboxPresentation
		program := presentation.Program
		program.Use()

		gl.UniformMatrix4fv(presentation.ProjectionMatrixLocation, 1, false, &ctx.projectionMatrix[0])
		gl.UniformMatrix4fv(presentation.ViewMatrixLocation, 1, false, &ctx.viewMatrix[0])

		gl.BindTextureUnit(0, texture.ID())
		gl.Uniform1i(presentation.AlbedoCubeTextureLocation, 0)

		gl.BindVertexArray(r.skyboxMesh.VertexArray.ID())
		gl.DrawElements(r.skyboxMesh.Primitive, r.skyboxMesh.IndexCount, gl.UNSIGNED_SHORT, gl.PtrOffset(r.skyboxMesh.IndexOffsetBytes))
	} else {
		gl.Enable(gl.CULL_FACE)

		presentation := r.skycolorPresentation
		program := presentation.Program
		program.Use()

		gl.UniformMatrix4fv(presentation.ProjectionMatrixLocation, 1, false, &ctx.projectionMatrix[0])
		gl.UniformMatrix4fv(presentation.ViewMatrixLocation, 1, false, &ctx.viewMatrix[0])

		gl.Uniform4f(presentation.AlbedoColorLocation,
			sky.backgroundColor.X,
			sky.backgroundColor.Y,
			sky.backgroundColor.Z,
			1.0,
		)

		gl.BindVertexArray(r.skyboxMesh.VertexArray.ID())
		gl.DrawElements(r.skyboxMesh.Primitive, r.skyboxMesh.IndexCount, gl.UNSIGNED_SHORT, gl.PtrOffset(r.skyboxMesh.IndexOffsetBytes))
	}
}

func (r *Renderer) renderExposureProbePass(ctx renderCtx) {
	if r.exposureSync != 0 {
		status := gl.ClientWaitSync(r.exposureSync, gl.SYNC_FLUSH_COMMANDS_BIT, 0)
		switch status {
		case gl.ALREADY_SIGNALED, gl.CONDITION_SATISFIED:
			data := make([]float32, 4)
			gl.GetNamedBufferSubData(r.exposureBuffer.ID(), 0, 4*4, gl.Ptr(&data[0]))
			brightness := 0.2126*data[0] + 0.7152*data[1] + 0.0722*data[2]
			if brightness < 0.001 {
				brightness = 0.001
			}
			r.exposureTarget = 1.0 / (3.14 * brightness)
			if r.exposureTarget > ctx.camera.maxExposure {
				r.exposureTarget = ctx.camera.maxExposure
			}
			if r.exposureTarget < ctx.camera.minExposure {
				r.exposureTarget = ctx.camera.minExposure
			}
			gl.DeleteSync(r.exposureSync)
			r.exposureSync = 0
		case gl.WAIT_FAILED:
			r.exposureSync = 0
		}
	}

	ctx.camera.exposure = mix(ctx.camera.exposure, r.exposureTarget, float32(0.01))

	if r.exposureSync == 0 {
		r.exposureFramebuffer.Use()

		gl.Viewport(0, 0, r.framebufferWidth, r.framebufferHeight)
		gl.Disable(gl.DEPTH_TEST)
		gl.DepthMask(false)
		gl.Enable(gl.CULL_FACE)

		r.exposureFramebuffer.ClearColor(0, sprec.ZeroVec4())

		presentation := r.exposurePresentation
		program := presentation.Program
		program.Use()

		textureUnit := uint32(0)

		gl.BindTextureUnit(textureUnit, r.lightingAlbedoTexture.ID())
		gl.Uniform1i(presentation.FramebufferDraw0Location, int32(textureUnit))
		textureUnit++

		gl.BindVertexArray(r.quadMesh.VertexArray.ID())
		gl.DrawElements(r.quadMesh.Primitive, r.quadMesh.IndexCount, gl.UNSIGNED_SHORT, gl.PtrOffset(r.quadMesh.IndexOffsetBytes))

		gl.TextureBarrier()

		gl.BindBuffer(gl.PIXEL_PACK_BUFFER, r.exposureBuffer.ID())
		gl.GetTextureImage(r.exposureAlbedoTexture.ID(), 0, gl.RGBA, gl.FLOAT, 4*4, gl.PtrOffset(0))
		r.exposureSync = gl.FenceSync(gl.SYNC_GPU_COMMANDS_COMPLETE, 0)
		gl.BindBuffer(gl.PIXEL_PACK_BUFFER, 0)
	}
}

func (r *Renderer) renderPostprocessingPass(ctx renderCtx) {
	r.screenFramebuffer.Use()
	gl.Viewport(int32(ctx.x), int32(ctx.y), int32(ctx.width), int32(ctx.height))
	gl.Scissor(int32(ctx.x), int32(ctx.y), int32(ctx.width), int32(ctx.height))

	gl.Disable(gl.DEPTH_TEST)
	gl.DepthMask(false)
	gl.DepthFunc(gl.ALWAYS)

	gl.Enable(gl.CULL_FACE)
	presentation := r.postprocessingPresentation
	presentation.Program.Use()

	gl.BindTextureUnit(0, r.lightingAlbedoTexture.ID())
	gl.Uniform1i(presentation.FramebufferDraw0Location, 0)
	gl.Uniform1f(presentation.ExposureLocation, ctx.camera.exposure)

	gl.BindVertexArray(r.quadMesh.VertexArray.ID())
	gl.DrawElements(r.quadMesh.Primitive, r.quadMesh.IndexCount, gl.UNSIGNED_SHORT, gl.PtrOffset(r.quadMesh.IndexOffsetBytes))
}

// TODO: Move to gomath
func mix(a, b, amount float32) float32 {
	return a*(1.0-amount) + b*amount
}
