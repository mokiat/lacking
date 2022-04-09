package graphics

import (
	"fmt"

	"github.com/go-gl/gl/v4.6-core/gl"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/graphics/renderapi/internal"
	"github.com/mokiat/lacking/game/graphics/renderapi/plugin"
	"github.com/mokiat/lacking/render"
)

const (
	coordAttributeIndex    = 0
	normalAttributeIndex   = 1
	tangentAttributeIndex  = 2
	texCoordAttributeIndex = 3
	colorAttributeIndex    = 4
)

func newRenderer(api render.API, shaders plugin.ShaderCollection) *Renderer {
	return &Renderer{
		api: api,

		framebufferWidth:  1920,
		framebufferHeight: 1080,

		exposureTarget: 1.0,

		screenFramebuffer: api.DefaultFramebuffer(),

		quadMesh: newQuadMesh(),

		skyboxMesh: newSkyboxMesh(),
	}
}

type Renderer struct {
	api      render.API
	commands render.CommandQueue

	framebufferWidth  int
	framebufferHeight int

	quadMesh *QuadMesh

	geometryAlbedoTexture render.Texture
	geometryNormalTexture render.Texture
	geometryDepthTexture  render.Texture
	geometryFramebuffer   render.Framebuffer

	lightingAlbedoTexture render.Texture
	lightingFramebuffer   render.Framebuffer

	exposureAlbedoTexture render.Texture
	exposureFramebuffer   render.Framebuffer
	exposurePresentation  *internal.LightingPresentation
	exposureBuffer        render.Buffer
	exposureSync          uintptr
	exposureTarget        float32

	screenFramebuffer render.Framebuffer

	postprocessingPresentation *internal.PostprocessingPresentation
	postprocessingPipeline     render.Pipeline

	directionalLightPresentation *internal.LightingPresentation
	directionalLightPipeline     render.Pipeline
	ambientLightPresentation     *internal.LightingPresentation
	ambientLightPipeline         render.Pipeline

	skyboxMesh           *SkyboxMesh
	skyboxPresentation   *internal.SkyboxPresentation
	skyboxPipeline       render.Pipeline
	skycolorPresentation *internal.SkyboxPresentation
	skycolorPipeline     render.Pipeline
}

