package render

// VertexArrayMarker marks a type as being a VertexArray.
type VertexArrayMarker interface {
	_isVertexArrayType()
}

// VertexArray represents the geometry for a mesh.
type VertexArray interface {
	VertexArrayMarker
	Resource
}

// VertexArrayInfo represents the information needed to create a VertexArray.
type VertexArrayInfo struct {

	// Bindings specifies the vertex buffers that make up this Vertex Array.
	Bindings []VertexArrayBinding

	// Attributes specifies the vertex attributes.
	Attributes []VertexArrayAttribute

	// IndexFormat specifies the format of the index buffer.
	IndexFormat IndexFormat

	// IndexBuffer specifies the index buffer.
	IndexBuffer Buffer
}

// NewVertexArrayBinding creates a new VertexArrayBinding with the specified
// buffer and stride.
func NewVertexArrayBinding(buffer Buffer, stride uint32) VertexArrayBinding {
	return VertexArrayBinding{
		VertexBuffer: buffer,
		Stride:       stride,
	}
}

// VertexArrayBinding represents a vertex buffer binding for a VertexArray.
type VertexArrayBinding struct {

	// VertexBuffer specifies the vertex buffer.
	VertexBuffer Buffer

	// Stride specifies the byte stride of subsequent elements in the buffer.
	Stride uint32
}

// NewVertexArrayAttribute creates a new VertexArrayAttribute with the specified
// binding, location, offset and format.
func NewVertexArrayAttribute(binding, location uint, offset uint32, format VertexAttributeFormat) VertexArrayAttribute {
	return VertexArrayAttribute{
		Binding:  binding,
		Location: location,
		Offset:   offset,
		Format:   format,
	}
}

// VertexArrayAttribute represents a vertex attribute for a VertexArray.
type VertexArrayAttribute struct {

	// Binding specifies the binding inside the Vertex Array that this uses.
	Binding uint

	// Location specifies the location of the attribute in the shader.
	Location uint

	// Format specifies the format of the attribute.
	Format VertexAttributeFormat

	// Offset specifies the byte offset of the attribute in the buffer.
	Offset uint32
}

