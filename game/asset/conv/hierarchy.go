package conv

import (
	"github.com/mokiat/gog/ds"
	"github.com/mokiat/lacking/game/asset/dto"
	"github.com/mokiat/lacking/game/asset/mdl"
	"github.com/mokiat/lacking/storage/chunked"
)

type HierarchySource interface {
	AllNodes() []*mdl.Node
}

func NewHierarchyConverter() *HierarchyConverter {
	return &HierarchyConverter{}
}

type HierarchyConverter struct{}

func (c *HierarchyConverter) Convert(target *ds.List[chunked.Chunk], asset any) error {
	src, ok := asset.(HierarchySource)
	if !ok {
		return nil
	}
	chunk, err := c.CreateHierarchyChunk(src)
	if err != nil {
		return err
	}
	target.Add(chunked.FromValue(dto.HierarchyChunkID, chunk))
	return nil
}

func (c *HierarchyConverter) CreateHierarchyChunk(src HierarchySource) (*dto.HierarchyChunk, error) {
	allNodes := src.AllNodes()
	dtoNodes := make([]dto.Node, len(allNodes))
	for i, node := range allNodes {
		parentID := dto.UnspecifiedNodeID
		if parent := node.Parent(); parent != nil {
			parentID = parent.ID()
		}
		dtoNodes[i] = dto.Node{
			ID:          node.ID(),
			ParentID:    parentID,
			Name:        node.Name(),
			Translation: node.Translation(),
			Rotation:    node.Rotation(),
			Scale:       node.Scale(),
			Mask:        dto.NodeMaskNone,
		}
	}
	return &dto.HierarchyChunk{
		Nodes: dtoNodes,
	}, nil
}
