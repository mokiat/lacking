package graphics

import (
	"github.com/mokiat/lacking/opengl"
)

type TwoDTexture struct {
	Texture *opengl.TwoDTexture
}

type TwoDTextureData struct {
	Width  int32
	Height int32
	Data   []byte
}

func (t *TwoDTexture) ID() uint32 {
	return t.Texture.ID()
}

func (t *TwoDTexture) Allocate(data TwoDTextureData) error {
	t.Texture = opengl.NewTwoDTexture()
	textureInfo := opengl.TwoDTextureAllocateInfo{
		Width:  data.Width,
		Height: data.Height,
		Data:   data.Data,
	}
	return t.Texture.Allocate(textureInfo)
}

func (t *TwoDTexture) Release() error {
	return t.Texture.Release()
}
