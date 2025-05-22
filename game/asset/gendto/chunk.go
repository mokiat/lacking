package gendto

import (
	"github.com/google/uuid"
	"github.com/mokiat/gog"
)

var genChunkID = gog.Must(uuid.Parse("a550a8a8-c2e9-4119-9fbb-9ad43b904cf9"))

type GenChunk struct {
	Digest string
}

func (c GenChunk) ChunkID() uuid.UUID {
	return genChunkID
}
