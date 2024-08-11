package internal

import (
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/util/blob"
)

type BloomBlurUniform struct {
	Horizontal float32
}

func (u BloomBlurUniform) Std140Plot(plotter *blob.Plotter) {
	plotter.PlotFloat32(u.Horizontal)
}

func (u BloomBlurUniform) Std140Size() uint32 {
	return render.SizeF32
}
