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
	}
}

// IntersectionResultSet is a structure that can be used to collection the
// result of an intersection test.
type IntersectionResultSet struct {
	intersections []Intersection
}

// Reset clears the buffer of this result set so that it can be reused.
func (s *IntersectionResultSet) Reset() {
	s.intersections = s.intersections[:0]
}

// Add adds a new Intersection to this set.
func (s *IntersectionResultSet) Add(intersection Intersection) {
	s.intersections = append(s.intersections, intersection)
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

// CheckIntersectionLineWithMesh checks if a StaticLine intersects with a
// StaticMesh shape.
func CheckIntersectionLineWithMesh(line StaticLine, mesh Placement[StaticMesh], resultSet *IntersectionResultSet) {
	for _, triangle := range mesh.shape.Triangles() {
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
			addIntersection(resultSet, false, Intersection{
				Depth:                -heightB,
				FirstContact:         projectedPoint,
				FirstDisplaceNormal:  triangle.Normal(),
				SecondContact:        projectedPoint,
				SecondDisplaceNormal: dprec.InverseVec3(triangle.Normal()),
			})
		}
	}
}

// CheckIntersectionSphereWithSphere checks if two StaticSphere shapes intersect.
func CheckIntersectionSphereWithSphere(first, second Placement[StaticSphere], resultSet *IntersectionResultSet) {
	if isIntersectionPossible(first, second) {
		checkIntersectionSphereWithSphere(first, second, resultSet)
	}
}

// CheckIntersectionBoxWithBox checks if two StaticBox shapes intersect.
func CheckIntersectionBoxWithBox(first, second Placement[StaticBox], resultSet *IntersectionResultSet) {
	if isIntersectionPossible(first, second) {
		checkIntersectionBoxWithBox(first, second, resultSet)
	}
}

// CheckIntersectionMeshWithMesh checks if two StaticMesh shapes intersect.
func CheckIntersectionMeshWithMesh(first, second Placement[StaticMesh], resultSet *IntersectionResultSet) {
	if isIntersectionPossible(first, second) {
		checkIntersectionMeshWithMesh(first, second, resultSet)
	}
}

// CheckIntersectionSphereWithBox checks if a StaticSphere shape intersects with
// a StaticBox shape.
func CheckIntersectionSphereWithBox(first Placement[StaticSphere], second Placement[StaticBox], resultSet *IntersectionResultSet) {
	if isIntersectionPossible(first, second) {
		checkIntersectionSphereWithBox(first, second, false, resultSet)
	}
}

// CheckIntersectionBoxWithSphere checks if a StaticBox shape intersects with
// a StaticSphere shape.
func CheckIntersectionBoxWithSphere(first Placement[StaticBox], second Placement[StaticSphere], resultSet *IntersectionResultSet) {
	if isIntersectionPossible(first, second) {
		checkIntersectionSphereWithBox(second, first, true, resultSet)
	}
}

// CheckIntersectionSphereWithMesh checks if a StaticSphere shape intersects with
// a StaticMesh shape.
func CheckIntersectionSphereWithMesh(first Placement[StaticSphere], second Placement[StaticMesh], resultSet *IntersectionResultSet) {
	if isIntersectionPossible(first, second) {
		checkIntersectionSphereWithMesh(first, second, false, resultSet)
	}
}

// CheckIntersectionMeshWithSphere checks if a StaticMesh shape intersects with
// a StaticSphere shape.
func CheckIntersectionMeshWithSphere(first Placement[StaticMesh], second Placement[StaticSphere], resultSet *IntersectionResultSet) {
	if isIntersectionPossible(first, second) {
		checkIntersectionSphereWithMesh(second, first, true, resultSet)
	}
}

// CheckIntersectionBoxWithMesh checks if a StaticBox shape intersects with
// a StaticMesh shape.
func CheckIntersectionBoxWithMesh(first Placement[StaticBox], second Placement[StaticMesh], resultSet *IntersectionResultSet) {
	if isIntersectionPossible(first, second) {
		checkIntersectionBoxWithMesh(first, second, false, resultSet)
	}
}

// CheckIntersectionMeshWithBox checks if a StaticMesh shape intersects with
// a StaticBox shape.
func CheckIntersectionMeshWithBox(first Placement[StaticMesh], second Placement[StaticBox], resultSet *IntersectionResultSet) {
	if isIntersectionPossible(first, second) {
		checkIntersectionBoxWithMesh(second, first, true, resultSet)
	}
}

