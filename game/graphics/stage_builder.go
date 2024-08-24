package graphics

import "github.com/mokiat/gog/opt"

// StageBuilderFunc is a function that creates a list of stages.
type StageBuilderFunc func(provider *StageProvider) []Stage

// DefaultStageBuilder is a default implementation of a stage builder.
func DefaultStageBuilder(provider *StageProvider) []Stage {
	depthSourceStage := provider.CreateDepthSourceStage()

	geometrySourceStage := provider.CreateGeometrySourceStage()

	forwardSourceStage := provider.CreateForwardSourceStage()

	shadowStage := provider.CreateShadowStage(ShadowStageInput{
		// TODO
	})

	geometryStage := provider.CreateGeometryStage(GeometryStageInput{
		AlbedoMetallicTexture:  geometrySourceStage.AlbedoMetallicTexture,
		NormalRoughnessTexture: geometrySourceStage.NormalRoughnessTexture,
		DepthTexture:           depthSourceStage.DepthTexture,
	})

	lightingStage := provider.CreateLightingStage(LightingStageInput{
		AlbedoMetallicTexture:  geometrySourceStage.AlbedoMetallicTexture,
		NormalRoughnessTexture: geometrySourceStage.NormalRoughnessTexture,
		DepthTexture:           depthSourceStage.DepthTexture,
		HDRTexture:             forwardSourceStage.HDRTexture,
	})

	forwardStage := provider.CreateForwardStage(ForwardStageInput{
		HDRTexture:   forwardSourceStage.HDRTexture,
		DepthTexture: depthSourceStage.DepthTexture,
	})

	exposureProbeStage := provider.CreateExposureProbeStage(ExposureProbeStageInput{
		HDRTexture: forwardSourceStage.HDRTexture,
	})

	bloomStage := provider.CreateBloomStage(BloomStageInput{
		HDRTexture: forwardSourceStage.HDRTexture,
	})

	toneMappingStage := provider.CreateToneMappingStage(ToneMappingStageInput{
		HDRTexture:   forwardSourceStage.HDRTexture,
		BloomTexture: opt.V[StageTextureParameter](bloomStage.BloomTexture),
	})

	return []Stage{
		depthSourceStage,
		geometrySourceStage,
		forwardSourceStage,
		shadowStage,
		geometryStage,
		lightingStage,
		forwardStage,
		exposureProbeStage,
		bloomStage,
		toneMappingStage,
	}
}
