package graphics

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/graphics/internal"
	"github.com/mokiat/lacking/render"
)

// Material determines the appearance of a mesh on the screen.
type Material struct {
	backfaceCulling bool
	alphaTesting    bool
	alphaBlending   bool
	alphaThreshold  float32

	geometryPresentation *internal.GeometryPresentation
	shadowPresentation   *internal.ShadowPresentation

	twoDTextures []render.Texture
	cubeTextures []render.Texture
	vectors      []sprec.Vec4
}

// Delete releases resources allocated for this material.
func (m *Material) Delete() {
	if m.geometryPresentation != nil {
		m.geometryPresentation.Delete()
	}
	if m.shadowPresentation != nil {
		m.shadowPresentation.Delete()
	}
}

// PBRMaterialDefinition contains the information needed to create
// a PBR Material.
type PBRMaterialDefinition struct {
	BackfaceCulling  bool
	AlphaBlending    bool
	AlphaTesting     bool
	AlphaThreshold   float32
	Metalness        float32
	MetalnessTexture *TwoDTexture
	Roughness        float32
	RoughnessTexture *TwoDTexture
	AlbedoColor      sprec.Vec4
	AlbedoTexture    *TwoDTexture
	NormalScale      float32
	NormalTexture    *TwoDTexture
}
