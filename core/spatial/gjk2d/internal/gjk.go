package internal

import "github.com/mokiat/gomath/dprec"

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
// an edge simplex to a point simplex, which is not part of the standard GJK
// algorithm.

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
	searchDirection     dprec.Vec2
	sqrSkinRadius       float64
	remainingIterations uint32
	containsOrigin      bool
	overlapsOrigin      bool
}

// NewGJKSolver creates a new [GJKSolver] instance.
func NewGJKSolver() *GJKSolver {
	return &GJKSolver{}
}

// Reset prepares the solver for a new query against the given shape.
func (s *GJKSolver) Reset(shape *MinkowskiShape) {
	s.simplex = EmptySimplex()
	s.searchDirection = dprec.BasisXVec2()
	s.sqrSkinRadius = shape.SkinRadius * shape.SkinRadius
	s.remainingIterations = uint32(shape.MaxIterations())
	s.containsOrigin = false
	s.overlapsOrigin = false
}

// Next runs a single GJK iteration. It returns false once the algorithm has
// converged (or the iteration budget is exhausted) and further calls would
// make no progress.
func (s *GJKSolver) Next(shape *MinkowskiShape) bool {
	if s.remainingIterations == 0 {
		return false
	}
	s.remainingIterations--

	point := shape.Support(s.searchDirection)
	if s.simplex.HasVertex(point) {
		return false // the simplex is not growing anymore
	}

	// Note: If we had a triangle simplex, we would have already completed.
	switch s.simplex.VertexCount {
	case 0: // currently empty
		return s.appendToEmpty(point)
	case 1: // currently point
		return s.appendToPoint(point)
	default: // currently edge
		return s.appendToEdge(point)
	}
}

// Simplex returns the current simplex. For simplexes with at least two
// vertices, the solver maintains the invariant that the origin lies on the
// left side of the directed edge from Vertices[0] to Vertices[1], which
// corresponds to a counter-clockwise winding around the origin. In
// particular, the triangle simplex produced when the origin is contained
// is wound counter-clockwise and encloses the origin.
func (s *GJKSolver) Simplex() Simplex {
	return s.simplex
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
	s.searchDirection = dprec.InverseVec2(vertex.Position)

	return true
}

