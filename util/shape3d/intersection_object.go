package shape3d

import "github.com/mokiat/gog/opt"

// ObjectIntersection represents the intersectio of two objects.
type ObjectIntersection struct {

	// SourceObjectID contains the ID of the first involved object.
	//
	// This ID is equal to NilObjectID() if the check was not performed with
	// a scene object.
	SourceObjectID ObjectID

	// TargetObjectID contains the ID of the second involved object.
	//
	// This ID is equal to NilObjectID() if the check was not performed with
	// a scene object.
	TargetObjectID ObjectID

	Intersection
}

// BestObjectIntersection is an implementation of IntersectionCollection that
// keeps track of the best (smallest depth) observed intersection.
type BestObjectIntersection struct {
	intersection opt.T[ObjectIntersection]
}

// Reset clears any observed intersection.
func (i *BestObjectIntersection) Reset() {
	i.intersection = opt.Unspecified[ObjectIntersection]()
}

// AddIntersection tracks the specified intersection.
func (i *BestObjectIntersection) AddIntersection(intersection ObjectIntersection) {
	if i.intersection.Specified {
		if intersection.Depth < i.intersection.Value.Depth {
			i.intersection.Value = intersection
		}
	} else {
		i.intersection = opt.V(intersection)
	}
}

// Intersection returns the worst observed intersection and a flag whether
// there was actually any intersection observed.
func (i *BestObjectIntersection) Intersection() (ObjectIntersection, bool) {
	return i.intersection.Value, i.intersection.Specified
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
func (i *WorstObjectIntersection) Intersection() (ObjectIntersection, bool) {
	return i.intersection.Value, i.intersection.Specified
}

// NewIntersectionBucket creates a new IntersectionBucket instance with
// the specified initial capacity.
func NewIntersectionBucket(initialCapacity int) *ObjectIntersectionBucket {
	return &ObjectIntersectionBucket{
		intersections: make([]ObjectIntersection, 0, initialCapacity),
	}
}

type ObjectIntersectionBucket struct {
	intersections []ObjectIntersection
}

// Reset clears the buffer of this result set so that it can be reused.
func (b *ObjectIntersectionBucket) Reset() {
	b.intersections = b.intersections[:0]
}

// Add adds a new Intersection to this set.
func (b *ObjectIntersectionBucket) AddIntersection(intersection ObjectIntersection) {
	b.intersections = append(b.intersections, intersection)
}

// IsEmpty returns whether no intersections were found.
func (b *ObjectIntersectionBucket) IsEmpty() bool {
	return len(b.intersections) == 0
}

// Intersections returns a slice of all intersections that have been observed.
//
// NOTE: The slice must not be modified or cached as it will be reused.
func (s *ObjectIntersectionBucket) Intersections() []ObjectIntersection {
	return s.intersections
}

func addIntersection(collection IntersectionCollection, flipped bool, intersection Intersection) {
	if flipped {
		collection.AddIntersection(intersection.Flipped())
	} else {
		collection.AddIntersection(intersection)
	}
}
