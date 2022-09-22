package game

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/dtos"
	"github.com/mokiat/lacking/game/ecs"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/physics"
	"github.com/mokiat/lacking/log"
	"github.com/mokiat/lacking/util/metrics"
)

func newScene(resourceSet *ResourceSet, physicsScene *physics.Scene, gfxScene *graphics.Scene, ecsScene *ecs.Scene) *Scene {
	return &Scene{
		physicsScene: physicsScene,
		gfxScene:     gfxScene,
		ecsScene:     ecsScene,
		root:         NewNode(),
	}
}

type Scene struct {
	physicsScene *physics.Scene
	gfxScene     *graphics.Scene
	ecsScene     *ecs.Scene
	root         *Node
}

func (s *Scene) Delete() {
	defer s.physicsScene.Delete()
	defer s.gfxScene.Delete()
	defer s.ecsScene.Delete()
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
		ambientLight := s.Graphics().CreateAmbientLight()
		ambientLight.SetReflectionTexture(definition.reflectionTexture.gfxTexture)
		ambientLight.SetRefractionTexture(definition.refractionTexture.gfxTexture)
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
		s.CreateModel(instance)
	}
}

func (s *Scene) Update(elapsedSeconds float64) {
	// TODO: Add OnUpdate hook here so that user code can modify stuff based off
	// of stable state.
	s.physicsScene.Update(elapsedSeconds)

	transform := metrics.BeginSpan("transform")
	s.applyPhysicsToNode(s.root)
	s.applyNodeToPhysics(s.root)
	s.applyNodeToGraphics(s.root)
	transform.End()
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
	// if info.IsDynamic {
	s.Root().AppendChild(modelNode)
	// }

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
		meshNode.SetMesh(mesh)
	}

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

func (s *Scene) applyPhysicsToNode(node *Node) {
	if body := node.body; body != nil {
		if !body.Static() {
			// FIXME: This should be SetAbsolutePosition and SetAbsoluteRotation
			node.SetPosition(body.Position())
			node.SetRotation(body.Orientation())
		}
	}
	for child := node.firstChild; child != nil; child = child.rightSibling {
		s.applyPhysicsToNode(child)
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
	if mesh := node.Mesh(); mesh != nil {
		mesh.SetMatrix(node.AbsoluteMatrix())
	}
	if camera := node.Camera(); camera != nil {
		camera.SetMatrix(node.AbsoluteMatrix())
	}
	if light := node.light; light != nil {
		light.SetMatrix(node.AbsoluteMatrix())
	}
	if armature := node.armature; armature != nil {
		armature.SetBone(node.armatureBone, dtos.Mat4(node.AbsoluteMatrix()))
	}
	for child := node.firstChild; child != nil; child = child.rightSibling {
		s.applyNodeToGraphics(child)
	}
}
