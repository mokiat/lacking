package graphics

import (
	"fmt"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/data"
	"github.com/mokiat/lacking/game/graphics/internal"
	"github.com/mokiat/lacking/render"
)

func newRenderer(api render.API, shaders ShaderCollection) *sceneRenderer {
	return &sceneRenderer{
		api:     api,
		shaders: shaders,

		framebufferWidth:  1920,
		framebufferHeight: 1080,

		exposureTarget: 1.0,

		quadMesh: internal.NewQuadMesh(),

		skyboxMesh: internal.NewSkyboxMesh(),
	}
}

type sceneRenderer struct {
	api     render.API
	shaders ShaderCollection

	commands render.CommandQueue

	framebufferWidth  int
	framebufferHeight int

	quadMesh *internal.QuadMesh

	geometryAlbedoTexture render.Texture
	geometryNormalTexture render.Texture
	geometryDepthTexture  render.Texture
	geometryFramebuffer   render.Framebuffer

	lightingAlbedoTexture render.Texture
	lightingFramebuffer   render.Framebuffer

	forwardFramebuffer render.Framebuffer

	exposureAlbedoTexture render.Texture
	exposureFramebuffer   render.Framebuffer
	exposurePresentation  *internal.LightingPresentation
	exposurePipeline      render.Pipeline
	exposureBuffer        render.Buffer
	exposureSync          render.Fence
	exposureTarget        float32

	postprocessingPresentation *internal.PostprocessingPresentation
	postprocessingPipeline     render.Pipeline

	directionalLightPresentation *internal.LightingPresentation
	directionalLightPipeline     render.Pipeline
	ambientLightPresentation     *internal.LightingPresentation
	ambientLightPipeline         render.Pipeline

	skyboxMesh           *internal.SkyboxMesh
	skyboxPresentation   *internal.SkyboxPresentation
	skyboxPipeline       render.Pipeline
	skycolorPresentation *internal.SkyboxPresentation
	skycolorPipeline     render.Pipeline
}