// CheckIntersection checks whether the two arbitrary shapes intersect.
//
// If you know the types of the shapes, you should use the specific Check
// methods instead.
func CheckIntersection(first, second Placement[Shape], resultSet *IntersectionResultSet) {
	if !isIntersectionPossible(first, second) {
		return
	}
	switch firstShape := first.shape.(type) {
	case StaticSphere:
		checkIntersectionSphereWithUnknown(NewPlacement(first.Transform, firstShape), second, resultSet)
	case StaticBox:
		checkIntersectionBoxWithUnknown(NewPlacement(first.Transform, firstShape), second, resultSet)
	case StaticMesh:
		checkIntersectionMeshWithUnknown(NewPlacement(first.Transform, firstShape), second, resultSet)
	}
}

// isIntersectionPossible performs a quick check whether two shapes can at all
// intersect, based on distances and bounding spheres.
func isIntersectionPossible[A, B Shape](first Placement[A], second Placement[B]) bool {
	r1 := first.Shape().BoundingSphereRadius()
	r2 := second.Shape().BoundingSphereRadius()
	sqrDistance := dprec.Vec3Diff(second.Position(), first.Position()).SqrLength()
	return sqrDistance <= (r1+r2)*(r1+r2)
}

// addIntersection is a helper function that adds an intersection to a result
// set and can flip it beforehand.
func addIntersection(resultSet *IntersectionResultSet, flipped bool, intersection Intersection) {
	if flipped {
		resultSet.Add(intersection.Flipped())
	} else {
		resultSet.Add(intersection)
	}
}

