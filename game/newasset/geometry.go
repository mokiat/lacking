package asset

const (
	UnspecifiedBufferIndex = int32(-1)
)

const (
	TopologyPoints Topology = iota
	TopologyLineList
	TopologyLineStrip
	TopologyTriangleList
	TopologyTriangleStrip
)

// Topology represents the way that the vertices of a mesh are connected.
type Topology uint8

const (
	IndexLayoutUint16 IndexLayout = iota
	IndexLayoutUint32
)

// IndexLayout represents the way that the indices of a mesh are stored.
type IndexLayout uint8

const (
	VertexAttributeFormatRGBA32F VertexAttributeFormat = iota
	VertexAttributeFormatRGB32F
	VertexAttributeFormatRG32F
	VertexAttributeFormatR32F

	VertexAttributeFormatRGBA16F
	VertexAttributeFormatRGB16F
	VertexAttributeFormatRG16F
	VertexAttributeFormatR16F

	VertexAttributeFormatRGBA16S
	VertexAttributeFormatRGB16S
	VertexAttributeFormatRG16S
	VertexAttributeFormatR16S

	VertexAttributeFormatRGBA16SN
	VertexAttributeFormatRGB16SN
	VertexAttributeFormatRG16SN
	VertexAttributeFormatR16SN

	VertexAttributeFormatRGBA16U
	VertexAttributeFormatRGB16U
	VertexAttributeFormatRG16U
	VertexAttributeFormatR16U

	VertexAttributeFormatRGBA16UN
	VertexAttributeFormatRGB16UN
	VertexAttributeFormatRG16UN
	VertexAttributeFormatR16UN

	VertexAttributeFormatRGBA8S
	VertexAttributeFormatRGB8S
	VertexAttributeFormatRG8S
	VertexAttributeFormatR8S

	VertexAttributeFormatRGBA8SN
	VertexAttributeFormatRGB8SN
	VertexAttributeFormatRG8SN
	VertexAttributeFormatR8SN

	VertexAttributeFormatRGBA8U
	VertexAttributeFormatRGB8U
	VertexAttributeFormatRG8U
	VertexAttributeFormatR8U

	VertexAttributeFormatRGBA8UN
	VertexAttributeFormatRGB8UN
	VertexAttributeFormatRG8UN
	VertexAttributeFormatR8UN

	VertexAttributeFormatRGBA8IU
	VertexAttributeFormatRGB8IU
	VertexAttributeFormatRG8IU
	VertexAttributeFormatR8IU
)

// VertexAttributeFormat represents the format of a vertex attribute.
type VertexAttributeFormat uint8

// IndexBuffer represents a buffer of index data.
type IndexBuffer struct {

	// IndexLayout specifies the data type that is used to represent individual
	// indices.
	IndexLayout IndexLayout

	// Data is the raw byte data that represents the indices.
	Data []byte
}

// VertexBuffer represents a buffer of vertex data.
type VertexBuffer struct {

	// Stride is the number of bytes that each vertex occupies within this buffer.
	Stride uint32

	// Data is the raw byte data that represents the vertices.
	Data []byte
}

// VertexAttribute represents a single attribute of a vertex.
type VertexAttribute struct {

	// BufferIndex specifies the index of the VertexBuffer that contains the
	// attribute data.
	//
	// If this value is set to UnspecifiedBufferIndex, it means that the
	// attribute is not present.
	BufferIndex int32

	// ByteOffset specifies the byte offset within the VertexBuffer where the
	// attribute data starts.
	ByteOffset uint32

	// Format specifies the format of the attribute data.
	Format VertexAttributeFormat
}

// VertexLayout describes how vertex data is positioned within the VertexData
// buffers.
type VertexLayout struct {

	// Coord specifies the layout of the vertex coordinate attribute.
	Coord VertexAttribute

	// Normal specifies the layout of the vertex normal attribute.
	Normal VertexAttribute

	// Tangent specifies the layout of the vertex tangent attribute.
	Tangent VertexAttribute

	// TexCoord specifies the layout of the vertex texture coordinate attribute.
	TexCoord VertexAttribute

	// Color specifies the layout of the vertex color attribute.
	Color VertexAttribute

	// Weights specifies the layout of the vertex weights attribute.
	Weights VertexAttribute

	// Joints specifies the layout of the vertex joints attribute.
	Joints VertexAttribute
}

// Fragment represents a portion of a mesh that is drawn with a specific
// material and topology.
type Fragment struct {

	// Name is the name of the fragment, often a hint to the material.
	Name string

	// Topology specifies the way that the vertices of the fragment are
	// connected.
	Topology Topology

	// IndexByteOffset specifies the byte offset within the IndexBuffer
	// where the indices of the fragment start.
	IndexByteOffset uint32

	// IndexCount specifies the number of indices that are used to draw the
	// fragment.
	IndexCount uint32
}

// Geometry represents a collection of vertex and index data that can be used
// to render a mesh.
type Geometry struct {

	// Name is the name of the geometry.
	Name string

	// VertexBuffers is the collection of buffers that contain the vertex data.
	VertexBuffers []VertexBuffer

	// VertexLayout describes how the vertex data is positioned within the
	// VertexBuffers.
	VertexLayout VertexLayout

	// IndexBuffer is the buffer that contains the index data.
	IndexBuffer IndexBuffer

	// Fragments is the collection of fragments that make up the geometry.
	Fragments []Fragment

	// BoundingSphereRadius is the radius of the sphere that encompasses the
	// entire geometry.
	BoundingSphereRadius float64
}