const (
	// VertexAttributeFormatR32F specifies that the vertex attribute is a
	// single 32-bit float.
	VertexAttributeFormatR32F VertexAttributeFormat = iota

	// VertexAttributeFormatRG32F specifies that the vertex attribute is a
	// two-component 32-bit float.
	VertexAttributeFormatRG32F

	// VertexAttributeFormatRGB32F specifies that the vertex attribute is a
	// three-component 32-bit float.
	VertexAttributeFormatRGB32F

	// VertexAttributeFormatRGBA32F specifies that the vertex attribute is a
	// four-component 32-bit float.
	VertexAttributeFormatRGBA32F

	// VertexAttributeFormatR16F specifies that the vertex attribute is a
	// single 16-bit float.
	VertexAttributeFormatR16F

	// VertexAttributeFormatRG16F specifies that the vertex attribute is a
	// two-component 16-bit float.
	VertexAttributeFormatRG16F

	// VertexAttributeFormatRGB16F specifies that the vertex attribute is a
	// three-component 16-bit float.
	VertexAttributeFormatRGB16F

	// VertexAttributeFormatRGBA16F specifies that the vertex attribute is a
	// four-component 16-bit float.
	VertexAttributeFormatRGBA16F

	// VertexAttributeFormatR16S specifies that the vertex attribute is a
	// single 16-bit signed integer.
	VertexAttributeFormatR16S

	// VertexAttributeFormatRG16S specifies that the vertex attribute is a
	// two-component 16-bit signed integer.
	VertexAttributeFormatRG16S

	// VertexAttributeFormatRGB16S specifies that the vertex attribute is a
	// three-component 16-bit signed integer.
	VertexAttributeFormatRGB16S

	// VertexAttributeFormatRGBA16S specifies that the vertex attribute is a
	// four-component 16-bit signed integer.
	VertexAttributeFormatRGBA16S

	// VertexAttributeFormatR16SN specifies that the vertex attribute is a
	// single 16-bit signed integer that is normalized to [-1, 1].
	VertexAttributeFormatR16SN

	// VertexAttributeFormatRG16SN specifies that the vertex attribute is a
	// two-component 16-bit signed integer that is normalized to [-1, 1].
	VertexAttributeFormatRG16SN

	// VertexAttributeFormatRGB16SN specifies that the vertex attribute is a
	// three-component 16-bit signed integer that is normalized to [-1, 1].
	VertexAttributeFormatRGB16SN

	// VertexAttributeFormatRGBA16SN specifies that the vertex attribute is a
	// four-component 16-bit signed integer that is normalized to [-1, 1].
	VertexAttributeFormatRGBA16SN

	// VertexAttributeFormatR16U specifies that the vertex attribute is a
	// single 16-bit unsigned integer.
	VertexAttributeFormatR16U

	// VertexAttributeFormatRG16U specifies that the vertex attribute is a
	// two-component 16-bit unsigned integer.
	VertexAttributeFormatRG16U

	// VertexAttributeFormatRGB16U specifies that the vertex attribute is a
	// three-component 16-bit unsigned integer.
	VertexAttributeFormatRGB16U

	// VertexAttributeFormatRGBA16U specifies that the vertex attribute is a
	// four-component 16-bit unsigned integer.
	VertexAttributeFormatRGBA16U

	// VertexAttributeFormatR16UN specifies that the vertex attribute is a
	// single 16-bit unsigned integer that is normalized to [0, 1].
	VertexAttributeFormatR16UN

	// VertexAttributeFormatRG16UN specifies that the vertex attribute is a
	// two-component 16-bit unsigned integer that is normalized to [0, 1].
	VertexAttributeFormatRG16UN

	// VertexAttributeFormatRGB16UN specifies that the vertex attribute is a
	// three-component 16-bit unsigned integer that is normalized to [0, 1].
	VertexAttributeFormatRGB16UN

	// VertexAttributeFormatRGBA16UN specifies that the vertex attribute is a
	// four-component 16-bit unsigned integer that is normalized to [0, 1].
	VertexAttributeFormatRGBA16UN

	// VertexAttributeFormatR8S specifies that the vertex attribute is a
	// single 8-bit signed integer.
	VertexAttributeFormatR8S

	// VertexAttributeFormatRG8S specifies that the vertex attribute is a
	// two-component 8-bit signed integer.
	VertexAttributeFormatRG8S

	// VertexAttributeFormatRGB8S specifies that the vertex attribute is a
	// three-component 8-bit signed integer.
	VertexAttributeFormatRGB8S

	// VertexAttributeFormatRGBA8S specifies that the vertex attribute is a
	// four-component 8-bit signed integer.
	VertexAttributeFormatRGBA8S

	// VertexAttributeFormatR8SN specifies that the vertex attribute is a
	// single 8-bit signed integer that is normalized to [-1, 1].
	VertexAttributeFormatR8SN

	// VertexAttributeFormatRG8SN specifies that the vertex attribute is a
	// two-component 8-bit signed integer that is normalized to [-1, 1].
	VertexAttributeFormatRG8SN

	// VertexAttributeFormatRGB8SN specifies that the vertex attribute is a
	// three-component 8-bit signed integer that is normalized to [-1, 1].
	VertexAttributeFormatRGB8SN

	// VertexAttributeFormatRGBA8SN specifies that the vertex attribute is a
	// four-component 8-bit signed integer that is normalized to [-1, 1].
	VertexAttributeFormatRGBA8SN

	// VertexAttributeFormatR8U specifies that the vertex attribute is a
	// single 8-bit unsigned integer.
	VertexAttributeFormatR8U

	// VertexAttributeFormatRG8U specifies that the vertex attribute is a
	// two-component 8-bit unsigned integer.
	VertexAttributeFormatRG8U

	// VertexAttributeFormatRGB8U specifies that the vertex attribute is a
	// three-component 8-bit unsigned integer.
	VertexAttributeFormatRGB8U

	// VertexAttributeFormatRGBA8U specifies that the vertex attribute is a
	// four-component 8-bit unsigned integer.
	VertexAttributeFormatRGBA8U

	// VertexAttributeFormatR8UN specifies that the vertex attribute is a
	// single 8-bit unsigned integer that is normalized to [0, 1].
	VertexAttributeFormatR8UN

	// VertexAttributeFormatRG8UN specifies that the vertex attribute is a
	// two-component 8-bit unsigned integer that is normalized to [0, 1].
	VertexAttributeFormatRG8UN

	// VertexAttributeFormatRGB8UN specifies that the vertex attribute is a
	// three-component 8-bit unsigned integer that is normalized to [0, 1].
	VertexAttributeFormatRGB8UN

	// VertexAttributeFormatRGBA8UN specifies that the vertex attribute is a
	// four-component 8-bit unsigned integer that is normalized to [0, 1].
	VertexAttributeFormatRGBA8UN

	// VertexAttributeFormatR8IU specifies that the vertex attribute is a single
	// 8-bit unsigned integer that should be treated as an integer on the GPU.
	VertexAttributeFormatR8IU

	// VertexAttributeFormatRG8IU specifies that the vertex attribute is a
	// two-component 8-bit unsigned integer that should be treated as an integer
	// on the GPU.
	VertexAttributeFormatRG8IU

	// VertexAttributeFormatRGB8IU specifies that the vertex attribute is a
	// three-component 8-bit unsigned integer that should be treated as an
	// integer on the GPU.
	VertexAttributeFormatRGB8IU

	// VertexAttributeFormatRGBA8IU specifies that the vertex attribute is a
	// four-component 8-bit unsigned integer that should be treated as an
	// integer on the GPU.
	VertexAttributeFormatRGBA8IU
)

