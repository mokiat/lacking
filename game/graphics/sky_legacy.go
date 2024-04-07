package graphics

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/render"
)

func newLegacySky() *OldSky {
	return &OldSky{}
}

// OldSky represents the Scene's background.
type OldSky struct {
	backgroundColor sprec.Vec3
	skyboxTexture   render.Texture
}

// BackgroundColor returns the color of the background.
func (s *OldSky) BackgroundColor() sprec.Vec3 {
	return s.backgroundColor
}

// SetBackgroundColor changes the color of the background.
func (s *OldSky) SetBackgroundColor(color sprec.Vec3) {
	s.backgroundColor = color
}

// // Skybox returns the cube texture to be used as the background.
// // If one has not been set, this method returns nil.
func (s *OldSky) Skybox() render.Texture {
	return s.skyboxTexture
}

// SetSkybox sets a cube texture to be used as the background.
// If nil is specified, then a texture will not be used and instead
// the background color will be drawn instead.
func (s *OldSky) SetSkybox(skybox render.Texture) {
	s.skyboxTexture = skybox
}
