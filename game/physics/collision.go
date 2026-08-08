package physics

import (
	"github.com/mokiat/lacking/core/spatial/placement3d"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

type CollisionShapeID struct {
	bodyID  BodyID
	shapeID placement3d.ShapeID
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
