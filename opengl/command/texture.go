package command

import "github.com/mokiat/lacking/opengl"

type BindTexture struct {
	Name    string
	Texture *opengl.Texture
}
