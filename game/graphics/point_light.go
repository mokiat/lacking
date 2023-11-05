package graphics

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/util/spatial"
)

// PointLightInfo contains the information needed to create a PointLight.
type PointLightInfo struct {
	Position  dprec.Vec3
	EmitRange float64
	EmitColor dprec.Vec3
}

func newPointLight(scene *Scene, info PointLightInfo) *PointLight {
	light := scene.pointLightPool.Fetch()

	light.scene = scene
	light.itemID = scene.pointLightSet.Insert(
		info.Position, info.EmitRange, light,
	)

	light.active = true
	light.position = info.Position
	light.emitRange = info.EmitRange
	light.emitColor = info.EmitColor

	light.matrix = sprec.IdentityMat4()
	light.matrixDirty = true
	return light
}

// PointLight represents a light source that is positioned at a point in
// space and emits light evenly in all directions up to a range.
type PointLight struct {
	scene  *Scene
	itemID spatial.DynamicSetItemID

	active    bool
	position  dprec.Vec3
	emitRange float64
	emitColor dprec.Vec3

	matrix      sprec.Mat4
	matrixDirty bool
}

// Active returns whether this light will be applied.
func (l *PointLight) Active() bool {
	return l.active
}

// SetActive changes whether this light will be applied.
func (l *PointLight) SetActive(active bool) {
	l.active = active
}

// Position returns the location of this light source.
func (l *PointLight) Position() dprec.Vec3 {
	return l.position
}

// SetPosition changes the position of this light source.
func (l *PointLight) SetPosition(position dprec.Vec3) {
	if position != l.position {
		l.position = position
		l.scene.pointLightSet.Update(
			l.itemID, l.position, l.emitRange,
		)
		l.matrixDirty = true
	}
}

// EmitRange returns the distance that this light source covers.
func (l *PointLight) EmitRange() float64 {
	return l.emitRange
}

// SetEmitRange changes the distance that this light source covers.
func (l *PointLight) SetEmitRange(emitRange float64) {
	if emitRange != l.emitRange {
		l.emitRange = dprec.Max(0.0, emitRange)
		l.scene.pointLightSet.Update(
			l.itemID, l.position, l.emitRange,
		)
		l.matrixDirty = true
	}
}

// EmitColor returns the linear color of this light.
func (l *PointLight) EmitColor() dprec.Vec3 {
	return l.emitColor
}

// SetEmitColor changes the linear color of this light. The values
// can be outside the [0.0, 1.0] range for higher intensity.
func (l *PointLight) SetEmitColor(color dprec.Vec3) {
	l.emitColor = color
}

// Delete removes this light from the scene.
func (l *PointLight) Delete() {
	if l.scene == nil {
		panic("light already deleted")
	}
	l.scene.pointLightSet.Remove(l.itemID)
	l.scene.pointLightPool.Restore(l)
	l.scene = nil
}

func (l *PointLight) gfxMatrix() sprec.Mat4 {
	if l.matrixDirty {
		l.matrix = sprec.Mat4Prod(
			sprec.TranslationMat4(
				float32(l.position.X),
				float32(l.position.Y),
				float32(l.position.Z),
			),
			sprec.ScaleMat4(
				float32(l.emitRange),
				float32(l.emitRange),
				float32(l.emitRange),
			),
		)
		l.matrixDirty = false
	}
	return l.matrix
}
