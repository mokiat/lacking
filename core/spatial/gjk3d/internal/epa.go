package internal

import (
	"math"

	"github.com/mokiat/gomath/dprec"
)

// EPASolver computes contact information (separation normal, penetration
// depth and closest feature) for two overlapping shapes, based on the final
// simplex produced by a [GJKSolver] run over the same [MinkowskiShape].
//
// Usage: call [EPASolver.Reset] with the shape and the final GJK simplex,
// then call [EPASolver.Next] repeatedly until it returns false, then obtain
// the result through [EPASolver.Solution].
type EPASolver struct {
	// polytope holds the faces of the expanding polytope. It is a reused slice
	// rather than a map value so that the large [PolytopeTriangle] (which
	// exceeds Go's 128-byte map element inline threshold) is not heap-allocated
	// on every insertion.
	polytope []PolytopeTriangle
	// polytopeIndex maps a face identity to its position within polytope,
	// providing the dedup-on-insert that the former map value type gave for
	// free. Its value is a small int, so it stays inline and allocation-free.
	polytopeIndex       map[PolytopeTriangleID]int32
	horizon             map[polytopeEdgeID]polytopeEdge
	solution            EPASolution
	skinRadius          float64
	remainingIterations uint32
}

// NewEPASolver creates a new [EPASolver] instance.
func NewEPASolver() *EPASolver {
	return &EPASolver{
		polytope:      make([]PolytopeTriangle, 0, 32),
		polytopeIndex: make(map[PolytopeTriangleID]int32, 32),
		horizon:       make(map[polytopeEdgeID]polytopeEdge, 32),
	}
}

// Reset prepares the solver for a new query against the given shape. The
// simplex must be the final simplex of a [GJKSolver] run over the same
// shape and containsOrigin must be the corresponding
// [GJKSolver.ContainsOrigin] result.
//
// When the origin is not contained, the simplex is the point, edge or
// triangle feature of the Minkowski difference that is closest to the origin,
// and the solution is computed immediately, requiring no iteration.
func (s *EPASolver) Reset(shape *MinkowskiShape, simplex Simplex, containsOrigin bool) {
	s.polytope = s.polytope[:0]
	clear(s.polytopeIndex)
	s.solution = EPASolution{
		VertexA: simplex.Vertices[0],
		VertexB: simplex.Vertices[0],
		VertexC: simplex.Vertices[0],
		Normal:  dprec.BasisXVec3(),
		BaryA:   1.0,
		Depth:   0.0,
	}
	s.skinRadius = shape.SkinRadius
	s.remainingIterations = uint32(shape.MaxIterations())

	if !containsOrigin {
		switch simplex.VertexCount {
		case 1:
			s.terminatePoint(shape, simplex.Vertices[0])
		case 2:
			s.terminateEdge(shape, simplex.Vertices[0], simplex.Vertices[1])
		case 3:
			s.terminateTriangle(simplex.Vertices[0], simplex.Vertices[1], simplex.Vertices[2])
		default:
			panic("unexpected simplex vertex count")
		}
	} else {
		// A contained origin yields either a tetrahedron simplex or, when the
		// origin happens to lie exactly on an intermediate triangle, a triangle
		// simplex. Seed a closed polytope enclosing the origin from whichever
		// was produced.
		switch simplex.VertexCount {
		case 4:
			s.seedTetrahedron(simplex.Vertices[0], simplex.Vertices[1], simplex.Vertices[2], simplex.Vertices[3])
		case 3:
			s.seedTriangle(shape, simplex.Vertices[0], simplex.Vertices[1], simplex.Vertices[2])
		default:
			panic("unexpected simplex vertex count")
		}
	}
}

