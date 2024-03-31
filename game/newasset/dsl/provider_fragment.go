package dsl

import (
	"fmt"

	"github.com/mokiat/lacking/game/newasset/model"
)

func CreateFragment(name string, operations ...Operation) Provider[*model.Scene] {
	get := func() (*model.Scene, error) {
		fragment := &model.Scene{}
		fragment.SetName(name)
		for _, operation := range operations {
			if err := operation.Apply(fragment); err != nil {
				return nil, err
			}
		}
		return fragment, nil
	}

	digest := func() ([]byte, error) {
		return digestItems("fragment", name, operations)
	}

	provider := OnceProvider(FuncProvider(get, digest))
	if _, ok := fragments[name]; ok {
		panic(fmt.Sprintf("fragment %q already exists", name))
	}
	fragments[name] = provider
	return provider
}
