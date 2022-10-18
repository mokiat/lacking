package shape

import "github.com/mokiat/gomath/dprec"

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
func (i Intersection) Flipped() Intersection {
	i.FirstContact, i.SecondContact = i.SecondContact, i.FirstContact
	i.FirstDisplaceNormal, i.SecondDisplaceNormal = i.SecondDisplaceNormal, i.FirstDisplaceNormal
	return i
}

// NewIntersectionResultSet creates a new IntersectionResultSet instance with
// the specified initial capacity. The set is allowed to grow past the
// specified capacity.
func NewIntersectionResultSet(capacity int) *IntersectionResultSet {
	return &IntersectionResultSet{
		intersections: make([]Intersection, 0, capacity),
		flipped:       false,
	}
}

// IntersectionResultSet is a structure that can be used to collection the
// result of an intersection test.
type IntersectionResultSet struct {
	intersections []Intersection
	flipped       bool
}

// Reset clears the buffer of this result set so that it can be reused.
func (s *IntersectionResultSet) Reset() {
	s.intersections = s.intersections[:0]
}

// AddFlipped controls whether newly added Intersections should be flipped
// beforehand.
func (s *IntersectionResultSet) AddFlipped(flipped bool) {
	// TODO: This method is used only internally and should be removed.
	s.flipped = flipped
}

// Add adds a new Intersection to this set.
func (s *IntersectionResultSet) Add(intersection Intersection) {
	if s.flipped {
		s.intersections = append(s.intersections, intersection.Flipped())
	} else {
		s.intersections = append(s.intersections, intersection)
	}
}

// Found returns whether this set contains any intersections.
func (s *IntersectionResultSet) Found() bool {
	return len(s.intersections) > 0
}

// Intersections returns a list of all intersections that have been found.
// The returned slice must not be modified or cached as it will be reused.
func (s *IntersectionResultSet) Intersections() []Intersection {
	return s.intersections
}

// CheckIntersection checks whether the two shapes intersect.
func CheckIntersection(first, second Shape, resultSet *IntersectionResultSet) {
	var firstPlacement Placement
	if placement, ok := first.(Placement); ok {
		firstPlacement = placement
	} else {
		firstPlacement = NewIdentityPlacement(first)
	}
	var secondPlacement Placement
	if placement, ok := second.(Placement); ok {
		secondPlacement = placement
	} else {
		secondPlacement = NewIdentityPlacement(second)
	}
	checkIntersectionPlacements(firstPlacement, secondPlacement, resultSet)
}

func checkIntersectionPlacements(first, second Placement, resultSet *IntersectionResultSet) {
	switch first.shape.(type) {
	case StaticSphere:
		checkIntersectionSphereUnknownPlacements(first, second, resultSet)
	case StaticBox:
		checkIntersectionBoxUnknownPlacements(first, second, resultSet)
	case StaticMesh:
		checkIntersectionMeshUnknownPlacements(first, second, resultSet)
	}
}

func checkIntersectionSphereUnknownPlacements(first, second Placement, resultSet *IntersectionResultSet) {
	switch second.shape.(type) {
	case StaticSphere:
		resultSet.AddFlipped(false)
		checkIntersectionSphereSpherePlacements(first, second, resultSet)
	case StaticBox:
		resultSet.AddFlipped(false)
		checkIntersectionSphereBoxPlacements(first, second, resultSet)
	case StaticMesh:
		resultSet.AddFlipped(false)
		checkIntersectionSphereMeshPlacements(first, second, resultSet)
	}
}

func checkIntersectionBoxUnknownPlacements(first, second Placement, resultSet *IntersectionResultSet) {
	switch second.shape.(type) {
	case StaticSphere:
		resultSet.AddFlipped(true)
		checkIntersectionSphereBoxPlacements(second, first, resultSet)
	case StaticBox:
		resultSet.AddFlipped(false)
		checkIntersectionBoxBoxPlacements(first, second, resultSet)
	case StaticMesh:
		resultSet.AddFlipped(false)
		checkIntersectionBoxMeshPlacements(first, second, resultSet)
	}
}

func checkIntersectionMeshUnknownPlacements(first, second Placement, resultSet *IntersectionResultSet) {
	switch second.shape.(type) {
	case StaticSphere:
		resultSet.AddFlipped(true)
		checkIntersectionSphereMeshPlacements(second, first, resultSet)
	case StaticBox:
		resultSet.AddFlipped(true)
		checkIntersectionBoxMeshPlacements(second, first, resultSet)
	case StaticMesh:
		resultSet.AddFlipped(false)
		checkIntersectionMeshMeshPlacements(first, second, resultSet)
	}
}

