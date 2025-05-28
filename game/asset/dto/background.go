package dto

const BackgroundChunkID = "lacking:background"

type BackgroundChunkHolder struct {
	BackgroundChunk *BackgroundChunk `chunk:"lacking:background"`
}

type BackgroundChunk struct {
	// Skies is the collection of skies that are part of the scene.
	Skies []Sky
}

// Sky represents the background of the scene.
type Sky struct {

	// ID is the unique identifier of the sky within the file.
	ID uint32

	// NodeID is the ID of the node that the sky is attached to.
	NodeID uint32

	// MaterialID is the ID of the material that will be used to render the
	// sky.
	MaterialID uint32
}
