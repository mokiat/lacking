package ui

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/util/blob"
)

type cameraUniform struct {
	Projection sprec.Mat4
}

func (u cameraUniform) Std140Plot(plotter *blob.Plotter) {
	plotter.PlotSPMat4(u.Projection)
}

func (u cameraUniform) Std140Size() uint32 {
	return 64
}

type modelUniform struct {
	Transform     sprec.Mat4
	ClipTransform sprec.Mat4
}

func (u modelUniform) Std140Plot(plotter *blob.Plotter) {
	plotter.PlotSPMat4(u.Transform)
	plotter.PlotSPMat4(u.ClipTransform)
}

func (u modelUniform) Std140Size() uint32 {
	return 64 + 64
}

type materialUniform struct {
	TextureTransform sprec.Mat4
	Color            sprec.Vec4
}

func (u materialUniform) Std140Plot(plotter *blob.Plotter) {
	plotter.PlotSPMat4(u.TextureTransform)
	plotter.PlotSPVec4(u.Color)
}

func (u materialUniform) Std140Size() uint32 {
	return 64 + 16
}
