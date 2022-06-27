package pack

import "github.com/mokiat/gomath/sprec"

type ModelProvider interface {
	Model() *Model
}

type Model struct {
	RootNodes       []*Node
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

type MeshInstance struct {
	Name       string
	Node       *Node
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
