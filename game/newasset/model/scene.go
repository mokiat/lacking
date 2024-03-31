package model

import asset "github.com/mokiat/lacking/game/newasset"

type Scene struct {
	name string

	nodes []*Node
}

func (f *Scene) Name() string {
	return f.name
}

func (f *Scene) SetName(name string) {
	f.name = name
}

func (f *Scene) Nodes() []*Node {
	return f.nodes
}

func (f *Scene) AddNode(node *Node) {
	f.nodes = append(f.nodes, node)
}

func (f *Scene) ToAsset() (asset.Scene, error) {
	nodes := make([]asset.Node, len(f.nodes))

	// FIXME: There is a problem with how we are handling the node index.
	// This does not work for hierarchies.
	nodeIndex := make(map[*Node]uint32)
	for i, node := range f.nodes {
		nodeAsset, err := node.ToAsset()
		if err != nil {
			return asset.Scene{}, err
		}
		nodes[i] = nodeAsset
		nodeIndex[node] = uint32(i)
	}

	pointLights := make([]asset.PointLight, 0)
	for _, node := range f.nodes {
		light, ok := node.Content().(*PointLight)
		if !ok {
			continue
		}
		lightAsset := light.ToAsset()
		lightAsset.NodeIndex = nodeIndex[node]
		pointLights = append(pointLights, lightAsset)
	}

	return asset.Scene{
		Nodes:       nodes,
		PointLights: pointLights,
	}, nil
}
