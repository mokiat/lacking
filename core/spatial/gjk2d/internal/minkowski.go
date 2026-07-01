package internal

import "github.com/mokiat/gomath/dprec"

type MinkowskiShape struct {
	Source     Polygon
	Target     Polygon
	Offset     dprec.Vec2
	SkinRadius float64
}

func (s *MinkowskiShape) MaxIterations() int {
	return len(s.Source.Points) + len(s.Target.Points)
}

func (s *MinkowskiShape) Support(dir dprec.Vec2) MinkowskiVertex {
	sourcePosition, sourceIndex := s.Source.Support(dprec.InverseVec2(dir))
	targetPosition, targetIndex := s.Target.Support(dir)
	return MinkowskiVertex{
		Position: dprec.Vec2Sum(s.Offset, dprec.Vec2Diff(targetPosition, sourcePosition)),
		Refs: RefPair{
			SourceIndex: sourceIndex,
			TargetIndex: targetIndex,
		},
	}
}

type MinkowskiVertex struct {
	Position dprec.Vec2
	Refs     RefPair
}

type RefPair struct {
	SourceIndex int
	TargetIndex int
}
