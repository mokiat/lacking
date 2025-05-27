package mdl

import (
	"github.com/mokiat/lacking/game/asset/dto/shadingdto"
)

const (
	CullModeNone         CullMode = shadingdto.CullModeNone
	CullModeFront        CullMode = shadingdto.CullModeFront
	CullModeBack         CullMode = shadingdto.CullModeBack
	CullModeFrontAndBack CullMode = shadingdto.CullModeFrontAndBack
)

type CullMode = shadingdto.CullMode

const (
	FaceOrientationCCW FaceOrientation = shadingdto.FaceOrientationCCW
	FaceOrientationCW  FaceOrientation = shadingdto.FaceOrientationCW
)

type FaceOrientation = shadingdto.FaceOrientation

const (
	ComparisonNever          Comparison = shadingdto.ComparisonNever
	ComparisonLess           Comparison = shadingdto.ComparisonLess
	ComparisonEqual          Comparison = shadingdto.ComparisonEqual
	ComparisonLessOrEqual    Comparison = shadingdto.ComparisonLessOrEqual
	ComparisonGreater        Comparison = shadingdto.ComparisonGreater
	ComparisonNotEqual       Comparison = shadingdto.ComparisonNotEqual
	ComparisonGreaterOrEqual Comparison = shadingdto.ComparisonGreaterOrEqual
	ComparisonAlways         Comparison = shadingdto.ComparisonAlways
)

type Comparison = shadingdto.Comparison

func NewMaterial(name string) *Material {
	return &Material{
		id:                   freeID.Add(1),
		name:                 name,
		metadata:             nil,
		samplers:             nil,
		properties:           nil,
		geometryPasses:       nil,
		shadowPasses:         nil,
		forwardPasses:        nil,
		skyPasses:            nil,
		postprocessingPasses: nil,
	}
}

type Material struct {
	id   uint32
	name string

	metadata Metadata

	samplers   map[string]*Sampler
	properties map[string]any

	geometryPasses       []*MaterialPass
	shadowPasses         []*MaterialPass
	forwardPasses        []*MaterialPass
	skyPasses            []*MaterialPass
	postprocessingPasses []*MaterialPass
}

func (m *Material) ID() uint32 {
	return m.id
}

func (m *Material) Clear() {
	clear(m.metadata)
	clear(m.samplers)
	clear(m.properties)
	m.geometryPasses = nil
	m.shadowPasses = nil
	m.forwardPasses = nil
	m.skyPasses = nil
	m.postprocessingPasses = nil
}

func (m *Material) Metadata() Metadata {
	return m.metadata
}

func (m *Material) SetMetadata(metadata Metadata) {
	m.metadata = metadata
}

func (m *Material) Name() string {
	return m.name
}

func (m *Material) SetName(name string) {
	m.name = name
}

func (m *Material) Samplers() map[string]*Sampler {
	return m.samplers
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

func (m *Material) Properties() map[string]any {
	return m.properties
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

func (m *Material) GeometryPasses() []*MaterialPass {
	return m.geometryPasses
}

func (m *Material) AddGeometryPass(pass *MaterialPass) {
	m.geometryPasses = append(m.geometryPasses, pass)
}

func (m *Material) ShadowPasses() []*MaterialPass {
	return m.shadowPasses
}

func (m *Material) AddShadowPass(pass *MaterialPass) {
	m.shadowPasses = append(m.shadowPasses, pass)
}

func (m *Material) ForwardPasses() []*MaterialPass {
	return m.forwardPasses
}

func (m *Material) AddForwardPass(pass *MaterialPass) {
	m.forwardPasses = append(m.forwardPasses, pass)
}

func (m *Material) SkyPasses() []*MaterialPass {
	return m.skyPasses
}

func (m *Material) AddSkyPass(pass *MaterialPass) {
	m.skyPasses = append(m.skyPasses, pass)
}

func (m *Material) PostprocessingPasses() []*MaterialPass {
	return m.postprocessingPasses
}

func (m *Material) AddPostprocessingPass(pass *MaterialPass) {
	m.postprocessingPasses = append(m.postprocessingPasses, pass)
}

func (m *Material) AllPasses() []*MaterialPass {
	var result []*MaterialPass
	result = append(result, m.geometryPasses...)
	result = append(result, m.shadowPasses...)
	result = append(result, m.forwardPasses...)
	result = append(result, m.skyPasses...)
	result = append(result, m.postprocessingPasses...)
	return result
}

func NewMaterialPass() *MaterialPass {
	return &MaterialPass{
		layer:           0,
		culling:         CullModeNone,
		frontFace:       FaceOrientationCCW,
		depthTest:       true,
		depthWrite:      true,
		depthComparison: ComparisonLess,
		blending:        false,
	}
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
