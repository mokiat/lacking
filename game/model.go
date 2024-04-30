package game

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/game/hierarchy"
	"github.com/mokiat/lacking/game/physics"
	"github.com/mokiat/lacking/render"
)

type ModelDefinition struct {
	nodes             []nodeDefinition
	animations        []*AnimationDefinition
	armatures         []armatureDefinition
	shaders           []*graphics.Shader
	textures          []render.Texture
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

func (d *ModelDefinition) FindAnimation(name string) *AnimationDefinition {
	for _, def := range d.animations {
		if def.name == name {
			return def
		}
	}
	return nil
}

type nodeDefinition struct {
	ParentIndex int
	Name        string
	Position    dprec.Vec3
	Rotation    dprec.Quat
	Scale       dprec.Vec3
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
	NodeIndex         int
	InverseBindMatrix sprec.Mat4
}

type meshInstance struct {
	NodeIndex       int
	DefinitionIndex int
	ArmatureIndex   int
}

type bodyInstance struct {
	NodeIndex       int
	DefinitionIndex int
}

type ambientLightInstance struct {
	nodeIndex         int
	reflectionTexture render.Texture
	refractionTexture render.Texture
	castShadow        bool
}

type pointLightInstance struct {
	nodeIndex    int
	emitColor    dprec.Vec3
	emitDistance float64
	castShadow   bool
}

type spotLightInstance struct {
	nodeIndex      int
	emitColor      dprec.Vec3
	emitDistance   float64
	emitAngleOuter dprec.Angle
	emitAngleInner dprec.Angle
	castShadow     bool
}

type directionalLightInstance struct {
	nodeIndex  int
	emitColor  dprec.Vec3
	castShadow bool
}

type skyInstance struct {
	nodeIndex       int
	definitionIndex int
}

// ModelInfo contains the information necessary to place a Model
// instance into a Scene.
type ModelInfo struct {

	// Name specifies the name of this instance. This should not be
	// confused with the name of the definition.
	Name string

	// Definition specifies the template from which this instance will
	// be created.
	Definition *ModelDefinition

	// Position is used to specify a location for the model instance.
	Position dprec.Vec3

	// Rotation is used to specify a rotation for the model instance.
	Rotation dprec.Quat

	// Scale is used to specify a scale for the model instance.
	Scale dprec.Vec3

	// IsDynamic determines whether the model can be repositioned once
	// placed in the Scene.
	// (i.e. whether it should be added to the scene hierarchy)
	IsDynamic bool
}

type Model struct {
	definition *ModelDefinition
	root       *hierarchy.Node

	nodes         []*hierarchy.Node
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