func (r *Renderer) Allocate() {
	r.commands = r.api.CreateCommandQueue()

	r.quadMesh.Allocate(r.api)

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
		Format:          render.DataFormatRGBA32F,
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
		Format:          render.DataFormatRGBA32F,
	})
	r.lightingFramebuffer = r.api.CreateFramebuffer(render.FramebufferInfo{
		ColorAttachments: [4]render.Texture{
			r.lightingAlbedoTexture,
		},
		DepthAttachment: r.geometryDepthTexture,
	})

	r.exposureAlbedoTexture = r.api.CreateColorTexture2D(render.ColorTexture2DInfo{
		Width:           1,
		Height:          1,
		Wrapping:        render.WrapModeClamp,
		Filtering:       render.FilterModeNearest,
		Mipmapping:      false,
		GammaCorrection: false,
		Format:          render.DataFormatRGBA32F,
	})
	r.exposureFramebuffer = r.api.CreateFramebuffer(render.FramebufferInfo{
		ColorAttachments: [4]render.Texture{
			r.exposureAlbedoTexture,
		},
	})

	r.exposurePresentation = internal.NewExposurePresentation(r.api)

	r.exposureBuffer = r.api.CreatePixelTransferBuffer(render.BufferInfo{
		Dynamic: true,
		Data:    make([]byte, 4*4),
		// Size:    4 * 4, // TODO
	})

	r.postprocessingPresentation = internal.NewTonePostprocessingPresentation(r.api, internal.ExponentialToneMapping)
	r.postprocessingPipeline = r.api.CreatePipeline(render.PipelineInfo{
		Program:         r.postprocessingPresentation.Program,
		VertexArray:     r.quadMesh.VertexArray,
		Topology:        r.quadMesh.Topology,
		Culling:         render.CullModeBack,
		FrontFace:       render.FaceOrientationCCW,
		LineWidth:       1.0,
		DepthTest:       false,
		DepthWrite:      false,
		DepthComparison: render.ComparisonAlways,
		StencilTest:     false,
		StencilFront: render.StencilOperationState{
			StencilFailOp:  render.StencilOperationKeep,
			DepthFailOp:    render.StencilOperationKeep,
			PassOp:         render.StencilOperationKeep,
			Comparison:     render.ComparisonAlways,
			ComparisonMask: 0xFF,
			Reference:      0x00,
			WriteMask:      0xFF,
		},
		StencilBack: render.StencilOperationState{
			StencilFailOp:  render.StencilOperationKeep,
			DepthFailOp:    render.StencilOperationKeep,
			PassOp:         render.StencilOperationKeep,
			Comparison:     render.ComparisonAlways,
			ComparisonMask: 0xFF,
			Reference:      0x00,
			WriteMask:      0xFF,
		},
		ColorWrite:   [4]bool{true, true, true, true},
		BlendEnabled: false,
	})

	r.directionalLightPresentation = internal.NewDirectionalLightPresentation(r.api)
	r.directionalLightPipeline = r.api.CreatePipeline(render.PipelineInfo{
		Program:         r.directionalLightPresentation.Program,
		VertexArray:     r.quadMesh.VertexArray,
		Topology:        r.quadMesh.Topology,
		Culling:         render.CullModeBack,
		FrontFace:       render.FaceOrientationCCW,
		LineWidth:       1.0,
		DepthTest:       false,
		DepthWrite:      false,
		DepthComparison: render.ComparisonAlways,
		StencilTest:     false,
		StencilFront: render.StencilOperationState{
			StencilFailOp:  render.StencilOperationKeep,
			DepthFailOp:    render.StencilOperationKeep,
			PassOp:         render.StencilOperationKeep,
			Comparison:     render.ComparisonAlways,
			ComparisonMask: 0xFF,
			Reference:      0x00,
			WriteMask:      0xFF,
		},
		StencilBack: render.StencilOperationState{
			StencilFailOp:  render.StencilOperationKeep,
			DepthFailOp:    render.StencilOperationKeep,
			PassOp:         render.StencilOperationKeep,
			Comparison:     render.ComparisonAlways,
			ComparisonMask: 0xFF,
			Reference:      0x00,
			WriteMask:      0xFF,
		},
		ColorWrite:                  [4]bool{true, true, true, true},
		BlendEnabled:                true,
		BlendColor:                  [4]float32{0.0, 0.0, 0.0, 0.0},
		BlendSourceColorFactor:      render.BlendFactorOne,
		BlendSourceAlphaFactor:      render.BlendFactorOne,
		BlendDestinationColorFactor: render.BlendFactorOne,
		BlendDestinationAlphaFactor: render.BlendFactorZero,
		BlendOpColor:                render.BlendOperationAdd,
		BlendOpAlpha:                render.BlendOperationAdd,
	})
	r.ambientLightPresentation = internal.NewAmbientLightPresentation(r.api)
	r.ambientLightPipeline = r.api.CreatePipeline(render.PipelineInfo{
		Program:         r.ambientLightPresentation.Program,
		VertexArray:     r.quadMesh.VertexArray,
		Topology:        r.quadMesh.Topology,
		Culling:         render.CullModeBack,
		FrontFace:       render.FaceOrientationCCW,
		LineWidth:       1.0,
		DepthTest:       false,
		DepthWrite:      false,
		DepthComparison: render.ComparisonAlways,
		StencilTest:     false,
		StencilFront: render.StencilOperationState{
			StencilFailOp:  render.StencilOperationKeep,
			DepthFailOp:    render.StencilOperationKeep,
			PassOp:         render.StencilOperationKeep,
			Comparison:     render.ComparisonAlways,
			ComparisonMask: 0xFF,
			Reference:      0x00,
			WriteMask:      0xFF,
		},
		StencilBack: render.StencilOperationState{
			StencilFailOp:  render.StencilOperationKeep,
			DepthFailOp:    render.StencilOperationKeep,
			PassOp:         render.StencilOperationKeep,
			Comparison:     render.ComparisonAlways,
			ComparisonMask: 0xFF,
			Reference:      0x00,
			WriteMask:      0xFF,
		},
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

	r.skyboxMesh.Allocate(r.api)
	r.skyboxPresentation = internal.NewCubeSkyboxPresentation(r.api)
	r.skyboxPipeline = r.api.CreatePipeline(render.PipelineInfo{
		Program:         r.skyboxPresentation.Program,
		VertexArray:     r.skyboxMesh.VertexArray,
		Topology:        r.skyboxMesh.Topology,
		Culling:         render.CullModeBack,
		FrontFace:       render.FaceOrientationCCW,
		LineWidth:       1.0,
		DepthTest:       true,
		DepthWrite:      false,
		DepthComparison: render.ComparisonLessOrEqual,
		StencilTest:     false,
		StencilFront: render.StencilOperationState{
			StencilFailOp:  render.StencilOperationKeep,
			DepthFailOp:    render.StencilOperationKeep,
			PassOp:         render.StencilOperationKeep,
			Comparison:     render.ComparisonAlways,
			ComparisonMask: 0xFF,
			Reference:      0x00,
			WriteMask:      0xFF,
		},
		StencilBack: render.StencilOperationState{
			StencilFailOp:  render.StencilOperationKeep,
			DepthFailOp:    render.StencilOperationKeep,
			PassOp:         render.StencilOperationKeep,
			Comparison:     render.ComparisonAlways,
			ComparisonMask: 0xFF,
			Reference:      0x00,
			WriteMask:      0xFF,
		},
		ColorWrite:   [4]bool{true, true, true, true},
		BlendEnabled: false,
	})
	r.skycolorPresentation = internal.NewColorSkyboxPresentation(r.api)
	r.skycolorPipeline = r.api.CreatePipeline(render.PipelineInfo{
		Program:         r.skycolorPresentation.Program,
		VertexArray:     r.skyboxMesh.VertexArray,
		Topology:        r.skyboxMesh.Topology,
		Culling:         render.CullModeBack,
		FrontFace:       render.FaceOrientationCCW,
		LineWidth:       1.0,
		DepthTest:       true,
		DepthWrite:      false,
		DepthComparison: render.ComparisonLessOrEqual,
		StencilTest:     false,
		StencilFront: render.StencilOperationState{
			StencilFailOp:  render.StencilOperationKeep,
			DepthFailOp:    render.StencilOperationKeep,
			PassOp:         render.StencilOperationKeep,
			Comparison:     render.ComparisonAlways,
			ComparisonMask: 0xFF,
			Reference:      0x00,
			WriteMask:      0xFF,
		},
		StencilBack: render.StencilOperationState{
			StencilFailOp:  render.StencilOperationKeep,
			DepthFailOp:    render.StencilOperationKeep,
			PassOp:         render.StencilOperationKeep,
			Comparison:     render.ComparisonAlways,
			ComparisonMask: 0xFF,
			Reference:      0x00,
			WriteMask:      0xFF,
		},
		ColorWrite:   [4]bool{true, true, true, true},
		BlendEnabled: false,
	})
}

func (r *Renderer) Release() {
	defer r.commands.Release()

	defer r.quadMesh.Release()

	defer r.geometryAlbedoTexture.Release()
	defer r.geometryNormalTexture.Release()
	defer r.geometryDepthTexture.Release()
	defer r.geometryFramebuffer.Release()

	defer r.lightingAlbedoTexture.Release()
	defer r.lightingFramebuffer.Release()

	defer r.postprocessingPresentation.Delete()
	defer r.postprocessingPipeline.Release()

	defer r.directionalLightPresentation.Delete()
	defer r.directionalLightPipeline.Release()
	defer r.ambientLightPresentation.Delete()
	defer r.ambientLightPipeline.Release()

	defer r.skyboxMesh.Release()
	defer r.skyboxPresentation.Delete()
	defer r.skyboxPipeline.Release()
	defer r.skycolorPresentation.Delete()
	defer r.skycolorPipeline.Release()

	r.exposureBuffer.Release()
	r.exposurePresentation.Delete()
	r.exposureFramebuffer.Release()
	r.exposureAlbedoTexture.Release()

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
	// TODO: Traverse octree
	for mesh := ctx.scene.firstMesh; mesh != nil; mesh = mesh.next {
		r.renderMesh(ctx, mesh.ModelMatrix().ColumnMajorArray(), mesh.template)
	}
	r.api.SubmitQueue(r.commands)
	r.api.EndRenderPass()
}

func (r *Renderer) renderMesh(ctx renderCtx, modelMatrix [16]float32, template *MeshTemplate) {
	for _, subMesh := range template.subMeshes {
		material := subMesh.material
		presentation := material.geometryPresentation

		cullMode := render.CullModeNone
		if subMesh.material.backfaceCulling {
			cullMode = render.CullModeBack
		}

		// FIXME: Don't use dynamic pipelines
		pipeline := r.api.CreatePipeline(render.PipelineInfo{
			Program:         presentation.Program,
			VertexArray:     template.vertexArray,
			Topology:        subMesh.topology,
			Culling:         cullMode,
			FrontFace:       render.FaceOrientationCCW,
			LineWidth:       1.0,
			DepthTest:       true,
			DepthWrite:      true,
			DepthComparison: render.ComparisonLessOrEqual,
			StencilTest:     false,
			StencilFront: render.StencilOperationState{
				StencilFailOp:  render.StencilOperationKeep,
				DepthFailOp:    render.StencilOperationKeep,
				PassOp:         render.StencilOperationKeep,
				Comparison:     render.ComparisonAlways,
				ComparisonMask: 0xFF,
				Reference:      0x00,
				WriteMask:      0xFF,
			},
			StencilBack: render.StencilOperationState{
				StencilFailOp:  render.StencilOperationKeep,
				DepthFailOp:    render.StencilOperationKeep,
				PassOp:         render.StencilOperationKeep,
				Comparison:     render.ComparisonAlways,
				ComparisonMask: 0xFF,
				Reference:      0x00,
				WriteMask:      0xFF,
			},
			ColorWrite:   render.ColorMaskTrue,
			BlendEnabled: false,
		})

		r.commands.BindPipeline(pipeline)
		r.commands.UniformMatrix4f(presentation.ProjectionMatrixLocation, ctx.projectionMatrix)
		r.commands.UniformMatrix4f(presentation.ViewMatrixLocation, ctx.viewMatrix)
		r.commands.UniformMatrix4f(presentation.ModelMatrixLocation, modelMatrix)
		r.commands.Uniform1f(presentation.MetalnessLocation, material.vectors[1].Y)
		r.commands.Uniform1f(presentation.RoughnessLocation, material.vectors[1].Z)
		r.commands.Uniform4f(presentation.AlbedoColorLocation, [4]float32{
			material.vectors[0].X,
			material.vectors[0].Y,
			material.vectors[0].Z,
			material.vectors[0].W,
		})
		if material.twoDTextures[0] != nil {
			r.commands.TextureUnit(0, material.twoDTextures[0])
			r.commands.Uniform1i(presentation.AlbedoTextureLocation, 0)

		}
		r.commands.DrawIndexed(subMesh.indexOffsetBytes, subMesh.indexCount, 1)

		pipeline.Release() // FIXME: This is not even correct
	}

}

func (r *Renderer) renderLightingPass(ctx renderCtx) {
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
		}
	}
	r.api.SubmitQueue(r.commands)
	r.api.EndRenderPass()
}