// VertexAttributeFormat specifies the format of a vertex attribute.
type VertexAttributeFormat uint8

// String returns a string representation of the VertexAttributeFormat.
func (f VertexAttributeFormat) String() string {
	switch f {
	case VertexAttributeFormatR32F:
		return "R32F"
	case VertexAttributeFormatRG32F:
		return "RG32F"
	case VertexAttributeFormatRGB32F:
		return "RGB32F"
	case VertexAttributeFormatRGBA32F:
		return "RGBA32F"
	case VertexAttributeFormatR16F:
		return "R16F"
	case VertexAttributeFormatRG16F:
		return "RG16F"
	case VertexAttributeFormatRGB16F:
		return "RGB16F"
	case VertexAttributeFormatRGBA16F:
		return "RGBA16F"
	case VertexAttributeFormatR16S:
		return "R16S"
	case VertexAttributeFormatRG16S:
		return "RG16S"
	case VertexAttributeFormatRGB16S:
		return "RGB16S"
	case VertexAttributeFormatRGBA16S:
		return "RGBA16S"
	case VertexAttributeFormatR16SN:
		return "R16SN"
	case VertexAttributeFormatRG16SN:
		return "RG16SN"
	case VertexAttributeFormatRGB16SN:
		return "RGB16SN"
	case VertexAttributeFormatRGBA16SN:
		return "RGBA16SN"
	case VertexAttributeFormatR16U:
		return "R16U"
	case VertexAttributeFormatRG16U:
		return "RG16U"
	case VertexAttributeFormatRGB16U:
		return "RGB16U"
	case VertexAttributeFormatRGBA16U:
		return "RGBA16U"
	case VertexAttributeFormatR16UN:
		return "R16UN"
	case VertexAttributeFormatRG16UN:
		return "RG16UN"
	case VertexAttributeFormatRGB16UN:
		return "RGB16UN"
	case VertexAttributeFormatRGBA16UN:
		return "RGBA16UN"
	case VertexAttributeFormatR8S:
		return "R8S"
	case VertexAttributeFormatRG8S:
		return "RG8S"
	case VertexAttributeFormatRGB8S:
		return "RGB8S"
	case VertexAttributeFormatRGBA8S:
		return "RGBA8S"
	case VertexAttributeFormatR8SN:
		return "R8SN"
	case VertexAttributeFormatRG8SN:
		return "RG8SN"
	case VertexAttributeFormatRGB8SN:
		return "RGB8SN"
	case VertexAttributeFormatRGBA8SN:
		return "RGBA8SN"
	case VertexAttributeFormatR8U:
		return "R8U"
	case VertexAttributeFormatRG8U:
		return "RG8U"
	case VertexAttributeFormatRGB8U:
		return "RGB8U"
	case VertexAttributeFormatRGBA8U:
		return "RGBA8U"
	case VertexAttributeFormatR8UN:
		return "R8UN"
	case VertexAttributeFormatRG8UN:
		return "RG8UN"
	case VertexAttributeFormatRGB8UN:
		return "RGB8UN"
	case VertexAttributeFormatRGBA8UN:
		return "RGBA8UN"
	case VertexAttributeFormatR8IU:
		return "R8IU"
	case VertexAttributeFormatRG8IU:
		return "RG8IU"
	case VertexAttributeFormatRGB8IU:
		return "RGB8IU"
	case VertexAttributeFormatRGBA8IU:
		return "RGBA8IU"
	default:
		return "UNKNOWN"
	}
}

const (
	// IndexFormatUnsignedU16 specifies that the index buffer is unsigned
	// short.
	IndexFormatUnsignedU16 IndexFormat = iota

	// IndexFormatUnsignedU32 specifies that the index buffer is unsigned int.
	IndexFormatUnsignedU32
)

// IndexFormat specifies the format of an index buffer.
type IndexFormat uint8

func (f IndexFormat) String() string {
	switch f {
	case IndexFormatUnsignedU16:
		return "UNSIGNED_SHORT"
	case IndexFormatUnsignedU32:
		return "UNSIGNED_INT"
	default:
		return "UNKNOWN"
	}
}
