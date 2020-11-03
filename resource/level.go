package resource

import (
	"fmt"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/async"
	"github.com/mokiat/lacking/data/asset"
	"github.com/mokiat/lacking/shape"
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

type Entity struct {
	Model  *Model
	Matrix sprec.Mat4
}

func NewLevelOperator(locator Locator, gfxWorker *async.Worker) *LevelOperator {
	return &LevelOperator{
		locator:   locator,
		gfxWorker: gfxWorker,
	}
}

type LevelOperator struct {
	locator   Locator
	gfxWorker *async.Worker
}

func (o *LevelOperator) Allocate(registry *Registry, name string) (interface{}, error) {
	in, err := o.locator.Open("assets", "levels", name)
	if err != nil {
		return nil, fmt.Errorf("failed to open level asset %q: %w", name, err)
	}
	defer in.Close()

	levelAsset := new(asset.Level)
	if err := asset.DecodeLevel(in, levelAsset); err != nil {
		return nil, fmt.Errorf("failed to decode level asset %q: %w", name, err)
	}

	level := &Level{
		Name: name,
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

	trianglesCenter := func(triangles []shape.StaticTriangle) sprec.Vec3 {
		var center sprec.Vec3
		count := 0
		for _, triangle := range triangles {
			center = sprec.Vec3Sum(center, triangle.A())
			center = sprec.Vec3Sum(center, triangle.B())
			center = sprec.Vec3Sum(center, triangle.C())
			count += 3
		}
		return sprec.Vec3Quot(center, float32(count))
	}

	convertCollisionMesh := func(collisionMeshAsset asset.LevelCollisionMesh) shape.Placement {
		var triangles []shape.StaticTriangle
		for _, triangleAsset := range collisionMeshAsset.Triangles {
			triangles = append(triangles, shape.NewStaticTriangle(
				sprec.NewVec3(triangleAsset[0][0], triangleAsset[0][1], triangleAsset[0][2]),
				sprec.NewVec3(triangleAsset[1][0], triangleAsset[1][1], triangleAsset[1][2]),
				sprec.NewVec3(triangleAsset[2][0], triangleAsset[2][1], triangleAsset[2][2]),
			))
		}
		center := trianglesCenter(triangles)
		for i := range triangles {
			triangles[i] = shape.NewStaticTriangle(
				sprec.Vec3Diff(triangles[i].A(), center),
				sprec.Vec3Diff(triangles[i].B(), center),
				sprec.Vec3Diff(triangles[i].C(), center),
			)
		}
		return shape.Placement{
			Position:    center,
			Orientation: sprec.IdentityQuat(),
			Shape:       shape.NewStaticMesh(triangles),
		}
	}

	collisionMeshes := make([]shape.Placement, len(levelAsset.CollisionMeshes))
	for i, collisionMeshAsset := range levelAsset.CollisionMeshes {
		collisionMeshes[i] = convertCollisionMesh(collisionMeshAsset)
	}
	level.CollisionMeshes = collisionMeshes

	staticMeshes := make([]*Mesh, len(levelAsset.StaticMeshes))
	for i, staticMeshAsset := range levelAsset.StaticMeshes {
		staticMesh, err := AllocateMesh(registry, staticMeshAsset.Name, o.gfxWorker, &staticMeshAsset)
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
			Matrix: floatArrayToMatrix(staticEntityAsset.Matrix),
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
		if err := ReleaseMesh(registry, o.gfxWorker, staticMesh); err != nil {
			return fmt.Errorf("failed to release mesh: %w", err)
		}
	}

	if result := registry.UnloadCubeTexture(level.SkyboxTexture).Wait(); result.Err != nil {
		return result.Err
	}
	return nil
}
