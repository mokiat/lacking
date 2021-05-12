package opengl

import (
	"fmt"
	"runtime"

	"github.com/go-gl/gl/v4.6-core/gl"
)

func NewTwoDTexture() *TwoDTexture {
	return &TwoDTexture{}
}

type TwoDTexture struct {
	Texture
}

func (t *TwoDTexture) Allocate(info TwoDTextureAllocateInfo) {
	if t.id != 0 {
		panic(fmt.Errorf("texture already allocated"))
	}
	gl.CreateTextures(gl.TEXTURE_2D, 1, &t.id)
	if t.id == 0 {
		panic(fmt.Errorf("failed to allocate texture"))
	}

	gl.TextureParameteri(t.id, gl.TEXTURE_WRAP_S, info.wrapS())
	gl.TextureParameteri(t.id, gl.TEXTURE_WRAP_T, info.wrapT())
	gl.TextureParameteri(t.id, gl.TEXTURE_MIN_FILTER, info.minFilter())
	gl.TextureParameteri(t.id, gl.TEXTURE_MAG_FILTER, info.magFilter())
	if info.UseAnisotropy {
		gl.TextureParameterf(t.id, gl.TEXTURE_MAX_ANISOTROPY, info.maxAnisotropy())
	}

	gl.TextureStorage2D(t.id, info.levels(), info.internalFormat(), info.Width, info.Height)
	if info.Data != nil {
		gl.TextureSubImage2D(t.id, 0, 0, 0, info.Width, info.Height, info.dataFormat(), info.dataComponentType(), gl.Ptr(info.Data))
		runtime.KeepAlive(info.Data)
	}
	if info.GenerateMipmaps {
		gl.GenerateTextureMipmap(t.id)
	}
}

func (t *TwoDTexture) Release() {
	if t.id == 0 {
		panic(fmt.Errorf("texture already released"))
	}
	gl.DeleteTextures(1, &t.id)
	t.id = 0
}

type TwoDTextureAllocateInfo struct {
	Width             int32
	Height            int32
	WrapS             int32
	WrapT             int32
	MinFilter         int32
	MagFilter         int32
	UseAnisotropy     bool
	GenerateMipmaps   bool
	InternalFormat    uint32
	DataFormat        uint32
	DataComponentType uint32
	Data              []byte
}

func (i TwoDTextureAllocateInfo) wrapS() int32 {
	if i.WrapS == 0 {
		return gl.REPEAT
	}
	return i.WrapS
}

func (i TwoDTextureAllocateInfo) wrapT() int32 {
	if i.WrapT == 0 {
		return gl.REPEAT
	}
	return i.WrapT
}

func (i TwoDTextureAllocateInfo) minFilter() int32 {
	if i.MinFilter == 0 {
		return gl.LINEAR_MIPMAP_LINEAR
	}
	return i.MinFilter
}

func (i TwoDTextureAllocateInfo) magFilter() int32 {
	if i.MagFilter == 0 {
		return gl.LINEAR
	}
	return i.MagFilter
}

func (i TwoDTextureAllocateInfo) maxAnisotropy() float32 {
	var maxAnisotropy float32
	gl.GetFloatv(gl.MAX_TEXTURE_MAX_ANISOTROPY, &maxAnisotropy)
	return maxAnisotropy
}

func (i TwoDTextureAllocateInfo) internalFormat() uint32 {
	if i.InternalFormat == 0 {
		return gl.SRGB8_ALPHA8
	}
	return i.InternalFormat
}

func (i TwoDTextureAllocateInfo) dataFormat() uint32 {
	if i.DataFormat == 0 {
		return gl.RGBA
	}
	return i.DataFormat
}

func (i TwoDTextureAllocateInfo) dataComponentType() uint32 {
	if i.DataComponentType == 0 {
		return gl.UNSIGNED_BYTE
	}
	return i.DataComponentType
}

func (i TwoDTextureAllocateInfo) levels() int32 {
	if !i.GenerateMipmaps {
		return 1
	}
	count := int32(1)
	width, height := i.Width, i.Height
	for width > 1 || height > 1 {
		width /= 2
		height /= 2
		count++
	}
	return count
}
