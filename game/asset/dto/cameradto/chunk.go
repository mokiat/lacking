package cameradto

var CameraChunkID = "lacking:camera"

type CameraChunkHolder struct {
	Camera *CameraChunk `chunk:"lacking:camera"`
}

type CameraChunk struct {
	// Cameras is the collection of cameras that are part of the scene.
	Cameras []Camera
}
