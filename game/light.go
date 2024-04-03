package game

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/render"
)

// AmbientLightInfo contains the information required to create an ambient
// light.
type AmbientLightInfo struct {
	ReflectionTexture render.Texture
	RefractionTexture render.Texture
	OuterRadius       opt.T[float64]
	InnerRadius       opt.T[float64]
	CastShadow        opt.T[bool]
}

// PointLightInfo contains the information required to create a point light.
type PointLightInfo struct {
	EmitColor    opt.T[dprec.Vec3]
	EmitDistance opt.T[float64]
	CastShadow   opt.T[bool]
}

// SpotLightInfo contains the information required to create a spot light.
type SpotLightInfo struct {
	EmitColor          opt.T[dprec.Vec3]
	EmitDistance       opt.T[float64]
	EmitOuterConeAngle opt.T[dprec.Angle]
	EmitInnerConeAngle opt.T[dprec.Angle]
	CastShadow         opt.T[bool]
}

// DirectionalLightInfo contains the information required to create a
// directional light.
type DirectionalLightInfo struct {
	EmitColor  opt.T[dprec.Vec3]
	CastShadow opt.T[bool]
}
