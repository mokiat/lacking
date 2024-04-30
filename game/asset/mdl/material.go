package mdl

import "github.com/mokiat/lacking/game/asset"

const (
	CullModeNone         CullMode = asset.CullModeNone
	CullModeFront        CullMode = asset.CullModeFront
	CullModeBack         CullMode = asset.CullModeBack
	CullModeFrontAndBack CullMode = asset.CullModeFrontAndBack
)

type CullMode = asset.CullMode

const (
	FaceOrientationCCW FaceOrientation = asset.FaceOrientationCCW
	FaceOrientationCW  FaceOrientation = asset.FaceOrientationCW
)

type FaceOrientation = asset.FaceOrientation

const (
	ComparisonNever          Comparison = asset.ComparisonNever
	ComparisonLess           Comparison = asset.ComparisonLess
	ComparisonEqual          Comparison = asset.ComparisonEqual
	ComparisonLessOrEqual    Comparison = asset.ComparisonLessOrEqual
	ComparisonGreater        Comparison = asset.ComparisonGreater
	ComparisonNotEqual       Comparison = asset.ComparisonNotEqual
	ComparisonGreaterOrEqual Comparison = asset.ComparisonGreaterOrEqual
	ComparisonAlways         Comparison = asset.ComparisonAlways
)

type Comparison = asset.Comparison

type Material struct {
	name string

	samplers   map[string]*Sampler
	properties map[string]any

	geometryPasses       []*MaterialPass
	shadowPasses         []*MaterialPass
	forwardPasses        []*MaterialPass
	skyPasses            []*MaterialPass
	postprocessingPasses []*MaterialPass
}

func (m *Material) Name() string {
	return m.name
}

func (m *Material) SetName(name string) {
	m.name = name
}

func (m *Material) Sampler(name string) *Sampler {
	if m.samplers == nil {
		return nil
	}
	return m.samplers[name]
}

func (m *Material) SetSampler(name string, sampler *Sampler) {
	if m.samplers == nil {
		m.samplers = make(map[string]*Sampler)
	}
	m.samplers[name] = sampler
}

func (m *Material) Property(name string) any {
	if m.properties == nil {
		return nil
	}
	return m.properties[name]
}

func (m *Material) SetProperty(name string, value any) {
	if m.properties == nil {
		m.properties = make(map[string]any)
	}
	m.properties[name] = value
}

func (m *Material) AddGeometryPass(pass *MaterialPass) {
	m.geometryPasses = append(m.geometryPasses, pass)
}

func (m *Material) AddShadowPass(pass *MaterialPass) {
	m.shadowPasses = append(m.shadowPasses, pass)
}

func (m *Material) AddForwardPass(pass *MaterialPass) {
	m.forwardPasses = append(m.forwardPasses, pass)
}

func (m *Material) AddSkyPass(pass *MaterialPass) {
	m.skyPasses = append(m.skyPasses, pass)
}

func (m *Material) AddPostprocessPass(pass *MaterialPass) {
	m.postprocessingPasses = append(m.postprocessingPasses, pass)
}

type MaterialPass struct {
	layer           int
	culling         CullMode
	frontFace       FaceOrientation
	depthTest       bool
	depthWrite      bool
	depthComparison Comparison
	blending        bool
	shader          *Shader
}

func (m *MaterialPass) Layer() int {
	return m.layer
}

func (m *MaterialPass) SetLayer(layer int) {
	m.layer = layer
}

func (m *MaterialPass) Culling() CullMode {
	return m.culling
}

func (m *MaterialPass) SetCulling(culling CullMode) {
	m.culling = culling
}

func (m *MaterialPass) FrontFace() FaceOrientation {
	return m.frontFace
}

func (m *MaterialPass) SetFrontFace(frontFace FaceOrientation) {
	m.frontFace = frontFace
}

func (m *MaterialPass) DepthTest() bool {
	return m.depthTest
}

func (m *MaterialPass) SetDepthTest(depthTest bool) {
	m.depthTest = depthTest
}

func (m *MaterialPass) DepthWrite() bool {
	return m.depthWrite
}

func (m *MaterialPass) SetDepthWrite(depthWrite bool) {
	m.depthWrite = depthWrite
}

func (m *MaterialPass) DepthComparison() Comparison {
	return m.depthComparison
}

func (m *MaterialPass) SetDepthComparison(depthComparison Comparison) {
	m.depthComparison = depthComparison
}

func (m *MaterialPass) Blending() bool {
	return m.blending
}

func (m *MaterialPass) SetBlending(blending bool) {
	m.blending = blending
}

func (m *MaterialPass) Shader() *Shader {
	return m.shader
}

func (m *MaterialPass) SetShader(shader *Shader) {
	m.shader = shader
}
