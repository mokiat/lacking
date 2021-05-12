package graphics

import (
	"fmt"

	"github.com/go-gl/gl/v4.6-core/gl"

	"github.com/mokiat/lacking/framework/opengl"
)

type DataFormat int

const (
	DataFormatRGBA8 DataFormat = iota
	DataFormatRGBA32F
)

type CubeTexture struct {
	Texture *opengl.CubeTexture
}

type CubeTextureData struct {
	Dimension      int32
	Format         DataFormat
	FrontSideData  []byte
	BackSideData   []byte
	LeftSideData   []byte
	RightSideData  []byte
	TopSideData    []byte
	BottomSideData []byte
}

func (t *CubeTexture) Allocate(data CubeTextureData) error {
	t.Texture = opengl.NewCubeTexture()
	textureInfo := opengl.CubeTextureAllocateInfo{
		Dimension:      data.Dimension,
		FrontSideData:  data.FrontSideData,
		BackSideData:   data.BackSideData,
		LeftSideData:   data.LeftSideData,
		RightSideData:  data.RightSideData,
		TopSideData:    data.TopSideData,
		BottomSideData: data.BottomSideData,
	}
	switch data.Format {
	case DataFormatRGBA8:
		textureInfo.InternalFormat = gl.SRGB8
		textureInfo.DataFormat = gl.RGBA
		textureInfo.DataComponentType = gl.UNSIGNED_BYTE
	case DataFormatRGBA32F:
		textureInfo.InternalFormat = gl.RGBA32F
		textureInfo.DataFormat = gl.RGBA
		textureInfo.DataComponentType = gl.FLOAT
	default:
		return fmt.Errorf("unknown format: %d", data.Format)
	}
	t.Texture.Allocate(textureInfo)
	return nil
}

func (t *CubeTexture) Release() error {
	t.Texture.Release()
	return nil
}
