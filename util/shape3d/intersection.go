package shape3d

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
)

type IntersectionCollection interface {
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

func (i *Intersection) EvalSourceContact() dprec.Vec3 {
	return dprec.Vec3Sum(i.TargetContact, dprec.Vec3Prod(i.TargetNormal, -i.Depth))
}

func (i *Intersection) EvalSourceNormal() dprec.Vec3 {
	return dprec.InverseVec3(i.TargetNormal)
}

func (i *Intersection) Flipped() Intersection {
	return Intersection{
		TargetContact: i.EvalSourceContact(),
		TargetNormal:  i.EvalSourceNormal(),
		Depth:         i.Depth,
	}
}

type OptIntersection opt.T[Intersection]

func (i *OptIntersection) AddIntersection(intersection Intersection) {
	*i = OptIntersection(opt.V(intersection))
}

func (i *OptIntersection) Intersection() opt.T[Intersection] {
	return opt.T[Intersection](*i)
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
func (i *LastIntersection) Intersection() opt.T[Intersection] {
	return i.intersection
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

func addIntersection(collection IntersectionCollection, flipped bool, intersection Intersection) {
	if flipped {
		collection.AddIntersection(intersection.Flipped())
	} else {
		collection.AddIntersection(intersection)
	}
}
