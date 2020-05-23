package shape

import "github.com/mokiat/gomath/sprec"

type Intersection struct {
	Depth                float32
	FirstContact         sprec.Vec3
	FirstDisplaceNormal  sprec.Vec3
	SecondContact        sprec.Vec3
	SecondDisplaceNormal sprec.Vec3
}

func (i Intersection) Flipped() Intersection {
	i.FirstContact, i.SecondContact = i.SecondContact, i.FirstContact
	i.FirstDisplaceNormal, i.SecondDisplaceNormal = i.SecondDisplaceNormal, i.FirstDisplaceNormal
	return i
}

func NewIntersectionResultSet(capacity int) *IntersectionResultSet {
	return &IntersectionResultSet{
		intersections: make([]Intersection, 0, capacity),
		flipped:       false,
	}
}

type IntersectionResultSet struct {
	intersections []Intersection
	flipped       bool
}

func (s *IntersectionResultSet) Reset() {
	s.intersections = s.intersections[:0]
}

func (s *IntersectionResultSet) AddFlipped(flipped bool) {
	s.flipped = true
}

func (s *IntersectionResultSet) Add(intersection Intersection) {
	if s.flipped {
		s.intersections = append(s.intersections, intersection.Flipped())
	} else {
		s.intersections = append(s.intersections, intersection)
	}
}

func (s *IntersectionResultSet) Found() bool {
	return len(s.intersections) > 0
}

func (s *IntersectionResultSet) Intersections() []Intersection {
	return s.intersections
}

func CheckIntersection(first, second Placement, resultSet *IntersectionResultSet) {
	switch first.Shape.(type) {
	case StaticSphere:
		CheckIntersectionSphereUnknown(first, second, resultSet)
	case StaticBox:
		CheckIntersectionBoxUnknown(first, second, resultSet)
	case StaticMesh:
		CheckIntersectionMeshUnknown(first, second, resultSet)
	}
}

func CheckIntersectionSphereUnknown(first, second Placement, resultSet *IntersectionResultSet) {
	switch second.Shape.(type) {
	case StaticSphere:
		resultSet.AddFlipped(false)
		CheckIntersectionSphereSphere(first, second, resultSet)
	case StaticBox:
		resultSet.AddFlipped(false)
		CheckIntersectionSphereBox(first, second, resultSet)
	case StaticMesh:
		resultSet.AddFlipped(false)
		CheckIntersectionSphereMesh(first, second, resultSet)
	}
}

func CheckIntersectionBoxUnknown(first, second Placement, resultSet *IntersectionResultSet) {
	switch second.Shape.(type) {
	case StaticSphere:
		resultSet.AddFlipped(true)
		CheckIntersectionSphereBox(second, first, resultSet)
	case StaticBox:
		resultSet.AddFlipped(false)
		CheckIntersectionBoxBox(first, second, resultSet)
	case StaticMesh:
		resultSet.AddFlipped(false)
		CheckIntersectionBoxMesh(first, second, resultSet)
	}
}

func CheckIntersectionMeshUnknown(first, second Placement, resultSet *IntersectionResultSet) {
	switch second.Shape.(type) {
	case StaticSphere:
		resultSet.AddFlipped(true)
		CheckIntersectionSphereMesh(second, first, resultSet)
	case StaticBox:
		resultSet.AddFlipped(true)
		CheckIntersectionBoxMesh(second, first, resultSet)
	case StaticMesh:
		resultSet.AddFlipped(false)
		CheckIntersectionMeshMesh(first, second, resultSet)
	}
}

func CheckIntersectionSphereSphere(first, second Placement, resultSet *IntersectionResultSet) {
}

func CheckIntersectionSphereBox(first, second Placement, resultSet *IntersectionResultSet) {
}

func CheckIntersectionSphereMesh(spherePlacement, meshPlacement Placement, resultSet *IntersectionResultSet) {
	sphere := spherePlacement.Shape.(StaticSphere)
	mesh := meshPlacement.Shape.(StaticMesh)

	// broad phase
	deltaPosition := sprec.Vec3Diff(meshPlacement.Position, spherePlacement.Position)
	if deltaPosition.Length() > sphere.Radius()+mesh.BoundingSphereRadius() {
		return
	}

	// narrow phase
	for _, triangle := range mesh.Triangles() {
		triangleWS := triangle.Transformed(meshPlacement.Position, meshPlacement.Orientation)
		height := sprec.Vec3Dot(triangleWS.Normal(), sprec.Vec3Diff(spherePlacement.Position, triangleWS.A()))
		if height > sphere.Radius() || height < -sphere.Radius() {
			continue
		}

		projectedPoint := sprec.Vec3Diff(spherePlacement.Position, sprec.Vec3Prod(triangle.Normal(), height))
		if triangleWS.ContainsPoint(projectedPoint) {
			resultSet.Add(Intersection{
				Depth:                sphere.Radius() - height,
				FirstContact:         projectedPoint, // TODO: Extrude to equal radius length
				FirstDisplaceNormal:  triangle.Normal(),
				SecondContact:        projectedPoint,
				SecondDisplaceNormal: sprec.InverseVec3(triangle.Normal()),
			})
		}
	}
}

