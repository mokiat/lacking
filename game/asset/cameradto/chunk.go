package cameradto

import (
	"github.com/google/uuid"
	"github.com/mokiat/gog"
)

var cameraChunkID = gog.Must(uuid.Parse("bdd1aea4-d20c-4c13-a626-7308818b79f7"))

type CameraChunk struct {
	// Cameras is the collection of cameras that are part of the scene.
	Cameras []Camera
}

func (c CameraChunk) ChunkID() uuid.UUID {
	return cameraChunkID
}
