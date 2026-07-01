package internal

import "github.com/mokiat/gomath/dprec"

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

	// NOTE: If we had a triangle simplex, we would have already completed.
	switch s.simplex.VertexCount {
	case 0: // currently empty
		return s.appendToEmpty(point)
	case 1: // currently point
		return s.appendToPoint(point)
	default: // currently edge
		return s.appendToEdge(point)
	}
}

// Simplex returns the current simplex. The solver maintains the invariant
// that the origin lies on the left side of the directed edge from
// Vertices[0] to Vertices[1], which corresponds to a counter-clockwise
// winding around the origin. In particular, the triangle simplex produced
// when the origin is contained is wound counter-clockwise and encloses
// the origin.
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
// stays true and iteration may be stopped early.
func (s *GJKSolver) OverlapsOrigin() bool {
	return s.overlapsOrigin
}

func (s *GJKSolver) appendToEmpty(vertex MinkowskiVertex) bool {
	// Check if the new vertex is within skin-radius distance of the origin.
	if vertex.Position.SqrLength() <= s.sqrSkinRadius {
		s.overlapsOrigin = true
	}

	// Configure next iteration.
	s.simplex = PointSimplex(vertex)
	s.searchDirection = dprec.InverseVec2(vertex.Position)

	return true
}

func (s *GJKSolver) appendToPoint(vertex MinkowskiVertex) bool {
	// Check if the new vertex is at all applicable.
	if !s.crossedSkinPlane(vertex.Position) {
		s.remainingIterations = 0 // ensure we can't be asked to iterate further
		return false
	}

	// Check if the new vertex is within skin-radius distance of the origin.
	if vertex.Position.SqrLength() <= s.sqrSkinRadius {
		s.overlapsOrigin = true
	}

	// Name vertices for clarity.
	vertA := s.simplex.Vertices[0]
	vertB := vertex

	edgeDir := dprec.Vec2Diff(
		vertB.Position,
		vertA.Position,
	)
	edgeNorm := transposeVec2(edgeDir)
	edgeDot := dprec.Vec2Dot(edgeNorm, vertA.Position)

	// Check if the edge is within skin-radius distance of the origin.
	if originProjectsToEdge(vertA.Position, vertB.Position) {
		if edgeDot*edgeDot <= edgeNorm.SqrLength()*s.sqrSkinRadius {
			s.overlapsOrigin = true
		}
	}

	isFacingOrigin := edgeDot < 0
	if isFacingOrigin {
		// Configure next iteration.
		s.simplex = EdgeSimplex(vertB, vertA) // origin on the left of B->A
		s.searchDirection = edgeNorm
	} else {
		// Configure next iteration.
		s.simplex = EdgeSimplex(vertA, vertB) // origin on the left of A->B
		s.searchDirection = dprec.InverseVec2(edgeNorm)
	}

	return true
}

func (s *GJKSolver) appendToEdge(vertex MinkowskiVertex) bool {
	// Check if the new vertex is at all applicable.
	if !s.crossedSkinPlane(vertex.Position) {
		s.remainingIterations = 0 // ensure we can't be asked to iterate further
		return false
	}

	// Check if the new vertex is within skin-radius distance of the origin.
	if vertex.Position.SqrLength() <= s.sqrSkinRadius {
		s.overlapsOrigin = true
	}

	// Name vertices for clarity.
	vertA := s.simplex.Vertices[0]
	vertB := s.simplex.Vertices[1]
	vertC := vertex

	normBC := transposeVec2(dprec.Vec2Diff(vertC.Position, vertB.Position))
	normCA := transposeVec2(dprec.Vec2Diff(vertA.Position, vertC.Position))

	dotBC := dprec.Vec2Dot(normBC, vertC.Position)
	dotCA := dprec.Vec2Dot(normCA, vertC.Position)

	isBCFacingOrigin := dotBC < 0
	isCAFacingOrigin := dotCA < 0

	switch {
	case !isBCFacingOrigin && !isCAFacingOrigin:
		s.simplex = TriangleSimplex(vertA, vertB, vertC)
		s.containsOrigin = true
		s.overlapsOrigin = true
		s.remainingIterations = 0 // ensure we can't be asked to iterate further
		return false

	case isBCFacingOrigin && !isCAFacingOrigin:
		// Check if the BC edge is within skin-radius distance of the origin.
		if originProjectsToEdge(vertB.Position, vertC.Position) {
			if dotBC*dotBC <= normBC.SqrLength()*s.sqrSkinRadius {
				s.overlapsOrigin = true
			}
		}

		// Configure next iteration.
		s.simplex = EdgeSimplex(vertC, vertB) // origin on the left of C->B
		s.searchDirection = normBC
		return true

	case isCAFacingOrigin && !isBCFacingOrigin:
		// Check if the CA edge is within skin-radius distance of the origin.
		if originProjectsToEdge(vertC.Position, vertA.Position) {
			if dotCA*dotCA <= normCA.SqrLength()*s.sqrSkinRadius {
				s.overlapsOrigin = true
			}
		}

		// Configure next iteration.
		s.simplex = EdgeSimplex(vertA, vertC) // origin on the left of A->C
		s.searchDirection = normCA
		return true

	default:
		// Check if the origin is within skin-radius distance of the BC edge.
		// This can happen if the C angle is larger than 90 degrees.
		if originProjectsToEdge(vertB.Position, vertC.Position) {
			if dotBC*dotBC <= normBC.SqrLength()*s.sqrSkinRadius {
				s.overlapsOrigin = true
			}

			// Configure next iteration.
			s.simplex = EdgeSimplex(vertC, vertB) // origin on the left of C->B
			s.searchDirection = normBC
			return true
		}

		// Check if the origin is within skin-radius distance of the CA edge.
		// This can happen if the C angle is larger than 90 degrees.
		if originProjectsToEdge(vertC.Position, vertA.Position) {
			if dotCA*dotCA <= normCA.SqrLength()*s.sqrSkinRadius {
				s.overlapsOrigin = true
			}

			// Configure next iteration.
			s.simplex = EdgeSimplex(vertA, vertC) // origin on the left of A->C
			s.searchDirection = normCA
			return true
		}

		// At this point there is no way that the origin can be contained
		// by the Minkowski difference. If we know that the origin is within
		// skin-radius distance of the Minkowski difference, we can stop here.
		if s.overlapsOrigin {
			s.remainingIterations = 0 // ensure we can't be asked to iterate further
			return false
		}

		// Otherwise we need to continue searching for the closest feature to
		// the origin. But since we can't construct a triangle simplex, we might
		// as well make our life easy by continuing from the a point-simplex at C.
		s.simplex = PointSimplex(vertC)
		s.searchDirection = dprec.InverseVec2(vertC.Position)
		return true
	}
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
