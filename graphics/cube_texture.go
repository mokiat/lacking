package graphics

import (
	"fmt"

	"github.com/go-gl/gl/v4.6-core/gl"
)

type DataFormat int

const (
	DataFormatRGBA8 DataFormat = iota
	DataFormatRGBA32F
)

type CubeTexture struct {
	ID uint32
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
	gl.GenTextures(1, &t.ID)
	gl.BindTexture(gl.TEXTURE_CUBE_MAP, t.ID)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_MAG_FILTER, gl.LINEAR)

	var (
		targetFormat  int32
		format        uint32
		componentType uint32
	)
	switch data.Format {
	case DataFormatRGBA8:
		targetFormat = gl.SRGB8
		format = gl.RGBA
		componentType = gl.UNSIGNED_BYTE
	case DataFormatRGBA32F:
		targetFormat = gl.RGBA32F
		format = gl.RGBA
		componentType = gl.FLOAT
	default:
		return fmt.Errorf("unknown format: %d", data.Format)
	}
	gl.TexImage2D(gl.TEXTURE_CUBE_MAP_POSITIVE_X, 0, targetFormat, data.Dimension, data.Dimension, 0, format, componentType, gl.Ptr(data.RightSideData))
	gl.TexImage2D(gl.TEXTURE_CUBE_MAP_NEGATIVE_X, 0, targetFormat, data.Dimension, data.Dimension, 0, format, componentType, gl.Ptr(data.LeftSideData))
	gl.TexImage2D(gl.TEXTURE_CUBE_MAP_POSITIVE_Z, 0, targetFormat, data.Dimension, data.Dimension, 0, format, componentType, gl.Ptr(data.FrontSideData))
	gl.TexImage2D(gl.TEXTURE_CUBE_MAP_NEGATIVE_Z, 0, targetFormat, data.Dimension, data.Dimension, 0, format, componentType, gl.Ptr(data.BackSideData))
	// Note: Top and Bottom are flipped due to OpenGL's renderman issue
	gl.TexImage2D(gl.TEXTURE_CUBE_MAP_POSITIVE_Y, 0, targetFormat, data.Dimension, data.Dimension, 0, format, componentType, gl.Ptr(data.BottomSideData))
	gl.TexImage2D(gl.TEXTURE_CUBE_MAP_NEGATIVE_Y, 0, targetFormat, data.Dimension, data.Dimension, 0, format, componentType, gl.Ptr(data.TopSideData))
	return nil
}

func (t *CubeTexture) Release() error {
	gl.DeleteTextures(1, &t.ID)
	t.ID = 0
	return nil
}
