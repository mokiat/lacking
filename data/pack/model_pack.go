package pack

import "github.com/mokiat/gomath/sprec"

type ModelProvider interface {
	Model() *Model
}

type Model struct {
	RootNodes []*Node
	Meshes    []*Mesh
}

type Node struct {
	Name        string
	Translation sprec.Vec3
	Scale       sprec.Vec3
	Rotation    sprec.Quat
	Mesh        *Mesh
	Children    []*Node
}

func (n *Node) Matrix() sprec.Mat4 {
	return sprec.Mat4MultiProd(
		sprec.TranslationMat4(n.Translation.X, n.Translation.Y, n.Translation.Z),
		sprec.TransformationMat4(
			n.Rotation.OrientationX(),
			n.Rotation.OrientationY(),
			n.Rotation.OrientationZ(),
			sprec.ZeroVec3(),
		),
		sprec.ScaleMat4(n.Scale.X, n.Scale.Y, n.Scale.Z),
	)
}

type Mesh struct {
	Name        string
	VertexCount int
	Coords      []sprec.Vec3
	Normals     []sprec.Vec3
	Tangents    []sprec.Vec3
	TexCoords   []sprec.Vec2
	Colors      []sprec.Vec4
	IndexCount  int
	Indices     []int
	SubMeshes   []SubMesh
}

type SubMesh struct {
	Primitive   Primitive
	IndexOffset int
	IndexCount  int
	Material    Material
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
	Type             string
	BackfaceCulling  bool
	AlphaTesting     bool
	AlphaThreshold   float32
	Metalness        float32
	MetalnessTexture string
	Roughness        float32
	RoughnessTexture string
	Color            sprec.Vec4
	ColorTexture     string
	NormalScale      float32
	NormalTexture    string
}