func (r *Renderer) renderAmbientLight(ctx renderCtx, light *Light) {
	r.commands.BindPipeline(r.ambientLightPipeline)
	r.commands.UniformMatrix4f(r.ambientLightPresentation.ProjectionMatrixLocation, ctx.projectionMatrix)
	r.commands.UniformMatrix4f(r.ambientLightPresentation.CameraMatrixLocation, ctx.cameraMatrix)
	r.commands.UniformMatrix4f(r.ambientLightPresentation.ViewMatrixLocation, ctx.viewMatrix)
	r.commands.TextureUnit(0, r.geometryAlbedoTexture)
	r.commands.Uniform1i(r.ambientLightPresentation.FramebufferDraw0Location, 0)
	r.commands.TextureUnit(1, r.geometryNormalTexture)
	r.commands.Uniform1i(r.ambientLightPresentation.FramebufferDraw1Location, 1)
	r.commands.TextureUnit(2, r.geometryDepthTexture)
	r.commands.Uniform1i(r.ambientLightPresentation.FramebufferDepthLocation, 2)
	r.commands.TextureUnit(3, light.reflectionTexture.Texture)
	r.commands.Uniform1i(r.ambientLightPresentation.ReflectionTextureLocation, 3)
	r.commands.TextureUnit(4, light.refractionTexture.Texture)
	r.commands.Uniform1i(r.ambientLightPresentation.RefractionTextureLocation, 4)
	r.commands.DrawIndexed(r.quadMesh.IndexOffsetBytes, r.quadMesh.IndexCount, 1)
}

