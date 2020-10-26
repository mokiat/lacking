package pack

import "github.com/mokiat/gomath/sprec"

type LevelProvider interface {
	Level() *Level
}

type Level struct {
	SkyboxTexture     string
	IrradianceTexture string
	StaticEntities    []LevelEntity
	StaticMeshes      []Mesh
	CollisionMeshes   []LevelCollisionMesh
}

type LevelEntity struct {
	Model  string
	Matrix sprec.Mat4
}

type LevelCollisionMesh struct {
	Triangles []Triangle
}

type Triangle struct {
	A sprec.Vec3
	B sprec.Vec3
	C sprec.Vec3
}
