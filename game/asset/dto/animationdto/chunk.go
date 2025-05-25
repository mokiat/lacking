package animationdto

type AnimationChunkHolder struct {
	AnimationChunk *AnimationChunk `chunk:"lacking:animation"`
}

type AnimationChunk struct {
	// Animations is the collection of animations that are part of the scene.
	Animations []Animation
}
