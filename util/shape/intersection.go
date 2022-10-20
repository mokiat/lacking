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
	CheckPlacementIntersection(
		NewPlacement(IdentityTransform(), first),
		NewPlacement(IdentityTransform(), second),
		resultSet,
	)
}

// CheckPlacementIntersection checks whether the two shapes contained in
// Placement objects intersect.
//
// This is an alternative to CheckIntersection that can be used when the
// shapes have a dynamic outside transform being applied to them.
func CheckPlacementIntersection(first, second Placement, resultSet *IntersectionResultSet) {
	switch firstShape := first.shape.(type) {
	case StaticSphere:
		firstShape = firstShape.Transformed(first.transform)
		checkIntersectionSphereWithUnknown(firstShape, second, resultSet)
	case StaticBox:
		firstShape = firstShape.Transformed(first.transform)
		checkIntersectionBoxWithUnknown(firstShape, second, resultSet)
	case StaticMesh:
		firstShape = firstShape.Transformed(first.transform)
		checkIntersectionMeshWithUnknown(firstShape, second, resultSet)
	}
}

func checkIntersectionSphereWithUnknown(first StaticSphere, second Placement, resultSet *IntersectionResultSet) {
	switch secondShape := second.shape.(type) {
	case StaticSphere:
		resultSet.AddFlipped(false)
		secondShape = secondShape.Transformed(second.transform)
		checkIntersectionSphereWithSphere(first, secondShape, resultSet)
	case StaticBox:
		resultSet.AddFlipped(false)
		secondShape = secondShape.Transformed(second.transform)
		checkIntersectionSphereWithBox(first, secondShape, resultSet)
	case StaticMesh:
		resultSet.AddFlipped(false)
		secondShape = secondShape.Transformed(second.transform)
		checkIntersectionSphereWithMesh(first, secondShape, resultSet)
	}
}

func checkIntersectionBoxWithUnknown(first StaticBox, second Placement, resultSet *IntersectionResultSet) {
	switch secondShape := second.shape.(type) {
	case StaticSphere:
		resultSet.AddFlipped(true)
		secondShape = secondShape.Transformed(second.transform)
		checkIntersectionSphereWithBox(secondShape, first, resultSet)
	case StaticBox:
		resultSet.AddFlipped(false)
		secondShape = secondShape.Transformed(second.transform)
		checkIntersectionBoxWithBox(first, secondShape, resultSet)
	case StaticMesh:
		resultSet.AddFlipped(false)
		secondShape = secondShape.Transformed(second.transform)
		checkIntersectionBoxWithMesh(first, secondShape, resultSet)
	}
}

func checkIntersectionMeshWithUnknown(first StaticMesh, second Placement, resultSet *IntersectionResultSet) {
	switch secondShape := second.shape.(type) {
	case StaticSphere:
		resultSet.AddFlipped(true)
		secondShape = secondShape.Transformed(second.transform)
		checkIntersectionSphereWithMesh(secondShape, first, resultSet)
	case StaticBox:
		resultSet.AddFlipped(true)
		secondShape = secondShape.Transformed(second.transform)
		checkIntersectionBoxWithMesh(secondShape, first, resultSet)
	case StaticMesh:
		resultSet.AddFlipped(false)
		secondShape = secondShape.Transformed(second.transform)
		checkIntersectionMeshWithMesh(first, secondShape, resultSet)
	}
}

func checkIntersectionSphereWithSphere(first, second StaticSphere, resultSet *IntersectionResultSet) {
	deltaPosition := dprec.Vec3Diff(second.Position(), first.Position())
	overlap := first.Radius() + second.Radius() - deltaPosition.Length()
	if overlap <= 0 {
		return
	}

	secondDisplaceNormal := dprec.UnitVec3(deltaPosition)
	firstDisplaceNormal := dprec.InverseVec3(secondDisplaceNormal)

	resultSet.Add(Intersection{
		Depth: overlap,
		FirstContact: dprec.Vec3Sum(
			first.Position(),
			dprec.Vec3Prod(secondDisplaceNormal, first.Radius()),
		),
		FirstDisplaceNormal: firstDisplaceNormal,
		SecondContact: dprec.Vec3Sum(
			second.Position(),
			dprec.Vec3Prod(firstDisplaceNormal, second.Radius()),
		),
		SecondDisplaceNormal: secondDisplaceNormal,
	})
}

func checkIntersectionSphereWithBox(first StaticSphere, second StaticBox, resultSet *IntersectionResultSet) {
}

