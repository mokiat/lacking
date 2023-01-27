package graphics

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/util/spatial"
)

type AmbientLightInfo struct {
	Position dprec.Vec3
	// TODO: Use a Box shape instead
	InnerRadius       float64
	OuterRadius       float64
	ReflectionTexture *CubeTexture
	RefractionTexture *CubeTexture
}

func newAmbientLight(scene *Scene, info AmbientLightInfo) *AmbientLight {
	light := scene.ambientLightPool.Fetch()
	light.scene = scene
	light.item = scene.ambientLightOctree.CreateItem(light)
	light.item.SetPosition(info.Position)
	light.item.SetRadius(info.OuterRadius)
	light.innerRadius = info.InnerRadius
	light.outerRadius = info.OuterRadius
	light.reflectionTexture = info.ReflectionTexture
	light.refractionTexture = info.RefractionTexture
	return light
}

type AmbientLight struct {
	scene *Scene
	item  *spatial.OctreeItem[*AmbientLight]

	innerRadius       float64
	outerRadius       float64
	reflectionTexture *CubeTexture
	refractionTexture *CubeTexture
}

// TODO: Set/Get Position
// TODO: Set/Get Inner Radius
// TODO: Set/Get Outer Radius
