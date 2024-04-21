package dsl

import (
	"fmt"

	"github.com/mokiat/lacking/game/asset/mdl"
)

// CreateModel creates a new model with the specified name and operations.
func CreateModel(name string, operations ...Operation) Provider[*mdl.Model] {
	if _, ok := modelProviders[name]; ok {
		panic(fmt.Sprintf("provider for model %q already exists", name))
	}

	modelProviders[name] = OnceProvider(FuncProvider(
		// get function
		func() (*mdl.Model, error) {
			var model mdl.Model
			model.SetName(name)
			for _, operation := range operations {
				if err := operation.Apply(&model); err != nil {
					return nil, err
				}
			}
			return &model, nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("create-model", name, operations)
		},
	))

	return modelProviders[name]
}
