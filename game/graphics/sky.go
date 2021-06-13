package graphics

import "github.com/mokiat/gomath/sprec"

// Sky represents the scene background.
type Sky interface {

	// BackgroundColor returns the color of the background.
	BackgroundColor() sprec.Vec3

	// SetBackgroundColor changes the color of the background.
	SetBackgroundColor(color sprec.Vec3)

	// Skybox returns the cube texture to be used as the background.
	// If one has not been set, this method returns nil.
	Skybox() CubeTexture

	// SetSkybox sets a cube texture to be used as the background.
	// If nil is specified, then a texture will not be used and instead
	// the background color will be drawn instead.
	SetSkybox(skybox CubeTexture)
}
