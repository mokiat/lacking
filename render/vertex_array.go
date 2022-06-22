package render

type VertexArrayInfo struct {
	Bindings    []VertexArrayBindingInfo
	Attributes  []VertexArrayAttributeInfo
	IndexFormat IndexFormat
	IndexBuffer Buffer
}

type VertexArrayBindingInfo struct {
	VertexBuffer Buffer
	Stride       int
}

type VertexArrayAttributeInfo struct {
	Binding  int
	Location int
	Format   VertexAttributeFormat
	Offset   int
}

type VertexAttributeFormat uint8

const (
	VertexAttributeFormatR32F VertexAttributeFormat = iota
	VertexAttributeFormatRG32F
	VertexAttributeFormatRGB32F
	VertexAttributeFormatRGBA32F

	VertexAttributeFormatR16F
	VertexAttributeFormatRG16F
	VertexAttributeFormatRGB16F
	VertexAttributeFormatRGBA16F

	VertexAttributeFormatR16S
	VertexAttributeFormatRG16S
	VertexAttributeFormatRGB16S
	VertexAttributeFormatRGBA16S

	VertexAttributeFormatR16SN
	VertexAttributeFormatRG16SN
	VertexAttributeFormatRGB16SN
	VertexAttributeFormatRGBA16SN

	VertexAttributeFormatR16U
	VertexAttributeFormatRG16U
	VertexAttributeFormatRGB16U
	VertexAttributeFormatRGBA16U

	VertexAttributeFormatR16UN
	VertexAttributeFormatRG16UN
	VertexAttributeFormatRGB16UN
	VertexAttributeFormatRGBA16UN

	VertexAttributeFormatR8S
	VertexAttributeFormatRG8S
	VertexAttributeFormatRGB8S
	VertexAttributeFormatRGBA8S

	VertexAttributeFormatR8SN
	VertexAttributeFormatRG8SN
	VertexAttributeFormatRGB8SN
	VertexAttributeFormatRGBA8SN

	VertexAttributeFormatR8U
	VertexAttributeFormatRG8U
	VertexAttributeFormatRGB8U
	VertexAttributeFormatRGBA8U

	VertexAttributeFormatR8UN
	VertexAttributeFormatRG8UN
	VertexAttributeFormatRGB8UN
	VertexAttributeFormatRGBA8UN

	VertexAttributeFormatR8IU
	VertexAttributeFormatRG8IU
	VertexAttributeFormatRGB8IU
	VertexAttributeFormatRGBA8IU
)

type IndexFormat uint8

const (
	// NOTE: Do not add IndexFormatUnsignedByte as it may be slow on some GPUs.

	IndexFormatUnsignedShort IndexFormat = iota
	IndexFormatUnsignedInt
)

type VertexArrayObject interface {
	_isVertexArrayObject() bool // ensures interface uniqueness
}

type VertexArray interface {
	VertexArrayObject
	Release()
}
