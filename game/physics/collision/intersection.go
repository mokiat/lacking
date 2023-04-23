package collision

import (
	"github.com/mokiat/gog/opt"
	"github.com/mokiat/gomath/dprec"
)

// Intersection represents the collision between two shapes.
type Intersection struct {

	// Depth returns the amount of penetration between the two shapes.
	Depth float64

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
}

// Flipped returns a new Intersection where the first and second shapes have
// their places swapped within the structure.
func (i *Intersection) Flipped() Intersection {
	return Intersection{
		Depth:                i.Depth,
		FirstContact:         i.SecondContact,
		FirstDisplaceNormal:  i.SecondDisplaceNormal,
		SecondContact:        i.SecondContact,
		SecondDisplaceNormal: i.SecondDisplaceNormal,
	}
}

// IntersectionCollection represents a data structure that can hold
// intersections.
type IntersectionCollection interface {
	AddIntersection(intersection Intersection)
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
func (i *WorstIntersection) Intersection() (Intersection, bool) {
	return i.intersection.Value, i.intersection.Specified
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
func (i *BestIntersection) Intersection() (Intersection, bool) {
	return i.intersection.Value, i.intersection.Specified
}

// NewIntersectionBucket creates a new IntersectionBucket instance with
// the specified initial capacity.
func NewIntersectionBucket(initialCapacity int) *IntersectionBucket {
	return &IntersectionBucket{
		intersections: make([]Intersection, 0, initialCapacity),
	}
}

// IntersectionBucket is a structure that can be used to collect the result of
// an intersection test.
type IntersectionBucket struct {
	intersections []Intersection
}

// Reset clears the buffer of this result set so that it can be reused.
func (b *IntersectionBucket) Reset() {
	b.intersections = b.intersections[:0]
}

// Add adds a new Intersection to this set.
func (b *IntersectionBucket) AddIntersection(intersection Intersection) {
	b.intersections = append(b.intersections, intersection)
}

// IsEmpty returns whether no intersections were found.
func (b *IntersectionBucket) IsEmpty() bool {
	return len(b.intersections) == 0
}

// Intersections returns a slice of all intersections that have been observed.
//
// NOTE: The slice must not be modified or cached as it will be reused.
func (s *IntersectionBucket) Intersections() []Intersection {
	return s.intersections
}

// IsSphereWithSphereIntersecting is a quick check to determine if two spheres
// are intersecting.
func IsSphereWithSphereIntersecting(bs1, bs2 Sphere) bool {
	sqrDistance := dprec.Vec3Diff(bs2.position, bs1.position).Length()
	return sqrDistance <= (bs1.radius + bs2.radius)
}

// CheckIntersectionSetWithSet checks for any intersections between the two
// collision Sets.
func CheckIntersectionSetWithSet(first, second Set, flipped bool, resultSet IntersectionCollection) {
	if !IsSphereWithSphereIntersecting(first.BoundingSphere(), second.BoundingSphere()) {
		return
	}
	for _, firstSphere := range first.Spheres() {
		for _, secondSphere := range second.Spheres() {
			if IsSphereWithSphereIntersecting(firstSphere, secondSphere) {
				CheckIntersectionSphereWithSphere(firstSphere, secondSphere, false, resultSet)
			}
		}
		for _, secondaryMesh := range second.Meshes() {
			if IsSphereWithSphereIntersecting(firstSphere, secondaryMesh.BoundingSphere()) {
				CheckIntersectionSphereWithMesh(firstSphere, secondaryMesh, false, resultSet)
			}
		}
	}
	for _, primaryBox := range first.Boxes() {
		for _, secondaryMesh := range second.Meshes() {
			if IsSphereWithSphereIntersecting(primaryBox.BoundingSphere(), secondaryMesh.BoundingSphere()) {
				CheckIntersectionBoxWithMesh(primaryBox, secondaryMesh, false, resultSet)
			}
		}
	}
	for _, primaryMesh := range first.Meshes() {
		for _, secondarySphere := range second.Spheres() {
			if IsSphereWithSphereIntersecting(primaryMesh.BoundingSphere(), secondarySphere) {
				CheckIntersectionSphereWithMesh(secondarySphere, primaryMesh, true, resultSet)
			}
		}
		for _, secondaryBox := range second.Boxes() {
			if IsSphereWithSphereIntersecting(primaryMesh.BoundingSphere(), secondaryBox.BoundingSphere()) {
				CheckIntersectionBoxWithMesh(secondaryBox, primaryMesh, true, resultSet)
			}
		}
	}
}

// CheckIntersectionSphereWithMesh checks if a Sphere shape intersects with
// a Mesh shape.
func CheckIntersectionSphereWithMesh(sphere Sphere, mesh Mesh, flipped bool, resultSet IntersectionCollection) {
	var worstIntersection WorstIntersection
	for _, triangle := range mesh.triangles {
		if IsSphereWithSphereIntersecting(sphere, triangle.BoundingSphere()) {
			CheckIntersectionSphereWithTriangle(sphere, triangle, flipped, &worstIntersection)
		}
	}
	if intersection, ok := worstIntersection.Intersection(); ok {
		addIntersection(resultSet, flipped, intersection)
	}
}

// CheckIntersectionLineWithTriangle checks if a Line shape intersects with a
// Triangle shape.
func CheckIntersectionLineWithTriangle(line Line, triangle Triangle, flipped bool, resultSet IntersectionCollection) {
	normal := triangle.Normal()
	pointA := line.A()
	pointB := line.B()

	heightA := dprec.Vec3Dot(normal, dprec.Vec3Diff(pointA, triangle.A()))
	heightB := dprec.Vec3Dot(normal, dprec.Vec3Diff(pointB, triangle.A()))

	if (heightA > 0.0 && heightB > 0.0) || (heightA < 0.0 && heightB < 0.0) {
		return
	}
	if heightA < 0.0 {
		pointA, pointB = pointB, pointA
		heightA, heightB = heightB, heightA
	}

	projectedPoint := dprec.Vec3Sum(
		dprec.Vec3Prod(pointA, -heightB/(heightA-heightB)),
		dprec.Vec3Prod(pointB, heightA/(heightA-heightB)),
	)

	if triangle.ContainsPoint(projectedPoint) {
		addIntersection(resultSet, flipped, Intersection{
			Depth:                -heightB,
			FirstContact:         projectedPoint,
			FirstDisplaceNormal:  normal,
			SecondContact:        projectedPoint,
			SecondDisplaceNormal: dprec.InverseVec3(normal),
		})
	}
}

// CheckIntersectionBoxWithMesh checks if a Box shape intersects with a Mesh
// shape.
func CheckIntersectionBoxWithMesh(box Box, mesh Mesh, flipped bool, resultSet IntersectionCollection) {
	boxPosition := box.Position()
	boxRotation := box.Rotation()

	maxX := dprec.Vec3Prod(boxRotation.OrientationX(), box.HalfWidth())
	minX := dprec.InverseVec3(maxX)
	maxY := dprec.Vec3Prod(boxRotation.OrientationY(), box.HalfHeight())
	minY := dprec.InverseVec3(maxY)
	maxZ := dprec.Vec3Prod(boxRotation.OrientationZ(), box.HalfLength())
	minZ := dprec.InverseVec3(maxZ)

	p1 := dprec.Vec3Sum(dprec.Vec3Sum(dprec.Vec3Sum(boxPosition, minX), minZ), maxY)
	p2 := dprec.Vec3Sum(dprec.Vec3Sum(dprec.Vec3Sum(boxPosition, minX), maxZ), maxY)
	p3 := dprec.Vec3Sum(dprec.Vec3Sum(dprec.Vec3Sum(boxPosition, maxX), maxZ), maxY)
	p4 := dprec.Vec3Sum(dprec.Vec3Sum(dprec.Vec3Sum(boxPosition, maxX), minZ), maxY)
	p5 := dprec.Vec3Sum(dprec.Vec3Sum(dprec.Vec3Sum(boxPosition, minX), minZ), minY)
	p6 := dprec.Vec3Sum(dprec.Vec3Sum(dprec.Vec3Sum(boxPosition, minX), maxZ), minY)
	p7 := dprec.Vec3Sum(dprec.Vec3Sum(dprec.Vec3Sum(boxPosition, maxX), maxZ), minY)
	p8 := dprec.Vec3Sum(dprec.Vec3Sum(dprec.Vec3Sum(boxPosition, maxX), minZ), minY)

	for _, triangle := range mesh.Triangles() {
		CheckIntersectionLineWithTriangle(NewLine(p1, p2), triangle, flipped, resultSet)
		CheckIntersectionLineWithTriangle(NewLine(p2, p3), triangle, flipped, resultSet)
		CheckIntersectionLineWithTriangle(NewLine(p3, p4), triangle, flipped, resultSet)
		CheckIntersectionLineWithTriangle(NewLine(p4, p1), triangle, flipped, resultSet)

		CheckIntersectionLineWithTriangle(NewLine(p5, p6), triangle, flipped, resultSet)
		CheckIntersectionLineWithTriangle(NewLine(p6, p7), triangle, flipped, resultSet)
		CheckIntersectionLineWithTriangle(NewLine(p7, p8), triangle, flipped, resultSet)
		CheckIntersectionLineWithTriangle(NewLine(p8, p5), triangle, flipped, resultSet)

		CheckIntersectionLineWithTriangle(NewLine(p1, p5), triangle, flipped, resultSet)
		CheckIntersectionLineWithTriangle(NewLine(p2, p6), triangle, flipped, resultSet)
		CheckIntersectionLineWithTriangle(NewLine(p3, p7), triangle, flipped, resultSet)
		CheckIntersectionLineWithTriangle(NewLine(p4, p8), triangle, flipped, resultSet)
	}
}

// CheckIntersectionSphereWithTriangle checks if a Sphere shape intersects
// with a Triangle shape.
func CheckIntersectionSphereWithTriangle(sphere Sphere, triangle Triangle, flipped bool, resultSet IntersectionCollection) {
	spherePosition := sphere.Position()
	sphereRadius := sphere.Radius()
	triangleA := triangle.A()
	triangleB := triangle.B()
	triangleC := triangle.C()
	triangleCenter := triangle.Center()
	triangleNormal := triangle.Normal()

	sphereOffset := dprec.Vec3Diff(spherePosition, triangleCenter)
	height := dprec.Vec3Dot(triangleNormal, sphereOffset)
	if height > sphereRadius || height < 0 {
		return
	}

	vecAB := dprec.Vec3Diff(triangleB, triangleA)
	vecBC := dprec.Vec3Diff(triangleC, triangleB)
	vecCA := dprec.Vec3Diff(triangleA, triangleC)
	tangentAB := dprec.UnitVec3(vecAB)
	tangentBC := dprec.UnitVec3(vecBC)
	tangentCA := dprec.UnitVec3(vecCA)
	normAB := dprec.Vec3Cross(tangentAB, triangleNormal)
	normBC := dprec.Vec3Cross(tangentBC, triangleNormal)
	normCA := dprec.Vec3Cross(tangentCA, triangleNormal)

	projectedPoint := dprec.Vec3Diff(spherePosition, dprec.Vec3Prod(triangleNormal, height))
	vecAP := dprec.Vec3Diff(projectedPoint, triangleA)
	vecBP := dprec.Vec3Diff(projectedPoint, triangleB)
	vecCP := dprec.Vec3Diff(projectedPoint, triangleC)

	distAB := dprec.Vec3Dot(normAB, vecAP)
	distBC := dprec.Vec3Dot(normBC, vecBP)
	distCA := dprec.Vec3Dot(normCA, vecCP)

	var (
		inside    bool
		outsideAB bool
		outsideBC bool
		outsideCA bool
		outsideA  bool
		outsideB  bool
		outsideC  bool
	)
	switch {
	case distAB >= 0:
		if dprec.Vec3Dot(tangentAB, vecAP) >= 0 {
			if dprec.Vec3Dot(tangentAB, vecBP) <= 0 {
				outsideAB = true
			} else {
				outsideB = true
			}
		} else {
			outsideA = true
		}
	case distBC >= 0:
		if dprec.Vec3Dot(tangentBC, vecBP) >= 0 {
			if dprec.Vec3Dot(tangentBC, vecCP) <= 0 {
				outsideBC = true
			} else {
				outsideC = true
			}
		} else {
			outsideB = true
		}
	case distCA >= 0:
		if dprec.Vec3Dot(tangentCA, vecCP) >= 0 {
			if dprec.Vec3Dot(tangentCA, vecAP) <= 0 {
				outsideCA = true
			} else {
				outsideA = true
			}
		} else {
			outsideC = true
		}
	default:
		inside = true
	}

	var (
		isIntersection       bool
		depth                float64
		sphereDisplaceNormal dprec.Vec3
	)
	switch {
	// TODO: Recover all cases once the physics engine is fixed to check
	// collisions precisely (via binary search or similar).

	case outsideA:
	// 	cornerOffset := dprec.Vec3Diff(spherePosition, triangleA)
	// 	cornerDistance := cornerOffset.Length()
	// 	if isIntersection = (cornerDistance <= sphereRadius); isIntersection {
	// 		depth = sphereRadius - cornerDistance
	// 		sphereDisplaceNormal = dprec.Vec3Quot(cornerOffset, cornerDistance)
	// 	}

	case outsideB:
	// 	cornerOffset := dprec.Vec3Diff(spherePosition, triangleB)
	// 	cornerDistance := cornerOffset.Length()
	// 	if isIntersection = (cornerDistance <= sphereRadius); isIntersection {
	// 		depth = sphereRadius - cornerDistance
	// 		sphereDisplaceNormal = dprec.Vec3Quot(cornerOffset, cornerDistance)
	// 	}

	case outsideC:
	// 	cornerOffset := dprec.Vec3Diff(spherePosition, triangleC)
	// 	cornerDistance := cornerOffset.Length()
	// 	if isIntersection = (cornerDistance <= sphereRadius); isIntersection {
	// 		depth = sphereRadius - cornerDistance
	// 		sphereDisplaceNormal = dprec.Vec3Quot(cornerOffset, cornerDistance)
	// 	}

	case outsideAB:
	// 	edgeOffset := dprec.Vec3Sum(dprec.Vec3Prod(normAB, distAB), dprec.Vec3Prod(triangleNormal, height))
	// 	edgeDistance := edgeOffset.Length()
	// 	if isIntersection = (edgeDistance <= sphereRadius); isIntersection {
	// 		depth = sphereRadius - edgeDistance
	// 		sphereDisplaceNormal = dprec.Vec3Quot(edgeOffset, edgeDistance)
	// 	}

	case outsideBC:
	// 	edgeOffset := dprec.Vec3Sum(dprec.Vec3Prod(normBC, distBC), dprec.Vec3Prod(triangleNormal, height))
	// 	edgeDistance := edgeOffset.Length()
	// 	if isIntersection = (edgeDistance <= sphereRadius); isIntersection {
	// 		depth = sphereRadius - edgeDistance
	// 		sphereDisplaceNormal = dprec.Vec3Quot(edgeOffset, edgeDistance)
	// 	}

	case outsideCA:
	// 	edgeOffset := dprec.Vec3Sum(dprec.Vec3Prod(normCA, distCA), dprec.Vec3Prod(triangleNormal, height))
	// 	edgeDistance := edgeOffset.Length()
	// 	if isIntersection = (edgeDistance <= sphereRadius); isIntersection {
	// 		depth = sphereRadius - edgeDistance
	// 		sphereDisplaceNormal = dprec.Vec3Quot(edgeOffset, edgeDistance)
	// 	}

	case inside:
		isIntersection = true
		depth = sphereRadius - height
		sphereDisplaceNormal = triangleNormal
	}

	if isIntersection {
		addIntersection(resultSet, flipped, Intersection{
			Depth:                depth,
			FirstContact:         dprec.Vec3Diff(spherePosition, dprec.Vec3Prod(sphereDisplaceNormal, sphereRadius)),
			FirstDisplaceNormal:  sphereDisplaceNormal,
			SecondContact:        dprec.Vec3Diff(spherePosition, dprec.Vec3Prod(sphereDisplaceNormal, sphereRadius-depth)),
			SecondDisplaceNormal: dprec.InverseVec3(sphereDisplaceNormal),
		})
	}
}

// CheckIntersectionSphereWithSphere checks if a Sphere shape intersects with
// another Sphere shape.
func CheckIntersectionSphereWithSphere(first, second Sphere, flipped bool, resultSet IntersectionCollection) {
	firstPosition := first.Position()
	firstRadius := first.Radius()

	secondPosition := second.Position()
	secondRadius := second.Radius()

	deltaPosition := dprec.Vec3Diff(secondPosition, firstPosition)
	distance := deltaPosition.Length()
	if overlap := (firstRadius + secondRadius) - distance; overlap > 0.0 {
		secondDisplaceNormal := dprec.Vec3Quot(deltaPosition, distance) // unit vec
		firstDisplaceNormal := dprec.InverseVec3(secondDisplaceNormal)

		addIntersection(resultSet, false, Intersection{
			Depth: overlap,
			FirstContact: dprec.Vec3Sum(
				firstPosition,
				dprec.Vec3Prod(secondDisplaceNormal, firstRadius),
			),
			FirstDisplaceNormal: firstDisplaceNormal,
			SecondContact: dprec.Vec3Sum(
				secondPosition,
				dprec.Vec3Prod(firstDisplaceNormal, secondRadius),
			),
			SecondDisplaceNormal: secondDisplaceNormal,
		})
	}
}

// addIntersection is a helper function that adds an intersection to a result
// set and can flip it beforehand.
func addIntersection(resultSet IntersectionCollection, flipped bool, intersection Intersection) {
	if flipped {
		resultSet.AddIntersection(intersection.Flipped())
	} else {
		resultSet.AddIntersection(intersection)
	}
}
