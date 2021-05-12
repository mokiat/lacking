package graphics

import (
	"github.com/go-gl/gl/v4.6-core/gl"

	"github.com/mokiat/lacking/framework/opengl"
)

type TwoDTexture struct {
	Texture *opengl.TwoDTexture
}

type TwoDTextureData struct {
	Width  int32
	Height int32
	Data   []byte
}

func (t *TwoDTexture) Allocate(data TwoDTextureData) error {
	t.Texture = opengl.NewTwoDTexture()
	textureInfo := opengl.TwoDTextureAllocateInfo{
		Width:             data.Width,
		Height:            data.Height,
		UseAnisotropy:     true,
		GenerateMipmaps:   true,
		InternalFormat:    gl.SRGB8_ALPHA8,
		DataFormat:        gl.RGBA,
		DataComponentType: gl.UNSIGNED_BYTE,
		Data:              data.Data,
	}
	t.Texture.Allocate(textureInfo)
	return nil
}

func (t *TwoDTexture) Release() error {
	t.Texture.Release()
	return nil
}
