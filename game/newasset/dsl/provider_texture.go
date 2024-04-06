package dsl

import (
	"github.com/mokiat/lacking/game/newasset/mdl"
)

func CreateTexture(operations ...Operation) Provider[*mdl.Texture] {
	get := func() (*mdl.Texture, error) {
		var texture mdl.Texture
		for _, op := range operations {
			if err := op.Apply(&texture); err != nil {
				return nil, err
			}
		}
		return &texture, nil
	}

	digest := func() ([]byte, error) {
		return digestItems("texture", operations)
	}

	return OnceProvider(FuncProvider(get, digest))
}
