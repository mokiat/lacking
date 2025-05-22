package lightingdto

import "github.com/google/uuid"

var lightingChunkID = uuid.Must(uuid.Parse("fb797d1c-9cc8-42e2-941c-a776d9c561de"))

type LightingChunk struct {
	// AmbientLights is the collection of ambient lights that are part of the
	// scene.
	AmbientLights []AmbientLight

	// PointLights is the collection of point lights that are part of the scene.
	PointLights []PointLight

	// SpotLights is the collection of spot lights that are part of the scene.
	SpotLights []SpotLight

	// DirectionalLights is the collection of directional lights that are part
	// of the scene.
	DirectionalLights []DirectionalLight
}

func (c LightingChunk) ChunkID() uuid.UUID {
	return lightingChunkID
}
