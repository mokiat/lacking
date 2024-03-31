package dsl

import (
	"fmt"

	"github.com/mokiat/lacking/game/newasset/model"
)

func CreateScene(name string, operations ...Operation) Provider[*model.Scene] {
	get := func() (*model.Scene, error) {
		scene := &model.Scene{}
		scene.SetName(name)
		for _, operation := range operations {
			if err := operation.Apply(scene); err != nil {
				return nil, err
			}
		}
		return scene, nil
	}

	digest := func() ([]byte, error) {
		return digestItems("scene", name, operations)
	}

	provider := OnceProvider(FuncProvider(get, digest))
	if _, ok := sceneProviders[name]; ok {
		panic(fmt.Sprintf("provider for scene %q already exists", name))
	}
	sceneProviders[name] = provider
	return provider
}
