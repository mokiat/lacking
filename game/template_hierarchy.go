package game

import (
	"log/slog"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/hierarchy"
)

type HierarchyTemplate struct {
	Nodes []HierarchyNodeTemplate
}

type HierarchyNodeTemplate struct {
	ID       uint32
	ParentID uint32
	Name     string
	Position dprec.Vec3
	Rotation dprec.Quat
	Scale    dprec.Vec3
}

type HierarchyInstanceInfo struct {
	Template      *HierarchyTemplate
	Name          opt.T[string]
	Position      opt.T[dprec.Vec3]
	Rotation      opt.T[dprec.Quat]
	Scale         opt.T[dprec.Vec3]
	SubTreeNode   opt.T[string]
	AttachToScene opt.T[bool]
}

type HierarchyInstance struct {
	RootNode *hierarchy.Node
	Nodes    IdentifiableList[*hierarchy.Node]
}

func (s *Scene) InstantiateHierarchy(info HierarchyInstanceInfo) HierarchyInstance {
	template := info.Template

	nodes := make(map[uint32]*hierarchy.Node, len(template.Nodes))
	for _, nodeDef := range template.Nodes {
		node := hierarchy.NewNode()
		node.SetName(nodeDef.Name)
		node.SetPosition(nodeDef.Position)
		node.SetRotation(nodeDef.Rotation)
		node.SetScale(nodeDef.Scale)
		nodes[nodeDef.ID] = node
	}

	rootNode := hierarchy.NewNode()
	for _, nodeDef := range template.Nodes {
		var parent *hierarchy.Node
		if nodeDef.ParentID != UnspecifiedID {
			parent = nodes[nodeDef.ParentID]
		} else {
			parent = rootNode
		}
		parent.AppendChild(nodes[nodeDef.ID])
	}

	if info.SubTreeNode.Specified {
		subTreeNode := rootNode.FindNode(info.SubTreeNode.Value)
		if subTreeNode == nil {
			logger.Error("Root node not found", slog.String("name", info.SubTreeNode.Value))
			subTreeNode = hierarchy.NewNode()
		}
		subTreeNode.Detach()
		for id, node := range nodes {
			if !node.IsDescendantOf(subTreeNode) {
				node.Delete()
				delete(nodes, id)
			}
		}
		rootNode = subTreeNode
	}

	if info.AttachToScene.ValueOrDefault(false) {
		s.Root().AppendChild(rootNode)
	}

	if info.Name.Specified {
		rootNode.SetName(info.Name.Value)
	}
	if info.Position.Specified {
		rootNode.SetPosition(info.Position.Value)
	}
	if info.Rotation.Specified {
		rootNode.SetRotation(info.Rotation.Value)
	}
	if info.Scale.Specified {
		rootNode.SetScale(info.Scale.Value)
	}

	nodeList := make(IdentifiableList[*hierarchy.Node], 0, len(nodes))
	for id, node := range nodes {
		nodeList = append(nodeList, Identifiable[*hierarchy.Node]{
			ID:    id,
			Value: node,
		})
	}

	return HierarchyInstance{
		RootNode: rootNode,
		Nodes:    nodeList,
	}
}
