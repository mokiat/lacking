package physicsdto

import (
	"github.com/google/uuid"
	"github.com/mokiat/gog"
)

var physicsChunkID = gog.Must(uuid.Parse("97f4233e-ae36-41fb-866e-de89ad2a0ee8"))

type PhysicsChunk struct {

	// BodyMaterials is the collection of body materials that are part of the
	// scene.
	BodyMaterials []BodyMaterial

	// BodyDefinitions is the collection of body definitions that are part of
	// the scene.
	BodyDefinitions []BodyDefinition

	// Bodies is the collection of body instances that are part of the scene.
	Bodies []Body
}

func (c PhysicsChunk) ChunkID() uuid.UUID {
	return physicsChunkID
}
