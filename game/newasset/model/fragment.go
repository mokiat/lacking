package model

import asset "github.com/mokiat/lacking/game/newasset"

type Fragment struct {
	name string

	nodes []*Node
}

func (f *Fragment) Name() string {
	return f.name
}

func (f *Fragment) SetName(name string) {
	f.name = name
}

func (f *Fragment) Nodes() []*Node {
	return f.nodes
}

func (f *Fragment) AddNode(node *Node) {
	f.nodes = append(f.nodes, node)
}

func (f *Fragment) ToAsset() (asset.Fragment, error) {
	nodes := make([]asset.Node, len(f.nodes))

	// FIXME: There is a problem with how we are handling the node index.
	// This does not work for hierarchies.
	nodeIndex := make(map[*Node]int)
	for i, node := range f.nodes {
		nodeAsset, err := node.ToAsset()
		if err != nil {
			return asset.Fragment{}, err
		}
		nodes[i] = nodeAsset
		nodeIndex[node] = i
	}

	pointLights := make([]asset.PointLight, 0)
	for _, node := range f.nodes {
		light, ok := node.Content().(*PointLight)
		if !ok {
			continue
		}
		lightAsset := light.ToAsset()
		lightAsset.NodeIndex = int32(nodeIndex[node])
		pointLights = append(pointLights, lightAsset)
	}

	return asset.Fragment{
		Nodes:       nodes,
		PointLights: pointLights,
	}, nil
}
