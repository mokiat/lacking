package lightingdto

var LightingChunkID = "lacking:lighting"

type LightingChunkHolder struct {
	LightingChunk *LightingChunk `chunk:"lacking:lighting"`
}

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
