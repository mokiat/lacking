package internal

import "github.com/mokiat/gomath/dprec"

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

func (s *GJKSolver) Reset(shape *MinkowskiShape) {
	s.simplex = EmptySimplex()
	s.searchDirection = dprec.BasisXVec2()
	s.sqrSkinRadius = shape.SkinRadius * shape.SkinRadius
	s.remainingIterations = uint32(shape.MaxIterations())
	s.containsOrigin = false
	s.overlapsOrigin = false
}

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

func (s *GJKSolver) Simplex() Simplex {
	// The simplex is always stored in CW order internally, which eases
	// contruction. Flip it to CCW order for external consumption.
	return Simplex{
		Vertices: [3]MinkowskiVertex{
			s.simplex.Vertices[1],
			s.simplex.Vertices[0],
			s.simplex.Vertices[2],
		},
		VertexCount: s.simplex.VertexCount,
	}
}

func (s *GJKSolver) ContainsOrigin() bool {
	return s.containsOrigin
}

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
		s.simplex = EdgeSimplex(vertB, vertA) // pointing away from origin
		s.searchDirection = edgeNorm
	} else {
		// Configure next iteration.
		s.simplex = EdgeSimplex(vertA, vertB) // pointing away from origin
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
		s.simplex = EdgeSimplex(vertC, vertB) // pointing away from origin
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
		s.simplex = EdgeSimplex(vertA, vertC) // pointing away from origin
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
			s.simplex = EdgeSimplex(vertC, vertB) // pointing away from origin
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
			s.simplex = EdgeSimplex(vertA, vertC) // pointing away from origin
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

// crossedSkinPlane checks if the point is past the plane defined by the
// origin and the skin radius along the inverse of the last search direction.
//
// If the furthers point along the last search direction never even reached
// anywhere past the plane skin-radius distance away from the origin, then the
// origin can never be touched by the simplex.
func (s *GJKSolver) crossedSkinPlane(point dprec.Vec2) bool {
	dot := dprec.Vec2Dot(point, s.searchDirection)
	if dot >= 0 {
		return true // the point is past the plane at the origin so we are good
	}
	return dot*dot <= s.searchDirection.SqrLength()*s.sqrSkinRadius
}
