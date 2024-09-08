package internal

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/util/blob"
)

type LightUniform struct {
	ShadowMatrices [8]sprec.Mat4
	ModelMatrix    sprec.Mat4

	ShadowCascades [8]sprec.Vec2

	Color     sprec.Vec3
	Intensity float32

	Range      float32
	OuterAngle float32
	InnerAngle float32
}

func (u LightUniform) Std140Plot(plotter *blob.Plotter) {
	// 8 x mat4
	for _, matrix := range u.ShadowMatrices {
		plotter.PlotSPMat4(matrix)
	}
	// mat4
	plotter.PlotSPMat4(u.ModelMatrix)

	// 8 x vec4
	for _, cascade := range u.ShadowCascades {
		plotter.PlotSPVec2(cascade)
		plotter.Skip(2 * 4)
	}

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
	return 8*64 + 64 + 8*4*4 + 4*4 + 4*4
}
