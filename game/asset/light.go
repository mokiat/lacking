package asset

import "github.com/mokiat/gomath/dprec"

type LightInstance struct {
	Name               string
	NodeIndex          int32
	Type               LightType
	EmitRange          float64
	EmitOuterConeAngle dprec.Angle
	EmitInnerConeAngle dprec.Angle
	EmitColor          dprec.Vec3
}

type LightType uint8

const (
	LightTypePoint LightType = 1 + iota
	LightTypeSpot
	LightTypeDirectional
	// TODO: Ambient Light
)
