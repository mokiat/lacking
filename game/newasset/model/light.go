package model

import (
	"github.com/mokiat/gomath/dprec"
	asset "github.com/mokiat/lacking/game/newasset"
)

type ColorEmitter interface {
	EmitColor() dprec.Vec3
	SetEmitColor(color dprec.Vec3)
}

type IntensityEmitter interface {
	EmitIntensity() float64
	SetEmitIntensity(intensity float64)
}

type DistanceEmitter interface {
	EmitDistance() float64
	SetEmitDistance(distance float64)
}

type PointLight struct {
	emitColor     dprec.Vec3
	emitIntensity float64
	emitDistance  float64
}

func (p *PointLight) EmitColor() dprec.Vec3 {
	return p.emitColor
}

func (p *PointLight) SetEmitColor(color dprec.Vec3) {
	p.emitColor = color
}

func (p *PointLight) EmitIntensity() float64 {
	return p.emitIntensity
}

func (p *PointLight) SetEmitIntensity(intensity float64) {
	p.emitIntensity = intensity
}

func (p *PointLight) EmitDistance() float64 {
	return p.emitDistance
}

func (p *PointLight) SetEmitDistance(distance float64) {
	p.emitDistance = distance
}

func (p *PointLight) ToAsset() asset.PointLight {
	return asset.PointLight{
		NodeIndex:     asset.UnspecifiedNodeIndex,
		EmitColor:     p.emitColor,
		EmitIntensity: p.emitIntensity,
		EmitRange:     p.emitDistance,
	}
}
