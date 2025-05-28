package dto

import "github.com/mokiat/gomath/sprec"

const (
	FoVModeFoVModeHorizontalPlus FoVMode = iota

	FoVModeFoVModeVertialMinus
)

var CameraChunkID = "lacking:camera"

type CameraChunkHolder struct {
	Camera *CameraChunk `chunk:"lacking:camera"`
}

type CameraChunk struct {
	// Cameras is the collection of cameras that are part of the scene.
	Cameras []Camera
}

// FoVMode determines how the camera field of view is calculated
// in the horizontal and vertical directions.
type FoVMode uint8

// Camera represents a camera that is part of a model.
type Camera struct {

	// ID is the unique identifier of the camera within the file.
	ID uint32

	// NodeIndex is the index of the node that is used by this camera.
	NodeIndex uint32

	// FoVMode determines how the camera field of view is calculated
	// in the horizontal and vertical directions.
	FoVMode FoVMode

	// FoVAngle is the field of view angle of the camera.
	FoVAngle sprec.Angle

	// Near is the distance to the near clipping plane.
	Near float32

	// Far is the distance to the far clipping plane.
	Far float32

	// Exposure is the exposure value of the camera.
	Exposure float32
}
