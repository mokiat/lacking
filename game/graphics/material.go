package graphics

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/render"
)

// MaterialDefinitionInfo contains the information needed to create a
// MaterialDefinition.
type MaterialDefinitionInfo struct {
	BackfaceCulling bool
	AlphaTesting    bool
	AlphaBlending   bool
	AlphaThreshold  float32
	Vectors         []sprec.Vec4
	TwoDTextures    []render.Texture
	CubeTextures    []render.Texture
	Shading         Shading
}

// MaterialDefinition represents a particular material template. Multiple meshes
// can share the same MaterialDefinition though their Material instances will
// differ.
type MaterialDefinition struct {
	revision int

	backfaceCulling bool
	alphaTesting    bool
	alphaBlending   bool
	alphaThreshold  float32

	uniformData  []byte
	twoDTextures []render.Texture
	cubeTextures []render.Texture

	shading Shading
}

// BackfaceCulling returns whether the back face of triangles should be
// skipped during rendering.
func (d *MaterialDefinition) BackfaceCulling() bool {
	return d.backfaceCulling
}

// SetBackfaceCulling changes the back face culling configuration of this
// material definition.
func (d *MaterialDefinition) SetBackfaceCulling(culling bool) {
	if culling != d.backfaceCulling {
		d.revision++
		d.backfaceCulling = culling
	}
}

// AlphaTesting returns whether the mesh will be checked for transparent
// sections.
func (d *MaterialDefinition) AlphaTesting() bool {
	return d.alphaTesting
}

// SetAlphaTesting changes whether alpha testing will be performed.
func (d *MaterialDefinition) SetAlphaTesting(testing bool) {
	if testing != d.alphaTesting {
		d.revision++
		d.alphaTesting = testing
	}
}

// AlphaBlending returns whether the mesh will be checked for translucency
// and will be mixed with the background.
func (d *MaterialDefinition) AlphaBlending() bool {
	return d.alphaBlending
}

// SetAlphaBlending changes whether alpha blending will be performed.
func (d *MaterialDefinition) SetAlphaBlending(blending bool) {
	if blending != d.alphaBlending {
		d.revision++
		d.alphaBlending = blending
	}
}

// AlphaThreshold returns the alpha value below which a pixel will be
// considered transparent.
func (d *MaterialDefinition) AlphaThreshold() float32 {
	return d.alphaThreshold
}

// SetAlphaThreshold changes the alpha threshold.
func (d *MaterialDefinition) SetAlphaThreshold(threshold float32) {
	if !sprec.Eq(threshold, d.alphaThreshold) {
		d.revision++
		d.alphaThreshold = threshold
	}
}

// Material determines the appearance of a mesh on the screen.
type Material struct {
	definitionRevision int
	definition         *MaterialDefinition

	shadowPipeline   render.Pipeline
	geometryPipeline render.Pipeline
	emissivePipeline render.Pipeline
	forwardPipeline  render.Pipeline
}

// PBRMaterialInfo contains the information needed to create a PBR Material.
type PBRMaterialInfo struct {
	BackfaceCulling          bool
	AlphaBlending            bool
	AlphaTesting             bool
	AlphaThreshold           float32
	Metallic                 float32
	Roughness                float32
	MetallicRoughnessTexture *TwoDTexture
	AlbedoColor              sprec.Vec4
	AlbedoTexture            *TwoDTexture
	NormalScale              float32
	NormalTexture            *TwoDTexture
}
