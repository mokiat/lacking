package graphics

import "github.com/mokiat/gomath/sprec"

func newSky() *Sky {
	return &Sky{}
}

// Sky represents the Scene's background.
type Sky struct {
	backgroundColor sprec.Vec3
	skyboxTexture   *CubeTexture
}

// BackgroundColor returns the color of the background.
func (s *Sky) BackgroundColor() sprec.Vec3 {
	return s.backgroundColor
}

// SetBackgroundColor changes the color of the background.
func (s *Sky) SetBackgroundColor(color sprec.Vec3) {
	s.backgroundColor = color
}

// 	// Skybox returns the cube texture to be used as the background.
// 	// If one has not been set, this method returns nil.
func (s *Sky) Skybox() *CubeTexture {
	return s.skyboxTexture
}

// SetSkybox sets a cube texture to be used as the background.
// If nil is specified, then a texture will not be used and instead
// the background color will be drawn instead.
func (s *Sky) SetSkybox(skybox *CubeTexture) {
	s.skyboxTexture = skybox
}
