package pack

import (
	"fmt"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/data/json"
)

type OpenLevelResourceAction struct {
	locator ResourceLocator
	uri     string
	level   *Level
}

func (a *OpenLevelResourceAction) Describe() string {
	return fmt.Sprintf("open_level_resource(uri: %q)", a.uri)
}

func (a *OpenLevelResourceAction) Level() *Level {
	if a.level == nil {
		panic("reading data from unprocessed action")
	}
	return a.level
}

func (a *OpenLevelResourceAction) Run() error {
	in, err := a.locator.Open(a.uri)
	if err != nil {
		return fmt.Errorf("failed to open level resource %q: %w", a.uri, err)
	}
	defer in.Close()

	jsonLevel, err := json.NewLevelDecoder().Decode(in)
	if err != nil {
		return fmt.Errorf("failed to decode level %q: %w", a.uri, err)
	}

	a.level = &Level{
		SkyboxTexture:            jsonLevel.SkyboxTexture,
		AmbientReflectionTexture: jsonLevel.AmbientReflectionTexture,
		AmbientRefractionTexture: jsonLevel.AmbientRefractionTexture,
		CollisionMeshes:          make([]LevelCollisionMesh, len(jsonLevel.CollisionMeshes)),
		StaticMeshes:             make([]Mesh, len(jsonLevel.StaticMeshes)),
		StaticEntities:           make([]LevelEntity, len(jsonLevel.StaticEntities)),
	}

	for i, jsonCollisionMesh := range jsonLevel.CollisionMeshes {
		collisionMesh := LevelCollisionMesh{
			Triangles: make([]Triangle, len(jsonCollisionMesh.Triangles)),
		}
		for j, jsonTriangle := range jsonCollisionMesh.Triangles {
			collisionMesh.Triangles[j] = Triangle{
				A: sprec.NewVec3(jsonTriangle[0][0], jsonTriangle[0][1], jsonTriangle[0][2]),
				B: sprec.NewVec3(jsonTriangle[1][0], jsonTriangle[1][1], jsonTriangle[1][2]),
				C: sprec.NewVec3(jsonTriangle[2][0], jsonTriangle[2][1], jsonTriangle[2][2]),
			}
		}
		a.level.CollisionMeshes[i] = collisionMesh
	}

	for i, jsonStaticMesh := range jsonLevel.StaticMeshes {
		staticMesh := Mesh{
			Name: "unnamed",
			VertexLayout: VertexLayout{
				HasCoords:    true,
				HasNormals:   true,
				HasTexCoords: true,
			},
			Vertices:  make([]Vertex, len(jsonStaticMesh.Coords)/3),
			Indices:   make([]int, len(jsonStaticMesh.Indices)),
			SubMeshes: make([]SubMesh, len(jsonStaticMesh.SubMeshes)),
		}
		for j := range staticMesh.Vertices {
			staticMesh.Vertices[j].Coord = sprec.NewVec3(
				jsonStaticMesh.Coords[j*3+0],
				jsonStaticMesh.Coords[j*3+1],
				jsonStaticMesh.Coords[j*3+2],
			)
		}
		for j := range staticMesh.Vertices {
			staticMesh.Vertices[j].Normal = sprec.NewVec3(
				jsonStaticMesh.Normals[j*3+0],
				jsonStaticMesh.Normals[j*3+1],
				jsonStaticMesh.Normals[j*3+2],
			)
		}
		for j := range staticMesh.Vertices {
			staticMesh.Vertices[j].TexCoord = sprec.NewVec2(
				jsonStaticMesh.TexCoords[j*2+0],
				jsonStaticMesh.TexCoords[j*2+1],
			)
		}
		for j := range staticMesh.Indices {
			staticMesh.Indices[j] = jsonStaticMesh.Indices[j]
		}
		for j, jsonSubMesh := range jsonStaticMesh.SubMeshes {
			staticMesh.SubMeshes[j] = SubMesh{
				Primitive:   PrimitiveTriangles,
				IndexOffset: jsonSubMesh.IndexOffset,
				IndexCount:  jsonSubMesh.IndexCount,
				Material: Material{
					Type:                     "pbr",
					BackfaceCulling:          true,
					AlphaTesting:             false,
					AlphaThreshold:           0.5,
					Metallic:                 0.0,
					Roughness:                0.8,
					MetallicRoughnessTexture: "",
					Color:                    sprec.ZeroVec4(),
					ColorTexture:             jsonSubMesh.DiffuseTexture,
					NormalScale:              1.0,
					NormalTexture:            "",
				},
			}
		}

		a.level.StaticMeshes[i] = staticMesh
	}

	for i, jsonStaticEntity := range jsonLevel.StaticEntities {
		a.level.StaticEntities[i] = LevelEntity{
			Model: jsonStaticEntity.Model,
			Matrix: sprec.NewMat4(
				jsonStaticEntity.Matrix[0], jsonStaticEntity.Matrix[4], jsonStaticEntity.Matrix[8], jsonStaticEntity.Matrix[12],
				jsonStaticEntity.Matrix[1], jsonStaticEntity.Matrix[5], jsonStaticEntity.Matrix[9], jsonStaticEntity.Matrix[13],
				jsonStaticEntity.Matrix[2], jsonStaticEntity.Matrix[6], jsonStaticEntity.Matrix[10], jsonStaticEntity.Matrix[14],
				jsonStaticEntity.Matrix[3], jsonStaticEntity.Matrix[7], jsonStaticEntity.Matrix[11], jsonStaticEntity.Matrix[15],
			),
		}
	}

	return nil
}