func checkIntersectionSphereWithSphere(first, second Placement[StaticSphere], resultSet *IntersectionResultSet) {
	firstPosition := first.Position()
	firstRadius := first.Shape().Radius()

	secondPosition := second.Position()
	secondRadius := second.Shape().Radius()

	deltaPosition := dprec.Vec3Diff(secondPosition, firstPosition)
	distance := deltaPosition.Length()
	overlap := (firstRadius + secondRadius) - distance

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

var (
	// TODO: Use SAT instead
	boxTriangles = make([]StaticTriangle, 2*6)
)

func checkIntersectionBoxWithBox(first, second Placement[StaticBox], resultSet *IntersectionResultSet) {
	halfWidth := second.shape.HalfWidth()
	halfHeight := second.shape.HalfHeight()
	halfLength := second.shape.HalfLength()

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

	secondAsMesh := NewPlacement(second.Transform, NewStaticMesh(boxTriangles))
	checkIntersectionBoxWithMesh(first, secondAsMesh, false, resultSet)
}

func checkIntersectionMeshWithMesh(first, second Placement[StaticMesh], resultSet *IntersectionResultSet) {
	// TODO
}

func checkIntersectionSphereWithBox(first Placement[StaticSphere], second Placement[StaticBox], flipped bool, resultSet *IntersectionResultSet) {
	boxAxisX := second.Rotation().OrientationX()
	boxAxisY := second.Rotation().OrientationY()
	boxAxisZ := second.Rotation().OrientationZ()

	deltaPosition := dprec.Vec3Diff(first.Position(), second.Position())
	distanceFront := dprec.Vec3Dot(boxAxisZ, deltaPosition) - second.Shape().HalfLength()
	distanceBack := -(second.Shape().Length() + distanceFront)
	distanceTop := dprec.Vec3Dot(boxAxisY, deltaPosition) - second.Shape().HalfHeight()
	distanceBottom := -(second.Shape().Height() + distanceTop)
	distanceRight := dprec.Vec3Dot(boxAxisX, deltaPosition) - second.Shape().HalfWidth()
	distanceLeft := -(second.Shape().Width() + distanceRight)

	isFront := distanceFront > 0
	isBack := distanceBack > 0
	isTop := distanceTop > 0
	isBottom := distanceBottom > 0
	isRight := distanceRight > 0
	isLeft := distanceLeft > 0

	// right check
	if isRight && !isFront && !isBack && !isTop && !isBottom {
		if depth := first.Shape().Radius() - distanceRight; depth > 0 {
			addIntersection(resultSet, flipped, Intersection{
				Depth:                depth,
				FirstContact:         dprec.Vec3Diff(first.Position(), dprec.Vec3Prod(boxAxisX, first.Shape().Radius())),
				FirstDisplaceNormal:  boxAxisX,
				SecondContact:        dprec.Vec3Diff(first.Position(), dprec.Vec3Prod(boxAxisX, distanceRight)),
				SecondDisplaceNormal: dprec.InverseVec3(boxAxisX),
			})
		}
		return
	}

	// left check
	if isLeft && !isFront && !isBack && !isTop && !isBottom {
		if depth := first.Shape().Radius() - distanceLeft; depth > 0 {
			addIntersection(resultSet, flipped, Intersection{
				Depth:                depth,
				FirstContact:         dprec.Vec3Sum(first.Position(), dprec.Vec3Prod(boxAxisX, first.Shape().Radius())),
				FirstDisplaceNormal:  dprec.InverseVec3(boxAxisX),
				SecondContact:        dprec.Vec3Sum(first.Position(), dprec.Vec3Prod(boxAxisX, distanceLeft)),
				SecondDisplaceNormal: boxAxisX,
			})
		}
		return
	}

	// top check
	if isTop && !isFront && !isBack && !isLeft && !isRight {
		if depth := first.Shape().Radius() - distanceTop; depth > 0 {
			addIntersection(resultSet, flipped, Intersection{
				Depth:                depth,
				FirstContact:         dprec.Vec3Diff(first.Position(), dprec.Vec3Prod(boxAxisY, first.Shape().Radius())),
				FirstDisplaceNormal:  boxAxisY,
				SecondContact:        dprec.Vec3Diff(first.Position(), dprec.Vec3Prod(boxAxisY, distanceTop)),
				SecondDisplaceNormal: dprec.InverseVec3(boxAxisY),
			})
		}
		return
	}

	// bottom check
	if isBottom && !isFront && !isBack && !isLeft && !isRight {
		if depth := first.Shape().Radius() - distanceBottom; depth > 0 {
			addIntersection(resultSet, flipped, Intersection{
				Depth:                depth,
				FirstContact:         dprec.Vec3Sum(first.Position(), dprec.Vec3Prod(boxAxisY, first.Shape().Radius())),
				FirstDisplaceNormal:  dprec.InverseVec3(boxAxisY),
				SecondContact:        dprec.Vec3Sum(first.Position(), dprec.Vec3Prod(boxAxisY, distanceBottom)),
				SecondDisplaceNormal: boxAxisY,
			})
		}
		return
	}

	// front check
	if isFront && !isTop && !isBottom && !isLeft && !isRight {
		if depth := first.Shape().Radius() - distanceFront; depth > 0 {
			addIntersection(resultSet, flipped, Intersection{
				Depth:                depth,
				FirstContact:         dprec.Vec3Diff(first.Position(), dprec.Vec3Prod(boxAxisZ, first.Shape().Radius())),
				FirstDisplaceNormal:  boxAxisZ,
				SecondContact:        dprec.Vec3Diff(first.Position(), dprec.Vec3Prod(boxAxisZ, distanceFront)),
				SecondDisplaceNormal: dprec.InverseVec3(boxAxisZ),
			})
		}
		return
	}

	// back check
	if isBack && !isTop && !isBottom && !isLeft && !isRight {
		if depth := first.Shape().Radius() - distanceBack; depth > 0 {
			addIntersection(resultSet, flipped, Intersection{
				Depth:                depth,
				FirstContact:         dprec.Vec3Sum(first.Position(), dprec.Vec3Prod(boxAxisZ, first.Shape().Radius())),
				FirstDisplaceNormal:  dprec.InverseVec3(boxAxisZ),
				SecondContact:        dprec.Vec3Sum(first.Position(), dprec.Vec3Prod(boxAxisZ, distanceBack)),
				SecondDisplaceNormal: boxAxisZ,
			})
		}
		return
	}

	// TODO: Top-Right Edge
	// TODO: Top-Left Edge
	// TODO: Top-Front Edge
	// TODO: Top-Back Edge
	// TODO: Bottom-Right Edge
	// TODO: Bottom-Left Edge
	// TODO: Bottom-Front Edge
	// TODO: Bottom-Back Edge

	// TODO: Top-Left-Front Vertex
	// TODO: Top-Left-Back Vertex
	// TODO: Top-Right-Front Vertex
	// TODO: Top-Right-Back Vertex
	// TODO: Bottom-Left-Front Vertex
	// TODO: Bottom-Left-Back Vertex
	// TODO: Bottom-Right-Front Vertex
	// TODO: Bottom-Right-Back Vertex
}

func checkIntersectionSphereWithMesh(sphere Placement[StaticSphere], mesh Placement[StaticMesh], flipped bool, resultSet *IntersectionResultSet) {
	for _, triangle := range mesh.Shape().Triangles() {
		triangle = triangle.Transformed(mesh.Transform)

		height := dprec.Vec3Dot(triangle.Normal(), dprec.Vec3Diff(sphere.Position(), dprec.Vec3(triangle.A())))
		if dprec.Abs(height) > sphere.Shape().Radius() {
			continue
		}

		projectedPoint := dprec.Vec3Diff(sphere.Position(), dprec.Vec3Prod(triangle.Normal(), height))
		if triangle.ContainsPoint(Point(projectedPoint)) {
			depth := sphere.Shape().Radius() - dprec.Abs(height)
			addIntersection(resultSet, flipped, Intersection{
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

func checkIntersectionBoxWithMesh(box Placement[StaticBox], mesh Placement[StaticMesh], flipped bool, resultSet *IntersectionResultSet) {
	minX := dprec.Vec3Prod(box.Rotation().OrientationX(), -box.Shape().HalfWidth())
	maxX := dprec.Vec3Prod(box.Rotation().OrientationX(), box.Shape().HalfWidth())
	minY := dprec.Vec3Prod(box.Rotation().OrientationY(), -box.Shape().HalfHeight())
	maxY := dprec.Vec3Prod(box.Rotation().OrientationY(), box.Shape().HalfHeight())
	minZ := dprec.Vec3Prod(box.Rotation().OrientationZ(), -box.Shape().HalfLength())
	maxZ := dprec.Vec3Prod(box.Rotation().OrientationZ(), box.Shape().HalfLength())

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
			addIntersection(resultSet, flipped, Intersection{
				Depth:                -heightB,
				FirstContact:         projectedPoint,
				FirstDisplaceNormal:  triangle.Normal(),
				SecondContact:        projectedPoint,
				SecondDisplaceNormal: triangle.Normal(),
			})
		}
	}

	for _, triangle := range mesh.Shape().Triangles() {
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

func checkIntersectionSphereWithUnknown(first Placement[StaticSphere], second Placement[Shape], resultSet *IntersectionResultSet) {
	switch secondShape := second.shape.(type) {
	case StaticSphere:
		checkIntersectionSphereWithSphere(first, NewPlacement(second.Transform, secondShape), resultSet)
	case StaticBox:
		checkIntersectionSphereWithBox(first, NewPlacement(second.Transform, secondShape), false, resultSet)
	case StaticMesh:
		checkIntersectionSphereWithMesh(first, NewPlacement(second.Transform, secondShape), false, resultSet)
	}
}

func checkIntersectionBoxWithUnknown(first Placement[StaticBox], second Placement[Shape], resultSet *IntersectionResultSet) {
	switch secondShape := second.shape.(type) {
	case StaticSphere:
		checkIntersectionSphereWithBox(NewPlacement(second.Transform, secondShape), first, true, resultSet)
	case StaticBox:
		checkIntersectionBoxWithBox(first, NewPlacement(second.Transform, secondShape), resultSet)
	case StaticMesh:
		checkIntersectionBoxWithMesh(first, NewPlacement(second.Transform, secondShape), false, resultSet)
	}
}

func checkIntersectionMeshWithUnknown(first Placement[StaticMesh], second Placement[Shape], resultSet *IntersectionResultSet) {
	switch secondShape := second.shape.(type) {
	case StaticSphere:
		checkIntersectionSphereWithMesh(NewPlacement(second.Transform, secondShape), first, true, resultSet)
	case StaticBox:
		checkIntersectionBoxWithMesh(NewPlacement(second.Transform, secondShape), first, true, resultSet)
	case StaticMesh:
		checkIntersectionMeshWithMesh(first, NewPlacement(second.Transform, secondShape), resultSet)
	}
}
