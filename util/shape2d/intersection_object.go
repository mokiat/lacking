package shape2d

import "github.com/mokiat/gog/opt"

// ObjectIntersectionCollection represents a data structure that can hold
// one or more object intersections.
type ObjectIntersectionCollection interface {

	// AddIntersection adds an intersection to the collection.
	AddIntersection(intersection ObjectIntersection)
}

// ObjectIntersection represents the intersectio of two objects.
type ObjectIntersection struct {

	// SourceObjectID contains the ID of the first involved object.
	//
	// This ID is equal to InvalidObjectID if the check was not performed with
	// a scene object.
	SourceObjectID ObjectID

	// SourceShapeID contains the ID of the shape from the first involved object.
	//
	// This ID is equal to InvalidShapeID if the check was not performed with
	// a scene object.
	SourceShapeID ShapeID

	// TargetObjectID contains the ID of the second involved object.
	//
	// This ID is equal to InvalidObjectID if the check was not performed with
	// a scene object.
	TargetObjectID ObjectID

	// TargetShapeID contains the ID of the shape from the second involved object.
	//
	// This ID is equal to InvalidShapeID if the check was not performed with
	// a scene object.
	TargetShapeID ShapeID

	// Intersection holds the underlying raw shape intersection.
	Intersection
}

// Flipped returns a flipped (source and target swapped) version of this
// intersection.
func (i *ObjectIntersection) Flipped() ObjectIntersection {
	return ObjectIntersection{
		SourceObjectID: i.TargetObjectID,
		SourceShapeID:  i.TargetShapeID,
		TargetObjectID: i.SourceObjectID,
		TargetShapeID:  i.SourceShapeID,
		Intersection:   i.Intersection.Flipped(),
	}
}

// SmallestObjectIntersection is an implementation of
// ObjectIntersectionCollection that keeps track of the nearest (smallest depth)
// observed intersection.
type SmallestObjectIntersection struct {
	intersection opt.T[ObjectIntersection]
}

// Reset clears any observed intersection.
func (i *SmallestObjectIntersection) Reset() {
	i.intersection = opt.Unspecified[ObjectIntersection]()
}

// AddIntersection tracks the specified intersection.
func (i *SmallestObjectIntersection) AddIntersection(intersection ObjectIntersection) {
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
func (i *SmallestObjectIntersection) Intersection() (ObjectIntersection, bool) {
	return i.intersection.Value, i.intersection.Specified
}

// LargestObjectIntersection is an implementation of
// ObjectIntersectionCollection that keeps track of the farthest (largest depth)
// observed intersection.
type LargestObjectIntersection struct {
	intersection opt.T[ObjectIntersection]
}

// Reset clears any observed intersection.
func (i *LargestObjectIntersection) Reset() {
	i.intersection = opt.Unspecified[ObjectIntersection]()
}

// AddIntersection tracks the specified intersection.
func (i *LargestObjectIntersection) AddIntersection(intersection ObjectIntersection) {
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
func (i *LargestObjectIntersection) Intersection() (ObjectIntersection, bool) {
	return i.intersection.Value, i.intersection.Specified
}

// NewIntersectionBucket creates a new IntersectionBucket instance with
// the specified initial capacity.
func NewIntersectionBucket(initialCapacity int) *ObjectIntersectionBucket {
	return &ObjectIntersectionBucket{
		intersections: make([]ObjectIntersection, 0, initialCapacity),
	}
}

// ObjectIntersectionBucket is a container for object intersections.
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
