package graphics

// MeshTemplate represents the definition of a mesh.
// Multiple mesh instances can be created off of one template
// reusing resources.
type MeshTemplate interface {

	// Delete releases any resources allocated by this
	// template.
	Delete()
}

// MeshTemplateDefinition contains everything needed to create
// a new MeshTemplate.
type MeshTemplateDefinition struct {
	VertexData   []byte
	VertexFormat VertexFormat
	IndexData    []byte
	IndexFormat  IndexFormat
	SubMeshes    []SubMeshTemplateDefinition
}

// SubMeshTemplateDefinition represents a portion of a mesh that
// is drawn with a specific material.
type SubMeshTemplateDefinition struct {
	Primitive   Primitive
	IndexOffset int
	IndexCount  int
	Material    Material
}

const (
	PrimitivePoints Primitive = 1 + iota
	PrimitiveLines
	PrimitiveLineStrip
	PrimitiveLineLoop
	PrimitiveTriangles
	PrimitiveTriangleStrip
	PrimitiveTriangleFan
)

type Primitive int

type VertexFormat struct {
	HasCoord            bool
	CoordOffsetBytes    int
	CoordStrideBytes    int
	HasNormal           bool
	NormalOffsetBytes   int
	NormalStrideBytes   int
	HasTangent          bool
	TangentOffsetBytes  int
	TangentStrideBytes  int
	HasTexCoord         bool
	TexCoordOffsetBytes int
	TexCoordStrideBytes int
	HasColor            bool
	ColorOffsetBytes    int
	ColorStrideBytes    int
}

const (
	IndexFormatU16 IndexFormat = 1 + iota
	IndexFormatU32
)

type IndexFormat int

// Mesh represents an instance of a 3D mesh.
type Mesh interface {
	Node

	// Delete removes this mesh from the scene.
	Delete()
}
