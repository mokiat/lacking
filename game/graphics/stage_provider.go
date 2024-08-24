package graphics

import "github.com/mokiat/lacking/render"

func newStageProvider(api render.API, shaders ShaderCollection, data *commonStageData, meshRenderer *meshRenderer) *StageProvider {
	return &StageProvider{
		api:          api,
		shaders:      shaders,
		data:         data,
		meshRenderer: meshRenderer,
	}
}

type StageProvider struct {
	api          render.API
	shaders      ShaderCollection
	data         *commonStageData
	meshRenderer *meshRenderer
}

// CreateDepthSourceStage creates a new DepthSourceStage.
func (p *StageProvider) CreateDepthSourceStage() *DepthSourceStage {
	return newDepthSourceStage(p.api)
}

// CreateGeometrySourceStage creates a new GeometrySourceStage.
func (p *StageProvider) CreateGeometrySourceStage() *GeometrySourceStage {
	return newGeometrySourceStage(p.api)
}

// CreateForwardSourceStage creates a new ForwardSourceStage.
func (p *StageProvider) CreateForwardSourceStage() *ForwardSourceStage {
	return newForwardSourceStage(p.api)
}

// CreateShadowStage creates a new ShadowStage using the specified input object.
func (p *StageProvider) CreateShadowStage(input ShadowStageInput) *ShadowStage {
	return newShadowStage(input)
}

// CreateGeometryStage creates a new GeometryStage using the specified input
// object.
func (p *StageProvider) CreateGeometryStage(input GeometryStageInput) *GeometryStage {
	return newGeometryStage(p.api, p.meshRenderer, input)
}

// CreateLightingStage creates a new LightingStage using the specified input
// object.
func (p *StageProvider) CreateLightingStage(input LightingStageInput) *LightingStage {
	return newLightingStage(p.api, p.shaders, p.data, input)
}

// CreateForwardStage creates a new ForwardStage using the specified input
// object.
func (p *StageProvider) CreateForwardStage(input ForwardStageInput) *ForwardStage {
	return newForwardStage(p.api, p.shaders, p.data, p.meshRenderer, input)
}

// CreateExposureProbeStage creates a new ExposureProbeStage using the specified
// input object.
func (p *StageProvider) CreateExposureProbeStage(input ExposureProbeStageInput) *ExposureProbeStage {
	return newExposureProbeStage(p.api, p.shaders, p.data, input)
}

// CreateBloomStage creates a new BloomStage using the specified input object.
func (p *StageProvider) CreateBloomStage(input BloomStageInput) *BloomStage {
	return newBloomStage(p.api, p.shaders, p.data, input)
}

// CreateToneMappingStage creates a new ToneMappingStage using the specified
// input object.
func (p *StageProvider) CreateToneMappingStage(input ToneMappingStageInput) *ToneMappingStage {
	return newToneMappingStage(p.api, p.shaders, p.data, input)
}
