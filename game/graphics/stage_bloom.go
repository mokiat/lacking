package graphics

import (
	"github.com/mokiat/lacking/game/graphics/internal"
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/render/ubo"
)

const (
	bloomDownsampleHDRImageSlot = 0

	bloomBlurSourceImageSlot = 0
)

const (
	blurIterations = 2
)

func newBloomRenderStage(api render.API, shaders ShaderCollection, data *commonStageData) *bloomRenderStage {
	return &bloomRenderStage{
		api:     api,
		shaders: shaders,
		data:    data,
	}
}

type bloomRenderStage struct {
	api     render.API
	shaders ShaderCollection
	data    *commonStageData

	framebufferWidth  uint32
	framebufferHeight uint32

	pingFramebuffer render.Framebuffer
	pingTexture     render.Texture

	pongFramebuffer render.Framebuffer
	pongTexture     render.Texture

	outputTexture render.Texture
	outputSampler render.Sampler

	downsampleProgram  render.Program
	downsamplePipeline render.Pipeline
	downsampleSampler  render.Sampler

	blurProgram  render.Program
	blurPipeline render.Pipeline
	blurSampler  render.Sampler
}

func (s *bloomRenderStage) Allocate() {
	quadShape := s.data.QuadShape()

	s.allocateTextures(1, 1)

	s.downsampleProgram = s.api.CreateProgram(render.ProgramInfo{
		SourceCode: s.shaders.BloomDownsampleSet(),
		TextureBindings: []render.TextureBinding{
			render.NewTextureBinding("lackingSourceImage", bloomDownsampleHDRImageSlot),
		},
	})
	s.downsamplePipeline = s.api.CreatePipeline(render.PipelineInfo{
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
		Wrapping:   render.WrapModeClamp,
		Filtering:  render.FilterModeNearest,
		Mipmapping: false,
	})

	s.blurProgram = s.api.CreateProgram(render.ProgramInfo{
		SourceCode: s.shaders.BloomBlurSet(),
		TextureBindings: []render.TextureBinding{
			render.NewTextureBinding("lackingSourceImage", bloomBlurSourceImageSlot),
		},
		UniformBindings: []render.UniformBinding{
			render.NewUniformBinding("BloomBlurData", internal.UniformBufferBindingBloom),
		},
	})
	s.blurPipeline = s.api.CreatePipeline(render.PipelineInfo{
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

	s.outputTexture = s.pingTexture
	s.outputSampler = s.api.CreateSampler(render.SamplerInfo{
		Wrapping:   render.WrapModeClamp,
		Filtering:  render.FilterModeLinear,
		Mipmapping: false,
	})
}

func (s *bloomRenderStage) Release() {
	defer s.releaseTextures()

	defer s.downsampleProgram.Release()
	defer s.downsamplePipeline.Release()
	defer s.downsampleSampler.Release()

	defer s.blurProgram.Release()
	defer s.blurPipeline.Release()
	defer s.blurSampler.Release()

	defer s.outputSampler.Release()
}

func (s *bloomRenderStage) Resize(width, height uint32) {
	width = max(1, width/2)
	height = max(1, height/2)

	s.releaseTextures()
	s.allocateTextures(width, height)
}

func (s *bloomRenderStage) Run(hdrImage render.Texture) {
	commandBuffer := s.data.CommandBuffer()
	uniformBuffer := s.data.UniformBuffer()
	quadShape := s.data.QuadShape()

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
	commandBuffer.TextureUnit(bloomDownsampleHDRImageSlot, hdrImage)
	commandBuffer.SamplerUnit(bloomDownsampleHDRImageSlot, s.downsampleSampler)
	commandBuffer.DrawIndexed(0, quadShape.IndexCount(), 1)
	commandBuffer.EndRenderPass()

	// Perform blur passes
	horizontal := float32(1.0)
	for range blurIterations * 2 {
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

	s.outputTexture = s.pingTexture
}

func (s *bloomRenderStage) OutputTexture() render.Texture {
	return s.outputTexture
}

func (s *bloomRenderStage) OutputSampler() render.Sampler {
	return s.outputSampler
}

func (s *bloomRenderStage) allocateTextures(width, height uint32) {
	s.framebufferWidth = width
	s.framebufferHeight = height

	s.pingTexture = s.api.CreateColorTexture2D(render.ColorTexture2DInfo{
		Width:           width,
		Height:          height,
		GenerateMipmaps: false,
		GammaCorrection: false,
		Format:          render.DataFormatRGBA16F,
	})
	s.pingFramebuffer = s.api.CreateFramebuffer(render.FramebufferInfo{
		ColorAttachments: [4]render.Texture{
			s.pingTexture,
		},
	})

	s.pongTexture = s.api.CreateColorTexture2D(render.ColorTexture2DInfo{
		Width:           width,
		Height:          height,
		GenerateMipmaps: false,
		GammaCorrection: false,
		Format:          render.DataFormatRGBA16F,
	})
	s.pongFramebuffer = s.api.CreateFramebuffer(render.FramebufferInfo{
		ColorAttachments: [4]render.Texture{
			s.pongTexture,
		},
	})

	s.outputTexture = s.pingTexture
}

func (s *bloomRenderStage) releaseTextures() {
	defer s.pingFramebuffer.Release()
	defer s.pingTexture.Release()

	defer s.pongFramebuffer.Release()
	defer s.pongTexture.Release()
}
