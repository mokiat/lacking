package graphics

// TwoDTexture represents a two-dimensional texture.
type TwoDTexture interface {

	// Delete releases any resources allocated for this texture.
	Delete()
}

// CubeTexture represents a cube texture.
type CubeTexture interface {

	// Delete releases any resources allocated for this texture.
	Delete()
}

// CubeTextureDefinition contains all the information needed
// to create a CubeTexture.
type CubeTextureDefinition struct {
	Dimension int

	WrapS Wrap
	WrapT Wrap

	MinFilter Filter
	MagFilter Filter

	InternalFormat InternalFormat
	DataFormat     DataFormat

	FrontSideData  []byte
	BackSideData   []byte
	LeftSideData   []byte
	RightSideData  []byte
	TopSideData    []byte
	BottomSideData []byte
}

const (
	WrapClampToEdge Wrap = 1 + iota
	WrapRepeat
)

type Wrap int

const (
	FilterNearest Filter = 1 + iota
	FilterLinear
	FilterNearestMipmapNearest
	FilterNearestMipmapLinear
	FilterLinearMipmapNearest
	FilterLinearMipmapLinear
)

type Filter int

const (
	DataFormatRGBA8 DataFormat = 1 + iota
	DataFormatRGBA32F
)

type DataFormat int

const (
	InternalFormatRGBA8 InternalFormat = 1 + iota
	InternalFormatRGBA32F
)

type InternalFormat int
