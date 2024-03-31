package model

import "github.com/mokiat/gomath/dprec"

type ColorEmitter interface {
	EmitColor() dprec.Vec3
	SetEmitColor(color dprec.Vec3)
}

type DistanceEmitter interface {
	EmitDistance() float64
	SetEmitDistance(distance float64)
}

type PointLight struct {
	BlankNode

	emitColor    dprec.Vec3
	emitDistance float64
	castShadow   bool
}

func (p *PointLight) EmitColor() dprec.Vec3 {
	return p.emitColor
}

func (p *PointLight) SetEmitColor(color dprec.Vec3) {
	p.emitColor = color
}

func (p *PointLight) EmitDistance() float64 {
	return p.emitDistance
}

func (p *PointLight) SetEmitDistance(distance float64) {
	p.emitDistance = distance
}

func (p *PointLight) CastShadow() bool {
	return p.castShadow
}

func (p *PointLight) SetCastShadow(castShadow bool) {
	p.castShadow = castShadow
}