func checkIntersectionSphereSpherePlacements(first, second Placement, resultSet *IntersectionResultSet) {
	firstSphere := first.shape.(StaticSphere)
	secondSphere := second.shape.(StaticSphere)

	deltaPosition := dprec.Vec3Diff(second.position, first.position)
	overlap := firstSphere.Radius() + secondSphere.Radius() - deltaPosition.Length()
	if overlap <= 0 {
		return
	}

	secondDisplaceNormal := dprec.UnitVec3(deltaPosition)
	firstDisplaceNormal := dprec.InverseVec3(secondDisplaceNormal)

	resultSet.Add(Intersection{
		Depth: overlap,
		FirstContact: dprec.Vec3Sum(
			first.position,
			dprec.Vec3Prod(secondDisplaceNormal, firstSphere.radius),
		),
		FirstDisplaceNormal: firstDisplaceNormal,
		SecondContact: dprec.Vec3Sum(
			second.position,
			dprec.Vec3Prod(firstDisplaceNormal, secondSphere.radius),
		),
		SecondDisplaceNormal: secondDisplaceNormal,
	})
}

func checkIntersectionSphereBoxPlacements(first, second Placement, resultSet *IntersectionResultSet) {
}

func checkIntersectionSphereMeshPlacements(spherePlacement, meshPlacement Placement, resultSet *IntersectionResultSet) {
	sphere := spherePlacement.shape.(StaticSphere)
	mesh := meshPlacement.shape.(StaticMesh)

	// broad phase
	deltaPosition := dprec.Vec3Diff(meshPlacement.position, spherePlacement.position)
	if deltaPosition.Length() > sphere.Radius()+mesh.BoundingSphereRadius() {
		return
	}

	// narrow phase
	for _, triangle := range mesh.Triangles() {
		triangleWS := triangle.Transformed(meshPlacement.position, meshPlacement.rotation)
		height := dprec.Vec3Dot(triangleWS.Normal(), dprec.Vec3Diff(spherePlacement.position, triangleWS.A()))
		if dprec.Abs(height) > sphere.Radius() {
			continue
		}

		projectedPoint := dprec.Vec3Diff(spherePlacement.position, dprec.Vec3Prod(triangle.Normal(), height))
		if triangleWS.ContainsPoint(projectedPoint) {
			depth := sphere.Radius() - dprec.Abs(height)
			resultSet.Add(Intersection{
				Depth:                depth,
				FirstContact:         dprec.Vec3Sum(projectedPoint, dprec.Vec3Prod(triangle.Normal(), -depth)),
				FirstDisplaceNormal:  triangle.Normal(),
				SecondContact:        projectedPoint,
				SecondDisplaceNormal: dprec.InverseVec3(triangle.Normal()),
			})
			// TODO: Handle cases where the point is not contained but the sphere touches the edge of the triangle
		}
	}
}

var (
	boxTriangles = make([]StaticTriangle, 2*6)
)

