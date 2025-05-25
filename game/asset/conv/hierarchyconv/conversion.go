package hierarchyconv

import (
	"iter"

	"github.com/mokiat/lacking/game/asset/dto/hierarchydto"
	"github.com/mokiat/lacking/game/asset/mdl"
)

type Source interface {
	NodesIter() iter.Seq2[int, *mdl.Node]
}

func CreateHierarchyChunk(src Source) *hierarchydto.HierarchyChunk {
	convertedNodes := make(map[*mdl.Node]uint32)
	for i, node := range src.NodesIter() {
		convertedNodes[node] = uint32(i)
	}
	dtoNodes := make([]hierarchydto.Node, len(convertedNodes))
	for i, node := range src.NodesIter() {
		parentIndex := hierarchydto.UnspecifiedNodeIndex
		if pIndex, ok := convertedNodes[node.Parent()]; ok {
			parentIndex = int32(pIndex)
		}
		dtoNodes[i] = hierarchydto.Node{
			Name:        node.Name(),
			ParentIndex: parentIndex,
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
