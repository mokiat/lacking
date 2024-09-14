package graphics

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/debug/metric"
	"github.com/mokiat/lacking/game/graphics/internal"
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/render/ubo"
)

const (
	bloomDownsampleHDRImageSlot = 0
	bloomBlurSourceImageSlot    = 0

	bloomDefaultBlurIterations = 2
)

// BloomStageInput is used to configure a BloomStage.
type BloomStageInput struct {
	HDRTexture StageTextureParameter
}

func newBloomStage(api render.API, shaders ShaderCollection, data *commonStageData, input BloomStageInput) *BloomStage {
	return &BloomStage{
		api:     api,
		shaders: shaders,
		data:    data,

		inHDRTexture: input.HDRTexture,
		iterations:   bloomDefaultBlurIterations,
	}
}

var _ Stage = (*BloomStage)(nil)

// BloomStage is a stage that produces a bloom overlay texture.
type BloomStage struct {
	api     render.API
	shaders ShaderCollection
	data    *commonStageData

	inHDRTexture StageTextureParameter
	iterations   int

	framebufferWidth  uint32
	framebufferHeight uint32

	pingFramebuffer render.Framebuffer
	pingTexture     render.Texture

	pongFramebuffer render.Framebuffer
	pongTexture     render.Texture

	downsampleProgram  render.Program
	downsamplePipeline render.Pipeline
	downsampleSampler  render.Sampler

	blurProgram  render.Program
	blurPipeline render.Pipeline
	blurSampler  render.Sampler
}

// Iterations returns the number of blur iterations that are
// performed on the bloom overlay texture.
func (s *BloomStage) Iterations() int {
	return s.iterations
}

// SetIterations sets the number of blur iterations that should be
// performed on the bloom overlay texture.
func (s *BloomStage) SetIterations(iterations int) {
	s.iterations = iterations
}

// BloomTexture returns the texture that contains the bloom overlay.
func (s *BloomStage) BloomTexture() render.Texture {
	return s.pingTexture
}

func (s *BloomStage) Allocate() {
	quadShape := s.data.QuadShape()

	s.allocateTextures(1, 1)

	s.downsampleProgram = s.api.CreateProgram(render.ProgramInfo{
		Label:      "Bloom Downsample Program",
		SourceCode: s.shaders.BloomDownsampleSet(),
		TextureBindings: []render.TextureBinding{
			render.NewTextureBinding("lackingSourceImage", bloomDownsampleHDRImageSlot),
		},
	})
	s.downsamplePipeline = s.api.CreatePipeline(render.PipelineInfo{
		Label:           "Bloom Downsample Pipeline",
		Program:         s.downsampleProgram,
		VertexArray:     quadShape.VertexArray(),
		Topology:        quadShape.Topology(),
		Culling:         render.CullModeBack,
		FrontFace:       render.FaceOrientationCCW,
		DepthTest:       false,
		DepthWrite:      false,
		DepthComparison: render.ComparisonAlways,
		StencilTest:     false,
		ColorWrite:      [4]bool{true, true, true, true},
		BlendEnabled:    false,
	})
	s.downsampleSampler = s.api.CreateSampler(render.SamplerInfo{
		Label:      "Bloom Downsample Sampler",
		Wrapping:   render.WrapModeClamp,
		Filtering:  render.FilterModeNearest,
		Mipmapping: false,
	})

	s.blurProgram = s.api.CreateProgram(render.ProgramInfo{
		Label:      "Bloom Blur Program",
		SourceCode: s.shaders.BloomBlurSet(),
		TextureBindings: []render.TextureBinding{
			render.NewTextureBinding("lackingSourceImage", bloomBlurSourceImageSlot),
		},
		UniformBindings: []render.UniformBinding{
			render.NewUniformBinding("BloomBlurData", internal.UniformBufferBindingBloom),
		},
	})
	s.blurPipeline = s.api.CreatePipeline(render.PipelineInfo{
		Label:           "Bloom Blur Pipeline",
		Program:         s.blurProgram,
		VertexArray:     quadShape.VertexArray(),
		Topology:        quadShape.Topology(),
		Culling:         render.CullModeBack,
		FrontFace:       render.FaceOrientationCCW,
		DepthTest:       false,
		DepthWrite:      false,
		DepthComparison: render.ComparisonAlways,
		StencilTest:     false,
		ColorWrite:      [4]bool{true, true, true, true},
		BlendEnabled:    false,
	})
	s.blurSampler = s.api.CreateSampler(render.SamplerInfo{
		Wrapping:   render.WrapModeClamp,
		Filtering:  render.FilterModeNearest,
		Mipmapping: false,
	})
}

func (s *BloomStage) Release() {
	defer s.releaseTextures()

	defer s.downsampleProgram.Release()
	defer s.downsamplePipeline.Release()
	defer s.downsampleSampler.Release()

	defer s.blurProgram.Release()
	defer s.blurPipeline.Release()
	defer s.blurSampler.Release()
}

func (s *BloomStage) PreRender(width, height uint32) {
	targetWidth := max(1, width/2)
	targetHeight := max(1, height/2)
	if s.framebufferWidth != targetWidth || s.framebufferHeight != targetHeight {
		s.releaseTextures()
		s.allocateTextures(targetWidth, targetHeight)
	}
}

