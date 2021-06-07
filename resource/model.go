package resource

import (
	"fmt"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/async"
	"github.com/mokiat/lacking/data/asset"
	"github.com/mokiat/lacking/game/graphics"
)

const ModelTypeName = TypeName("model")

func InjectModel(target **Model) func(value interface{}) {
	return func(value interface{}) {
		*target = value.(*Model)
	}
}

type Model struct {
	Name   string
	Nodes  []*Node
	meshes []*Mesh
}

func (m Model) FindNode(name string) (*Node, bool) {
	for _, node := range m.Nodes {
		if node.Name == name {
			return node, true
		}
		if child, found := node.FindNode(name); found {
			return child, true
		}
	}
	return nil, false
}

type Node struct {
	Name     string
	Matrix   sprec.Mat4
	Mesh     *Mesh
	Parent   *Node
	Children []*Node
}

func (n Node) FindNode(name string) (*Node, bool) {
	for _, node := range n.Children {
		if node.Name == name {
			return node, true
		}
		if child, found := node.FindNode(name); found {
			return child, true
		}
	}
	return nil, false
}

func NewModelOperator(locator Locator, gfxEngine graphics.Engine, gfxWorker *async.Worker) *ModelOperator {
	return &ModelOperator{
		locator:   locator,
		gfxEngine: gfxEngine,
		gfxWorker: gfxWorker,
	}
}

type ModelOperator struct {
	locator   Locator
	gfxEngine graphics.Engine
	gfxWorker *async.Worker
}

func (o *ModelOperator) Allocate(registry *Registry, name string) (interface{}, error) {
	in, err := o.locator.Open("assets", "models", name)
	if err != nil {
		return nil, fmt.Errorf("failed to open model asset %q: %w", name, err)
	}
	defer in.Close()

	modelAsset := new(asset.Model)
	if err := asset.DecodeModel(in, modelAsset); err != nil {
		return nil, fmt.Errorf("failed to decode model asset %q: %w", name, err)
	}

	model := &Model{
		Name: name,
	}

	meshes := make([]*Mesh, len(modelAsset.Meshes))
	for i, meshAsset := range modelAsset.Meshes {
		mesh, err := AllocateMesh(registry, meshAsset.Name, o.gfxWorker, o.gfxEngine, &meshAsset)
		if err != nil {
			return nil, fmt.Errorf("failed to allocate mesh: %w", err)
		}
		meshes[i] = mesh
	}
	model.meshes = meshes

	nodes := make([]*Node, len(modelAsset.Nodes))
	for i := range nodes {
		nodes[i] = &Node{}
	}
	rootNodes := make([]*Node, 0)
	for i, nodeAsset := range modelAsset.Nodes {
		if nodeAsset.ParentIndex != -1 {
			nodes[i].Parent = nodes[nodeAsset.ParentIndex]
			nodes[nodeAsset.ParentIndex].Children = append(nodes[nodeAsset.ParentIndex].Children, nodes[i])
		} else {
			rootNodes = append(rootNodes, nodes[i])
		}
		nodes[i].Name = nodeAsset.Name
		nodes[i].Matrix = sprec.ColumnMajorArrayMat4(nodeAsset.Matrix)
		nodes[i].Mesh = meshes[nodeAsset.MeshIndex]
	}
	model.Nodes = rootNodes

	return model, nil
}

func (o *ModelOperator) Release(registry *Registry, resource interface{}) error {
	model := resource.(*Model)

	for _, mesh := range model.meshes {
		if err := ReleaseMesh(registry, o.gfxWorker, mesh); err != nil {
			return fmt.Errorf("failed to release mesh: %w", err)
		}
	}

	return nil
}
