package game

import (
	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/dtos"
	"github.com/mokiat/lacking/game/ecs"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/physics"
	"github.com/mokiat/lacking/log"
	"github.com/mokiat/lacking/util/datastruct"
)

func newScene(resourceSet *ResourceSet, physicsScene *physics.Scene, gfxScene *graphics.Scene, ecsScene *ecs.Scene) *Scene {
	return &Scene{
		physicsScene: physicsScene,
		gfxScene:     gfxScene,
		ecsScene:     ecsScene,
		root:         NewNode(),

		playbackPool: datastruct.NewDynamicPool[Playback](),
		playbacks:    ds.NewList[*Playback](4),
	}
}

type Scene struct {
	physicsScene *physics.Scene
	gfxScene     *graphics.Scene
	ecsScene     *ecs.Scene
	root         *Node
	models       []*Model

	playbackPool datastruct.Pool[Playback]
	playbacks    *ds.List[*Playback]

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
		s.Graphics().CreateAmbientLightNew(graphics.AmbientLightInfo{
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

func (s *Scene) Update(elapsedSeconds float64) {
	if !s.frozen {
		s.applyPlaybacks(elapsedSeconds)
		s.applyNodeToPhysics(s.root)
		s.physicsScene.Update(elapsedSeconds)
		s.applyPhysicsToNode(s.root)
	}
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
		body := s.physicsScene.CreateBody(physics.BodyInfo{
			Name:       instance.Name,
			Definition: bodyDefinition,
			Position:   dprec.ZeroVec3(),
			Rotation:   dprec.IdentityQuat(),
			IsDynamic:  info.IsDynamic,
		})
		bodyNode.SetBody(body)
		bodyInstances = append(bodyInstances, body)
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
			// TODO: Use single method SetArmatureBinding(armature, joint)
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
		mesh := s.gfxScene.CreateMesh(graphics.MeshInfo{
			Definition: meshDefinition,
			Armature:   armature,
		})
		meshNode.SetAttachable(mesh)
	}

	if info.IsDynamic {
		s.Root().AppendChild(modelNode)
	}
	s.applyPhysicsToNode(modelNode)
	s.applyNodeToPhysics(modelNode)
	s.applyNodeToGraphics(modelNode)

	result := &Model{
		definition:    definition,
		root:          modelNode,
		bodyInstances: bodyInstances,
		nodes:         nodes,
		armatures:     armatures,
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
	for _, playback := range s.playbacks.Items() {
		if playback.name == name {
			return playback
		}
	}
	return nil
}

func (s *Scene) applyPhysicsToNode(node *Node) {
	if body := node.body; body != nil {
		if !body.Static() {
			absMatrix := dprec.TRSMat4(
				body.VisualPosition(),
				body.VisualOrientation(),
				dprec.NewVec3(1.0, 1.0, 1.0),
			)
			node.SetAbsoluteMatrix(absMatrix)
		}
	}
	for child := node.firstChild; child != nil; child = child.rightSibling {
		s.applyPhysicsToNode(child)
	}
}

func (s *Scene) applyPlaybacks(elapsedSeconds float64) {
	for _, playback := range s.playbacks.Items() {
		if playback.playing {
			playback.Advance(elapsedSeconds)
			playback.animation.Apply(playback.head)
		}
	}
}

func (s *Scene) applyNodeToPhysics(node *Node) {
	if body := node.body; body != nil {
		if body.Static() {
			absMatrix := node.AbsoluteMatrix()
			translation, rotation, _ := absMatrix.TRS()
			body.SetPosition(translation)
			body.SetOrientation(rotation)
		}
	}
	for child := node.firstChild; child != nil; child = child.rightSibling {
		s.applyNodeToPhysics(child)
	}
}

func (s *Scene) applyNodeToGraphics(node *Node) {
	// NOTE: call AbsoluteMatrix regardless if there is anything attached or not,
	// since this Node could be used as an armature.
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