func CheckIntersectionBoxBox(first, second Placement, resultSet *IntersectionResultSet) {
}

func CheckIntersectionBoxMesh(boxPlacement, meshPlacement Placement, resultSet *IntersectionResultSet) {
	box := boxPlacement.Shape.(StaticBox)
	mesh := meshPlacement.Shape.(StaticMesh)

	// broad phase
	deltaPosition := sprec.Vec3Diff(meshPlacement.Position, boxPlacement.Position)
	boxBoundingSphereRadius := sprec.Sqrt(box.Width()*box.Width()+box.Height()*box.Height()+box.Length()*box.Length()) / 2.0
	if deltaPosition.Length() > boxBoundingSphereRadius+mesh.BoundingSphereRadius() {
		return
	}

	minX := sprec.Vec3Prod(boxPlacement.Orientation.OrientationX(), -box.Width()/2.0)
	maxX := sprec.Vec3Prod(boxPlacement.Orientation.OrientationX(), box.Width()/2.0)
	minY := sprec.Vec3Prod(boxPlacement.Orientation.OrientationY(), -box.Height()/2.0)
	maxY := sprec.Vec3Prod(boxPlacement.Orientation.OrientationY(), box.Height()/2.0)
	minZ := sprec.Vec3Prod(boxPlacement.Orientation.OrientationZ(), -box.Length()/2.0)
	maxZ := sprec.Vec3Prod(boxPlacement.Orientation.OrientationZ(), box.Length()/2.0)

	p1 := sprec.Vec3Sum(sprec.Vec3Sum(sprec.Vec3Sum(boxPlacement.Position, minX), minZ), maxY)
	p2 := sprec.Vec3Sum(sprec.Vec3Sum(sprec.Vec3Sum(boxPlacement.Position, minX), maxZ), maxY)
	p3 := sprec.Vec3Sum(sprec.Vec3Sum(sprec.Vec3Sum(boxPlacement.Position, maxX), maxZ), maxY)
	p4 := sprec.Vec3Sum(sprec.Vec3Sum(sprec.Vec3Sum(boxPlacement.Position, maxX), minZ), maxY)
	p5 := sprec.Vec3Sum(sprec.Vec3Sum(sprec.Vec3Sum(boxPlacement.Position, minX), minZ), minY)
	p6 := sprec.Vec3Sum(sprec.Vec3Sum(sprec.Vec3Sum(boxPlacement.Position, minX), maxZ), minY)
	p7 := sprec.Vec3Sum(sprec.Vec3Sum(sprec.Vec3Sum(boxPlacement.Position, maxX), maxZ), minY)
	p8 := sprec.Vec3Sum(sprec.Vec3Sum(sprec.Vec3Sum(boxPlacement.Position, maxX), minZ), minY)

	checkLineIntersection := func(line StaticLine, triangle StaticTriangle) {
		heightA := sprec.Vec3Dot(triangle.Normal(), sprec.Vec3Diff(line.A(), triangle.A()))
		heightB := sprec.Vec3Dot(triangle.Normal(), sprec.Vec3Diff(line.B(), triangle.A()))
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

		projectedPoint := sprec.Vec3Sum(
			sprec.Vec3Prod(line.A(), -heightB/(heightA-heightB)),
			sprec.Vec3Prod(line.B(), heightA/(heightA-heightB)),
		)

		if triangle.ContainsPoint(projectedPoint) {
			resultSet.Add(Intersection{
				Depth:                -heightB,
				FirstContact:         projectedPoint,
				FirstDisplaceNormal:  triangle.Normal(),
				SecondContact:        projectedPoint,
				SecondDisplaceNormal: sprec.InverseVec3(triangle.Normal()),
			})
		}
	}

	// narrow phase
	for _, triangle := range mesh.Triangles() {
		triangleWS := triangle.Transformed(meshPlacement.Position, meshPlacement.Orientation)
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

func CheckIntersectionMeshMesh(first, second Placement, resultSet *IntersectionResultSet) {
}
