package internal

import (
	"github.com/mokiat/lacking/render"
	"github.com/mokiat/lacking/util/blob"
)

type BloomBlurUniform struct {
	Horizontal float32
	Steps      float32
}

func (u BloomBlurUniform) Std140Plot(plotter *blob.Plotter) {
	plotter.PlotFloat32(u.Horizontal)
	plotter.PlotFloat32(u.Steps)
}

func (u BloomBlurUniform) Std140Size() int {
	return render.SizeF16 + render.SizeF16
}
