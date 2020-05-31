package graphics

import "github.com/go-gl/gl/v4.1-core/gl"

type TwoDTexture struct {
	ID uint32
}

type TwoDTextureData struct {
	Width  int32
	Height int32
	Data   []byte
}

func (t *TwoDTexture) Allocate(data TwoDTextureData) error {
	gl.GenTextures(1, &t.ID)
	gl.BindTexture(gl.TEXTURE_2D, t.ID)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.SRGB8_ALPHA8, data.Width, data.Height, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(data.Data))
	return nil
}

func (t *TwoDTexture) Release() error {
	gl.DeleteTextures(1, &t.ID)
	t.ID = 0
	return nil
}