// Next runs a single EPA iteration. It returns false once the algorithm
// has converged (or the iteration budget is exhausted) and further calls
// would make no progress.
func (s *EPASolver) Next(shape *MinkowskiShape) bool {
	if s.remainingIterations == 0 {
		return false
	}
	s.remainingIterations--

	minTriangle, ok := s.closestTriangle()
	if !ok {
		s.remainingIterations = 0
		return false
	}
	// Record the best solution so far, so that a valid result is available even
	// if the iteration budget is exhausted before convergence.
	s.solvePolytopeTriangle(minTriangle)

	support := shape.Support(minTriangle.Normal)
	advance := dprec.Vec3Dot(minTriangle.Normal, dprec.Vec3Diff(support.Position, minTriangle.VertexA.Position))
	if advance <= dprec.Epsilon || s.polytopeContainsVertex(support.Refs) {
		// The support does not extend the polytope past its closest face, so
		// the closest face is the closest feature of the Minkowski difference.
		s.remainingIterations = 0
		return false
	}

	// Expand the polytope: remove every face visible from the support point and
	// re-triangulate the resulting horizon loop toward the support. The visible
	// faces are dropped by compacting the survivors to the front of the slice
	// in place, rebuilding polytopeIndex for them as we go.
	clear(s.horizon)
	clear(s.polytopeIndex)
	kept := s.polytope[:0]
	for i := range s.polytope {
		triangle := s.polytope[i]
		visible := dprec.Vec3Dot(triangle.Normal, dprec.Vec3Diff(support.Position, triangle.VertexA.Position))
		if visible > 0 {
			addHorizonEdge(s.horizon, triangle.VertexA, triangle.VertexB)
			addHorizonEdge(s.horizon, triangle.VertexB, triangle.VertexC)
			addHorizonEdge(s.horizon, triangle.VertexC, triangle.VertexA)
			continue
		}
		id := PolytopeTriangleID{triangle.VertexA.Refs, triangle.VertexB.Refs, triangle.VertexC.Refs}
		s.polytopeIndex[id] = int32(len(kept))
		kept = append(kept, triangle)
	}
	s.polytope = kept
	for _, edge := range s.horizon {
		s.addPolytopeTriangle(edge.VertexA, edge.VertexB, support)
	}

	return true
}

// Solution returns the computed [EPASolution]. The result is only reliable
// once [EPASolver.Next] has been iterated until it returned false.
func (s *EPASolver) Solution() EPASolution {
	return s.solution
}

// closestTriangle returns the polytope face that is nearest to the origin.
func (s *EPASolver) closestTriangle() (PolytopeTriangle, bool) {
	var (
		result   PolytopeTriangle
		distance = math.MaxFloat64
		found    bool
	)
	for _, triangle := range s.polytope {
		if triangle.Distance < distance {
			distance = triangle.Distance
			result = triangle
			found = true
		}
	}
	return result, found
}

// solvePolytopeTriangle records the solution implied by a polytope face, whose
// outward normal is the separation direction and whose distance to the origin
// (plus the skin radius) is the penetration depth. This is the contained-origin
// regime, where the origin projects onto the interior of the closest face.
func (s *EPASolver) solvePolytopeTriangle(triangle PolytopeTriangle) {
	baryA, baryB, baryC := barycentricClosestToOrigin(triangle.VertexA.Position, triangle.VertexB.Position, triangle.VertexC.Position)
	s.setSolution(triangle.VertexA, triangle.VertexB, triangle.VertexC, triangle.Normal, baryA, baryB, baryC)
}

// terminateTriangle completes the solve with a triangle as the closest feature
// of the Minkowski difference to the origin (the non-contained regime). The
// normal points from the feature toward the origin.
func (s *EPASolver) terminateTriangle(vertexA, vertexB, vertexC MinkowskiVertex) {
	baryA, baryB, baryC := barycentricClosestToOrigin(vertexA.Position, vertexB.Position, vertexC.Position)
	closest := closestPoint(vertexA.Position, vertexB.Position, vertexC.Position, baryA, baryB, baryC)

	var normal dprec.Vec3
	if closest.SqrLength() < 1e-12 {
		// The origin lies on the triangle, so any origin-to-feature direction is
		// undefined. Fall back to the triangle normal implied by the GJK winding,
		// which points toward the origin side.
		edgeAB := dprec.Vec3Diff(vertexB.Position, vertexA.Position)
		edgeAC := dprec.Vec3Diff(vertexC.Position, vertexA.Position)
		normal = dprec.UnitVec3(dprec.Vec3Cross(edgeAB, edgeAC))
	} else {
		normal = dprec.InverseVec3(dprec.UnitVec3(closest))
	}

	s.remainingIterations = 0 // ensure we can't be asked to iterate further
	s.setSolution(vertexA, vertexB, vertexC, normal, baryA, baryB, baryC)
}

