package dsl

import "github.com/mokiat/lacking/game/newasset/model"

func CreatePointLight(operations ...Operation) Provider[*model.PointLight] {
	get := func() (*model.PointLight, error) {
		var pointLight model.PointLight
		for _, operation := range operations {
			if err := operation.Apply(&pointLight); err != nil {
				return nil, err
			}
		}
		return &pointLight, nil
	}

	digest := func() ([]byte, error) {
		return digestItems("point-light", operations)
	}

	return OnceProvider(FuncProvider(get, digest))
}
