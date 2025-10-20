package shape3d

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
)

// IntersectionCollection represents a data structure that can hold
// one or more intersections.
type IntersectionCollection interface {

	// AddIntersection adds an intersection to the collection.
	AddIntersection(intersection Intersection)
}

// Intersection represents the intersection of two shapes.
type Intersection struct {

	// TargetContact contains the point on the target shape where the
	// intersection first occurred.
	TargetContact dprec.Vec3

	// TargetNormal contains the best direction in which the source shape
	// must be translated to avoid the intersection.
	TargetNormal dprec.Vec3

	// Depth returns the amount of penetration between the two shapes.
	Depth float64
}

// EvalSourceContact calculates the contact point on the source shape.
func (i *Intersection) EvalSourceContact() dprec.Vec3 {
	return dprec.Vec3Sum(i.TargetContact, dprec.Vec3Prod(i.TargetNormal, -i.Depth))
}

// EvalSourceNormal calculates the normal on the source shape along which
// the target shape needs to be repositioned to resolve the intersection.
func (i *Intersection) EvalSourceNormal() dprec.Vec3 {
	return dprec.InverseVec3(i.TargetNormal)
}

// Flipped returns a flipped (source and target swapped) version of this
// intersection.
func (i *Intersection) Flipped() Intersection {
	return Intersection{
		TargetContact: i.EvalSourceContact(),
		TargetNormal:  i.EvalSourceNormal(),
		Depth:         i.Depth,
	}
}

// LastIntersection is an implementation of IntersectionCollection that keeps
// track of the last observed intersection.
type LastIntersection struct {
	intersection opt.T[Intersection]
}

// Reset clears any observed intersection.
func (i *LastIntersection) Reset() {
	i.intersection = opt.Unspecified[Intersection]()
}

// AddIntersection tracks the specified intersection.
func (i *LastIntersection) AddIntersection(intersection Intersection) {
	i.intersection = opt.V(intersection)
}

// Intersection returns the last observed intersection and a flag whether
// there was actually any intersection observed.
func (i *LastIntersection) Intersection() (Intersection, bool) {
	return i.intersection.Value, i.intersection.Specified
}

// SmallestIntersection is an implementation of IntersectionCollection that keeps
// track of the farthest (smallest depth) observed intersection.
type SmallestIntersection struct {
	intersection opt.T[Intersection]
}

// Reset clears any observed intersection.
func (i *SmallestIntersection) Reset() {
	i.intersection = opt.Unspecified[Intersection]()
}

// AddIntersection tracks the specified intersection.
func (i *SmallestIntersection) AddIntersection(intersection Intersection) {
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
func (i *SmallestIntersection) Intersection() (Intersection, bool) {
	return i.intersection.Value, i.intersection.Specified
}

// LargestIntersection is an implementation of IntersectionCollection that
// keeps track of the closest (largest depth) observed intersection.
type LargestIntersection struct {
	intersection opt.T[Intersection]
}

// Reset clears any observed intersection.
func (i *LargestIntersection) Reset() {
	i.intersection = opt.Unspecified[Intersection]()
}

// AddIntersection tracks the specified intersection.
func (i *LargestIntersection) AddIntersection(intersection Intersection) {
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
func (i *LargestIntersection) Intersection() (Intersection, bool) {
	return i.intersection.Value, i.intersection.Specified
}
