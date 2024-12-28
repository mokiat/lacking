package game

import (
	"time"

	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/debug/log"
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
	s.placeAmbientLight(node, ambientLightInstance{
		nodeIndex:         0,
		reflectionTexture: info.ReflectionTexture,
		refractionTexture: info.RefractionTexture,
		castShadow:        info.CastShadow.ValueOrDefault(false),
	})
	return node
}

// CreatePointLight creates a new point light and appends it to the root of the
// scene.
func (s *Scene) CreatePointLight(info PointLightInfo) *hierarchy.Node {
	node := s.CreateNode()
	s.placePointLight(node, pointLightInstance{
		nodeIndex:    0,
		emitColor:    info.EmitColor.ValueOrDefault(dprec.NewVec3(10.0, 0.0, 10.0)),
		emitDistance: info.EmitDistance.ValueOrDefault(20.0),
		castShadow:   info.CastShadow.ValueOrDefault(false),
	})
	return node
}

// CreateSpotLight creates a new spot light and appends it to the root of the
// scene.
func (s *Scene) CreateSpotLight(info SpotLightInfo) *hierarchy.Node {
	node := s.CreateNode()
	s.placeSpotLight(node, spotLightInstance{
		nodeIndex:      0,
		emitColor:      info.EmitColor.ValueOrDefault(dprec.NewVec3(10.0, 0.0, 10.0)),
		emitDistance:   info.EmitDistance.ValueOrDefault(20.0),
		emitAngleOuter: info.EmitOuterConeAngle.ValueOrDefault(dprec.Degrees(60)),
		emitAngleInner: info.EmitInnerConeAngle.ValueOrDefault(dprec.Degrees(30)),
		castShadow:     info.CastShadow.ValueOrDefault(false),
	})
	return node
}

