package graphics

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/framework/opengl"
	"github.com/mokiat/lacking/framework/opengl/game/graphics/internal"
)

type Material struct {
	backfaceCulling bool
	alphaTesting    bool
	alphaBlending   bool
	alphaThreshold  float32

	geometryPresentation *internal.GeometryPresentation
	shadowPresentation   *internal.ShadowPresentation

	twoDTextures []*opengl.TwoDTexture
	cubeTextures []*opengl.CubeTexture
	vectors      []sprec.Vec4
}

func (m *Material) Delete() {
	if m.geometryPresentation != nil {
		m.geometryPresentation.Delete()
	}
	if m.shadowPresentation != nil {
		m.shadowPresentation.Delete()
	}
	for _, texture := range m.twoDTextures {
		texture.Release()
	}
	for _, texture := range m.cubeTextures {
		texture.Release()
	}
}
