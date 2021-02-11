package ui

import (
	"github.com/mokiat/lacking/opengl"
	"github.com/mokiat/lacking/ui"
)

type Image struct {
	texture *opengl.TwoDTexture
	size    ui.Size
}

func (i *Image) Size() ui.Size {
	return i.size
}
