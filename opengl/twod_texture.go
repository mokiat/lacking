package opengl

import (
	"fmt"

	"github.com/go-gl/gl/v4.6-core/gl"
)

func NewTwoDTexture() *TwoDTexture {
	return &TwoDTexture{}
}

type TwoDTexture struct {
	Texture
}

func (t *TwoDTexture) ID() uint32 {
	return t.id
}

func (t *TwoDTexture) Allocate(info TwoDTextureAllocateInfo) error {
	gl.CreateTextures(gl.TEXTURE_2D, 1, &t.id)
	if t.id == 0 {
		return fmt.Errorf("failed to allocate texture")
	}

	gl.TextureParameteri(t.id, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TextureParameteri(t.id, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.TextureParameteri(t.id, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR)
	gl.TextureParameteri(t.id, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TextureParameterf(t.id, gl.TEXTURE_MAX_ANISOTROPY, t.maxAnisotropy())

	gl.TextureStorage2D(t.id, info.levels(), gl.SRGB8_ALPHA8, info.Width, info.Height)
	gl.TextureSubImage2D(t.id, 0, 0, 0, info.Width, info.Height, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(info.Data))
	gl.GenerateTextureMipmap(t.id)
	return nil
}

func (t *TwoDTexture) Release() error {
	gl.DeleteTextures(1, &t.id)
	t.id = 0
	return nil
}

func (t *TwoDTexture) maxAnisotropy() float32 {
	var maxAnisotropy float32
	gl.GetFloatv(gl.MAX_TEXTURE_MAX_ANISOTROPY, &maxAnisotropy)
	return maxAnisotropy
}

type TwoDTextureAllocateInfo struct {
	Width  int32
	Height int32
	Data   []byte
}

func (i TwoDTextureAllocateInfo) levels() int32 {
	count := int32(1)
	width, height := i.Width, i.Height
	for width > 1 || height > 1 {
		width /= 2
		height /= 2
		count++
	}
	return count
}
