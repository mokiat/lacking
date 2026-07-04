package internal

import (
	"math"

	"github.com/mokiat/gomath/dprec"
)

// Implementation note:
//
// We are following the standard GJK algorithm by exploring Voronoi regions
// towards the origin.
//
// The only difference is that if we determine that the origin cannot be
// contained within the Minkowski difference, we continue searching for
// the closest feature until we find one or we determine that even that
// is more than skin-radius away from the origin.
// While searching for such a feature, it is possible to downgrade from
// a triangle or edge simplex to a point simplex, which is not part of the
// standard GJK algorithm.

// GJKSolver implements the GJK algorithm over a [MinkowskiShape]. It answers
// whether the origin lies inside the shape (containment) and whether the
// origin lies within the shape's skin radius (overlap).
//
// Usage: call [GJKSolver.Reset] with the shape, then call [GJKSolver.Next]
// repeatedly until it returns false, then inspect [GJKSolver.OverlapsOrigin]
// and [GJKSolver.ContainsOrigin]. Callers interested only in overlap may
// stop iterating as soon as OverlapsOrigin returns true.
type GJKSolver struct {
	simplex             Simplex
	searchDirection     dprec.Vec3
	sqrSkinRadius       float64
	remainingIterations uint32
	containsOrigin      bool
	overlapsOrigin      bool

	// closestSimplex tracks the simplex whose closest point to the origin is
	// nearest, across all iterations. The custom closest-feature search can
	// oscillate between features without the current simplex settling on the
	// closest one, so the nearest simplex seen is retained and reported for the
	// non-contained case (see [GJKSolver.Simplex]). This makes the reported
	// feature robust; in rare grazing configurations the oscillation may never
	// visit the exact closest feature, in which case the retained simplex is a
	// very close approximation.
	closestSimplex     Simplex
	closestSqrDistance float64
}

// NewGJKSolver creates a new [GJKSolver] instance.
func NewGJKSolver() *GJKSolver {
	return &GJKSolver{}
}

// Reset prepares the solver for a new query against the given shape.
func (s *GJKSolver) Reset(shape *MinkowskiShape) {
	s.simplex = EmptySimplex()
	s.searchDirection = dprec.BasisXVec3()
	s.sqrSkinRadius = shape.SkinRadius * shape.SkinRadius
	s.remainingIterations = uint32(shape.MaxIterations())
	s.containsOrigin = false
	s.overlapsOrigin = false
	s.closestSimplex = EmptySimplex()
	s.closestSqrDistance = math.MaxFloat64
}

// Next runs a single GJK iteration. It returns false once the algorithm has
// converged (or the iteration budget is exhausted) and further calls would
// make no progress.
func (s *GJKSolver) Next(shape *MinkowskiShape) bool {
	s.recordClosest()
	if s.remainingIterations == 0 {
		return false
	}
	s.remainingIterations--

	point := shape.Support(s.searchDirection)
	if s.simplex.HasVertex(point) {
		return false // the simplex is not growing anymore
	}

	// Note: If we had a tetrahedron simplex, we would have already completed.
	switch s.simplex.VertexCount {
	case 0: // currently empty
		return s.appendToEmpty(point)
	case 1: // currently point
		return s.appendToPoint(point)
	case 2: // currently edge
		return s.appendToEdge(point)
	default: // currently triangle
		return s.appendToTriangle(point)
	}
}

// Simplex returns the terminal simplex of the query. When the origin is
// contained, this is the tetrahedron (or, when the origin lies exactly on an
// intermediate triangle, the triangle) that witnesses the containment. When
// the origin is not contained, it is the point, edge or triangle feature of
// the Minkowski difference closest to the origin.
//
// For triangle simplexes, the solver maintains the invariant that the normal
// implied by the vertex order (the cross product of the Vertices[0] to
// Vertices[1] and Vertices[0] to Vertices[2] edges) points towards the origin.
// The tetrahedron simplex produced when the origin is contained preserves that
// order for its first three vertices, with the fourth vertex on the same side
// as the origin.
func (s *GJKSolver) Simplex() Simplex {
	if s.containsOrigin {
		return s.simplex
	}
	// The closest-feature search may not leave the current simplex on the
	// closest feature, so return the nearest simplex seen, including the
	// current one.
	if simplexSqrDistance(s.simplex) < s.closestSqrDistance {
		return s.simplex
	}
	return s.closestSimplex
}

