package dsl

import (
	"fmt"

	"github.com/mokiat/lacking/game/newasset/mdl"
)

func CreateModel(name string, operations ...Operation) Provider[*mdl.Model] {
	get := func() (*mdl.Model, error) {
		var model mdl.Model
		model.SetName(name)
		for _, operation := range operations {
			if err := operation.Apply(&model); err != nil {
				return nil, err
			}
		}
		return &model, nil
	}

	digest := func() ([]byte, error) {
		return CreateDigest("create-model", name, operations)
	}

	provider := OnceProvider(FuncProvider(get, digest))
	if _, ok := modelProviders[name]; ok {
		panic(fmt.Sprintf("provider for model %q already exists", name))
	}
	modelProviders[name] = provider
	return provider
}
