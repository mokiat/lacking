package graphics

import (
	"encoding/binary"
	"fmt"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/gomath/stod"
	"github.com/mokiat/lacking/data"
	"github.com/mokiat/lacking/data/buffer"
	"github.com/mokiat/lacking/game/graphics/internal"
	"github.com/mokiat/lacking/log"
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/util/spatial"
	"github.com/x448/float16"
	"golang.org/x/exp/slices"
)

func newRenderer(api render.API, shaders ShaderCollection) *sceneRenderer {
	return &sceneRenderer{
		api:     api,
		shaders: shaders,

		exposureBufferData: make([]byte, 4*4), // Worst case RGBA32F
		exposureTarget:     1.0,

		quadMesh: internal.NewQuadMesh(),

		skyboxMesh: internal.NewSkyboxMesh(),

		visibleMeshes: spatial.NewVisitorBucket[*Mesh](2_000_000),
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
	exposureBufferData    data.Buffer
	exposureBuffer        render.Buffer
	exposureFormat        render.DataFormat
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

	cameraUniformBufferData data.Buffer
	cameraUniformBuffer     render.Buffer

	modelUniformBufferData data.Buffer
	modelUniformBuffer     render.Buffer

	materialUniformBufferData data.Buffer
	materialUniformBuffer     render.Buffer

	visibleMeshes *spatial.VisitorBucket[*Mesh]
	submeshItems  []submeshItem
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

	r.quadMesh.Allocate(r.api)

	r.createFramebuffers(800, 600)

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
		Size:    len(r.exposureBufferData),
	})
	r.exposureFormat = r.api.DetermineContentFormat(r.exposureFramebuffer)

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
		ColorWrite:   render.ColorMaskTrue,
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
		ColorWrite:   render.ColorMaskTrue,
		BlendEnabled: false,
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
}

func (r *sceneRenderer) Release() {
	defer r.commands.Release()

	defer r.quadMesh.Release()

	defer r.releaseFramebuffers()

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

	defer r.skyboxMesh.Release()
	defer r.skyboxPresentation.Delete()
	defer r.skyboxPipeline.Release()
	defer r.skycolorPresentation.Delete()
	defer r.skycolorPipeline.Release()

	defer r.cameraUniformBuffer.Release()
	defer r.modelUniformBuffer.Release()
	defer r.materialUniformBuffer.Release()
}

func (r *sceneRenderer) Render(framebuffer render.Framebuffer, viewport Viewport, scene *Scene, camera *Camera) {
	if viewport.Width != r.framebufferWidth || viewport.Height != r.framebufferHeight {
		r.releaseFramebuffers()
		r.createFramebuffers(viewport.Width, viewport.Height)
	}

	projectionMatrix := r.evaluateProjectionMatrix(camera, viewport.Width, viewport.Height)
	cameraMatrix := camera.gfxMatrix()
	viewMatrix := sprec.InverseMat4(cameraMatrix)
	projectionViewMatrix := sprec.Mat4Prod(projectionMatrix, viewMatrix)
	frustum := spatial.ProjectionRegion(stod.Mat4(projectionViewMatrix))
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

	cameraPlotter := buffer.NewPlotter(r.cameraUniformBufferData, binary.LittleEndian)
	cameraPlotter.PlotMat4(projectionMatrix)
	cameraPlotter.PlotMat4(viewMatrix)
	cameraPlotter.PlotMat4(cameraMatrix)
	r.cameraUniformBuffer.Update(render.BufferUpdateInfo{
		Data: r.cameraUniformBufferData,
	})

	r.api.UniformBufferUnit(internal.UniformBufferBindingCamera, r.cameraUniformBuffer)
	r.api.UniformBufferUnit(internal.UniformBufferBindingModel, r.modelUniformBuffer)
	r.api.UniformBufferUnit(internal.UniformBufferBindingMaterial, r.materialUniformBuffer)

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

	r.submeshItems = r.submeshItems[:0]
	ctx.scene.meshOctree.VisitHexahedronRegion(&ctx.frustum, r.visibleMeshes)
	for _, mesh := range r.visibleMeshes.Items() {
		r.queueMesh(ctx, mesh.matrixData, mesh.template)
	}
	slices.SortFunc(r.submeshItems, func(a, b submeshItem) bool {
		return a.subMesh.id < b.subMesh.id
	})
	r.renderMeshesList(ctx)

	r.api.SubmitQueue(r.commands)
	r.api.EndRenderPass()
}

