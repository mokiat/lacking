package hierarchyconv

import (
	"github.com/mokiat/lacking/game/asset/dto/hierarchydto"
	"github.com/mokiat/lacking/game/asset/mdl"
)

type Source interface {
	AllNodes() []*mdl.Node
}

func CreateHierarchyChunk(src Source) *hierarchydto.HierarchyChunk {
	allNodes := src.AllNodes()
	dtoNodes := make([]hierarchydto.Node, len(allNodes))
	for i, node := range allNodes {
		parentID := hierarchydto.UnspecifiedNodeID
		if parent := node.Parent(); parent != nil {
			parentID = parent.ID()
		}
		dtoNodes[i] = hierarchydto.Node{
			ID:          node.ID(),
			ParentID:    parentID,
			Name:        node.Name(),
			Translation: node.Translation(),
			Rotation:    node.Rotation(),
			Scale:       node.Scale(),
			Mask:        hierarchydto.NodeMaskNone,
		}
	}
	return &hierarchydto.HierarchyChunk{
		Nodes: dtoNodes,
	}
}
