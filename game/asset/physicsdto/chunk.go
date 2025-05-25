package physicsdto

type PhysicsChunkHolder struct {
	PhysicsChunk *PhysicsChunk `chunk:"lacking:physics"`
}

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
