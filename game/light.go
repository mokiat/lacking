package game

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
)

type AmbientLightInfo struct {
}

type SpotLightInfo struct {
	EmitColor          opt.T[dprec.Vec3]
	EmitDistance       opt.T[float64]
	EmitOuterConeAngle opt.T[dprec.Angle]
	EmitInnerConeAngle opt.T[dprec.Angle]
}

/// DEPRECATED BELOW

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
	EmitDistance       float64
	EmitOuterConeAngle dprec.Angle
	EmitInnerConeAngle dprec.Angle
}

// DirectionalLightDefinition contains the properties of a directional light.
type DirectionalLightDefinition struct {
	EmitColor dprec.Vec3
	EmitRange float64
}
