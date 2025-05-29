package game

import (
	"github.com/mokiat/gog/ds"
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/hierarchy"
	"github.com/mokiat/lacking/game/physics"
	"github.com/mokiat/lacking/render"
)

type ModelDefinition struct {
	hierarchy         *HierarchyTemplate
	animationSet      *AnimationSetTemplate
	armatures         []armatureDefinition
	shaders           []*graphics.Shader
	textures          map[uint32]render.Texture
	materials         []*graphics.Material
	meshGeometries    []*graphics.MeshGeometry
	meshDefinitions   []*graphics.MeshDefinition
	meshes            []meshInstance
	bodyMaterials     []*physics.Material
	bodyDefinitions   []*physics.BodyDefinition
	bodies            []bodyInstance
	ambientLights     []ambientLightInstance
	pointLights       []pointLightInstance
	spotLights        []spotLightInstance
	directionalLights []directionalLightInstance
	skyDefinitions    []*graphics.SkyDefinition
	skies             []skyInstance
}

type armatureDefinition struct {
	Joints []armatureJoint
}

func (d armatureDefinition) InverseBindMatrices() []sprec.Mat4 {
	result := make([]sprec.Mat4, len(d.Joints))
	for i, joint := range d.Joints {
		result[i] = joint.InverseBindMatrix
	}
	return result
}

type armatureJoint struct {
	NodeID            uint32
	InverseBindMatrix sprec.Mat4
}

type meshInstance struct {
	NodeID          uint32
	DefinitionIndex int
	ArmatureIndex   int
}

type bodyInstance struct {
	NodeID          uint32
	DefinitionIndex int
}

type ambientLightInstance struct {
	nodeID              uint32
	reflectionTextureID uint32
	refractionTextureID uint32
	castShadow          bool
}

type pointLightInstance struct {
	nodeID       uint32
	emitColor    dprec.Vec3
	emitDistance float64
	castShadow   bool
}

type spotLightInstance struct {
	nodeID         uint32
	emitColor      dprec.Vec3
	emitDistance   float64
	emitAngleOuter dprec.Angle
	emitAngleInner dprec.Angle
	castShadow     bool
}

type directionalLightInstance struct {
	nodeID     uint32
	emitColor  dprec.Vec3
	castShadow bool
}

type skyInstance struct {
	nodeID          uint32
	definitionIndex int
}

// ModelInfo contains the information necessary to place a Model
// instance into a Scene.
type ModelInfo struct {

	// Name specifies the name of this instance. This should not be
	// confused with the name of the definition.
	Name string

	// RootNode specifies the name of the root node of the model to use, in which
	// case a wrapper root node will not be created. The selected root node will
	// be renamed to Name if it is specified.
	RootNode opt.T[string]

	// Definition specifies the template from which this instance will
	// be created.
	Definition *ModelDefinition

	// Position is used to specify a location for the model instance.
	Position opt.T[dprec.Vec3]

	// Rotation is used to specify a rotation for the model instance.
	Rotation opt.T[dprec.Quat]

	// Scale is used to specify a scale for the model instance.
	Scale opt.T[dprec.Vec3]

	// IsDynamic determines whether the model can be repositioned once
	// placed in the Scene - whether it should be added to the scene hierarchy.
	IsDynamic bool
}

type Model struct {
	definition *ModelDefinition
	root       *hierarchy.Node

	armatures     []*graphics.Armature
	animations    []*Animation
	bodyInstances []physics.Body
}

func (m *Model) Root() *hierarchy.Node {
	return m.root
}

func (m *Model) FindNode(name string) *hierarchy.Node {
	return m.root.FindNode(name)
}

func (m *Model) BodyInstances() []physics.Body {
	return m.bodyInstances
}

func (m *Model) Animations() []*Animation {
	return m.animations
}

func (m *Model) FindAnimation(name string) *Animation {
	for _, animation := range m.animations {
		if animation.name == name {
			return animation
		}
	}
	return nil
}

func (m *Model) AnimatedNodes() []*hierarchy.Node {
	result := ds.NewSet[*hierarchy.Node](0)
	for _, animation := range m.animations {
		for nodeName := range animation.bindings {
			if node := m.FindNode(nodeName); node != nil {
				result.Add(node)
			}
		}
	}
	return result.Items()
}

func (m *Model) BindAnimationSource(source AnimationSource) {
	for _, node := range m.AnimatedNodes() {
		node.SetSource(AnimationNodeSource{
			Source: source,
		})
	}
}
