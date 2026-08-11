package physics

import (
	"github.com/mokiat/lacking/core/spatial/placement3d"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

type Mask = placement3d.Mask

const FullMask = placement3d.FullMask

type Filter = placement3d.Filter

type BodyCollisionShapeID struct {
	bodyID  BodyID
	shapeID placement3d.ObjectShapeID
}

type TerrainCollisionShapeID struct {
	terrainID TerrainID
	shapeID   placement3d.TerrainShapeID
}

type CollisionShape[T any] struct {
	Shape                  T
	FrictionCoefficient    float64
	RestitutionCoefficient float64
	Filtering              placement3d.FilterInfo
}

type CollisionSphere CollisionShape[shape3d.Sphere]

type CollisionBox CollisionShape[shape3d.Box]

type CollisionMesh CollisionShape[shape3d.Mesh]
