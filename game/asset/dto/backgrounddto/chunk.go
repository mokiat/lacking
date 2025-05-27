package backgrounddto

var BackgroundChunkID = "lacking:background"

type BackgroundChunkHolder struct {
	BackgroundChunk *BackgroundChunk `chunk:"lacking:background"`
}

type BackgroundChunk struct {
	// Skies is the collection of skies that are part of the scene.
	Skies []Sky
}
