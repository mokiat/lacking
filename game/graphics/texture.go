package graphics

import "github.com/mokiat/lacking/render"

// Deprecated: Remove this file. Graphics should use render.Texture
// directly. The `game` package should deal with higher-level texture
// management / modification capabilities.

type Texture interface {
	Name() string
}

func newTwoDTexture(texture render.Texture) *TwoDTexture {
	return &TwoDTexture{
		texture: texture,
	}
}

// TwoDTexture represents a two-dimensional texture.
type TwoDTexture struct {
	texture render.Texture
}

func (t *TwoDTexture) Texture() render.Texture {
	return t.texture
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

	GenerateMipmaps bool
	GammaCorrection bool

	InternalFormat InternalFormat
	DataFormat     DataFormat
	Data           []byte
}

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
