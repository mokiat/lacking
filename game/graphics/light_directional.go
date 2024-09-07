package graphics

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/dtos"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/util/spatial"
)

const dirLightRadius = 16000.0

type DirectionalLightInfo struct {
	Position   dprec.Vec3
	Rotation   dprec.Quat
	EmitColor  dprec.Vec3
	CastShadow bool
}

func newDirectionalLight(scene *Scene, info DirectionalLightInfo) *DirectionalLight {
	light := scene.directionalLightPool.Fetch()

	light.scene = scene
	light.itemID = scene.directionalLightSet.Insert(
		info.Position, dirLightRadius, light,
	)

	light.active = true
	light.position = info.Position
	light.rotation = info.Rotation
	light.emitColor = info.EmitColor
	light.castShadow = info.CastShadow

	light.matrix = sprec.IdentityMat4()
	light.matrixDirty = true
	return light
}

type DirectionalLight struct {
	scene  *Scene
	itemID spatial.DynamicSetItemID

	active     bool
	position   dprec.Vec3
	rotation   dprec.Quat
	emitColor  dprec.Vec3
	castShadow bool

	matrix      sprec.Mat4
	matrixDirty bool
}

// Active returns whether this light will be applied.
func (l *DirectionalLight) Active() bool {
	return l.active
}

// SetActive changes whether this light will be applied.
func (l *DirectionalLight) SetActive(active bool) {
	l.active = active
}

// Position returns the location of this light source.
func (l *DirectionalLight) Position() dprec.Vec3 {
	return l.position
}

// SetPosition changes the position of this light source.
func (l *DirectionalLight) SetPosition(position dprec.Vec3) {
	if position != l.position {
		l.position = position
		l.scene.directionalLightSet.Update(
			l.itemID, l.position, dirLightRadius,
		)
		l.matrixDirty = true
	}
}

// Rotation returns the orientation of this light source.
func (l *DirectionalLight) Rotation() dprec.Quat {
	return l.rotation
}

// SetRotation changes the orientation of this light source.
func (l *DirectionalLight) SetRotation(rotation dprec.Quat) {
	if rotation != l.rotation {
		l.rotation = rotation
		l.matrixDirty = true
	}
}

// EmitColor returns the linear color of this light.
func (l *DirectionalLight) EmitColor() dprec.Vec3 {
	return l.emitColor
}

// SetEmitColor changes the linear color of this light. The values
// can be outside the [0.0, 1.0] range for higher intensity.
func (l *DirectionalLight) SetEmitColor(color dprec.Vec3) {
	l.emitColor = color
}

// CastShadow returns whether this light will cast a shadow.
func (l *DirectionalLight) CastShadow() bool {
	return l.castShadow
}

// SetCastShadow changes whether this light will cast a shadow.
func (l *DirectionalLight) SetCastShadow(castShadow bool) {
	l.castShadow = castShadow
}

// Delete removes this light from the scene.
func (l *DirectionalLight) Delete() {
	if l.scene == nil {
		panic("directional light already deleted")
	}
	l.scene.directionalLightSet.Remove(l.itemID)
	l.scene.directionalLightPool.Restore(l)
	l.scene = nil
}

func (l *DirectionalLight) gfxMatrix() sprec.Mat4 {
	if l.matrixDirty {
		l.matrix = sprec.TRSMat4(
			dtos.Vec3(l.position),
			dtos.Quat(l.rotation),
			sprec.NewVec3(1.0, 1.0, 1.0),
		)
		l.matrixDirty = false
	}
	return l.matrix
}
