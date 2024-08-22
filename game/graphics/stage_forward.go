package graphics

import (
	"github.com/mokiat/gog/ds"
	"github.com/mokiat/lacking/debug/metric"
	"github.com/mokiat/lacking/game/graphics/internal"
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/render/ubo"
	"github.com/mokiat/lacking/util/blob"
)

const (
	debugVertexSize   = 3*render.SizeF32 + 4*render.SizeU8
	debugBufferSize   = 1024 * 1024
	debugMaxLineCount = debugBufferSize / ((debugVertexSize * 2) * 2) // dobule buffer
)

type ForwardStageInput struct {
	HDRTexture   StageTextureParameter
	DepthTexture StageTextureParameter
}

func newForwardStage(api render.API, shaders ShaderCollection, data *commonStageData, meshRenderer *meshRenderer, input ForwardStageInput) *ForwardStage {
	return &ForwardStage{
		api:          api,
		shaders:      shaders,
		data:         data,
		input:        input,
		meshRenderer: meshRenderer,
	}
}

var _ Stage = (*ForwardStage)(nil)

type ForwardStage struct {
	api          render.API
	shaders      ShaderCollection
	data         *commonStageData
	meshRenderer *meshRenderer
	input        ForwardStageInput

	hdrTexture   render.Texture
	depthTexture render.Texture
	framebuffer  render.Framebuffer

	debugVertexData         []byte
	debugVertexBuffer       render.Buffer
	debugVertexBufferOffset uint32
	debugVertexArray        render.VertexArray
	debugProgram            render.Program
	debugPipeline           render.Pipeline
}

func (s *ForwardStage) Allocate() {
	s.allocateFramebuffer()
	s.allocateDebug()
}

func (s *ForwardStage) Release() {
	defer s.releaseFramebuffer()
	defer s.releaseDebug()
}

func (s *ForwardStage) PreRender(width, height uint32) {
	hdrTexture := s.input.HDRTexture()
	depthTexture := s.input.DepthTexture()
	if hdrTexture != s.hdrTexture || depthTexture != s.depthTexture {
		s.releaseFramebuffer()
		s.allocateFramebuffer()
	}
}

func (s *ForwardStage) Render(ctx StageContext) {
	// TODO: Use built-in tracing. It does not actually allocate memory
	// when tracing is not enabled. Furthermore, the context can be Background
	// and nesting should still work. The context is used for tasks.
	defer metric.BeginRegion("forward").End()

	commandBuffer := s.data.CommandBuffer()
	commandBuffer.BeginRenderPass(render.RenderPassInfo{
		Framebuffer: s.framebuffer,
		Viewport: render.Area{
			// TODO: This should be based on HDR texture size.
			Width:  ctx.Viewport.Width,
			Height: ctx.Viewport.Height,
		},
		DepthLoadOp:    render.LoadOperationLoad,
		DepthStoreOp:   render.StoreOperationStore,
		StencilLoadOp:  render.LoadOperationLoad,
		StencilStoreOp: render.StoreOperationDiscard,
		Colors: [4]render.ColorAttachmentInfo{
			{
				LoadOp:  render.LoadOperationLoad,
				StoreOp: render.StoreOperationStore,
			},
		},
	})
	s.renderSky(ctx)
	s.renderDebug(ctx)
	s.renderMeshes(ctx)
	commandBuffer.EndRenderPass()
}

func (s *ForwardStage) PostRender() {
	// Nothing to do here.
}

func (s *ForwardStage) allocateFramebuffer() {
	s.hdrTexture = s.input.HDRTexture()
	s.depthTexture = s.input.DepthTexture()

	s.framebuffer = s.api.CreateFramebuffer(render.FramebufferInfo{
		ColorAttachments: [4]render.Texture{
			s.hdrTexture,
		},
		DepthAttachment: s.depthTexture,
	})
}

func (s *ForwardStage) releaseFramebuffer() {
	defer s.framebuffer.Release()
}

