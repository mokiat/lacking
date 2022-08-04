package pack

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/sprec"
)

type ModelProvider interface {
	Model() *Model
}

type Model struct {
	RootNodes       []*Node
	Animations      []*Animation
	Armatures       []*Armature
	Materials       []*Material
	MeshDefinitions []*MeshDefinition
	MeshInstances   []*MeshInstance
}

type Node struct {
	Name        string
	Translation sprec.Vec3
	Scale       sprec.Vec3
	Rotation    sprec.Quat
	Children    []*Node
}

type MeshDefinition struct {
	Name         string
	VertexLayout VertexLayout
	Vertices     []Vertex
	Indices      []int
	Fragments    []MeshFragment
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
	ColorTexture             string
	Metallic                 float32
	Roughness                float32
	MetallicRoughnessTexture string
	NormalScale              float32
	NormalTexture            string
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