func checkIntersectionBoxBoxPlacements(first, second Placement, resultSet *IntersectionResultSet) {
	box := second.shape.(StaticBox)
	halfWidth := box.Width() / 2.0
	halfHeight := box.Height() / 2.0
	halfLength := box.Length() / 2.0

	// TOP
	boxTriangles[0] = StaticTriangle{
		a: dprec.NewVec3(-halfWidth, halfHeight, -halfLength),
		b: dprec.NewVec3(-halfWidth, halfHeight, halfLength),
		c: dprec.NewVec3(halfWidth, halfHeight, halfLength),
	}
	boxTriangles[1] = StaticTriangle{
		a: dprec.NewVec3(-halfWidth, halfHeight, -halfLength),
		b: dprec.NewVec3(halfWidth, halfHeight, halfLength),
		c: dprec.NewVec3(halfWidth, halfHeight, -halfLength),
	}

	// BOTTOM
	boxTriangles[2] = StaticTriangle{
		a: dprec.NewVec3(-halfWidth, -halfHeight, -halfLength),
		b: dprec.NewVec3(halfWidth, -halfHeight, halfLength),
		c: dprec.NewVec3(-halfWidth, -halfHeight, halfLength),
	}
	boxTriangles[3] = StaticTriangle{
		a: dprec.NewVec3(-halfWidth, -halfHeight, -halfLength),
		b: dprec.NewVec3(halfWidth, -halfHeight, -halfLength),
		c: dprec.NewVec3(halfWidth, -halfHeight, halfLength),
	}

	// FRONT
	boxTriangles[4] = StaticTriangle{
		a: dprec.NewVec3(-halfWidth, halfHeight, halfLength),
		b: dprec.NewVec3(-halfWidth, -halfHeight, halfLength),
		c: dprec.NewVec3(halfWidth, -halfHeight, halfLength),
	}
	boxTriangles[5] = StaticTriangle{
		a: dprec.NewVec3(-halfWidth, halfHeight, halfLength),
		b: dprec.NewVec3(halfWidth, -halfHeight, halfLength),
		c: dprec.NewVec3(halfWidth, halfHeight, halfLength),
	}

	// REAR
	boxTriangles[6] = StaticTriangle{
		a: dprec.NewVec3(-halfWidth, halfHeight, -halfLength),
		b: dprec.NewVec3(halfWidth, -halfHeight, -halfLength),
		c: dprec.NewVec3(-halfWidth, -halfHeight, -halfLength),
	}
	boxTriangles[7] = StaticTriangle{
		a: dprec.NewVec3(-halfWidth, halfHeight, -halfLength),
		b: dprec.NewVec3(halfWidth, halfHeight, -halfLength),
		c: dprec.NewVec3(halfWidth, -halfHeight, -halfLength),
	}

	// LEFT
	boxTriangles[8] = StaticTriangle{
		a: dprec.NewVec3(-halfWidth, halfHeight, -halfLength),
		b: dprec.NewVec3(-halfWidth, -halfHeight, -halfLength),
		c: dprec.NewVec3(-halfWidth, -halfHeight, halfLength),
	}
	boxTriangles[9] = StaticTriangle{
		a: dprec.NewVec3(-halfWidth, halfHeight, -halfLength),
		b: dprec.NewVec3(-halfWidth, -halfHeight, halfLength),
		c: dprec.NewVec3(-halfWidth, halfHeight, halfLength),
	}

	// RIGHT
	boxTriangles[10] = StaticTriangle{
		a: dprec.NewVec3(halfWidth, halfHeight, -halfLength),
		b: dprec.NewVec3(halfWidth, -halfHeight, halfLength),
		c: dprec.NewVec3(halfWidth, -halfHeight, -halfLength),
	}
	boxTriangles[11] = StaticTriangle{
		a: dprec.NewVec3(halfWidth, halfHeight, -halfLength),
		b: dprec.NewVec3(halfWidth, halfHeight, halfLength),
		c: dprec.NewVec3(halfWidth, -halfHeight, halfLength),
	}

	second = Placement{
		shape:    NewStaticMesh(boxTriangles),
		position: second.position,
		rotation: second.rotation,
	}
	checkIntersectionBoxMeshPlacements(first, second, resultSet)
}