// terminateEdge completes the solve with an edge as the closest feature of the
// Minkowski difference to the origin, falling back to one of its vertices when
// the edge is degenerate. The normal points from the feature toward the origin.
func (s *EPASolver) terminateEdge(shape *MinkowskiShape, vertexA, vertexB MinkowskiVertex) {
	edge := dprec.Vec3Diff(vertexB.Position, vertexA.Position)
	sqrLength := edge.SqrLength()
	if sqrLength < 1e-12 {
		s.terminatePoint(shape, vertexA) // treat degenerate edges as points
		return
	}

	lerp := max(0.0, min(-dprec.Vec3Dot(edge, vertexA.Position)/sqrLength, 1.0))
	closest := dprec.Vec3Lerp(vertexA.Position, vertexB.Position, lerp)

	var normal dprec.Vec3
	if closest.SqrLength() < 1e-12 {
		// The origin lies on the edge line, so the origin-to-feature direction is
		// undefined. Any direction perpendicular to the edge is a valid normal.
		normal = dprec.UnitVec3(perpendicularVec3(edge))
	} else {
		normal = dprec.InverseVec3(dprec.UnitVec3(closest))
	}

	s.remainingIterations = 0 // ensure we can't be asked to iterate further
	s.setSolution(vertexA, vertexB, vertexB, normal, 1.0-lerp, lerp, 0.0)
}

// terminatePoint completes the solve with a single vertex as the closest
// feature of the Minkowski difference to the origin.
func (s *EPASolver) terminatePoint(shape *MinkowskiShape, vertex MinkowskiVertex) {
	var normal dprec.Vec3
	if vertex.Position.SqrLength() < 1e-12 {
		normal = shape.VertexNormal(vertex)
	} else {
		normal = dprec.InverseVec3(dprec.UnitVec3(vertex.Position))
	}

	s.remainingIterations = 0 // ensure we can't be asked to iterate further
	s.setSolution(vertex, vertex, vertex, normal, 1.0, 0.0, 0.0)
}

// setSolution records the solution described by the given closest feature,
// separation normal and barycentric weights. The penetration depth is the skin
// radius plus the signed distance from the origin to the feature along the
// normal (positive when the origin is contained, negative when it lies outside
// the core but within the skin).
func (s *EPASolver) setSolution(vertexA, vertexB, vertexC MinkowskiVertex, normal dprec.Vec3, baryA, baryB, baryC float64) {
	closest := closestPoint(vertexA.Position, vertexB.Position, vertexC.Position, baryA, baryB, baryC)
	s.solution = EPASolution{
		VertexA: vertexA,
		VertexB: vertexB,
		VertexC: vertexC,
		Normal:  normal,
		BaryA:   baryA,
		BaryB:   baryB,
		BaryC:   baryC,
		Depth:   s.skinRadius + dprec.Vec3Dot(normal, closest),
	}
}

// seedTetrahedron seeds the polytope with the four faces of a tetrahedron that
// encloses the origin, each wound outward (away from the opposite vertex). This
// establishes the globally consistent outward winding that the horizon
// expansion in [EPASolver.Next] relies on.
func (s *EPASolver) seedTetrahedron(vertA, vertB, vertC, vertD MinkowskiVertex) {
	s.seedPolytopeFace(vertA, vertB, vertC, vertD)
	s.seedPolytopeFace(vertA, vertB, vertD, vertC)
	s.seedPolytopeFace(vertA, vertC, vertD, vertB)
	s.seedPolytopeFace(vertB, vertC, vertD, vertA)
}

