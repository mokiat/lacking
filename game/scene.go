package game

import (
	"time"

	"github.com/mokiat/gog/ds"
	"github.com/mokiat/lacking/debug/metric"
	"github.com/mokiat/lacking/game/animation"
	"github.com/mokiat/lacking/game/ecs"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/hierarchy"
	"github.com/mokiat/lacking/game/physics"
	"github.com/mokiat/lacking/game/timestep"
	"github.com/mokiat/lacking/render"
)

func newScene(engine *Engine, physicsScene *physics.Scene, gfxScene *graphics.Scene, ecsScene *ecs.Scene) *Scene {
	return &Scene{
		engine: engine,

		physicsScene: physicsScene,
		gfxScene:     gfxScene,
		ecsScene:     ecsScene,
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

	physicsScene *physics.Scene
	gfxScene     *graphics.Scene
	ecsScene     *ecs.Scene
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
	defer s.physicsScene.Delete()
	defer s.gfxScene.Delete()
	defer s.ecsScene.Delete()
	s.engine.SetActiveScene(nil)
	s.engine = nil
}

// Engine returns the engine associated with the scene.
func (s *Scene) Engine() *Engine {
	return s.engine
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

// Physics returns the physics scene associated with the scene.
func (s *Scene) Physics() *physics.Scene {
	return s.physicsScene
}

// Graphics returns the graphics scene associated with the scene.
func (s *Scene) Graphics() *graphics.Scene {
	return s.gfxScene
}

// ECS returns the ECS scene associated with the scene.
func (s *Scene) ECS() *ecs.Scene {
	return s.ecsScene
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

	s.gfxScene.Update(elapsedTime)
}

// Render draws the scene to the provided viewport.
func (s *Scene) Render(framebuffer render.Framebuffer, viewport graphics.Viewport) {
	stageSpan := metric.BeginRegion("stage")
	s.root.ApplyToTarget(true)
	stageSpan.End()

	renderSpan := metric.BeginRegion("render")
	s.gfxScene.Render(framebuffer, viewport)
	renderSpan.End()
}

func (s *Scene) updatePhysics(elapsedTime time.Duration) {
	prePhysicsSpan := metric.BeginRegion("pre-physics")
	s.prePhysicsSubscriptions.Each(func(callback timestep.UpdateCallback) {
		callback(elapsedTime)
	})
	prePhysicsSpan.End()

	physicsSpan := metric.BeginRegion("physics")
	s.physicsScene.Update(elapsedTime)
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

	nodeSpan := metric.BeginRegion("node")
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
