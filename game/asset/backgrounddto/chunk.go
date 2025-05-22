package backgrounddto

import (
	"github.com/google/uuid"
	"github.com/mokiat/gog"
)

var backgroundChunkID = gog.Must(uuid.Parse("316e7b52-1d28-4b5b-b6f6-0f45ac7b1ac9"))

type BackgroundChunk struct {
	// Skies is the collection of skies that are part of the scene.
	Skies []Sky
}

func (c BackgroundChunk) ChunkID() uuid.UUID {
	return backgroundChunkID
}
