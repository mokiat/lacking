package opengl

import (
	"fmt"

	"github.com/go-gl/gl/v4.6-core/gl"
)

func NewCubeTexture() *CubeTexture {
	return &CubeTexture{}
}

type CubeTexture struct {
	id uint32
}

func (t *CubeTexture) ID() uint32 {
	return t.id
}

func (t *CubeTexture) Allocate(info CubeTextureAllocateInfo) error {
	gl.CreateTextures(gl.TEXTURE_CUBE_MAP, 1, &t.id)
	if t.id == 0 {
		return fmt.Errorf("failed to allocate texture")
	}

	gl.TextureParameteri(t.id, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TextureParameteri(t.id, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TextureParameteri(t.id, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TextureParameteri(t.id, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	// Note: Top and Bottom are flipped due to OpenGL's renderman issue
	gl.TextureStorage2D(t.id, 1, info.InternalFormat, info.Dimension, info.Dimension)
	gl.TextureSubImage3D(t.id, 0, 0, 0, 0, info.Dimension, info.Dimension, 1, info.DataFormat, info.DataComponentType, gl.Ptr(info.RightSideData))
	gl.TextureSubImage3D(t.id, 0, 0, 0, 1, info.Dimension, info.Dimension, 1, info.DataFormat, info.DataComponentType, gl.Ptr(info.LeftSideData))
	gl.TextureSubImage3D(t.id, 0, 0, 0, 2, info.Dimension, info.Dimension, 1, info.DataFormat, info.DataComponentType, gl.Ptr(info.BottomSideData))
	gl.TextureSubImage3D(t.id, 0, 0, 0, 3, info.Dimension, info.Dimension, 1, info.DataFormat, info.DataComponentType, gl.Ptr(info.TopSideData))
	gl.TextureSubImage3D(t.id, 0, 0, 0, 4, info.Dimension, info.Dimension, 1, info.DataFormat, info.DataComponentType, gl.Ptr(info.FrontSideData))
	gl.TextureSubImage3D(t.id, 0, 0, 0, 5, info.Dimension, info.Dimension, 1, info.DataFormat, info.DataComponentType, gl.Ptr(info.BackSideData))
	return nil
}

func (t *CubeTexture) Release() error {
	gl.DeleteTextures(1, &t.id)
	t.id = 0
	return nil
}

type CubeTextureAllocateInfo struct {
	Dimension         int32
	InternalFormat    uint32
	DataFormat        uint32
	DataComponentType uint32
	FrontSideData     []byte
	BackSideData      []byte
	LeftSideData      []byte
	RightSideData     []byte
	TopSideData       []byte
	BottomSideData    []byte
}

func (i CubeTextureAllocateInfo) levels() int32 {
	count := int32(1)
	dimension := i.Dimension
	for dimension > 1 {
		dimension /= 2
		count++
	}
	return count
}