// appendToPoint grows the point simplex into an edge simplex, or replaces
// it with a point simplex at the new vertex when the origin lies in that
// vertex's Voronoi region.
func (s *GJKSolver) appendToPoint(vertex MinkowskiVertex) bool {
	if !s.crossedSkinPlane(vertex.Position) {
		// The new vertex is not past the plane that lies skin-radius distance
		// behind the origin, opposite the search direction.
		// Not only are we not able to construct a triangle simplex that contains
		// the origin, but we also won't be able to find a closest feature that is
		// within skin-radius distance of the origin.
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

	edgeDir := dprec.Vec2Diff(vertB.Position, vertA.Position)

	isPastOrigin := dprec.Vec2Dot(edgeDir, vertB.Position) > 0
	if !isPastOrigin {
		// At this point we know that we cannot construct a triangle simplex that
		// contains the origin. Additionally, the origin does not project onto the
		// edge, hence AB is not the closest feature. However, it could be vertex B
		// or some other edge. We downgrade to a point simplex at B and continue
		// searching for the closest feature.
		s.simplex = PointSimplex(vertB)
		s.searchDirection = dprec.InverseVec2(vertB.Position)
		return true
	}

	edgeNorm := transposeVec2(edgeDir)
	edgeDot := dprec.Vec2Dot(edgeNorm, vertA.Position)

	isEdgeWithinSkinRadius := edgeDot*edgeDot <= edgeNorm.SqrLength()*s.sqrSkinRadius
	if isEdgeWithinSkinRadius {
		// The edge is within skin-radius distance of the origin, which means that at
		// minimum the two shapes touch at their skin radius.
		s.overlapsOrigin = true
	}

	isFacingOrigin := edgeDot < 0
	if isFacingOrigin {
		// Configure next iteration.
		s.simplex = EdgeSimplex(vertB, vertA) // flip the edge so that the origin is behind
		s.searchDirection = edgeNorm
	} else {
		// Configure next iteration.
		s.simplex = EdgeSimplex(vertA, vertB) // preserve the edge so that the origin is behind
		s.searchDirection = dprec.InverseVec2(edgeNorm)
	}

	return true
}

// appendToEdge completes the edge simplex into a triangle simplex when the
// origin is contained. Otherwise it advances the simplex toward the origin,
// keeping the edge closest to it, or downgrading to a point simplex at the
// new vertex when the origin lies in that vertex's Voronoi region.
func (s *GJKSolver) appendToEdge(vertex MinkowskiVertex) bool {
	if !s.crossedSkinPlane(vertex.Position) {
		// The new vertex is not past the plane that lies skin-radius distance
		// behind the origin, opposite the search direction.
		// Not only are we not able to construct a triangle simplex that contains
		// the origin, but we also won't be able to find a closest feature that is
		// within skin-radius distance of the origin.
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

	edgeAC := dprec.Vec2Diff(vertC.Position, vertA.Position)
	edgeBC := dprec.Vec2Diff(vertC.Position, vertB.Position)

	cross := dprec.Vec2Cross(edgeAC, edgeBC)
	if cross*cross <= dprec.Epsilon*edgeAC.SqrLength()*edgeBC.SqrLength() {
		// The three vertices are (near) collinear, so the Minkowski difference is
		// flat here and the support toward the origin did not advance past edge
		// AB. The origin therefore cannot be strictly contained by the simplex,
		// and since no progress was made we stop iterating, keeping edge AB as
		// the closest feature. The comparison is inclusive so that it also
		// catches a support vertex C that coincides with A or B: the zeroed edge
		// length makes both sides zero, which a strict comparison would miss.
		s.terminate(false)
		return false
	}

	projectsToAC := dprec.Vec2Dot(edgeAC, vertC.Position) > 0
	projectsToBC := dprec.Vec2Dot(edgeBC, vertC.Position) > 0

	if !projectsToAC && !projectsToBC {
		// We are in the Voronoi region of C, which means that the origin cannot
		// be contained by the simplex. Nevertheless, we may still be able to find a
		// closest feature that is within skin-radius distance of the origin.
		// However, we can be sure that neither edge BC nor edge AC is the closest.
		// At best it is vertex C or some other edge. We downgrade to a point
		// simplex at C and continue searching for the closest feature.
		s.simplex = PointSimplex(vertC)
		s.searchDirection = dprec.InverseVec2(vertC.Position)
		return true
	}

	normCA := transposeVec2(dprec.InverseVec2(edgeAC))
	normBC := transposeVec2(edgeBC)

	dotBC := dprec.Vec2Dot(normBC, vertC.Position)
	dotCA := dprec.Vec2Dot(normCA, vertC.Position)

	isBCFacingOrigin := dotBC < 0
	isCAFacingOrigin := dotCA < 0

	switch {
	case isBCFacingOrigin && projectsToBC:
		isEdgeWithinSkinRadius := dotBC*dotBC <= normBC.SqrLength()*s.sqrSkinRadius
		if isEdgeWithinSkinRadius {
			// The edge is within skin-radius distance of the origin, which means that at
			// minimum the two shapes touch at their skin radius.
			s.overlapsOrigin = true
		}

		// Configure next iteration.
		s.simplex = EdgeSimplex(vertC, vertB) // flip the edge so that the origin is behind
		s.searchDirection = normBC
		return true

	case isCAFacingOrigin && projectsToAC:
		isEdgeWithinSkinRadius := dotCA*dotCA <= normCA.SqrLength()*s.sqrSkinRadius
		if isEdgeWithinSkinRadius {
			// The edge is within skin-radius distance of the origin, which means that at
			// minimum the two shapes touch at their skin radius.
			s.overlapsOrigin = true
		}

		// Configure next iteration.
		s.simplex = EdgeSimplex(vertA, vertC) // preserve the edge so that the origin is behind
		s.searchDirection = normCA
		return true

	case !isBCFacingOrigin && !isCAFacingOrigin:
		s.simplex = TriangleSimplex(vertA, vertB, vertC)
		s.terminate(true)
		return false

	default:
		// Handle precision issues by downgrading to a point simplex at C.
		s.simplex = PointSimplex(vertC)
		s.searchDirection = dprec.InverseVec2(vertC.Position)
		return true
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
func (s *GJKSolver) isWithinSkinRadius(point dprec.Vec2) bool {
	return point.SqrLength() <= s.sqrSkinRadius
}

// crossedSkinPlane checks if the point is past the plane that lies
// skin-radius distance behind the origin, opposite the search direction.
//
// The point is the furthest point of the Minkowski difference along the
// search direction. If even it does not reach that plane, then the whole
// Minkowski difference is more than skin-radius away from the origin and
// the query can be answered negatively right away.
func (s *GJKSolver) crossedSkinPlane(point dprec.Vec2) bool {
	dot := dprec.Vec2Dot(point, s.searchDirection)
	if dot >= 0 {
		return true // the point is past the plane at the origin so we are good
	}
	return dot*dot <= s.searchDirection.SqrLength()*s.sqrSkinRadius
}
