package game

import (
	"time"

	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/lacking/debug/metric"
	"github.com/mokiat/lacking/game/animation"
	"github.com/mokiat/lacking/game/ecs"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/hierarchy"
	"github.com/mokiat/lacking/game/physics"
	"github.com/mokiat/lacking/game/timestep"
	"github.com/mokiat/lacking/render"
)

// ECSScope is the default scope used for all ECS scenes in the game.
var ECSScope = ecs.NewScope()

// SceneInfo specifies details regarding the scene to be created.
type SceneInfo struct {

	// IncludeECS indicates whether an ECS sub-scene would be required.
	//
	// Defaults to `true`.
	IncludeECS opt.T[bool]

	// IncludePhysics indicates whether a Physics sub-scene would be required.
	//
	// Defaults to `true`.
	IncludePhysics opt.T[bool]

	// IncludeGraphics indicates whether a Graphics sub-scene would be required.
	//
	// Defaults to `true`.
	IncludeGraphics opt.T[bool]

	// ECSScope specifies the scope to be used for the ECS sub-scene of the scene.
	//
	// Defaults to the global `ECSScope`.
	ECSScope opt.T[*ecs.Scope]

	// FixedTimestep determines the duration of a single fixed-step tick.
	FixedTimestep opt.T[time.Duration]
}

func newScene(engine *Engine, info SceneInfo) *Scene {
	var (
		includeHierarchy = true
		includeECS       = info.IncludeECS.ValueOrDefault(true)
		includePhysics   = info.IncludePhysics.ValueOrDefault(true)
		includeGraphics  = info.IncludeGraphics.ValueOrDefault(true)
	)

	var hierarchyScene *hierarchy.Scene
	if includeHierarchy {
		hierarchyScene = hierarchy.NewScene()
	}

	var ecsScene *ecs.Scene
	if includeECS {
		ecsScene = ecs.NewScene(info.ECSScope.ValueOrDefault(ECSScope))
	}

	var physicsScene *physics.Scene
	if includePhysics {
		physicsScene = physics.NewScene()
	}

	var gfxScene *graphics.Scene
	if gfxEngine := engine.Graphics(); gfxEngine != nil && includeGraphics {
		gfxScene = gfxEngine.CreateScene()
	}

	fixedTimestep := info.FixedTimestep.ValueOrDefault(16 * time.Millisecond)

	armatureBindingSet := hierarchy.NewBinding[*animation.Player](hierarchyScene, NewAnimationBindingSolver())
	bodyBindingSet := hierarchy.NewBinding[physics.BodyID](hierarchyScene, NewBodyBindingSolver(physicsScene))
	skyBindingSet := hierarchy.NewBinding[*graphics.Sky](hierarchyScene, NewSkyBindingSolver())
	ambientLightBindingSet := hierarchy.NewBinding[*graphics.AmbientLight](hierarchyScene, NewAmbientLightBindingSolver())
	pointLightBindingSet := hierarchy.NewBinding[*graphics.PointLight](hierarchyScene, NewPointLightBindingSolver())
	spotLightBindingSet := hierarchy.NewBinding[*graphics.SpotLight](hierarchyScene, NewSpotLightBindingSolver())
	directionalLightBindingSet := hierarchy.NewBinding[*graphics.DirectionalLight](hierarchyScene, NewDirectionalLightBindingSolver())
	meshBindingSet := hierarchy.NewBinding[*graphics.Mesh](hierarchyScene, NewMeshBindingSolver())
	boneBindingSet := hierarchy.NewBinding[BoneTarget](hierarchyScene, NewBoneBindingSolver())
	cameraBindingSet := hierarchy.NewBinding[*graphics.Camera](hierarchyScene, NewCameraBindingSolver())

	return &Scene{
		engine: engine,

		hierarchyScene: hierarchyScene,
		ecsScene:       ecsScene,
		physicsScene:   physicsScene,
		gfxScene:       gfxScene,

		armatureBindingSet:         armatureBindingSet,
		bodyBindingSet:             bodyBindingSet,
		skyBindingSet:              skyBindingSet,
		ambientLightBindingSet:     ambientLightBindingSet,
		pointLightBindingSet:       pointLightBindingSet,
		spotLightBindingSet:        spotLightBindingSet,
		directionalLightBindingSet: directionalLightBindingSet,
		meshBindingSet:             meshBindingSet,
		boneBindingSet:             boneBindingSet,
		cameraBindingSet:           cameraBindingSet,

		timeSegmenter: timestep.NewSegmenter(fixedTimestep),

		animationTrees: ds.EmptyList[*animation.Player](),

		stepUpdateSubscriptions:    timestep.NewStepSubscriptionSet(),
		fixedUpdateSubscriptions:   timestep.NewUpdateSubscriptionSet(),
		interpolationSubscriptions: timestep.NewInterpolationSubscriptionSet(),
		updateSubscriptions:        timestep.NewUpdateSubscriptionSet(),

		frozen: false,
	}
}

