package resource

import (
	"fmt"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/game/asset"
	"github.com/mokiat/lacking/game/graphics"
)

const ModelTypeName = TypeName("model")

func InjectModel(target **Model) func(value interface{}) {
	return func(value interface{}) {
		*target = value.(*Model)
	}
}

type Model struct {
	Name          string
	Nodes         []*Node
	Armatures     []*Armature
	MeshInstances []*MeshInstance
	meshes        []*Mesh
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

func (m Model) FindMeshInstance(name string) (*MeshInstance, bool) {
	for _, instance := range m.MeshInstances {
		if instance.Name == name {
			return instance, true
		}
	}
	return nil, false
}

type Node struct {
	Name     string
	Matrix   sprec.Mat4
	Parent   *Node
	Children []*Node
}

func (n Node) AbsoluteMatrix() sprec.Mat4 {
	if n.Parent == nil {
		return n.Matrix
	}
	return sprec.Mat4Prod(n.Parent.AbsoluteMatrix(), n.Matrix)
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

type Armature struct {
	Joints []Joint
}

type Joint struct {
	Node              *Node
	InverseBindMatrix sprec.Mat4
}

type MeshInstance struct {
	Name           string
	Node           *Node
	MeshDefinition *Mesh
}

func NewModelOperator(delegate asset.Registry, gfxEngine *graphics.Engine) *ModelOperator {
	return &ModelOperator{
		delegate:  delegate,
		gfxEngine: gfxEngine,
	}
}

type ModelOperator struct {
	delegate  asset.Registry
	gfxEngine *graphics.Engine
}

func (o *ModelOperator) Allocate(registry *Registry, id string) (interface{}, error) {
	modelAsset := new(asset.Model)
	resource := o.delegate.ResourceByID(id)
	if resource == nil {
		return nil, fmt.Errorf("cannot find asset %q", id)
	}
	if err := resource.ReadContent(modelAsset); err != nil {
		return nil, fmt.Errorf("failed to open model asset %q: %w", id, err)
	}

	model := &Model{
		Name: id,
	}

	materials := make([]*Material, len(modelAsset.Materials))
	for i, assetMaterial := range modelAsset.Materials {
		material, err := AllocateMaterial(registry, o.gfxEngine, &assetMaterial)
		if err != nil {
			return nil, fmt.Errorf("failed to allocate material: %w", err)
		}
		materials[i] = material
	}

	meshes := make([]*Mesh, len(modelAsset.MeshDefinitions))
	for i, meshAsset := range modelAsset.MeshDefinitions {
		mesh, err := AllocateMesh(registry, o.gfxEngine, materials, &meshAsset)
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
		rotation := sprec.NewQuat(
			nodeAsset.Rotation[3],
			nodeAsset.Rotation[0],
			nodeAsset.Rotation[1],
			nodeAsset.Rotation[2],
		)
		nodes[i].Name = nodeAsset.Name
		nodes[i].Matrix = sprec.TRSMat4(
			sprec.ArrayToVec3(nodeAsset.Translation),
			rotation,
			sprec.ArrayToVec3(nodeAsset.Scale),
		)
	}
	model.Nodes = rootNodes

	model.Armatures = make([]*Armature, len(modelAsset.Armatures))
	for i, assetArmature := range modelAsset.Armatures {
		joints := make([]Joint, len(assetArmature.Joints))
		for j, assetJoint := range assetArmature.Joints {
			joints[j] = Joint{
				Node:              nodes[assetJoint.NodeIndex],
				InverseBindMatrix: sprec.ColumnMajorArrayToMat4(assetJoint.InverseBindMatrix),
			}
		}
		model.Armatures[i] = &Armature{
			Joints: joints,
		}
	}

	model.MeshInstances = make([]*MeshInstance, len(modelAsset.MeshInstances))
	for i, instance := range modelAsset.MeshInstances {
		model.MeshInstances[i] = &MeshInstance{
			Name:           instance.Name,
			Node:           nodes[instance.NodeIndex],
			MeshDefinition: model.meshes[instance.DefinitionIndex],
		}
	}

	return model, nil
}

func (o *ModelOperator) Release(registry *Registry, resource interface{}) error {
	model := resource.(*Model)

	for _, mesh := range model.meshes {
		if err := ReleaseMesh(registry, mesh); err != nil {
			return fmt.Errorf("failed to release mesh: %w", err)
		}
	}

	return nil
}