func (r *Renderer) renderDirectionalLight(ctx renderCtx, light *Light) {
	r.commands.BindPipeline(r.directionalLightPipeline)
	r.commands.UniformMatrix4f(r.directionalLightPresentation.ProjectionMatrixLocation, ctx.projectionMatrix)
	r.commands.UniformMatrix4f(r.directionalLightPresentation.CameraMatrixLocation, ctx.cameraMatrix)
	r.commands.UniformMatrix4f(r.directionalLightPresentation.ViewMatrixLocation, ctx.viewMatrix)
	direction := light.Rotation().OrientationZ()
	r.commands.Uniform3f(r.directionalLightPresentation.LightDirection, [3]float32{
		direction.X, direction.Y, direction.Z, // TODO: TO FLOAT ARRAY
	})
	intensity := light.intensity
	r.commands.Uniform3f(r.directionalLightPresentation.LightIntensity, [3]float32{
		intensity.X, intensity.Y, intensity.Z, // TODO: TO FLOAT ARRAY
	})
	r.commands.TextureUnit(0, r.geometryAlbedoTexture)
	r.commands.Uniform1i(r.directionalLightPresentation.FramebufferDraw0Location, 0)
	r.commands.TextureUnit(1, r.geometryNormalTexture)
	r.commands.Uniform1i(r.directionalLightPresentation.FramebufferDraw1Location, 1)
	r.commands.TextureUnit(2, r.geometryDepthTexture)
	r.commands.Uniform1i(r.directionalLightPresentation.FramebufferDepthLocation, 2)
	r.commands.DrawIndexed(r.quadMesh.IndexOffsetBytes, r.quadMesh.IndexCount, 1)
}

