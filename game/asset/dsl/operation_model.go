package dsl

import (
	"fmt"

	"github.com/mokiat/lacking/game/asset/mdl"
)

// AppendModel creates an operation that appends the contents
// of the provided model to the target model.
func AppendModel(modelProvider Provider[*mdl.Model]) Operation {
	return FuncOperation(
		// apply function
		func(target any) error {
			model, err := modelProvider.Get()
			if err != nil {
				return fmt.Errorf("error getting model: %w", err)
			}

			targetModel, ok := target.(*mdl.Model)
			if !ok {
				return fmt.Errorf("target %T is not a model", target)
			}

			for _, node := range model.Nodes() {
				targetModel.AddNode(node)
			}
			for _, animation := range model.Animations() {
				targetModel.AddAnimation(animation)
			}
			return nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("append-model", modelProvider)
		},
	)
}

// ForceCollision creates an operation that forces the target
// model to have a collision mesh.
func ForceCollision() Operation {
	type collisionConfigurable interface {
		SetForceCollision(bool)
	}
	return FuncOperation(
		// apply function
		func(target any) error {
			configurable, ok := target.(collisionConfigurable)
			if !ok {
				return fmt.Errorf("target %T is not configurable for collision", target)
			}
			configurable.SetForceCollision(true)
			return nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("force-collision")
		},
	)
}
