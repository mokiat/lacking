package graphics

import (
	"time"

	"github.com/mokiat/gblob"
	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/util/spatial"
)

const (
	maxSceneSize = 32_000.0
)

func newScene(engine *Engine, renderer *sceneRenderer) *Scene {
	return &Scene{
		engine:   engine,
		renderer: renderer,

		skies: ds.NewList[*Sky](1),

		staticMeshOctree: spatial.NewStaticOctree[uint32](spatial.StaticOctreeSettings{
			Size:                opt.V(maxSceneSize),
			MaxDepth:            opt.V(int32(15)),
			BiasRatio:           opt.V(4.0),
			InitialNodeCapacity: opt.V(int32(128 * 1024)),
			InitialItemCapacity: opt.V(int32(1024 * 1024)),
		}),

		dynamicMeshPool: ds.NewPool[Mesh](),
		dynamicMeshSet: spatial.NewDynamicSet[*Mesh](spatial.DynamicSetSettings{
			InitialItemCapacity: opt.V(int32(1024)),
		}),

		ambientLightPool: ds.NewPool[AmbientLight](),
		ambientLightSet: spatial.NewDynamicSet[*AmbientLight](spatial.DynamicSetSettings{
			InitialItemCapacity: opt.V(int32(4)),
		}),

		pointLightPool: ds.NewPool[PointLight](),
		pointLightSet: spatial.NewDynamicSet[*PointLight](spatial.DynamicSetSettings{
			InitialItemCapacity: opt.V(int32(128)),
		}),

		spotLightPool: ds.NewPool[SpotLight](),
		spotLightSet: spatial.NewDynamicSet[*SpotLight](spatial.DynamicSetSettings{
			InitialItemCapacity: opt.V(int32(128)),
		}),

		directionalLightPool: ds.NewPool[DirectionalLight](),
		directionalLightSet: spatial.NewDynamicSet[*DirectionalLight](spatial.DynamicSetSettings{
			InitialItemCapacity: opt.V(int32(16)),
		}),
	}
}

// Scene represents a collection of 3D render entities
// that comprise a single visual scene.
type Scene struct {
	engine   *Engine
	renderer *sceneRenderer

	time float32

	skies *ds.List[*Sky]

	staticMeshes     []StaticMesh
	staticMeshOctree *spatial.StaticOctree[uint32]

	dynamicMeshPool *ds.Pool[Mesh]
	dynamicMeshSet  *spatial.DynamicSet[*Mesh]

	ambientLightPool *ds.Pool[AmbientLight]
	ambientLightSet  *spatial.DynamicSet[*AmbientLight]

	pointLightPool *ds.Pool[PointLight]
	pointLightSet  *spatial.DynamicSet[*PointLight]

	spotLightPool *ds.Pool[SpotLight]
	spotLightSet  *spatial.DynamicSet[*SpotLight]

	directionalLightPool *ds.Pool[DirectionalLight]
	directionalLightSet  *spatial.DynamicSet[*DirectionalLight]

	activeCamera *Camera
}

// Engine returns the graphics engine that owns this scene.
func (s *Scene) Engine() *Engine {
	return s.engine
}

func (s *Scene) Time() float32 {
	return s.time
}

// ActiveCamera returns the currently active camera for this scene.
func (s *Scene) ActiveCamera() *Camera {
	return s.activeCamera
}

// SetActiveCamera changes the active camera for this scene.
func (s *Scene) SetActiveCamera(camera *Camera) {
	s.activeCamera = camera
}

// CreateCamera creates a new camera object to be
// used with this scene.
func (s *Scene) CreateCamera() *Camera {
	result := newCamera()
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

// CreateSky creates a new Sky object to be used within this scene.
func (s *Scene) CreateSky(info SkyInfo) *Sky {
	return newSky(s, info)
}

// CreateStaticMesh creates a new static mesh to be rendered in this scene.
//
// Static meshes cannot be removed from a scene but are rendered more
// efficiently.
func (s *Scene) CreateStaticMesh(info StaticMeshInfo) {
	createStaticMesh(s, info)
}

// CreateMesh creates a new mesh instance from the specified
// template and places it in the scene.
func (s *Scene) CreateMesh(info MeshInfo) *Mesh {
	return newMesh(s, info)
}

// CreateArmature creates an armature to be used with meshes.
func (s *Scene) CreateArmature(info ArmatureInfo) *Armature {
	boneCount := len(info.InverseMatrices)
	return &Armature{
		inverseMatrices:   info.InverseMatrices,
		uniformBufferData: make(gblob.LittleEndianBlock, boneCount*64),
	}
}

func (s *Scene) Ray(viewport Viewport, camera *Camera, x, y int) (dprec.Vec3, dprec.Vec3) {
	return s.renderer.Ray(viewport, camera, x, y)
}

func (s *Scene) Point(viewport Viewport, camera *Camera, position dprec.Vec3) dprec.Vec2 {
	return s.renderer.Point(viewport, camera, position)
}

func (s *Scene) Update(elapsedTime time.Duration) {
	s.time += float32(elapsedTime.Seconds())
}

// Render draws this scene to the specified viewport
// looking through the specified camera.
func (s *Scene) Render(framebuffer render.Framebuffer, viewport Viewport) {
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
