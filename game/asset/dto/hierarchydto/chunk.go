package hierarchydto

var HierarchyChunkID = "lacking:hierarchy"

type HierarchyChunkHolder struct {
	HierarchyChunk *HierarchyChunk `chunk:"lacking:hierarchy"`
}

type HierarchyChunk struct {
	Nodes []Node
}