func (r *sceneRenderer) queueMesh(ctx renderCtx, modelMatrix []byte, template *MeshTemplate) {
	for i := range template.subMeshes {
		subMesh := &template.subMeshes[i]
		// TODO: Move pipeline creation to mesh instantiation
		// (where also materials can be instantiated).
		if subMesh.pipeline == nil {
			material := subMesh.material
			presentation := material.geometryPresentation
			cullMode := render.CullModeNone
			if subMesh.material.backfaceCulling {
				cullMode = render.CullModeBack
			}
			subMesh.pipeline = r.api.CreatePipeline(render.PipelineInfo{
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
		}

		r.submeshItems = append(r.submeshItems, submeshItem{
			subMesh:     subMesh,
			modelMatrix: modelMatrix,
		})
	}
}

func (r *sceneRenderer) renderMeshesList(ctx renderCtx) {
	var lastSubmesh *subMeshTemplate
	count := 0
	for _, item := range r.submeshItems {
		subMesh := item.subMesh

		// just append the modelmatrix
		if subMesh == lastSubmesh {
			copy(r.modelUniformBufferData[count*64:], item.modelMatrix)
			count++

			// flush batch
			if count >= 256 {
				r.commands.UpdateBufferData(r.modelUniformBuffer, render.BufferUpdateInfo{
					Data:   r.modelUniformBufferData,
					Offset: 0,
				})
				r.commands.DrawIndexed(lastSubmesh.indexOffsetBytes, lastSubmesh.indexCount, count)
				count = 0
			}
		} else {
			// flush batch
			if lastSubmesh != nil && count > 0 {
				r.commands.UpdateBufferData(r.modelUniformBuffer, render.BufferUpdateInfo{
					Data:   r.modelUniformBufferData,
					Offset: 0,
				})
				r.commands.DrawIndexed(lastSubmesh.indexOffsetBytes, lastSubmesh.indexCount, count)
			}

			count = 0
			lastSubmesh = subMesh
			material := subMesh.material
			plotter := buffer.NewPlotter(r.materialUniformBufferData, binary.LittleEndian)
			plotter.PlotVec4(material.vectors[0])        // albedo color
			plotter.PlotFloat32(material.alphaThreshold) // alpha threshold
			plotter.PlotFloat32(material.vectors[1].X)   // normal scale
			plotter.PlotFloat32(material.vectors[1].Y)   // metallic
			plotter.PlotFloat32(material.vectors[1].Z)   // roughness
			r.commands.BindPipeline(subMesh.pipeline)
			if material.twoDTextures[0] != nil {
				r.commands.TextureUnit(internal.TextureBindingGeometryAlbedoTexture, material.twoDTextures[0])
			}
			r.commands.UpdateBufferData(r.materialUniformBuffer, render.BufferUpdateInfo{
				Data:   r.materialUniformBufferData,
				Offset: 0,
			})
			copy(r.modelUniformBufferData[count*64:], item.modelMatrix)
			count++
		}
	}

	// flush remainder
	if lastSubmesh != nil {
		r.commands.UpdateBufferData(r.modelUniformBuffer, render.BufferUpdateInfo{
			Data:   r.modelUniformBufferData,
			Offset: 0,
		})
		r.commands.DrawIndexed(lastSubmesh.indexOffsetBytes, lastSubmesh.indexCount, count)
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
	r.commands.TextureUnit(internal.TextureBindingLightingFramebufferColor0, r.geometryAlbedoTexture)
	r.commands.TextureUnit(internal.TextureBindingLightingFramebufferColor1, r.geometryNormalTexture)
	r.commands.TextureUnit(internal.TextureBindingLightingFramebufferDepth, r.geometryDepthTexture)
	r.commands.TextureUnit(internal.TextureBindingLightingReflectionTexture, light.reflectionTexture.texture)
	r.commands.TextureUnit(internal.TextureBindingLightingRefractionTexture, light.refractionTexture.texture)
	r.commands.DrawIndexed(r.quadMesh.IndexOffsetBytes, r.quadMesh.IndexCount, 1)
}

func (r *sceneRenderer) renderDirectionalLight(ctx renderCtx, light *Light) {
	r.commands.BindPipeline(r.directionalLightPipeline)
	direction := light.gfxMatrix().OrientationZ()
	r.commands.Uniform3f(r.directionalLightPresentation.LightDirection, direction.Array())
	intensity := light.intensity
	r.commands.Uniform3f(r.directionalLightPresentation.LightIntensity, intensity.Array())
	r.commands.TextureUnit(internal.TextureBindingLightingFramebufferColor0, r.geometryAlbedoTexture)
	r.commands.TextureUnit(internal.TextureBindingLightingFramebufferColor1, r.geometryNormalTexture)
	r.commands.TextureUnit(internal.TextureBindingLightingFramebufferDepth, r.geometryDepthTexture)
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
		r.commands.TextureUnit(internal.TextureBindingSkyboxAlbedoTexture, texture.texture)
		r.commands.DrawIndexed(r.skyboxMesh.IndexOffsetBytes, r.skyboxMesh.IndexCount, 1)
	} else {
		r.commands.BindPipeline(r.skycolorPipeline)
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
			var colorR, colorG, colorB float32
			switch r.exposureFormat {
			case render.DataFormatRGBA16F:
				colorR = float16.Frombits(r.exposureBufferData.Uint16(0 * 2)).Float32()
				colorG = float16.Frombits(r.exposureBufferData.Uint16(1 * 2)).Float32()
				colorB = float16.Frombits(r.exposureBufferData.Uint16(2 * 2)).Float32()
			case render.DataFormatRGBA32F:
				colorR = data.Buffer(r.exposureBufferData).Float32(0 * 4)
				colorG = data.Buffer(r.exposureBufferData).Float32(1 * 4)
				colorB = data.Buffer(r.exposureBufferData).Float32(2 * 4)
			}
			brightness := 0.2126*colorR + 0.7152*colorG + 0.0722*colorB
			if brightness < 0.001 {
				brightness = 0.001
			}
			// TODO: This needs to take elapsed time into consideration, otherwise
			// without vsync it jumps from dark to bright in an instant.
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
		r.commands.TextureUnit(internal.TextureBindingLightingFramebufferColor0, r.lightingAlbedoTexture)
		r.commands.DrawIndexed(r.quadMesh.IndexOffsetBytes, r.quadMesh.IndexCount, 1)
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
	r.commands.DrawIndexed(r.quadMesh.IndexOffsetBytes, r.quadMesh.IndexCount, 1)
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

type submeshItem struct {
	subMesh     *subMeshTemplate
	modelMatrix []byte
}
