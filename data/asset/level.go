package asset

import (
	"io"

	"github.com/mokiat/lacking/game/asset"
)

type Level struct {
	SkyboxTexture            string
	AmbientReflectionTexture string
	AmbientRefractionTexture string
	Materials                []asset.Material
	StaticEntities           []LevelEntity
	StaticMeshes             []asset.MeshDefinition
	CollisionMeshes          []LevelCollisionMesh
}

func (l *Level) EncodeTo(out io.Writer) error {
	return Encode(out, l)
}

func (l *Level) DecodeFrom(in io.Reader) error {
	return Decode(in, l)
}

type LevelEntity struct {
	Model  string
	Matrix [16]float32
}

type LevelCollisionMesh struct {
	Triangles []Triangle
}

type Triangle [3]Point

type Point [3]float32

func EncodeLevel(out io.Writer, level *Level) error {
	return Encode(out, level)
}

func DecodeLevel(in io.Reader, level *Level) error {
	return Decode(in, level)
}
