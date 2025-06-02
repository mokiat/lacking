package game

import (
	"log/slog"

	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/lacking/game/asset/dto"
	"github.com/mokiat/lacking/game/hierarchy"
)

// NodeTemplate represents a template for a node in the scene hierarchy.
type NodeTemplate struct {
	ParentID uint32
	Name     string
	Position dprec.Vec3
	Rotation dprec.Quat
	Scale    dprec.Vec3
}

// LoadNodeTemplate resolves a node template from the given asset data.
//
// This is a blocking operation and should be called from a worker thread.
func LoadNodeTemplate(loader *AssetLoader, assetNode dto.Node) (Identifiable[NodeTemplate], error) {
	return Identifiable[NodeTemplate]{
		ID: assetNode.ID,
		Value: NodeTemplate{
			ParentID: assetNode.ParentID,
			Name:     assetNode.Name,
			Position: assetNode.Translation,
			Rotation: assetNode.Rotation,
			Scale:    assetNode.Scale,
		},
	}, nil
}

// LoadNodeTemplates resolves a list of node templates from the given asset
// nodes.
//
// This is a blocking operation and should be called from a worker thread.
func LoadNodeTemplates(loader *AssetLoader, assetNodes []dto.Node) (IdentifiableList[NodeTemplate], error) {
	templates := make(IdentifiableList[NodeTemplate], len(assetNodes))
	for i, assetNode := range assetNodes {
		template, err := LoadNodeTemplate(loader, assetNode)
		if err != nil {
			return nil, err
		}
		templates[i] = template
	}
	return templates, nil
}

// UnloadNodeTemplate unloads a node template from the asset loader.
//
// This is a blocking operation and should be called from a worker thread.
func UnloadNodeTemplate(loader *AssetLoader, idNode Identifiable[NodeTemplate]) error {
	// At the time being this is a no-op.
	return nil
}

// UnloadNodeTemplates unloads a list of node templates from the asset loader.
//
// This is a blocking operation and should be called from a worker thread.
func UnloadNodeTemplates(loader *AssetLoader, idNodes IdentifiableList[NodeTemplate]) error {
	for _, idNode := range idNodes {
		if err := UnloadNodeTemplate(loader, idNode); err != nil {
			return err
		}
	}
	return nil
}

// HierarchyInfo contains information about a hierarchy to be instantiated in
// a scene.
type HierarchyInfo struct {
	NodeTemplates IdentifiableList[NodeTemplate]
	Name          opt.T[string]
	Position      opt.T[dprec.Vec3]
	Rotation      opt.T[dprec.Quat]
	Scale         opt.T[dprec.Vec3]
	SubTreeNode   opt.T[string]
	AttachToScene opt.T[bool]
}

// Hierarchy represents a scene hierarchy that has been instantiated from a
// HierarchyInfo.
type Hierarchy struct {
	RootNode *hierarchy.Node
	Nodes    IdentifiableList[*hierarchy.Node]
}

// InstantiateHierarchy instantiates a hierarchy in the given scene based on
// the provided info.
func InstantiateHierarchy(scene *Scene, info HierarchyInfo) *Hierarchy {
	nodes := make(map[uint32]*hierarchy.Node, len(info.NodeTemplates))
	for nodeID, nodeTemplate := range info.NodeTemplates.Iter() {
		node := hierarchy.NewNode()
		node.SetName(nodeTemplate.Name)
		node.SetPosition(nodeTemplate.Position)
		node.SetRotation(nodeTemplate.Rotation)
		node.SetScale(nodeTemplate.Scale)
		nodes[nodeID] = node
	}

	rootNode := hierarchy.NewNode()
	for nodeID, nodeTemplate := range info.NodeTemplates.Iter() {
		var parent *hierarchy.Node
		if nodeTemplate.ParentID != UnspecifiedID {
			parent = nodes[nodeTemplate.ParentID]
		} else {
			parent = rootNode
		}
		parent.AppendChild(nodes[nodeID])
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
		scene.Root().AppendChild(rootNode)
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

	return &Hierarchy{
		RootNode: rootNode,
		Nodes:    nodeList,
	}
}
