package graphics

import (
	"github.com/mokiat/gblob"
	"github.com/mokiat/gog/ds"
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/util/shape"
	"github.com/mokiat/lacking/util/spatial"
)

const (
	maxSceneSize = 32_000.0
)

func newScene(renderer *sceneRenderer) *Scene {
	return &Scene{
		renderer: renderer,

		sky: newSky(),

		meshOctree: spatial.NewOctree[*Mesh](maxSceneSize, 15),
		meshPool:   ds.NewPool[Mesh](),

		ambientLightOctree: spatial.NewOctree[*AmbientLight](maxSceneSize, 15),
		ambientLightPool:   ds.NewPool[AmbientLight](),

		pointLightOctree: spatial.NewOctree[*PointLight](maxSceneSize, 15),
		pointLightPool:   ds.NewPool[PointLight](),

		spotLightOctree: spatial.NewOctree[*SpotLight](maxSceneSize, 15),
		spotLightPool:   ds.NewPool[SpotLight](),

		directionalLightOctree: spatial.NewOctree[*DirectionalLight](maxSceneSize, 15),
		directionalLightPool:   ds.NewPool[DirectionalLight](),
	}
}

// Scene represents a collection of 3D render entities
// that comprise a single visual scene.
type Scene struct {
	renderer *sceneRenderer

	sky *Sky

	meshOctree *spatial.Octree[*Mesh]
	meshPool   *ds.Pool[Mesh]

	ambientLightOctree *spatial.Octree[*AmbientLight]
	ambientLightPool   *ds.Pool[AmbientLight]

	pointLightOctree *spatial.Octree[*PointLight]
	pointLightPool   *ds.Pool[PointLight]

	spotLightOctree *spatial.Octree[*SpotLight]
	spotLightPool   *ds.Pool[SpotLight]

	directionalLightOctree *spatial.Octree[*DirectionalLight]
	directionalLightPool   *ds.Pool[DirectionalLight]

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

// CreateAmbientLight creates a new AmbientLight object to be used
// within this scene.
func (s *Scene) CreateAmbientLight(info AmbientLightInfo) *AmbientLight {
	return newAmbientLight(s, info)
}

// CreatePointLight creates a new PointLight object to be used
// within this scene.
func (s *Scene) CreatePointLight(info PointLightInfo) *PointLight {
	return newPointLight(s, info)
}

// CreateSpotLight creates a new SpotLight object to be used
// within this scene.
func (s *Scene) CreateSpotLight(info SpotLightInfo) *SpotLight {
	return newSpotLight(s, info)
}

// CreateDirectionalLight creates a new DirectionalLight object to be
// used within this scene.
func (s *Scene) CreateDirectionalLight(info DirectionalLightInfo) *DirectionalLight {
	return newDirectionalLight(s, info)
}

// CreateMesh creates a new mesh instance from the specified
// template and places it in the scene.
func (s *Scene) CreateMesh(info MeshInfo) *Mesh {
	return newMesh(s, info)
}

func (s *Scene) CreateArmature(info ArmatureInfo) *Armature {
	boneCount := len(info.InverseMatrices)
	return &Armature{
		inverseMatrices:   info.InverseMatrices,
		uniformBufferData: make(gblob.LittleEndianBlock, boneCount*64),
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
	s.ambientLightPool = nil
	s.pointLightPool = nil
	s.spotLightPool = nil
	s.directionalLightPool = nil
}
