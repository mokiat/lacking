package graphics

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/util/blob"
	"github.com/mokiat/lacking/util/shape"
	"github.com/mokiat/lacking/util/spatial"
)

func newScene(renderer *sceneRenderer) *Scene {
	return &Scene{
		renderer: renderer,

		sky: newSky(),

		meshOctree: spatial.NewOctree[*Mesh](32000.0, 9, 2_000_000),
	}
}

// Scene represents a collection of 3D render entities
// that comprise a single visual scene.
type Scene struct {
	renderer *sceneRenderer

	sky *Sky

	meshOctree *spatial.Octree[*Mesh]
	firstMesh  *Mesh
	lastMesh   *Mesh
	cachedMesh *Mesh

	firstLight  *Light
	lastLight   *Light
	cachedLight *Light

	activeCamera *Camera
}

func (s *Scene) ActiveCamera() *Camera {
	return s.activeCamera
}

func (s *Scene) SetActiveCamera(camera *Camera) {
	s.activeCamera = camera
}

// Sky returns this scene's sky object.
// You can use the Sky object to control the
// background appearance.
func (s *Scene) Sky() *Sky {
	return s.sky
}

// CreateCamera creates a new camera object to be
// used with this scene.
func (s *Scene) CreateCamera() *Camera {
	result := newCamera(s)
	if s.activeCamera == nil {
		s.activeCamera = result
	}
	return result
}

// CreateDirectionalLight creates a new directional light object to be
// used within this scene.
func (s *Scene) CreateDirectionalLight() *Light {
	var light *Light
	if s.cachedLight != nil {
		light = s.cachedLight
		s.cachedLight = s.cachedLight.next
	} else {
		light = &Light{}
	}
	light.mode = LightModeDirectional
	light.Node = *newNode()
	light.scene = s
	light.prev = nil
	light.next = nil
	light.intensity = sprec.NewVec3(1.0, 1.0, 1.0)
	s.attachLight(light)
	return light
}

// CreateAmbientLight creates a new ambient light object to be used
// within this scene.
func (s *Scene) CreateAmbientLight() *Light {
	var light *Light
	if s.cachedLight != nil {
		light = s.cachedLight
		s.cachedLight = s.cachedLight.next
	} else {
		light = &Light{}
	}
	light.mode = LightModeAmbient
	light.Node = *newNode()
	light.scene = s
	light.prev = nil
	light.next = nil
	light.intensity = sprec.NewVec3(1.0, 1.0, 1.0)
	s.attachLight(light)
	return light
}

// CreatePointLight creates a new ambient light object to be used
// within this scene.
func (s *Scene) CreatePointLight() *Light {
	var light *Light
	if s.cachedLight != nil {
		light = s.cachedLight
		s.cachedLight = s.cachedLight.next
	} else {
		light = &Light{}
	}
	light.mode = LightModePoint
	light.Node = *newNode()
	light.scene = s
	light.prev = nil
	light.next = nil
	light.intensity = sprec.NewVec3(1.0, 1.0, 1.0)
	s.attachLight(light)
	return light
}

// CreateMesh creates a new mesh instance from the specified
// template and places it in the scene.
func (s *Scene) CreateMesh(info MeshInfo) *Mesh {
	var mesh *Mesh
	if s.cachedMesh != nil {
		mesh = s.cachedMesh
		s.cachedMesh = s.cachedMesh.next
	} else {
		mesh = &Mesh{}
	}

	definition := info.Definition
	mesh.Node = *newNode()
	mesh.item = s.meshOctree.CreateItem(mesh)
	mesh.item.SetRadius(definition.boundingSphereRadius)
	mesh.scene = s
	mesh.prev = nil
	mesh.next = nil
	mesh.definition = definition
	mesh.armature = info.Armature
	s.attachMesh(mesh)
	return mesh
}

func (s *Scene) CreateArmature(info ArmatureInfo) *Armature {
	boneCount := len(info.InverseMatrices)
	return &Armature{
		inverseMatrices:   info.InverseMatrices,
		uniformBufferData: make(blob.Buffer, boneCount*64),
	}
}

// Render draws this scene to the specified viewport
// looking through the specified camera.
func (s *Scene) Render(viewport Viewport) {
	if s.activeCamera != nil {
		s.renderer.Render(s.renderer.api.DefaultFramebuffer(), viewport, s, s.activeCamera)
	}
}

func (s *Scene) Ray(viewport Viewport, camera *Camera, x, y int) shape.StaticLine {
	return s.renderer.Ray(viewport, camera, x, y)
}

// Render draws this scene to the specified viewport
// looking through the specified camera.
func (s *Scene) RenderFramebuffer(framebuffer render.Framebuffer, viewport Viewport) {
	if s.activeCamera != nil {
		s.renderer.Render(framebuffer, viewport, s, s.activeCamera)
	}
}

// Delete removes this scene and releases all
// entities allocated for it.
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
