package game

import "github.com/mokiat/gomath/dprec"

type AmbientLightDefinition struct {
	ReflectionTexture *CubeTexture
	RefractionTexture *CubeTexture
	OuterRadius       float64
	InnerRadius       float64
}

type DirectionalLightDefinition struct {
	EmitColor dprec.Vec3
	EmitRange float64
}
