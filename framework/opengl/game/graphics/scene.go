package graphics

import "github.com/mokiat/lacking/game/graphics"

func newScene(renderer *Renderer) *Scene {
	return &Scene{
		renderer: renderer,

		sky: newSky(),
	}
}

var _ graphics.Scene = (*Scene)(nil)

type Scene struct {
	renderer *Renderer

	sky *Sky
}

func (s *Scene) Sky() graphics.Sky {
	return s.sky
}

func (s *Scene) CreateCamera() graphics.Camera {
	return newCamera(s)
}

func (s *Scene) CreateLight() graphics.Light {
	return nil
}

func (s *Scene) Render(viewport graphics.Viewport, camera graphics.Camera) {
	gfxCamera := camera.(*Camera)
	s.renderer.Render(viewport, s, gfxCamera)
}

func (s *Scene) Delete() {
}
