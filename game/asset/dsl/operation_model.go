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

// EditMaterial creates an operation that edits the material with the
// provided name in the target node holder.
func EditMaterial(name string, opts ...Operation) Operation {
	type nodeHolder interface {
		Nodes() []*mdl.Node
	}
	return FuncOperation(
		// apply function
		func(target any) error {
			nodeHolder, ok := target.(nodeHolder)
			if !ok {
				return fmt.Errorf("target %T is not a node holder", target)
			}
			material := findMaterial(nodeHolder.Nodes(), name)
			if material == nil {
				return fmt.Errorf("material %q not found", name)
			}
			for _, opt := range opts {
				if err := opt.Apply(material); err != nil {
					return fmt.Errorf("error applying material operation: %w", err)
				}
			}
			return nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("edit-material", name, opts)
		},
	)
}

// AddBlob creates an operation that adds a blob to the target model.
func AddBlob(blobProvider Provider[*mdl.Blob]) Operation {
	return FuncOperation(
		// apply function
		func(target any) error {
			blob, err := blobProvider.Get()
			if err != nil {
				return fmt.Errorf("error getting blob: %w", err)
			}

			targetModel, ok := target.(*mdl.Model)
			if !ok {
				return fmt.Errorf("target %T is not a model", target)
			}
			targetModel.AddBlob(blob)

			return nil
		},

		// digest function
		func() ([]byte, error) {
			return CreateDigest("add-blob", blobProvider)
		},
	)
}

func findMaterial(nodes []*mdl.Node, name string) *mdl.Material {
	nodes = flattenNodes(nodes)
	for _, node := range nodes {
		if mesh, ok := node.Target().(*mdl.Mesh); ok {
			materials := mesh.Definition().Materials()
			for _, material := range materials {
				if material.Name() == name {
					return material
				}
			}
		}
	}
	return nil
}

func flattenNodes(nodes []*mdl.Node) []*mdl.Node {
	var flattened []*mdl.Node
	visitNodes(nodes, func(node *mdl.Node) {
		flattened = append(flattened, node)
	})
	return flattened
}

func visitNodes(nodes []*mdl.Node, visitor func(*mdl.Node)) {
	for _, node := range nodes {
		visitor(node)
		visitNodes(node.Nodes(), visitor)
	}
}
