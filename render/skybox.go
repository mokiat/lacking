package render

import "github.com/mokiat/lacking/graphics"

type Skybox struct {
	SkyboxTexture            *graphics.CubeTexture
	AmbientReflectionTexture *graphics.CubeTexture
	AmbientRefractionTexture *graphics.CubeTexture
}
