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

	// FixedTimestep determines the duration of a single fixed-step tick.
	FixedTimestep opt.T[time.Duration]
}

func newScene(engine *Engine, info SceneInfo) *Scene {
	var (
		includeECS      = info.IncludeECS.ValueOrDefault(true)
		includePhysics  = info.IncludePhysics.ValueOrDefault(true)
		includeGraphics = info.IncludeGraphics.ValueOrDefault(true)
	)

	var ecsScene *ecs.Scene
	if ecsEngine := engine.ECS(); ecsEngine != nil && includeECS {
		ecsScene = ecsEngine.CreateScene()
	}

	var physicsScene *physics.Scene
	if physicsEngine := engine.Physics(); physicsEngine != nil && includePhysics {
		physicsScene = physicsEngine.CreateScene()
	}

	var gfxScene *graphics.Scene
	if gfxEngine := engine.Graphics(); gfxEngine != nil && includeGraphics {
		gfxScene = gfxEngine.CreateScene()
	}

	fixedTimestep := info.FixedTimestep.ValueOrDefault(16 * time.Millisecond)

	return &Scene{
		engine: engine,

		ecsScene:     ecsScene,
		physicsScene: physicsScene,
		gfxScene:     gfxScene,

		fixedTimestep: fixedTimestep,
		timeSegmenter: timestep.NewSegmenter(fixedTimestep),

		root:           hierarchy.NewNode(), // TODO: Make this node stationary
		animationTrees: ds.NewList[animation.Source](0),

		fixedUpdateSubscriptions:   timestep.NewUpdateSubscriptionSet(),
		interpolationSubscriptions: timestep.NewInterpolationSubscriptionSet(),
		updateSubscriptions:        timestep.NewUpdateSubscriptionSet(),

		frozen: false,
	}
}

// Scene is the main container for all game objects and systems.
type Scene struct {
	engine *Engine

	ecsScene     *ecs.Scene
	physicsScene *physics.Scene
	gfxScene     *graphics.Scene

	fixedTimestep time.Duration
	timeSegmenter *timestep.Segmenter

	root           *hierarchy.Node
	animationTrees *ds.List[animation.Source]

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
	if s.physicsScene != nil {
		defer s.physicsScene.Delete()
	}
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

// Root returns the root node of the scene.
func (s *Scene) Root() *hierarchy.Node {
	return s.root
}

// CreateNode creates a new node and appends it to the root of the scene.
func (s *Scene) CreateNode() *hierarchy.Node {
	result := hierarchy.NewNode()
	s.root.AppendChild(result)
	return result
}

// PlayAnimationTree adds the provided animation tree to the scene.
func (s *Scene) PlayAnimationTree(tree animation.Source) {
	s.animationTrees.Add(tree)
}

// StopAnimationTree removes the provided animation tree from the scene.
func (s *Scene) StopAnimationTree(tree animation.Source) {
	s.animationTrees.Remove(tree)
}

// Update advances the scene by the provided time.
func (s *Scene) Update(elapsedTime time.Duration) {
	if !s.frozen {
		s.timeSegmenter.Update(elapsedTime, s.doFixedUpdate, s.doInterpolationUpdate)
		s.doUpdate(elapsedTime)
	}
}

func (s *Scene) doFixedUpdate(elapsedTime time.Duration) {
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

	if s.ecsScene != nil {
		s.ecsScene.Purge()
	}
}

func (s *Scene) doInterpolationUpdate(fraction float64) {
	nodeSpan := metric.BeginRegion("node-fetch")
	s.root.ApplyFromSource(fraction, true)
	nodeSpan.End()

	callbackSpan := metric.BeginRegion("interp-cb")
	s.interpolationSubscriptions.Each(func(callback timestep.InterpolationCallback) {
		callback(fraction)
	})
	callbackSpan.End()
}

func (s *Scene) doUpdate(elapsedTime time.Duration) {
	animationSpan := metric.BeginRegion("anim")
	s.updateAnimationTrees(elapsedTime)
	animationSpan.End()

	callbackSpan := metric.BeginRegion("update-cb")
	s.updateSubscriptions.Each(func(callback timestep.UpdateCallback) {
		callback(elapsedTime)
	})
	callbackSpan.End()

	nodeSpan := metric.BeginRegion("node-apply")
	s.root.ApplyToTarget(true)
	nodeSpan.End()

	if s.ecsScene != nil {
		s.ecsScene.Purge()
	}

	if s.gfxScene != nil {
		s.gfxScene.Update(elapsedTime)
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

func (s *Scene) updateAnimationTrees(elapsedTime time.Duration) {
	for _, tree := range s.animationTrees.Unbox() {
		tree.SetPosition(tree.Position() + elapsedTime.Seconds())
	}
}