// recordClosest updates the nearest simplex seen so far with the current
// simplex, so that the non-contained case can report the closest feature even
// when the closest-feature search oscillates.
func (s *GJKSolver) recordClosest() {
	if distance := simplexSqrDistance(s.simplex); distance < s.closestSqrDistance {
		s.closestSqrDistance = distance
		s.closestSimplex = s.simplex
	}
}

// ContainsOrigin reports whether the origin is inside the Minkowski
// difference core (ignoring the skin radius). The result is only reliable
// once [GJKSolver.Next] has been iterated until it returned false.
func (s *GJKSolver) ContainsOrigin() bool {
	return s.containsOrigin
}

// OverlapsOrigin reports whether the origin is within skin-radius distance
// of the Minkowski difference (which includes containment). Once true, it
// stays true and iteration may be stopped early. A false result is only
// reliable once [GJKSolver.Next] has been iterated until it returned false.
func (s *GJKSolver) OverlapsOrigin() bool {
	return s.overlapsOrigin
}

// appendToEmpty seeds the empty simplex with the first support vertex.
func (s *GJKSolver) appendToEmpty(vertex MinkowskiVertex) bool {
	if s.isWithinSkinRadius(vertex.Position) {
		// The new point is within skin-radius distance of the origin, which means
		// that at minimum the two shapes touch at their skin radius.
		s.overlapsOrigin = true
	}

	// Configure next iteration.
	s.simplex = PointSimplex(vertex)
	s.searchDirection = dprec.InverseVec3(vertex.Position)

	return true
}

// appendToPoint grows the point simplex into an edge simplex, or replaces
// it with a point simplex at the new vertex when the origin lies in that
// vertex's Voronoi region.
func (s *GJKSolver) appendToPoint(vertex MinkowskiVertex) bool {
	if !s.crossedSkinPlane(vertex.Position) {
		// The new vertex is not past the plane that lies skin-radius distance
		// behind the origin, opposite the search direction.
		// Not only are we not able to construct a tetrahedron simplex that
		// contains the origin, but we also won't be able to find a closest
		// feature that is within skin-radius distance of the origin.
		s.terminate(false)
		return false
	}

	if s.isWithinSkinRadius(vertex.Position) {
		// The new point is within skin-radius distance of the origin, which means
		// that at minimum the two shapes touch at their skin radius.
		s.overlapsOrigin = true
	}

	// Name vertices for clarity.
	vertA := s.simplex.Vertices[0]
	vertB := vertex

	edgeDir := dprec.Vec3Diff(vertB.Position, vertA.Position)

	isPastOrigin := dprec.Vec3Dot(edgeDir, vertB.Position) > 0
	if !isPastOrigin {
		// At this point we know that we cannot construct a tetrahedron simplex
		// that contains the origin. Additionally, the origin does not project
		// onto the edge, hence AB is not the closest feature. However, it could
		// be vertex B or some other edge. We downgrade to a point simplex at B
		// and continue searching for the closest feature.
		s.simplex = PointSimplex(vertB)
		s.searchDirection = dprec.InverseVec3(vertB.Position)
		return true
	}

	// Configure next iteration.
	s.advanceToEdge(vertA, vertB, edgeDir)
	return true
}

// appendToEdge grows the edge simplex into a triangle simplex, or advances
// the simplex toward the origin, keeping the feature closest to it, or
// downgrading to a point simplex at the new vertex when the origin lies in
// that vertex's Voronoi region.
func (s *GJKSolver) appendToEdge(vertex MinkowskiVertex) bool {
	if !s.crossedSkinPlane(vertex.Position) {
		// The new vertex is not past the plane that lies skin-radius distance
		// behind the origin, opposite the search direction.
		// Not only are we not able to construct a tetrahedron simplex that
		// contains the origin, but we also won't be able to find a closest
		// feature that is within skin-radius distance of the origin.
		s.terminate(false)
		return false
	}

	// Check if the new vertex is within skin-radius distance of the origin.
	if s.isWithinSkinRadius(vertex.Position) {
		// The new point is within skin-radius distance of the origin, which means
		// that at minimum the two shapes touch at their skin radius.
		s.overlapsOrigin = true
	}

	// Name vertices for clarity.
	vertA := s.simplex.Vertices[0]
	vertB := s.simplex.Vertices[1]
	vertC := vertex

	edgeAC := dprec.Vec3Diff(vertC.Position, vertA.Position)
	edgeBC := dprec.Vec3Diff(vertC.Position, vertB.Position)

	projectsToAC := dprec.Vec3Dot(edgeAC, vertC.Position) > 0
	projectsToBC := dprec.Vec3Dot(edgeBC, vertC.Position) > 0

	if !projectsToAC && !projectsToBC {
		// We are in the Voronoi region of C, which means that the origin cannot
		// be contained by the simplex. Nevertheless, we may still be able to find a
		// closest feature that is within skin-radius distance of the origin.
		// However, we can be sure that neither edge BC nor edge AC is the closest.
		// At best it is vertex C or some other feature. We downgrade to a point
		// simplex at C and continue searching for the closest feature.
		s.simplex = PointSimplex(vertC)
		s.searchDirection = dprec.InverseVec3(vertC.Position)
		return true
	}

	return s.resolveTriangle(vertA, vertB, vertC, projectsToAC, projectsToBC)
}

