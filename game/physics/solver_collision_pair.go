package physics

var _ PairConstraintSolver = (*PairCollisionSolver)(nil)

type PairCollisionSolver struct{}

func (s *PairCollisionSolver) Reset(ctx PairConstraintContext) {

}
func (s *PairCollisionSolver) ApplyImpulses(ctx PairConstraintContext) {

}
func (s *PairCollisionSolver) ApplyNudges(ctx PairConstraintContext) {

}
