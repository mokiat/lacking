package graphics

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dtos"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/debug/metric"
	"github.com/mokiat/lacking/game/graphics/internal"
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/render/ubo"
)

// LightingStageInput is the input data for the LightingStage.
type LightingStageInput struct {
	AlbedoMetallicTexture  StageTextureParameter
	NormalRoughnessTexture StageTextureParameter
	DepthTexture           StageTextureParameter
	HDRTexture             StageTextureParameter
}

func newLightingStage(api render.API, shaders ShaderCollection, data *commonStageData, input LightingStageInput) *LightingStage {
	return &LightingStage{
		api:     api,
		shaders: shaders,
		data:    data,
		input:   input,
	}
}

var _ Stage = (*LightingStage)(nil)

// LightingStage is responsible for rendering the lighting of the scene.
type LightingStage struct {
	api     render.API
	shaders ShaderCollection
	data    *commonStageData
	input   LightingStageInput

	hdrTexture  render.Texture
	framebuffer render.Framebuffer

	noShadowTexture render.Texture

	ambientLightProgram  render.Program
	ambientLightPipeline render.Pipeline

	pointLightProgram  render.Program
	pointLightPipeline render.Pipeline

	spotLightProgram  render.Program
	spotLightPipeline render.Pipeline

	directionalLightProgram  render.Program
	directionalLightPipeline render.Pipeline
}

func (s *LightingStage) Allocate() {
	s.allocateFramebuffer()
	s.allocatePipelines()
}

func (s *LightingStage) Release() {
	defer s.releaseFramebuffer()
	defer s.releasePipelines()
}

func (s *LightingStage) PreRender(width, height uint32) {
	hdrTexture := s.input.HDRTexture()
	if hdrTexture != s.hdrTexture {
		s.releaseFramebuffer()
		s.allocateFramebuffer()
	}
}

func (s *LightingStage) Render(ctx StageContext) {
	defer metric.BeginRegion("lighting").End()

	commandBuffer := ctx.CommandBuffer
	commandBuffer.BeginRenderPass(render.RenderPassInfo{
		Framebuffer: s.framebuffer,
		Viewport: render.Area{
			Width:  s.hdrTexture.Width(),
			Height: s.hdrTexture.Height(),
		},
		DepthLoadOp:    render.LoadOperationLoad,
		DepthStoreOp:   render.StoreOperationStore,
		StencilLoadOp:  render.LoadOperationLoad,
		StencilStoreOp: render.StoreOperationDiscard,
		Colors: [4]render.ColorAttachmentInfo{
			{
				LoadOp:     render.LoadOperationClear,
				StoreOp:    render.StoreOperationStore,
				ClearValue: [4]float32{0.0, 0.0, 0.0, 1.0},
			},
		},
	})

	for _, ambientLight := range ctx.VisibleAmbientLights {
		if ambientLight.active {
			s.renderAmbientLight(ctx, ambientLight)
		}
	}
	for _, pointLight := range ctx.VisiblePointLights {
		if pointLight.active {
			s.renderPointLight(ctx, pointLight)
		}
	}
	for _, spotLight := range ctx.VisibleSpotLights {
		if spotLight.active {
			s.renderSpotLight(ctx, spotLight)
		}
	}
	for _, directionalLight := range ctx.VisibleDirectionalLights {
		if directionalLight.active {
			s.renderDirectionalLight(ctx, directionalLight)
		}
	}

	commandBuffer.EndRenderPass()
}

func (s *LightingStage) PostRender() {
	// Nothing to do here.
}

func (s *LightingStage) allocateFramebuffer() {
	s.hdrTexture = s.input.HDRTexture()

	s.framebuffer = s.api.CreateFramebuffer(render.FramebufferInfo{
		Label: "Lighting Framebuffer",
		ColorAttachments: [4]opt.T[render.TextureAttachment]{
			opt.V(render.PlainTextureAttachment(s.hdrTexture)),
		},
	})
}

func (s *LightingStage) releaseFramebuffer() {
	defer s.framebuffer.Release()
}

