package graphics

// Engine represents an entrypoint to 3D graphics rendering.
type Engine interface {

	// Create initializes this 3D engine.
	Create()

	// CreateScene creates a new 3D Scene. Entities managed
	// within a given scene are isolated within that scene.
	CreateScene() Scene

	// CreateTwoDTexture creates a new TwoDTexture using the
	// specified definition.
	CreateTwoDTexture(definition TwoDTextureDefinition) TwoDTexture

	// CreateCubeTexture creates a new CubeTexture using the
	// specified definition.
	CreateCubeTexture(definition CubeTextureDefinition) CubeTexture

	// CreateMeshTemplate creates a new MeshTemplate using the specified
	// definition.
	CreateMeshTemplate(definition MeshTemplateDefinition) MeshTemplate

	// CreatePBRMaterial creates a new Material that is based on PBR
	// definition.
	CreatePBRMaterial(definition PBRMaterialDefinition) Material

	// Destroy releases resources allocated by this
	// 3D engine.
	Destroy()
}