func checkIntersectionBoxMeshPlacements(boxPlacement, meshPlacement Placement, resultSet *IntersectionResultSet) {
	box := boxPlacement.shape.(StaticBox)
	mesh := meshPlacement.shape.(StaticMesh)

	// broad phase
	deltaPosition := dprec.Vec3Diff(meshPlacement.position, boxPlacement.position)
	boxBoundingSphereRadius := dprec.Sqrt(box.Width()*box.Width()+box.Height()*box.Height()+box.Length()*box.Length()) / 2.0
	if deltaPosition.Length() > boxBoundingSphereRadius+mesh.BoundingSphereRadius() {
		return
	}

	minX := dprec.Vec3Prod(boxPlacement.rotation.OrientationX(), -box.Width()/2.0)
	maxX := dprec.Vec3Prod(boxPlacement.rotation.OrientationX(), box.Width()/2.0)
	minY := dprec.Vec3Prod(boxPlacement.rotation.OrientationY(), -box.Height()/2.0)
	maxY := dprec.Vec3Prod(boxPlacement.rotation.OrientationY(), box.Height()/2.0)
	minZ := dprec.Vec3Prod(boxPlacement.rotation.OrientationZ(), -box.Length()/2.0)
	maxZ := dprec.Vec3Prod(boxPlacement.rotation.OrientationZ(), box.Length()/2.0)

	p1 := dprec.Vec3Sum(dprec.Vec3Sum(dprec.Vec3Sum(boxPlacement.position, minX), minZ), maxY)
	p2 := dprec.Vec3Sum(dprec.Vec3Sum(dprec.Vec3Sum(boxPlacement.position, minX), maxZ), maxY)
	p3 := dprec.Vec3Sum(dprec.Vec3Sum(dprec.Vec3Sum(boxPlacement.position, maxX), maxZ), maxY)
	p4 := dprec.Vec3Sum(dprec.Vec3Sum(dprec.Vec3Sum(boxPlacement.position, maxX), minZ), maxY)
	p5 := dprec.Vec3Sum(dprec.Vec3Sum(dprec.Vec3Sum(boxPlacement.position, minX), minZ), minY)
	p6 := dprec.Vec3Sum(dprec.Vec3Sum(dprec.Vec3Sum(boxPlacement.position, minX), maxZ), minY)
	p7 := dprec.Vec3Sum(dprec.Vec3Sum(dprec.Vec3Sum(boxPlacement.position, maxX), maxZ), minY)
	p8 := dprec.Vec3Sum(dprec.Vec3Sum(dprec.Vec3Sum(boxPlacement.position, maxX), minZ), minY)

	checkLineIntersection := func(line StaticLine, triangle StaticTriangle) {
		heightA := dprec.Vec3Dot(triangle.Normal(), dprec.Vec3Diff(line.A(), triangle.A()))
		heightB := dprec.Vec3Dot(triangle.Normal(), dprec.Vec3Diff(line.B(), triangle.A()))
		if heightA > 0.0 && heightB > 0.0 {
			return
		}
		if heightA < 0.0 && heightB < 0.0 {
			return
		}
		if heightA < 0.0 {
			line = NewStaticLine(line.B(), line.A())
			heightA, heightB = heightB, heightA
		}

		projectedPoint := dprec.Vec3Sum(
			dprec.Vec3Prod(line.A(), -heightB/(heightA-heightB)),
			dprec.Vec3Prod(line.B(), heightA/(heightA-heightB)),
		)

		if triangle.ContainsPoint(projectedPoint) {
			resultSet.Add(Intersection{
				Depth:                -heightB,
				FirstContact:         projectedPoint,
				FirstDisplaceNormal:  triangle.Normal(),
				SecondContact:        projectedPoint,
				SecondDisplaceNormal: triangle.Normal(),
			})
		}
	}

	// narrow phase
	for _, triangle := range mesh.Triangles() {
		triangleWS := triangle.Transformed(meshPlacement.position, meshPlacement.rotation)
		checkLineIntersection(NewStaticLine(p1, p2), triangleWS)
		checkLineIntersection(NewStaticLine(p2, p3), triangleWS)
		checkLineIntersection(NewStaticLine(p3, p4), triangleWS)
		checkLineIntersection(NewStaticLine(p4, p1), triangleWS)

		checkLineIntersection(NewStaticLine(p5, p6), triangleWS)
		checkLineIntersection(NewStaticLine(p6, p7), triangleWS)
		checkLineIntersection(NewStaticLine(p7, p8), triangleWS)
		checkLineIntersection(NewStaticLine(p8, p5), triangleWS)

		checkLineIntersection(NewStaticLine(p1, p5), triangleWS)
		checkLineIntersection(NewStaticLine(p2, p6), triangleWS)
		checkLineIntersection(NewStaticLine(p3, p7), triangleWS)
		checkLineIntersection(NewStaticLine(p4, p8), triangleWS)
	}
}

func checkIntersectionMeshMeshPlacements(first, second Placement, resultSet *IntersectionResultSet) {
}

func CheckLineIntersection(line StaticLine, mesh StaticMesh, resultSet *IntersectionResultSet) {
	for _, triangle := range mesh.Triangles() {
		heightA := dprec.Vec3Dot(triangle.Normal(), dprec.Vec3Diff(line.A(), triangle.A()))
		heightB := dprec.Vec3Dot(triangle.Normal(), dprec.Vec3Diff(line.B(), triangle.A()))
		if heightA > 0.0 && heightB > 0.0 {
			return
		}
		if heightA < 0.0 && heightB < 0.0 {
			return
		}
		if heightA < 0.0 {
			line = NewStaticLine(line.B(), line.A())
			heightA, heightB = heightB, heightA
		}

		projectedPoint := dprec.Vec3Sum(
			dprec.Vec3Prod(line.A(), -heightB/(heightA-heightB)),
			dprec.Vec3Prod(line.B(), heightA/(heightA-heightB)),
		)

		if triangle.ContainsPoint(projectedPoint) {
			resultSet.Add(Intersection{
				Depth:                -heightB,
				FirstContact:         projectedPoint,
				FirstDisplaceNormal:  triangle.Normal(),
				SecondContact:        projectedPoint,
				SecondDisplaceNormal: dprec.InverseVec3(triangle.Normal()),
			})
		}
	}
}
