package graphics

import (
	"fmt"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/util/spatial"
)

// SpotLightInfo contains the information needed to create a SpotLight.
type SpotLightInfo struct {
	Position           dprec.Vec3
	EmitRange          float64
	EmitOuterConeAngle dprec.Angle
	EmitInnerConeAngle dprec.Angle
	EmitColor          dprec.Vec3
}

func newSpotLight(scene *Scene, info SpotLightInfo) *SpotLight {
	light := scene.spotLightPool.Fetch()

	light.scene = scene
	light.item = scene.spotLightOctree.CreateItem(light)
	light.item.SetPosition(info.Position)
	light.item.SetRadius(info.EmitRange)

	light.active = true
	light.position = info.Position
	light.emitRange = info.EmitRange
	light.emitOuterConeAngle = info.EmitOuterConeAngle
	light.emitInnerConeAngle = info.EmitInnerConeAngle
	light.emitColor = info.EmitColor

	light.matrix = sprec.IdentityMat4()
	light.matrixDirty = true
	return light
}

// SpotLight represents a light source that is positioned at a point in
// space and emits a light cone in down the -Z axis up to a range.
type SpotLight struct {
	scene *Scene
	item  *spatial.OctreeItem[*SpotLight]

	active             bool
	position           dprec.Vec3
	emitRange          float64
	emitOuterConeAngle dprec.Angle
	emitInnerConeAngle dprec.Angle
	emitColor          dprec.Vec3

	matrix      sprec.Mat4
	matrixDirty bool
}

// Active returns whether this light will be applied.
func (l *SpotLight) Active() bool {
	return l.active
}

// SetActive changes whether this light will be applied.
func (l *SpotLight) SetActive(active bool) {
	l.active = active
}

// Position returns the location of this light source.
func (l *SpotLight) Position() dprec.Vec3 {
	return l.position
}

// SetPosition changes the position of this light source.
func (l *SpotLight) SetPosition(position dprec.Vec3) {
	if position != l.position {
		l.position = position
		l.item.SetPosition(l.position)
		l.matrixDirty = true
	}
}

// EmitRange returns the distance that this light source covers.
func (l *SpotLight) EmitRange() float64 {
	return l.emitRange
}

// SetEmitRange changes the distance that this light source covers.
func (l *SpotLight) SetEmitRange(emitRange float64) {
	if emitRange != l.emitRange {
		l.emitRange = dprec.Max(0.0, emitRange)
		l.item.SetRadius(l.emitRange)
		l.matrixDirty = true
	}
}

// EmitColor returns the linear color of this light.
func (l *SpotLight) EmitColor() dprec.Vec3 {
	return l.emitColor
}

// SetEmitColor changes the linear color of this light. The values
// can be outside the [0.0, 1.0] range for higher intensity.
func (l *SpotLight) SetEmitColor(color dprec.Vec3) {
	l.emitColor = color
}

// Delete removes this light from the scene.
func (l *SpotLight) Delete() {
	if l.scene == nil {
		panic(fmt.Errorf("spot light already deleted"))
	}
	l.item.Delete()
	l.item = nil
	l.scene.spotLightPool.Restore(l)
	l.scene = nil
}

func (l *SpotLight) gfxMatrix() sprec.Mat4 {
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
