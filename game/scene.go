package game

import (
	"time"

	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/dtos"
	"github.com/mokiat/lacking/game/ecs"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/physics"
	"github.com/mokiat/lacking/game/physics/collision"
	"github.com/mokiat/lacking/log"
)

func newScene(resourceSet *ResourceSet, physicsScene *physics.Scene, gfxScene *graphics.Scene, ecsScene *ecs.Scene) *Scene {
	return &Scene{
		physicsScene: physicsScene,
		gfxScene:     gfxScene,
		ecsScene:     ecsScene,
		root:         NewNode(),

		playbackPool: ds.NewPool[Playback](),
		playbacks:    ds.NewList[*Playback](4),

		preUpdateSubscriptions:  NewSubscriptionSet[UpdateCallback](),
		postUpdateSubscriptions: NewSubscriptionSet[UpdateCallback](),
	}
}

type Scene struct {
	physicsScene *physics.Scene
	gfxScene     *graphics.Scene
	ecsScene     *ecs.Scene
	root         *Node
	models       []*Model

	playbackPool *ds.Pool[Playback]
	playbacks    *ds.List[*Playback]

	preUpdateSubscriptions  *SubscriptionSet[UpdateCallback]
	postUpdateSubscriptions *SubscriptionSet[UpdateCallback]

	frozen bool
}

func (s *Scene) Delete() {
	defer s.physicsScene.Delete()
	defer s.gfxScene.Delete()
	defer s.ecsScene.Delete()
}

func (s *Scene) IsFrozen() bool {
	return s.frozen
}

func (s *Scene) Freeze() {
	s.frozen = true
}

func (s *Scene) Unfreeze() {
	s.frozen = false
}

func (s *Scene) Physics() *physics.Scene {
	return s.physicsScene
}

func (s *Scene) Graphics() *graphics.Scene {
	return s.gfxScene
}

func (s *Scene) ECS() *ecs.Scene {
	return s.ecsScene
}

func (s *Scene) Root() *Node {
	return s.root
}

func (s *Scene) Initialize(definition *SceneDefinition) {
	if definition.skyboxTexture != nil {
		s.Graphics().Sky().SetSkybox(definition.skyboxTexture.gfxTexture)
	}

	if definition.reflectionTexture != nil && definition.refractionTexture != nil {
		s.Graphics().CreateAmbientLight(graphics.AmbientLightInfo{
			ReflectionTexture: definition.reflectionTexture.gfxTexture,
			RefractionTexture: definition.refractionTexture.gfxTexture,
			Position:          dprec.ZeroVec3(),
			InnerRadius:       25000.0,
			OuterRadius:       25000.0, // FIXME
		})
	}

	s.CreateModel(ModelInfo{
		Name:              "scene",
		Definition:        definition.model,
		Position:          dprec.ZeroVec3(),
		Rotation:          dprec.IdentityQuat(),
		Scale:             dprec.NewVec3(1.0, 1.0, 1.0),
		IsDynamic:         false,
		PrepareAnimations: false,
	})

	for _, instance := range definition.modelInstances {
		model := s.CreateModel(instance)
		s.models = append(s.models, model)
	}
}

func (s *Scene) FindModel(name string) *Model {
	for _, model := range s.models {
		if model.root.name == name {
			return model
		}
	}
	return nil
}

func (s *Scene) SubscribePreUpdate(callback UpdateCallback) *UpdateSubscription {
	return s.preUpdateSubscriptions.Subscribe(callback)
}

func (s *Scene) SubscribePostUpdate(callback UpdateCallback) *UpdateSubscription {
	return s.postUpdateSubscriptions.Subscribe(callback)
}

func (s *Scene) Update(elapsedTime time.Duration) {
	if s.frozen {
		return
	}

	s.preUpdateSubscriptions.Each(func(callback UpdateCallback) {
		callback(elapsedTime)
	})

	// TODO: Pre-physics
	s.physicsScene.Update(elapsedTime.Seconds()) // FIXME: use ticker
	// TODO: Post-physics

	// TDOO: Pre-node
	s.applyPlaybacks(elapsedTime)
	s.applyPhysicsToNode(s.root)
	// TODO: Post-node

	s.postUpdateSubscriptions.Each(func(callback UpdateCallback) {
		callback(elapsedTime)
	})
}

