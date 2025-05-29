package game

import (
	"time"

	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/debug/metric"
	"github.com/mokiat/lacking/game/ecs"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/hierarchy"
	"github.com/mokiat/lacking/game/physics"
	"github.com/mokiat/lacking/game/physics/collision"
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

		animationTrees: ds.NewList[AnimationSource](0),

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

	animationTrees *ds.List[AnimationSource]

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

// CreateAmbientLight creates a new ambient light and appends it to the root of
// the scene.
func (s *Scene) CreateAmbientLight(info AmbientLightInfo) *hierarchy.Node {
	node := s.CreateNode()
	s.PlaceAmbientLight(node, info)
	return node
}

// PlaceAmbientLight places an ambient light on the provided node.
func (s *Scene) PlaceAmbientLight(node *hierarchy.Node, info AmbientLightInfo) {
	light := s.gfxScene.CreateAmbientLight(graphics.AmbientLightInfo{
		Position:          dprec.ZeroVec3(),
		InnerRadius:       25000.0,
		OuterRadius:       25000.0,
		ReflectionTexture: info.ReflectionTexture,
		RefractionTexture: info.RefractionTexture,
		CastShadow:        info.CastShadow.ValueOrDefault(false),
	})
	node.SetTarget(AmbientLightNodeTarget{
		Light: light,
	})
	node.ApplyToTarget(false)
}

// CreatePointLight creates a new point light and appends it to the root of the
// scene.
func (s *Scene) CreatePointLight(info PointLightInfo) *hierarchy.Node {
	node := s.CreateNode()
	s.PlacePointLight(node, info)
	return node
}

// PlacePointLight places a point light on the provided node.
func (s *Scene) PlacePointLight(node *hierarchy.Node, info PointLightInfo) {
	light := s.gfxScene.CreatePointLight(graphics.PointLightInfo{
		Position:   dprec.ZeroVec3(),
		EmitColor:  info.EmitColor.ValueOrDefault(dprec.NewVec3(10.0, 0.0, 10.0)),
		EmitRange:  info.EmitDistance.ValueOrDefault(20.0),
		CastShadow: info.CastShadow.ValueOrDefault(false),
	})
	node.SetTarget(PointLightNodeTarget{
		Light: light,
	})
	node.ApplyToTarget(false)
}

// CreateSpotLight creates a new spot light and appends it to the root of the
// scene.
func (s *Scene) CreateSpotLight(info SpotLightInfo) *hierarchy.Node {
	node := s.CreateNode()
	s.PlaceSpotLight(node, info)
	return node
}

// PlaceSpotLight places a spot light on the provided node.
func (s *Scene) PlaceSpotLight(node *hierarchy.Node, info SpotLightInfo) {
	light := s.gfxScene.CreateSpotLight(graphics.SpotLightInfo{
		Position:           dprec.ZeroVec3(),
		Rotation:           dprec.IdentityQuat(),
		EmitColor:          info.EmitColor.ValueOrDefault(dprec.NewVec3(10.0, 0.0, 10.0)),
		EmitRange:          info.EmitDistance.ValueOrDefault(20.0),
		EmitOuterConeAngle: info.EmitOuterConeAngle.ValueOrDefault(dprec.Degrees(60)),
		EmitInnerConeAngle: info.EmitInnerConeAngle.ValueOrDefault(dprec.Degrees(30)),
		CastShadow:         info.CastShadow.ValueOrDefault(false),
	})
	node.SetTarget(SpotLightNodeTarget{
		Light: light,
	})
	node.ApplyToTarget(false)
}

// CreateDirectionalLight creates a new directional light and appends it to the
// root of the scene.
func (s *Scene) CreateDirectionalLight(info DirectionalLightInfo) *hierarchy.Node {
	node := s.CreateNode()
	s.PlaceDirectionalLight(node, info)
	return node
}

// PlaceDirectionalLight places a directional light on the provided node.
func (s *Scene) PlaceDirectionalLight(node *hierarchy.Node, info DirectionalLightInfo) {
	light := s.gfxScene.CreateDirectionalLight(graphics.DirectionalLightInfo{
		Position:   dprec.ZeroVec3(),
		Rotation:   dprec.IdentityQuat(),
		EmitColor:  info.EmitColor.ValueOrDefault(dprec.NewVec3(10.0, 0.0, 10.0)),
		CastShadow: info.CastShadow.ValueOrDefault(false),
	})
	node.SetTarget(DirectionalLightNodeTarget{
		Light: light,
	})
	node.ApplyToTarget(false)
}

// CreateAnimation creates a new animation based on the provided information.
func (s *Scene) CreateAnimation(info AnimationInfo) *Animation {
	def := info.Definition
	return &Animation{
		name:      def.name,
		startTime: info.ClipStart.ValueOrDefault(def.startTime),
		endTime:   info.ClipEnd.ValueOrDefault(def.endTime),
		loop:      info.Loop.ValueOrDefault(def.loop),
		bindings:  def.bindings,
	}
}

// PlayAnimationTree adds the provided animation tree to the scene.
func (s *Scene) PlayAnimationTree(tree AnimationSource) {
	s.animationTrees.Add(tree)
}

// StopAnimationTree removes the provided animation tree from the scene.
func (s *Scene) StopAnimationTree(tree AnimationSource) {
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

// TODO: Return the node instead and have the Model be a target?
func (s *Scene) CreateModel(info ModelInfo) *Model {
	hierarchyInfo := HierarchyInstanceInfo{
		Template:      info.Definition.hierarchy,
		Name:          opt.V(info.Name),
		Position:      info.Position,
		Rotation:      info.Rotation,
		Scale:         info.Scale,
		SubTreeNode:   info.RootNode,
		AttachToScene: opt.V(info.IsDynamic),
	}

	hierarchyInstance := s.InstantiateHierarchy(hierarchyInfo)
	modelNode := hierarchyInstance.RootNode
	nodes := hierarchyInstance.Nodes

	definition := info.Definition

	// TODO: Move after bodies are created? But maybe only after pos/rot of bodies
	// is implemented correctly. Right now it does not seem to do anything.
	modelNode.ApplyFromSource(true)

	animations := make([]*Animation, len(definition.animations))
	for i, animationDef := range definition.animations {
		animations[i] = s.CreateAnimation(AnimationInfo{
			Definition: animationDef,
		})
	}

	armatures := make([]*graphics.Armature, len(definition.armatures))
	for i, instance := range definition.armatures {
		armature := s.gfxScene.CreateArmature(graphics.ArmatureInfo{
			InverseMatrices: instance.InverseBindMatrices(),
		})
		for j, joint := range instance.Joints {
			if jointNode, ok := nodes.FindByID(joint.NodeID); ok {
				jointNode.SetTarget(BoneNodeTarget{
					Armature:  armature,
					BoneIndex: j,
				})
			}
		}
		armatures[i] = armature
	}

	// TODO: Track mesh instances?
	for _, instance := range definition.meshes {
		if meshNode, ok := nodes.FindByID(instance.NodeID); ok {
			var armature *graphics.Armature
			if instance.ArmatureIndex >= 0 {
				armature = armatures[instance.ArmatureIndex]
			}
			meshDefinition := definition.meshDefinitions[instance.DefinitionIndex]

			// TODO: Base this on node flags
			if info.IsDynamic {
				mesh := s.gfxScene.CreateMesh(graphics.MeshInfo{
					Definition: meshDefinition,
					Armature:   armature,
				})
				mesh.SetMatrix(meshNode.AbsoluteMatrix())
				meshNode.SetTarget(MeshNodeTarget{
					Mesh: mesh,
				})
			} else {
				s.gfxScene.CreateStaticMesh(graphics.StaticMeshInfo{
					Definition: meshDefinition,
					Armature:   armature,
					Matrix:     meshNode.AbsoluteMatrix(),
				})
			}
		}
	}

	var bodyInstances []physics.Body
	for _, instance := range definition.bodies {
		if bodyNode, ok := nodes.FindByID(instance.NodeID); ok {
			bodyDefinition := definition.bodyDefinitions[instance.DefinitionIndex]
			if info.IsDynamic {
				body := s.physicsScene.CreateBody(physics.BodyInfo{
					Name:       bodyNode.Name(),
					Definition: bodyDefinition,
					// TODO: Initialize from body node matrix?
					Position: dprec.ZeroVec3(),
					Rotation: dprec.IdentityQuat(),
				})
				bodyNode.SetSource(BodyNodeSource{
					Body: body,
				})
				bodyInstances = append(bodyInstances, body)
			} else {
				absMatrix := bodyNode.AbsoluteMatrix()
				transform := collision.TRTransform(absMatrix.Translation(), absMatrix.Rotation())
				collisionSet := collision.NewSet()
				collisionSet.Replace(bodyDefinition.CollisionSet(), transform)
				s.physicsScene.CreateProp(physics.PropInfo{
					Name:         bodyNode.Name(),
					CollisionSet: collisionSet,
				})
			}
		}
	}

	for _, instance := range definition.ambientLights {
		if node, ok := nodes.FindByID(instance.nodeID); ok {
			info := AmbientLightInfo{
				ReflectionTexture: definition.textures[instance.reflectionTextureID],
				RefractionTexture: definition.textures[instance.refractionTextureID],
				OuterRadius:       opt.Unspecified[float64](),
				InnerRadius:       opt.Unspecified[float64](),
				CastShadow:        opt.V(instance.castShadow),
			}
			s.PlaceAmbientLight(node, info)
		}
	}
	for _, instance := range definition.pointLights {
		if node, ok := nodes.FindByID(instance.nodeID); ok {
			info := PointLightInfo{
				EmitColor:    opt.V(instance.emitColor),
				EmitDistance: opt.V(instance.emitDistance),
				CastShadow:   opt.V(instance.castShadow),
			}
			s.PlacePointLight(node, info)
		}
	}
	for _, instance := range definition.spotLights {
		if node, ok := nodes.FindByID(instance.nodeID); ok {
			info := SpotLightInfo{
				EmitColor:          opt.V(instance.emitColor),
				EmitDistance:       opt.V(instance.emitDistance),
				EmitOuterConeAngle: opt.V(instance.emitAngleOuter),
				EmitInnerConeAngle: opt.V(instance.emitAngleInner),
				CastShadow:         opt.V(instance.castShadow),
			}
			s.PlaceSpotLight(node, info)
		}
	}
	for _, instance := range definition.directionalLights {
		if node, ok := nodes.FindByID(instance.nodeID); ok {
			info := DirectionalLightInfo{
				EmitColor:  opt.V(instance.emitColor),
				CastShadow: opt.V(instance.castShadow),
			}
			s.PlaceDirectionalLight(node, info)
		}
	}
	for _, instance := range definition.skies {
		if node, ok := nodes.FindByID(instance.nodeID); ok {
			definition := definition.skyDefinitions[instance.definitionIndex]
			s.placeSky(node, definition)
		}
	}

	// NOTE: This needs to happen after armatures are initialized!
	modelNode.ApplyToTarget(true)

	return &Model{
		definition:    definition,
		root:          modelNode,
		bodyInstances: bodyInstances,
		armatures:     armatures,
		animations:    animations,
	}
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

func (s *Scene) placeSky(node *hierarchy.Node, definition *graphics.SkyDefinition) {
	sky := s.gfxScene.CreateSky(graphics.SkyInfo{
		Definition: definition,
	})
	node.SetTarget(SkyNodeTarget{
		Sky: sky,
	})
}
