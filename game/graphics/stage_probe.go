package graphics

import (
	"time"

	"github.com/mokiat/gblob"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/debug/metric"
	"github.com/mokiat/lacking/game/graphics/internal"
	"github.com/mokiat/lacking/render"
	"github.com/x448/float16"
)

// ExposureProbeStageInput is used to configure an ExposureProbeStage.
type ExposureProbeStageInput struct {
	HDRTexture StageTextureParameter
}

func newExposureProbeStage(api render.API, shaders ShaderCollection, data *commonStageData, input ExposureProbeStageInput) *ExposureProbeStage {
	return &ExposureProbeStage{
		api:     api,
		shaders: shaders,
		data:    data,

		hdrTexture: input.HDRTexture,

		exposureBufferData: make([]byte, 4*render.SizeF32), // worst case RGBA32F
		exposureTarget:     1.0,
	}
}

var _ Stage = (*ExposureProbeStage)(nil)

// ExposureProbeStage is a stage that measures the brightness of the scene
// and adjusts the exposure of the camera accordingly.
type ExposureProbeStage struct {
	api     render.API
	shaders ShaderCollection
	data    *commonStageData

	hdrTexture StageTextureParameter

	exposureAlbedoTexture   render.Texture
	exposureFramebuffer     render.Framebuffer
	exposureFormat          render.DataFormat
	exposureProgram         render.Program
	exposurePipeline        render.Pipeline
	exposureBufferData      gblob.LittleEndianBlock
	exposureBuffer          render.Buffer
	exposureSync            render.Fence
	exposureSyncNeeded      bool
	exposureTarget          float32
	exposureUpdateTimestamp time.Time
}

func (s *ExposureProbeStage) Allocate() {
	quadShape := s.data.QuadShape()

	s.exposureAlbedoTexture = s.api.CreateColorTexture2D(render.ColorTexture2DInfo{
		Width:           1,
		Height:          1,
		GenerateMipmaps: false,
		GammaCorrection: false,
		Format:          render.DataFormatRGBA16F,
	})
	s.exposureFramebuffer = s.api.CreateFramebuffer(render.FramebufferInfo{
		ColorAttachments: [4]render.Texture{
			s.exposureAlbedoTexture,
		},
	})

	s.exposureProgram = s.api.CreateProgram(render.ProgramInfo{
		SourceCode: s.shaders.ExposureSet(),
		TextureBindings: []render.TextureBinding{
			render.NewTextureBinding("fbColor0TextureIn", internal.TextureBindingLightingFramebufferColor0),
		},
		UniformBindings: []render.UniformBinding{},
	})
	s.exposurePipeline = s.api.CreatePipeline(render.PipelineInfo{
		Program:      s.exposureProgram,
		VertexArray:  quadShape.VertexArray(),
		Topology:     quadShape.Topology(),
		Culling:      render.CullModeBack,
		FrontFace:    render.FaceOrientationCCW,
		DepthTest:    false,
		DepthWrite:   false,
		StencilTest:  false,
		ColorWrite:   render.ColorMaskTrue,
		BlendEnabled: false,
	})

	s.exposureBuffer = s.api.CreatePixelTransferBuffer(render.BufferInfo{
		Dynamic: true,
		Size:    uint32(len(s.exposureBufferData)),
	})
	s.exposureFormat = s.api.DetermineContentFormat(s.exposureFramebuffer)
	if s.exposureFormat == render.DataFormatUnsupported {
		// This happens on MacOS on native; fallback to a default format and
		// hope for the best.
		s.exposureFormat = render.DataFormatRGBA32F
	}
}

func (s *ExposureProbeStage) Release() {
	defer s.exposureAlbedoTexture.Release()
	defer s.exposureFramebuffer.Release()

	defer s.exposureProgram.Release()
	defer s.exposurePipeline.Release()

	defer s.exposureBuffer.Release()
}

func (s *ExposureProbeStage) PreRender(width, height uint32) {
	// Nothing to do here.
}

func (s *ExposureProbeStage) Render(ctx StageContext) {
	if !ctx.Camera.AutoExposure() {
		return
	}

	defer metric.BeginRegion("exposure").End()

	if s.exposureSync != nil && s.exposureSync.Status() == render.FenceStatusSuccess {
		s.api.Queue().ReadBuffer(s.exposureBuffer, 0, s.exposureBufferData)

		var brightness float32
		switch s.exposureFormat {
		case render.DataFormatRGBA16F:
			brightness = float16.Frombits(s.exposureBufferData.Uint16(0)).Float32()
		case render.DataFormatRGBA32F:
			brightness = s.exposureBufferData.Float32(0)
		}
		brightness = sprec.Clamp(brightness, 0.001, 1000.0)

		s.exposureTarget = 1.0 / (2 * 3.14 * brightness)
		s.exposureTarget = sprec.Clamp(s.exposureTarget, ctx.Camera.MinimumExposure(), ctx.Camera.MaximumExposure())

		s.exposureSync.Release()
		s.exposureSync = nil
	}

	if !s.exposureUpdateTimestamp.IsZero() {
		elapsedSeconds := float32(time.Since(s.exposureUpdateTimestamp).Seconds())
		currentExposure := ctx.Camera.Exposure()
		alpha := sprec.Clamp(ctx.Camera.AutoExposureSpeed()*elapsedSeconds, 0.0, 1.0)
		ctx.Camera.SetExposure(sprec.Mix(currentExposure, s.exposureTarget, alpha))
	}
	s.exposureUpdateTimestamp = time.Now()

	if s.exposureSync == nil {
		s.exposureSyncNeeded = true
		quadShape := s.data.QuadShape()
		nearestSampler := s.data.NearestSampler()
		commandBuffer := ctx.CommandBuffer
		commandBuffer.BeginRenderPass(render.RenderPassInfo{
			Framebuffer: s.exposureFramebuffer,
			Viewport: render.Area{
				X:      0,
				Y:      0,
				Width:  1,
				Height: 1,
			},
			DepthLoadOp:    render.LoadOperationLoad,
			DepthStoreOp:   render.StoreOperationDiscard,
			StencilLoadOp:  render.LoadOperationLoad,
			StencilStoreOp: render.StoreOperationDiscard,
			Colors: [4]render.ColorAttachmentInfo{
				{
					LoadOp:     render.LoadOperationClear,
					StoreOp:    render.StoreOperationDiscard,
					ClearValue: [4]float32{0.0, 0.0, 0.0, 0.0},
				},
			},
		})
		commandBuffer.BindPipeline(s.exposurePipeline)
		commandBuffer.TextureUnit(internal.TextureBindingLightingFramebufferColor0, s.hdrTexture())
		commandBuffer.SamplerUnit(internal.TextureBindingLightingFramebufferColor0, nearestSampler)
		commandBuffer.UniformBufferUnit(
			internal.UniformBufferBindingCamera,
			ctx.CameraPlacement.Buffer,
			ctx.CameraPlacement.Offset,
			ctx.CameraPlacement.Size,
		)
		commandBuffer.DrawIndexed(0, quadShape.IndexCount(), 1)
		commandBuffer.CopyFramebufferToBuffer(render.CopyFramebufferToBufferInfo{
			Buffer: s.exposureBuffer,
			X:      0,
			Y:      0,
			Width:  1,
			Height: 1,
			Format: s.exposureFormat,
		})
		commandBuffer.EndRenderPass()
	}
}

func (s *ExposureProbeStage) PostRender() {
	if s.exposureSyncNeeded && s.exposureSync == nil {
		s.exposureSync = s.api.Queue().TrackSubmittedWorkDone()
	}
}
