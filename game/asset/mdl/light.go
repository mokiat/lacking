package mdl

import (
	"github.com/mokiat/gomath/dprec"
)

type ColorEmitter interface {
	EmitColor() dprec.Vec3
	SetEmitColor(color dprec.Vec3)
}

type BaseColorEmitter struct {
	emitColor dprec.Vec3
}

func (b *BaseColorEmitter) EmitColor() dprec.Vec3 {
	return b.emitColor
}

func (b *BaseColorEmitter) SetEmitColor(color dprec.Vec3) {
	b.emitColor = color
}

type DistanceEmitter interface {
	EmitDistance() float64
	SetEmitDistance(distance float64)
}

type BaseDistanceEmitter struct {
	emitDistance float64
}

func (b *BaseDistanceEmitter) EmitDistance() float64 {
	return b.emitDistance
}

func (b *BaseDistanceEmitter) SetEmitDistance(distance float64) {
	b.emitDistance = distance
}

type ConeEmitter interface {
	EmitAngleOuter() dprec.Angle
	SetEmitAngleOuter(angle dprec.Angle)
	EmitAngleInner() dprec.Angle
	SetEmitAngleInner(angle dprec.Angle)
}

type BaseConeEmitter struct {
	emitAngleOuter dprec.Angle
	emitAngleInner dprec.Angle
}

func (b *BaseConeEmitter) EmitAngleOuter() dprec.Angle {
	return b.emitAngleOuter
}

func (b *BaseConeEmitter) SetEmitAngleOuter(angle dprec.Angle) {
	b.emitAngleOuter = angle
}

func (b *BaseConeEmitter) EmitAngleInner() dprec.Angle {
	return b.emitAngleInner
}

func (b *BaseConeEmitter) SetEmitAngleInner(angle dprec.Angle) {
	b.emitAngleInner = angle
}

type ShadowCaster interface {
	CastShadow() bool
	SetCastShadow(castShadow bool)
}

type BaseShadowCaster struct {
	castShadow bool
}

func (b *BaseShadowCaster) CastShadow() bool {
	return b.castShadow
}

func (b *BaseShadowCaster) SetCastShadow(castShadow bool) {
	b.castShadow = castShadow
}

type AmbientLight struct {
	BaseNode
	BaseShadowCaster
	reflectionTexture *Texture
	refractionTexture *Texture
}

func (l *AmbientLight) ReflectionTexture() *Texture {
	return l.reflectionTexture
}

func (l *AmbientLight) SetReflectionTexture(texture *Texture) {
	l.reflectionTexture = texture
}

func (l *AmbientLight) RefractionTexture() *Texture {
	return l.refractionTexture
}

func (l *AmbientLight) SetRefractionTexture(texture *Texture) {
	l.refractionTexture = texture
}

type PointLight struct {
	BaseNode
	BaseColorEmitter
	BaseDistanceEmitter
	BaseShadowCaster
}

type SpotLight struct {
	BaseNode
	BaseColorEmitter
	BaseDistanceEmitter
	BaseConeEmitter
	BaseShadowCaster
}

type DirectionalLight struct {
	BaseNode
	BaseColorEmitter
	BaseShadowCaster
}
