package graphics

import (
	"fmt"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/dtos"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/util/spatial"
)

type DirectionalLightInfo struct {
	Position    dprec.Vec3
	Orientation dprec.Quat
	EmitColor   dprec.Vec3
	EmitRange   float64
}

func newDirectionalLight(scene *Scene, info DirectionalLightInfo) *DirectionalLight {
	light := scene.directionalLightPool.Fetch()

	light.scene = scene
	light.itemID = scene.directionalLightSet.Insert(
		info.Position, info.EmitRange, light,
	)

	light.active = true
	light.position = info.Position
	light.orientation = info.Orientation
	light.emitRange = info.EmitRange
	light.emitColor = info.EmitColor

	light.matrix = sprec.IdentityMat4()
	light.matrixDirty = true
	return light
}

type DirectionalLight struct {
	scene  *Scene
	itemID spatial.DynamicSetItemID

	active      bool
	position    dprec.Vec3
	orientation dprec.Quat
	emitRange   float64
	emitColor   dprec.Vec3

	matrix      sprec.Mat4
	matrixDirty bool
}

func (l *DirectionalLight) SetMatrix(matrix dprec.Mat4) { // FIXME
	t, r, _ := matrix.TRS()
	l.position = t
	l.orientation = r
	l.scene.directionalLightSet.Update(
		l.itemID, l.position, l.emitRange,
	)
	l.matrixDirty = true
}

func (l *DirectionalLight) Active() bool {
	return l.active
}

func (l *DirectionalLight) SetActive(active bool) {
	l.active = active
}

// Delete removes this light from the scene.
func (l *DirectionalLight) Delete() {
	if l.scene == nil {
		panic(fmt.Errorf("directional light already deleted"))
	}
	l.scene.directionalLightSet.Remove(l.itemID)
	l.scene.directionalLightPool.Restore(l)
	l.scene = nil
}

func (l *DirectionalLight) gfxMatrix() sprec.Mat4 {
	if l.matrixDirty || true { // FIXME
		l.matrix = sprec.TRSMat4(
			dtos.Vec3(l.position),
			dtos.Quat(l.orientation),
			sprec.NewVec3(1.0, 1.0, 1.0),
		)
		// l.matrix = sprec.Mat4Prod(
		// 	sprec.TranslationMat4(
		// 		float32(l.position.X),
		// 		float32(l.position.Y),
		// 		float32(l.position.Z),
		// 	),
		// 	sprec.ScaleMat4(
		// 		float32(l.emitRange),
		// 		float32(l.emitRange),
		// 		float32(l.emitRange),
		// 	),
		// )
		// FIXME
		// l.matrixDirty = false
	}
	return l.matrix
}