// seedTriangle seeds the polytope when GJK reports containment with only a
// triangle simplex, which happens when the origin lies exactly on the triangle.
// It probes the shape on both sides of the triangle plane. When the difference
// extends to both sides, the triangle plus the two support points form a
// closed bipyramid enclosing the origin. Otherwise the origin lies on a genuine
// boundary face and the contact is resolved immediately against that face.
func (s *EPASolver) seedTriangle(shape *MinkowskiShape, vertA, vertB, vertC MinkowskiVertex) {
	faceNormal := dprec.UnitVec3(dprec.Vec3Cross(
		dprec.Vec3Diff(vertB.Position, vertA.Position),
		dprec.Vec3Diff(vertC.Position, vertA.Position),
	))
	front := shape.Support(faceNormal)
	back := shape.Support(dprec.InverseVec3(faceNormal))
	frontDistance := dprec.Vec3Dot(faceNormal, dprec.Vec3Diff(front.Position, vertA.Position))
	backDistance := dprec.Vec3Dot(faceNormal, dprec.Vec3Diff(back.Position, vertA.Position))

	const planeEpsilon = 1e-9
	frontExtends := frontDistance > planeEpsilon
	backExtends := backDistance < -planeEpsilon
	switch {
	case frontExtends && backExtends:
		// Bipyramid: the two apexes cap the triangle from opposite sides.
		s.seedPolytopeFace(vertA, vertB, front, vertC)
		s.seedPolytopeFace(vertB, vertC, front, vertA)
		s.seedPolytopeFace(vertC, vertA, front, vertB)
		s.seedPolytopeFace(vertA, vertB, back, vertC)
		s.seedPolytopeFace(vertB, vertC, back, vertA)
		s.seedPolytopeFace(vertC, vertA, back, vertB)
	case frontExtends:
		// The difference extends only toward the front, so the triangle is a
		// boundary face and its outward normal points backward.
		s.terminateContainedFace(vertA, vertB, vertC, dprec.InverseVec3(faceNormal))
	default:
		// The difference extends only toward the back (or is flat), so the
		// triangle is a boundary face and its outward normal points forward.
		s.terminateContainedFace(vertA, vertB, vertC, faceNormal)
	}
}

// terminateContainedFace completes the solve with a boundary face on which the
// contained origin lies. The normal is the outward face normal.
func (s *EPASolver) terminateContainedFace(vertA, vertB, vertC MinkowskiVertex, normal dprec.Vec3) {
	baryA, baryB, baryC := barycentricClosestToOrigin(vertA.Position, vertB.Position, vertC.Position)
	s.remainingIterations = 0 // ensure we can't be asked to iterate further
	s.setSolution(vertA, vertB, vertC, normal, baryA, baryB, baryC)
}

// seedPolytopeFace inserts a tetrahedron face into the polytope, choosing the
// winding so that the face normal points away from the opposite vertex (that
// is, outward). This is used only while seeding, to bootstrap the consistent
// outward winding that later faces inherit through the horizon expansion.
func (s *EPASolver) seedPolytopeFace(vertexA, vertexB, vertexC, opposite MinkowskiVertex) {
	edgeAB := dprec.Vec3Diff(vertexB.Position, vertexA.Position)
	edgeAC := dprec.Vec3Diff(vertexC.Position, vertexA.Position)
	normal := dprec.Vec3Cross(edgeAB, edgeAC)
	if dprec.Vec3Dot(normal, dprec.Vec3Diff(vertexA.Position, opposite.Position)) < 0 {
		vertexB, vertexC = vertexC, vertexB // flip so the normal points outward
	}
	s.addPolytopeTriangle(vertexA, vertexB, vertexC)
}

