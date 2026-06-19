package gjk2d

import (
	"github.com/mokiat/gomath/sprec"
	"github.com/mokiat/lacking/core/spatial/gjk2d/internal"
)

type Solver struct{}

func NewSolver() *Solver {
	return &Solver{}
}

func (s *Solver) Intersect(shapeA, shapeB Shape) bool {
	if len(shapeA.Points) == 0 || len(shapeB.Points) == 0 {
		return false
	}

	simplex := internal.NewSimplex(
		len(shapeA.Points)+len(shapeB.Points),
		shapeA.SkinRadius+shapeB.SkinRadius,
	)

	dir := s.pickInitialDirection(&shapeA, &shapeB)
	support := s.minkowskiSupport(&shapeA, &shapeB, dir)
	simplex.Append(support, dir)

	for simplex.CanProgress() {
		dir = simplex.SearchDirection()
		support = s.minkowskiSupport(&shapeA, &shapeB, dir)
		simplex.Append(support, dir)
	}

	return simplex.TouchesOrigin()
}

func (s *Solver) pickInitialDirection(shapeA, shapeB *Shape) sprec.Vec2 {
	return sprec.Vec2Diff(shapeB.Points[0], shapeA.Points[0])
}

func (s *Solver) minkowskiSupport(shapeA, shapeB *Shape, dir sprec.Vec2) sprec.Vec2 {
	supportA := s.shapeSupport(shapeA, sprec.InverseVec2(dir))
	supportB := s.shapeSupport(shapeB, dir)
	return sprec.Vec2Diff(supportB, supportA)
}

func (s *Solver) shapeSupport(shape *Shape, dir sprec.Vec2) sprec.Vec2 {
	best := shape.Points[0]
	bestDot := sprec.Vec2Dot(best, dir)
	for _, v := range shape.Points[1:] {
		if dot := sprec.Vec2Dot(v, dir); dot > bestDot {
			bestDot = dot
			best = v
		}
	}
	return best
}
