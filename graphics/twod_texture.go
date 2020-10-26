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
	var maxAnisotropy float32
	gl.GetFloatv(gl.MAX_TEXTURE_MAX_ANISOTROPY, &maxAnisotropy)

	gl.GenTextures(1, &t.ID)
	gl.BindTexture(gl.TEXTURE_2D, t.ID)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameterf(gl.TEXTURE_2D, gl.TEXTURE_MAX_ANISOTROPY, maxAnisotropy)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.SRGB8_ALPHA8, data.Width, data.Height, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(data.Data))
	gl.GenerateMipmap(gl.TEXTURE_2D)
	return nil
}

func (t *TwoDTexture) Release() error {
	gl.DeleteTextures(1, &t.ID)
	t.ID = 0
	return nil
}
