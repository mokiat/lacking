package resource

import (
	"fmt"

	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/async"
	"github.com/mokiat/lacking/data/asset"
	"github.com/mokiat/lacking/shape"
)

type Level struct {
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

func (o *LevelOperator) Allocator(uri string) Allocator {
	return AllocatorFunc(func(set *Set) (interface{}, error) {
		in, err := o.locator.Open(uri)
		if err != nil {
			return nil, fmt.Errorf("failed to open level asset %q: %w", uri, err)
		}
		defer in.Close()

		levelAsset := new(asset.Level)
		if err := asset.DecodeLevel(in, levelAsset); err != nil {
			return nil, fmt.Errorf("failed to decode level asset %q: %w", uri, err)
		}

		level := &Level{}

		uri := fmt.Sprintf("assets/textures/cube/%s.dat", levelAsset.SkyboxTexture)
		if err := set.OpenCubeTexture(uri, &level.SkyboxTexture).Wait(); err != nil {
			return nil, err
		}
		uri = fmt.Sprintf("assets/textures/cube/%s.dat", levelAsset.AmbientReflectionTexture)
		if err := set.OpenCubeTexture(uri, &level.AmbientReflectionTexture).Wait(); err != nil {
			return nil, err
		}
		uri = fmt.Sprintf("assets/textures/cube/%s.dat", levelAsset.AmbientRefractionTexture)
		if err := set.OpenCubeTexture(uri, &level.AmbientRefractionTexture).Wait(); err != nil {
			return nil, err
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
			staticMesh, err := AllocateMesh(set, o.gfxWorker, &staticMeshAsset)
			if err != nil {
				return nil, fmt.Errorf("failed to allocate mesh: %w", err)
			}
			staticMeshes[i] = staticMesh
		}
		level.StaticMeshes = staticMeshes

		staticEntities := make([]*Entity, len(levelAsset.StaticEntities))
		for i, staticEntityAsset := range levelAsset.StaticEntities {
			var model *Model
			uri = fmt.Sprintf("assets/models/%s.dat", staticEntityAsset.Model)
			if err := set.OpenModel(uri, &model).Wait(); err != nil {
				return nil, err
			}
			staticEntities[i] = &Entity{
				Model:  model,
				Matrix: floatArrayToMatrix(staticEntityAsset.Matrix),
			}
		}
		level.StaticEntities = staticEntities

		return level, nil
	})
}

func (o *LevelOperator) Releaser() Releaser {
	return ReleaserFunc(func(resource interface{}) error {
		level := resource.(*Level)

		for _, staticMesh := range level.StaticMeshes {
			if err := ReleaseMesh(o.gfxWorker, staticMesh); err != nil {
				return fmt.Errorf("failed to release mesh: %w", err)
			}
		}
		return nil
	})
}