// Scene is the main container for all game objects and systems.
type Scene struct {
	engine *Engine

	hierarchyScene *hierarchy.Scene
	ecsScene       *ecs.Scene
	physicsScene   *physics.Scene
	gfxScene       *graphics.Scene

	armatureBindingSet         *hierarchy.Binding[*animation.Player]
	bodyBindingSet             *hierarchy.Binding[physics.BodyID]
	skyBindingSet              *hierarchy.Binding[*graphics.Sky]
	ambientLightBindingSet     *hierarchy.Binding[*graphics.AmbientLight]
	pointLightBindingSet       *hierarchy.Binding[*graphics.PointLight]
	spotLightBindingSet        *hierarchy.Binding[*graphics.SpotLight]
	directionalLightBindingSet *hierarchy.Binding[*graphics.DirectionalLight]
	meshBindingSet             *hierarchy.Binding[*graphics.Mesh]
	boneBindingSet             *hierarchy.Binding[BoneTarget]
	cameraBindingSet           *hierarchy.Binding[*graphics.Camera]

	timeSegmenter *timestep.Segmenter

	animationTrees *ds.List[*animation.Player]

	stepUpdateSubscriptions    *timestep.StepSubscriptionSet
	fixedUpdateSubscriptions   *timestep.UpdateSubscriptionSet
	interpolationSubscriptions *timestep.InterpolationSubscriptionSet
	updateSubscriptions        *timestep.UpdateSubscriptionSet

	frozen bool
}

// Delete removes all resources associated with the scene.
func (s *Scene) Delete() {
	if s.ecsScene != nil {
		defer s.ecsScene.Delete()
	}
	s.physicsScene = nil
	if s.gfxScene != nil {
		defer s.gfxScene.Delete()
	}
	s.engine.SetActiveScene(nil)
	s.engine = nil
}

// Engine returns the engine associated with the scene.
func (s *Scene) Engine() *Engine {
	return s.engine
}

// Hierarchy returns the Hierarchy sub-scene associated with this scene.
func (s *Scene) Hierarchy() *hierarchy.Scene {
	return s.hierarchyScene
}

// ECS returns the ECS sub-scene associated with this scene.
//
// Returns `nil` if this scene does not have ECS enabled.
func (s *Scene) ECS() *ecs.Scene {
	return s.ecsScene
}

// Physics returns the Physics sub-scene associated with this scene.
//
// Returns `nil` if this scene does not have Physics enabled.
func (s *Scene) Physics() *physics.Scene {
	return s.physicsScene
}

// Graphics returns the Graphics sub-scene associated with the scene.
//
// Returns `nil` if this scene does not have Graphics enabled.
func (s *Scene) Graphics() *graphics.Scene {
	return s.gfxScene
}

// SubscribeStepUpdate adds a callback to be executed before fixed time updates
func (s *Scene) SubscribeStepUpdate(callback timestep.StepCallback) *timestep.StepSubscription {
	return s.stepUpdateSubscriptions.Subscribe(callback)
}

// SubscribeFixedUpdate adds a callback to be executed after each fixed time
// update.
func (s *Scene) SubscribeFixedUpdate(callback timestep.UpdateCallback) *timestep.UpdateSubscription {
	return s.fixedUpdateSubscriptions.Subscribe(callback)
}

// SubscribeInterpolation adds a callback to be executed after a series of
// fixed time updates are performed and interpolation is to be performed.
func (s *Scene) SubscribeInterpolation(callback timestep.InterpolationCallback) *timestep.InterpolationSubscription {
	return s.interpolationSubscriptions.Subscribe(callback)
}

// SubscribeUpdate adds a callback to be executed after each dynamic time
// update.
func (s *Scene) SubscribeUpdate(callback timestep.UpdateCallback) *timestep.UpdateSubscription {
	return s.updateSubscriptions.Subscribe(callback)
}

// ArmatureBindingSet returns the binding set that binds animation players
// to armatures.
func (s *Scene) ArmatureBindingSet() *hierarchy.Binding[*animation.Player] {
	return s.armatureBindingSet
}

// BodyBindingSet returns the binding set that binds physics bodies.
func (s *Scene) BodyBindingSet() *hierarchy.Binding[physics.BodyID] {
	return s.bodyBindingSet
}

// SkyBindingSet returns the binding set that binds sky objects.
func (s *Scene) SkyBindingSet() *hierarchy.Binding[*graphics.Sky] {
	return s.skyBindingSet
}

// AmbientLightBindingSet returns the binding set that binds ambient light objects.
func (s *Scene) AmbientLightBindingSet() *hierarchy.Binding[*graphics.AmbientLight] {
	return s.ambientLightBindingSet
}

// PointLightBindingSet returns the binding set that binds point light objects.
func (s *Scene) PointLightBindingSet() *hierarchy.Binding[*graphics.PointLight] {
	return s.pointLightBindingSet
}