func (r *sceneRenderer) Allocate() {
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
	})

	r.forwardFramebuffer = r.api.CreateFramebuffer(render.FramebufferInfo{
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
	r.exposurePresentation = internal.NewLightingPresentation(r.api,
		r.shaders.ExposureSet().VertexShader(),
		r.shaders.ExposureSet().FragmentShader(),
	)
	r.exposurePipeline = r.api.CreatePipeline(render.PipelineInfo{
		Program:      r.exposurePresentation.Program,
		VertexArray:  r.quadMesh.VertexArray,
		Topology:     r.quadMesh.Topology,
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
		Size:    4 * 4,
	})

	r.postprocessingPresentation = internal.NewPostprocessingPresentation(r.api,
		r.shaders.PostprocessingSet(ExponentialToneMapping).VertexShader(),
		r.shaders.PostprocessingSet(ExponentialToneMapping).FragmentShader(),
	)
	r.postprocessingPipeline = r.api.CreatePipeline(render.PipelineInfo{
		Program:         r.postprocessingPresentation.Program,
		VertexArray:     r.quadMesh.VertexArray,
		Topology:        r.quadMesh.Topology,
		Culling:         render.CullModeBack,
		FrontFace:       render.FaceOrientationCCW,
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

	r.directionalLightPresentation = internal.NewLightingPresentation(r.api,
		r.shaders.DirectionalLightSet().VertexShader(),
		r.shaders.DirectionalLightSet().FragmentShader(),
	)
	r.directionalLightPipeline = r.api.CreatePipeline(render.PipelineInfo{
		Program:         r.directionalLightPresentation.Program,
		VertexArray:     r.quadMesh.VertexArray,
		Topology:        r.quadMesh.Topology,
		Culling:         render.CullModeBack,
		FrontFace:       render.FaceOrientationCCW,
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
	r.ambientLightPresentation = internal.NewLightingPresentation(r.api,
		r.shaders.AmbientLightSet().VertexShader(),
		r.shaders.AmbientLightSet().FragmentShader(),
	)
	r.ambientLightPipeline = r.api.CreatePipeline(render.PipelineInfo{
		Program:         r.ambientLightPresentation.Program,
		VertexArray:     r.quadMesh.VertexArray,
		Topology:        r.quadMesh.Topology,
		Culling:         render.CullModeBack,
		FrontFace:       render.FaceOrientationCCW,
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
	r.skyboxPresentation = internal.NewSkyboxPresentation(r.api,
		r.shaders.SkyboxSet().VertexShader(),
		r.shaders.SkyboxSet().FragmentShader(),
	)
	r.skyboxPipeline = r.api.CreatePipeline(render.PipelineInfo{
		Program:         r.skyboxPresentation.Program,
		VertexArray:     r.skyboxMesh.VertexArray,
		Topology:        r.skyboxMesh.Topology,
		Culling:         render.CullModeBack,
		FrontFace:       render.FaceOrientationCCW,
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
	r.skycolorPresentation = internal.NewSkyboxPresentation(r.api,
		r.shaders.SkycolorSet().VertexShader(),
		r.shaders.SkycolorSet().FragmentShader(),
	)
	r.skycolorPipeline = r.api.CreatePipeline(render.PipelineInfo{
		Program:         r.skycolorPresentation.Program,
		VertexArray:     r.skyboxMesh.VertexArray,
		Topology:        r.skyboxMesh.Topology,
		Culling:         render.CullModeBack,
		FrontFace:       render.FaceOrientationCCW,
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

func (r *sceneRenderer) Release() {
	defer r.commands.Release()

	defer r.quadMesh.Release()

	defer r.geometryAlbedoTexture.Release()
	defer r.geometryNormalTexture.Release()
	defer r.geometryDepthTexture.Release()
	defer r.geometryFramebuffer.Release()

	defer r.lightingAlbedoTexture.Release()
	defer r.lightingFramebuffer.Release()

	defer r.forwardFramebuffer.Release()

	defer r.exposureBuffer.Release()
	defer r.exposurePresentation.Delete()
	defer r.exposureFramebuffer.Release()
	defer r.exposureAlbedoTexture.Release()

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
}

type renderCtx struct {
	framebuffer      render.Framebuffer
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

func (r *sceneRenderer) Render(framebuffer render.Framebuffer, viewport Viewport, scene *Scene, camera *Camera) {
	projectionMatrix := r.evaluateProjectionMatrix(camera, viewport.Width, viewport.Height)
	cameraMatrix := camera.Matrix()
	viewMatrix := sprec.InverseMat4(cameraMatrix)
	ctx := renderCtx{
		framebuffer:      framebuffer,
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
		far  = float32(900.0)
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

func (r *sceneRenderer) renderGeometryPass(ctx renderCtx) {
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
		r.renderMesh(ctx, mesh.Matrix().ColumnMajorArray(), mesh.template)
	}
	r.api.SubmitQueue(r.commands)
	r.api.EndRenderPass()
}

func (r *sceneRenderer) renderMesh(ctx renderCtx, modelMatrix [16]float32, template *MeshTemplate) {
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

func (r *sceneRenderer) renderLightingPass(ctx renderCtx) {
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

func (r *sceneRenderer) renderAmbientLight(ctx renderCtx, light *Light) {
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
	r.commands.TextureUnit(3, light.reflectionTexture.texture)
	r.commands.Uniform1i(r.ambientLightPresentation.ReflectionTextureLocation, 3)
	r.commands.TextureUnit(4, light.refractionTexture.texture)
	r.commands.Uniform1i(r.ambientLightPresentation.RefractionTextureLocation, 4)
	r.commands.DrawIndexed(r.quadMesh.IndexOffsetBytes, r.quadMesh.IndexCount, 1)
}

func (r *sceneRenderer) renderDirectionalLight(ctx renderCtx, light *Light) {
	r.commands.BindPipeline(r.directionalLightPipeline)
	r.commands.UniformMatrix4f(r.directionalLightPresentation.ProjectionMatrixLocation, ctx.projectionMatrix)
	r.commands.UniformMatrix4f(r.directionalLightPresentation.CameraMatrixLocation, ctx.cameraMatrix)
	r.commands.UniformMatrix4f(r.directionalLightPresentation.ViewMatrixLocation, ctx.viewMatrix)
	direction := light.Rotation().OrientationZ()
	r.commands.Uniform3f(r.directionalLightPresentation.LightDirection, direction.Array())
	intensity := light.intensity
	r.commands.Uniform3f(r.directionalLightPresentation.LightIntensity, intensity.Array())
	r.commands.TextureUnit(0, r.geometryAlbedoTexture)
	r.commands.Uniform1i(r.directionalLightPresentation.FramebufferDraw0Location, 0)
	r.commands.TextureUnit(1, r.geometryNormalTexture)
	r.commands.Uniform1i(r.directionalLightPresentation.FramebufferDraw1Location, 1)
	r.commands.TextureUnit(2, r.geometryDepthTexture)
	r.commands.Uniform1i(r.directionalLightPresentation.FramebufferDepthLocation, 2)
	r.commands.DrawIndexed(r.quadMesh.IndexOffsetBytes, r.quadMesh.IndexCount, 1)
}

func (r *sceneRenderer) renderForwardPass(ctx renderCtx) {
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
		r.commands.UniformMatrix4f(r.skyboxPresentation.ProjectionMatrixLocation, ctx.projectionMatrix)
		r.commands.UniformMatrix4f(r.skyboxPresentation.ViewMatrixLocation, ctx.viewMatrix)
		r.commands.TextureUnit(0, texture.texture)
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

func (r *sceneRenderer) renderExposureProbePass(ctx renderCtx) {
	if r.exposureSync != nil {
		switch r.exposureSync.Status() {
		case render.FenceStatusSuccess:
			colorData := data.Buffer(make([]byte, 4*4)) // TODO: Prevent allocation
			r.exposureBuffer.Fetch(render.BufferFetchInfo{
				Offset: 0,
				Target: colorData,
			})
			colorR := colorData.Float32(0 * 4)
			colorG := colorData.Float32(1 * 4)
			colorB := colorData.Float32(2 * 4)
			brightness := 0.2126*colorR + 0.7152*colorG + 0.0722*colorB
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
			fallthrough

		case render.FenceStatusDeviceLost:
			r.exposureSync.Delete()
			r.exposureSync = nil

		case render.FenceStatusNotReady:
			// wait until next frame
		}
	}

	ctx.camera.exposure = sprec.Mix(ctx.camera.exposure, r.exposureTarget, float32(0.01))

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
		r.commands.TextureUnit(0, r.lightingAlbedoTexture)
		r.commands.Uniform1i(r.exposurePresentation.FramebufferDraw0Location, 0)
		r.commands.DrawIndexed(r.quadMesh.IndexOffsetBytes, r.quadMesh.IndexCount, 1)
		r.commands.CopyContentToBuffer(render.CopyContentToBufferInfo{
			Buffer: r.exposureBuffer,
			X:      0,
			Y:      0,
			Width:  1,
			Height: 1,
			Format: render.DataFormatRGBA32F,
		})
		r.api.SubmitQueue(r.commands)
		r.exposureSync = r.api.CreateFence()
		r.api.EndRenderPass()
	}
}

func (r *sceneRenderer) renderPostprocessingPass(ctx renderCtx) {
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
	r.commands.TextureUnit(0, r.lightingAlbedoTexture)
	r.commands.Uniform1i(r.postprocessingPresentation.FramebufferDraw0Location, 0)
	r.commands.Uniform1f(r.postprocessingPresentation.ExposureLocation, ctx.camera.exposure)
	r.commands.DrawIndexed(r.quadMesh.IndexOffsetBytes, r.quadMesh.IndexCount, 1)
	r.api.SubmitQueue(r.commands)
	r.api.EndRenderPass()
}
