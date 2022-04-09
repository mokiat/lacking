package graphics

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/graphics"
)

func newSky() *Sky {
	return &Sky{}
}

var _ graphics.Sky = (*Sky)(nil)

type Sky struct {
	backgroundColor sprec.Vec3
	skyboxTexture   *CubeTexture
}

func (s *Sky) BackgroundColor() sprec.Vec3 {
	return s.backgroundColor
}

func (s *Sky) SetBackgroundColor(color sprec.Vec3) {
	s.backgroundColor = color
}

func (s *Sky) Skybox() graphics.CubeTexture {
	return s.skyboxTexture
}

func (s *Sky) SetSkybox(skybox graphics.CubeTexture) {
	if skybox == nil {
		s.skyboxTexture = nil
	} else {
		s.skyboxTexture = skybox.(*CubeTexture)
	}
}