func (s *LightingStage) allocatePipelines() {
	quadShape := s.data.QuadShape()
	sphereShape := s.data.SphereShape()
	coneShape := s.data.ConeShape()

	s.noShadowTexture = s.api.CreateDepthTexture2D(render.DepthTexture2DInfo{
		Label:      "No Shadow Texture",
		Width:      1,
		Height:     1,
		Comparable: true,
		// TODO: Initialize to furthest possible depth value
	})

	s.ambientLightProgram = s.api.CreateProgram(render.ProgramInfo{
		Label:      "Ambient Light Program",
		SourceCode: s.shaders.AmbientLightSet(),
		TextureBindings: []render.TextureBinding{
			render.NewTextureBinding("fbColor0TextureIn", internal.TextureBindingLightingFramebufferColor0),
			render.NewTextureBinding("fbColor1TextureIn", internal.TextureBindingLightingFramebufferColor1),
			render.NewTextureBinding("fbDepthTextureIn", internal.TextureBindingLightingFramebufferDepth),
			render.NewTextureBinding("reflectionTextureIn", internal.TextureBindingLightingReflectionTexture),
			render.NewTextureBinding("refractionTextureIn", internal.TextureBindingLightingRefractionTexture),
		},
		UniformBindings: []render.UniformBinding{
			render.NewUniformBinding("Camera", internal.UniformBufferBindingCamera),
		},
	})
	s.ambientLightPipeline = s.api.CreatePipeline(render.PipelineInfo{
		Label:                       "Ambient Light Pipeline",
		Program:                     s.ambientLightProgram,
		VertexArray:                 quadShape.VertexArray(),
		Topology:                    quadShape.Topology(),
		Culling:                     render.CullModeBack,
		FrontFace:                   render.FaceOrientationCCW,
		DepthTest:                   false,
		DepthWrite:                  false,
		DepthComparison:             render.ComparisonAlways,
		StencilTest:                 false,
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

	s.pointLightProgram = s.api.CreateProgram(render.ProgramInfo{
		Label:      "Point Light Program",
		SourceCode: s.shaders.PointLightSet(),
		TextureBindings: []render.TextureBinding{
			render.NewTextureBinding("fbColor0TextureIn", internal.TextureBindingLightingFramebufferColor0),
			render.NewTextureBinding("fbColor1TextureIn", internal.TextureBindingLightingFramebufferColor1),
			render.NewTextureBinding("fbDepthTextureIn", internal.TextureBindingLightingFramebufferDepth),
		},
		UniformBindings: []render.UniformBinding{
			render.NewUniformBinding("Camera", internal.UniformBufferBindingCamera),
			render.NewUniformBinding("Light", internal.UniformBufferBindingLight),
		},
	})
	s.pointLightPipeline = s.api.CreatePipeline(render.PipelineInfo{
		Label:                       "Point Light Pipeline",
		Program:                     s.pointLightProgram,
		VertexArray:                 sphereShape.VertexArray(),
		Topology:                    sphereShape.Topology(),
		Culling:                     render.CullModeFront,
		FrontFace:                   render.FaceOrientationCCW,
		DepthTest:                   false,
		DepthWrite:                  false,
		DepthComparison:             render.ComparisonAlways,
		StencilTest:                 false,
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

	s.spotLightProgram = s.api.CreateProgram(render.ProgramInfo{
		Label:      "Spot Light Program",
		SourceCode: s.shaders.SpotLightSet(),
		TextureBindings: []render.TextureBinding{
			render.NewTextureBinding("fbColor0TextureIn", internal.TextureBindingLightingFramebufferColor0),
			render.NewTextureBinding("fbColor1TextureIn", internal.TextureBindingLightingFramebufferColor1),
			render.NewTextureBinding("fbDepthTextureIn", internal.TextureBindingLightingFramebufferDepth),
		},
		UniformBindings: []render.UniformBinding{
			render.NewUniformBinding("Camera", internal.UniformBufferBindingCamera),
			render.NewUniformBinding("Light", internal.UniformBufferBindingLight),
		},
	})
	s.spotLightPipeline = s.api.CreatePipeline(render.PipelineInfo{
		Label:                       "Spot Light Pipeline",
		Program:                     s.spotLightProgram,
		VertexArray:                 coneShape.VertexArray(),
		Topology:                    coneShape.Topology(),
		Culling:                     render.CullModeFront,
		FrontFace:                   render.FaceOrientationCCW,
		DepthTest:                   false,
		DepthWrite:                  false,
		DepthComparison:             render.ComparisonAlways,
		StencilTest:                 false,
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

	s.directionalLightProgram = s.api.CreateProgram(render.ProgramInfo{
		Label:      "Directional Light Program",
		SourceCode: s.shaders.DirectionalLightSet(),
		TextureBindings: []render.TextureBinding{
			render.NewTextureBinding("fbColor0TextureIn", internal.TextureBindingLightingFramebufferColor0),
			render.NewTextureBinding("fbColor1TextureIn", internal.TextureBindingLightingFramebufferColor1),
			render.NewTextureBinding("fbDepthTextureIn", internal.TextureBindingLightingFramebufferDepth),
			render.NewTextureBinding("lackingShadowMap", internal.TextureBindingLightingShadowMap),
		},
		UniformBindings: []render.UniformBinding{
			render.NewUniformBinding("Camera", internal.UniformBufferBindingCamera),
			render.NewUniformBinding("Light", internal.UniformBufferBindingLight),
		},
	})
	s.directionalLightPipeline = s.api.CreatePipeline(render.PipelineInfo{
		Label:                       "Directional Light Pipeline",
		Program:                     s.directionalLightProgram,
		VertexArray:                 quadShape.VertexArray(),
		Topology:                    quadShape.Topology(),
		Culling:                     render.CullModeBack,
		FrontFace:                   render.FaceOrientationCCW,
		DepthTest:                   false,
		DepthWrite:                  false,
		DepthComparison:             render.ComparisonAlways,
		StencilTest:                 false,
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
}

func (s *LightingStage) releasePipelines() {
	defer s.ambientLightProgram.Release()
	defer s.ambientLightPipeline.Release()

	defer s.pointLightProgram.Release()
	defer s.pointLightPipeline.Release()

	defer s.spotLightProgram.Release()
	defer s.spotLightPipeline.Release()

	defer s.directionalLightProgram.Release()
	defer s.directionalLightPipeline.Release()
}

func (s *LightingStage) renderAmbientLight(ctx StageContext, light *AmbientLight) {
	quadShape := s.data.QuadShape()

	nearestSampler := s.data.NearestSampler()
	linearSampler := s.data.LinearSampler()
	mipmapSampler := s.data.MipmapSampler()

	albedoMetallicTexture := s.input.AlbedoMetallicTexture()
	normalRoughnessTexture := s.input.NormalRoughnessTexture()
	depthTexture := s.input.DepthTexture()

	commandBuffer := ctx.CommandBuffer
	// TODO: Ambient light intensity based on distance and inner and outer radius
	commandBuffer.BindPipeline(s.ambientLightPipeline)
	commandBuffer.TextureUnit(internal.TextureBindingLightingFramebufferColor0, albedoMetallicTexture)
	commandBuffer.SamplerUnit(internal.TextureBindingLightingFramebufferColor0, nearestSampler)
	commandBuffer.TextureUnit(internal.TextureBindingLightingFramebufferColor1, normalRoughnessTexture)
	commandBuffer.SamplerUnit(internal.TextureBindingLightingFramebufferColor1, nearestSampler)
	commandBuffer.TextureUnit(internal.TextureBindingLightingFramebufferDepth, depthTexture)
	commandBuffer.SamplerUnit(internal.TextureBindingLightingFramebufferDepth, nearestSampler)
	commandBuffer.TextureUnit(internal.TextureBindingLightingReflectionTexture, light.reflectionTexture)
	commandBuffer.SamplerUnit(internal.TextureBindingLightingReflectionTexture, mipmapSampler)
	commandBuffer.TextureUnit(internal.TextureBindingLightingRefractionTexture, light.refractionTexture)
	commandBuffer.SamplerUnit(internal.TextureBindingLightingRefractionTexture, linearSampler)
	commandBuffer.UniformBufferUnit(
		internal.UniformBufferBindingCamera,
		ctx.CameraPlacement.Buffer,
		ctx.CameraPlacement.Offset,
		ctx.CameraPlacement.Size,
	)
	commandBuffer.DrawIndexed(0, quadShape.IndexCount(), 1)
}

func (s *LightingStage) renderPointLight(ctx StageContext, light *PointLight) {
	sphereShape := s.data.SphereShape()
	nearestSampler := s.data.NearestSampler()

	commandBuffer := ctx.CommandBuffer
	uniformBuffer := ctx.UniformBuffer

	albedoMetallicTexture := s.input.AlbedoMetallicTexture()
	normalRoughnessTexture := s.input.NormalRoughnessTexture()
	depthTexture := s.input.DepthTexture()

	lightPlacement := ubo.WriteUniform(uniformBuffer, internal.LightUniform{
		ShadowMatrices: [4]sprec.Mat4{}, // irrelevant
		ShadowCascades: [4]sprec.Vec2{}, // irrelevant
		ModelMatrix:    light.gfxMatrix(),
		Color:          dtos.Vec3(light.emitColor),
		Intensity:      1.0,
		Range:          float32(light.emitRange),
	})

	commandBuffer.BindPipeline(s.pointLightPipeline)
	commandBuffer.TextureUnit(internal.TextureBindingLightingFramebufferColor0, albedoMetallicTexture)
	commandBuffer.SamplerUnit(internal.TextureBindingLightingFramebufferColor0, nearestSampler)
	commandBuffer.TextureUnit(internal.TextureBindingLightingFramebufferColor1, normalRoughnessTexture)
	commandBuffer.SamplerUnit(internal.TextureBindingLightingFramebufferColor1, nearestSampler)
	commandBuffer.TextureUnit(internal.TextureBindingLightingFramebufferDepth, depthTexture)
	commandBuffer.SamplerUnit(internal.TextureBindingLightingFramebufferDepth, nearestSampler)
	commandBuffer.UniformBufferUnit(
		internal.UniformBufferBindingCamera,
		ctx.CameraPlacement.Buffer,
		ctx.CameraPlacement.Offset,
		ctx.CameraPlacement.Size,
	)
	commandBuffer.UniformBufferUnit(
		internal.UniformBufferBindingLight,
		lightPlacement.Buffer,
		lightPlacement.Offset,
		lightPlacement.Size,
	)
	commandBuffer.DrawIndexed(0, sphereShape.IndexCount(), 1)
}

func (s *LightingStage) renderSpotLight(ctx StageContext, light *SpotLight) {
	coneShape := s.data.ConeShape()
	nearestSampler := s.data.NearestSampler()

	uniformBuffer := ctx.UniformBuffer
	commandBuffer := ctx.CommandBuffer

	albedoMetallicTexture := s.input.AlbedoMetallicTexture()
	normalRoughnessTexture := s.input.NormalRoughnessTexture()
	depthTexture := s.input.DepthTexture()

	lightPlacement := ubo.WriteUniform(uniformBuffer, internal.LightUniform{
		ShadowMatrices: [4]sprec.Mat4{}, // irrelevant
		ShadowCascades: [4]sprec.Vec2{}, // irrelevant
		ModelMatrix:    light.gfxMatrix(),
		Color:          dtos.Vec3(light.emitColor),
		Intensity:      1.0,
		Range:          float32(light.emitRange),
		OuterAngle:     float32(light.emitOuterConeAngle.Radians()),
		InnerAngle:     float32(light.emitInnerConeAngle.Radians()),
	})

	commandBuffer.BindPipeline(s.spotLightPipeline)
	commandBuffer.TextureUnit(internal.TextureBindingLightingFramebufferColor0, albedoMetallicTexture)
	commandBuffer.SamplerUnit(internal.TextureBindingLightingFramebufferColor0, nearestSampler)
	commandBuffer.TextureUnit(internal.TextureBindingLightingFramebufferColor1, normalRoughnessTexture)
	commandBuffer.SamplerUnit(internal.TextureBindingLightingFramebufferColor1, nearestSampler)
	commandBuffer.TextureUnit(internal.TextureBindingLightingFramebufferDepth, depthTexture)
	commandBuffer.SamplerUnit(internal.TextureBindingLightingFramebufferDepth, nearestSampler)
	commandBuffer.UniformBufferUnit(
		internal.UniformBufferBindingCamera,
		ctx.CameraPlacement.Buffer,
		ctx.CameraPlacement.Offset,
		ctx.CameraPlacement.Size,
	)
	commandBuffer.UniformBufferUnit(
		internal.UniformBufferBindingLight,
		lightPlacement.Buffer,
		lightPlacement.Offset,
		lightPlacement.Size,
	)
	commandBuffer.DrawIndexed(0, coneShape.IndexCount(), 1)
}

func (s *LightingStage) renderDirectionalLight(ctx StageContext, light *DirectionalLight) {
	quadShape := s.data.QuadShape()
	nearestSampler := s.data.NearestSampler()
	depthSampler := s.data.DepthSampler()

	uniformBuffer := ctx.UniformBuffer
	commandBuffer := ctx.CommandBuffer

	albedoMetallicTexture := s.input.AlbedoMetallicTexture()
	normalRoughnessTexture := s.input.NormalRoughnessTexture()
	depthTexture := s.input.DepthTexture()

	lightUniform := internal.LightUniform{
		ShadowMatrices: [4]sprec.Mat4{},
		ShadowCascades: [4]sprec.Vec2{},
		ModelMatrix:    light.gfxMatrix(),
		Color:          dtos.Vec3(light.emitColor),
		Intensity:      1.0,
		Range:          0.0, // irrelevant
		OuterAngle:     0.0, // irrelevant
		InnerAngle:     0.0, // irrelevant
	}

	shadowTexture := s.noShadowTexture
	if shadowMap := s.data.GetDirectionalShadowMap(light); shadowMap != nil {
		shadowTexture = shadowMap.ArrayTexture
		for i, cascade := range shadowMap.Cascades {
			lightUniform.ShadowMatrices[i] = cascade.ProjectionMatrix
			lightUniform.ShadowCascades[i] = sprec.NewVec2(cascade.Near, cascade.Far)
		}
	}

	lightPlacement := ubo.WriteUniform(uniformBuffer, lightUniform)

	commandBuffer.BindPipeline(s.directionalLightPipeline)
	commandBuffer.TextureUnit(internal.TextureBindingLightingFramebufferColor0, albedoMetallicTexture)
	commandBuffer.SamplerUnit(internal.TextureBindingLightingFramebufferColor0, nearestSampler)
	commandBuffer.TextureUnit(internal.TextureBindingLightingFramebufferColor1, normalRoughnessTexture)
	commandBuffer.SamplerUnit(internal.TextureBindingLightingFramebufferColor1, nearestSampler)
	commandBuffer.TextureUnit(internal.TextureBindingLightingFramebufferDepth, depthTexture)
	commandBuffer.SamplerUnit(internal.TextureBindingLightingFramebufferDepth, nearestSampler)
	commandBuffer.TextureUnit(internal.TextureBindingLightingShadowMap, shadowTexture)
	commandBuffer.SamplerUnit(internal.TextureBindingLightingShadowMap, depthSampler)
	commandBuffer.UniformBufferUnit(
		internal.UniformBufferBindingCamera,
		ctx.CameraPlacement.Buffer,
		ctx.CameraPlacement.Offset,
		ctx.CameraPlacement.Size,
	)
	commandBuffer.UniformBufferUnit(
		internal.UniformBufferBindingLight,
		lightPlacement.Buffer,
		lightPlacement.Offset,
		lightPlacement.Size,
	)
	commandBuffer.DrawIndexed(0, quadShape.IndexCount(), 1)
}
