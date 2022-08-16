package asset

const (
	UnspecifiedOffset        = int32(-1)
	UnspecifiedMaterialIndex = int32(-1)
	UnspecifiedArmatureIndex = int32(-1)
)

type MeshDefinition struct {
	Name                 string
	VertexLayout         VertexLayout
	VertexData           []byte
	IndexLayout          IndexLayout
	IndexData            []byte
	Fragments            []MeshFragment
	BoundingSphereRadius float64
}

type MeshInstance struct {
	Name            string
	NodeIndex       int32
	ArmatureIndex   int32
	DefinitionIndex int32
}

type Armature struct {
	Joints []Joint
}

type Joint struct {
	NodeIndex         int32
	InverseBindMatrix [16]float32
}

// VertexLayout describes how vertex data is positioned within the VertexData
// buffer.
//
// Coords are represented by RGB32F (i.e. three 32-bit float values).
// Since coords can span a bigger range, 32-bit floats are preferred.
//
// Normals and Tangents are represented by RGB16F (i.e. three 16-bit float
// values). This should be more than sufficient for normals.
//
// TexCoords are represented by RG16F (i.e. two 16-bit float values).
// Since texture coordinates are usually close to the zero-one range,
// 16-bit floats should provide sufficient precision.
//
// Colors are represented by RGBA8UN (i.e. four 8-bit unsigned normalized
// values). This is sufficient for an sRGB color with alpha.
//
// Weights are represented by RGBA8UN (i.e. four 8-bit unsigned normalized
// values). This should provide sufficient precision while still being
// fairly compact.
//
// Joints are represented by RGBA8IU (i.e. four 8-bit integer unsigned values).
// This means that there can be at most 256 joints in an Armature.
type VertexLayout struct {
	CoordOffset    int32
	CoordStride    int32
	NormalOffset   int32
	NormalStride   int32
	TangentOffset  int32
	TangentStride  int32
	TexCoordOffset int32
	TexCoordStride int32
	ColorOffset    int32
	ColorStride    int32
	WeightsOffset  int32
	WeightsStride  int32
	JointsOffset   int32
	JointsStride   int32
}

const (
	IndexLayoutUint16 IndexLayout = iota
	IndexLayoutUint32
)

type IndexLayout uint8

type MeshFragment struct {
	Topology      MeshTopology
	IndexOffset   uint32
	IndexCount    uint32
	MaterialIndex int32
}

const (
	MeshTopologyPoints MeshTopology = iota
	MeshTopologyLineStrip
	MeshTopologyLineLoop
	MeshTopologyLines
	MeshTopologyTriangleStrip
	MeshTopologyTriangleFan
	MeshTopologyTriangles
)

type MeshTopology uint8