func (r *Renderer) renderForwardPass(ctx renderCtx) {
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
				LoadOp:  render.LoadOperationDontCare, // TODO: LoadOperationLoad: we do care
				StoreOp: render.StoreOperationStore,
			},
		},
	})

	sky := ctx.scene.sky
	if texture := sky.skyboxTexture; texture != nil {
		r.commands.BindPipeline(r.skyboxPipeline)
		r.commands.UniformMatrix4f(r.skyboxPresentation.ProjectionMatrixLocation, ctx.projectionMatrix)
		r.commands.UniformMatrix4f(r.skyboxPresentation.ViewMatrixLocation, ctx.viewMatrix)
		r.commands.TextureUnit(0, texture.Texture)
		r.commands.Uniform1i(r.skyboxPresentation.AlbedoCubeTextureLocation, 0)
		r.commands.DrawIndexed(r.skyboxMesh.IndexOffsetBytes, r.skyboxMesh.IndexCount, 1)
	} else {
		r.commands.BindPipeline(r.skycolorPipeline)
		r.commands.UniformMatrix4f(r.skycolorPresentation.ProjectionMatrixLocation, ctx.projectionMatrix)
		r.commands.UniformMatrix4f(r.skycolorPresentation.ViewMatrixLocation, ctx.viewMatrix)
		r.commands.Uniform4f(r.skycolorPresentation.AlbedoColorLocation, [4]float32{
			sky.backgroundColor.X,
			sky.backgroundColor.Y,
			sky.backgroundColor.Z,
			1.0,
		})
		r.commands.DrawIndexed(r.skyboxMesh.IndexOffsetBytes, r.skyboxMesh.IndexCount, 1)
	}

	r.api.SubmitQueue(r.commands)
	r.api.EndRenderPass()
}

