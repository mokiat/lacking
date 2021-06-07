package graphics

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/framework/opengl"
)

type Material struct {
	backfaceCulling bool
	alphaTesting    bool
	alphaBlending   bool
	alphaThreshold  float32

	geometryProgram *opengl.Program
	shadowProgram   *opengl.Program // TODO: One that doesn't draw pixels

	twoDTextures []*opengl.TwoDTexture
	cubeTextures []*opengl.CubeTexture
	vectors      []sprec.Vec4
}

func (m *Material) Delete() {
	if m.geometryProgram != nil {
		m.geometryProgram.Release()
	}
	if m.shadowProgram != nil {
		m.shadowProgram.Release()
	}
	for _, texture := range m.twoDTextures {
		texture.Release()
	}
	for _, texture := range m.cubeTextures {
		texture.Release()
	}
}
