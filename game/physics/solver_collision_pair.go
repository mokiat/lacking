package physics

import "github.com/mokiat/gomath/dprec"

type PairCollisionSolverConfig struct {
	PrimaryFrictionCoefficient    float64
	PrimaryRestitutionCoefficient float64
	PrimaryContactNormal          dprec.Vec3
	PrimaryContactPoint           dprec.Vec3

	SecondaryFrictionCoefficient    float64
	SecondaryRestitutionCoefficient float64
	SecondaryContactNormal          dprec.Vec3
	SecondaryContactPoint           dprec.Vec3

	Depth float64
}

type PairCollisionSolver struct{}

var _ PairConstraintSolver = (*PairCollisionSolver)(nil)

func (s *PairCollisionSolver) Init(config PairCollisionSolverConfig) {

}

func (s *PairCollisionSolver) Reset(ctx PairConstraintContext) {

}
func (s *PairCollisionSolver) ApplyImpulses(ctx PairConstraintContext) {

}
func (s *PairCollisionSolver) ApplyNudges(ctx PairConstraintContext) {

}
