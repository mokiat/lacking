package graphics

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/debug/metric"
	"github.com/mokiat/lacking/game/graphics/internal"
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/render/ubo"
)

type ToneMappingStageInput struct {
	HDRTexture   StageTextureParameter
	BloomTexture opt.T[StageTextureParameter]
}

func newToneMappingStage(api render.API, shaders ShaderCollection, data *commonStageData, input ToneMappingStageInput) *ToneMappingStage {
	return &ToneMappingStage{
		api:     api,
		shaders: shaders,
		data:    data,

		toneMapping: ExponentialToneMapping,

		inHDRTexture:   input.HDRTexture,
		inBloomTexture: input.BloomTexture.ValueOrDefault(nil),
	}
}

var _ Stage = (*ToneMappingStage)(nil)

// ToneMappingStage is a stage that applies tone mapping to the input
// HDR texture and outputs the result to the framebuffer.
type ToneMappingStage struct {
	api     render.API
	shaders ShaderCollection
	data    *commonStageData

	program  render.Program
	pipeline render.Pipeline

	toneMapping ToneMapping

	inHDRTexture   StageTextureParameter
	inBloomTexture StageTextureParameter
}

// ToneMapping represents the tone mapping algorithm that should be used
// when rendering the output of the stage.
func (s *ToneMappingStage) ToneMapping() ToneMapping {
	return s.toneMapping
}

// SetToneMapping sets the tone mapping algorithm that should be used
// when rendering the output of the stage.
func (s *ToneMappingStage) SetToneMapping(toneMapping ToneMapping) {
	s.toneMapping = toneMapping
}

func (s *ToneMappingStage) Allocate() {
	quadShape := s.data.QuadShape()

	s.program = s.api.CreateProgram(render.ProgramInfo{
		SourceCode: s.shaders.PostprocessingSet(PostprocessingShaderConfig{
			ToneMapping: s.toneMapping,
			Bloom:       s.inBloomTexture != nil,
		}),
		TextureBindings: []render.TextureBinding{
			render.NewTextureBinding("fbColor0TextureIn", internal.TextureBindingPostprocessFramebufferColor0),
			render.NewTextureBinding("lackingBloomTexture", internal.TextureBindingPostprocessBloom),
		},
		UniformBindings: []render.UniformBinding{
			render.NewUniformBinding("Postprocess", internal.UniformBufferBindingPostprocess),
		},
	})
	s.pipeline = s.api.CreatePipeline(render.PipelineInfo{
		Program:         s.program,
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
}

func (s *ToneMappingStage) Release() {
	defer s.program.Release()
	defer s.pipeline.Release()
}

func (s *ToneMappingStage) PreRender(width, height uint32) {
	// Nothing to do here.
}

func (s *ToneMappingStage) Render(ctx StageContext) {
	// TODO: Use built-in tracing. It does not actually allocate memory
	// when tracing is not enabled. Furthermore, the context can be Background
	// and nesting should still work. The context is used for tasks.
	defer metric.BeginRegion("tone-mapping").End()

	quadShape := s.data.QuadShape()
	nearestSampler := s.data.NearestSampler()
	linearSampler := s.data.LinearSampler()

	commandBuffer := ctx.CommandBuffer
	uniformBuffer := ctx.UniformBuffer

	postprocessPlacement := ubo.WriteUniform(uniformBuffer, internal.PostprocessUniform{
		Exposure: ctx.Camera.Exposure(),
	})

	commandBuffer.BeginRenderPass(render.RenderPassInfo{
		Framebuffer:    ctx.Framebuffer,
		Viewport:       ctx.Viewport,
		DepthLoadOp:    render.LoadOperationLoad,
		DepthStoreOp:   render.StoreOperationDiscard,
		StencilLoadOp:  render.LoadOperationLoad,
		StencilStoreOp: render.StoreOperationDiscard,
		Colors: [4]render.ColorAttachmentInfo{
			{
				LoadOp:  render.LoadOperationLoad,
				StoreOp: render.StoreOperationStore,
			},
		},
	})

	commandBuffer.BindPipeline(s.pipeline)
	commandBuffer.TextureUnit(internal.TextureBindingPostprocessFramebufferColor0, s.inHDRTexture())
	commandBuffer.SamplerUnit(internal.TextureBindingPostprocessFramebufferColor0, nearestSampler)
	if s.inBloomTexture != nil {
		commandBuffer.TextureUnit(internal.TextureBindingPostprocessBloom, s.inBloomTexture())
		commandBuffer.SamplerUnit(internal.TextureBindingPostprocessBloom, linearSampler)
	}
	commandBuffer.UniformBufferUnit(
		internal.UniformBufferBindingPostprocess,
		postprocessPlacement.Buffer,
		postprocessPlacement.Offset,
		postprocessPlacement.Size,
	)
	commandBuffer.DrawIndexed(0, quadShape.IndexCount(), 1)

	commandBuffer.EndRenderPass()
}

func (s *ToneMappingStage) PostRender() {
	// Nothing to do here.
}