// appendToTriangle completes the triangle simplex into a tetrahedron simplex
// when the origin is contained. Otherwise it advances the simplex toward the
// origin, keeping the face or edge closest to it, or downgrading to a point
// simplex at the new vertex when the origin lies in that vertex's Voronoi
// region.
func (s *GJKSolver) appendToTriangle(vertex MinkowskiVertex) bool {
	if !s.crossedSkinPlane(vertex.Position) {
		// The new vertex is not past the plane that lies skin-radius distance
		// behind the origin, opposite the search direction.
		// Not only are we not able to construct a tetrahedron simplex that
		// contains the origin, but we also won't be able to find a closest
		// feature that is within skin-radius distance of the origin.
		s.terminate(false)
		return false
	}

	// Check if the new vertex is within skin-radius distance of the origin.
	if s.isWithinSkinRadius(vertex.Position) {
		// The new point is within skin-radius distance of the origin, which means
		// that at minimum the two shapes touch at their skin radius.
		s.overlapsOrigin = true
	}

	// Name vertices for clarity.
	vertA := s.simplex.Vertices[0]
	vertB := s.simplex.Vertices[1]
	vertC := s.simplex.Vertices[2]
	vertD := vertex

	edgeAD := dprec.Vec3Diff(vertD.Position, vertA.Position)
	edgeBD := dprec.Vec3Diff(vertD.Position, vertB.Position)
	edgeCD := dprec.Vec3Diff(vertD.Position, vertC.Position)

	volume := dprec.Vec3Dot(edgeAD, dprec.Vec3Cross(edgeBD, edgeCD))
	if volume*volume <= dprec.Epsilon*edgeAD.SqrLength()*edgeBD.SqrLength()*edgeCD.SqrLength() {
		// The four vertices are (near) coplanar, so the Minkowski difference is
		// flat here and the support toward the origin did not advance past the
		// ABC plane. The origin therefore cannot be strictly contained by the
		// simplex, and since no progress was made we stop iterating, keeping
		// triangle ABC as the closest feature. The comparison is inclusive so
		// that it also catches a support vertex D that coincides with A, B or C:
		// the zeroed edge length makes both sides zero, which a strict
		// comparison would miss.
		s.terminate(false)
		return false
	}

	projectsToAD := dprec.Vec3Dot(edgeAD, vertD.Position) > 0
	projectsToBD := dprec.Vec3Dot(edgeBD, vertD.Position) > 0
	projectsToCD := dprec.Vec3Dot(edgeCD, vertD.Position) > 0

	if !projectsToAD && !projectsToBD && !projectsToCD {
		// We are in the Voronoi region of D, which means that the origin cannot
		// be contained by the simplex. Nevertheless, we may still be able to find
		// a closest feature that is within skin-radius distance of the origin.
		// However, we can be sure that none of the edges AD, BD and CD is the
		// closest. At best it is vertex D or some other feature. We downgrade to
		// a point simplex at D and continue searching for the closest feature.
		s.simplex = PointSimplex(vertD)
		s.searchDirection = dprec.InverseVec3(vertD.Position)
		return true
	}

	// Outward normals of the three tetrahedron faces that contain D. The
	// triangle simplex invariant (implied normal facing the origin) together
	// with D lying on the origin side of the ABC plane guarantees that these
	// cross products point away from the opposite tetrahedron vertex.
	normABD := dprec.Vec3Cross(edgeAD, edgeBD)
	normBCD := dprec.Vec3Cross(edgeBD, edgeCD)
	normCAD := dprec.Vec3Cross(edgeCD, edgeAD)

	dotABD := dprec.Vec3Dot(normABD, vertD.Position)
	dotBCD := dprec.Vec3Dot(normBCD, vertD.Position)
	dotCAD := dprec.Vec3Dot(normCAD, vertD.Position)

	switch {
	case dotABD < 0:
		// The origin is outside the face ABD, so the closest feature is a
		// feature of that face.
		return s.resolveTriangle(vertA, vertB, vertD, projectsToAD, projectsToBD)

	case dotBCD < 0:
		// The origin is outside the face BCD, so the closest feature is a
		// feature of that face.
		return s.resolveTriangle(vertB, vertC, vertD, projectsToBD, projectsToCD)

	case dotCAD < 0:
		// The origin is outside the face CAD, so the closest feature is a
		// feature of that face.
		return s.resolveTriangle(vertC, vertA, vertD, projectsToCD, projectsToAD)

	default:
		// The origin is on the inner side of all three faces that contain D and,
		// by the triangle simplex invariant, on the inner side of the ABC face
		// as well, hence it is contained by the tetrahedron.
		s.simplex = TetrahedronSimplex(vertA, vertB, vertC, vertD)
		s.terminate(true)
		return false
	}
}