func (s *ForwardStage) allocateDebug() {
	s.debugVertexData = make([]byte, debugBufferSize)
	s.debugVertexBuffer = s.api.CreateVertexBuffer(render.BufferInfo{
		Dynamic: true,
		Data:    s.debugVertexData,
	})

	const coordOffset = 0
	const colorOffset = coordOffset + 3*render.SizeF32
	s.debugVertexArray = s.api.CreateVertexArray(render.VertexArrayInfo{
		Bindings: []render.VertexArrayBinding{
			render.NewVertexArrayBinding(s.debugVertexBuffer, debugVertexSize),
		},
		Attributes: []render.VertexArrayAttribute{
			render.NewVertexArrayAttribute(0, internal.CoordAttributeIndex, coordOffset, render.VertexAttributeFormatRGB32F),
			render.NewVertexArrayAttribute(0, internal.ColorAttributeIndex, colorOffset, render.VertexAttributeFormatRGBA8UN),
		},
	})

	s.debugProgram = s.api.CreateProgram(render.ProgramInfo{
		SourceCode:      s.shaders.DebugSet(),
		TextureBindings: []render.TextureBinding{},
		UniformBindings: []render.UniformBinding{
			render.NewUniformBinding("Camera", internal.UniformBufferBindingCamera),
		},
	})
	s.debugPipeline = s.api.CreatePipeline(render.PipelineInfo{
		Program:         s.debugProgram,
		VertexArray:     s.debugVertexArray,
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
}

func (s *ForwardStage) releaseDebug() {
	defer s.debugVertexBuffer.Release()
	defer s.debugVertexArray.Release()
	defer s.debugProgram.Release()
	defer s.debugPipeline.Release()
}

func (s *ForwardStage) renderSky(ctx StageContext) {
	sky := s.findActiveSky(ctx.Scene.skies)
	if sky == nil {
		return
	}

	commandBuffer := s.data.CommandBuffer()
	uniformBuffer := s.data.UniformBuffer()

	for _, pass := range sky.definition.renderPasses {
		commandBuffer.BindPipeline(pass.Pipeline)
		commandBuffer.UniformBufferUnit(
			internal.UniformBufferBindingCamera,
			ctx.CameraPlacement.Buffer,
			ctx.CameraPlacement.Offset,
			ctx.CameraPlacement.Size,
		)
		if !pass.UniformSet.IsEmpty() {
			materialData := ubo.WriteUniform(uniformBuffer, internal.MaterialUniform{
				Data: pass.UniformSet.Data(),
			})
			commandBuffer.UniformBufferUnit(
				internal.UniformBufferBindingMaterial,
				materialData.Buffer,
				materialData.Offset,
				materialData.Size,
			)
		}
		for i := range pass.TextureSet.TextureCount() {
			if texture := pass.TextureSet.TextureAt(i); texture != nil {
				commandBuffer.TextureUnit(uint(i), texture)
			}
			if sampler := pass.TextureSet.SamplerAt(i); sampler != nil {
				commandBuffer.SamplerUnit(uint(i), sampler)
			}
		}
		commandBuffer.DrawIndexed(pass.IndexByteOffset, pass.IndexCount, 1)
	}
}

func (s *ForwardStage) renderDebug(ctx StageContext) {
	count := len(ctx.DebugLines)
	if count == 0 {
		return
	}

	plotter := blob.NewPlotter(s.debugVertexData)
	for _, line := range ctx.DebugLines {
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
	vertexData := s.debugVertexData[:plotter.Offset()]
	s.api.Queue().WriteBuffer(s.debugVertexBuffer, s.debugVertexBufferOffset, vertexData)

	commandBuffer := s.data.CommandBuffer()
	commandBuffer.BindPipeline(s.debugPipeline)
	commandBuffer.UniformBufferUnit(
		internal.UniformBufferBindingCamera,
		ctx.CameraPlacement.Buffer,
		ctx.CameraPlacement.Offset,
		ctx.CameraPlacement.Size,
	)
	commandBuffer.Draw(0, uint32(count)*2, 1)

	// Double buffer the vertex buffer through offset and hope the driver
	// is smart enough to figure this out.
	if s.debugVertexBufferOffset == 0 {
		s.debugVertexBufferOffset = uint32(len(s.debugVertexData)) / 2
	} else {
		s.debugVertexBufferOffset = 0
	}
}

func (s *ForwardStage) renderMeshes(ctx StageContext) {
	s.meshRenderer.DiscardRenderItems()
	for _, mesh := range ctx.VisibleMeshes {
		s.meshRenderer.QueueMeshRenderItems(ctx, mesh, internal.MeshRenderPassTypeForward)
	}
	for _, meshIndex := range ctx.VisibleStaticMeshIndices {
		staticMesh := &ctx.Scene.staticMeshes[meshIndex]
		s.meshRenderer.QueueStaticMeshRenderItems(ctx, staticMesh, internal.MeshRenderPassTypeForward)
	}
	s.meshRenderer.Render(ctx)
}

func (s *ForwardStage) findActiveSky(skies *ds.List[*Sky]) *Sky {
	for _, sky := range skies.Unbox() {
		if sky.Active() {
			return sky
		}
	}
	return nil
}
