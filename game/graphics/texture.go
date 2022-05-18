package graphics

import "github.com/mokiat/lacking/render"

func newTwoDTexture(texture render.Texture) *TwoDTexture {
	return &TwoDTexture{
		texture: texture,
	}
}

// TwoDTexture represents a two-dimensional texture.
type TwoDTexture struct {
	texture render.Texture
}

// Delete releases any resources allocated for this texture.
func (t *TwoDTexture) Delete() {
	t.texture.Release()
}

// TwoDTextureDefinition contains all the information needed
// to create a TwoDTexture.
type TwoDTextureDefinition struct {
	Width  int
	Height int

	WrapS Wrap
	WrapT Wrap

	MinFilter     Filter
	MagFilter     Filter
	UseAnisotropy bool

	InternalFormat InternalFormat
	DataFormat     DataFormat

	Data []byte
}

func newCubeTexture(texture render.Texture) *CubeTexture {
	return &CubeTexture{
		texture: texture,
	}
}

// CubeTexture represents a cube texture.
type CubeTexture struct {
	texture render.Texture
}

// Delete releases any resources allocated for this texture.
func (t *CubeTexture) Delete() {
	t.texture.Release()
}

// CubeTextureDefinition contains all the information needed
// to create a CubeTexture.
type CubeTextureDefinition struct {
	Dimension int

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
	WrapMirroredRepat
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
	DataFormatRGBA16F
	DataFormatRGBA32F
)

type DataFormat int

const (
	InternalFormatRGBA8 InternalFormat = 1 + iota
	InternalFormatRGBA16F
	InternalFormatRGBA32F
)

type InternalFormat int
