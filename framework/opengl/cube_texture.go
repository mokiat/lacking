package opengl

import (
	"fmt"
	"runtime"

	"github.com/go-gl/gl/v4.6-core/gl"
)

func NewCubeTexture() *CubeTexture {
	return &CubeTexture{}
}

type CubeTexture struct {
	Texture
}

func (t *CubeTexture) Allocate(info CubeTextureAllocateInfo) {
	if t.id != 0 {
		panic(fmt.Errorf("texture already allocated"))
	}
	gl.CreateTextures(gl.TEXTURE_CUBE_MAP, 1, &t.id)
	if t.id == 0 {
		panic(fmt.Errorf("failed to allocate texture"))
	}

	gl.TextureParameteri(t.id, gl.TEXTURE_WRAP_S, info.wrapS())
	gl.TextureParameteri(t.id, gl.TEXTURE_WRAP_T, info.wrapT())
	gl.TextureParameteri(t.id, gl.TEXTURE_MIN_FILTER, info.minFilter())
	gl.TextureParameteri(t.id, gl.TEXTURE_MAG_FILTER, info.magFilter())

	// Note: Top and Bottom are flipped due to OpenGL's renderman issue
	gl.TextureStorage2D(t.id, 1, info.internalFormat(), info.Dimension, info.Dimension)
	gl.TextureSubImage3D(t.id, 0, 0, 0, 0, info.Dimension, info.Dimension, 1, info.dataFormat(), info.dataComponentType(), gl.Ptr(info.RightSideData))
	runtime.KeepAlive(info.RightSideData)
	gl.TextureSubImage3D(t.id, 0, 0, 0, 1, info.Dimension, info.Dimension, 1, info.dataFormat(), info.dataComponentType(), gl.Ptr(info.LeftSideData))
	runtime.KeepAlive(info.LeftSideData)
	gl.TextureSubImage3D(t.id, 0, 0, 0, 2, info.Dimension, info.Dimension, 1, info.dataFormat(), info.dataComponentType(), gl.Ptr(info.BottomSideData))
	runtime.KeepAlive(info.BottomSideData)
	gl.TextureSubImage3D(t.id, 0, 0, 0, 3, info.Dimension, info.Dimension, 1, info.dataFormat(), info.dataComponentType(), gl.Ptr(info.TopSideData))
	runtime.KeepAlive(info.TopSideData)
	gl.TextureSubImage3D(t.id, 0, 0, 0, 4, info.Dimension, info.Dimension, 1, info.dataFormat(), info.dataComponentType(), gl.Ptr(info.FrontSideData))
	runtime.KeepAlive(info.FrontSideData)
	gl.TextureSubImage3D(t.id, 0, 0, 0, 5, info.Dimension, info.Dimension, 1, info.dataFormat(), info.dataComponentType(), gl.Ptr(info.BackSideData))
	runtime.KeepAlive(info.BackSideData)
}

func (t *CubeTexture) Release() {
	if t.id == 0 {
		panic(fmt.Errorf("texture already released"))
	}
	gl.DeleteTextures(1, &t.id)
	t.id = 0
}

type CubeTextureAllocateInfo struct {
	Dimension         int32
	WrapS             int32
	WrapT             int32
	MinFilter         int32
	MagFilter         int32
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

func (i CubeTextureAllocateInfo) wrapS() int32 {
	if i.WrapS == 0 {
		return gl.CLAMP_TO_EDGE
	}
	return i.WrapS
}

func (i CubeTextureAllocateInfo) wrapT() int32 {
	if i.WrapT == 0 {
		return gl.CLAMP_TO_EDGE
	}
	return i.WrapT
}

func (i CubeTextureAllocateInfo) minFilter() int32 {
	if i.MinFilter == 0 {
		return gl.LINEAR_MIPMAP_LINEAR
	}
	return i.MinFilter
}

func (i CubeTextureAllocateInfo) magFilter() int32 {
	if i.MagFilter == 0 {
		return gl.LINEAR
	}
	return i.MagFilter
}

func (i CubeTextureAllocateInfo) internalFormat() uint32 {
	if i.InternalFormat == 0 {
		return gl.SRGB8_ALPHA8
	}
	return i.InternalFormat
}

func (i CubeTextureAllocateInfo) dataFormat() uint32 {
	if i.DataFormat == 0 {
		return gl.RGBA
	}
	return i.DataFormat
}

func (i CubeTextureAllocateInfo) dataComponentType() uint32 {
	if i.DataComponentType == 0 {
		return gl.UNSIGNED_BYTE
	}
	return i.DataComponentType
}
