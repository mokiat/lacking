package graphics

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/graphics/internal"
)

// MaterialInfo contains the information needed to create a Material.
type MaterialInfo struct {

	// Name specifies the name of the material.
	Name string

	// GeometryPasses specifies a list of geometry passes to be applied.
	// This is used for opaque materials that go through deferred shading.
	GeometryPasses []GeometryRenderPassInfo

	// ShadowPasses specifies a list of shadow passes to be applied.
	// This should be omitted if the material does not cast shadows.
	ShadowPasses []ShadowRenderPassInfo

	// ForwardPasses specifies a list of forward passes to be applied.
	// This is useful when dealing with translucent materials or when special
	// effects are needed.
	ForwardPasses []ForwardRenderPassInfo
}

// Material determines the appearance of a mesh on the screen.
type Material struct {
	name string

	// TODO: Restructure in fixed array
	geometryPasses []internal.MaterialRenderPassDefinition
	shadowPasses   []internal.MaterialRenderPassDefinition
	forwardPasses  []internal.MaterialRenderPassDefinition
}

func (m *Material) Name() string {
	return m.name
}

/// OLD STUFF BELOW

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
