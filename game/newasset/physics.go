package asset

import "github.com/mokiat/gomath/dprec"

// Body represents a physical body.
type Body struct {

	// NodeIndex is the index of the node that this body is attached to.
	NodeIndex uint32

	// BodyDefinitionIndex is the index of the body definition that this
	// body uses.
	BodyDefinitionIndex uint32
}

// BodyDefinition represents the physical properties of a body.
type BodyDefinition struct {

	// MaterialIndex is the index of the material that this body uses.
	MaterialIndex uint32

	// Mass is the mass of the body.
	Mass float64

	// MomentOfInertia is the moment of inertia of the body represented
	// as 3x3 tensor.
	MomentOfInertia dprec.Mat3

	// DragFactor is the linear drag factor of the body.
	DragFactor float64

	// AngularDragFactor is the angular drag factor of the body.
	AngularDragFactor float64

	// CollisionBoxes is a list of collision boxes that define the
	// collision shape of the body.
	CollisionBoxes []CollisionBox

	// CollisionSpheres is a list of collision spheres that define the
	// collision shape of the body.
	CollisionSpheres []CollisionSphere

	// CollisionMeshes is a list of collision meshes that define the
	// collision shape of the body.
	CollisionMeshes []CollisionMesh
}

// BodyMaterial represents a physical material.
type BodyMaterial struct {

	// FrictionCoefficient is the coefficient of friction of this material.
	// Lower values mean more slippery surfaces.
	FrictionCoefficient float64

	// RestitutionCoefficient is the coefficient of restitution of this material.
	// Higher values mean more bouncy surfaces.
	RestitutionCoefficient float64
}

// CollisionBox represents a box-shaped collision volume.
type CollisionBox struct {

	// Translation is the position of the box.
	Translation dprec.Vec3

	// Rotation is the orientation of the box.
	Rotation dprec.Quat

	// Width is the width of the box.
	Width float64

	// Height is the height of the box.
	Height float64

	// Lenght is the length of the box.
	Lenght float64
}

// CollisionSphere represents a sphere-shaped collision volume.
type CollisionSphere struct {

	// Translation is the position of the sphere.
	Translation dprec.Vec3

	// Radius is the radius of the sphere.
	Radius float64
}

// CollisionMesh represents a mesh-shaped collision volume.
type CollisionMesh struct {

	// Translation is the position of the mesh.
	Translation dprec.Vec3

	// Rotation is the orientation of the mesh.
	Rotation dprec.Quat

	// Triangles is a list of triangles that define the collision shape
	Triangles []CollisionTriangle
}

// CollisionTriangle represents a triangle-shaped collision surface.
//
// Ordering of the vertices determines the normal direction.
type CollisionTriangle struct {

	// A is the first vertex of the triangle.
	A dprec.Vec3

	// B is the second vertex of the triangle.
	B dprec.Vec3

	// C is the third vertex of the triangle.
	C dprec.Vec3
}