// SpotLightBindingSet returns the binding set that binds spot light objects.
func (s *Scene) SpotLightBindingSet() *hierarchy.Binding[*graphics.SpotLight] {
	return s.spotLightBindingSet
}

// DirectionalLightBindingSet returns the binding set that binds directional
// light objects.
func (s *Scene) DirectionalLightBindingSet() *hierarchy.Binding[*graphics.DirectionalLight] {
	return s.directionalLightBindingSet
}

// MeshBindingSet returns the binding set that binds mesh objects.
func (s *Scene) MeshBindingSet() *hierarchy.Binding[*graphics.Mesh] {
	return s.meshBindingSet
}

// BoneBindingSet returns the binding set that binds bone target objects.
func (s *Scene) BoneBindingSet() *hierarchy.Binding[BoneTarget] {
	return s.boneBindingSet
}

// CameraBindingSet returns the binding set that binds camera objects.
func (s *Scene) CameraBindingSet() *hierarchy.Binding[*graphics.Camera] {
	return s.cameraBindingSet
}

// IsFrozen returns whether the scene is currently frozen. A frozen scene
// will not update any of its systems.
func (s *Scene) IsFrozen() bool {
	return s.frozen
}

// Freeze stops the scene from updating any of its systems.
func (s *Scene) Freeze() {
	s.frozen = true
}

// Unfreeze allows the scene to update its systems.
func (s *Scene) Unfreeze() {
	s.frozen = false
}

// PlayAnimation adds the provided animation player to the scene.
func (s *Scene) PlayAnimation(player *animation.Player) {
	s.animationTrees.Add(player)
}

// StopAnimationTree removes the provided animation player from the scene.
func (s *Scene) StopAnimation(player *animation.Player) {
	s.animationTrees.Remove(player)
}

// Update advances the scene by the provided time.
func (s *Scene) Update(elapsedTime time.Duration) {
	if !s.frozen {
		s.timeSegmenter.SetStepCallback(s.doStepUpdate)
		s.timeSegmenter.SetFixedCallback(s.doFixedUpdate)
		s.timeSegmenter.SetInterpCallback(s.doInterpolationUpdate)
		s.timeSegmenter.Update(elapsedTime)
		s.doUpdate(elapsedTime)
	}
}

// Render draws the scene to the provided viewport.
func (s *Scene) Render(framebuffer render.Framebuffer, viewport graphics.Viewport) {
	if s.gfxScene != nil {
		renderSpan := metric.BeginRegion("render")
		s.gfxScene.Render(framebuffer, viewport)
		renderSpan.End()
	}
}

func (s *Scene) doStepUpdate(steps float64) {
	callbackSpan := metric.BeginRegion("step-cb")
	s.stepUpdateSubscriptions.Each(func(callback timestep.StepCallback) {
		callback(steps)
	})
	callbackSpan.End()
}

func (s *Scene) doFixedUpdate(elapsedTime time.Duration) {
	resetSpan := metric.BeginRegion("node-fixed")
	s.hierarchyScene.AdvanceStep()
	resetSpan.End()

	physicsSpan := metric.BeginRegion("physics")
	if s.physicsScene != nil {
		s.physicsScene.Update(elapsedTime)
	}
	physicsSpan.End()

	callbackSpan := metric.BeginRegion("fixed-cb")
	s.fixedUpdateSubscriptions.Each(func(callback timestep.UpdateCallback) {
		callback(elapsedTime)
	})
	callbackSpan.End()

	nodeSpan := metric.BeginRegion("node-source")
	s.hierarchyScene.ApplySourcesToNodes()
	nodeSpan.End()

	animationSpan := metric.BeginRegion("anim")
	s.updateAnimationTrees(elapsedTime)
	animationSpan.End()

	nodeSpan = metric.BeginRegion("node-target")
	s.hierarchyScene.ApplyTargetsFromNodes()
	nodeSpan.End()
}

func (s *Scene) doInterpolationUpdate(fraction float64) {
	nodeSpan := metric.BeginRegion("node-interp")
	s.hierarchyScene.ApplyInterpolationsFromNodes(fraction)
	nodeSpan.End()

	callbackSpan := metric.BeginRegion("interp-cb")
	s.interpolationSubscriptions.Each(func(callback timestep.InterpolationCallback) {
		callback(fraction)
	})
	callbackSpan.End()
}

func (s *Scene) doUpdate(elapsedTime time.Duration) {
	callbackSpan := metric.BeginRegion("update-cb")
	s.updateSubscriptions.Each(func(callback timestep.UpdateCallback) {
		callback(elapsedTime)
	})
	callbackSpan.End()

	if s.gfxScene != nil {
		s.gfxScene.Update(elapsedTime)
	}
}

func (s *Scene) updateAnimationTrees(elapsedTime time.Duration) {
	for _, player := range s.animationTrees.Unbox() {
		player.Update(elapsedTime)
	}
}