func (r *Renderer) renderExposureProbePass(ctx renderCtx) {
	// if r.exposureSync != 0 {
	// 	status := gl.ClientWaitSync(r.exposureSync, gl.SYNC_FLUSH_COMMANDS_BIT, 0)
	// 	switch status {
	// 	case gl.ALREADY_SIGNALED, gl.CONDITION_SATISFIED:
	// 		data := make([]float32, 4)
	// 		gl.GetNamedBufferSubData(r.exposureBuffer.ID(), 0, 4*4, gl.Ptr(&data[0]))
	// 		brightness := 0.2126*data[0] + 0.7152*data[1] + 0.0722*data[2]
	// 		if brightness < 0.001 {
	// 			brightness = 0.001
	// 		}
	// 		r.exposureTarget = 1.0 / (3.14 * brightness)
	// 		if r.exposureTarget > ctx.camera.maxExposure {
	// 			r.exposureTarget = ctx.camera.maxExposure
	// 		}
	// 		if r.exposureTarget < ctx.camera.minExposure {
	// 			r.exposureTarget = ctx.camera.minExposure
	// 		}
	// 		gl.DeleteSync(r.exposureSync)
	// 		r.exposureSync = 0
	// 	case gl.WAIT_FAILED:
	// 		r.exposureSync = 0
	// 	}
	// }

	// ctx.camera.exposure = mix(ctx.camera.exposure, r.exposureTarget, float32(0.01))

	// if r.exposureSync == 0 {
	// 	r.exposureFramebuffer.Use()

	// 	gl.Viewport(0, 0, r.framebufferWidth, r.framebufferHeight)
	// 	gl.Disable(gl.DEPTH_TEST)
	// 	gl.DepthMask(false)
	// 	gl.Enable(gl.CULL_FACE)

	// 	r.exposureFramebuffer.ClearColor(0, sprec.ZeroVec4())

	// 	presentation := r.exposurePresentation
	// 	program := presentation.Program
	// 	program.Use()

	// 	textureUnit := uint32(0)

	// 	gl.BindTextureUnit(textureUnit, r.lightingAlbedoTexture.ID())
	// 	gl.Uniform1i(presentation.FramebufferDraw0Location, int32(textureUnit))
	// 	textureUnit++

	// 	gl.BindVertexArray(r.quadMesh.VertexArray.ID())
	// 	gl.DrawElements(r.quadMesh.Primitive, r.quadMesh.IndexCount, gl.UNSIGNED_SHORT, gl.PtrOffset(r.quadMesh.IndexOffsetBytes))

	// 	gl.TextureBarrier()

	// 	gl.BindBuffer(gl.PIXEL_PACK_BUFFER, r.exposureBuffer.ID())
	// 	gl.GetTextureImage(r.exposureAlbedoTexture.ID(), 0, gl.RGBA, gl.FLOAT, 4*4, gl.PtrOffset(0))
	// 	r.exposureSync = gl.FenceSync(gl.SYNC_GPU_COMMANDS_COMPLETE, 0)
	// 	gl.BindBuffer(gl.PIXEL_PACK_BUFFER, 0)
	// }
}

func (r *Renderer) renderPostprocessingPass(ctx renderCtx) {
	r.api.BeginRenderPass(render.RenderPassInfo{
		Framebuffer: r.screenFramebuffer,
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
	r.commands.TextureUnit(0, r.lightingAlbedoTexture)
	r.commands.Uniform1i(r.postprocessingPresentation.FramebufferDraw0Location, 0)
	r.commands.Uniform1f(r.postprocessingPresentation.ExposureLocation, ctx.camera.exposure)
	r.commands.DrawIndexed(r.quadMesh.IndexOffsetBytes, r.quadMesh.IndexCount, 1)
	r.api.SubmitQueue(r.commands)
	r.api.EndRenderPass()
}

// TODO: Move to gomath
func mix(a, b, amount float32) float32 {
	return a*(1.0-amount) + b*amount
}
