package animationdto

import (
	"github.com/google/uuid"
	"github.com/mokiat/gog"
)

var animationChunkID = gog.Must(uuid.Parse("1336d701-72ba-4043-93ff-733033dfe838"))

type AnimationChunk struct {
	// Animations is the collection of animations that are part of the scene.
	Animations []Animation
}

func (c AnimationChunk) ChunkID() uuid.UUID {
	return animationChunkID
}
