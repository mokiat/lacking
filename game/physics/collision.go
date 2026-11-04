package physics

import (
	"github.com/mokiat/lacking/core/spatial/placement3d"
	"github.com/mokiat/lacking/core/spatial/shape3d"
)

// Mask is a bitmask over the collision layers that a shape can occupy.
//
// See [placement3d.Mask] for details.
type Mask = placement3d.Mask

// FullMask is a [Mask] with all layer bits set.
//
// See [placement3d.FullMask] for details.
const FullMask = placement3d.FullMask

// Filter narrows down the collision shapes that a segment query, such as
// [Scene.CollectSegmentBodyIntersections], considers. Its zero value
// considers every shape.
//
// See [placement3d.Filter] for details.
type Filter = placement3d.Filter

// BodyCollisionShapeID identifies a collision shape that was attached to a
// body through [BodyView.AttachCollisionSphere] or
// [BodyView.AttachCollisionBox]. It can be passed to
// [BodyView.DetachCollisionShape] to remove the shape again.
type BodyCollisionShapeID struct {
	bodyID  BodyID
	shapeID placement3d.ObjectShapeID
}

// TerrainCollisionShapeID identifies a collision shape that was attached to
// a terrain through [TerrainView.AttachCollisionMesh]. It can be passed to
// [TerrainView.DetachCollisionShape] to remove the shape again.
type TerrainCollisionShapeID struct {
	terrainID TerrainID
	shapeID   placement3d.TerrainShapeID
}

// CollisionShape describes a collision shape of geometry T, together with
// the surface material and filtering settings that are used whenever the
// shape participates in a collision.
type CollisionShape[T any] struct {

	// Shape is the geometry of the collision shape. For a body, it is
	// expressed in the body's local space; for a terrain, it is expressed
	// in world space.
	Shape T

	// FrictionCoefficient is the friction coefficient of the shape's
	// surface. It is combined with that of the other shape in a contact
	// through [CombinedFrictionCoefficient].
	FrictionCoefficient float64

	// RestitutionCoefficient is the restitution (bounciness) coefficient
	// of the shape's surface. It is combined with that of the other shape
	// in a contact through [CombinedRestitutionCoefficient].
	RestitutionCoefficient float64

	// Filtering determines which other shapes this shape can collide with,
	// through collision layers and reject groups (see
	// [Scene.NextCollisionRejectGroup]), as well as which segment queries
	// it is visible to.
	Filtering placement3d.FilterInfo
}

// CollisionSphere is a sphere-shaped [CollisionShape], which can be
// attached to a body through [BodyView.AttachCollisionSphere].
type CollisionSphere CollisionShape[shape3d.Sphere]

// CollisionBox is a box-shaped [CollisionShape], which can be attached to a
// body through [BodyView.AttachCollisionBox].
type CollisionBox CollisionShape[shape3d.Box]

// CollisionMesh is a triangle-mesh [CollisionShape], which can be attached
// to a terrain through [TerrainView.AttachCollisionMesh].
type CollisionMesh CollisionShape[shape3d.Mesh]
