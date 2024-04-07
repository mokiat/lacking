package internal

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/util/blob"
)

type CameraUniform struct {
	ProjectionMatrix sprec.Mat4
	ViewMatrix       sprec.Mat4
	CameraMatrix     sprec.Mat4
	Viewport         sprec.Vec4
}

func (u CameraUniform) Std140Plot(plotter *blob.Plotter) {
	plotter.PlotSPMat4(u.ProjectionMatrix)
	plotter.PlotSPMat4(u.ViewMatrix)
	plotter.PlotSPMat4(u.CameraMatrix)
	plotter.PlotSPVec4(u.Viewport)
}

func (u CameraUniform) Std140Size() uint32 {
	return 64 + 64 + 64 + 16
}