// resolveTriangle advances the simplex given the triangle (P, Q, N), where N
// is the most recently discovered vertex and the origin is known not to lie
// in the Voronoi region of N. It selects between the two edges that contain
// N and the triangle face, and configures the next iteration accordingly.
func (s *GJKSolver) resolveTriangle(vertP, vertQ, vertN MinkowskiVertex, projectsToPN, projectsToQN bool) bool {
	edgePN := dprec.Vec3Diff(vertN.Position, vertP.Position)
	edgeQN := dprec.Vec3Diff(vertN.Position, vertQ.Position)

	faceNorm := dprec.Vec3Cross(edgePN, edgeQN)
	if faceNorm.SqrLength() <= dprec.Epsilon*edgePN.SqrLength()*edgeQN.SqrLength() {
		// The triangle is (near) degenerate (its vertices are collinear), so its
		// Voronoi regions cannot be evaluated. Handle this by downgrading to a
		// point simplex at N and continuing the search. The comparison is
		// inclusive so that it also catches a vertex N that coincides with P or
		// Q: the zeroed edge length makes both sides zero, which a strict
		// comparison would miss.
		s.simplex = PointSimplex(vertN)
		s.searchDirection = dprec.InverseVec3(vertN.Position)
		return true
	}

	// In-plane normals of the two edges that contain N, pointing away from
	// the triangle interior.
	outPN := dprec.Vec3Cross(faceNorm, edgePN)
	outQN := dprec.Vec3Cross(edgeQN, faceNorm)

	dotPN := dprec.Vec3Dot(outPN, vertN.Position)
	dotQN := dprec.Vec3Dot(outQN, vertN.Position)

	isBeyondPN := dotPN < 0
	isBeyondQN := dotQN < 0

	switch {
	case isBeyondQN && projectsToQN:
		// The origin projects onto edge QN and is outside the triangle across
		// that edge, so edge QN is the closest feature candidate.
		s.advanceToEdge(vertQ, vertN, edgeQN)
		return true

	case isBeyondPN && projectsToPN:
		// The origin projects onto edge PN and is outside the triangle across
		// that edge, so edge PN is the closest feature candidate.
		s.advanceToEdge(vertP, vertN, edgePN)
		return true

	case !isBeyondPN && !isBeyondQN:
		// The origin projects onto the triangle interior.
		faceDot := dprec.Vec3Dot(faceNorm, vertN.Position)

		isFaceWithinSkinRadius := faceDot*faceDot <= faceNorm.SqrLength()*s.sqrSkinRadius
		if isFaceWithinSkinRadius {
			// The face is within skin-radius distance of the origin, which means
			// that at minimum the two shapes touch at their skin radius.
			s.overlapsOrigin = true
		}

		switch {
		case faceDot < 0:
			// Configure next iteration.
			s.simplex = TriangleSimplex(vertP, vertQ, vertN) // implied normal faces the origin
			s.searchDirection = faceNorm
			return true
		case faceDot > 0:
			// Configure next iteration.
			s.simplex = TriangleSimplex(vertQ, vertP, vertN) // flip the winding so that the implied normal faces the origin
			s.searchDirection = dprec.InverseVec3(faceNorm)
			return true
		default:
			// The origin lies exactly on the triangle, hence it touches the
			// Minkowski difference.
			s.simplex = TriangleSimplex(vertP, vertQ, vertN)
			s.terminate(true)
			return false
		}

	default:
		// Handle precision issues by downgrading to a point simplex at N.
		s.simplex = PointSimplex(vertN)
		s.searchDirection = dprec.InverseVec3(vertN.Position)
		return true
	}
}

