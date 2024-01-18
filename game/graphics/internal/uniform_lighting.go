package internal

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/util/blob"
)

type LightUniform struct {
	ProjectionMatrix sprec.Mat4
	ViewMatrix       sprec.Mat4
	LightMatrix      sprec.Mat4
}

func (u LightUniform) Std140Plot(plotter *blob.Plotter) {
	plotter.PlotSPMat4(u.ProjectionMatrix)
	plotter.PlotSPMat4(u.ViewMatrix)
	plotter.PlotSPMat4(u.LightMatrix)
}

func (u LightUniform) Std140Size() int {
	return 64 + 64 + 64
}

type LightPropertiesUniform struct {
	Color     sprec.Vec3
	Intensity float32

	Range      float32
	OuterAngle float32
	InnerAngle float32
}

func (u LightPropertiesUniform) Std140Plot(plotter *blob.Plotter) {
	// vec4
	plotter.PlotSPVec3(u.Color)
	plotter.PlotFloat32(u.Intensity)

	// vec4
	plotter.PlotFloat32(u.Range)
	plotter.PlotFloat32(u.OuterAngle)
	plotter.PlotFloat32(u.InnerAngle)
	plotter.Skip(4)
}

func (u LightPropertiesUniform) Std140Size() int {
	return 16 + 16
}
