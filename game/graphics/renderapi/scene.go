package graphics

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/graphics/renderapi/internal"
)

func newScene(renderer *Renderer) *Scene {
	return &Scene{
		renderer: renderer,

		sky: newSky(),
	}
}

var _ graphics.Scene = (*Scene)(nil)

type Scene struct {
	renderer *Renderer

	sky *Sky

	firstMesh  *Mesh
	lastMesh   *Mesh
	cachedMesh *Mesh

	firstLight  *Light
	lastLight   *Light
	cachedLight *Light
}

func (s *Scene) Sky() graphics.Sky {
	return s.sky
}

func (s *Scene) CreateCamera() graphics.Camera {
	return newCamera(s)
}

func (s *Scene) CreateDirectionalLight() graphics.DirectionalLight {
	var light *Light
	if s.cachedLight != nil {
		light = s.cachedLight
		s.cachedLight = s.cachedLight.next
	} else {
		light = &Light{}
	}
	light.mode = LightModeDirectional
	light.Node = internal.NewNode()
	light.scene = s
	light.prev = nil
	light.next = nil
	light.intensity = sprec.NewVec3(1.0, 1.0, 1.0)
	s.attachLight(light)
	return light
}

func (s *Scene) CreateAmbientLight() graphics.AmbientLight {
	var light *Light
	if s.cachedLight != nil {
		light = s.cachedLight
		s.cachedLight = s.cachedLight.next
	} else {
		light = &Light{}
	}
	light.mode = LightModeAmbient
	light.Node = internal.NewNode()
	light.scene = s
	light.prev = nil
	light.next = nil
	light.intensity = sprec.NewVec3(1.0, 1.0, 1.0)
	s.attachLight(light)
	return light
}

func (s *Scene) CreateMesh(template graphics.MeshTemplate) graphics.Mesh {
	var mesh *Mesh
	if s.cachedMesh != nil {
		mesh = s.cachedMesh
		s.cachedMesh = s.cachedMesh.next
	} else {
		mesh = &Mesh{}
	}
	mesh.Node = internal.NewNode()
	mesh.scene = s
	mesh.template = template.(*MeshTemplate)
	mesh.prev = nil
	mesh.next = nil
	s.attachMesh(mesh)
	return mesh
}

func (s *Scene) Render(viewport graphics.Viewport, camera graphics.Camera) {
	gfxCamera := camera.(*Camera)
	s.renderer.Render(viewport, s, gfxCamera)
}

func (s *Scene) Delete() {
	s.firstMesh = nil
	s.lastMesh = nil
	s.cachedMesh = nil

	s.firstLight = nil
	s.lastLight = nil
	s.cachedLight = nil
}

func (s *Scene) attachMesh(mesh *Mesh) {
	if s.firstMesh == nil {
		s.firstMesh = mesh
	}
	if s.lastMesh != nil {
		s.lastMesh.next = mesh
		mesh.prev = s.lastMesh
	}
	mesh.next = nil
	s.lastMesh = mesh
}

func (s *Scene) detachMesh(mesh *Mesh) {
	if s.firstMesh == mesh {
		s.firstMesh = mesh.next
	}
	if s.lastMesh == mesh {
		s.lastMesh = mesh.prev
	}
	if mesh.next != nil {
		mesh.next.prev = mesh.prev
	}
	if mesh.prev != nil {
		mesh.prev.next = mesh.next
	}
	mesh.prev = nil
	mesh.next = nil
}

func (s *Scene) cacheMesh(mesh *Mesh) {
	mesh.next = s.cachedMesh
	s.cachedMesh = mesh
}

func (s *Scene) attachLight(light *Light) {
	if s.firstLight == nil {
		s.firstLight = light
	}
	if s.lastLight != nil {
		s.lastLight.next = light
		light.prev = s.lastLight
	}
	light.next = nil
	s.lastLight = light
}

func (s *Scene) detachLight(light *Light) {
	if s.firstLight == light {
		s.firstLight = light.next
	}
	if s.lastLight == light {
		s.lastLight = light.prev
	}
	if light.next != nil {
		light.next.prev = light.prev
	}
	if light.prev != nil {
		light.prev.next = light.next
	}
	light.prev = nil
	light.next = nil
}

func (s *Scene) cacheLight(light *Light) {
	light.next = s.cachedLight
	s.cachedLight = light
}
