package internal

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/util/blob"
)

type LightPropertiesUniform struct {
	Color     sprec.Vec3
	Intensity float32

	Range      float32
	OuterAngle float32
	InnerAngle float32
}

func (u LightPropertiesUniform) Plot(plotter *blob.Plotter, padding int) {
	// vec4
	plotter.PlotSPVec3(u.Color)
	plotter.PlotFloat32(u.Intensity)

	// vec4
	plotter.PlotFloat32(u.Range)
	plotter.PlotFloat32(u.OuterAngle)
	plotter.PlotFloat32(u.InnerAngle)
	plotter.Skip(4)

	plotter.Skip(padding)
}

func (u LightPropertiesUniform) Std140Size() int {
	return 16 + 16
}
