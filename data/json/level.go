package json

import (
	"encoding/json"
	"fmt"
	"io"
)

type Level struct {
	SkyboxTexture      string          `json:"skybox_texture"`
	IrradianceTexture  string          `json:"irradiance_texture"`
	StartCollisionMesh CollisionMesh   `json:"start_collision_mesh"`
	Waypoints          []Position      `json:"waypoints"`
	StaticMeshes       []Mesh          `json:"static_meshes"`
	CollisionMeshes    []CollisionMesh `json:"collision_meshes"`
	StaticEntities     []Entity        `json:"static_entities"`
}

type CollisionMesh struct {
	Triangles []Triangle `json:"triangles"`
}

type Triangle [3]Position

type Entity struct {
	Model  string      `json:"model"`
	Matrix [16]float32 `json:"matrix"`
}

type Position [3]float32

type Mesh struct {
	Coords    []float32 `json:"coords"`
	Normals   []float32 `json:"normals"`
	TexCoords []float32 `json:"tex_coords"`
	Indices   []int     `json:"indices"`
	SubMeshes []SubMesh `json:"sub_meshes"`
}

type SubMesh struct {
	Name           string `json:"name"`
	IndexOffset    int    `json:"index_offset"`
	IndexCount     int    `json:"index_count"`
	DiffuseTexture string `json:"diffuse_texture"`
}

func NewLevelDecoder() *LevelDecoder {
	return &LevelDecoder{}
}

type LevelDecoder struct{}

func (d *LevelDecoder) Decode(in io.Reader) (*Level, error) {
	var level Level
	if err := json.NewDecoder(in).Decode(&level); err != nil {
		return nil, fmt.Errorf("failed to decode json: %w", err)
	}
	return &level, nil
}

func NewLevelEncoder() *LevelEncoder {
	return &LevelEncoder{}
}

type LevelEncoder struct{}

func (e *LevelEncoder) Encode(out io.Writer, level *Level) error {
	if err := json.NewEncoder(out).Encode(level); err != nil {
		return fmt.Errorf("failed to encode json: %w", err)
	}
	return nil
}
