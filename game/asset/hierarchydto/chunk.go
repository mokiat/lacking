package hierarchydto

import (
	"github.com/google/uuid"
	"github.com/mokiat/gog"
)

var hierarchyChunkID = gog.Must(uuid.Parse("4e43db56-0910-4731-bd19-37107cbbac75"))

type HierarchyChunk struct {
	Nodes []Node
}

func (c HierarchyChunk) ChunkID() uuid.UUID {
	return hierarchyChunkID
}