// addPolytopeTriangle inserts a face into the polytope, skipping degenerate
// faces. The winding must already be outward (its normal, computed by the
// right-hand rule from the vertex order, pointing away from the enclosed
// origin); the caller is responsible for providing it, either through
// [EPASolver.seedPolytopeFace] or through the horizon expansion, which
// preserves the outward winding.
func (s *EPASolver) addPolytopeTriangle(vertexA, vertexB, vertexC MinkowskiVertex) {
	edgeAB := dprec.Vec3Diff(vertexB.Position, vertexA.Position)
	edgeAC := dprec.Vec3Diff(vertexC.Position, vertexA.Position)
	cross := dprec.Vec3Cross(edgeAB, edgeAC)
	if cross.SqrLength() < 1e-12 {
		return // ignore degenerate (zero-area) faces
	}

	normal := dprec.UnitVec3(cross)
	id := PolytopeTriangleID{vertexA.Refs, vertexB.Refs, vertexC.Refs}
	triangle := PolytopeTriangle{
		VertexA:  vertexA,
		VertexB:  vertexB,
		VertexC:  vertexC,
		Normal:   normal,
		Distance: dprec.Vec3Dot(normal, vertexA.Position),
	}
	if index, ok := s.polytopeIndex[id]; ok {
		s.polytope[index] = triangle // replace an existing face with the same winding
	} else {
		s.polytopeIndex[id] = int32(len(s.polytope))
		s.polytope = append(s.polytope, triangle)
	}
}

// polytopeContainsVertex reports whether any polytope face already uses the
// vertex identified by the given refs.
func (s *EPASolver) polytopeContainsVertex(refs RefPair) bool {
	for _, triangle := range s.polytope {
		if triangle.VertexA.Refs == refs || triangle.VertexB.Refs == refs || triangle.VertexC.Refs == refs {
			return true
		}
	}
	return false
}

// addHorizonEdge accumulates a directed edge of a face that is being removed.
// An edge shared by two removed faces appears in both orientations and cancels
// out, so only the boundary of the removed region (the horizon) survives.
func addHorizonEdge(horizon map[polytopeEdgeID]polytopeEdge, vertexA, vertexB MinkowskiVertex) {
	reverse := polytopeEdgeID{vertexB.Refs, vertexA.Refs}
	if _, ok := horizon[reverse]; ok {
		delete(horizon, reverse)
		return
	}
	horizon[polytopeEdgeID{vertexA.Refs, vertexB.Refs}] = polytopeEdge{
		VertexA: vertexA,
		VertexB: vertexB,
	}
}

// closestPoint returns the point with the given barycentric weights over the
// triangle (vertexA, vertexB, vertexC).
func closestPoint(pointA, pointB, pointC dprec.Vec3, baryA, baryB, baryC float64) dprec.Vec3 {
	return dprec.Vec3MultiSum(
		dprec.Vec3Prod(pointA, baryA),
		dprec.Vec3Prod(pointB, baryB),
		dprec.Vec3Prod(pointC, baryC),
	)
}

