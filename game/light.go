package game

import "github.com/mokiat/gomath/dprec"

// AmbientLightDefinition contains the properties of an ambient light.
type AmbientLightDefinition struct {
	ReflectionTexture *CubeTexture
	RefractionTexture *CubeTexture
	OuterRadius       float64
	InnerRadius       float64
}

// PointLightDefinition contains the properties of a point light.
type PointLightDefinition struct {
	EmitColor dprec.Vec3
	EmitRange float64
}

// SpotLightDefinition contains the properties of a spot light.
type SpotLightDefinition struct {
	EmitColor          dprec.Vec3
	EmitOuterConeAngle dprec.Angle
	EmitInnerConeAngle dprec.Angle
	EmitRange          float64
}

// DirectionalLightDefinition contains the properties of a directional light.
type DirectionalLightDefinition struct {
	EmitColor dprec.Vec3
	EmitRange float64
}
