package internal

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/util/blob"
)

type LightUniform struct {
	ShadowMatrixNear sprec.Mat4
	ShadowMatrixMid  sprec.Mat4
	ShadowMatrixFar  sprec.Mat4
	ModelMatrix      sprec.Mat4

	ShadowCascades sprec.Vec4

	Color     sprec.Vec3
	Intensity float32

	Range      float32
	OuterAngle float32
	InnerAngle float32
}

func (u LightUniform) Std140Plot(plotter *blob.Plotter) {
	// mat4
	plotter.PlotSPMat4(u.ShadowMatrixNear)
	// mat4
	plotter.PlotSPMat4(u.ShadowMatrixMid)
	// mat4
	plotter.PlotSPMat4(u.ShadowMatrixFar)
	// mat4
	plotter.PlotSPMat4(u.ModelMatrix)

	// vec4
	plotter.PlotSPVec4(u.ShadowCascades)

	// vec4
	plotter.PlotSPVec3(u.Color)
	plotter.PlotFloat32(u.Intensity)

	// vec4
	plotter.PlotFloat32(u.Range)
	plotter.PlotFloat32(u.OuterAngle)
	plotter.PlotFloat32(u.InnerAngle)
	plotter.Skip(4)
}

func (u LightUniform) Std140Size() uint32 {
	return 4*64 + 3*16
}