func checkIntersectionSphereWithMesh(sphere StaticSphere, mesh StaticMesh, resultSet *IntersectionResultSet) {
	// broad phase
	deltaPosition := dprec.Vec3Diff(mesh.Position(), sphere.Position())
	if deltaPosition.Length() > sphere.Radius()+mesh.BoundingSphereRadius() {
		return
	}

	// narrow phase
	for _, triangle := range mesh.Triangles() {
		triangle = triangle.Transformed(mesh.Transform)

		height := dprec.Vec3Dot(triangle.Normal(), dprec.Vec3Diff(sphere.Position(), dprec.Vec3(triangle.A())))
		if dprec.Abs(height) > sphere.Radius() {
			continue
		}

		projectedPoint := dprec.Vec3Diff(sphere.Position(), dprec.Vec3Prod(triangle.Normal(), height))
		if triangle.ContainsPoint(Point(projectedPoint)) {
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

func checkIntersectionBoxWithBox(first, second StaticBox, resultSet *IntersectionResultSet) {
	halfWidth := second.HalfWidth()
	halfHeight := second.HalfHeight()
	halfLength := second.HalfLength()

	// TOP
	boxTriangles[0] = StaticTriangle{
		a: Point(dprec.NewVec3(-halfWidth, halfHeight, -halfLength)),
		b: Point(dprec.NewVec3(-halfWidth, halfHeight, halfLength)),
		c: Point(dprec.NewVec3(halfWidth, halfHeight, halfLength)),
	}
	boxTriangles[1] = StaticTriangle{
		a: Point(dprec.NewVec3(-halfWidth, halfHeight, -halfLength)),
		b: Point(dprec.NewVec3(halfWidth, halfHeight, halfLength)),
		c: Point(dprec.NewVec3(halfWidth, halfHeight, -halfLength)),
	}

	// BOTTOM
	boxTriangles[2] = StaticTriangle{
		a: Point(dprec.NewVec3(-halfWidth, -halfHeight, -halfLength)),
		b: Point(dprec.NewVec3(halfWidth, -halfHeight, halfLength)),
		c: Point(dprec.NewVec3(-halfWidth, -halfHeight, halfLength)),
	}
	boxTriangles[3] = StaticTriangle{
		a: Point(dprec.NewVec3(-halfWidth, -halfHeight, -halfLength)),
		b: Point(dprec.NewVec3(halfWidth, -halfHeight, -halfLength)),
		c: Point(dprec.NewVec3(halfWidth, -halfHeight, halfLength)),
	}

	// FRONT
	boxTriangles[4] = StaticTriangle{
		a: Point(dprec.NewVec3(-halfWidth, halfHeight, halfLength)),
		b: Point(dprec.NewVec3(-halfWidth, -halfHeight, halfLength)),
		c: Point(dprec.NewVec3(halfWidth, -halfHeight, halfLength)),
	}
	boxTriangles[5] = StaticTriangle{
		a: Point(dprec.NewVec3(-halfWidth, halfHeight, halfLength)),
		b: Point(dprec.NewVec3(halfWidth, -halfHeight, halfLength)),
		c: Point(dprec.NewVec3(halfWidth, halfHeight, halfLength)),
	}

	// REAR
	boxTriangles[6] = StaticTriangle{
		a: Point(dprec.NewVec3(-halfWidth, halfHeight, -halfLength)),
		b: Point(dprec.NewVec3(halfWidth, -halfHeight, -halfLength)),
		c: Point(dprec.NewVec3(-halfWidth, -halfHeight, -halfLength)),
	}
	boxTriangles[7] = StaticTriangle{
		a: Point(dprec.NewVec3(-halfWidth, halfHeight, -halfLength)),
		b: Point(dprec.NewVec3(halfWidth, halfHeight, -halfLength)),
		c: Point(dprec.NewVec3(halfWidth, -halfHeight, -halfLength)),
	}

	// LEFT
	boxTriangles[8] = StaticTriangle{
		a: Point(dprec.NewVec3(-halfWidth, halfHeight, -halfLength)),
		b: Point(dprec.NewVec3(-halfWidth, -halfHeight, -halfLength)),
		c: Point(dprec.NewVec3(-halfWidth, -halfHeight, halfLength)),
	}
	boxTriangles[9] = StaticTriangle{
		a: Point(dprec.NewVec3(-halfWidth, halfHeight, -halfLength)),
		b: Point(dprec.NewVec3(-halfWidth, -halfHeight, halfLength)),
		c: Point(dprec.NewVec3(-halfWidth, halfHeight, halfLength)),
	}

	// RIGHT
	boxTriangles[10] = StaticTriangle{
		a: Point(dprec.NewVec3(halfWidth, halfHeight, -halfLength)),
		b: Point(dprec.NewVec3(halfWidth, -halfHeight, halfLength)),
		c: Point(dprec.NewVec3(halfWidth, -halfHeight, -halfLength)),
	}
	boxTriangles[11] = StaticTriangle{
		a: Point(dprec.NewVec3(halfWidth, halfHeight, -halfLength)),
		b: Point(dprec.NewVec3(halfWidth, halfHeight, halfLength)),
		c: Point(dprec.NewVec3(halfWidth, -halfHeight, halfLength)),
	}

	secondAsMesh := NewStaticMesh(boxTriangles).WithTransform(second.Transform)
	checkIntersectionBoxWithMesh(first, secondAsMesh, resultSet)
}

func checkIntersectionBoxWithMesh(box StaticBox, mesh StaticMesh, resultSet *IntersectionResultSet) {
	// broad phase
	deltaPosition := dprec.Vec3Diff(mesh.Position(), box.Position())
	if deltaPosition.Length() > box.BoundingSphereRadius()+mesh.BoundingSphereRadius() {
		return
	}

	// narrow phase
	minX := dprec.Vec3Prod(box.Rotation().OrientationX(), -box.Width()/2.0)
	maxX := dprec.Vec3Prod(box.Rotation().OrientationX(), box.Width()/2.0)
	minY := dprec.Vec3Prod(box.Rotation().OrientationY(), -box.Height()/2.0)
	maxY := dprec.Vec3Prod(box.Rotation().OrientationY(), box.Height()/2.0)
	minZ := dprec.Vec3Prod(box.Rotation().OrientationZ(), -box.Length()/2.0)
	maxZ := dprec.Vec3Prod(box.Rotation().OrientationZ(), box.Length()/2.0)

	p1 := dprec.Vec3Sum(dprec.Vec3Sum(dprec.Vec3Sum(box.Position(), minX), minZ), maxY)
	p2 := dprec.Vec3Sum(dprec.Vec3Sum(dprec.Vec3Sum(box.Position(), minX), maxZ), maxY)
	p3 := dprec.Vec3Sum(dprec.Vec3Sum(dprec.Vec3Sum(box.Position(), maxX), maxZ), maxY)
	p4 := dprec.Vec3Sum(dprec.Vec3Sum(dprec.Vec3Sum(box.Position(), maxX), minZ), maxY)
	p5 := dprec.Vec3Sum(dprec.Vec3Sum(dprec.Vec3Sum(box.Position(), minX), minZ), minY)
	p6 := dprec.Vec3Sum(dprec.Vec3Sum(dprec.Vec3Sum(box.Position(), minX), maxZ), minY)
	p7 := dprec.Vec3Sum(dprec.Vec3Sum(dprec.Vec3Sum(box.Position(), maxX), maxZ), minY)
	p8 := dprec.Vec3Sum(dprec.Vec3Sum(dprec.Vec3Sum(box.Position(), maxX), minZ), minY)

	checkLineIntersection := func(line StaticLine, triangle StaticTriangle) {
		heightA := dprec.Vec3Dot(triangle.Normal(), dprec.Vec3Diff(dprec.Vec3(line.A()), dprec.Vec3(triangle.A())))
		heightB := dprec.Vec3Dot(triangle.Normal(), dprec.Vec3Diff(dprec.Vec3(line.B()), dprec.Vec3(triangle.A())))
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
			dprec.Vec3Prod(dprec.Vec3(line.A()), -heightB/(heightA-heightB)),
			dprec.Vec3Prod(dprec.Vec3(line.B()), heightA/(heightA-heightB)),
		)

		if triangle.ContainsPoint(Point(projectedPoint)) {
			resultSet.Add(Intersection{
				Depth:                -heightB,
				FirstContact:         projectedPoint,
				FirstDisplaceNormal:  triangle.Normal(),
				SecondContact:        projectedPoint,
				SecondDisplaceNormal: triangle.Normal(),
			})
		}
	}

	for _, triangle := range mesh.Triangles() {
		triangle := triangle.Transformed(mesh.Transform)
		checkLineIntersection(NewStaticLine(Point(p1), Point(p2)), triangle)
		checkLineIntersection(NewStaticLine(Point(p2), Point(p3)), triangle)
		checkLineIntersection(NewStaticLine(Point(p3), Point(p4)), triangle)
		checkLineIntersection(NewStaticLine(Point(p4), Point(p1)), triangle)

		checkLineIntersection(NewStaticLine(Point(p5), Point(p6)), triangle)
		checkLineIntersection(NewStaticLine(Point(p6), Point(p7)), triangle)
		checkLineIntersection(NewStaticLine(Point(p7), Point(p8)), triangle)
		checkLineIntersection(NewStaticLine(Point(p8), Point(p5)), triangle)

		checkLineIntersection(NewStaticLine(Point(p1), Point(p5)), triangle)
		checkLineIntersection(NewStaticLine(Point(p2), Point(p6)), triangle)
		checkLineIntersection(NewStaticLine(Point(p3), Point(p7)), triangle)
		checkLineIntersection(NewStaticLine(Point(p4), Point(p8)), triangle)
	}
}

func checkIntersectionMeshWithMesh(first, second StaticMesh, resultSet *IntersectionResultSet) {
}

func CheckLineIntersection(line StaticLine, mesh StaticMesh, resultSet *IntersectionResultSet) {
	for _, triangle := range mesh.Triangles() {
		heightA := dprec.Vec3Dot(triangle.Normal(), dprec.Vec3Diff(dprec.Vec3(line.A()), dprec.Vec3(triangle.A())))
		heightB := dprec.Vec3Dot(triangle.Normal(), dprec.Vec3Diff(dprec.Vec3(line.B()), dprec.Vec3(triangle.A())))
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
			dprec.Vec3Prod(dprec.Vec3(line.A()), -heightB/(heightA-heightB)),
			dprec.Vec3Prod(dprec.Vec3(line.B()), heightA/(heightA-heightB)),
		)

		if triangle.ContainsPoint(Point(projectedPoint)) {
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
