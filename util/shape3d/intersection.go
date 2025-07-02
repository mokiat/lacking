package shape3d

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
)

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

// WorstIntersection is an implementation of IntersectionCollection that keeps
// track of the worst (largest depth) observed intersection.
type WorstIntersection struct {
	intersection opt.T[Intersection]
}

// Reset clears any observed intersection.
func (i *WorstIntersection) Reset() {
	i.intersection = opt.Unspecified[Intersection]()
}

// AddIntersection tracks the specified intersection.
func (i *WorstIntersection) AddIntersection(intersection Intersection) {
	if i.intersection.Specified {
		if intersection.Depth > i.intersection.Value.Depth {
			i.intersection.Value = intersection
		}
	} else {
		i.intersection = opt.V(intersection)
	}
}

// Intersection returns the worst observed intersection and a flag whether
// there was actually any intersection observed.
func (i *WorstIntersection) Intersection() opt.T[Intersection] {
	return i.intersection
}

// BestIntersection is an implementation of IntersectionCollection that keeps
// track of the best (smallest depth) observed intersection.
type BestIntersection struct {
	intersection opt.T[Intersection]
}

// Reset clears any observed intersection.
func (i *BestIntersection) Reset() {
	i.intersection = opt.Unspecified[Intersection]()
}

// AddIntersection tracks the specified intersection.
func (i *BestIntersection) AddIntersection(intersection Intersection) {
	if i.intersection.Specified {
		if intersection.Depth < i.intersection.Value.Depth {
			i.intersection.Value = intersection
		}
	} else {
		i.intersection = opt.V(intersection)
	}
}

// Intersection returns the best observed intersection and a flag whether
// there was actually any intersection observed.
func (i *BestIntersection) Intersection() opt.T[Intersection] {
	return i.intersection
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

// WorstObjectIntersection is an implementation of IntersectionCollection that keeps
// track of the worst (largest depth) observed intersection.
type WorstObjectIntersection struct {
	intersection opt.T[ObjectIntersection]
}

// Reset clears any observed intersection.
func (i *WorstObjectIntersection) Reset() {
	i.intersection = opt.Unspecified[ObjectIntersection]()
}

// AddIntersection tracks the specified intersection.
func (i *WorstObjectIntersection) AddIntersection(intersection ObjectIntersection) {
	if i.intersection.Specified {
		if intersection.Depth > i.intersection.Value.Depth {
			i.intersection.Value = intersection
		}
	} else {
		i.intersection = opt.V(intersection)
	}
}

// Intersection returns the worst observed intersection and a flag whether
// there was actually any intersection observed.
func (i *WorstObjectIntersection) Intersection() opt.T[ObjectIntersection] {
	return i.intersection
}
