package graphics

import (
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/render"
)

func newTwoDTexture(texture render.Texture) *TwoDTexture {
	return &TwoDTexture{
		Texture: texture,
	}
}

var _ graphics.TwoDTexture = (*TwoDTexture)(nil)

type TwoDTexture struct {
	render.Texture
}

func (t *TwoDTexture) Delete() {
	t.Release()
}

func newCubeTexture(texture render.Texture) *CubeTexture {
	return &CubeTexture{
		Texture: texture,
	}
}

var _ graphics.CubeTexture = (*CubeTexture)(nil)

type CubeTexture struct {
	render.Texture
}

func (t *CubeTexture) Delete() {
	t.Release()
}