// CreateDirectionalLight creates a new directional light and appends it to the
// root of the scene.
func (s *Scene) CreateDirectionalLight(info DirectionalLightInfo) *hierarchy.Node {
	node := s.CreateNode()
	s.placeDirectionalLight(node, directionalLightInstance{
		nodeIndex:  0,
		emitColor:  info.EmitColor.ValueOrDefault(dprec.NewVec3(10.0, 0.0, 10.0)),
		castShadow: info.CastShadow.ValueOrDefault(false),
	})
	return node
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
	modelNode := hierarchy.NewNode()

	definition := info.Definition
	nodes := make(map[int]*hierarchy.Node, len(definition.nodes))
	for i, nodeDef := range definition.nodes {
		node := hierarchy.NewNode()
		node.SetName(nodeDef.Name)
		node.SetPosition(nodeDef.Position)
		node.SetRotation(nodeDef.Rotation)
		node.SetScale(nodeDef.Scale)
		nodes[i] = node
	}
	for i, nodeDef := range definition.nodes {
		var parent *hierarchy.Node
		if nodeDef.ParentIndex >= 0 {
			parent = nodes[nodeDef.ParentIndex]
		} else {
			parent = modelNode
		}
		parent.AppendChild(nodes[i])
	}
	if info.RootNode.Specified {
		if name := info.RootNode.Value; name != "" {
			modelNode = modelNode.FindNode(info.RootNode.Value)
		} else {
			modelNode = nil
		}
		if modelNode == nil {
			log.Error("Root node %q not found", info.RootNode.Value)
			modelNode = hierarchy.NewNode()
		}
		modelNode.Detach()
		for i := range definition.nodes {
			if node := nodes[i]; !node.IsDescendantOf(modelNode) {
				delete(nodes, i)
			}
		}
	}

	modelNode.SetName(info.Name)
	modelNode.SetPosition(info.Position.ValueOrDefault(dprec.ZeroVec3()))
	modelNode.SetRotation(info.Rotation.ValueOrDefault(dprec.IdentityQuat()))
	modelNode.SetScale(info.Scale.ValueOrDefault(dprec.NewVec3(1.0, 1.0, 1.0)))
	if info.IsDynamic {
		s.Root().AppendChild(modelNode)
	}

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
			if jointNode := nodes[joint.NodeIndex]; jointNode != nil {
				jointNode.SetTarget(BoneNodeTarget{
					Armature:  armature,
					BoneIndex: j,
				})
			}
		}
		armatures[i] = armature
	}

	// NOTE: This needs to happen after armatures are initialized!
	modelNode.ApplyToTarget(true)

	// TODO: Track mesh instances?
	for _, instance := range definition.meshes {
		if meshNode := nodes[instance.NodeIndex]; meshNode != nil {
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
		if bodyNode := nodes[instance.NodeIndex]; bodyNode != nil {
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
		if node := nodes[instance.nodeIndex]; node != nil {
			s.placeAmbientLight(node, instance)
		}
	}
	for _, instance := range definition.pointLights {
		if node := nodes[instance.nodeIndex]; node != nil {
			s.placePointLight(node, instance)
		}
	}
	for _, instance := range definition.spotLights {
		if node := nodes[instance.nodeIndex]; node != nil {
			s.placeSpotLight(node, instance)
		}
	}
	for _, instance := range definition.directionalLights {
		if node := nodes[instance.nodeIndex]; node != nil {
			s.placeDirectionalLight(node, instance)
		}
	}
	for _, instance := range definition.skies {
		if node := nodes[instance.nodeIndex]; node != nil {
			definition := definition.skyDefinitions[instance.definitionIndex]
			s.placeSky(node, definition)
		}
	}

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

func (s *Scene) placeAmbientLight(node *hierarchy.Node, instance ambientLightInstance) {
	light := s.gfxScene.CreateAmbientLight(graphics.AmbientLightInfo{
		Position:          dprec.ZeroVec3(),
		InnerRadius:       25000.0,
		OuterRadius:       25000.0,
		ReflectionTexture: instance.reflectionTexture,
		RefractionTexture: instance.refractionTexture,
		CastShadow:        instance.castShadow,
	})
	node.SetTarget(AmbientLightNodeTarget{
		Light: light,
	})
	node.ApplyToTarget(false)
}

func (s *Scene) placePointLight(node *hierarchy.Node, instance pointLightInstance) {
	light := s.gfxScene.CreatePointLight(graphics.PointLightInfo{
		Position:   dprec.ZeroVec3(),
		EmitColor:  instance.emitColor,
		EmitRange:  instance.emitDistance,
		CastShadow: instance.castShadow,
	})
	node.SetTarget(PointLightNodeTarget{
		Light: light,
	})
	node.ApplyToTarget(false)
}

func (s *Scene) placeSpotLight(node *hierarchy.Node, instance spotLightInstance) {
	light := s.gfxScene.CreateSpotLight(graphics.SpotLightInfo{
		Position:           dprec.ZeroVec3(),
		Rotation:           dprec.IdentityQuat(),
		EmitColor:          instance.emitColor,
		EmitRange:          instance.emitDistance,
		EmitOuterConeAngle: instance.emitAngleOuter,
		EmitInnerConeAngle: instance.emitAngleInner,
		CastShadow:         instance.castShadow,
	})
	node.SetTarget(SpotLightNodeTarget{
		Light: light,
	})
	node.ApplyToTarget(false)
}

func (s *Scene) placeDirectionalLight(node *hierarchy.Node, instance directionalLightInstance) {
	light := s.gfxScene.CreateDirectionalLight(graphics.DirectionalLightInfo{
		Position:   dprec.ZeroVec3(),
		Rotation:   dprec.IdentityQuat(),
		EmitColor:  instance.emitColor,
		CastShadow: instance.castShadow,
	})
	node.SetTarget(DirectionalLightNodeTarget{
		Light: light,
	})
	node.ApplyToTarget(false)
}

func (s *Scene) placeSky(node *hierarchy.Node, definition *graphics.SkyDefinition) {
	sky := s.gfxScene.CreateSky(graphics.SkyInfo{
		Definition: definition,
	})
	node.SetTarget(SkyNodeTarget{
		Sky: sky,
	})
}