func (s *BloomStage) Render(ctx StageContext) {
	defer metric.BeginRegion("bloom").End()

	quadShape := s.data.QuadShape()
	commandBuffer := ctx.CommandBuffer
	uniformBuffer := ctx.UniformBuffer

	// Perform a downsample into the Ping texture
	commandBuffer.BeginRenderPass(render.RenderPassInfo{
		Framebuffer: s.pingFramebuffer,
		Viewport: render.Area{
			Width:  s.framebufferWidth,
			Height: s.framebufferHeight,
		},
		DepthLoadOp:    render.LoadOperationLoad,
		DepthStoreOp:   render.StoreOperationDiscard,
		StencilLoadOp:  render.LoadOperationLoad,
		StencilStoreOp: render.StoreOperationDiscard,
		Colors: [4]render.ColorAttachmentInfo{
			{
				LoadOp:     render.LoadOperationLoad,
				StoreOp:    render.StoreOperationStore,
				ClearValue: [4]float32{0.0, 0.0, 0.0, 0.0},
			},
		},
	})
	commandBuffer.BindPipeline(s.downsamplePipeline)
	commandBuffer.TextureUnit(bloomDownsampleHDRImageSlot, s.inHDRTexture())
	commandBuffer.SamplerUnit(bloomDownsampleHDRImageSlot, s.downsampleSampler)
	commandBuffer.DrawIndexed(0, quadShape.IndexCount(), 1)
	commandBuffer.EndRenderPass()

	// Perform blur passes
	horizontal := float32(1.0)
	for range s.iterations * 2 { // times 2 because we do horizontal and vertical passes
		commandBuffer.BeginRenderPass(render.RenderPassInfo{
			Framebuffer: s.pongFramebuffer,
			Viewport: render.Area{
				Width:  s.framebufferWidth,
				Height: s.framebufferHeight,
			},
			DepthLoadOp:    render.LoadOperationLoad,
			DepthStoreOp:   render.StoreOperationDiscard,
			StencilLoadOp:  render.LoadOperationLoad,
			StencilStoreOp: render.StoreOperationDiscard,
			Colors: [4]render.ColorAttachmentInfo{
				{
					LoadOp:     render.LoadOperationLoad,
					StoreOp:    render.StoreOperationStore,
					ClearValue: [4]float32{0.0, 0.0, 0.0, 0.0},
				},
			},
		})
		commandBuffer.BindPipeline(s.blurPipeline)
		commandBuffer.TextureUnit(bloomBlurSourceImageSlot, s.pingTexture)
		commandBuffer.SamplerUnit(bloomBlurSourceImageSlot, s.blurSampler)
		uniformPlacement := ubo.WriteUniform(uniformBuffer, internal.BloomBlurUniform{
			Horizontal: horizontal,
		})
		commandBuffer.UniformBufferUnit(internal.UniformBufferBindingBloom, uniformPlacement.Buffer, uniformPlacement.Offset, uniformPlacement.Size)
		commandBuffer.DrawIndexed(0, quadShape.IndexCount(), 1)
		commandBuffer.EndRenderPass()

		s.pingFramebuffer, s.pongFramebuffer = s.pongFramebuffer, s.pingFramebuffer
		s.pingTexture, s.pongTexture = s.pongTexture, s.pingTexture
		horizontal = 1.0 - horizontal
	}
}

func (s *BloomStage) PostRender() {
	// Nothing to do here.
}

func (s *BloomStage) allocateTextures(width, height uint32) {
	s.framebufferWidth = width
	s.framebufferHeight = height

	s.pingTexture = s.api.CreateColorTexture2D(render.ColorTexture2DInfo{
		Label:           "Bloom Ping Texture",
		Width:           width,
		Height:          height,
		GenerateMipmaps: false,
		GammaCorrection: false,
		Format:          render.DataFormatRGBA16F,
	})
	s.pingFramebuffer = s.api.CreateFramebuffer(render.FramebufferInfo{
		Label: "Bloom Ping Framebuffer",
		ColorAttachments: [4]opt.T[render.TextureAttachment]{
			opt.V(render.PlainTextureAttachment(s.pingTexture)),
		},
	})

	s.pongTexture = s.api.CreateColorTexture2D(render.ColorTexture2DInfo{
		Label:           "Bloom Pong Texture",
		Width:           width,
		Height:          height,
		GenerateMipmaps: false,
		GammaCorrection: false,
		Format:          render.DataFormatRGBA16F,
	})
	s.pongFramebuffer = s.api.CreateFramebuffer(render.FramebufferInfo{
		Label: "Bloom Pong Framebuffer",
		ColorAttachments: [4]opt.T[render.TextureAttachment]{
			opt.V(render.PlainTextureAttachment(s.pongTexture)),
		},
	})
}

func (s *BloomStage) releaseTextures() {
	defer s.pingFramebuffer.Release()
	defer s.pingTexture.Release()

	defer s.pongFramebuffer.Release()
	defer s.pongTexture.Release()
}
