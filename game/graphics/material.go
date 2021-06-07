package graphics

import "github.com/mokiat/gomath/sprec"

// Material determines the appearance of a mesh on the screen.
type Material interface {

	// Delete releases resources allocated for this material.
	Delete()
}

// PBRMaterialDefinition contains the information needed to create
// a PBR Material.
type PBRMaterialDefinition struct {
	BackfaceCulling  bool
	AlphaBlending    bool
	AlphaTesting     bool
	AlphaThreshold   float32
	Metalness        float32
	MetalnessTexture TwoDTexture
	Roughness        float32
	RoughnessTexture TwoDTexture
	AlbedoColor      sprec.Vec4
	AlbedoTexture    TwoDTexture
	NormalScale      float32
	NormalTexture    TwoDTexture
}
