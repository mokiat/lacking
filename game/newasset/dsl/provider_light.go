package dsl

import (
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/newasset/model"
)

func CreatePointLight(name string, operations ...Operation) Provider[model.Node] {
	get := func() (model.Node, error) {
		var pointLight model.PointLight
		pointLight.SetEmitColor(dprec.NewVec3(1.0, 1.0, 1.0))
		pointLight.SetEmitDistance(10.0)
		for _, operation := range operations {
			if err := operation.Apply(&pointLight); err != nil {
				return nil, err
			}
		}
		return &pointLight, nil
	}

	digest := func() ([]byte, error) {
		return digestItems("point-light", name, operations)
	}

	return OnceProvider(FuncProvider(get, digest))
}
