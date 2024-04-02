package game

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/render"
)

type AmbientLightInfo struct {
	ReflectionTexture render.Texture
	RefractionTexture render.Texture
	OuterRadius       opt.T[float64]
	InnerRadius       opt.T[float64]
}

type PointLightInfo struct {
	EmitColor    opt.T[dprec.Vec3]
	EmitDistance opt.T[float64]
}

type SpotLightInfo struct {
	EmitColor          opt.T[dprec.Vec3]
	EmitDistance       opt.T[float64]
	EmitOuterConeAngle opt.T[dprec.Angle]
	EmitInnerConeAngle opt.T[dprec.Angle]
}

/// DEPRECATED BELOW

// DirectionalLightDefinition contains the properties of a directional light.
type DirectionalLightDefinition struct {
	EmitColor dprec.Vec3
	EmitRange float64
}
