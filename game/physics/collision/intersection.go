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
		SecondContact:        i.FirstContact,
		SecondDisplaceNormal: i.FirstDisplaceNormal,
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
func CheckIntersectionSetWithSet(first, second Set, resultSet IntersectionCollection) {
	if !IsSphereWithSphereIntersecting(first.BoundingSphere(), second.BoundingSphere()) {
		return
	}
	for _, firstSphere := range first.Spheres() {
		for _, secondSphere := range second.Spheres() {
			if IsSphereWithSphereIntersecting(firstSphere, secondSphere) {
				CheckIntersectionSphereWithSphere(firstSphere, secondSphere, false, resultSet)
			}
		}
		for _, secondBox := range second.Boxes() {
			if IsSphereWithSphereIntersecting(firstSphere, secondBox.BoundingSphere()) {
				CheckIntersectionSphereWithBox(firstSphere, secondBox, false, resultSet)
			}
		}
		for _, secondMesh := range second.Meshes() {
			if IsSphereWithSphereIntersecting(firstSphere, secondMesh.BoundingSphere()) {
				CheckIntersectionSphereWithMesh(firstSphere, secondMesh, false, resultSet)
			}
		}
	}
	for _, primaryBox := range first.Boxes() {
		for _, secondSphere := range second.Spheres() {
			if IsSphereWithSphereIntersecting(primaryBox.BoundingSphere(), secondSphere) {
				CheckIntersectionSphereWithBox(secondSphere, primaryBox, true, resultSet)
			}
		}
		for _, secondMesh := range second.Meshes() {
			if IsSphereWithSphereIntersecting(primaryBox.BoundingSphere(), secondMesh.BoundingSphere()) {
				CheckIntersectionBoxWithMesh(primaryBox, secondMesh, false, resultSet)
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

// CheckIntersectionSphereWithBox checks if a Sphere shape intersects with
// a Box shape.
func CheckIntersectionSphereWithBox(sphere Sphere, box Box, flipped bool, resultSet IntersectionCollection) {
	spherePosition := sphere.Position()
	sphereRadius := sphere.Radius()

	boxPosition := box.Position()
	boxRotation := box.Rotation()
	boxAxisX := boxRotation.OrientationX()
	boxAxisY := boxRotation.OrientationY()
	boxAxisZ := boxRotation.OrientationZ()
	boxHalfWidth := box.HalfWidth()
	boxHalfHeight := box.HalfHeight()
	boxHalfLength := box.HalfLength()

	deltaPosition := dprec.Vec3Diff(spherePosition, boxPosition)
	distanceX := dprec.Vec3Dot(boxAxisX, deltaPosition)
	distanceY := dprec.Vec3Dot(boxAxisY, deltaPosition)
	distanceZ := dprec.Vec3Dot(boxAxisZ, deltaPosition)

	distanceRight := distanceX - boxHalfWidth
	distanceLeft := -(2.0*boxHalfWidth + distanceRight)
	distanceTop := distanceY - boxHalfHeight
	distanceBottom := -(2.0*boxHalfHeight + distanceTop)
	distanceFront := distanceZ - boxHalfLength
	distanceBack := -(2.0*boxHalfLength + distanceFront)

	var (
		isIntersection    bool
		depth             float64
		boxContact        dprec.Vec3
		boxDisplaceNormal dprec.Vec3
	)

	const (
		maskLeft   = 0b100000
		maskRight  = 0b010000
		maskBottom = 0b001000
		maskTop    = 0b000100
		maskBack   = 0b000010
		maskFront  = 0b000001
	)
	var mask uint8
	if distanceLeft > 0 {
		mask |= maskLeft
	}
	if distanceRight > 0 {
		mask |= maskRight
	}
	if distanceBottom > 0 {
		mask |= maskBottom
	}
	if distanceTop > 0 {
		mask |= maskTop
	}
	if distanceBack > 0 {
		mask |= maskBack
	}
	if distanceFront > 0 {
		mask |= maskFront
	}
	switch mask {
	case maskLeft:
		if depth = sphereRadius - distanceLeft; depth > 0 {
			isIntersection = true
			boxDisplaceNormal = boxAxisX
			boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, sphereRadius-depth))
		}

	case maskRight:
		if depth = sphereRadius - distanceRight; depth > 0 {
			isIntersection = true
			boxDisplaceNormal = dprec.InverseVec3(boxAxisX)
			boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, sphereRadius-depth))
		}

	case maskBottom:
		if depth = sphereRadius - distanceBottom; depth > 0 {
			isIntersection = true
			boxDisplaceNormal = boxAxisY
			boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, sphereRadius-depth))
		}

	case maskTop:
		if depth = sphereRadius - distanceTop; depth > 0 {
			isIntersection = true
			boxDisplaceNormal = dprec.InverseVec3(boxAxisY)
			boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, sphereRadius-depth))
		}

	case maskBack:
		if depth = sphereRadius - distanceBack; depth > 0 {
			isIntersection = true
			boxDisplaceNormal = boxAxisZ
			boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, sphereRadius-depth))
		}

	case maskFront:
		if depth = sphereRadius - distanceFront; depth > 0 {
			isIntersection = true
			boxDisplaceNormal = dprec.InverseVec3(boxAxisZ)
			boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, sphereRadius-depth))
		}

	case maskLeft | maskBottom:
		sqrDistance := distanceLeft*distanceLeft + distanceBottom*distanceBottom
		if isIntersection = sqrDistance < sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxDisplaceNormal = dprec.UnitVec3(dprec.Vec3Sum(
				dprec.Vec3Prod(boxAxisX, distanceLeft),
				dprec.Vec3Prod(boxAxisY, distanceBottom),
			))
			boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, distance))
		}

	case maskLeft | maskTop:
		sqrDistance := distanceLeft*distanceLeft + distanceTop*distanceTop
		if isIntersection = sqrDistance < sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxDisplaceNormal = dprec.UnitVec3(dprec.Vec3Sum(
				dprec.Vec3Prod(boxAxisX, distanceLeft),
				dprec.Vec3Prod(boxAxisY, -distanceTop),
			))
			boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, distance))
		}

	case maskRight | maskBottom:
		sqrDistance := distanceRight*distanceRight + distanceBottom*distanceBottom
		if isIntersection = sqrDistance < sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxDisplaceNormal = dprec.UnitVec3(dprec.Vec3Sum(
				dprec.Vec3Prod(boxAxisX, -distanceRight),
				dprec.Vec3Prod(boxAxisY, distanceBottom),
			))
			boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, distance))
		}

	case maskRight | maskTop:
		sqrDistance := distanceRight*distanceRight + distanceTop*distanceTop
		if isIntersection = sqrDistance < sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxDisplaceNormal = dprec.UnitVec3(dprec.Vec3Sum(
				dprec.Vec3Prod(boxAxisX, -distanceRight),
				dprec.Vec3Prod(boxAxisY, -distanceTop),
			))
			boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, distance))
		}

	case maskBack | maskBottom:
		sqrDistance := distanceBack*distanceBack + distanceBottom*distanceBottom
		if isIntersection = sqrDistance < sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxDisplaceNormal = dprec.UnitVec3(dprec.Vec3Sum(
				dprec.Vec3Prod(boxAxisY, distanceBottom),
				dprec.Vec3Prod(boxAxisZ, distanceBack),
			))
			boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, distance))
		}

	case maskBack | maskTop:
		sqrDistance := distanceBack*distanceBack + distanceTop*distanceTop
		if isIntersection = sqrDistance < sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxDisplaceNormal = dprec.UnitVec3(dprec.Vec3Sum(
				dprec.Vec3Prod(boxAxisY, -distanceTop),
				dprec.Vec3Prod(boxAxisZ, distanceBack),
			))
			boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, distance))
		}

	case maskFront | maskBottom:
		sqrDistance := distanceFront*distanceFront + distanceBottom*distanceBottom
		if isIntersection = sqrDistance < sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxDisplaceNormal = dprec.UnitVec3(dprec.Vec3Sum(
				dprec.Vec3Prod(boxAxisY, distanceBottom),
				dprec.Vec3Prod(boxAxisZ, -distanceFront),
			))
			boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, distance))
		}

	case maskFront | maskTop:
		sqrDistance := distanceFront*distanceFront + distanceTop*distanceTop
		if isIntersection = sqrDistance < sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxDisplaceNormal = dprec.UnitVec3(dprec.Vec3Sum(
				dprec.Vec3Prod(boxAxisY, -distanceTop),
				dprec.Vec3Prod(boxAxisZ, -distanceFront),
			))
			boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, distance))
		}

	case maskBack | maskLeft:
		sqrDistance := distanceBack*distanceBack + distanceLeft*distanceLeft
		if isIntersection = sqrDistance < sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxDisplaceNormal = dprec.UnitVec3(dprec.Vec3Sum(
				dprec.Vec3Prod(boxAxisX, distanceLeft),
				dprec.Vec3Prod(boxAxisZ, distanceBack),
			))
			boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, distance))
		}

	case maskBack | maskRight:
		sqrDistance := distanceBack*distanceBack + distanceRight*distanceRight
		if isIntersection = sqrDistance < sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxDisplaceNormal = dprec.UnitVec3(dprec.Vec3Sum(
				dprec.Vec3Prod(boxAxisX, -distanceRight),
				dprec.Vec3Prod(boxAxisZ, distanceBack),
			))
			boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, distance))
		}

	case maskFront | maskLeft:
		sqrDistance := distanceFront*distanceFront + distanceLeft*distanceLeft
		if isIntersection = sqrDistance < sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxDisplaceNormal = dprec.UnitVec3(dprec.Vec3Sum(
				dprec.Vec3Prod(boxAxisX, distanceLeft),
				dprec.Vec3Prod(boxAxisZ, -distanceFront),
			))
			boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, distance))
		}

	case maskFront | maskRight:
		sqrDistance := distanceFront*distanceFront + distanceRight*distanceRight
		if isIntersection = sqrDistance < sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxDisplaceNormal = dprec.UnitVec3(dprec.Vec3Sum(
				dprec.Vec3Prod(boxAxisX, -distanceRight),
				dprec.Vec3Prod(boxAxisZ, -distanceFront),
			))
			boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, distance))
		}

	case maskLeft | maskBottom | maskBack:
		sqrDistance := distanceLeft*distanceLeft + distanceBottom*distanceBottom + distanceBack*distanceBack
		if isIntersection = sqrDistance < sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxContact = dprec.Vec3MultiSum(boxPosition, dprec.Vec3Prod(boxAxisX, -boxHalfWidth), dprec.Vec3Prod(boxAxisY, -boxHalfHeight), dprec.Vec3Prod(boxAxisZ, -boxHalfLength))
			boxDisplaceNormal = dprec.Vec3Quot(dprec.Vec3Diff(boxContact, spherePosition), distance)
		}

	case maskLeft | maskBottom | maskFront:
		sqrDistance := distanceLeft*distanceLeft + distanceBottom*distanceBottom + distanceFront*distanceFront
		if isIntersection = sqrDistance < sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxContact = dprec.Vec3MultiSum(boxPosition, dprec.Vec3Prod(boxAxisX, -boxHalfWidth), dprec.Vec3Prod(boxAxisY, -boxHalfHeight), dprec.Vec3Prod(boxAxisZ, boxHalfLength))
			boxDisplaceNormal = dprec.Vec3Quot(dprec.Vec3Diff(boxContact, spherePosition), distance)
		}

	case maskLeft | maskTop | maskBack:
		sqrDistance := distanceLeft*distanceLeft + distanceTop*distanceTop + distanceBack*distanceBack
		if isIntersection = sqrDistance < sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxContact = dprec.Vec3MultiSum(boxPosition, dprec.Vec3Prod(boxAxisX, -boxHalfWidth), dprec.Vec3Prod(boxAxisY, boxHalfHeight), dprec.Vec3Prod(boxAxisZ, -boxHalfLength))
			boxDisplaceNormal = dprec.Vec3Quot(dprec.Vec3Diff(boxContact, spherePosition), distance)
		}

	case maskLeft | maskTop | maskFront:
		sqrDistance := distanceLeft*distanceLeft + distanceTop*distanceTop + distanceFront*distanceFront
		if isIntersection = sqrDistance < sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxContact = dprec.Vec3MultiSum(boxPosition, dprec.Vec3Prod(boxAxisX, -boxHalfWidth), dprec.Vec3Prod(boxAxisY, boxHalfHeight), dprec.Vec3Prod(boxAxisZ, boxHalfLength))
			boxDisplaceNormal = dprec.Vec3Quot(dprec.Vec3Diff(boxContact, spherePosition), distance)
		}

	case maskRight | maskBottom | maskBack:
		sqrDistance := distanceRight*distanceRight + distanceBottom*distanceBottom + distanceBack*distanceBack
		if isIntersection = sqrDistance < sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxContact = dprec.Vec3MultiSum(boxPosition, dprec.Vec3Prod(boxAxisX, boxHalfWidth), dprec.Vec3Prod(boxAxisY, -boxHalfHeight), dprec.Vec3Prod(boxAxisZ, -boxHalfLength))
			boxDisplaceNormal = dprec.Vec3Quot(dprec.Vec3Diff(boxContact, spherePosition), distance)
		}

	case maskRight | maskBottom | maskFront:
		sqrDistance := distanceRight*distanceRight + distanceBottom*distanceBottom + distanceFront*distanceFront
		if isIntersection = sqrDistance < sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxContact = dprec.Vec3MultiSum(boxPosition, dprec.Vec3Prod(boxAxisX, boxHalfWidth), dprec.Vec3Prod(boxAxisY, -boxHalfHeight), dprec.Vec3Prod(boxAxisZ, boxHalfLength))
			boxDisplaceNormal = dprec.Vec3Quot(dprec.Vec3Diff(boxContact, spherePosition), distance)
		}

	case maskRight | maskTop | maskBack:
		sqrDistance := distanceRight*distanceRight + distanceTop*distanceTop + distanceBack*distanceBack
		if isIntersection = sqrDistance < sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxContact = dprec.Vec3MultiSum(boxPosition, dprec.Vec3Prod(boxAxisX, boxHalfWidth), dprec.Vec3Prod(boxAxisY, boxHalfHeight), dprec.Vec3Prod(boxAxisZ, -boxHalfLength))
			boxDisplaceNormal = dprec.Vec3Quot(dprec.Vec3Diff(boxContact, spherePosition), distance)
		}

	case maskRight | maskTop | maskFront:
		sqrDistance := distanceRight*distanceRight + distanceTop*distanceTop + distanceFront*distanceFront
		if isIntersection = sqrDistance < sphereRadius*sphereRadius; isIntersection {
			distance := dprec.Sqrt(sqrDistance)
			depth = sphereRadius - distance
			boxContact = dprec.Vec3MultiSum(boxPosition, dprec.Vec3Prod(boxAxisX, boxHalfWidth), dprec.Vec3Prod(boxAxisY, boxHalfHeight), dprec.Vec3Prod(boxAxisZ, boxHalfLength))
			boxDisplaceNormal = dprec.Vec3Quot(dprec.Vec3Diff(boxContact, spherePosition), distance)
		}

	default:
		// Note: This branch is unlikely to occur so no need to be extremely optimal.
		isIntersection = true
		var (
			displaceX float64
			displaceY float64
			displaceZ float64
		)
		if distanceLeft > distanceRight {
			displaceX = distanceLeft
		} else {
			displaceX = -distanceRight
		}
		if distanceBottom > distanceTop {
			displaceY = distanceBottom
		} else {
			displaceY = -distanceTop
		}
		if distanceBack > distanceFront {
			displaceZ = distanceBack
		} else {
			displaceZ = -distanceFront
		}
		if dprec.Abs(displaceX) < dprec.Abs(displaceY) {
			if dprec.Abs(displaceX) < dprec.Abs(displaceZ) {
				depth = dprec.Abs(displaceX) + sphereRadius
				boxDisplaceNormal = dprec.Vec3Prod(boxAxisX, -dprec.Sign(displaceX))
			} else {
				depth = dprec.Abs(displaceZ) + sphereRadius
				boxDisplaceNormal = dprec.Vec3Prod(boxAxisZ, -dprec.Sign(displaceZ))
			}
		} else {
			if dprec.Abs(displaceY) < dprec.Abs(displaceZ) {
				depth = dprec.Abs(displaceY) + sphereRadius
				boxDisplaceNormal = dprec.Vec3Prod(boxAxisY, -dprec.Sign(displaceY))
			} else {
				depth = dprec.Abs(displaceZ) + sphereRadius
				boxDisplaceNormal = dprec.Vec3Prod(boxAxisZ, -dprec.Sign(displaceZ))
			}
		}
		boxContact = dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, sphereRadius-depth))
	}

	if isIntersection {
		addIntersection(resultSet, flipped, Intersection{
			Depth:                depth,
			FirstContact:         dprec.Vec3Sum(spherePosition, dprec.Vec3Prod(boxDisplaceNormal, sphereRadius)),
			FirstDisplaceNormal:  dprec.InverseVec3(boxDisplaceNormal),
			SecondContact:        boxContact,
			SecondDisplaceNormal: boxDisplaceNormal,
		})
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

// CheckIntersectionLineWithMesh checks if a Line shape intersects with a Mesh
// shape.
func CheckIntersectionLineWithMesh(line Line, mesh Mesh, flipped bool, resultSet IntersectionCollection) {
	for _, triangle := range mesh.Triangles() {
		CheckIntersectionLineWithTriangle(line, triangle, flipped, resultSet)
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

// LineWithSurfaceIntersectionPoint returns the point where a line intersects
// with a surface, if there is one.
func LineWithSurfaceIntersectionPoint(line Line, point, normal dprec.Vec3) (dprec.Vec3, bool) {
	h1 := dprec.Vec3Dot(normal, dprec.Vec3Diff(line.A(), point))
	h2 := dprec.Vec3Dot(normal, dprec.Vec3Diff(point, line.B()))
	if h1 > 0.0 && h2 < 0.0 {
		return dprec.Vec3{}, false
	}
	if h1 < 0.0 && h2 > 0.0 {
		return dprec.Vec3{}, false
	}
	return dprec.Vec3Sum(
		dprec.Vec3Prod(line.A(), h2/(h1+h2)),
		dprec.Vec3Prod(line.B(), h1/(h1+h2)),
	), true
}

// LineWithSphereIntersectionPoints returns the two intersection points of
// a line with a sphere, if there are such.
func LineWithSphereIntersectionPoints(line Line, sphere Sphere) (dprec.Vec3, dprec.Vec3, bool) {
	tangent := dprec.UnitVec3(dprec.Vec3Diff(line.A(), line.B()))
	h1 := dprec.Vec3Dot(tangent, dprec.Vec3Diff(line.A(), sphere.Position()))
	h2 := dprec.Vec3Dot(tangent, dprec.Vec3Diff(sphere.Position(), line.B()))
	point := dprec.Vec3Sum(
		dprec.Vec3Prod(line.A(), h2/(h1+h2)),
		dprec.Vec3Prod(line.B(), h1/(h1+h2)),
	)

	diff := dprec.Vec3Diff(point, sphere.Position())
	height := diff.Length()
	r := sphere.Radius()
	if height >= r {
		return dprec.Vec3{}, dprec.Vec3{}, false
	}
	shift := dprec.Sqrt(r*r - height*height)

	var first dprec.Vec3
	if h1 > shift {
		first = dprec.Vec3Sum(
			dprec.Vec3Prod(line.A(), (h2+shift)/(h1+h2)),
			dprec.Vec3Prod(line.B(), (h1-shift)/(h1+h2)),
		)
	} else {
		first = line.A()
	}

	var second dprec.Vec3
	if h2 > shift {
		second = dprec.Vec3Sum(
			dprec.Vec3Prod(line.A(), (h2-shift)/(h1+h2)),
			dprec.Vec3Prod(line.B(), (h1+shift)/(h1+h2)),
		)
	} else {
		second = line.B()
	}

	return first, second, true
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
