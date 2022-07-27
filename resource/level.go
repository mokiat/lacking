package resource

import (
	"fmt"

	"github.com/mokiat/gomath/dprec"
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/data/asset"
	gameasset "github.com/mokiat/lacking/game/asset"
	"github.com/mokiat/lacking/game/graphics"
	"github.com/mokiat/lacking/util/shape"
)

const LevelTypeName = TypeName("level")

func InjectLevel(target **Level) func(value interface{}) {
	return func(value interface{}) {
		*target = value.(*Level)
	}
}

type Level struct {
	Name                     string
	SkyboxTexture            *CubeTexture
	AmbientReflectionTexture *CubeTexture
	AmbientRefractionTexture *CubeTexture
	CollisionMeshes          []shape.Placement
	StaticMeshes             []*Mesh
	StaticEntities           []*Entity
}

func (l *Level) FindStaticEntity(id string) *Entity {
	for _, entity := range l.StaticEntities {
		if entity.Model.Name == id {
			return entity
		}
	}
	return nil
}

type Entity struct {
	Model  *Model
	Matrix sprec.Mat4
}

func NewLevelOperator(delegate gameasset.Registry, gfxEngine *graphics.Engine) *LevelOperator {
	return &LevelOperator{
		delegate:  delegate,
		gfxEngine: gfxEngine,
	}
}

type LevelOperator struct {
	delegate  gameasset.Registry
	gfxEngine *graphics.Engine
}

func (o *LevelOperator) Allocate(registry *Registry, id string) (interface{}, error) {
	levelAsset := new(asset.Level)
	resource := o.delegate.ResourceByID(id)
	if resource == nil {
		return nil, fmt.Errorf("cannot find asset %q", id)
	}
	if err := resource.ReadContent(levelAsset); err != nil {
		return nil, fmt.Errorf("failed to open level asset %q: %w", id, err)
	}

	level := &Level{
		Name: id,
	}

	if result := registry.LoadCubeTexture(levelAsset.SkyboxTexture).OnSuccess(InjectCubeTexture(&level.SkyboxTexture)).Wait(); result.Err != nil {
		return nil, result.Err
	}
	if result := registry.LoadCubeTexture(levelAsset.AmbientReflectionTexture).OnSuccess(InjectCubeTexture(&level.AmbientReflectionTexture)).Wait(); result.Err != nil {
		return nil, result.Err
	}
	if result := registry.LoadCubeTexture(levelAsset.AmbientRefractionTexture).OnSuccess(InjectCubeTexture(&level.AmbientRefractionTexture)).Wait(); result.Err != nil {
		return nil, result.Err
	}

	trianglesCenter := func(triangles []shape.StaticTriangle) dprec.Vec3 {
		var center dprec.Vec3
		count := 0
		for _, triangle := range triangles {
			center = dprec.Vec3Sum(center, triangle.A())
			center = dprec.Vec3Sum(center, triangle.B())
			center = dprec.Vec3Sum(center, triangle.C())
			count += 3
		}
		return dprec.Vec3Quot(center, float64(count))
	}

	convertCollisionMesh := func(collisionMeshAsset asset.LevelCollisionMesh) shape.Placement {
		var triangles []shape.StaticTriangle
		for _, triangleAsset := range collisionMeshAsset.Triangles {
			triangles = append(triangles, shape.NewStaticTriangle(
				dprec.NewVec3(float64(triangleAsset[0][0]), float64(triangleAsset[0][1]), float64(triangleAsset[0][2])),
				dprec.NewVec3(float64(triangleAsset[1][0]), float64(triangleAsset[1][1]), float64(triangleAsset[1][2])),
				dprec.NewVec3(float64(triangleAsset[2][0]), float64(triangleAsset[2][1]), float64(triangleAsset[2][2])),
			))
		}
		center := trianglesCenter(triangles)
		for i := range triangles {
			triangles[i] = shape.NewStaticTriangle(
				dprec.Vec3Diff(triangles[i].A(), center),
				dprec.Vec3Diff(triangles[i].B(), center),
				dprec.Vec3Diff(triangles[i].C(), center),
			)
		}
		return shape.NewPlacement(
			shape.NewStaticMesh(triangles),
			center,
			dprec.IdentityQuat(),
		)
	}

	collisionMeshes := make([]shape.Placement, len(levelAsset.CollisionMeshes))
	for i, collisionMeshAsset := range levelAsset.CollisionMeshes {
		collisionMeshes[i] = convertCollisionMesh(collisionMeshAsset)
	}
	level.CollisionMeshes = collisionMeshes

	staticMaterials := make([]*Material, len(levelAsset.Materials))
	for i, staticMaterial := range levelAsset.Materials {
		material, err := AllocateMaterial(registry, o.gfxEngine, &staticMaterial)
		if err != nil {
			return nil, fmt.Errorf("failed to allocate material: %w", err)
		}
		staticMaterials[i] = material
	}

	staticMeshes := make([]*Mesh, len(levelAsset.StaticMeshes))
	for i, staticMeshAsset := range levelAsset.StaticMeshes {
		staticMesh, err := AllocateMesh(registry, o.gfxEngine, staticMaterials, &staticMeshAsset)
		if err != nil {
			return nil, fmt.Errorf("failed to allocate mesh: %w", err)
		}
		staticMeshes[i] = staticMesh
	}
	level.StaticMeshes = staticMeshes

	staticEntities := make([]*Entity, len(levelAsset.StaticEntities))
	for i, staticEntityAsset := range levelAsset.StaticEntities {
		var model *Model
		if result := registry.LoadModel(staticEntityAsset.Model).OnSuccess(InjectModel(&model)).Wait(); result.Err != nil {
			return nil, result.Err
		}
		staticEntities[i] = &Entity{
			Model:  model,
			Matrix: sprec.ColumnMajorArrayToMat4(staticEntityAsset.Matrix),
		}
	}
	level.StaticEntities = staticEntities

	return level, nil
}

func (o *LevelOperator) Release(registry *Registry, res interface{}) error {
	level := res.(*Level)

	for _, staticEntity := range level.StaticEntities {
		if result := registry.UnloadModel(staticEntity.Model).Wait(); result.Err != nil {
			return result.Err
		}
	}
	for _, staticMesh := range level.StaticMeshes {
		if err := ReleaseMesh(registry, staticMesh); err != nil {
			return fmt.Errorf("failed to release mesh: %w", err)
		}
	}

	if result := registry.UnloadCubeTexture(level.SkyboxTexture).Wait(); result.Err != nil {
		return result.Err
	}
	return nil
}
