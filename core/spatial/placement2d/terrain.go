package placement2d

// NilTerrainID indicates a terrain that can never be part of the scene.
const NilTerrainID = TerrainID(nilIndex)

// TerrainID is a reference to a terrain in the scene.
type TerrainID int32

// TerrainInfo contains the information needed to create a terrain in a scene.
//
// Unlike an object, a terrain has no transform of its own. The shapes that are
// attached to it are specified directly in world space.
type TerrainInfo[T any] struct {

	// UserData allows one to attach custom user data to a terrain.
	UserData T
}

type terrainState[T any] struct {
	firstShapeIndex int32
	lastShapeIndex  int32
	userData        T
}
