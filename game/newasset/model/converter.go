package model

import asset "github.com/mokiat/lacking/game/newasset"

func NewConverter(scene *Scene) *Converter {
	return &Converter{
		scene: scene,
	}
}

type Converter struct {
	scene *Scene
}

func (c *Converter) Convert() (asset.Scene, error) {
	return c.convertScene(c.scene)
}

func (c *Converter) convertScene(s *Scene) (asset.Scene, error) {
	var (
		assetNodes       []asset.Node
		assetPointLights []asset.PointLight
		assetSpotLights  []asset.SpotLight
	)

	nodes := s.FlattenNodes()

	nodeIndex := make(map[Node]uint32)

	for i, node := range nodes {
		nodeIndex[node] = uint32(i)

		parentIndex := asset.UnspecifiedNodeIndex
		if pIndex, ok := nodeIndex[node.Parent()]; ok {
			parentIndex = int32(pIndex)
		}

		assetNodes = append(assetNodes, asset.Node{
			Name:        node.Name(),
			ParentIndex: parentIndex,
			Translation: node.Translation(),
			Rotation:    node.Rotation(),
			Scale:       node.Scale(),
			Mask:        asset.NodeMaskNone,
		})

		switch essence := node.(type) {
		case *PointLight:
			assetPointLights = append(assetPointLights, asset.PointLight{
				NodeIndex:    uint32(i),
				EmitColor:    essence.EmitColor(),
				EmitDistance: essence.EmitDistance(),
				CastShadow:   essence.CastShadow(),
			})
		case *SpotLight:
			assetSpotLights = append(assetSpotLights, asset.SpotLight{
				NodeIndex:      uint32(i),
				EmitColor:      essence.EmitColor(),
				EmitDistance:   essence.EmitDistance(),
				EmitAngleOuter: essence.EmitAngleOuter(),
				EmitAngleInner: essence.EmitAngleInner(),
				CastShadow:     essence.CastShadow(),
			})
		}
	}

	return asset.Scene{
		Nodes:       assetNodes,
		PointLights: assetPointLights,
		SpotLights:  assetSpotLights,
	}, nil
}
