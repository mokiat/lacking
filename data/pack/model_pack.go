package pack

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/newasset/mdl"
)

type ModelProvider interface {
	Model() *Model
}

type Properties map[string]string

func (p Properties) HasCollision() bool {
	return p.IsSet("collidable")
}

func (p Properties) HasSkipCollision() bool {
	return p.IsSet("non-collidable")
}

func (p Properties) IsInvisible() bool {
	return p.IsSet("invisible")
}

func (p Properties) IsSet(key string) bool {
	if p == nil {
		return false
	}
	_, ok := p[key]
	return ok
}

type Model struct {
	RootNodes        []*Node
	Animations       []*Animation
	Armatures        []*Armature
	Materials        []*Material
	MeshDefinitions  []*MeshDefinition
	MeshInstances    []*MeshInstance
	LightDefinitions []*LightDefinition
	LightInstances   []*LightInstance
	Textures         []*mdl.Image
	Properties       Properties
}

type Node struct {
	Name        string
	Translation dprec.Vec3
	Scale       dprec.Vec3
	Rotation    dprec.Quat
	Children    []*Node
	Properties  Properties
}

type MeshDefinition struct {
	Name         string
	VertexLayout VertexLayout
	Vertices     []Vertex
	Indices      []int
	Fragments    []MeshFragment
	Properties   Properties
}

func (d MeshDefinition) HasCollision() bool {
	return d.Properties.HasCollision()
}

type Armature struct {
	Joints []Joint
}

type Joint struct {
	Node              *Node
	InverseBindMatrix sprec.Mat4
}

type MeshInstance struct {
	Name       string
	Node       *Node
	Armature   *Armature
	Definition *MeshDefinition
}

func (i MeshInstance) HasCollision() bool {
	return i.Definition.HasCollision()
}

type VertexLayout struct {
	HasCoords    bool
	HasNormals   bool
	HasTangents  bool
	HasTexCoords bool
	HasColors    bool
	HasWeights   bool
	HasJoints    bool
}

type Vertex struct {
	Coord    sprec.Vec3
	Normal   sprec.Vec3
	Tangent  sprec.Vec3
	TexCoord sprec.Vec2
	Color    sprec.Vec4
	Weights  sprec.Vec4
	Joints   [4]uint8
}

type MeshFragment struct {
	Primitive   Primitive
	IndexOffset int
	IndexCount  int
	Material    *Material
}

type Primitive int

const (
	PrimitivePoints Primitive = iota
	PrimitiveLines
	PrimitiveLineStrip
	PrimitiveLineLoop
	PrimitiveTriangles
	PrimitiveTriangleStrip
	PrimitiveTriangleFan
)

type Material struct {
	Name                     string
	BackfaceCulling          bool
	AlphaTesting             bool
	AlphaThreshold           float32
	Blending                 bool
	Color                    sprec.Vec4
	ColorTexture             *TextureRef
	Metallic                 float32
	Roughness                float32
	MetallicRoughnessTexture *TextureRef
	NormalScale              float32
	NormalTexture            *TextureRef
	Properties               Properties
}

func (m Material) HasSkipCollision() bool {
	return m.Properties.HasSkipCollision()
}

func (m Material) IsInvisible() bool {
	return m.Properties.IsInvisible()
}

type TextureRef struct {
	TextureIndex int
}

type Animation struct {
	Name      string
	StartTime float64
	EndTime   float64
	Bindings  []*AnimationBinding
}

type AnimationBinding struct {
	Node                 *Node
	TranslationKeyframes []TranslationKeyframe
	RotationKeyframes    []RotationKeyframe
	ScaleKeyframes       []ScaleKeyframe
}

type TranslationKeyframe struct {
	Timestamp   float64
	Translation dprec.Vec3
}

type RotationKeyframe struct {
	Timestamp float64
	Rotation  dprec.Quat
}

type ScaleKeyframe struct {
	Timestamp float64
	Scale     dprec.Vec3
}

type LightDefinition struct {
	Name               string
	Type               LightType
	EmitRange          float64
	EmitOuterConeAngle dprec.Angle
	EmitInnerConeAngle dprec.Angle
	EmitColor          dprec.Vec3
}

type LightType string

const (
	LightTypePoint       LightType = "point"
	LightTypeSpot        LightType = "spot"
	LightTypeDirectional LightType = "directional"
	// TODO: Ambient Light
)

type LightInstance struct {
	Name       string
	Node       *Node
	Definition *LightDefinition
}
