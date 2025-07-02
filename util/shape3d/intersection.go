package shape3d

import "github.com/mokiat/gomath/dprec"

// Intersection represents the intersection of two shapes.
type Intersection struct {

	// FirstContact returns the point of contact on the first shape.
	FirstContact dprec.Vec3

	// FirstDisplaceNormal returns the normal along which the second shape
	// needs to be moved in order to separate the two shapes the fastest.
	FirstDisplaceNormal dprec.Vec3

	// SecondContact returns the point of contact on the second shape.
	SecondContact dprec.Vec3

	// SecondDisplaceNormal returns the normal along which the first shape
	// needs to be moved in order to separate the two shapes the fastest.
	SecondDisplaceNormal dprec.Vec3

	// Depth returns the amount of penetration between the two shapes.
	Depth float64
}

// ObjectIntersection represents the intersectio of two objects.
type ObjectIntersection struct {

	// FirstObjectID contains the ID of the first involved object.
	//
	// This ID is equal to NilObjectID() if the check was not performed with
	// a scene object.
	FirstObjectID ObjectID

	// SecondObjectID contains the ID of the second involved object.
	//
	// This ID is equal to NilObjectID() if the check was not performed with
	// a scene object.
	SecondObjectID ObjectID

	Intersection
}
