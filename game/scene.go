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

	return &Scene{
		engine: engine,

		ecsScene:     ecsScene,
		physicsScene: physicsScene,
		gfxScene:     gfxScene,
		root:         hierarchy.NewNode(), // TODO: Make this node stationary

		animationTrees: ds.NewList[animation.Source](0),

		preUpdateSubscriptions:  timestep.NewUpdateSubscriptionSet(),
		postUpdateSubscriptions: timestep.NewUpdateSubscriptionSet(),

		prePhysicsSubscriptions:  timestep.NewUpdateSubscriptionSet(),
		postPhysicsSubscriptions: timestep.NewUpdateSubscriptionSet(),

		preAnimationSubscriptions:  timestep.NewUpdateSubscriptionSet(),
		postAnimationSubscriptions: timestep.NewUpdateSubscriptionSet(),

		preNodeSubscriptions:  timestep.NewUpdateSubscriptionSet(),
		postNodeSubscriptions: timestep.NewUpdateSubscriptionSet(),
	}
}

// Scene is the main container for all game objects and systems.
type Scene struct {
	engine *Engine

	ecsScene     *ecs.Scene
	physicsScene *physics.Scene
	gfxScene     *graphics.Scene
	root         *hierarchy.Node

	animationTrees *ds.List[animation.Source]

	preUpdateSubscriptions  *timestep.UpdateSubscriptionSet
	postUpdateSubscriptions *timestep.UpdateSubscriptionSet

	prePhysicsSubscriptions  *timestep.UpdateSubscriptionSet
	postPhysicsSubscriptions *timestep.UpdateSubscriptionSet

	preAnimationSubscriptions  *timestep.UpdateSubscriptionSet
	postAnimationSubscriptions *timestep.UpdateSubscriptionSet

	preNodeSubscriptions  *timestep.UpdateSubscriptionSet
	postNodeSubscriptions *timestep.UpdateSubscriptionSet

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

// SubscribePreUpdate adds a callback to be executed before the scene updates.
func (s *Scene) SubscribePreUpdate(callback timestep.UpdateCallback) *timestep.UpdateSubscription {
	return s.preUpdateSubscriptions.Subscribe(callback)
}

// SubscribePostUpdate adds a callback to be executed after the scene updates.
func (s *Scene) SubscribePostUpdate(callback timestep.UpdateCallback) *timestep.UpdateSubscription {
	return s.postUpdateSubscriptions.Subscribe(callback)
}

// SubscribePrePhysics adds a callback to be executed before the physics scene
// updates.
func (s *Scene) SubscribePrePhysics(callback timestep.UpdateCallback) *timestep.UpdateSubscription {
	return s.prePhysicsSubscriptions.Subscribe(callback)
}

// SubscribePostPhysics adds a callback to be executed after the physics scene
// updates.
func (s *Scene) SubscribePostPhysics(callback timestep.UpdateCallback) *timestep.UpdateSubscription {
	return s.postPhysicsSubscriptions.Subscribe(callback)
}

// SubscribePreAnimation adds a callback to be executed before the animations
// are updated.
func (s *Scene) SubscribePreAnimation(callback timestep.UpdateCallback) *timestep.UpdateSubscription {
	return s.preAnimationSubscriptions.Subscribe(callback)
}

// SubscribePostAnimation adds a callback to be executed after the animations
// are updated.
func (s *Scene) SubscribePostAnimation(callback timestep.UpdateCallback) *timestep.UpdateSubscription {
	return s.postAnimationSubscriptions.Subscribe(callback)
}

// SubscribePreNode adds a callback to be executed before the nodes are updated.
func (s *Scene) SubscribePreNode(callback timestep.UpdateCallback) *timestep.UpdateSubscription {
	return s.preNodeSubscriptions.Subscribe(callback)
}

// SubscribePostNode adds a callback to be executed after the nodes are updated.
func (s *Scene) SubscribePostNode(callback timestep.UpdateCallback) *timestep.UpdateSubscription {
	return s.postNodeSubscriptions.Subscribe(callback)
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
	if s.frozen {
		return
	}

	preUpdateSpan := metric.BeginRegion("pre-update")
	s.preUpdateSubscriptions.Each(func(callback timestep.UpdateCallback) {
		callback(elapsedTime)
	})
	preUpdateSpan.End()

	updateSpan := metric.BeginRegion("update")
	s.updatePhysics(elapsedTime)
	s.updateAnimations(elapsedTime)
	s.updateNodes(elapsedTime)
	updateSpan.End()

	postUpdateSpan := metric.BeginRegion("post-update")
	s.postUpdateSubscriptions.Each(func(callback timestep.UpdateCallback) {
		callback(elapsedTime)
	})
	postUpdateSpan.End()

	if s.ecsScene != nil {
		s.ecsScene.Purge()
	}
	if s.gfxScene != nil {
		s.gfxScene.Update(elapsedTime)
	}
}

// Render draws the scene to the provided viewport.
func (s *Scene) Render(framebuffer render.Framebuffer, viewport graphics.Viewport) {
	// NOTE: This needs to be here right now because camera systems operate
	// between ApplyFromSource and ApplyToTarget, hence this can't be moved
	// to main Update loop quite yet.
	nodeSpan := metric.BeginRegion("node-apply")
	s.root.ApplyToTarget(true)
	nodeSpan.End()

	if s.gfxScene != nil {
		renderSpan := metric.BeginRegion("render")
		s.gfxScene.Render(framebuffer, viewport)
		renderSpan.End()
	}
}

func (s *Scene) updatePhysics(elapsedTime time.Duration) {
	prePhysicsSpan := metric.BeginRegion("pre-physics")
	s.prePhysicsSubscriptions.Each(func(callback timestep.UpdateCallback) {
		callback(elapsedTime)
	})
	prePhysicsSpan.End()

	physicsSpan := metric.BeginRegion("physics")
	if s.physicsScene != nil {
		s.physicsScene.Update(elapsedTime)
	}
	physicsSpan.End()

	postPhysicsSpan := metric.BeginRegion("post-physics")
	s.postPhysicsSubscriptions.Each(func(callback timestep.UpdateCallback) {
		callback(elapsedTime)
	})
	postPhysicsSpan.End()
}

func (s *Scene) updateAnimations(elapsedTime time.Duration) {
	preAnimationSpan := metric.BeginRegion("pre-anim")
	s.preAnimationSubscriptions.Each(func(callback timestep.UpdateCallback) {
		callback(elapsedTime)
	})
	preAnimationSpan.End()

	animationSpan := metric.BeginRegion("anim")
	s.updateAnimationTrees(elapsedTime)
	animationSpan.End()

	postAnimationSpan := metric.BeginRegion("post-anim")
	s.postAnimationSubscriptions.Each(func(callback timestep.UpdateCallback) {
		callback(elapsedTime)
	})
	postAnimationSpan.End()

}

func (s *Scene) updateNodes(elapsedTime time.Duration) {
	preNodeSpan := metric.BeginRegion("pre-node")
	s.preNodeSubscriptions.Each(func(callback timestep.UpdateCallback) {
		callback(elapsedTime)
	})
	preNodeSpan.End()

	nodeSpan := metric.BeginRegion("node-fetch")
	s.root.ApplyFromSource(true)
	nodeSpan.End()

	postNodeSpan := metric.BeginRegion("post-node")
	s.postNodeSubscriptions.Each(func(callback timestep.UpdateCallback) {
		callback(elapsedTime)
	})
	postNodeSpan.End()
}

func (s *Scene) updateAnimationTrees(elapsedTime time.Duration) {
	for _, tree := range s.animationTrees.Unbox() {
		tree.SetPosition(tree.Position() + elapsedTime.Seconds())
	}
}