func (s *Scene) Render(viewport graphics.Viewport) {
	s.applyNodeToGraphics(s.root)
	s.gfxScene.Render(viewport)
}

func (s *Scene) CreateAnimation(info AnimationInfo) *Animation {
	def := info.Definition
	bindings := make([]animationBinding, len(def.bindings))
	for i, bindingDef := range def.bindings {
		var target *Node
		if bindingDef.NodeIndex >= 0 {
			target = info.Model.nodes[bindingDef.NodeIndex]
		} else {
			target = info.Model.root.FindNode(bindingDef.NodeName)
		}
		if target == nil {
			log.Warn("Animation cannot find target node %q", bindingDef.NodeName)
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

func (s *Scene) CreateModel(info ModelInfo) *Model {
	modelNode := NewNode()
	modelNode.SetName(info.Name)
	modelNode.SetPosition(info.Position)
	modelNode.SetRotation(info.Rotation)
	modelNode.SetScale(info.Scale)

	definition := info.Definition
	nodes := make([]*Node, len(definition.nodes))
	for i, nodeDef := range definition.nodes {
		node := NewNode()
		node.SetName(nodeDef.Name)
		node.SetPosition(nodeDef.Position)
		node.SetRotation(nodeDef.Rotation)
		node.SetScale(nodeDef.Scale)
		nodes[i] = node
	}
	for i, nodeDef := range definition.nodes {
		var parent *Node
		if nodeDef.ParentIndex >= 0 {
			parent = nodes[nodeDef.ParentIndex]
		} else {
			parent = modelNode
		}
		parent.AppendChild(nodes[i])
	}

	var bodyInstances []*physics.Body
	for _, instance := range definition.bodyInstances {
		var bodyNode *Node
		if instance.NodeIndex >= 0 {
			bodyNode = nodes[instance.NodeIndex]
		} else {
			bodyNode = modelNode
		}
		bodyDefinition := definition.bodyDefinitions[instance.DefinitionIndex]
		if info.IsDynamic {
			body := s.physicsScene.CreateBody(physics.BodyInfo{
				Name:       instance.Name,
				Definition: bodyDefinition,
				Position:   dprec.ZeroVec3(),
				Rotation:   dprec.IdentityQuat(),
			})
			bodyNode.SetBody(body)
			bodyInstances = append(bodyInstances, body)
		} else {
			absMatrix := bodyNode.AbsoluteMatrix()
			transform := collision.TRTransform(absMatrix.Translation(), absMatrix.Rotation())
			collisionSet := collision.NewSet()
			collisionSet.Replace(bodyDefinition.CollisionSet(), transform)
			s.physicsScene.CreateProp(physics.PropInfo{
				CollisionSet: collisionSet,
			})
		}
	}

	pointLightInstances := make([]*graphics.PointLight, len(definition.pointLightInstances))
	for i, instance := range definition.pointLightInstances {
		var lightNode *Node
		if instance.NodeIndex >= 0 {
			lightNode = nodes[instance.NodeIndex]
		} else {
			lightNode = modelNode
		}
		light := s.gfxScene.CreatePointLight(graphics.PointLightInfo{
			Position:  dprec.ZeroVec3(),
			EmitRange: instance.EmitRange,
			EmitColor: instance.EmitColor,
		})
		lightNode.SetAttachable(light)
		pointLightInstances[i] = light
	}

	spotLightInstances := make([]*graphics.SpotLight, len(definition.spotLightInstances))
	for i, instance := range definition.spotLightInstances {
		var lightNode *Node
		if instance.NodeIndex >= 0 {
			lightNode = nodes[instance.NodeIndex]
		} else {
			lightNode = modelNode
		}
		light := s.gfxScene.CreateSpotLight(graphics.SpotLightInfo{
			Position:           dprec.ZeroVec3(),
			Rotation:           dprec.IdentityQuat(),
			EmitRange:          instance.EmitRange,
			EmitOuterConeAngle: instance.EmitOuterConeAngle,
			EmitInnerConeAngle: instance.EmitInnerConeAngle,
			EmitColor:          instance.EmitColor,
		})
		lightNode.SetAttachable(light)
		spotLightInstances[i] = light
	}

	directionalLightInstances := make([]*graphics.DirectionalLight, len(definition.directionalLightInstances))
	for i, instance := range definition.directionalLightInstances {
		var lightNode *Node
		if instance.NodeIndex >= 0 {
			lightNode = nodes[instance.NodeIndex]
		} else {
			lightNode = modelNode
		}
		light := s.gfxScene.CreateDirectionalLight(graphics.DirectionalLightInfo{
			Position:    dprec.ZeroVec3(),
			Orientation: dprec.IdentityQuat(),
			EmitRange:   instance.EmitRange,
			EmitColor:   instance.EmitColor,
		})
		lightNode.SetAttachable(light)
		directionalLightInstances[i] = light
	}

	armatures := make([]*graphics.Armature, len(definition.armatures))
	for i, instance := range definition.armatures {
		armature := s.gfxScene.CreateArmature(graphics.ArmatureInfo{
			InverseMatrices: instance.InverseBindMatrices(),
		})
		for j, joint := range instance.Joints {
			var jointNode *Node
			if joint.NodeIndex >= 0 {
				jointNode = nodes[joint.NodeIndex]
			} else {
				jointNode = modelNode
			}
			// TODO: Use single method SetAttachment(BoneAttachment{armature, joint})
			jointNode.SetArmature(armature)
			jointNode.SetArmatureBone(j)
		}
		armatures[i] = armature
	}

	for _, instance := range definition.meshInstances {
		var meshNode *Node
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

		if info.IsDynamic {
			mesh := s.gfxScene.CreateMesh(graphics.MeshInfo{
				Definition: meshDefinition,
				Armature:   armature,
			})
			meshNode.SetAttachable(mesh)
		} else {
			s.gfxScene.CreateStaticMesh(graphics.StaticMeshInfo{
				Definition: meshDefinition,
				Matrix:     meshNode.AbsoluteMatrix(),
			})
		}
	}

	if info.IsDynamic {
		s.Root().AppendChild(modelNode)
	}
	s.applyPhysicsToNode(modelNode)
	s.applyNodeToGraphics(modelNode)

	result := &Model{
		definition:                definition,
		root:                      modelNode,
		bodyInstances:             bodyInstances,
		nodes:                     nodes,
		armatures:                 armatures,
		pointLightInstances:       pointLightInstances,
		spotLightInstances:        spotLightInstances,
		directionalLightInstances: directionalLightInstances,
	}
	if info.PrepareAnimations {
		animations := make([]*Animation, len(definition.animations))
		for i, animationDef := range definition.animations {
			animations[i] = s.CreateAnimation(AnimationInfo{
				Model:      result,
				Definition: animationDef,
			})
		}
		result.animations = animations
	}
	return result
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

func (s *Scene) applyPhysicsToNode(node *Node) {
	if body := node.body; body != nil {
		absMatrix := dprec.TRSMat4(
			body.VisualPosition(),
			body.VisualOrientation(),
			dprec.NewVec3(1.0, 1.0, 1.0),
		)
		node.SetAbsoluteMatrix(absMatrix)
	}
	for child := node.firstChild; child != nil; child = child.rightSibling {
		s.applyPhysicsToNode(child)
	}
}

func (s *Scene) applyPlaybacks(elapsedTime time.Duration) {
	for _, playback := range s.playbacks.Unbox() {
		if playback.playing {
			playback.Advance(elapsedTime.Seconds())
			playback.animation.Apply(playback.head)
		}
	}
}

func (s *Scene) applyNodeToGraphics(node *Node) {
	absMatrix := node.AbsoluteMatrix()
	if armature := node.armature; armature != nil {
		armature.SetBone(node.armatureBone, dtos.Mat4(absMatrix))
	}
	if attachable := node.attachable; attachable != nil {
		attachable.SetMatrix(absMatrix)
	}
	for child := node.firstChild; child != nil; child = child.rightSibling {
		s.applyNodeToGraphics(child)
	}
}
