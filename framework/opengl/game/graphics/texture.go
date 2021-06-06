package graphics

import (
	"github.com/mokiat/lacking/framework/opengl"
	"github.com/mokiat/lacking/game/graphics"
)

func newTwoDTexture() *TwoDTexture {
	return &TwoDTexture{
		TwoDTexture: opengl.NewTwoDTexture(),
	}
}

var _ graphics.TwoDTexture = (*TwoDTexture)(nil)

type TwoDTexture struct {
	*opengl.TwoDTexture
}

func newCubeTexture() *CubeTexture {
	return &CubeTexture{
		CubeTexture: opengl.NewCubeTexture(),
	}
}

var _ graphics.CubeTexture = (*CubeTexture)(nil)

type CubeTexture struct {
	*opengl.CubeTexture
}