// barycentricClosestToOrigin returns the barycentric weights (over the triangle
// A, B, C) of the point on the triangle that is closest to the origin. The
// weights are clamped to the triangle and sum to one. This follows the standard
// closest-point-on-triangle decomposition into vertex, edge and face regions.
func barycentricClosestToOrigin(pointA, pointB, pointC dprec.Vec3) (float64, float64, float64) {
	edgeAB := dprec.Vec3Diff(pointB, pointA)
	edgeAC := dprec.Vec3Diff(pointC, pointA)

	// Vectors from each vertex to the origin (origin - vertex).
	toOriginA := dprec.InverseVec3(pointA)
	d1 := dprec.Vec3Dot(edgeAB, toOriginA)
	d2 := dprec.Vec3Dot(edgeAC, toOriginA)
	if d1 <= 0 && d2 <= 0 {
		return 1, 0, 0 // vertex A region
	}

	toOriginB := dprec.InverseVec3(pointB)
	d3 := dprec.Vec3Dot(edgeAB, toOriginB)
	d4 := dprec.Vec3Dot(edgeAC, toOriginB)
	if d3 >= 0 && d4 <= d3 {
		return 0, 1, 0 // vertex B region
	}

	vc := d1*d4 - d3*d2
	if vc <= 0 && d1 >= 0 && d3 <= 0 {
		v := d1 / (d1 - d3)
		return 1 - v, v, 0 // edge AB region
	}

	toOriginC := dprec.InverseVec3(pointC)
	d5 := dprec.Vec3Dot(edgeAB, toOriginC)
	d6 := dprec.Vec3Dot(edgeAC, toOriginC)
	if d6 >= 0 && d5 <= d6 {
		return 0, 0, 1 // vertex C region
	}

	vb := d5*d2 - d1*d6
	if vb <= 0 && d2 >= 0 && d6 <= 0 {
		w := d2 / (d2 - d6)
		return 1 - w, 0, w // edge AC region
	}

	va := d3*d6 - d5*d4
	if va <= 0 && (d4-d3) >= 0 && (d5-d6) >= 0 {
		w := (d4 - d3) / ((d4 - d3) + (d5 - d6))
		return 0, 1 - w, w // edge BC region
	}

	denom := 1.0 / (va + vb + vc)
	v := vb * denom
	w := vc * denom
	return 1 - v - w, v, w // face interior
}

// EPASolution describes how two overlapping shapes can be separated.
type EPASolution struct {
	// VertexA, VertexB and VertexC are the Minkowski difference vertices of the
	// closest feature. For a point feature all three are equal; for an edge
	// feature VertexC equals VertexB.
	VertexA MinkowskiVertex
	// VertexB is the second vertex of the closest feature.
	VertexB MinkowskiVertex
	// VertexC is the third vertex of the closest feature.
	VertexC MinkowskiVertex
	// Normal is the unit direction along which the source shape must be
	// moved by Depth to separate the shapes. In world space it points from
	// the target shape toward the source shape. When the origin is outside
	// the Minkowski difference core it points from the closest feature
	// toward the origin; when the origin is contained it is the outward
	// normal of the closest boundary feature.
	Normal dprec.Vec3
	// BaryA, BaryB and BaryC are the barycentric weights of the closest point
	// to the origin over VertexA, VertexB and VertexC. They are non-negative
	// and sum to one.
	BaryA float64
	// BaryB is the barycentric weight of VertexB.
	BaryB float64
	// BaryC is the barycentric weight of VertexC.
	BaryC float64
	// Depth is the amount by which the shapes, inflated by their skin
	// radii, overlap along Normal.
	Depth float64
}

// PolytopeTriangleID identifies a polytope face by the ref pairs of its three
// vertices in winding order.
type PolytopeTriangleID [3]RefPair

// PolytopeTriangle is a face of the expanding polytope. Its normal points
// outward (away from the enclosed origin) and Distance is the perpendicular
// distance from the origin to the face plane.
type PolytopeTriangle struct {
	// VertexA is the first vertex of the face, in outward winding order.
	VertexA MinkowskiVertex
	// VertexB is the second vertex of the face, in outward winding order.
	VertexB MinkowskiVertex
	// VertexC is the third vertex of the face, in outward winding order.
	VertexC MinkowskiVertex
	// Normal is the outward unit normal of the face, pointing away from the
	// enclosed origin.
	Normal dprec.Vec3
	// Distance is the perpendicular distance from the origin to the face plane.
	Distance float64
}

// polytopeEdgeID identifies a directed polytope edge by the ref pairs of its
// two endpoints.
type polytopeEdgeID [2]RefPair

// polytopeEdge is a directed edge used while computing the horizon of the
// polytope region that is visible from a support point.
type polytopeEdge struct {
	// VertexA is the origin endpoint of the directed edge.
	VertexA MinkowskiVertex
	// VertexB is the destination endpoint of the directed edge.
	VertexB MinkowskiVertex
}
