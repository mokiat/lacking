package graphics

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/util/spatial"
)

type AmbientLightInfo struct {
	Position dprec.Vec3
	// TODO: Use a Box shape instead
	// Width
	// Height
	// Length
	// Overflow (for linear falloff into neighboring lights)
	InnerRadius       float64
	OuterRadius       float64
	ReflectionTexture render.Texture
	RefractionTexture render.Texture
}

func newAmbientLight(scene *Scene, info AmbientLightInfo) *AmbientLight {
	light := scene.ambientLightPool.Fetch()
	light.scene = scene
	light.itemID = scene.ambientLightSet.Insert(
		info.Position, info.OuterRadius, light,
	)
	light.innerRadius = info.InnerRadius
	light.outerRadius = info.OuterRadius
	light.reflectionTexture = info.ReflectionTexture
	light.refractionTexture = info.RefractionTexture
	light.active = true
	return light
}

type AmbientLight struct {
	scene  *Scene
	itemID spatial.DynamicSetItemID

	innerRadius       float64
	outerRadius       float64
	reflectionTexture render.Texture
	refractionTexture render.Texture

	active bool
}

func (l *AmbientLight) Active() bool {
	return l.active
}

func (l *AmbientLight) SetActive(active bool) {
	l.active = active
}

func (l *AmbientLight) Delete() {
	if l.scene == nil {
		panic("ambient light already deleted")
	}
	l.scene.ambientLightSet.Remove(l.itemID)
	l.scene.ambientLightPool.Restore(l)
	l.scene = nil
}

// TODO: Set/Get Position
// TODO: Set/Get Inner Radius
// TODO: Set/Get Outer Radius
