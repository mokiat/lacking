package hierarchydto

type HierarchyChunkHolder struct {
	HierarchyChunk *HierarchyChunk `chunk:"lacking:hierarchy"`
}

type HierarchyChunk struct {
	Nodes []Node
}
