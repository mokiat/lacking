package game

import (
	"github.com/mokiat/gomath/dtos"
	"github.com/mokiat/lacking/game/ecs"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/physics"
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
	if info.IsDynamic {
		s.Root().AppendChild(modelNode)
	}

	for _, instance := range definition.bodyInstances {
		var bodyNode *Node
		if instance.NodeIndex >= 0 {
			bodyNode = nodes[instance.NodeIndex]
		} else {
			bodyNode = modelNode
		}
		bodyDefinition := definition.bodyDefinitions[instance.DefinitionIndex]
		body := s.physicsScene.CreateBody2(physics.BodyInfo{
			Name:       instance.Name,
			Definition: bodyDefinition,
			Position:   instance.Position,
			Rotation:   instance.Rotation,
			IsDynamic:  instance.IsDynamic,
		})
		bodyNode.SetBody(body)
	}

	// armatures := make([]*graphics.Armature, len(def.Armatures))
	// defToArmature := make(map[*ArmatureDefinition]*graphics.Armature)
	// for i, armatureDef := range def.Armatures {
	// 	// TODO: ECS Component that maps between armature and nodes and that
	// 	// updates the respective armature?
	// 	armature := s.gfxScene.CreateArmature(graphics.ArmatureTemplate{
	// 		// FIXME
	// 	})
	// 	armatures[i] = armature
	// 	defToArmature[armatureDef] = armature
	// }

	// materials := make([]*graphics.Material, len(def.Materials))

	// meshInstances := make([]*graphics.Mesh, len(def.MeshInstances))
	// for i, meshInstanceDef := range def.MeshInstances {
	// 	meshInstance := s.gfxScene.CreateMesh(meshInstanceDef.GraphicsTemplate)
	// 	if meshInstanceDef.Node != nil {
	// 		node := defToNode[meshInstanceDef.Node]
	// 		meshInstance.SetMatrix(node.AbsoluteMatrix()) // TODO: Do only if entity is static
	// 	}
	// 	if meshInstanceDef.Armature != nil {
	// 		armature := defToArmature[meshInstanceDef.Armature]
	// 		meshInstance.SetArmature(armature)
	// 	}
	// 	meshInstances[i] = meshInstance
	// }

	// // TODO:

	return &Model{
		// nodes:     nodes,
		// armatures: armatures,
		// materials: materials,
	}
}

// func (s *Scene) CreateModel(def *ModelDefinition) *Model {
// 	nodes := make([]*Node, len(def.Nodes))
// 	defToNode := make(map[*NodeDefinition]*Node)
// 	for i, nodeDef := range def.Nodes {
// 		node := NewNode()
// 		nodes[i] = node
// 		defToNode[nodeDef] = node
// 	}
// 	for _, nodeDef := range def.Nodes {
// 		node := defToNode[nodeDef]
// 		if nodeDef.Parent != nil {
// 			parentNode := defToNode[nodeDef.Parent]
// 			parentNode.AppendChild(node)
// 		}
// 	}

// 	armatures := make([]*graphics.Armature, len(def.Armatures))
// 	defToArmature := make(map[*ArmatureDefinition]*graphics.Armature)
// 	for i, armatureDef := range def.Armatures {
// 		// TODO: ECS Component that maps between armature and nodes and that
// 		// updates the respective armature?
// 		armature := s.gfxScene.CreateArmature(graphics.ArmatureTemplate{
// 			// FIXME
// 		})
// 		armatures[i] = armature
// 		defToArmature[armatureDef] = armature
// 	}

// 	materials := make([]*graphics.Material, len(def.Materials))

// 	meshInstances := make([]*graphics.Mesh, len(def.MeshInstances))
// 	for i, meshInstanceDef := range def.MeshInstances {
// 		meshInstance := s.gfxScene.CreateMesh(meshInstanceDef.GraphicsTemplate)
// 		if meshInstanceDef.Node != nil {
// 			node := defToNode[meshInstanceDef.Node]
// 			meshInstance.SetMatrix(node.AbsoluteMatrix()) // TODO: Do only if entity is static
// 		}
// 		if meshInstanceDef.Armature != nil {
// 			armature := defToArmature[meshInstanceDef.Armature]
// 			meshInstance.SetArmature(armature)
// 		}
// 		meshInstances[i] = meshInstance
// 	}

// 	// TODO:

// 	return &Model{
// 		nodes:     nodes,
// 		armatures: armatures,
// 		materials: materials,
// 	}
// }

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