// advanceToEdge replaces the simplex with the edge (F, N), where edgeDir is
// the direction from F to N, and points the search direction from the edge
// toward the origin. It also checks whether the edge line is within
// skin-radius distance of the origin.
func (s *GJKSolver) advanceToEdge(vertF, vertN MinkowskiVertex, edgeDir dprec.Vec3) {
	// edgeCross is perpendicular to the plane spanned by the edge direction
	// and the position of the edge relative to the origin. Its length encodes
	// the distance from the origin to the edge line (scaled by the edge
	// direction length).
	edgeCross := dprec.Vec3Cross(edgeDir, vertN.Position)

	isEdgeWithinSkinRadius := edgeCross.SqrLength() <= edgeDir.SqrLength()*s.sqrSkinRadius
	if isEdgeWithinSkinRadius {
		// The edge is within skin-radius distance of the origin, which means that at
		// minimum the two shapes touch at their skin radius.
		s.overlapsOrigin = true
	}

	// Configure next iteration.
	s.simplex = EdgeSimplex(vertF, vertN)
	if edgeCross.SqrLength() == 0 {
		// The origin lies on the edge line, so any direction perpendicular to
		// the edge is a valid search direction.
		s.searchDirection = perpendicularVec3(edgeDir)
	} else {
		s.searchDirection = dprec.Vec3Cross(edgeDir, edgeCross) // points from the edge toward the origin
	}
}

// terminate prevents any further iteration and records whether the origin
// was determined to be contained in the Minkowski difference.
func (s *GJKSolver) terminate(containsOrigin bool) {
	s.remainingIterations = 0 // ensure we can't be asked to iterate further
	if containsOrigin {
		s.containsOrigin = true
		s.overlapsOrigin = true
	}
}

// isWithinSkinRadius reports whether the point is within skin-radius
// distance of the origin.
func (s *GJKSolver) isWithinSkinRadius(point dprec.Vec3) bool {
	return point.SqrLength() <= s.sqrSkinRadius
}

// crossedSkinPlane checks if the point is past the plane that lies
// skin-radius distance behind the origin, opposite the search direction.
//
// The point is the furthest point of the Minkowski difference along the
// search direction. If even it does not reach that plane, then the whole
// Minkowski difference is more than skin-radius away from the origin and
// the query can be answered negatively right away.
func (s *GJKSolver) crossedSkinPlane(point dprec.Vec3) bool {
	dot := dprec.Vec3Dot(point, s.searchDirection)
	if dot >= 0 {
		return true // the point is past the plane at the origin so we are good
	}
	return dot*dot <= s.searchDirection.SqrLength()*s.sqrSkinRadius
}

// simplexSqrDistance returns the squared distance from the origin to the point
// of the simplex that is closest to it. An empty simplex is treated as
// infinitely far.
func simplexSqrDistance(simplex Simplex) float64 {
	switch simplex.VertexCount {
	case 1:
		return simplex.Vertices[0].Position.SqrLength()
	case 2:
		pointA := simplex.Vertices[0].Position
		pointB := simplex.Vertices[1].Position
		edge := dprec.Vec3Diff(pointB, pointA)
		sqrLength := edge.SqrLength()
		if sqrLength < 1e-12 {
			return pointA.SqrLength()
		}
		lerp := max(0.0, min(-dprec.Vec3Dot(edge, pointA)/sqrLength, 1.0))
		return dprec.Vec3Lerp(pointA, pointB, lerp).SqrLength()
	case 3:
		pointA := simplex.Vertices[0].Position
		pointB := simplex.Vertices[1].Position
		pointC := simplex.Vertices[2].Position
		baryA, baryB, baryC := barycentricClosestToOrigin(pointA, pointB, pointC)
		return closestPoint(pointA, pointB, pointC, baryA, baryB, baryC).SqrLength()
	default:
		return math.MaxFloat64
	}
}
