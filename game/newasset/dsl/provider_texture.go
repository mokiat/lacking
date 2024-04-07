package dsl

import (
	"github.com/mokiat/lacking/game/newasset/mdl"
)

func CreateCubeTexture(size int, format mdl.TextureFormat, operations ...Operation) Provider[*mdl.Texture] {
	get := func() (*mdl.Texture, error) {
		var texture mdl.Texture
		texture.SetKind(mdl.TextureKindCube)
		texture.Resize(size, size)
		texture.SetFormat(format)
		for _, op := range operations {
			if err := op.Apply(&texture); err != nil {
				return nil, err
			}
		}
		return &texture, nil
	}

	digest := func() ([]byte, error) {
		return digestItems("cube-texture", operations)
	}

	return OnceProvider(FuncProvider(get, digest))
}
