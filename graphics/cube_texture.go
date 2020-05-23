package graphics

import "github.com/go-gl/gl/v4.1-core/gl"

type CubeTexture struct {
	ID uint32
}

type CubeTextureData struct {
	Dimension      int32
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
	// TODO: Configure unpack alignment (it works now, since texture is at least 4 pixels wide and power of 2)
	gl.TexImage2D(gl.TEXTURE_CUBE_MAP_POSITIVE_X, 0, gl.RGB, data.Dimension, data.Dimension, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(data.RightSideData))
	gl.TexImage2D(gl.TEXTURE_CUBE_MAP_NEGATIVE_X, 0, gl.RGB, data.Dimension, data.Dimension, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(data.LeftSideData))
	gl.TexImage2D(gl.TEXTURE_CUBE_MAP_POSITIVE_Z, 0, gl.RGB, data.Dimension, data.Dimension, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(data.FrontSideData))
	gl.TexImage2D(gl.TEXTURE_CUBE_MAP_NEGATIVE_Z, 0, gl.RGB, data.Dimension, data.Dimension, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(data.BackSideData))
	// Note: Top and Bottom are flipped due to OpenGL's renderman issue
	gl.TexImage2D(gl.TEXTURE_CUBE_MAP_POSITIVE_Y, 0, gl.RGB, data.Dimension, data.Dimension, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(data.BottomSideData))
	gl.TexImage2D(gl.TEXTURE_CUBE_MAP_NEGATIVE_Y, 0, gl.RGB, data.Dimension, data.Dimension, 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(data.TopSideData))
	return nil
}

func (t *CubeTexture) Release() error {
	gl.DeleteTextures(1, &t.ID)
	t.ID = 0
	return nil
}
