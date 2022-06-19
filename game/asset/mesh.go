package asset

const UnspecifiedOffset = int32(-1)

type MeshDefinition struct {
	Name         string
	VertexData   []byte
	VertexLayout VertexLayout
	IndexData    []byte
	IndexLayout  IndexLayout
	Fragments    []MeshFragment
}

type MeshInstance struct {
	Name            string
	NodeIndex       int32
	DefinitionIndex int32
}

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
}

const (
	IndexLayoutUint16 IndexLayout = iota
	IndexLayoutUint32
)

type IndexLayout uint8

type MeshFragment struct {
	Topology      MeshTopology
	IndexOffset   int32
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
