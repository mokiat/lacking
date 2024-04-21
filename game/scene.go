package game

import (
	"time"

	"github.com/mokiat/gog/ds"
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

func newScene(physicsScene *physics.Scene, gfxScene *graphics.Scene, ecsScene *ecs.Scene) *Scene {
	return &Scene{
		physicsScene: physicsScene,
		gfxScene:     gfxScene,
		ecsScene:     ecsScene,
		root:         hierarchy.NewNode(), // TODO: Make this node stationary

		playbackPool: ds.NewPool[Playback](),
		playbacks:    ds.NewList[*Playback](4),

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
	physicsScene *physics.Scene
	gfxScene     *graphics.Scene
	ecsScene     *ecs.Scene
	root         *hierarchy.Node

	playbackPool *ds.Pool[Playback]
	playbacks    *ds.List[*Playback]

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
	s.placeAmbientLight(placementData{
		Nodes: []*hierarchy.Node{node},
	}, ambientLightInstance{
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
	s.placePointLight(placementData{
		Nodes: []*hierarchy.Node{node},
	}, pointLightInstance{
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
	s.placeSpotLight(placementData{
		Nodes: []*hierarchy.Node{node},
	}, spotLightInstance{
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
	s.placeDirectionalLight(placementData{
		Nodes: []*hierarchy.Node{node},
	}, directionalLightInstance{
		nodeIndex:  0,
		emitColor:  info.EmitColor.ValueOrDefault(dprec.NewVec3(10.0, 0.0, 10.0)),
		castShadow: info.CastShadow.ValueOrDefault(false),
	})
	return node
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
}

// Render draws the scene to the provided viewport.
func (s *Scene) Render(viewport graphics.Viewport) {
	stageSpan := metric.BeginRegion("stage")
	s.root.ApplyToTarget(true)
	stageSpan.End()

	renderSpan := metric.BeginRegion("render")
	s.gfxScene.Render(viewport)
	renderSpan.End()
}

func (s *Scene) CreateAnimation(info AnimationInfo) *Animation {
	def := info.Definition
	bindings := make([]animationBinding, len(def.bindings))
	for i, bindingDef := range def.bindings {
		target := info.Root.FindNode(bindingDef.NodeName)
		if target == nil {
			logger.Warn("Animation cannot find target node (%q)!", bindingDef.NodeName)
		}
		bindings[i] = animationBinding{
			node:                 target,
			translationKeyframes: bindingDef.TranslationKeyframes,
			rotationKeyframes:    bindingDef.RotationKeyframes,
			scaleKeyframes:       bindingDef.ScaleKeyframes,
		}
	}
	return &Animation{
		name:       def.name,
		definition: def,
		bindings:   bindings,
	}
}

// TODO: Return the node instead and have the Model be a target?
func (s *Scene) CreateModel(info ModelInfo) *Model {
	modelNode := hierarchy.NewNode()
	modelNode.SetName(info.Name)
	modelNode.SetPosition(info.Position)
	modelNode.SetRotation(info.Rotation)
	modelNode.SetScale(info.Scale)

	definition := info.Definition
	nodes := make([]*hierarchy.Node, len(definition.nodes))
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

	animations := make([]*Animation, len(definition.animations))
	for i, animationDef := range definition.animations {
		animations[i] = s.CreateAnimation(AnimationInfo{
			Root:       modelNode,
			Definition: animationDef,
		})
	}

	armatures := make([]*graphics.Armature, len(definition.armatures))
	for i, instance := range definition.armatures {
		armature := s.gfxScene.CreateArmature(graphics.ArmatureInfo{
			InverseMatrices: instance.InverseBindMatrices(),
		})
		for j, joint := range instance.Joints {
			var jointNode *hierarchy.Node
			if joint.NodeIndex >= 0 {
				jointNode = nodes[joint.NodeIndex]
			} else {
				jointNode = modelNode
			}
			jointNode.SetTarget(BoneNodeTarget{
				Armature:  armature,
				BoneIndex: j,
			})
		}
		armatures[i] = armature
	}

	// TODO: Track mesh instances?
	for _, instance := range definition.meshes {
		var meshNode *hierarchy.Node
		if instance.NodeIndex >= 0 {
			meshNode = nodes[instance.NodeIndex]
		} else {
			meshNode = modelNode
		}
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
				Matrix:     meshNode.AbsoluteMatrix(),
			})
		}
	}

	var bodyInstances []physics.Body
	for _, instance := range definition.bodies {
		var bodyNode *hierarchy.Node
		if instance.NodeIndex >= 0 {
			bodyNode = nodes[instance.NodeIndex]
		} else {
			bodyNode = modelNode
		}
		bodyDefinition := definition.bodyDefinitions[instance.DefinitionIndex]
		if info.IsDynamic {
			body := s.physicsScene.CreateBody(physics.BodyInfo{
				Name:       bodyNode.Name(),
				Definition: bodyDefinition,
				Position:   dprec.ZeroVec3(),
				Rotation:   dprec.IdentityQuat(),
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

	for _, instance := range definition.ambientLights {
		s.placeAmbientLight(placementData{
			Nodes:    nodes,
			Textures: definition.textures,
		}, instance)
	}
	for _, instance := range definition.pointLights {
		s.placePointLight(placementData{
			Nodes: nodes,
		}, instance)
	}
	for _, instance := range definition.spotLights {
		s.placeSpotLight(placementData{
			Nodes: nodes,
		}, instance)
	}
	for _, instance := range definition.directionalLights {
		s.placeDirectionalLight(placementData{
			Nodes: nodes,
		}, instance)
	}
	for _, instance := range definition.skies {
		s.placeSky(placementData{
			Nodes:          nodes,
			SkyDefinitions: definition.skyDefinitions,
		}, instance)
	}

	if info.IsDynamic {
		s.Root().AppendChild(modelNode)
	}
	modelNode.ApplyFromSource(true)
	modelNode.ApplyToTarget(true)

	return &Model{
		definition:    definition,
		root:          modelNode,
		bodyInstances: bodyInstances,
		nodes:         nodes,
		armatures:     armatures,
		animations:    animations,
	}
}

func (s *Scene) PlayAnimation(animation *Animation) *Playback {
	result := s.playbackPool.Fetch()
	result.scene = s
	result.animation = animation
	result.head = animation.StartTime()
	result.startTime = animation.StartTime()
	result.endTime = animation.EndTime()
	result.speed = 1.0
	result.Play()
	s.playbacks.Add(result)
	return result
}

func (s *Scene) FindPlayback(name string) *Playback {
	for _, playback := range s.playbacks.Unbox() {
		if playback.name == name {
			return playback
		}
	}
	return nil
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
	s.applyPlaybacks(elapsedTime)
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

func (s *Scene) applyPlaybacks(elapsedTime time.Duration) {
	for _, playback := range s.playbacks.Unbox() {
		if playback.playing {
			playback.Advance(elapsedTime.Seconds())
			playback.animation.Apply(playback.head)
		}
	}
}

func (s *Scene) placeAmbientLight(data placementData, instance ambientLightInstance) {
	node := data.Nodes[instance.nodeIndex]
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

func (s *Scene) placePointLight(data placementData, instance pointLightInstance) {
	node := data.Nodes[instance.nodeIndex]
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

func (s *Scene) placeSpotLight(data placementData, instance spotLightInstance) {
	node := data.Nodes[instance.nodeIndex]
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

func (s *Scene) placeDirectionalLight(data placementData, instance directionalLightInstance) {
	node := data.Nodes[instance.nodeIndex]
	light := s.gfxScene.CreateDirectionalLight(graphics.DirectionalLightInfo{
		Position:   dprec.ZeroVec3(),
		Rotation:   dprec.IdentityQuat(),
		EmitColor:  instance.emitColor,
		EmitRange:  25000.0,
		CastShadow: instance.castShadow,
	})
	node.SetTarget(DirectionalLightNodeTarget{
		Light: light,
	})
	node.ApplyToTarget(false)
}

func (s *Scene) placeSky(data placementData, instance skyInstance) {
	node := data.Nodes[instance.nodeIndex]
	sky := s.gfxScene.CreateSky(graphics.SkyInfo{
		Definition: data.SkyDefinitions[instance.definitionIndex],
	})
	node.SetTarget(SkyNodeTarget{
		Sky: sky,
	})
}

type placementData struct {
	SkyDefinitions []*graphics.SkyDefinition

	Nodes    []*hierarchy.Node
	Textures []render.Texture
}
